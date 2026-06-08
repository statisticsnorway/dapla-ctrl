package parquedit

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/labreconciler"
)

const (
	reconcilerName = "dapla:ffunk:parquedit"

	configExternalReconcilerEndpointKey = "labReconcilerEndpoint"
	configExternalReconcilerSecretKey   = "labReconcilerSecret"
)

type reconciler struct {
	config       config
	client       client
	staticClient client
}

type config struct {
	LabReconcilerEndpoint string
	LabReconcilerSecret   string
}

type client interface {
	EnableParquedit(ctx context.Context, team string) error
	HasParquedit(ctx context.Context, team string) (bool, error)
	DisableParquedit(ctx context.Context, team string) error
}

type optFunc func(*reconciler)

func New(ctx context.Context, opts ...optFunc) (*reconciler, error) {
	r := new(reconciler)

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "Parquedit",
		Description: "Manage Parquedit resources in ffunk's database",
		MemberAware: true,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         configExternalReconcilerEndpointKey,
				DisplayName: "Lab Reconciler Endpoint",
				Description: "The endpoint to use to access the external (in Dapla Lab) reconciler application",
				Secret:      false,
			},
			{
				Key:         configExternalReconcilerSecretKey,
				DisplayName: "Lab Reconciler Secret",
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
	parqueditClient, err := r.getClient(ctx, client)
	if err != nil {
		return fmt.Errorf("error getting reconciler config: %w", err)
	}

	hasParquedit, err := parqueditClient.HasParquedit(ctx, naisTeam.Slug)
	if err != nil {
		return fmt.Errorf("check if team has parquedit: %w", err)
	}

	shouldHaveParquedit := naisTeam.HasParquedit

	if hasParquedit == shouldHaveParquedit {
		return nil
	}

	if shouldHaveParquedit {
		return parqueditClient.EnableParquedit(ctx, naisTeam.Slug)
	}

	return parqueditClient.DisableParquedit(ctx, naisTeam.Slug)
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
			rc.LabReconcilerEndpoint = c.Value
		case configExternalReconcilerSecretKey:
			rc.LabReconcilerSecret = c.Value
		default:
			return nil, fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if rc == r.config && r.client != nil {
		return r.client, nil
	}

	newClient, err := labreconciler.New(rc.LabReconcilerEndpoint, rc.LabReconcilerSecret)
	if err != nil {
		return nil, fmt.Errorf("create labreconciler client: %w", err)
	}

	r.client = newClient
	r.config = rc

	return newClient, nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error {
	parqueditClient, err := r.getClient(ctx, client)
	if err != nil {
		return err
	}

	hasParquedit, err := parqueditClient.HasParquedit(ctx, naisTeam.Slug)
	if err != nil {
		return err
	}

	if hasParquedit {
		return parqueditClient.DisableParquedit(ctx, naisTeam.Slug)
	}

	return nil
}
