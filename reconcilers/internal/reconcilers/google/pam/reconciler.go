package pam

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers"
)

const (
	reconcilerName = "google:pam"

	configRoleKey                 = "role"
	configPrincipalsKey           = "principals"
	configMaxGrantDurationKey     = "max_grant_duration"
	configConditionTemplateKey    = "condition_template"
	configTeamAllowlistKey        = "team_allowlist"
	configEnvironmentAllowlistKey = "environment_allowlist"
)

type Config struct {
	Environments []string
}

type pamConfig struct {
	Principals           []string
	Role                 string
	MaxGrantDuration     time.Duration
	ConditionTemplate    *template.Template
	TeamAllowlist        []string
	EnvironmentAllowlist []string
}

func (c pamConfig) configured() bool {
	return len(c.Principals) != 0 && c.Role != "" && c.MaxGrantDuration != 0 && c.ConditionTemplate != nil
}

func (c Config) validate() error {
	if len(c.Environments) == 0 {
		return fmt.Errorf("at least one PAM environment must be configured")
	}
	for _, env := range c.Environments {
		if strings.TrimSpace(env) == "" {
			return fmt.Errorf("PAM environments must not contain empty entries")
		}
	}
	return nil
}

func (c pamConfig) validate() error {
	for _, teamSlug := range c.TeamAllowlist {
		if teamSlug == "" {
			return fmt.Errorf("PAM team allowlist must not contain empty entries")
		}
	}
	for _, env := range c.EnvironmentAllowlist {
		if env == "" {
			return fmt.Errorf("PAM environment allowlist must not contain empty entries")
		}
	}
	for _, principal := range c.Principals {
		if strings.TrimSpace(principal) == "" {
			return fmt.Errorf("PAM principals must not contain empty entries")
		}
	}
	if c.MaxGrantDuration != 0 && (c.MaxGrantDuration < 30*time.Minute || c.MaxGrantDuration > 168*time.Hour) {
		return fmt.Errorf("PAM maximum grant duration must be between 30m and 168h")
	}
	return nil
}

type reconciler struct {
	client PrivilegedAccessManager
	cfg    Config
}

type optFunc func(*reconciler)

func WithPrivilegedAccessManager(c PrivilegedAccessManager) optFunc {
	return func(r *reconciler) {
		r.client = c
	}
}

func New(ctx context.Context, cfg Config, services *google.Services, opts ...optFunc) (reconcilers.Reconciler, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	cfg.Environments = slices.Clone(cfg.Environments)

	r := &reconciler{cfg: cfg}
	if services != nil {
		r.client = &GooglePrivilegedAccessManager{client: services.PrivilegedAccessManager}
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.client == nil {
		return nil, fmt.Errorf("PAM client is not configured")
	}
	return r, nil
}

func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "PAM reconciler",
		Description: "Manage PAM entitlements on team folders for Managed teams",
		MemberAware: false,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         configRoleKey,
				DisplayName: "Managed team folder IAM role",
				Description: "Full IAM role name granted by PAM. No default is applied.",
				Secret:      false,
			},
			{
				Key:         configPrincipalsKey,
				DisplayName: "Managed team folder eligible principals",
				Description: "Comma-separated IAM principal identifiers eligible to request access.",
				Secret:      false,
			},
			{
				Key:         configMaxGrantDurationKey,
				DisplayName: "Managed team folder maximum grant duration",
				Description: "Go duration between 30m and 168h, e.g. 1h.",
				Secret:      false,
			},
			{
				Key:         configConditionTemplateKey,
				DisplayName: "Managed team folder IAM condition template",
				Description: "Go text template rendering an IAM CEL condition, using .TeamSlug and .Env.",
				Secret:      false,
			},
			{
				Key:         configTeamAllowlistKey,
				DisplayName: "Team allowlist",
				Description: "Comma-separated list of exact, case-sensitive team slugs. Empty allows all Managed teams.",
				Secret:      false,
			},
			{
				Key:         configEnvironmentAllowlistKey,
				DisplayName: "Environment allowlist",
				Description: "Comma-separated list of exact, case-sensitive environment names. Empty allows all configured environments.",
				Secret:      false,
			},
		},
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	if !daplaTeam.IsManaged {
		return nil
	}
	cfg, err := r.getConfig(ctx, client)
	if err != nil {
		return fmt.Errorf("error getting reconciler config: %w", err)
	}
	if len(cfg.TeamAllowlist) > 0 && !slices.Contains(cfg.TeamAllowlist, daplaTeam.Slug) {
		log.Debug("skipping PAM entitlements: team is not allowlisted")
		return nil
	}
	if !cfg.configured() {
		log.Warn("skipping PAM entitlements: eligible principals, IAM role, maximum grant duration and IAM condition template must all be configured")
		return nil
	}

	for _, env := range r.cfg.Environments {
		if len(cfg.EnvironmentAllowlist) > 0 && !slices.Contains(cfg.EnvironmentAllowlist, env) {
			log.WithField("env", env).Debug("skipping PAM entitlement: environment is not allowlisted")
			continue
		}
		resp, err := client.GcpTeamResources().GetTeamFolder(ctx, &protoapi.GetGcpTeamFolderRequest{
			TeamSlug: daplaTeam.Slug,
			Env:      env,
		})
		if err != nil {
			return fmt.Errorf("get stored folder for team %q in env %q: %w", daplaTeam.Slug, env, err)
		}
		folderID := resp.GetFolder().GetFolderId()
		if folderID == "" {
			return fmt.Errorf("stored folder ID for team %q in env %q is empty", daplaTeam.Slug, env)
		}
		conditionExpression, err := cfg.renderPAMCondition(daplaTeam.Slug, env)
		if err != nil {
			return fmt.Errorf("PAM condition for folder %q in env %q: %w", folderID, env, err)
		}
		if err := r.client.EnsureEntitlement(ctx, folderID, cfg.Role, conditionExpression, cfg.Principals, cfg.MaxGrantDuration); err != nil {
			return fmt.Errorf("reconcile PAM entitlement for folder %q in env %q: %w", folderID, env, err)
		}
	}
	return nil
}

func (r *reconciler) getConfig(ctx context.Context, client *apiclient.APIClient) (pamConfig, error) {
	config, err := client.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return pamConfig{}, fmt.Errorf("get reconciler config: %w", err)
	}

	cfg := pamConfig{}
	var conditionTemplate string
	for _, c := range config.Nodes {
		value := strings.TrimSpace(c.Value)
		switch c.Key {
		case configRoleKey:
			cfg.Role = value
		case configPrincipalsKey:
			cfg.Principals = parseCommaSeparatedList(value)
		case configMaxGrantDurationKey:
			if value != "" {
				cfg.MaxGrantDuration, err = time.ParseDuration(value)
				if err != nil {
					return pamConfig{}, fmt.Errorf("parse PAM maximum grant duration: %w", err)
				}
			}
		case configConditionTemplateKey:
			conditionTemplate = value
		case configTeamAllowlistKey:
			cfg.TeamAllowlist = parseCommaSeparatedList(value)
		case configEnvironmentAllowlistKey:
			cfg.EnvironmentAllowlist = parseCommaSeparatedList(value)
		default:
			return pamConfig{}, fmt.Errorf("unknown config key %q", c.Key)
		}
	}
	if err := cfg.validate(); err != nil {
		return pamConfig{}, fmt.Errorf("validate reconciler config: %w", err)
	}
	if conditionTemplate != "" {
		cfg.ConditionTemplate, err = template.New("PAM IAM condition").Option("missingkey=error").Parse(conditionTemplate)
		if err != nil {
			return pamConfig{}, fmt.Errorf("parse PAM IAM condition template: %w", err)
		}
	}
	return cfg, nil
}

func parseCommaSeparatedList(value string) []string {
	if value == "" {
		return nil
	}
	entries := strings.Split(value, ",")
	for i, entry := range entries {
		entries[i] = strings.TrimSpace(entry)
	}
	return entries
}

func (c pamConfig) renderPAMCondition(teamSlug, env string) (string, error) {
	var rendered bytes.Buffer
	if err := c.ConditionTemplate.Execute(&rendered, struct {
		TeamSlug string
		Env      string
	}{TeamSlug: teamSlug, Env: env}); err != nil {
		return "", fmt.Errorf("render PAM IAM condition: %w", err)
	}
	expression := strings.TrimSpace(rendered.String())
	if expression == "" {
		return "", fmt.Errorf("PAM IAM condition rendered to an empty expression")
	}
	return expression, nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("PAM entitlement deletion is not implemented")
	return nil
}
