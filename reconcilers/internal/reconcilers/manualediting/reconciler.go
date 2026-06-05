package manualediting

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/labreconciler"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers"
)

const (
	reconcilerName = "dapla:ffunk:manualediting"

	configExternalReconcilerEndpointKey = "externalReconcilerEndpoint"
	configExternalReconcilerSecretKey   = "externalReconcilerSecret"
)

type reconciler struct {
	config       config
	client       client
	staticClient client
}

type config struct {
	ExternalReconcilerEndpoint string
	ExternalReconcilerSecret   string
}

type client interface {
	EnableFfunkEditing(ctx context.Context, team string) error
	HasFfunkEditing(ctx context.Context, team string) (bool, error)
	DisableFfunkEditing(ctx context.Context, team string) error
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
		DisplayName: "Ffunk editering reconciler",
		Description: "Manage manual editing resources in ffunk's database",
		MemberAware: true,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         configExternalReconcilerEndpointKey,
				DisplayName: "External Reconciler Endpoint",
				Description: "The endpoint to use to access the external (in Dapla Lab) reconciler application",
				Secret:      false,
			},
			{
				Key:         configExternalReconcilerSecretKey,
				DisplayName: "External Reconciler Secret",
				Description: "The secret to use for authorization of request to the external reconciler",
				Secret:      true,
			},
		},
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error {
	editingClient, err := r.getClient(ctx, client)
	if err != nil {
		return fmt.Errorf("error getting reconciler config: %w", err)
	}

	hasEditing, err := editingClient.HasFfunkEditing(ctx, naisTeam.Slug)
	if err != nil {
		return fmt.Errorf("check if team has ffunk editing: %w", err)
	}

	shouldHaveEditing := naisTeam.HasManualEditing

	if hasEditing == shouldHaveEditing {
		return nil
	}

	if shouldHaveEditing {
		return editingClient.EnableFfunkEditing(ctx, naisTeam.Slug)
	}

	return editingClient.DisableFfunkEditing(ctx, naisTeam.Slug)
}

func (r *reconciler) getClient(ctx context.Context, client *apiclient.APIClient) (client, error) {
	if r.staticClient != nil {
		return r.staticClient, nil
	}

	cfg, err := client.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return nil, fmt.Errorf("get reconciler config: %w", err)
	}

	rc := config{}

	for _, c := range cfg.Nodes {
		switch c.Key {
		case configExternalReconcilerEndpointKey:
			rc.ExternalReconcilerEndpoint = c.Value
		case configExternalReconcilerSecretKey:
			rc.ExternalReconcilerSecret = c.Value
		default:
			return nil, fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if rc == r.config && r.client != nil {
		return r.client, nil
	}

	newClient, err := labreconciler.New(rc.ExternalReconcilerEndpoint, rc.ExternalReconcilerSecret)
	if err != nil {
		return nil, fmt.Errorf("create labreconciler client: %w", err)
	}

	r.client = newClient
	r.config = rc

	return newClient, nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error {
	editingClient, err := r.getClient(ctx, client)
	if err != nil {
		return err
	}

	hasEditing, err := editingClient.HasFfunkEditing(ctx, naisTeam.Slug)
	if err != nil {
		return err
	}

	if hasEditing {
		return editingClient.DisableFfunkEditing(ctx, naisTeam.Slug)
	}

	return nil
}
