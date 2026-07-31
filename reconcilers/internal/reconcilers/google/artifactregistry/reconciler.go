package artifactregistry

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers"
)

const (
	reconcilerName = "google:artifactregistry"

	configRegionKey = "region"
)

type reconciler struct {
	config groupSaConfig
}

type groupSaConfig struct {
	Region string
}

type optFunc func(*reconciler)

func New(ctx context.Context, opts ...optFunc) (reconcilers.Reconciler, error) {
	r := new(reconciler)

	for _, opt := range opts {
		opt(r)
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
				Key:         configRegionKey,
				DisplayName: "Artifact Registry region",
				Description: "Region to manage AR repos in",
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
		case configRegionKey:
			gac.Region = c.Value
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
