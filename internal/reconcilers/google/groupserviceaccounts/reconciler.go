package groupserviceaccounts

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/iterator"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/protoapi"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
)

const (
	reconcilerName = "google:groupserviceaccounts"

	saDescription = "source:dapla-api"

	configDaplaGroupSaProjectIdKey = "daplaGroupSaProjectId"
)

type reconciler struct {
	mainCtx context.Context
	client  *iam.Service
	config  groupSaConfig
}

type groupSaConfig struct {
	DaplaGroupSaProjectId string
}

func New(ctx context.Context) (reconcilers.Reconciler, error) {
	client, err := iam.NewService(ctx)
	if err != nil {
		return nil, err
	}

	r := &reconciler{
		mainCtx: ctx,
		client:  client,
	}

	return r, nil
}

func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "Google Group SA reconciler",
		Description: "Create Group SAs in Google for Dapla groups",
		MemberAware: true,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         configDaplaGroupSaProjectIdKey,
				DisplayName: "Dapla Group SA Project ID",
				Description: "The Google project to create Dapla Group SAs in",
				Secret:      false,
			},
		},
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error {

	// Iterate through all the groups in the team, and reconcile them one by one
	groupsIt := iterator.New(ctx, 100, func(limit, offset int64) (*protoapi.ListTeamGroupsResponse, error) {
		return client.Teams().Groups(ctx, &protoapi.ListTeamGroupsRequest{
			Slug: naisTeam.Slug,
		})
	})

	for groupsIt.Next() {
		if err := r.reconcileGroup(groupsIt.Value().Group.Name); err != nil {
			return fmt.Errorf("reconcile group %q: %w", groupsIt.Value().Group.Name, err)
		}
	}

	return nil
}

func (r *reconciler) reconcileGroup(groupName string) error {
	saName := fmt.Sprintf("projects/-/serviceAccounts/%s@%s.iam.gserviceaccount.com", groupName, r.config.DaplaGroupSaProjectId)

	sa, err := r.client.Projects.ServiceAccounts.Get(saName).Do()
	var apiError *googleapi.Error
	if errors.As(err, &apiError) {
		if apiError.Code == http.StatusNotFound {
			return r.createServiceAccount(groupName, r.config.DaplaGroupSaProjectId)
		}
		return fmt.Errorf("unexpected status code getting service account: %w", err)
	} else if err != nil {
		return fmt.Errorf("unexpected error getting service account: %w", err)
	}

	if sa.Description == saDescription {
		return nil
	}

	req := iam.PatchServiceAccountRequest{
		ServiceAccount: &iam.ServiceAccount{
			Description: saDescription,
		},
		UpdateMask: "description",
	}

	if _, err := r.client.Projects.ServiceAccounts.Patch(saName, &req).Do(); err != nil {
		return fmt.Errorf("unexpected error patching sa description: %w", err)
	}

	return nil
}

func (r *reconciler) createServiceAccount(groupName string, projectId string) error {
	req := iam.CreateServiceAccountRequest{
		AccountId: groupName,
		ServiceAccount: &iam.ServiceAccount{
			Description: saDescription,
		},
	}

	if _, err := r.client.Projects.ServiceAccounts.Create(fmt.Sprintf("projects/%s", projectId), &req).Do(); err != nil {
		return fmt.Errorf("unexpected error creating service account: %w", err)
	}

	return nil
}

func (r *reconciler) updateConfig(ctx context.Context, client *apiclient.APIClient) (*groupSaConfig, error) {
	config, err := client.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return nil, fmt.Errorf("get reconciler config: %w", err)
	}

	gac := groupSaConfig{}

	for _, c := range config.Nodes {
		switch c.Key {
		case configDaplaGroupSaProjectIdKey:
			gac.DaplaGroupSaProjectId = c.Value
		default:
			return nil, fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	return &gac, nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
