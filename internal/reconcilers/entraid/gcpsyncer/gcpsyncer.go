package gcpsyncer

import (
	"context"
	"fmt"
	"slices"
	"time"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphserviceprincipals "github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/queue"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/protoapi"
	"k8s.io/utils/ptr"
)

const (
	reconcilerName = "entraid:gcpsync"

	configClientIdKey               = "clientId"
	configClientSecretKey           = "clientSecret"
	configTenantIdKey               = "tenantId"
	configGcpProvisioningResourceId = "gcpProvisioningResourceId"
	configGcpSyncJobId              = "gcpSyncJobId"
	configGcpSyncRuleId             = "gcpSyncRuleId"
	configGcpSyncInterval           = "gcpSyncInterval"
)

type gcpSyncConfig struct {
	ClientId                         string
	ClientSecret                     string
	TenantId                         string
	GoogleSyncProvisioningResourceId uuid.UUID
	GoogleSyncJobId                  string
	GoogleSyncRuleId                 string
	SyncInterval                     time.Duration
}

type gcpSyncReconciler struct {
	apiClient    *apiclient.APIClient
	Config       gcpSyncConfig
	Queue        queue.Queue[SyncRequest]
	queueChannel <-chan SyncRequest
	entraId      *msgraphsdkgo.GraphServiceClient
	log          logrus.FieldLogger
}

type SyncRequest struct {
	Group string
	User  *string
}

func New(client *apiclient.APIClient) *gcpSyncReconciler {
	queue, channel := queue.NewQueue[SyncRequest]()
	return &gcpSyncReconciler{
		apiClient:    client,
		Queue:        queue,
		queueChannel: channel,
		log: logrus.New().
			WithField("subsystem", "gcpSyncer"),
	}
}

func (r *gcpSyncReconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "GCP Syncer",
		Description: "Syncs groups and group memberships to GCP ASAP",
		MemberAware: true,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         configClientIdKey,
				DisplayName: "Entra ID Client ID",
				Description: "Client ID of the Entra ID client to use",
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
				Description: "Client secret of the Entra ID client to use",
				Secret:      true,
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
			{
				Key:         configGcpSyncInterval,
				DisplayName: "GCP Sync interval",
				Description: "Max duration before a sync is triggered after first request, as a Go duration",
				Secret:      false,
			},
		},
	}
}

func (r *gcpSyncReconciler) Name() string {
	return reconcilerName
}

func (s *gcpSyncReconciler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		groups, err := s.CollectRequests(ctx)
		if err != nil {
			s.log.Errorf("error collecting sync requests: %s", err)
			continue
		}
		if err := s.Sync(ctx, groups); err != nil {
			s.log.Errorf("error running gcp sync: %s", err)
		}
	}
}

func (s *gcpSyncReconciler) CollectRequests(ctx context.Context) (map[string][]string, error) {
	syncTimer := &time.Timer{}
	groupsToSync := make(map[string][]string)
	for {
		select {
		case r := <-s.queueChannel:
			if syncTimer.C == nil {
				s.log.Info("received first sync request, starting collection interval", "interval", s.Config.SyncInterval)
				syncTimer = time.NewTimer(s.Config.SyncInterval)
			}
			if r.User == nil {
				if _, ok := groupsToSync[r.Group]; !ok {
					groupsToSync[r.Group] = nil
				}
			} else if !slices.Contains(groupsToSync[r.Group], *r.User) {
				groupsToSync[r.Group] = append(groupsToSync[r.Group], *r.User)
			}
			// Remove duplicates, stop if over 4 users in group (5 is limit for sync job)
			if len(groupsToSync[r.Group]) > 4 {
				s.log.Info("group has 5 or more member updates, splitting..")
				return groupsToSync, nil
			}

		case <-ctx.Done():
			return nil, context.Canceled
		case <-syncTimer.C:
			return groupsToSync, nil
		}
	}
}

func (s *gcpSyncReconciler) Sync(ctx context.Context, groups map[string][]string) error {
	config, err := s.apiClient.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: s.Name(),
	})
	if err != nil {
		return fmt.Errorf("get reocnciler config: %w", err)
	}

	entraId, err := s.getEntraIdClient(config)
	if err != nil {
		return fmt.Errorf("get entra id client: %w", err)
	}

	if len(groups) == 0 {
		s.log.Info("no groups to sync")
		return nil
	}

	requestBody := graphserviceprincipals.NewItemSynchronizationJobsItemProvisionOnDemandPostRequestBody()

	var parameters []models.SynchronizationJobApplicationParametersable

	for group, users := range groups {
		syncParamSet := syncJobParameterSet(s.Config.GoogleSyncRuleId, group, users)
		parameters = append(parameters, syncParamSet)
	}

	requestBody.SetParameters(parameters)

	_, err = entraId.ServicePrincipals().ByServicePrincipalId(s.Config.GoogleSyncProvisioningResourceId.String()).
		Synchronization().Jobs().BySynchronizationJobId(s.Config.GoogleSyncJobId).ProvisionOnDemand().
		Post(ctx, requestBody, nil)

	return err
}
func (r *gcpSyncReconciler) getEntraIdClient(config *protoapi.ConfigReconcilerResponse) (*msgraphsdk.GraphServiceClient, error) {
	gc := gcpSyncConfig{}
	for _, c := range config.Nodes {
		switch c.Key {
		case configClientIdKey:
			gc.ClientId = c.Value
		case configClientSecretKey:
			gc.ClientSecret = c.Value
		case configTenantIdKey:
			gc.TenantId = c.Value
		case configGcpSyncJobId:
			gc.GoogleSyncJobId = c.Value
		case configGcpSyncRuleId:
			gc.GoogleSyncRuleId = c.Value
		case configGcpProvisioningResourceId:
			id, err := uuid.Parse(c.Value)
			if err != nil {
				return nil, fmt.Errorf("parse provisioning resource id: %w", err)
			}
			gc.GoogleSyncProvisioningResourceId = id
		case configGcpSyncInterval:
			interval, err := time.ParseDuration(c.Value)
			if err != nil {
				return nil, fmt.Errorf("could not parse sync interval duration: %w", err)
			}
			gc.SyncInterval = interval
		default:
			return nil, fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if gc == r.Config {
		return r.entraId, nil
	}

	creds, err := azidentity.NewClientSecretCredential(gc.TenantId, gc.ClientId, gc.ClientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("create credentials: %w", err)
	}

	service, err := msgraphsdk.NewGraphServiceClientWithCredentials(creds, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, fmt.Errorf("create graph service client: %w", err)
	}

	r.entraId = service
	r.Config = gc

	return service, nil
}
func syncJobParameterSet(syncRuleId string, groupId string, userIds []string) *graphmodels.SynchronizationJobApplicationParameters {
	mutatedMembers := make([]models.SynchronizationJobSubjectable, 0, len(userIds))
	for _, u := range userIds {
		member := models.NewSynchronizationJobSubject()
		member.SetObjectId(&u)
		member.SetObjectTypeName(ptr.To("User"))
		mutatedMembers = append(mutatedMembers, member)
	}
	links := models.NewSynchronizationLinkedObjects()
	links.SetMembers(mutatedMembers)

	groupSubject := models.NewSynchronizationJobSubject()
	groupSubject.SetObjectId(&groupId)
	groupSubject.SetObjectTypeName(ptr.To("Group"))
	groupSubject.SetLinks(links)

	parameters := models.NewSynchronizationJobApplicationParameters()
	parameters.SetRuleId(&syncRuleId)
	parameters.SetSubjects([]models.SynchronizationJobSubjectable{groupSubject})

	return parameters
}

func (r *gcpSyncReconciler) Add(group string, member *string) {
	r.Queue.Add(SyncRequest{Group: group, User: member})
}

// These methods are no-ops, just there to satisfy the Reconciler interface.
// The GCP Syncer runs "independently"
func (r *gcpSyncReconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error {
	return nil
}

func (r *gcpSyncReconciler) Delete(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error {
	return nil
}
