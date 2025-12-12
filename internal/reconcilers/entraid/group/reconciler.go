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
	configGcpSyncJobId              = "gcpSyncJobId"
	configGcpSyncRuleId             = "gcpSyncRuleId"

	entraIdGroupPrefix = "dapla-api-test-"
)

type entraIdGroupReconciler struct {
	mainCtx       context.Context
	service       *msgraphsdk.GraphServiceClient
	entraIdConfig entraIdClientConfig
	gcpSyncer     gcpSyncer
	gcpCancelFunc func()
}

type entraIdClientConfig struct {
	ClientId     string
	ClientSecret string
	TenantId     string
}

func (c entraIdClientConfig) Equal(o entraIdClientConfig) bool {
	return (c.ClientId == o.ClientId) && (c.ClientSecret == o.ClientSecret) && (c.TenantId == o.TenantId)
}

func New(ctx context.Context) reconcilers.Reconciler {
	r := &entraIdGroupReconciler{
		mainCtx:   ctx,
		gcpSyncer: NewGcpSyncer(),
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
			{
				Key:         configGcpAppRoleId,
				DisplayName: "GCP Sync App Role ID",
				Description: "ID of App Role to grant on Google SSO/Provisioning Apps in Entra ID",
				Secret:      false,
			},
			{
				Key:         configGcpSSOResourceId,
				DisplayName: "GCP SSO App Resource ID",
				Description: "Resource ID of the Google SSO App in Entra ID",
				Secret:      false,
			},
			{
				Key:         configGcpProvisioningResourceId,
				DisplayName: "GCP Provisioning App Resource ID",
				Description: "Resource ID of the Google Provisioning App in Entra ID",
				Secret:      false,
			},
			{
				Key:         configGcpSyncJobId,
				DisplayName: "GCP Sync Job ID",
				Description: "ID of the Entra ID-Google sync job",
				Secret:      false,
			},
			{
				Key:         configGcpSyncRuleId,
				DisplayName: "GCP Sync Rule Id",
				Description: "ID of the Entra ID GCP Sync Job Rule ID",
				Secret:      false,
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

	// If the GCP Sync settings have changed, we can hotswap the config,
	// since the syncer just reads these as its syncing. In a worst case
	// race condition, the sync will just fail this once.
	gcpSyncConfig, err := r.getGcpSyncConfig(config)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("get gcp sync config")
	}
	if !r.gcpSyncer.Config.Equal(*gcpSyncConfig) {
		r.gcpSyncer.Config = *gcpSyncConfig
	}

	entraId, changed, err := r.getEntraIdClient(config)
	if err != nil {
		return fmt.Errorf("get entra id client: %w", err)
	}

	// If the Entra ID credentials have changed, we need to cancel the currently
	// running GCP sync loop and restart it with the new client.
	if changed {
		log.Info("updating gcpsyncer entra id client")
		r.restartGcpSyncer(entraId)
	}

	// Iterate through all the groups in the team, and reconcile them one by one
	groupsIt := iterator.New(ctx, 100, func(limit, offset int64) (*protoapi.ListTeamGroupsResponse, error) {
		return client.Teams().Groups(ctx, &protoapi.ListTeamGroupsRequest{
			Slug: naisTeam.Slug,
		})
	})

	for groupsIt.Next() {
		if err := r.reconcileGroup(ctx, entraId, client, gcpSyncConfig, groupsIt.Value().Group.Name, log); err != nil {
			return fmt.Errorf("reconcile group %q: %w", groupsIt.Value().Group.Name, err)
		}
	}

	return nil
}

func (r *entraIdGroupReconciler) restartGcpSyncer(entraId *msgraphsdk.GraphServiceClient) {
	if r.gcpCancelFunc != nil {
		r.gcpCancelFunc()
	}
	gcpSyncerCtx, cancel := context.WithCancel(context.Background())
	r.gcpCancelFunc = cancel
	r.gcpSyncer.EntraId = entraId
	go r.gcpSyncer.Run(gcpSyncerCtx)
}

func (r *entraIdGroupReconciler) reconcileGroup(ctx context.Context, entraId *msgraphsdk.GraphServiceClient, client *apiclient.APIClient, gcpSyncConfig *gcpSyncConfig, groupName string, log logrus.FieldLogger) error {
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

		if gcpSyncConfig != nil {
			log.Info("assigning app roles")
			if err := assignAppRoles(ctx, entraId, *group.GetId(), &gcpSyncConfig.GoogleSyncAppRoleId, &gcpSyncConfig.GoogleSyncProvisioningResourceId, &gcpSyncConfig.GoogleSyncSSOResourceId); err != nil {
				return fmt.Errorf("assign provisioning app role: %w", err)
			}
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

	if gcpSyncConfig != nil && (len(usersToAdd) > 0 || len(usersToRemove) > 0) {
		r.gcpSyncer.Queue(*group.GetId(), toUniformIdList(usersToAdd, usersToRemove))
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

func (r *entraIdGroupReconciler) getGcpSyncConfig(config *protoapi.ConfigReconcilerResponse) (*gcpSyncConfig, error) {
	gc := &gcpSyncConfig{}
	for _, c := range config.Nodes {
		switch c.Key {
		case configGcpAppRoleId:
			id, err := uuid.Parse(c.Value)
			if err != nil {
				return nil, fmt.Errorf("parse app role id: %w", err)
			}
			gc.GoogleSyncAppRoleId = id
		case configGcpSSOResourceId:
			id, err := uuid.Parse(c.Value)
			if err != nil {
				return nil, fmt.Errorf("parse sso resource id: %w", err)
			}
			gc.GoogleSyncSSOResourceId = id
		case configGcpProvisioningResourceId:
			id, err := uuid.Parse(c.Value)
			if err != nil {
				return nil, fmt.Errorf("parse provisioning resource id: %w", err)
			}
			gc.GoogleSyncProvisioningResourceId = id
		case configGcpSyncJobId:
			gc.GoogleSyncJobId = c.Value
		case configGcpSyncRuleId:
			gc.GoogleSyncRuleId = c.Value
		case configClientIdKey, configClientSecretKey, configTenantIdKey:
		default:
			return nil, fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if gc.GoogleSyncAppRoleId == uuid.Nil ||
		gc.GoogleSyncProvisioningResourceId == uuid.Nil ||
		gc.GoogleSyncSSOResourceId == uuid.Nil {
		if gc.GoogleSyncAppRoleId == uuid.Nil &&
			gc.GoogleSyncProvisioningResourceId == uuid.Nil &&
			gc.GoogleSyncSSOResourceId == uuid.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("missing one of %s, %s, %s", configGcpAppRoleId, configGcpProvisioningResourceId, configGcpSSOResourceId)
	}

	return gc, nil
}

func (r *entraIdGroupReconciler) getEntraIdClient(config *protoapi.ConfigReconcilerResponse) (*msgraphsdk.GraphServiceClient, bool, error) {
	rc := entraIdClientConfig{}
	for _, c := range config.Nodes {
		switch c.Key {
		case configClientIdKey:
			rc.ClientId = c.Value
		case configClientSecretKey:
			rc.ClientSecret = c.Value
		case configTenantIdKey:
			rc.TenantId = c.Value
		case configGcpAppRoleId, configGcpProvisioningResourceId, configGcpSSOResourceId, configGcpSyncJobId, configGcpSyncRuleId:
		default:
			return nil, false, fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if rc.Equal(r.entraIdConfig) {
		return r.service, false, nil
	}

	creds, err := azidentity.NewClientSecretCredential(rc.TenantId, rc.ClientId, rc.ClientSecret, nil)
	if err != nil {
		return nil, false, fmt.Errorf("create credentials: %w", err)
	}

	r.service, err = msgraphsdk.NewGraphServiceClientWithCredentials(creds, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, false, fmt.Errorf("create graph service client: %w", err)
	}

	r.entraIdConfig = rc

	return r.service, true, nil
}

func (r *entraIdGroupReconciler) Delete(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
