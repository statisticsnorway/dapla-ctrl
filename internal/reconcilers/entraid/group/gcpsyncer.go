package group

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphserviceprincipals "github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/sirupsen/logrus"
	"k8s.io/utils/ptr"
)

const (
	gcpSyncInterval  = 2 * time.Minute
	gcpSyncQueueSize = 1024
)

type gcpSyncConfig struct {
	GoogleSyncAppRoleId              uuid.UUID
	GoogleSyncProvisioningResourceId uuid.UUID
	GoogleSyncSSOResourceId          uuid.UUID
	GoogleSyncJobId                  string
	GoogleSyncRuleId                 string
}

func (c gcpSyncConfig) Equal(other gcpSyncConfig) bool {
	return c.GoogleSyncAppRoleId == other.GoogleSyncAppRoleId &&
		c.GoogleSyncProvisioningResourceId == other.GoogleSyncProvisioningResourceId &&
		c.GoogleSyncSSOResourceId == other.GoogleSyncSSOResourceId &&
		c.GoogleSyncJobId == other.GoogleSyncJobId &&
		c.GoogleSyncRuleId == other.GoogleSyncRuleId
}

type gcpSyncer struct {
	Config  gcpSyncConfig
	queue   chan gcpSyncRequest
	EntraId *msgraphsdkgo.GraphServiceClient
	log     logrus.FieldLogger
}

func NewGcpSyncer() gcpSyncer {
	return gcpSyncer{
		queue: make(chan gcpSyncRequest, gcpSyncQueueSize),
		log: logrus.New().
			WithField("reconciler", reconcilerName).
			WithField("subsystem", "gcpSyncer"),
	}
}

type gcpSyncRequest struct {
	Created time.Time
	Group   string
	User    string
}

func (s *gcpSyncer) Queue(group string, users []string) {
	for _, user := range users {
		s.queue <- gcpSyncRequest{
			Created: time.Now(),
			Group:   group,
			User:    user,
		}
	}
}

func (s *gcpSyncer) Run(ctx context.Context) {
	for {
		if err := s.run(ctx); err != nil {
			s.log.Errorf("error while performing gcp sync: %w", err)
		}
	}
}

func (s *gcpSyncer) run(ctx context.Context) error {
	syncTimer := &time.Timer{}

	groupsToSync := make(map[string][]string)

outerLoop:
	for {
		select {
		case r := <-s.queue:
			if syncTimer.C == nil {
				s.log.Info("received first sync request, starting collection interval", "interval", gcpSyncInterval)
				syncTimer = time.NewTimer(gcpSyncInterval)
			}
			if !slices.Contains(groupsToSync[r.Group], r.User) {
				groupsToSync[r.Group] = append(groupsToSync[r.Group], r.User)
			}
			// Remove duplicates, stop if over 4 users in group (5 is limit for sync job)
			if len(groupsToSync[r.Group]) > 4 {
				s.log.Info("group has 5 or more member updates, splitting..")
				break outerLoop
			}

		case <-ctx.Done():
			return nil
		case <-syncTimer.C:
			break outerLoop
		}
	}

	if len(groupsToSync) == 0 {
		s.log.Info("no groups to sync")
		return nil
	}

	requestBody := graphserviceprincipals.NewItemSynchronizationJobsItemProvisionOnDemandPostRequestBody()

	var parameters []models.SynchronizationJobApplicationParametersable

	for group, users := range groupsToSync {
		syncParamSet := syncJobParameterSet(s.Config.GoogleSyncRuleId, group, users)
		parameters = append(parameters, syncParamSet)
	}

	s.log.Info("syncing these groups and users", groupsToSync)

	requestBody.SetParameters(parameters)

	_, err := s.EntraId.ServicePrincipals().ByServicePrincipalId(s.Config.GoogleSyncProvisioningResourceId.String()).
		Synchronization().Jobs().BySynchronizationJobId(s.Config.GoogleSyncJobId).ProvisionOnDemand().
		Post(ctx, requestBody, nil)

	return err
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
