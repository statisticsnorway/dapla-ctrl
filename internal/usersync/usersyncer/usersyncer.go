package usersyncer

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/internal/usersync/usersyncsql"
	"k8s.io/utils/ptr"
)

type Usersynchronizer struct {
	pool          *pgxpool.Pool
	querier       *usersyncsql.Queries
	adminGroup    string
	allUsersGroup string
	service       *msgraphsdk.GraphServiceClient

	log logrus.FieldLogger
}

type userMap struct {
	byID         map[uuid.UUID]*usersyncsql.User
	byExternalID map[string]*usersyncsql.User
	byEmail      map[string]*usersyncsql.User
}

type entraIdUser struct {
	ID          string
	Email       string
	Name        string
	SectionCode *string
	JobTitle    *string
}

func New(pool *pgxpool.Pool, allUsersGroup, adminGroup string, service *msgraphsdk.GraphServiceClient, log logrus.FieldLogger) *Usersynchronizer {
	return &Usersynchronizer{
		pool:          pool,
		querier:       usersyncsql.New(pool),
		allUsersGroup: allUsersGroup,
		adminGroup:    adminGroup,
		service:       service,
		log:           log,
	}
}

func NewFromConfig(ctx context.Context, pool *pgxpool.Pool, clientId, tenantId, allUsersGroup, adminGroup string, log logrus.FieldLogger) (*Usersynchronizer, error) {
	creds, err := azidentity.NewClientAssertionCredential(tenantId, clientId, func(ctx context.Context) (string, error) {
		creds, err := idtoken.NewCredentials(&idtoken.Options{Audience: "api://AzureADTokenExchange"})
		if err != nil {
			return "", err
		}
		token, err := creds.Token(ctx)
		if err != nil {
			return "", err
		}
		return token.Value, nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("exchange for azure credentials: %w", err)
	}

	srv, err := msgraphsdk.NewGraphServiceClientWithCredentials(creds, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, fmt.Errorf("create graph service client: %w", err)
	}

	return New(pool, allUsersGroup, adminGroup, srv, log), nil
}

// Sync fetches all users from Entra ID and adds them as users in Nais API.
//
// If a user already exist in Nais API the user will get the name and email potentially updated if it has changed in
// Entra ID
//
// After all users have been synced, users that have an email address that matches the tenant domain that no longer
// exist in Entra ID will be removed.
//
// All users present in the admin group in Entra ID will also be granted the admin role in Nais API, and
// existing admins that no longer exist in the admin group will get the admin role revoked.
func (s *Usersynchronizer) Sync(ctx context.Context) error {
	entraIdUsers, err := s.getEntraIdUsers(ctx, s.log)
	if err != nil {
		return fmt.Errorf("get users from entra id: %w", err)
	}

	sections, err := s.querier.GetSections(ctx)
	if err != nil {
		return fmt.Errorf("get sections: %w", err)
	}

	entraIdUsers = sanitizeUserSections(sections, entraIdUsers)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err == nil {
			return
		} else if !errors.Is(err, pgx.ErrTxClosed) {
			s.log.WithError(err).Errorf("rollback transaction")
		}
	}()
	querier := s.querier.WithTx(tx)

	users, err := getDbUsers(ctx, querier)
	if err != nil {
		return fmt.Errorf("get existing users: %w", err)
	}

	entraIdUserMap := make(map[string]*usersyncsql.User)
	for _, eu := range entraIdUsers {
		user, err := getOrCreateUserFromEntraIdUser(ctx, querier, eu, users, s.log)
		if err != nil {
			return fmt.Errorf("get or create user %q: %w", eu.Email, err)
		}

		if userIsOutdated(user, eu) {
			if err := querier.Update(ctx, usersyncsql.UpdateParams{
				ID:          user.ID,
				Name:        eu.Name,
				Email:       eu.Email,
				ExternalID:  eu.ID,
				SectionCode: eu.SectionCode,
			}); err != nil {
				return fmt.Errorf("update user %q: %w", eu.Email, err)
			}

			if err := querier.CreateLogEntry(ctx, usersyncsql.CreateLogEntryParams{
				Action:       usersyncsql.UsersyncLogEntryActionUpdateUser,
				UserID:       user.ID,
				UserName:     eu.Name,
				UserEmail:    eu.Email,
				OldUserName:  &user.Name,
				OldUserEmail: &user.Email,
			}); err != nil {
				s.log.WithError(err).Errorf("create user sync log entry")
			}
		}

		entraIdUserMap[eu.ID] = user

		// remove user from map to keep track of users that no longer exist in Entra ID
		delete(users.byID, user.ID)
	}

	if err := deleteUnknownUsers(ctx, querier, users.byID, s.log); err != nil {
		return err
	}

	if err := s.assignAdmins(ctx, querier, entraIdUserMap, s.log); err != nil {
		return err
	}

	if err := s.assignSectionManagers(ctx, sections, querier, entraIdUsers, entraIdUserMap, s.log); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func AssignDefaultPermissionsToUser(ctx context.Context, querier usersyncsql.Querier, userID uuid.UUID) error {
	defaultUserRoles := []string{}
	for _, roleName := range defaultUserRoles {
		if err := querier.AssignGlobalRole(ctx, usersyncsql.AssignGlobalRoleParams{
			UserID:   userID,
			RoleName: roleName,
		}); err != nil {
			return err
		}
	}
	return nil
}

// deleteUnknownUsers will delete users from Nais API that does not exist in Entra ID.
func deleteUnknownUsers(ctx context.Context, querier usersyncsql.Querier, unknownUsers map[uuid.UUID]*usersyncsql.User, log logrus.FieldLogger) error {
	for _, user := range unknownUsers {
		if err := querier.Delete(ctx, user.ID); err != nil {
			return fmt.Errorf("delete user %q: %w", user.Email, err)
		}
		if err := querier.CreateLogEntry(ctx, usersyncsql.CreateLogEntryParams{
			Action:    usersyncsql.UsersyncLogEntryActionDeleteUser,
			UserID:    user.ID,
			UserName:  user.Name,
			UserEmail: user.Email,
		}); err != nil {
			log.WithError(err).Errorf("create user sync log entry")
		}
	}

	return nil
}

// assignAdmins assigns the global admin role to members of the admin group in the Entra ID.
// Existing admins that is no longer a member of the admin group will have the admin role revoked.
func (s *Usersynchronizer) assignAdmins(ctx context.Context, querier usersyncsql.Querier, entraIdUsers map[string]*usersyncsql.User, log logrus.FieldLogger) error {
	admins, err := s.getAdminGroupMembers(ctx, entraIdUsers, log)
	if err != nil {
		return err
	}

	existingAdmins, err := querier.ListGlobalAdmins(ctx)
	if err != nil {
		return err
	}

	for _, existingAdmin := range existingAdmins {
		if _, shouldBeAdmin := admins[existingAdmin.ID]; !shouldBeAdmin {
			log.WithField("email", existingAdmin.Email).Infof("revoke admin role")
			if err := querier.RevokeGlobalAdmin(ctx, existingAdmin.ID); err != nil {
				return err
			}

			if err := querier.CreateLogEntry(ctx, usersyncsql.CreateLogEntryParams{
				Action:    usersyncsql.UsersyncLogEntryActionRevokeRole,
				UserID:    existingAdmin.ID,
				UserName:  existingAdmin.Name,
				UserEmail: existingAdmin.Email,
				RoleName:  ptr.To("Admin"),
			}); err != nil {
				log.WithError(err).Errorf("create user sync log entry")
			}
		}
	}

	for _, admin := range admins {
		if !admin.Admin {
			log.WithField("email", admin.Email).Infof("assign admin role")
			if err := querier.AssignGlobalAdmin(ctx, admin.ID); err != nil {
				return err
			}

			if err := querier.CreateLogEntry(ctx, usersyncsql.CreateLogEntryParams{
				Action:    usersyncsql.UsersyncLogEntryActionAssignRole,
				UserID:    admin.ID,
				UserName:  admin.Name,
				UserEmail: admin.Email,
				RoleName:  ptr.To("Admin"),
			}); err != nil {
				log.WithError(err).Errorf("create user sync log entry")
			}
		}
	}

	return nil
}

// assignSectionManagers sets the section manager for each section in the database, if it finds
// a matching user in Entra ID (someone in the given section with the job title '^Seksjonssjef.*')
func (s *Usersynchronizer) assignSectionManagers(ctx context.Context, sections []*usersyncsql.Section, querier usersyncsql.Querier, entraIdUsers []*entraIdUser, entraIdUserMap map[string]*usersyncsql.User, log logrus.FieldLogger) error {
	// Get the definitive list of section managers from Entra ID
	entraIdSectionManagers := parseEntraIdSectionManagers(entraIdUsers, entraIdUserMap, log)

	// Compare the managers in the DB to the ones in Entra ID and get a
	// changeset of sections and their new managers
	updatedSectionManagers := getSectionManagerChanges(sections, entraIdSectionManagers)

	for sectionCode, manager := range updatedSectionManagers {
		if err := querier.UpdateSectionManager(ctx, usersyncsql.UpdateSectionManagerParams{
			ManagerID:   manager,
			SectionCode: sectionCode,
		}); err != nil {
			return fmt.Errorf("update section manager for section %s: %w", sectionCode, err)
		}
	}
	return nil
}

// Saved for fun! lern it
// func sanitizedManagers(allManagers map[string][]*usersyncsql.User, log logrus.FieldLogger) func(yield func(k string, v *usersyncsql.User) bool) {
// 	return func(yield func(k string, v *usersyncsql.User) bool) {
// 		for section, managers := range allManagers {
// 			if len(managers) > 1 {
// 				var mgs []string
// 				for _, m := range managers {
// 					mgs = append(mgs, m.Email)
// 				}
// 				log.Warnf("section %s has multiple managers: %s", section, strings.Join(mgs, ", "))
// 				continue
// 			}
// 			if ok := yield(section, managers[0]); !ok {
// 				return
// 			}
// 		}
// 	}
// }

// getAdminGroupMembers fetches all users in the admin group from the Entra ID group of the tenant.
func (s *Usersynchronizer) getAdminGroupMembers(ctx context.Context, entraIdUsers map[string]*usersyncsql.User, log logrus.FieldLogger) (map[uuid.UUID]*usersyncsql.User, error) {
	collection, err := s.service.Groups().ByGroupId(s.adminGroup).TransitiveMembers().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get admin group members: %w", err)
	}

	pageIterator, _ := msgraphcore.NewPageIterator[models.Userable](collection, s.service.GetAdapter(), models.CreateUserCollectionResponseFromDiscriminatorValue)

	groupMembers := make([]models.Userable, 0)
	if err := pageIterator.Iterate(ctx, func(user models.Userable) bool {
		groupMembers = append(groupMembers, user)
		return true
	}); err != nil {
		return nil, fmt.Errorf("iterate through admin group members: %w", err)
	}

	admins := make(map[uuid.UUID]*usersyncsql.User)

	for _, member := range groupMembers {
		admin, exists := entraIdUsers[*member.GetId()]
		if !exists {
			log.WithField("email", *member.GetUserPrincipalName()).Errorf("unknown user in admins groups")
			continue
		}

		admins[admin.ID] = admin
	}

	return admins, nil
}

// userIsOutdated checks if a user needs to get its name or its email address updated.
func userIsOutdated(user *usersyncsql.User, eu *entraIdUser) bool {
	if user.Name != eu.Name {
		return true
	}

	if !strings.EqualFold(user.Email, eu.Email) {
		return true
	}

	if user.ExternalID != eu.ID {
		return true
	}

	if user.SectionCode != eu.SectionCode {
		return true
	}

	return false
}

// getOrCreateUserFromEntraIdUser will return a user for an Entra ID user, creating it first if needed.
func getOrCreateUserFromEntraIdUser(ctx context.Context, querier usersyncsql.Querier, entraIdUser *entraIdUser, existingUsers *userMap, log logrus.FieldLogger) (*usersyncsql.User, error) {
	if existingUser, exists := existingUsers.byExternalID[entraIdUser.ID]; exists {
		return existingUser, nil
	}

	if existingUser, exists := existingUsers.byEmail[entraIdUser.Email]; exists {
		return existingUser, nil
	}

	createdUser, err := querier.Create(ctx, usersyncsql.CreateParams{
		Name:        entraIdUser.Name,
		Email:       entraIdUser.Email,
		ExternalID:  entraIdUser.ID,
		SectionCode: entraIdUser.SectionCode,
	})
	if err != nil {
		return nil, err
	}

	if err := AssignDefaultPermissionsToUser(ctx, querier, createdUser.ID); err != nil {
		return nil, err
	}

	if err := querier.CreateLogEntry(ctx, usersyncsql.CreateLogEntryParams{
		Action:    usersyncsql.UsersyncLogEntryActionCreateUser,
		UserID:    createdUser.ID,
		UserName:  createdUser.Name,
		UserEmail: createdUser.Email,
	}); err != nil {
		log.WithError(err).Errorf("create user sync log entry")
	}

	return createdUser, nil
}

// getEntraIdUsers fetches all users from Entra ID.
func (s *Usersynchronizer) getEntraIdUsers(ctx context.Context, log logrus.FieldLogger) ([]*entraIdUser, error) {
	users := make([]*entraIdUser, 0)

	log.Debugf("start fetching users from Entra ID")
	t := time.Now()

	usersResponse, err := s.service.Groups().ByGroupId(s.allUsersGroup).TransitiveMembers().Get(ctx, &groups.ItemTransitiveMembersRequestBuilderGetRequestConfiguration{
		QueryParameters: &groups.ItemTransitiveMembersRequestBuilderGetQueryParameters{
			Select: []string{"department", "jobTitle", "id", "email", "displayName", "userPrincipalName"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get all users group: %w", err)
	}

	pageIterator, err := msgraphcore.NewPageIterator[models.Userable](usersResponse, s.service.GetAdapter(), models.CreateUserCollectionResponseFromDiscriminatorValue)
	if err != nil {
		return nil, fmt.Errorf("create pageiterator: %w", err)
	}

	if err := pageIterator.Iterate(ctx, func(user models.Userable) bool {
		users = append(users, &entraIdUser{
			ID:          *user.GetId(),
			Email:       strings.ToLower(*user.GetUserPrincipalName()),
			Name:        *user.GetDisplayName(),
			SectionCode: parseSectionCode(user.GetDepartment()),
			JobTitle:    user.GetJobTitle(),
		})
		return true
	}); err != nil {
		return nil, fmt.Errorf("list all users group members: %w", err)
	}

	log.WithFields(logrus.Fields{
		"duration":  time.Since(t),
		"num_users": len(users),
	}).Infof("finished fetching users from Entra ID")
	return users, err
}

// getDbUsers return a collection of maps of users by ID, external ID and email.
func getDbUsers(ctx context.Context, querier usersyncsql.Querier) (*userMap, error) {
	users, err := querier.List(ctx)
	if err != nil {
		return nil, err
	}
	ret := &userMap{
		byID:         make(map[uuid.UUID]*usersyncsql.User, len(users)),
		byExternalID: make(map[string]*usersyncsql.User, len(users)),
		byEmail:      make(map[string]*usersyncsql.User, len(users)),
	}
	for _, user := range users {
		ret.byID[user.ID] = user
		ret.byExternalID[user.ExternalID] = user
		ret.byEmail[user.Email] = user
	}

	return ret, nil
}

// parseEntraIdSectionManagers creates a map of sections to their manager, for all
// sections with one definite manager
func parseEntraIdSectionManagers(entraIdUsers []*entraIdUser, entraIdUserMap map[string]*usersyncsql.User, log logrus.FieldLogger) map[string]*usersyncsql.User {
	entraIdSectionManagers := make(map[string][]*usersyncsql.User)
	for _, eu := range entraIdUsers {
		if isSectionManager(eu) {
			entraIdSectionManagers[*eu.SectionCode] = append(entraIdSectionManagers[*eu.SectionCode], entraIdUserMap[eu.ID])
		}
	}
	return sanitizeSectionManagers(entraIdSectionManagers, log)
}

// parseEntraIdSectionManager checks whether the user has the Seksjonssjef job title,
// and whether they have a valid section. If so, we return the section code, otherwise we
// return nil.
func isSectionManager(eu *entraIdUser) bool {
	if eu.SectionCode == nil || eu.JobTitle == nil {
		return false
	}
	return strings.HasPrefix(*eu.JobTitle, "Seksjonssjef")
}

// parseSectionCode tries to extract the section code from the given section name,
// using the format "O|K xxx Seksjon for ...". Returns nil if it does not work.
func parseSectionCode(section *string) *string {
	if section == nil {
		return nil
	}
	parts := strings.Split(*section, " ")
	// Format should be "O|K xxx Seksjon for ..."
	if len(parts) < 3 {
		return nil
	}

	return &parts[1]
}

// sanitizeSectionManagers goes through the list of the (potentially multiple) managers for each section,
// and returns a map of only those sections with one definite manager.
func sanitizeSectionManagers(allSectionManagers map[string][]*usersyncsql.User, log logrus.FieldLogger) map[string]*usersyncsql.User {
	sectionManagers := make(map[string]*usersyncsql.User, len(allSectionManagers))
	for section, managers := range allSectionManagers {
		// There is a definite manager for this section
		if len(managers) == 1 {
			sectionManagers[section] = managers[0]
			continue
		}
		if len(managers) == 0 {
			continue
		}
		// There are multiple managers for the section, which we treat as there being none
		var mgs []string
		for _, m := range managers {
			mgs = append(mgs, m.Email)
		}
		log.WithFields(logrus.Fields{"section": section, "managers": strings.Join(mgs, ", ")}).Warnf("section has multiple managers")
	}
	return sectionManagers
}

// getSectionManagerChanges creates a map of the changes in section managers in Entra ID compared to the database.
func getSectionManagerChanges(sections []*usersyncsql.Section, entraIdSectionManagers map[string]*usersyncsql.User) map[string]*uuid.UUID {
	changes := make(map[string]*uuid.UUID)
	for _, section := range sections {
		newManager := entraIdSectionManagers[section.Code]
		var newManagerId *uuid.UUID
		if newManager != nil {
			newManagerId = &newManager.ID
		}
		if managerHasChanged(section.ManagerID, newManagerId) {
			changes[section.Code] = newManagerId
		}
	}
	return changes
}

// managerHasChanged checks whether the manager in the database is outdated compared to Entra ID
func managerHasChanged(oldManagerId *uuid.UUID, newManagerId *uuid.UUID) bool {
	if oldManagerId == newManagerId {
		return false
	}
	if oldManagerId == nil || newManagerId == nil {
		return true
	}
	return *oldManagerId != *newManagerId
}

func sanitizeUserSections(validSections []*usersyncsql.Section, users []*entraIdUser) []*entraIdUser {
	for _, u := range users {
		if u.SectionCode != nil && !slices.ContainsFunc(validSections, func(s *usersyncsql.Section) bool {
			return *u.SectionCode == s.Code
		}) {
			u.SectionCode = nil
		}
	}
	return users
}
