// Reconciler which enables Google's Vertex AI (Gemini Enterprise Agent Platform)
// and sets the right IAM permissions for Dapla teams.
package ai

import (
	"context"
	"fmt"
	"slices"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/serviceusage/v1"
)

const (
	reconcilerName     = "google:ai"
	aiPlatformUserRole = "ssb.aiplatform.user"
	// For SAs in the test environment
	groupSANameTemplate = "developers@dapla-group-sa-t-57.iam.gserviceaccount.com"
)

type reconciler struct{}

func New() *reconciler {
	return &reconciler{}
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

// Reconcile implements [reconcilers.Reconciler].
func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	resourceManagerService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return err
	}

	serviceUsageService, err := serviceusage.NewService(ctx)
	if err != nil {
		return err
	}

	testProjectID, err := projectExists(resourceManagerService, daplaTeam.Slug)

	if err != nil {
		return err
	}

	if testProjectID == nil {
		return fmt.Errorf("No test project found for team %s", daplaTeam.Slug)
	}

	aiFeatureIsEnabled, err := isAIFeatureEnabled(ctx, client, daplaTeam.Slug)

	if err != nil {
		return err
	}

	vertexAIEnabled, err := isVertexAIEnabled(serviceUsageService, *testProjectID)

	if err != nil {
		return err
	}

	// TODO: Set or remove IAM permissions:
	// - Give ssb.aiplatform.user to dapla SA
	if aiFeatureIsEnabled && !vertexAIEnabled {
		if err := setVertexAIEnabled(serviceUsageService, *testProjectID, true); err != nil {
			return err
		}

		if err := reconcileAIPlatformUserBinding(ctx, resourceManagerService, client, daplaTeam.Slug, *testProjectID); err != nil {
			return err
		}

	} else if !aiFeatureIsEnabled && vertexAIEnabled {
		if err := setVertexAIEnabled(serviceUsageService, *testProjectID, false); err != nil {
			return err
		}
	}

	return nil
}

// Get the Dapla teams' developers group name if it exists
func getDevelopersGroup(ctx context.Context, client *apiclient.APIClient, daplaTeamSlug string) (*string, error) {
	teamDevelopersGroup, err := client.Groups().Get(ctx, &protoapi.GetGroupRequest{
		Name: daplaTeamSlug + "-developers",
	})

	if err != nil {
		return nil, err
	}

	return &teamDevelopersGroup.Group.Name, nil
}

// Ensures a dapla team's developers group and corresponding SA has the AI Platform user role on the project.
func reconcileAIPlatformUserBinding(ctx context.Context, svc *cloudresourcemanager.Service, client *apiclient.APIClient, daplaTeamSlug, projectID string) error {
	// TODO: Perhaps we can assume these standard dapla groups will always exist
	// so we don't need to query google cloud?
	daplaDevelopersGroup_, err := getDevelopersGroup(ctx, client, daplaTeamSlug)
	if err != nil {
		return fmt.Errorf("get developers group: %w", err)
	}

	daplaDevelopersGroup := fmt.Sprintf("group:%s", *daplaDevelopersGroup_)
	// This SA is always guaranteed to exist as long as we run this reconciler **after**
	// the groupSA reconciler
	daplaDevelopersGroupSA := fmt.Sprintf("serviceAccount:%s-%s", daplaTeamSlug, groupSANameTemplate)
	members := []string{daplaDevelopersGroup, daplaDevelopersGroupSA}
	policy, err := svc.Projects.GetIamPolicy(projectID, &cloudresourcemanager.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("get IAM policy for project %q: %w", projectID, err)
	}

	bindingIndex := slices.IndexFunc(policy.Bindings, func(binding *cloudresourcemanager.Binding) bool {
		return binding.Role == aiPlatformUserRole
	})

	if bindingIndex != -1 {
		binding := policy.Bindings[bindingIndex]
		missingMembers := slices.DeleteFunc(slices.Clone(members), func(member string) bool {
			return slices.Contains(binding.Members, member)
		})

		if len(missingMembers) == 0 {
			return nil
		}

		binding.Members = append(binding.Members, missingMembers...)
		_, err = svc.Projects.SetIamPolicy(projectID, &cloudresourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("set IAM policy for project %q: %w", projectID, err)
		}
		return nil
	} else {
		// If the binding doesn't exist, create it.
		policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
			Role:    aiPlatformUserRole,
			Members: members,
		})

		_, err = svc.Projects.SetIamPolicy(projectID, &cloudresourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("set IAM policy for project %q: %w", projectID, err)
		}
		return nil
	}
}

// Check if the AI feature is enabled in the API database for a given team.
func isAIFeatureEnabled(ctx context.Context, client *apiclient.APIClient, daplaTeamSlug string) (bool, error) {

	resp, err := client.Teams().GetFeatures(ctx, &protoapi.GetFeaturesRequest{Slug: daplaTeamSlug})
	if err != nil {
		return false, err
	}

	var aiFeatureIsEnabled bool

	for _, feat := range resp.Features {
		// For now we only check if it's enabled in the test environment
		if feat.Name == "ai" && feat.Env == "test" {
			aiFeatureIsEnabled = true
		}
	}

	return aiFeatureIsEnabled, nil
}

// Check if the vertexai API is enabled in a given Google project
func isVertexAIEnabled(svc *serviceusage.Service, projectID string) (bool, error) {
	name := fmt.Sprintf("projects/%s/services/aiplatform.googleapis.com", projectID)
	service, err := svc.Services.Get(name).Do()
	if err != nil {
		return false, err
	}
	return service.State == "ENABLED", nil
}

// Enable or disable the vertexAI API in a given google project
func setVertexAIEnabled(svc *serviceusage.Service, projectID string, enabled bool) error {
	name := fmt.Sprintf("projects/%s/services/aiplatform.googleapis.com", projectID)
	if enabled {
		op, err := svc.Services.Enable(name, &serviceusage.EnableServiceRequest{}).Do()
		if err != nil {
			return err
		}
		if op.Error != nil {
			return fmt.Errorf("enable Vertex AI API for project %q: %s", projectID, op.Error.Message)
		}
		return nil
	}

	op, err := svc.Services.Disable(name, &serviceusage.DisableServiceRequest{}).Do()
	if err != nil {
		return err
	}
	if op.Error != nil {
		return fmt.Errorf("disable Vertex AI API for project %q: %s", projectID, op.Error.Message)
	}
	return nil
}

// Check if a dapla test project exists for a given dapla team
func projectExists(svc *cloudresourcemanager.Service, daplaTeamSlug string) (*string, error) {
	resp, err := svc.Projects.Search().Query(fmt.Sprintf("projectId:%s", daplaTeamSlug)).Do()
	if err != nil {
		return nil, err
	}

	for _, project := range resp.Projects {
		if testProjectID := daplaTeamSlug + "-t"; project.ProjectId == testProjectID {
			return &testProjectID, nil
		}
	}

	return nil, nil
}

// Delete implements [reconcilers.Reconciler].
func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	panic("unimplemented")
}
