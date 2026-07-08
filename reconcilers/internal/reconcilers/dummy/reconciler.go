package dummy

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
)

const (
	reconcilerName = "stat:dummy"
)

type reconciler struct{}

func New() *reconciler {
	return &reconciler{}
}

// Configuration implements [reconcilers.Reconciler].
func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "Our new dummy reconciler",
		Description: "n/a",
	}
}

// Name implements [reconcilers.Reconciler].
func (r *reconciler) Name() string {
	return reconcilerName
}

// Reconcile implements [reconcilers.Reconciler].
func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	f, err := os.Create("teams/" + daplaTeam.Slug)
	if err != nil {
		return fmt.Errorf("create team file: %w", err)
	}

	resp, err := client.Teams().GetFeatures(ctx, &protoapi.GetFeaturesRequest{Slug: daplaTeam.Slug})
	if err != nil {
		return err
	}

	for _, feat := range resp.Features {
		fmt.Fprintf(f, "%s\t%s\n", feat.Name, feat.Env)
	}

	return nil
}

// Delete implements [reconcilers.Reconciler].
func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	panic("unimplemented")
}
