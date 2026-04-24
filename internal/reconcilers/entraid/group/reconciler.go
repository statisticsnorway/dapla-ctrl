package group

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/entraidclient"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers/entraid/group/master"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers/entraid/group/master/database"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers/entraid/group/master/entraid"
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

type entraIdClient interface {
	AddUserToGroup(ctx context.Context, groupId, userId string) error
	RemoveUserFromGroup(ctx context.Context, groupId, userId string) error
	CreateGroup(ctx context.Context, groupName string) (models.Groupable, error)
	GetGroup(ctx context.Context, groupId string) (models.Groupable, error)
	GetTransitiveMembers(ctx context.Context, groupId string) ([]models.Userable, error)
	AssignAppRoleToGroup(ctx context.Context, groupId string, resourceId *uuid.UUID, appRoleId *uuid.UUID) error
	GetAppRolesForGroup(ctx context.Context, groupId string) ([]models.AppRoleAssignmentable, error)
}

type entraIdGroupReconciler struct {
	mainCtx             context.Context
	entraIdClient       entraIdClient
	entraIdConfig       entraIdConfig
	staticEntraIdClient bool
	syncQueuer          syncQueuer
	masterHandler       master.Handler
	memberMasterConfig  memberMasterConfig
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

type OptFunc func(*entraIdGroupReconciler)

func WithEntraIdClient(client entraIdClient) OptFunc {
	return func(r *entraIdGroupReconciler) {
		r.entraIdClient = client
		r.staticEntraIdClient = true
	}
}

func New(ctx context.Context, sq syncQueuer, opts ...OptFunc) reconcilers.Reconciler {
	r := &entraIdGroupReconciler{
		mainCtx:    ctx,
		syncQueuer: sq,
	}

	for _, opt := range opts {
		opt(r)
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

func (r *entraIdGroupReconciler) reconcileGroup(ctx context.Context, entraId entraIdClient, client *apiclient.APIClient, teamSlug string, groupName string, log logrus.FieldLogger) error {
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

	entraIdUsers, err := entraId.GetTransitiveMembers(ctx, group.ExternalId)
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

func getOrCreateGroup(ctx context.Context, entraId entraIdClient, client *apiclient.APIClient, groupName string, groupPrefix string) (_ *master.Group, created bool, err error) {
	dbGroup, err := client.Groups().Get(ctx, &protoapi.GetGroupRequest{
		Name: groupName,
	})
	if err != nil {
		return nil, false, fmt.Errorf("get group from database: %w", err)
	}
	entraIdGroupName := fmt.Sprintf("%s%s", groupPrefix, groupName)
	if dbGroup.Group.ExternalId != nil {
		if _, err := entraId.GetGroup(ctx, *dbGroup.Group.ExternalId); err == nil {
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

	group, err := entraId.CreateGroup(ctx, entraIdGroupName)
	if err != nil {
		return nil, false, fmt.Errorf("create group: %w", err)
	}

	return &master.Group{
		ExternalId: *group.GetId(),
		Name:       entraIdGroupName,
	}, true, nil
}

// ensureAppRoles assigns any app roles which may be missing on the given group
func ensureAppRoles(ctx context.Context, entraId entraIdClient, groupId string, appRoleId *uuid.UUID, resourceIds ...*uuid.UUID) error {
	appRoleAssignments, err := entraId.GetAppRolesForGroup(ctx, groupId)
	if err != nil {
		return fmt.Errorf("get app role assignments for group %q: %w", groupId, err)
	}

	for _, resourceId := range resourceIds {
		if !slices.ContainsFunc(appRoleAssignments, func(apa models.AppRoleAssignmentable) bool {
			return apa.GetAppRoleId() != nil && *apa.GetAppRoleId() == *appRoleId && apa.GetResourceId() != nil && *apa.GetResourceId() == *resourceId
		}) {
			if err := entraId.AssignAppRoleToGroup(ctx, groupId, resourceId, appRoleId); err != nil {
				return fmt.Errorf("assign app role %q on resourceId %q for group %q: %w", appRoleId.String(), resourceId.String(), groupId, err)
			}
		}
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

func (r *entraIdGroupReconciler) getEntraIdClient(config *protoapi.ConfigReconcilerResponse) (entraIdClient, error) {
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
		case configMasterKey:
		default:
			return nil, fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if rc == r.entraIdConfig {
		return r.entraIdClient, nil
	}

	client, err := entraidclient.New(rc.TenantId, rc.ClientId)
	if err != nil {
		return nil, fmt.Errorf("create entraid client: %w", err)
	}

	r.entraIdClient = client
	r.entraIdConfig = rc

	return client, nil
}

func (r *entraIdGroupReconciler) configureMemberMasters(apiClient *apiclient.APIClient, entraidClient entraIdClient, config *protoapi.ConfigReconcilerResponse) error {
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
