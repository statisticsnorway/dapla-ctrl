// Reconciler which enables Google's Vertex AI (Gemini Enterprise Agent Platform)
// and sets the right IAM permissions for Dapla teams.
package ai

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/serviceusage/v1"
)

const (
	reconcilerName = "stat:ai"
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
	// - Give ssb.aiplatform.user to $TEAM_NAME-developers group
	// - Give ssb.aiplatform.user to dapla SA
	if aiFeatureIsEnabled && !vertexAIEnabled {
		setVertexAIEnabled(serviceUsageService, *testProjectID, true)
	} else if !aiFeatureIsEnabled && vertexAIEnabled {
		setVertexAIEnabled(serviceUsageService, *testProjectID, false)
	}

	return nil
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
