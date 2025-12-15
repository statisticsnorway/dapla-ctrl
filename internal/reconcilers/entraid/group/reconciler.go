package group

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/iterator"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/protoapi"
	"k8s.io/utils/ptr"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

const (
	reconcilerName = "entraid:group"

	configClientIdKey               = "clientId"
	configClientSecretKey           = "clientSecret"
	configTenantIdKey               = "tenantId"
	configGcpAppRoleId              = "gcpSyncAppRoleId"
	configGcpProvisioningResourceId = "gcpProvisioningResourceId"
	configGcpSSOResourceId          = "gcpSSOResourceId"

	entraIdGroupPrefix = "dapla-api-test-"
)

type syncQueuer interface {
	Add(group string, member *string)
}

type entraIdGroupReconciler struct {
	mainCtx       context.Context
	service       *msgraphsdk.GraphServiceClient
	entraIdConfig entraIdConfig
	syncQueuer    syncQueuer
}

type entraIdConfig struct {
	ClientId               string
	ClientSecret           string
	TenantId               string
	SSOResourceId          uuid.UUID
	ProvisioningResourceId uuid.UUID
	AppRoleId              uuid.UUID
}

func New(ctx context.Context, sq syncQueuer) reconcilers.Reconciler {
	r := &entraIdGroupReconciler{
		mainCtx:    ctx,
		syncQueuer: sq,
	}

	return r
}

func (r *entraIdGroupReconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "ABC reconciler",
		Description: "Do stupid things",
		MemberAware: true,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         configClientIdKey,
				DisplayName: "Entra ID Client ID",
				Description: "Client ID of the Entra ID client to use for group administration",
				Secret:      false,
			},
			{
				Key:         configTenantIdKey,
				DisplayName: "Entra ID tenant ID",
				Description: "ID of the Entra ID tenant to use",
				Secret:      false,
			},
			{
				Key:         configClientSecretKey,
				DisplayName: "Entra ID Client secret",
				Description: "Client secret of the Entra ID client to use for group administration",
				Secret:      true,
			},
		},
	}
}

func (r *entraIdGroupReconciler) Name() string {
	return reconcilerName
}

func (r *entraIdGroupReconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error {
	config, err := client.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return fmt.Errorf("get reconciler config: %w", err)
	}

	entraId, _, err := r.getEntraIdClient(config)
	if err != nil {
		return fmt.Errorf("get entra id client: %w", err)
	}

	// Iterate through all the groups in the team, and reconcile them one by one
	groupsIt := iterator.New(ctx, 100, func(limit, offset int64) (*protoapi.ListTeamGroupsResponse, error) {
		return client.Teams().Groups(ctx, &protoapi.ListTeamGroupsRequest{
			Slug: naisTeam.Slug,
		})
	})

	for groupsIt.Next() {
		if err := r.reconcileGroup(ctx, entraId, client, groupsIt.Value().Group.Name, log); err != nil {
			return fmt.Errorf("reconcile group %q: %w", groupsIt.Value().Group.Name, err)
		}
	}

	return nil
}

func (r *entraIdGroupReconciler) reconcileGroup(ctx context.Context, entraId *msgraphsdk.GraphServiceClient, client *apiclient.APIClient, groupName string, log logrus.FieldLogger) error {
	group, created, err := getOrCreateGroup(ctx, entraId, client, groupName)
	if err != nil {
		return fmt.Errorf("get or create group: %w", err)
	}

	if created {
		if _, err := client.Groups().SetExternalId(ctx, &protoapi.SetExternalIdRequest{
			Name:       groupName,
			ExternalId: *group.GetId(),
		}); err != nil {
			return fmt.Errorf("update external id: %w", err)
		}

		log.Info("assigning app roles")
		if err := assignAppRoles(ctx, entraId, *group.GetId(), &r.entraIdConfig.AppRoleId, &r.entraIdConfig.ProvisioningResourceId, &r.entraIdConfig.SSOResourceId); err != nil {
			return fmt.Errorf("assign provisioning app role: %w", err)
		}

		if r.syncQueuer != nil {
			r.syncQueuer.Add(*group.GetId(), nil)
		}
	}

	dbMembers, err := getDatabaseMembers(ctx, client, groupName)
	if err != nil {
		return fmt.Errorf("get database members: %w", err)
	}

	entraIdUsers, err := getEntraIdMembers(ctx, entraId, *group.GetId())
	if err != nil {
		return fmt.Errorf("get entra id members: %w", err)
	}

	usersToAdd := getDatabaseOnlyUsers(dbMembers, entraIdUsers)
	usersToRemove := getRemoteOnlyUsers(dbMembers, entraIdUsers)

	if len(usersToAdd) == 0 && len(usersToRemove) == 0 {
		return nil
	}

	for _, u := range usersToRemove {
		if err := entraId.Groups().ByGroupId(*group.GetId()).Members().ByDirectoryObjectId(*u.GetId()).Ref().Delete(ctx, nil); err != nil {
			return fmt.Errorf("remove user %q from group %q: %w", *u.GetUserPrincipalName(), groupName, err)
		}
	}

	for _, u := range usersToAdd {
		requestBody := graphmodels.NewReferenceCreate()
		odataId := fmt.Sprintf("https://graph.microsoft.com/v1.0/directoryObjects/%s", u.User.ExternalId)
		requestBody.SetOdataId(&odataId)
		if err := entraId.Groups().ByGroupId(*group.GetId()).Members().Ref().Post(context.Background(), requestBody, nil); err != nil {
			return fmt.Errorf("add user %q to group %q: %w", u.User.Email, groupName, err)
		}
	}

	if r.syncQueuer != nil && (len(usersToAdd) > 0 || len(usersToRemove) > 0) {
		for _, u := range toUniformIdList(usersToAdd, usersToRemove) {
			r.syncQueuer.Add(*group.GetId(), &u)
		}
	}

	return nil
}

func getDatabaseMembers(ctx context.Context, client *apiclient.APIClient, group string) ([]*protoapi.GroupMember, error) {
	dbMembersIt := iterator.New(ctx, 100, func(limit, offset int64) (*protoapi.ListGroupMembersResponse, error) {
		return client.Groups().Members(ctx, &protoapi.ListGroupMembersRequest{
			Name: group,
		})
	})

	var dbMembers []*protoapi.GroupMember
	for dbMembersIt.Next() {
		if err := dbMembersIt.Err(); err != nil {
			return nil, err
		}
		dbMembers = append(dbMembers, dbMembersIt.Value())
	}
	return dbMembers, nil
}

func getEntraIdMembers(ctx context.Context, entraId *msgraphsdk.GraphServiceClient, groupId string) ([]models.Userable, error) {
	var entraIdUsers []models.Userable
	entraIdUsersReq, err := entraId.Groups().ByGroupId(groupId).Members().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get entra id group members: %w", err)
	}

	pageIterator, err := msgraphcore.NewPageIterator[models.Userable](entraIdUsersReq, entraId.GetAdapter(), models.CreateUserCollectionResponseFromDiscriminatorValue)
	if err != nil {
		return nil, fmt.Errorf("create entra id users pageiterator: %w", err)
	}

	if err := pageIterator.Iterate(ctx, func(user models.Userable) bool {
		entraIdUsers = append(entraIdUsers, user)
		return true
	}); err != nil {
		return nil, fmt.Errorf("list all users group members: %w", err)
	}
	return entraIdUsers, nil
}

func getOrCreateGroup(ctx context.Context, entraId *msgraphsdk.GraphServiceClient, client *apiclient.APIClient, groupName string) (_ models.Groupable, created bool, err error) {
	dbGroup, err := client.Groups().Get(ctx, &protoapi.GetGroupRequest{
		Name: groupName,
	})
	if err != nil {
		return nil, false, fmt.Errorf("get group from database: %w", err)
	}
	if dbGroup.Group.ExternalId != nil {
		entraIdGroup, err := entraId.Groups().ByGroupId(*dbGroup.Group.ExternalId).Get(ctx, nil)
		// TODO: Do we want to handle a 404 (external broup deletion) differently,
		// or do we want to keep it as an error to notify us of something shady going on?
		if err != nil {
			return nil, false, fmt.Errorf("get group from entra id: %w", err)
		}
		return entraIdGroup, false, nil
	}

	// TODO: Remove before prod!
	entraIdGroupName := fmt.Sprintf("%s%s", entraIdGroupPrefix, groupName)

	requestBody := graphmodels.NewGroup()
	requestBody.SetDisplayName(&entraIdGroupName)
	requestBody.SetSecurityEnabled(ptr.To(true))
	requestBody.SetMailEnabled(ptr.To(false))
	requestBody.SetMailNickname(&entraIdGroupName)
	requestBody.SetDescription(ptr.To("source:dapla-api"))

	group, err := entraId.Groups().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, false, fmt.Errorf("create group: %w", err)
	}

	return group, true, nil
}

// toUniformIdList takes two lists of protoapi.GroupMember (database users)
// and models.Userable (Entra ID users) and returns a list of their Entra ID IDs.
// The lists are assumed to be disjunct, so no deduplication of entries is performed.
func toUniformIdList(toAdd []*protoapi.GroupMember, toRemove []models.Userable) []string {
	userIds := make([]string, len(toAdd)+len(toRemove))
	for _, u := range toAdd {
		userIds = append(userIds, u.User.ExternalId)
	}
	for _, u := range toRemove {
		userIds = append(userIds, *u.GetId())
	}
	return userIds
}

// assignAppRoles is a convenience function to assign multiple app roles,
// see assignAppRole
func assignAppRoles(ctx context.Context, entraId *msgraphsdk.GraphServiceClient, groupId string, appRoleId *uuid.UUID, resourceIds ...*uuid.UUID) error {
	for _, resourceId := range resourceIds {
		if err := assignAppRole(ctx, entraId, groupId, resourceId, appRoleId); err != nil {
			return err
		}
	}
	return nil
}

// assignAppRole sets the necessary App Role on the given Entra ID group,
// so that it can be synced by the GCP provisioning app.
func assignAppRole(ctx context.Context, entraId *msgraphsdk.GraphServiceClient, groupId string, resourceId *uuid.UUID, appRoleId *uuid.UUID) error {
	gcpSyncAssignment := models.NewAppRoleAssignment()
	gcpSyncAssignment.SetAppRoleId(appRoleId)
	gcpSyncAssignment.SetResourceId(resourceId)
	groupUuid, err := uuid.Parse(groupId)
	if err != nil {
		return fmt.Errorf("parse group id: %w", err)
	}
	gcpSyncAssignment.SetPrincipalId(&groupUuid)
	if _, err := entraId.Groups().ByGroupId(groupId).AppRoleAssignments().Post(ctx, gcpSyncAssignment, nil); err != nil {
		return fmt.Errorf("create app role assignment: %w", err)
	}

	return nil
}

// getDatabaseOnlyUsers takes a list of database users and remote/Entra ID users and returns
// those users which are only present in the database. These are the users that need to be added
// to the Entra ID group.
func getDatabaseOnlyUsers(dbUsers []*protoapi.GroupMember, remoteUsers []models.Userable) []*protoapi.GroupMember {
	dbUserMap := make(map[string]*protoapi.GroupMember)
	for _, u := range dbUsers {
		dbUserMap[u.User.ExternalId] = u
	}
	for _, u := range remoteUsers {
		delete(dbUserMap, *u.GetId())
	}
	var dbOnly []*protoapi.GroupMember
	for _, u := range dbUserMap {
		dbOnly = append(dbOnly, u)
	}
	return dbOnly
}

// getRemoteOnlyUsers takes a list of database users and remote/Entra ID users and returns
// those users which are only present in Entra ID. These are the users that need to be removed
// from the Entra ID group.
func getRemoteOnlyUsers(dbUsers []*protoapi.GroupMember, remoteUsers []models.Userable) []models.Userable {
	remoteUserMap := make(map[string]models.Userable)
	for _, u := range remoteUsers {
		remoteUserMap[*u.GetId()] = u
	}
	for _, u := range dbUsers {
		delete(remoteUserMap, u.User.ExternalId)
	}
	var remoteOnly []models.Userable
	for _, u := range remoteUserMap {
		remoteOnly = append(remoteOnly, u)
	}
	return remoteOnly
}

func (r *entraIdGroupReconciler) getEntraIdClient(config *protoapi.ConfigReconcilerResponse) (*msgraphsdk.GraphServiceClient, bool, error) {
	rc := entraIdConfig{}
	for _, c := range config.Nodes {
		switch c.Key {
		case configClientIdKey:
			rc.ClientId = c.Value
		case configClientSecretKey:
			rc.ClientSecret = c.Value
		case configTenantIdKey:
			rc.TenantId = c.Value
		case configGcpAppRoleId:
			id, err := uuid.Parse(c.Value)
			if err != nil {
				return nil, false, fmt.Errorf("parse app role id: %w", err)
			}
			rc.AppRoleId = id
		case configGcpSSOResourceId:
			id, err := uuid.Parse(c.Value)
			if err != nil {
				return nil, false, fmt.Errorf("parse sso resource id: %w", err)
			}
			rc.SSOResourceId = id
		case configGcpProvisioningResourceId:
			id, err := uuid.Parse(c.Value)
			if err != nil {
				return nil, false, fmt.Errorf("parse provisioning resource id: %w", err)
			}
			rc.ProvisioningResourceId = id
		default:
			return nil, false, fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if rc == r.entraIdConfig {
		return r.service, false, nil
	}

	creds, err := azidentity.NewClientSecretCredential(rc.TenantId, rc.ClientId, rc.ClientSecret, nil)
	if err != nil {
		return nil, false, fmt.Errorf("create credentials: %w", err)
	}

	service, err := msgraphsdk.NewGraphServiceClientWithCredentials(creds, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, false, fmt.Errorf("create graph service client: %w", err)
	}

	r.service = service
	r.entraIdConfig = rc

	return service, true, nil
}

func (r *entraIdGroupReconciler) Delete(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
