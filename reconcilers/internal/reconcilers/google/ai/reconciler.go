package ai

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
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
	reconcilerName            = "google:ai"
	aiBudgetDisplayNameSuffix = "AI budget"
	aiBudgetCurrencyCode      = "EUR"
	aiBudgetNotificationType  = "email"
	aiBudgetNotificationLabel = "email_address"
	vertexAIServiceName       = "services/C7E2-9256-1C43"
	environment               = "test"

	aiPlatformUserRoleKey              = "ai_platform_user_role"
	aiBudgetThresholdsKey              = "ai_budget_thresholds"
	aiBudgetBillingAccountKey          = "ai_budget_billing_account"
	aiBudgetNotificationChannelNameKey = "ai_budget_notification_channel_name"
	groupSANameTemplateKey             = "group_sa_name_template"
)

type reconciler struct {
	AIPlatformUserRole            string
	GroupSANameTemplate           string
	AIBudgetBillingAccount        string
	AIBudgetThresholds            []float64
	AIBudgetNotificationName      string
	BudgetNotificationEmails      []string
	AIBudgetDeveloperBillingGroup string
}

type optFunc func(*reconciler) error

func WithDaplaStatBudgetNotifications(ctx context.Context, apiclient *apiclient.APIClient) optFunc {
	return func(r *reconciler) error {
		var limit uint = 4
		members, err := getGroupMembers(ctx, apiclient, r.AIBudgetDeveloperBillingGroup, limit)
		if err != nil {
			return err
		}

		developerEmails := make([]string, limit)

		for i, member := range members {
			developerEmails[i] = member.User.Email
		}

		r.BudgetNotificationEmails = developerEmails

		return nil
	}
}

func New(ctx context.Context, opts ...optFunc) (reconcilers.Reconciler, error) {
	r := new(reconciler)

	r.AIBudgetThresholds = []float64{0.5, 0.9, 1}
	r.AIPlatformUserRole = "ssb.aiplatform.user"
	r.AIBudgetBillingAccount = "billingAccounts/018A21-E69CB3-A95FA4"

	r.AIBudgetNotificationName = "Vertex AI budget notification channel"
	r.AIBudgetDeveloperBillingGroup = "group:dapla-stat-developers@groups.ssb.no"
	// For SAs in the test environment
	r.GroupSANameTemplate = "developers@dapla-group-sa-t-57.iam.gserviceaccount.com"

	for _, opt := range opts {
		err := opt(r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

// Configuration implements [reconcilers.Reconciler].
func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "Vertex AI reconciler",
		Description: "Enables Vertex AI (Gemini Enterprise Agent Platform) in Dapla Teams",
		Config: []*protoapi.ReconcilerConfigSpec{
			&protoapi.ReconcilerConfigSpec{
				Key:         aiBudgetThresholdsKey,
				DisplayName: "AI Budget Notification Thresholds",
				Description: "The threshhold values for when to notify users about exceeded budget limits.",
				Secret:      false,
			},
			&protoapi.ReconcilerConfigSpec{
				Key:         aiPlatformUserRoleKey,
				DisplayName: "AI Platform User Role",
				Description: "The name of the custom user role for Vertex AI.",
				Secret:      false,
			},
			&protoapi.ReconcilerConfigSpec{
				Key:         aiBudgetBillingAccountKey,
				DisplayName: "AI Budget Billing Account",
				Description: "The ID of the SSB org billing account.",
				Secret:      false,
			},
			&protoapi.ReconcilerConfigSpec{
				Key:         aiBudgetNotificationChannelNameKey,
				DisplayName: "AI Budget Notification Channel Name",
				Description: "The name of the budget notificaiton channel resource.",
				Secret:      false,
			},

			&protoapi.ReconcilerConfigSpec{
				Key:         groupSANameTemplateKey,
				DisplayName: "Group Service Account Name Template",
				Description: "Template for the name of the Group Service Account.",
				Secret:      false,
			},
		},
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
	resourceManagerService, err := resourcemanager.NewProjectsClient(ctx)
	if err != nil {
		return nil, err
	}
	serviceUsageService, err := serviceusage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	budgetService, err := budgets.NewBudgetClient(ctx)
	if err != nil {
		return nil, err
	}
	ncService, err := monitoring.NewNotificationChannelClient(ctx)
	if err != nil {
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
	if err := r.updateConfig(ctx, client); err != nil {
		return fmt.Errorf("error getting reconciler config: %w", err)
	}

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
		Env:      environment,
	})
	if err != nil {
		return err
	}

	teamFolder := resp.GetFolder()
	projectID, err := getStandardProjectID(ctx, services.Project, teamFolder.FolderId, daplaTeam.Slug, teamFolder.Env)
	if err != nil {
		return err
	}

	aiFeature, err := client.Teams().HasFeature(ctx, &protoapi.HasFeatureRequest{
		Slug: daplaTeam.Slug,
		Feature: &protoapi.Feature{
			Name: "ai",
			Env:  environment,
		},
	})
	if err != nil {
		return err
	}
	aiFeatureIsEnabled := aiFeature.HasFeature
	vertexAIEnabled, err := isVertexAIEnabled(ctx, services.ServiceUsage, projectID)
	if err != nil {
		return err
	}
	membersHaveIAM, err := membersHaveAIPlatformUserBinding(r, ctx, services.Project, daplaTeam.Slug, projectID)
	if err != nil {
		return err
	}
	budget, err := getExistingAIBudget(r, ctx, services.CloudBudget, fmt.Sprintf("%s %s", daplaTeam.Slug, aiBudgetDisplayNameSuffix))
	if err != nil {
		return err
	}
	var budgetNotificationEmails []string
	notificationChannels, err := getExistingAIBudgetNotificationChannel(ctx, services.NotificationChannel, "projects/"+projectID, budgetNotificationEmails)
	if err != nil {
		return err
	}

	if vertexAIEnabled != aiFeatureIsEnabled {
		if err := reconcileVertexAIAPI(ctx, services.ServiceUsage, projectID, aiFeatureIsEnabled); err != nil {
			return err
		}
	}

	if membersHaveIAM != aiFeatureIsEnabled {
		if err := reconcileAIPlatformUserBinding(r, ctx, services.Project, daplaTeam.Slug, projectID, aiFeatureIsEnabled); err != nil {
			return err
		}
	}

	budgetExists := budget != nil
	allNotificationChannelsExist := len(notificationChannels) == len(budgetNotificationEmails)
	// Run the reconciler if 'budgetExists', 'allNotificationChannelsExist' and 'aiFeatureIsEnabled' aren't all equal
	needToBeReconciled := !(budgetExists == aiFeatureIsEnabled && allNotificationChannelsExist == aiFeatureIsEnabled)
	if needToBeReconciled {
		if err := reconcileAIBudget(r, ctx, client, services, daplaTeam.Slug, projectID, budget, r.BudgetNotificationEmails, aiFeatureIsEnabled); err != nil {
			return err
		}
	}

	return nil
}

// Get principals that should be have the AIPlatform user role
func getIAMMembers(r *reconciler, daplaTeamSlug string) []string {
	// We assume the developers group will exist
	daplaDevelopersGroup := fmt.Sprintf("group:%s-developers@groups.ssb.no", daplaTeamSlug)
	// This SA is always guaranteed to exist as long as we run this reconciler **after**
	// the groupSA reconciler
	daplaDevelopersGroupSA := fmt.Sprintf("serviceAccount:%s-%s", daplaTeamSlug, r.GroupSANameTemplate)
	return []string{daplaDevelopersGroup, daplaDevelopersGroupSA}
}

// Return true if all members are part of an IAM binding for the `AIPlatformUserRole`, otherwise return false
func membersHaveAIPlatformUserBinding(r *reconciler, ctx context.Context, projectsClient *resourcemanager.ProjectsClient, daplaTeamSlug, projectID string) (bool, error) {
	policy, err := projectsClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: "projects/" + projectID})
	if err != nil {
		return false, fmt.Errorf("get IAM policy for project %q: %w", projectID, err)
	}

	members := getIAMMembers(r, daplaTeamSlug)

	bindingIndex := slices.IndexFunc(policy.Bindings, func(binding *iampb.Binding) bool {
		return binding.Role == r.AIPlatformUserRole
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
func reconcileAIPlatformUserBinding(r *reconciler, ctx context.Context, projectsClient *resourcemanager.ProjectsClient, daplaTeamSlug, projectID string, enabled bool) error {
	members := getIAMMembers(r, daplaTeamSlug)

	projectName := "projects/" + projectID
	policy, err := projectsClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: projectName})
	if err != nil {
		return fmt.Errorf("get IAM policy for project %q: %w", projectID, err)
	}

	bindingIndex := slices.IndexFunc(policy.Bindings, func(binding *iampb.Binding) bool {
		return binding.Role == r.AIPlatformUserRole
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
			Role:    r.AIPlatformUserRole,
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

func (r *reconciler) updateConfig(ctx context.Context, client *apiclient.APIClient) error {
	config, err := client.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return fmt.Errorf("get reconciler config: %w", err)
	}

	for _, c := range config.Nodes {
		switch c.Key {
		case aiPlatformUserRoleKey:
			r.AIPlatformUserRole = c.Value
		case aiBudgetThresholdsKey:
			parts := strings.Split(c.Value, ",")
			values := make([]float64, len(parts))

			for i, part := range parts {
				var err error
				values[i], err = strconv.ParseFloat(strings.TrimSpace(part), 64)
				if err != nil {
					return fmt.Errorf("invalid float %q: %w", part, err)
				}
			}
			r.AIBudgetThresholds = values
		case aiBudgetNotificationChannelNameKey:
			r.AIBudgetNotificationName = c.Value
		case aiBudgetBillingAccountKey:
			r.AIBudgetBillingAccount = c.Value
		case groupSANameTemplateKey:
			r.GroupSANameTemplate = c.Value
		default:
			return fmt.Errorf("unknown config key %q", c.Key)

		}
	}

	return nil

}

func reconcileAIBudget(r *reconciler, ctx context.Context, client *apiclient.APIClient, services *googleServices, daplaTeamSlug, projectID string, existingBudget *budgetspb.Budget, daplaStatNotificationEmails []string, enabled bool) error {
	if !enabled {
		if existingBudget != nil {
			if err := services.CloudBudget.DeleteBudget(ctx, &budgetspb.DeleteBudgetRequest{
				Name: existingBudget.GetName(),
			}); err != nil {
				return fmt.Errorf("delete AI budget %q: %w", existingBudget.Name, err)
			}
		}

		_, err := reconcileAIBudgetNotificationChannels(ctx, r, services.NotificationChannel, projectID, nil)
		return err
	}

	teamDevelopers, err := getGroupMembers(ctx, client, fmt.Sprintf("%s-developers", daplaTeamSlug), 100)
	if err != nil {
		return err
	}

	// The Google Billing API only allows 5 notification channels to be attached to a billing budget. Therefore we only pick one developer from the team + 4 dapla-stat developers to recieve billing alerts
	budgetNotificationEmails := slices.Concat([]string{teamDevelopers[0].User.Email}, daplaStatNotificationEmails)

	project, err := services.Project.GetProject(ctx, &resourcemanagerpb.GetProjectRequest{Name: "projects/" + projectID})
	if err != nil {
		return fmt.Errorf("get project %q: %w", projectID, err)
	}

	notificationChannelNames, err := reconcileAIBudgetNotificationChannels(ctx, r, services.NotificationChannel, projectID, budgetNotificationEmails)
	if err != nil {
		return err
	}

	// (developers * 10) EUR per month
	monthlyBudgetLimit := int64(len(teamDevelopers)) * 10

	budget := getAIBudget(r, daplaTeamSlug, strings.TrimPrefix(project.Name, "projects/"), monthlyBudgetLimit, notificationChannelNames)
	if existingBudget == nil {
		_, err = services.CloudBudget.CreateBudget(ctx, &budgetspb.CreateBudgetRequest{Parent: r.AIBudgetBillingAccount, Budget: budget})
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

func reconcileAIBudgetNotificationChannels(ctx context.Context, r *reconciler, ncClient *monitoring.NotificationChannelClient, projectID string, budgetNotificationEmails []string) ([]string, error) {
	projectName := "projects/" + projectID
	filter := fmt.Sprintf(`display_name = "%s" AND type = "%s"`, r.AIBudgetNotificationName, aiBudgetNotificationType)
	var channels []*monitoringpb.NotificationChannel
	it := ncClient.ListNotificationChannels(ctx, &monitoringpb.ListNotificationChannelsRequest{Name: projectName, Filter: filter})
	for channel, err := range it.All() {
		if err != nil {
			return nil, fmt.Errorf("list AI budget notification channels for %q: %w", projectName, err)
		}
		channels = append(channels, channel)
	}

	// We want to disable channels which exist in Google, but not locally
	toDisable := slices.DeleteFunc(slices.Clone(channels), func(c *monitoringpb.NotificationChannel) bool {
		return slices.Contains(budgetNotificationEmails, c.Labels[aiBudgetNotificationLabel])
	})

	// We want to enable channels which don't exist in Google, but do exist locally
	toEnable := slices.DeleteFunc(slices.Clone(budgetNotificationEmails), func(s string) bool {
		return slices.ContainsFunc(channels, func(c *monitoringpb.NotificationChannel) bool {
			return c.Labels[aiBudgetNotificationLabel] == s
		})
	})

	for _, channel := range toDisable {
		if err := ncClient.DeleteNotificationChannel(ctx, &monitoringpb.DeleteNotificationChannelRequest{Name: channel.Name}); err != nil {
			return nil, fmt.Errorf("delete AI budget notification channel %q: %w", channel.Name, err)
		}
	}
	channels = slices.DeleteFunc(channels, func(old *monitoringpb.NotificationChannel) bool {
		return slices.ContainsFunc(toDisable, func(c *monitoringpb.NotificationChannel) bool {
			return old.Name == c.Name
		})
	})

	for _, email := range toEnable {
		channel, err := ncClient.CreateNotificationChannel(ctx, &monitoringpb.CreateNotificationChannelRequest{
			Name: projectName,
			NotificationChannel: &monitoringpb.NotificationChannel{
				DisplayName: r.AIBudgetNotificationName,
				Type:        aiBudgetNotificationType,
				Labels: map[string]string{
					aiBudgetNotificationLabel: email,
				},
			}})
		if err != nil {
			return nil, fmt.Errorf("create AI budget notification channel for %q in project %q: %w", email, projectID, err)
		}

		channels = append(channels, channel)

	}

	channelNames := make([]string, 0, len(channels))
	for _, channel := range channels {
		channelNames = append(channelNames, channel.Name)
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

func getExistingAIBudget(r *reconciler, ctx context.Context, budgetClient *budgets.BudgetClient, displayName string) (*budgetspb.Budget, error) {
	it := budgetClient.ListBudgets(ctx, &budgetspb.ListBudgetsRequest{Parent: r.AIBudgetBillingAccount})
	for budget, err := range it.All() {
		if err != nil {
			return nil, fmt.Errorf("list AI budgets: %w", err)
		}
		if budget.DisplayName == displayName {
			return budget, nil
		}
	}
	return nil, fmt.Errorf("list AI budgets: none matching name %s", displayName)
}

func getAIBudget(r *reconciler, daplaTeamSlug, projectNumber string, budgetNotificationLimitUnits int64, notificationChannelNames []string) *budgetspb.Budget {
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
			{ThresholdPercent: r.AIBudgetThresholds[0]},
			{ThresholdPercent: r.AIBudgetThresholds[1]},
			{ThresholdPercent: r.AIBudgetThresholds[2]},
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

// Get the project ID of the standard project given a folderID, team slug and environment
func getStandardProjectID(ctx context.Context, client *resourcemanager.ProjectsClient, folderID, daplaTeamSlug, env string) (string, error) {
	it := client.SearchProjects(ctx, &resourcemanagerpb.SearchProjectsRequest{
		Query: fmt.Sprintf("parent:folders/%s", folderID),
	})

	projectID := ""
	projectIDPrefix := fmt.Sprintf("%s-%s-", daplaTeamSlug, string([]rune(env)[0]))
	for project, err := range it.All() {
		if err != nil {
			return "", err
		}

		if strings.HasPrefix(project.ProjectId, projectIDPrefix) {
			projectID = project.ProjectId
			break
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
