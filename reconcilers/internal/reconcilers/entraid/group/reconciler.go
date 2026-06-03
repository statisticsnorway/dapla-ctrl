package group

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers/entraid/group/master"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers/entraid/group/master/database"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers/entraid/group/master/entraid"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/iterator"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/protoapi"
)

const (
	reconcilerName = "entraid:group"

	configClientIdKey                  = "clientId"
	configTenantIdKey                  = "tenantId"
	configGcpAppRoleIdKey              = "gcpSyncAppRoleId"
	configGcpProvisioningResourceIdKey = "gcpProvisioningResourceId"
	configGcpSSOResourceIdKey          = "gcpSSOResourceId"
	configGroupPrefixKey               = "groupPrefix"
	configMasterKey                    = "master"
	configMasterOverridesKey           = "masterOverrides"
)

var validMasters = []string{"entraid", "database"}

type syncQueuer interface {
	Add(group string, member *string) error
}

type entraIdGroupReconciler struct {
	mainCtx            context.Context
	service            *msgraphsdk.GraphServiceClient
	entraIdConfig      entraIdConfig
	syncQueuer         syncQueuer
	masterHandler      master.Handler
	memberMasterConfig memberMasterConfig
}

type entraIdConfig struct {
	ClientId               string
	TenantId               string
	SSOResourceId          uuid.UUID
	ProvisioningResourceId uuid.UUID
	AppRoleId              uuid.UUID
	GroupPrefix            string
}

type memberMasterConfig struct {
	defaultMaster string
	overrides     string
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
		DisplayName: "Entra ID reconciler",
		Description: "Synchronize team and groups to Entra ID",
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
				Key:         configGcpAppRoleIdKey,
				DisplayName: "GCP Sync App Role ID",
				Description: "ID of App Role to grant on Google SSO/Provisioning Apps in Entra ID",
				Secret:      false,
			},
			{
				Key:         configGcpSSOResourceIdKey,
				DisplayName: "GCP SSO App Resource ID",
				Description: "Resource ID of the Google SSO App in Entra ID",
				Secret:      false,
			},
			{
				Key:         configGcpProvisioningResourceIdKey,
				DisplayName: "GCP Provisioning App Resource ID",
				Description: "Resource ID of the Google Provisioning App in Entra ID",
				Secret:      false,
			},
			{
				Key:         configGroupPrefixKey,
				DisplayName: "Entra ID Group Prefix",
				Description: "Prefix to be added to any group created in Entra ID. Used for testing.",
				Secret:      false,
			},
			{
				Key:         configMasterKey,
				DisplayName: "Master",
				Description: "Which system is source of truth for group membership. Value can be 'entraid' or 'database'.",
				Secret:      false,
			},
			{
				Key:         configMasterOverridesKey,
				DisplayName: "Master Overrides",
				Description: "Override the member master for specific team/groups. Comma-separated list of the form `<team/group>:<name>:<masterName>`. E.g. `team:my-team:database,group:my-team-developers:entraid`",
				Secret:      false,
			},
		},
	}
}

func (r *entraIdGroupReconciler) Name() string {
	return reconcilerName
}

func (r *entraIdGroupReconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	config, err := client.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return fmt.Errorf("get reconciler config: %w", err)
	}

	entraId, err := r.getEntraIdClient(config)
	if err != nil {
		return fmt.Errorf("get entra id client: %w", err)
	}

	if err := r.configureMemberMasters(client, entraId, config); err != nil {
		return fmt.Errorf("configure member masters: %w", err)
	}

	// Iterate through all the groups in the team, and reconcile them one by one
	groupsIt := iterator.New(ctx, 100, func(limit, offset int64) (*protoapi.ListTeamGroupsResponse, error) {
		return client.Teams().Groups(ctx, &protoapi.ListTeamGroupsRequest{
			Slug: daplaTeam.Slug,
		})
	})

	for groupsIt.Next() {
		if err := r.reconcileGroup(ctx, entraId, client, daplaTeam.Slug, groupsIt.Value().Group.Name, log); err != nil {
			return fmt.Errorf("reconcile group %q: %w", groupsIt.Value().Group.Name, err)
		}
	}

	return nil
}

func (r *entraIdGroupReconciler) reconcileGroup(ctx context.Context, entraId *msgraphsdk.GraphServiceClient, client *apiclient.APIClient, teamSlug string, groupName string, log logrus.FieldLogger) error {
	log = log.WithField("groupName", groupName)
	group, created, err := getOrCreateGroup(ctx, entraId, client, groupName, r.entraIdConfig.GroupPrefix)
	if err != nil {
		return fmt.Errorf("get or create group: %w", err)
	}

	if created {
		if _, err := client.Groups().SetExternalId(ctx, &protoapi.SetExternalIdRequest{
			Name:       groupName,
			ExternalId: group.ExternalId,
		}); err != nil {
			return fmt.Errorf("update external id: %w", err)
		}

		if r.syncQueuer != nil {
			if err := r.syncQueuer.Add(group.ExternalId, nil); err != nil {
				return fmt.Errorf("add to sync queue: %w", err)
			}
		}
	}

	if err := ensureAppRoles(ctx, entraId, group.ExternalId, &r.entraIdConfig.AppRoleId, &r.entraIdConfig.ProvisioningResourceId, &r.entraIdConfig.SSOResourceId); err != nil {
		return fmt.Errorf("assign provisioning app role: %w", err)
	}

	dbMembers, err := getDatabaseMembers(ctx, client, groupName)
	if err != nil {
		return fmt.Errorf("get database members: %w", err)
	}

	entraIdUsers, err := getEntraIdMembers(ctx, entraId, group.ExternalId)
	if err != nil {
		return fmt.Errorf("get entra id members: %w", err)
	}

	localOnlyUsers := getDatabaseOnlyUsers(dbMembers, entraIdUsers)
	remoteOnlyUsers := getRemoteOnlyUsers(dbMembers, entraIdUsers)

	if len(localOnlyUsers) == 0 && len(remoteOnlyUsers) == 0 {
		return nil
	}

	master := r.masterHandler.GetMasterFor(teamSlug, groupName)

	if err := master.RemoveUsers(ctx, *group, localOnlyUsers, remoteOnlyUsers, log); err != nil {
		return err
	}

	if err := master.AddUsers(ctx, *group, localOnlyUsers, remoteOnlyUsers, log); err != nil {
		return err
	}

	if r.syncQueuer != nil && (len(localOnlyUsers) > 0 || len(remoteOnlyUsers) > 0) {
		var joinedError error
		for _, u := range slices.Concat(localOnlyUsers, remoteOnlyUsers) {
			err := r.syncQueuer.Add(group.ExternalId, &u.ExternalId)
			if err != nil {
				joinedError = errors.Join(joinedError, err)
			}
		}

		if joinedError != nil {
			return joinedError
		}
	}

	return nil
}

func getDatabaseMembers(ctx context.Context, client *apiclient.APIClient, group string) ([]*protoapi.GroupMember, error) {
	dbMembersIt := iterator.New(ctx, 100, func(limit, offset int64) (*protoapi.ListGroupMembersResponse, error) {
		return client.Groups().Members(ctx, &protoapi.ListGroupMembersRequest{
			Name:   group,
			Limit:  limit,
			Offset: offset,
		})
	})

	var dbMembers []*protoapi.GroupMember
	for dbMembersIt.Next() {
		dbMembers = append(dbMembers, dbMembersIt.Value())
	}
	return dbMembers, dbMembersIt.Err()
}

func getEntraIdMembers(ctx context.Context, entraId *msgraphsdk.GraphServiceClient, groupId string) ([]models.Userable, error) {
	var entraIdUsers []models.Userable
	entraIdUsersReq, err := entraId.Groups().ByGroupId(groupId).TransitiveMembers().GraphUser().Get(ctx, nil)
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

func getOrCreateGroup(ctx context.Context, entraId *msgraphsdk.GraphServiceClient, client *apiclient.APIClient, groupName string, groupPrefix string) (_ *master.Group, created bool, err error) {
	dbGroup, err := client.Groups().Get(ctx, &protoapi.GetGroupRequest{
		Name: groupName,
	})
	if err != nil {
		return nil, false, fmt.Errorf("get group from database: %w", err)
	}
	entraIdGroupName := fmt.Sprintf("%s%s", groupPrefix, groupName)
	if dbGroup.Group.ExternalId != nil {
		_, err := entraId.Groups().ByGroupId(*dbGroup.Group.ExternalId).Get(ctx, nil)
		if err == nil {
			return &master.Group{
				ExternalId: *dbGroup.Group.ExternalId,
				Name:       entraIdGroupName,
			}, false, nil
		}
		var odError odataerrors.ODataErrorable
		if !errors.As(err, &odError) {
			return nil, false, fmt.Errorf("get group from entra id: %w", err)
		} else if code := odError.GetErrorEscaped().GetCode(); code != nil && *code != "404" {
			return nil, false, fmt.Errorf("non-404 status code on get group from entra id: %w", err)
		}
	}

	requestBody := models.NewGroup()
	requestBody.SetDisplayName(&entraIdGroupName)
	requestBody.SetSecurityEnabled(new(true))
	requestBody.SetMailEnabled(new(false))
	requestBody.SetMailNickname(&entraIdGroupName)
	requestBody.SetDescription(new("source:dapla-api"))

	group, err := entraId.Groups().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, false, fmt.Errorf("create group: %w", err)
	}

	return &master.Group{
		ExternalId: *group.GetId(),
		Name:       entraIdGroupName,
	}, true, nil
}

// ensureAppRoles assigns any app roles which may be missing on the given group
func ensureAppRoles(ctx context.Context, entraId *msgraphsdk.GraphServiceClient, groupId string, appRoleId *uuid.UUID, resourceIds ...*uuid.UUID) error {
	appRolesResponse, err := entraId.Groups().ByGroupId(groupId).AppRoleAssignments().Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("get app role assignments for group %q: %w", groupId, err)
	}

	pageIterator, _ := msgraphcore.NewPageIterator[models.AppRoleAssignmentable](appRolesResponse, entraId.GetAdapter(), models.CreateAppRoleAssignmentCollectionResponseFromDiscriminatorValue)

	var appRoleAssignments []models.AppRoleAssignmentable
	if err := pageIterator.Iterate(ctx, func(apa models.AppRoleAssignmentable) bool {
		appRoleAssignments = append(appRoleAssignments, apa)
		return true
	}); err != nil {
		return fmt.Errorf("iterate through app role assignments for group %q: %w", groupId, err)
	}

	for _, resourceId := range resourceIds {
		if !slices.ContainsFunc(appRoleAssignments, func(apa models.AppRoleAssignmentable) bool {
			return apa.GetAppRoleId() != nil && *apa.GetAppRoleId() == *appRoleId && apa.GetResourceId() != nil && *apa.GetResourceId() == *resourceId
		}) {
			if err := assignAppRole(ctx, entraId, groupId, resourceId, appRoleId); err != nil {
				return fmt.Errorf("assign app role %q on resourceId %q for group %q: %w", appRoleId.String(), resourceId.String(), groupId, err)
			}
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
func getDatabaseOnlyUsers(dbUsers []*protoapi.GroupMember, remoteUsers []models.Userable) []master.User {
	dbUserMap := make(map[string]master.User)
	for _, u := range dbUsers {
		dbUserMap[u.User.ExternalId] = master.User{
			ExternalId: u.User.ExternalId,
			Email:      u.User.Email,
		}
	}
	for _, u := range remoteUsers {
		delete(dbUserMap, *u.GetId())
	}
	var dbOnly []master.User
	for _, u := range dbUserMap {
		dbOnly = append(dbOnly, u)
	}
	return dbOnly
}

// getRemoteOnlyUsers takes a list of database users and remote/Entra ID users and returns
// those users which are only present in Entra ID. These are the users that need to be removed
// from the Entra ID group.
func getRemoteOnlyUsers(dbUsers []*protoapi.GroupMember, remoteUsers []models.Userable) []master.User {
	remoteUserMap := make(map[string]master.User)
	for _, u := range remoteUsers {
		remoteUserMap[*u.GetId()] = master.User{
			ExternalId: *u.GetId(),
			Email:      *u.GetUserPrincipalName(),
		}
	}
	for _, u := range dbUsers {
		delete(remoteUserMap, u.User.ExternalId)
	}
	var remoteOnly []master.User
	for _, u := range remoteUserMap {
		remoteOnly = append(remoteOnly, u)
	}
	return remoteOnly
}

func (r *entraIdGroupReconciler) getEntraIdClient(config *protoapi.ConfigReconcilerResponse) (*msgraphsdk.GraphServiceClient, error) {
	rc := entraIdConfig{}
	for _, c := range config.Nodes {
		switch c.Key {
		case configClientIdKey:
			rc.ClientId = c.Value
		case configTenantIdKey:
			rc.TenantId = c.Value
		case configGcpAppRoleIdKey:
			id, err := uuid.Parse(c.Value)
			if err != nil {
				return nil, fmt.Errorf("parse app role id: %w", err)
			}
			rc.AppRoleId = id
		case configGcpSSOResourceIdKey:
			id, err := uuid.Parse(c.Value)
			if err != nil {
				return nil, fmt.Errorf("parse sso resource id: %w", err)
			}
			rc.SSOResourceId = id
		case configGcpProvisioningResourceIdKey:
			id, err := uuid.Parse(c.Value)
			if err != nil {
				return nil, fmt.Errorf("parse provisioning resource id: %w", err)
			}
			rc.ProvisioningResourceId = id
		case configGroupPrefixKey:
			rc.GroupPrefix = c.Value
		case configMasterKey, configMasterOverridesKey:
		default:
			return nil, fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if rc == r.entraIdConfig {
		return r.service, nil
	}

	creds, err := azidentity.NewClientAssertionCredential(rc.TenantId, rc.ClientId, func(ctx context.Context) (string, error) {
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

	service, err := msgraphsdk.NewGraphServiceClientWithCredentials(creds, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, fmt.Errorf("create graph service client: %w", err)
	}

	r.service = service
	r.entraIdConfig = rc

	return service, nil
}

func (r *entraIdGroupReconciler) configureMemberMasters(apiClient *apiclient.APIClient, entraidClient *msgraphsdk.GraphServiceClient, config *protoapi.ConfigReconcilerResponse) error {
	newConfig := memberMasterConfig{}
	for _, c := range config.Nodes {
		switch c.Key {
		case configMasterKey:
			if !slices.Contains(validMasters, c.Value) {
				return fmt.Errorf("unknown master type %q, master must be 'entraid' or 'database'", c.Value)
			}
			newConfig.defaultMaster = c.Value
		case configMasterOverridesKey:
			newConfig.overrides = c.Value
		}
	}

	if newConfig.defaultMaster == "" {
		return errors.New("default member master has to be set")
	}

	if newConfig == r.memberMasterConfig {
		return nil
	}

	masters := []master.MemberMaster{
		entraid.New(apiClient),
		database.New(entraidClient),
	}

	var defaultMaster master.MemberMaster
	for _, m := range masters {
		if m.Name() == newConfig.defaultMaster {
			defaultMaster = m
		}
	}
	if defaultMaster == nil {
		return fmt.Errorf("master %q does not exist", newConfig.defaultMaster)
	}

	handler, err := master.NewHandler(newConfig.overrides, defaultMaster, masters...)
	if err != nil {
		return fmt.Errorf("could not create member master handler: %w", err)
	}

	r.masterHandler = *handler
	r.memberMasterConfig = newConfig

	return nil
}

func (r *entraIdGroupReconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
