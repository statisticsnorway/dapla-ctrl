package groupserviceaccounts

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/iterator"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/protoapi"
)

const (
	reconcilerName = "google:groupserviceaccounts"

	saDescription = "source:dapla-api"

	configDaplaGroupSaProjectIdKey = "daplaGroupSaProjectId"
)

type reconciler struct {
	client GroupServiceAccounts
	config groupSaConfig
}

type groupSaConfig struct {
	DaplaGroupSaProjectId string
}

type optFunc func(*reconciler)

func WithGroupServiceAccounts(gsa GroupServiceAccounts) optFunc {
	return func(r *reconciler) {
		r.client = gsa
	}
}

func New(ctx context.Context, opts ...optFunc) (reconcilers.Reconciler, error) {
	r := new(reconciler)

	for _, opt := range opts {
		opt(r)
	}

	if r.client == nil {
		client, err := NewGoogleServiceAccounts(ctx)
		if err != nil {
			return nil, err
		}
		r.client = client
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
	if err := r.updateConfig(ctx, client); err != nil {
		return fmt.Errorf("error getting reconciler config: %w", err)
	}

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
	sa, err := r.client.GetOrCreate(groupName, r.config.DaplaGroupSaProjectId)
	if err != nil {
		return err
	}

	if sa.Description == saDescription {
		return nil
	}

	if err := r.client.UpdateDescription(groupName, saDescription, r.config.DaplaGroupSaProjectId); err != nil {
		return err
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

	gac := groupSaConfig{}

	for _, c := range config.Nodes {
		switch c.Key {
		case configDaplaGroupSaProjectIdKey:
			gac.DaplaGroupSaProjectId = c.Value
		default:
			return fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	r.config = gac
	return nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
