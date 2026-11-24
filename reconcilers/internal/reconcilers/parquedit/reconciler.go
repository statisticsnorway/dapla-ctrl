package parquedit

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

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
	environments     []string
	labradoorConfigs map[string]labradoorConfig
	labradoorClients map[string]client
}

type labradoorConfig struct {
	Endpoint string
	Secret   string
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
	specs := make([]*protoapi.ReconcilerConfigSpec, 0, 2*len(r.environments))
	for _, env := range r.environments {
		env = strings.ToLower(env)
		specs = append(specs,
			&protoapi.ReconcilerConfigSpec{
				Key:         env + "_" + configExternalReconcilerEndpointKey,
				DisplayName: "Labradoor endpoint (" + env + ")",
				Description: "Labradoor endpoint (" + env + ")",
				Secret:      false,
			},
			&protoapi.ReconcilerConfigSpec{
				Key:         env + "_" + configExternalReconcilerSecretKey,
				DisplayName: "Labradoor token (" + env + ")",
				Description: "Labradoor bearer token (" + env + ")",
				Secret:      true,
			},
		)
	}

	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "Parquedit",
		Description: "Manage Parquedit resources in dapla-ffunk's databases",
		MemberAware: true,
		Config:      specs,
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	return r.syncTeam(ctx, client, daplaTeam.Slug, daplaTeam.HasParquedit)
}

func (r *reconciler) syncTeam(ctx context.Context, apiClient *apiclient.APIClient, team string, enabled bool) error {
	clients, err := r.getClients(ctx, apiClient)
	if err != nil {
		return fmt.Errorf("error getting reconciler config: %w", err)
	}

	var errs []error
	for env, client := range clients {
		hasParquedit, err := client.HasParquedit(ctx, team)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: check if team has parquedit: %w", env, err))
			continue
		}
		if hasParquedit == enabled {
			continue
		}

		if enabled {
			err = client.EnableParquedit(ctx, team)
		} else {
			err = client.DisableParquedit(ctx, team)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: set parquedit enabled=%t: %w", env, enabled, err))
		}
	}
	return errors.Join(errs...)
}

func (r *reconciler) getClients(ctx context.Context, apiClient *apiclient.APIClient) (map[string]client, error) {
	cfg, err := apiClient.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return nil, fmt.Errorf("get reconciler config: %w", err)
	}

	configs := make(map[string]labradoorConfig, len(r.environments))

	for _, c := range cfg.Nodes {
		parts := strings.Split(c.Key, "_")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid config key: %q", c.Key)
		}
		env := parts[0]
		key := parts[1]

		config := configs[env]
		switch key {
		case configExternalReconcilerEndpointKey:
			config.Endpoint = c.Value
		case configExternalReconcilerSecretKey:
			config.Secret = c.Value
		default:
			return nil, fmt.Errorf("unknown config key: %q", c.Key)
		}
		configs[env] = config
	}

	if maps.Equal(r.labradoorConfigs, configs) && r.labradoorClients != nil {
		return r.labradoorClients, nil
	}

	clients := make(map[string]client, len(configs))

	for env, config := range configs {
		if config.Endpoint == "" || config.Secret == "" {
			continue
		}
		clients[env], err = labreconciler.New(config.Endpoint, config.Secret)
		if err != nil {
			return nil, fmt.Errorf("create labreconciler client: %w", err)
		}
	}

	r.labradoorClients = clients
	r.labradoorConfigs = configs

	return clients, nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	return r.syncTeam(ctx, client, daplaTeam.Slug, false)
}
