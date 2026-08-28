package gcpsyncer

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"cloud.google.com/go/auth/credentials/idtoken"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	graphserviceprincipals "github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/queue"
)

const (
	reconcilerName = "entraid:gcpsync"

	configClientIdKey               = "clientId"
	configTenantIdKey               = "tenantId"
	configGcpProvisioningResourceId = "gcpProvisioningResourceId"
	configGcpSyncJobId              = "gcpSyncJobId"
	configGcpSyncRuleId             = "gcpSyncRuleId"
)

type gcpSyncConfig struct {
	ClientId                         string
	TenantId                         string
	GoogleSyncProvisioningResourceId uuid.UUID
	GoogleSyncJobId                  string
	GoogleSyncRuleId                 string
}

type gcpSyncReconciler struct {
	apiClient    *apiclient.APIClient
	config       gcpSyncConfig
	queue        queue.Queue[SyncRequest]
	queueChannel <-chan SyncRequest
	entraId      *msgraphsdk.GraphServiceClient
	log          logrus.FieldLogger
}

type SyncRequest struct {
	Group string
	Users []string
}

func New(client *apiclient.APIClient) *gcpSyncReconciler {
	queue, channel := queue.NewQueue[SyncRequest]()
	return &gcpSyncReconciler{
		apiClient:    client,
		queue:        queue,
		queueChannel: channel,
		log: logrus.New().
			WithField("subsystem", "gcpSyncer"),
	}
}

func (s *gcpSyncReconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        s.Name(),
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

func (s *gcpSyncReconciler) Name() string {
	return reconcilerName
}

func (s *gcpSyncReconciler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-s.queueChannel:
			if err := s.Sync(ctx, r); err != nil {
				s.log.Errorf("error running gcp sync: %s", err)
			}
		}
	}
}

func (s *gcpSyncReconciler) Sync(ctx context.Context, req SyncRequest) error {
	if err := s.parseConfig(ctx); err != nil {
		return err
	}
	if len(req.Users) == 0 {
		return s.syncGroupMembers(ctx, req.Group, nil)
	}
	var err error
	for usersChunk := range slices.Chunk(req.Users, 5) {
		err = errors.Join(err, s.syncGroupMembers(ctx, req.Group, usersChunk))
	}
	return err
}

func (s *gcpSyncReconciler) syncGroupMembers(ctx context.Context, group string, users []string) error {
	requestBody := graphserviceprincipals.NewItemSynchronizationJobsItemProvisionOnDemandPostRequestBody()

	var parameters []models.SynchronizationJobApplicationParametersable
	syncParamSet := syncJobParameterSet(s.config.GoogleSyncRuleId, group, users)
	parameters = append(parameters, syncParamSet)
	requestBody.SetParameters(parameters)

	_, err := s.entraId.ServicePrincipals().ByServicePrincipalId(s.config.GoogleSyncProvisioningResourceId.String()).
		Synchronization().Jobs().BySynchronizationJobId(s.config.GoogleSyncJobId).ProvisionOnDemand().
		Post(ctx, requestBody, nil)
	return err
}

func (s *gcpSyncReconciler) parseConfig(ctx context.Context) error {
	config, err := s.apiClient.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: s.Name(),
	})
	if err != nil {
		return fmt.Errorf("get reconciler config: %w", err)
	}
	gc := gcpSyncConfig{}
	for _, c := range config.Nodes {
		switch c.Key {
		case configClientIdKey:
			gc.ClientId = c.Value
		case configTenantIdKey:
			gc.TenantId = c.Value
		case configGcpSyncJobId:
			gc.GoogleSyncJobId = c.Value
		case configGcpSyncRuleId:
			gc.GoogleSyncRuleId = c.Value
		case configGcpProvisioningResourceId:
			id, err := uuid.Parse(c.Value)
			if err != nil {
				return fmt.Errorf("parse provisioning resource id: %w", err)
			}
			gc.GoogleSyncProvisioningResourceId = id
		default:
			return fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if gc == s.config {
		return nil
	}

	creds, err := azidentity.NewClientAssertionCredential(gc.TenantId, gc.ClientId, func(ctx context.Context) (string, error) {
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
		return fmt.Errorf("exchange for azure credentials: %w", err)
	}

	service, err := msgraphsdk.NewGraphServiceClientWithCredentials(creds, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return fmt.Errorf("create graph service client: %w", err)
	}

	s.entraId = service
	s.config = gc

	return nil
}

func syncJobParameterSet(syncRuleId string, groupId string, userIds []string) *models.SynchronizationJobApplicationParameters {
	links := models.NewSynchronizationLinkedObjects()

	if len(userIds) != 0 {
		mutatedMembers := make([]models.SynchronizationJobSubjectable, 0, len(userIds))
		for _, u := range userIds {
			member := models.NewSynchronizationJobSubject()
			member.SetObjectId(&u)
			member.SetObjectTypeName(new("User"))
			mutatedMembers = append(mutatedMembers, member)
		}
		links.SetMembers(mutatedMembers)
	}

	groupSubject := models.NewSynchronizationJobSubject()
	groupSubject.SetObjectId(&groupId)
	groupSubject.SetObjectTypeName(new("Group"))
	groupSubject.SetLinks(links)

	parameters := models.NewSynchronizationJobApplicationParameters()
	parameters.SetRuleId(&syncRuleId)
	parameters.SetSubjects([]models.SynchronizationJobSubjectable{groupSubject})

	return parameters
}

func (s *gcpSyncReconciler) Add(group string, members []string) error {
	return s.queue.Add(SyncRequest{Group: group, Users: members})
}

// These methods are no-ops, just there to satisfy the Reconciler interface.
// The GCP Syncer runs "independently"
func (s *gcpSyncReconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	return nil
}

func (s *gcpSyncReconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	return nil
}
