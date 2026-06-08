package groupserviceaccounts

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/iterator"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers"
)

const (
	reconcilerName = "google:groupserviceaccounts"

	saDescription = "source:dapla-api"

	configDaplaGroupSaProjectIdsKey = "daplaGroupSaProjectIds"
)

type reconciler struct {
	client GroupServiceAccounts
	config groupSaConfig
}

type groupSaConfig struct {
	DaplaGroupSaProjectIds []string
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
				Key:         configDaplaGroupSaProjectIdsKey,
				DisplayName: "Dapla Group SA Project IDs",
				Description: "Comma-separated list of Google projects to create Dapla Group SAs in",
				Secret:      false,
			},
		},
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	if err := r.updateConfig(ctx, client); err != nil {
		return fmt.Errorf("error getting reconciler config: %w", err)
	}

	// Iterate through all the groups in the team, and reconcile them one by one
	groupsIt := iterator.New(ctx, 100, func(limit, offset int64) (*protoapi.ListTeamGroupsResponse, error) {
		return client.Teams().Groups(ctx, &protoapi.ListTeamGroupsRequest{
			Slug: daplaTeam.Slug,
		})
	})

	for groupsIt.Next() {
		groupName := groupsIt.Value().Group.Name
		for _, envProjectId := range r.config.DaplaGroupSaProjectIds {
			if err := r.reconcileEnv(ctx, envProjectId, groupName); err != nil {
				return fmt.Errorf("reconcile group %q: %w", groupName, err)
			}
		}
	}

	return nil
}

func (r *reconciler) reconcileEnv(ctx context.Context, envProjectId, groupName string) error {
	sa, err := r.client.GetOrCreate(ctx, groupName, envProjectId)
	if err != nil {
		return err
	}

	if sa.Description == saDescription {
		return nil
	}

	if err := r.client.UpdateDescription(ctx, groupName, saDescription, envProjectId); err != nil {
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
		case configDaplaGroupSaProjectIdsKey:
			gac.DaplaGroupSaProjectIds = strings.Split(c.Value, ",")
		default:
			return fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	r.config = gac
	return nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
