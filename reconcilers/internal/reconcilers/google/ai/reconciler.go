// Reconciles these resources for a Dapla Team:
// - Google's Vertex AI (Gemini Enterprise Agent Platform)
// - IAM permissions for Dapla teams
// - CloudBilling budget and notification channels
package ai

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"cloud.google.com/go/billing/budgets/apiv1"
	"cloud.google.com/go/billing/budgets/apiv1/budgetspb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"cloud.google.com/go/serviceusage/apiv1"
	"cloud.google.com/go/serviceusage/apiv1/serviceusagepb"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	iter "github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/iterator"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers"

	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/type/money"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	reconcilerName                 = "google:ai"
	aiPlatformUserRole             = "ssb.aiplatform.user"
	aiBudgetBillingAccount         = "billingAccounts/018A21-E69CB3-A95FA4"
	aiBudgetDisplayNameSuffix      = "AI budget"
	aiBudgetCurrencyCode           = "EUR"
	aiBudgetNotificationType       = "email"
	aiBudgetNotificationLabel      = "email_address"
	aiBudgetNotificationName       = "Vertex AI budget notification channel"
	aiBudgetThresholdHalf          = 0.5
	aiBudgetThresholdNinetyPercent = 0.9
	aiBudgetThresholdFull          = 1.0
	vertexAIServiceName            = "services/C7E2-9256-1C43"
	// For SAs in the test environment
	groupSANameTemplate = "developers@dapla-group-sa-t-57.iam.gserviceaccount.com"
)

type reconciler struct {
	BudgetNotificationEmails []string
}

type optFunc func(*reconciler)

func WithDaplaStatBudgetNotifications() optFunc {
	return func(r *reconciler) {
		r.BudgetNotificationEmails = []string{}
	}
}

func New(ctx context.Context, opts ...optFunc) (reconcilers.Reconciler, error) {
	r := new(reconciler)

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

// Configuration implements [reconcilers.Reconciler].
func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "Vertex AI reconciler",
		Description: "Enables Vertex AI (Gemini Enterprise Agent Platform) in Dapla Teams",
	}
}

// Name implements [reconcilers.Reconciler].
func (r *reconciler) Name() string {
	return reconcilerName
}

type googleServices struct {
	Project             *resourcemanager.ProjectsClient
	ServiceUsage        *serviceusage.Client
	CloudBudget         *budgets.BudgetClient
	NotificationChannel *monitoring.NotificationChannelClient
}

func (services googleServices) close() error {
	return errors.Join(
		services.Project.Close(),
		services.ServiceUsage.Close(),
		services.CloudBudget.Close(),
		services.NotificationChannel.Close(),
	)
}

func createGoogleClients(ctx context.Context) (*googleServices, error) {
	resourceManagerService, err1 := resourcemanager.NewProjectsClient(ctx)
	serviceUsageService, err2 := serviceusage.NewClient(ctx)
	budgetService, err3 := budgets.NewBudgetClient(ctx)
	ncService, err4 := monitoring.NewNotificationChannelClient(ctx)

	if err := cmp.Or(err1, err2, err3, err4); err != nil {
		return nil, err
	}

	return &googleServices{
		Project:             resourceManagerService,
		ServiceUsage:        serviceUsageService,
		CloudBudget:         budgetService,
		NotificationChannel: ncService,
	}, nil

}

// Reconcile implements [reconcilers.Reconciler].
func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	services, err := createGoogleClients(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := services.close(); err != nil {
			log.WithError(err).Error("close Google clients")
		}
	}()

	resp, err := client.GcpTeamResources().GetTeamFolder(ctx, &protoapi.GetGcpTeamFolderRequest{
		TeamSlug: daplaTeam.Slug,
		Env:      "test",
	})

	if err != nil {
		return err
	}

	teamTestFolder := resp.GetFolder()
	testProjectID, err := getProjectID(ctx, services.Project, teamTestFolder.FolderId, daplaTeam.Slug, teamTestFolder.Env)
	if err != nil {
		return err
	}

	aiFeature, err1 := client.Teams().HasFeature(ctx, &protoapi.HasFeatureRequest{
		Slug: daplaTeam.Slug,
		Feature: &protoapi.Feature{
			Name: "ai",
			Env:  "test",
		},
	})
	aiFeatureIsEnabled := aiFeature.HasFeature
	vertexAIEnabled, err2 := isVertexAIEnabled(ctx, services.ServiceUsage, testProjectID)
	membersHaveIAM, err3 := membersHaveAIPlatformUserBinding(ctx, services.Project, daplaTeam.Slug, testProjectID)
	budget, err4 := getExistingAIBudget(ctx, services.CloudBudget, fmt.Sprintf("%s %s", daplaTeam.Slug, aiBudgetDisplayNameSuffix))

	var budgetNotificationEmails []string
	notificationChannels, err5 := getExistingAIBudgetNotificationChannel(ctx, services.NotificationChannel, "projects/"+testProjectID, budgetNotificationEmails)

	if err := cmp.Or(err1, err2, err3, err4, err5); err != nil {
		return err
	}

	if vertexAIEnabled != aiFeatureIsEnabled {
		if err := reconcileVertexAIAPI(ctx, services.ServiceUsage, testProjectID, aiFeatureIsEnabled); err != nil {
			return err
		}
	}

	if membersHaveIAM != aiFeatureIsEnabled {
		if err := reconcileAIPlatformUserBinding(ctx, services.Project, daplaTeam.Slug, testProjectID, aiFeatureIsEnabled); err != nil {
			return err
		}
	}

	// Run the reconciler if the values of 'budgetExists', 'allNotificationChannelsExist' and 'aiFeatureIsEnabled' aren't all equal
	budgetExists := budget != nil
	allNotificationChannelsExist := len(notificationChannels) == len(budgetNotificationEmails)

	if !(budgetExists == aiFeatureIsEnabled && allNotificationChannelsExist == aiFeatureIsEnabled) {
		if err := reconcileAIBudget(ctx, client, services, daplaTeam.Slug, testProjectID, budget, r.BudgetNotificationEmails, aiFeatureIsEnabled); err != nil {
			return err
		}
	}

	return nil
}

// Get principals that should be have the AIPlatform user role
func getIAMMembers(daplaTeamSlug string) []string {
	// We assume the developers group will exist
	daplaDevelopersGroup := fmt.Sprintf("group:%s-developers@groups.ssb.no", daplaTeamSlug)
	// This SA is always guaranteed to exist as long as we run this reconciler **after**
	// the groupSA reconciler
	daplaDevelopersGroupSA := fmt.Sprintf("serviceAccount:%s-%s", daplaTeamSlug, groupSANameTemplate)
	return []string{daplaDevelopersGroup, daplaDevelopersGroupSA}
}

// Return true if all members are part of an IAM binding for the `aiPlatformUserRole`, otherwise return false
func membersHaveAIPlatformUserBinding(ctx context.Context, projectsClient *resourcemanager.ProjectsClient, daplaTeamSlug, projectID string) (bool, error) {
	policy, err := projectsClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: "projects/" + projectID})
	if err != nil {
		return false, fmt.Errorf("get IAM policy for project %q: %w", projectID, err)
	}

	members := getIAMMembers(daplaTeamSlug)

	bindingIndex := slices.IndexFunc(policy.Bindings, func(binding *iampb.Binding) bool {
		return binding.Role == aiPlatformUserRole
	})
	if bindingIndex == -1 {
		return false, nil
	}

	binding := policy.Bindings[bindingIndex]
	return !slices.ContainsFunc(members, func(member string) bool {
		return !slices.Contains(binding.Members, member)
	}), nil
}

// Ensures a dapla team's developers group and corresponding SA has the AI Platform user role on the project.
func reconcileAIPlatformUserBinding(ctx context.Context, projectsClient *resourcemanager.ProjectsClient, daplaTeamSlug, projectID string, enabled bool) error {

	members := getIAMMembers(daplaTeamSlug)

	projectName := "projects/" + projectID
	policy, err := projectsClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: projectName})
	if err != nil {
		return fmt.Errorf("get IAM policy for project %q: %w", projectID, err)
	}

	bindingIndex := slices.IndexFunc(policy.Bindings, func(binding *iampb.Binding) bool {
		return binding.Role == aiPlatformUserRole
	})

	switch {
	case enabled && bindingIndex != -1:
		// Binding exists: add any missing AI Platform user members.
		binding := policy.Bindings[bindingIndex]
		missingMembers := slices.DeleteFunc(slices.Clone(members), func(member string) bool {
			return slices.Contains(binding.Members, member)
		})

		if len(missingMembers) == 0 {
			return nil
		}

		binding.Members = append(binding.Members, missingMembers...)

	case enabled && bindingIndex == -1:
		// Binding is missing: create it with the required AI Platform user members.
		policy.Bindings = append(policy.Bindings, &iampb.Binding{
			Role:    aiPlatformUserRole,
			Members: members,
		})

	case !enabled && bindingIndex != -1:
		// Binding exists: remove this team's AI Platform user members.
		binding := policy.Bindings[bindingIndex]
		binding.Members = slices.DeleteFunc(binding.Members, func(member string) bool {
			return slices.Contains(members, member)
		})

		if len(binding.Members) == 0 {
			policy.Bindings = slices.Delete(policy.Bindings, bindingIndex, bindingIndex+1)
		}

	case !enabled && bindingIndex == -1:
		// Binding is already missing: there is nothing to remove.
		return nil
	}

	_, err = projectsClient.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{Resource: projectName, Policy: policy})
	if err != nil {
		return fmt.Errorf("set IAM policy for project %q: %w", projectID, err)
	}
	return nil
}

func reconcileAIBudget(ctx context.Context, client *apiclient.APIClient, services *googleServices, daplaTeamSlug, projectID string, existingBudget *budgetspb.Budget, daplaStatNotificationEmails []string, enabled bool) error {
	if !enabled {
		if existingBudget != nil {
			if err := services.CloudBudget.DeleteBudget(ctx, &budgetspb.DeleteBudgetRequest{
				Name: existingBudget.GetName(),
			}); err != nil {
				return fmt.Errorf("delete AI budget %q: %w", existingBudget.Name, err)
			}
		}

		_, err := reconcileAIBudgetNotificationChannels(ctx, services.NotificationChannel, projectID, nil, enabled)
		return err
	}

	budgetNotificationUsers, err := getGroupMembers(ctx, client, fmt.Sprintf("%s-developers", daplaTeamSlug), 1)
	if err != nil {
		return err
	}

	budgetNotificationEmail := make([]string, len(budgetNotificationUsers))

	for idx, user := range budgetNotificationUsers {
		budgetNotificationEmail[idx] = user.User.Email
	}

	budgetNotificationEmails := slices.Concat(budgetNotificationEmail, daplaStatNotificationEmails)

	project, err := services.Project.GetProject(ctx, &resourcemanagerpb.GetProjectRequest{Name: "projects/" + projectID})
	if err != nil {
		return fmt.Errorf("get project %q: %w", projectID, err)
	}

	notificationChannelNames, err := reconcileAIBudgetNotificationChannels(ctx, services.NotificationChannel, projectID, budgetNotificationEmails, enabled)
	if err != nil {
		return err
	}

	budget := getAIBudget(daplaTeamSlug, strings.TrimPrefix(project.Name, "projects/"), 0, notificationChannelNames)
	if existingBudget == nil {
		_, err = services.CloudBudget.CreateBudget(ctx, &budgetspb.CreateBudgetRequest{Parent: aiBudgetBillingAccount, Budget: budget})
		if err != nil {
			return fmt.Errorf("create AI budget for project %q: %w", projectID, err)
		}
		return nil
	}

	budget.Name = existingBudget.Name
	budget.Etag = existingBudget.Etag
	_, err = services.CloudBudget.UpdateBudget(ctx, &budgetspb.UpdateBudgetRequest{
		Budget: budget,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"display_name", "budget_filter", "amount", "threshold_rules", "notifications_rule"},
		},
	})
	if err != nil {
		return fmt.Errorf("update AI budget %q: %w", existingBudget.Name, err)
	}
	return nil
}

func reconcileAIBudgetNotificationChannels(ctx context.Context, ncClient *monitoring.NotificationChannelClient, projectID string, budgetNotificationEmails []string, enabled bool) ([]string, error) {
	projectName := "projects/" + projectID
	if !enabled {
		filter := fmt.Sprintf(`display_name = "%s" AND type = "%s"`, aiBudgetNotificationName, aiBudgetNotificationType)
		var channelNames []string
		it := ncClient.ListNotificationChannels(ctx, &monitoringpb.ListNotificationChannelsRequest{Name: projectName, Filter: filter})
		for channel, err := range it.All() {
			if err != nil {
				return nil, fmt.Errorf("list AI budget notification channels for %q: %w", projectName, err)
			}
			channelNames = append(channelNames, channel.Name)
		}

		for _, channelName := range channelNames {
			if err := ncClient.DeleteNotificationChannel(ctx, &monitoringpb.DeleteNotificationChannelRequest{Name: channelName}); err != nil {
				return nil, fmt.Errorf("delete AI budget notification channel %q: %w", channelName, err)
			}
		}
		return nil, nil
	}

	for _, budgetNotificationEmail := range budgetNotificationEmails {
		filter := fmt.Sprintf(`type = "%s" AND labels.%s = "%s"`, aiBudgetNotificationType, aiBudgetNotificationLabel, budgetNotificationEmail)
		_, err := ncClient.ListNotificationChannels(ctx, &monitoringpb.ListNotificationChannelsRequest{Name: projectName, Filter: filter}).Next()
		if err != nil && err != iterator.Done {
			return nil, fmt.Errorf("list AI budget notification channels for %q: %w", projectName, err)
		}

		if err == iterator.Done {
			_, err = ncClient.CreateNotificationChannel(ctx, &monitoringpb.CreateNotificationChannelRequest{
				Name: projectName,
				NotificationChannel: &monitoringpb.NotificationChannel{
					DisplayName: aiBudgetNotificationName,
					Type:        aiBudgetNotificationType,
					Labels: map[string]string{
						aiBudgetNotificationLabel: budgetNotificationEmail,
					},
				},
			})
			if err != nil {
				return nil, fmt.Errorf("create AI budget notification channel for %q in project %q: %w", budgetNotificationEmail, projectID, err)
			}
		}

	}

	channels, err := getExistingAIBudgetNotificationChannel(ctx, ncClient, projectName, budgetNotificationEmails)
	if err != nil {
		return nil, err
	}

	channelNames := make([]string, len(channels))
	for i, channel := range channels {
		channelNames[i] = channel.Name
	}
	return channelNames, nil
}

// Get existing notification channels, if any budgetNotificationEmail is missing a channel return an error
func getExistingAIBudgetNotificationChannel(ctx context.Context, ncClient *monitoring.NotificationChannelClient, projectName string, emails []string) ([]*monitoringpb.NotificationChannel, error) {
	channels := make([]*monitoringpb.NotificationChannel, 0, len(emails))
	for _, email := range emails {
		filter := fmt.Sprintf(`type = "%s" AND labels.%s = "%s"`, aiBudgetNotificationType, aiBudgetNotificationLabel, email)
		channel, err := ncClient.ListNotificationChannels(ctx, &monitoringpb.ListNotificationChannelsRequest{Name: projectName, Filter: filter}).Next()
		if err == iterator.Done {
			return nil, fmt.Errorf("AI budget notification channel for %q does not exist in %q", email, projectName)
		}
		if err != nil {
			return nil, fmt.Errorf("list AI budget notification channels for %q: %w", projectName, err)
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func getExistingAIBudget(ctx context.Context, budgetClient *budgets.BudgetClient, displayName string) (*budgetspb.Budget, error) {
	it := budgetClient.ListBudgets(ctx, &budgetspb.ListBudgetsRequest{Parent: aiBudgetBillingAccount})
	for budget, err := range it.All() {
		if err != nil {
			return nil, fmt.Errorf("list AI budgets: %w", err)
		}
		if budget.DisplayName == displayName {
			return budget, nil
		}
	}
	return nil, nil
}

func getAIBudget(daplaTeamSlug, projectNumber string, budgetNotificationLimitUnits int64, notificationChannelNames []string) *budgetspb.Budget {
	return &budgetspb.Budget{
		DisplayName: fmt.Sprintf("%s %s", daplaTeamSlug, aiBudgetDisplayNameSuffix),
		BudgetFilter: &budgetspb.Filter{
			Projects:             []string{"projects/" + projectNumber},
			CreditTypesTreatment: budgetspb.Filter_INCLUDE_ALL_CREDITS,
			Services:             []string{vertexAIServiceName},
			UsagePeriod: &budgetspb.Filter_CalendarPeriod{
				CalendarPeriod: budgetspb.CalendarPeriod_MONTH,
			},
		},
		Amount: &budgetspb.BudgetAmount{
			BudgetAmount: &budgetspb.BudgetAmount_SpecifiedAmount{
				SpecifiedAmount: &money.Money{
					CurrencyCode: aiBudgetCurrencyCode,
					Units:        budgetNotificationLimitUnits,
				},
			},
		},
		ThresholdRules: []*budgetspb.ThresholdRule{
			{ThresholdPercent: aiBudgetThresholdHalf},
			{ThresholdPercent: aiBudgetThresholdNinetyPercent},
			{ThresholdPercent: aiBudgetThresholdFull},
		},
		NotificationsRule: &budgetspb.NotificationsRule{
			MonitoringNotificationChannels: notificationChannelNames,
			DisableDefaultIamRecipients:    true,
		},
	}
}

func getGroupMembers(ctx context.Context, client *apiclient.APIClient, group string, limit uint) ([]*protoapi.GroupMember, error) {
	dbMembersIt := iter.New(ctx, int64(limit), func(limit, offset int64) (*protoapi.ListGroupMembersResponse, error) {
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

// Check if the vertexai API is enabled in a given Google project
func isVertexAIEnabled(ctx context.Context, client *serviceusage.Client, projectID string) (bool, error) {
	name := fmt.Sprintf("projects/%s/services/aiplatform.googleapis.com", projectID)
	service, err := client.GetService(ctx, &serviceusagepb.GetServiceRequest{Name: name})
	if err != nil {
		return false, err
	}
	return service.State == serviceusagepb.State_ENABLED, nil
}

// Enable or disable the vertexAI API in a given google project
func reconcileVertexAIAPI(ctx context.Context, client *serviceusage.Client, projectID string, enabled bool) error {
	name := fmt.Sprintf("projects/%s/services/aiplatform.googleapis.com", projectID)
	if enabled {
		op, err := client.EnableService(ctx, &serviceusagepb.EnableServiceRequest{Name: name})
		if err != nil {
			return err
		}
		if _, err := op.Wait(ctx); err != nil {
			return fmt.Errorf("enable Vertex AI API for project %q: %w", projectID, err)
		}
		return nil
	}

	op, err := client.DisableService(ctx, &serviceusagepb.DisableServiceRequest{Name: name})
	if err != nil {
		return err
	}
	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("disable Vertex AI API for project %q: %w", projectID, err)
	}
	return nil
}

func getProjectID(ctx context.Context, client *resourcemanager.ProjectsClient, folderID, daplaTeamSlug, env string) (string, error) {
	it := client.SearchProjects(ctx, &resourcemanagerpb.SearchProjectsRequest{
		Query: fmt.Sprintf("parent:folders/%s", folderID),
	})

	projectID := ""
	projectIDPrefix := daplaTeamSlug + "-" + string([]rune("Hello")[0])
	for project, err := range it.All() {
		if err != nil {
			return "", err
		}

		if strings.HasPrefix(project.ProjectId, projectIDPrefix) {
			projectID = project.ProjectId
		}
	}

	if projectID == "" {
		return "", fmt.Errorf("no standard project found in folder %q for environment %q", folderID, env)
	}
	return projectID, nil
}

// Delete implements [reconcilers.Reconciler].
func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	panic("unimplemented")
}
