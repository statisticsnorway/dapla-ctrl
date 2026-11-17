package parquedit

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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
	config  config
	clients map[string]client
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
		Description: "Manage Parquedit resources in ffunk's production and test databases",
		MemberAware: true,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         configExternalReconcilerEndpointKey,
				DisplayName: "Lab Reconciler Endpoint",
				Description: "Labradoor endpoints: prod:https://labradoor.lab.dapla.ssb.no,test:https://labradoor.lab.dapla-test.ssb.no.",
				Secret:      false,
			},
			{
				Key:         configExternalReconcilerSecretKey,
				DisplayName: "Lab Reconciler Secret",
				Description: "Labradoor bearer tokens: prod:<production token>,test:<test token>. Use token values, not secret names.",
				Secret:      true,
			},
		},
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
	for _, env := range []string{"prod", "test"} {
		parqueditClient, ok := clients[env]
		if !ok {
			continue
		}

		hasParquedit, err := parqueditClient.HasParquedit(ctx, team)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: check if team has parquedit: %w", env, err))
			continue
		}
		if hasParquedit == enabled {
			continue
		}

		if enabled {
			err = parqueditClient.EnableParquedit(ctx, team)
		} else {
			err = parqueditClient.DisableParquedit(ctx, team)
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

	if rc == r.config && r.clients != nil {
		return r.clients, nil
	}

	clients, err := rc.newClients()
	if err != nil {
		return nil, err
	}

	r.clients = clients
	r.config = rc

	return clients, nil
}

func (c config) newClients() (map[string]client, error) {
	endpoints, err := parseEnvironmentValues(c.LabReconcilerEndpoint, configExternalReconcilerEndpointKey)
	if err != nil {
		return nil, err
	}
	secrets, err := parseEnvironmentValues(c.LabReconcilerSecret, configExternalReconcilerSecretKey)
	if err != nil {
		return nil, err
	}

	clients := make(map[string]client, len(endpoints))
	for _, env := range []string{"prod", "test"} {
		endpoint, hasEndpoint := endpoints[env]
		secret, hasSecret := secrets[env]
		if hasEndpoint != hasSecret {
			return nil, fmt.Errorf("%s: both %s and %s must be configured", env, configExternalReconcilerEndpointKey, configExternalReconcilerSecretKey)
		}
		if !hasEndpoint {
			continue
		}
		u, err := url.Parse(endpoint)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("%s: %s must be an HTTP(S) URL without credentials, query or fragment", env, configExternalReconcilerEndpointKey)
		}
		if strings.ContainsAny(secret, "\r\n") {
			return nil, fmt.Errorf("%s: %s must not contain line breaks", env, configExternalReconcilerSecretKey)
		}
		clients[env], err = labreconciler.New(strings.TrimRight(endpoint, "/"), secret)
		if err != nil {
			return nil, fmt.Errorf("%s: create labreconciler client: %w", env, err)
		}
	}
	return clients, nil
}

func parseEnvironmentValues(value, key string) (map[string]string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("%s must not be empty", key)
	}
	if !strings.Contains(value, ",") && !strings.HasPrefix(value, "prod:") && !strings.HasPrefix(value, "test:") {
		return map[string]string{"prod": value}, nil
	}

	values := make(map[string]string)
	for entry := range strings.SplitSeq(value, ",") {
		env, v, ok := strings.Cut(strings.TrimSpace(entry), ":")
		env, v = strings.TrimSpace(env), strings.TrimSpace(v)
		if !ok || (env != "prod" && env != "test") || v == "" {
			return nil, fmt.Errorf("%s must contain comma-separated prod:<value> and test:<value> entries", key)
		}
		if _, exists := values[env]; exists {
			return nil, fmt.Errorf("%s contains duplicate %s entries", key, env)
		}
		values[env] = v
	}
	if _, ok := values["prod"]; !ok {
		return nil, fmt.Errorf("%s must include prod", key)
	}
	return values, nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	return r.syncTeam(ctx, client, daplaTeam.Slug, false)
}
