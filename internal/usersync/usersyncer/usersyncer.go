package usersyncer

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

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
	tenantDomain  string
	service       *msgraphsdk.GraphServiceClient

	log logrus.FieldLogger
}

type userMap struct {
	byID         map[uuid.UUID]*usersyncsql.User
	byExternalID map[string]*usersyncsql.User
	byEmail      map[string]*usersyncsql.User
}

type entraIdUser struct {
	ID       string
	Email    string
	Name     string
	Section  *string
	JobTitle *string
}

func New(pool *pgxpool.Pool, allUsersGroup, adminGroup, tenantDomain string, service *msgraphsdk.GraphServiceClient, log logrus.FieldLogger) *Usersynchronizer {
	return &Usersynchronizer{
		pool:          pool,
		querier:       usersyncsql.New(pool),
		allUsersGroup: allUsersGroup,
		adminGroup:    adminGroup,
		tenantDomain:  tenantDomain,
		service:       service,
		log:           log,
	}
}

func NewFromConfig(ctx context.Context, pool *pgxpool.Pool, clientId, clientSecret, tenantId, tenantDomain, allUsersGroup, adminGroup string, log logrus.FieldLogger) (*Usersynchronizer, error) {
	cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("create credentials: %w", err)
	}

	srv, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, fmt.Errorf("create graph service client: %w", err)
	}

	return New(pool, allUsersGroup, adminGroup, tenantDomain, srv, log), nil
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

	users, err := getUsers(ctx, querier)
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
				ID:         user.ID,
				Name:       eu.Name,
				Email:      eu.Email,
				ExternalID: eu.ID,
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

	if err := s.assignSectionManagers(ctx, querier, entraIdUsers, entraIdUserMap, s.log); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func AssignDefaultPermissionsToUser(ctx context.Context, querier usersyncsql.Querier, userID uuid.UUID) error {
	defaultUserRoles := []string{
		"Team creator",
	}
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
func (s *Usersynchronizer) assignSectionManagers(ctx context.Context, querier usersyncsql.Querier, entraIdUsers []*entraIdUser, entraIdUserMap map[string]*usersyncsql.User, log logrus.FieldLogger) error {
	sectionCodes, err := querier.GetSectionCodes(ctx)
	if err != nil {
		return fmt.Errorf("get section codes: %w", err)
	}

	sectionManagers := make(map[string][]*usersyncsql.User)
	for _, eu := range entraIdUsers {
		if code := parseSectionManager(sectionCodes, eu, log); code != nil {
			sectionManagers[*code] = append(sectionManagers[*code], entraIdUserMap[eu.ID])
		}
	}

	for sectionCode, manager := range sanitizeSectionManagers(sectionManagers, log) {
		if err := querier.UpdateSectionManager(ctx, usersyncsql.UpdateSectionManagerParams{
			ManagerID:   &manager.ID,
			SectionCode: sectionCode,
		}); err != nil {
			return fmt.Errorf("update section manager for section %s to %s: %w", sectionCode, manager.Email, err)
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

func sanitizeSectionManagers(allSectionManagers map[string][]*usersyncsql.User, log logrus.FieldLogger) map[string]*usersyncsql.User {
	sectionManagers := make(map[string]*usersyncsql.User, len(allSectionManagers))
	for section, managers := range allSectionManagers {
		if len(managers) > 1 {
			var mgs []string
			for _, m := range managers {
				mgs = append(mgs, m.Email)
			}
			log.Warnf("section %s has multiple managers: %s", section, strings.Join(mgs, ", "))
			continue
		}
		sectionManagers[section] = managers[0]
	}
	return sectionManagers
}

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
		Name:       entraIdUser.Name,
		Email:      entraIdUser.Email,
		ExternalID: entraIdUser.ID,
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
			ID:       *user.GetId(),
			Email:    strings.ToLower(*user.GetUserPrincipalName()),
			Name:     *user.GetDisplayName(),
			Section:  user.GetDepartment(),
			JobTitle: user.GetJobTitle(),
		})
		return true
	}); err != nil {
		return nil, fmt.Errorf("list all users group members: %w", err)
	}

	log.WithFields(logrus.Fields{
		"duration":  time.Since(t),
		"num_users": len(users),
	}).Debugf("finished fetching users from Entra ID")
	return users, err
}

// getUsers return a collection of maps of users by ID, external ID and email.
func getUsers(ctx context.Context, querier usersyncsql.Querier) (*userMap, error) {
	users, err := querier.List(ctx)
	if err != nil {
		return nil, err
	}
	ret := &userMap{
		byID:         make(map[uuid.UUID]*usersyncsql.User),
		byExternalID: make(map[string]*usersyncsql.User),
		byEmail:      make(map[string]*usersyncsql.User),
	}
	for _, user := range users {
		ret.byID[user.ID] = user
		ret.byExternalID[user.ExternalID] = user
		ret.byEmail[user.Email] = user
	}

	return ret, nil
}

func parseSectionManager(sectionCodes []string, eu *entraIdUser, log logrus.FieldLogger) (code *string) {
	if eu.Section == nil || eu.JobTitle == nil {
		return nil
	}
	if !strings.HasPrefix(*eu.JobTitle, "Seksjonssjef") {
		return nil
	}
	parts := strings.Split(*eu.Section, " ")
	// Format should be "O|K xxx Seksjon for ..."
	if len(parts) < 3 {
		return nil
	}
	sectionCode := parts[1]
	if !slices.Contains(sectionCodes, sectionCode) {
		log.Infof("encountered section %q, not present in database", sectionCode)
		return nil
	}
	return &sectionCode
}
