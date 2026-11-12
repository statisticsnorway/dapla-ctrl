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

const reconcilerName = "google:pam"

type Config struct {
	Environments      []string
	Principals        []string
	Role              string
	MaxGrantDuration  time.Duration
	ConditionTemplate string
}

func (c Config) configured() bool {
	return len(c.Principals) != 0 && c.Role != "" && c.MaxGrantDuration != 0 && c.ConditionTemplate != ""
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
	client               PrivilegedAccessManager
	cfg                  Config
	pamConditionTemplate *template.Template
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
	cfg.Role = strings.TrimSpace(cfg.Role)
	cfg.ConditionTemplate = strings.TrimSpace(cfg.ConditionTemplate)
	cfg.Environments = slices.Clone(cfg.Environments)
	cfg.Principals = slices.Clone(cfg.Principals)
	for i, principal := range cfg.Principals {
		cfg.Principals[i] = strings.TrimSpace(principal)
	}

	r := &reconciler{cfg: cfg}
	if cfg.ConditionTemplate != "" {
		var err error
		r.pamConditionTemplate, err = template.New("PAM IAM condition").Option("missingkey=error").Parse(cfg.ConditionTemplate)
		if err != nil {
			return nil, fmt.Errorf("parse PAM IAM condition template: %w", err)
		}
	}
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
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	if !daplaTeam.IsManaged {
		return nil
	}
	if !r.cfg.configured() {
		log.Warn("skipping PAM entitlements: eligible principals, IAM role, maximum grant duration and IAM condition template must all be configured")
		return nil
	}

	for _, env := range r.cfg.Environments {
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
		conditionExpression, err := r.renderPAMCondition(daplaTeam.Slug, env)
		if err != nil {
			return fmt.Errorf("PAM condition for folder %q in env %q: %w", folderID, env, err)
		}
		if err := r.client.EnsureEntitlement(ctx, folderID, r.cfg.Role, conditionExpression, r.cfg.Principals, r.cfg.MaxGrantDuration); err != nil {
			return fmt.Errorf("reconcile PAM entitlement for folder %q in env %q: %w", folderID, env, err)
		}
	}
	return nil
}

func (r *reconciler) renderPAMCondition(teamSlug, env string) (string, error) {
	var rendered bytes.Buffer
	if err := r.pamConditionTemplate.Execute(&rendered, struct {
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
