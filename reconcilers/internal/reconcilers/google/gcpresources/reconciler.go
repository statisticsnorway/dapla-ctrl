package gcpresources

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const reconcilerName = "google:gcpresources"

type Config struct {
	TagKeyNamespacedName string
	EnvParentFolders     map[string]string
	PAM                  PAMConfig
}

type reconciler struct {
	client ResourceManager
	pam    PrivilegedAccessManager
	cfg    Config
}

type optFunc func(*reconciler)

func WithResourceManager(c ResourceManager) optFunc {
	return func(r *reconciler) {
		r.client = c
	}
}

func WithPrivilegedAccessManager(c PrivilegedAccessManager) optFunc {
	return func(r *reconciler) {
		r.pam = c
	}
}

func New(ctx context.Context, cfg Config, services *google.Services, opts ...optFunc) (reconcilers.Reconciler, error) {
	err := validateConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	cfg.PAM.Role = strings.TrimSpace(cfg.PAM.Role)
	cfg.PAM.Principals = slices.Clone(cfg.PAM.Principals)
	for i, principal := range cfg.PAM.Principals {
		cfg.PAM.Principals[i] = strings.TrimSpace(principal)
	}

	r := &reconciler{cfg: cfg}
	if services != nil {
		r.pam = &GooglePrivilegedAccessManager{client: services.PrivilegedAccessManager}
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.client == nil {
		c, err := NewGoogleResourceManager(ctx, services)
		if err != nil {
			return nil, fmt.Errorf("create GCP client: %w", err)
		}
		r.client = c
	}

	return r, nil
}

func validateConfig(cfg Config) error {
	cfg.TagKeyNamespacedName = strings.TrimSpace(cfg.TagKeyNamespacedName)
	if cfg.TagKeyNamespacedName == "" {
		return fmt.Errorf("tag key name is required")
	}

	tagKeyID := strings.TrimPrefix(cfg.TagKeyNamespacedName, "tagKeys/")
	if tagKeyID == cfg.TagKeyNamespacedName || tagKeyID == "" {
		return fmt.Errorf("tag key name must be in format tagKeys/{id}")
	}

	if len(cfg.EnvParentFolders) == 0 {
		return fmt.Errorf("at least one environment parent folder must be configured")
	}
	return cfg.PAM.validate()
}

func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "GCP Resources reconciler",
		Description: "Create team folders, attach team tags and manage PAM entitlements for Managed teams",
		MemberAware: false,
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	teamSlug := daplaTeam.Slug
	teamFoldersClient := client.GcpTeamResources()

	var pamCfg PAMConfig
	if daplaTeam.IsManaged {
		pamCfg = r.cfg.PAM
		if !pamCfg.configured() {
			log.Warn("skipping PAM entitlements: eligible principals, IAM role and maximum grant duration must all be configured")
		}
	}

	tagValueName, err := r.client.GetTagValue(ctx, r.cfg.TagKeyNamespacedName, teamSlug)
	if errors.Is(err, ErrNotFound) {
		tagValueName, err = r.client.CreateTagValue(ctx, r.cfg.TagKeyNamespacedName, teamSlug)
		if err != nil {
			return fmt.Errorf("create tag value for team %q: %w", teamSlug, err)
		}
		log.WithField("tag_value_name", tagValueName).Info("created tag value")
	} else if err != nil {
		return fmt.Errorf("get tag value for team %q: %w", teamSlug, err)
	} else {
		log.WithField("tag_value_name", tagValueName).Debug("resolved GCP tag value")
	}

	for env, parentFolderNumber := range r.cfg.EnvParentFolders {
		if err := r.reconcileEnvFolder(ctx, teamFoldersClient, teamSlug, env, parentFolderNumber, tagValueName, pamCfg, log); err != nil {
			return err
		}
	}

	return nil
}

func (r *reconciler) reconcileEnvFolder(
	ctx context.Context,
	teamFoldersClient protoapi.GcpTeamResourcesClient,
	teamSlug, env, parentFolderNumber, tagValueName string,
	pamCfg PAMConfig,
	log logrus.FieldLogger,
) error {
	log = log.WithField("env", env)

	resp, err := teamFoldersClient.GetTeamFolder(ctx, &protoapi.GetGcpTeamFolderRequest{
		TeamSlug: teamSlug,
		Env:      env,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		return fmt.Errorf("get stored folder for env %q: %w", env, err)
	}

	var fID string
	if err == nil {
		fID = resp.Folder.FolderId
		log.WithField("folder_id", fID).Debug("folder already exist in db")
	} else {
		parent := fmt.Sprintf("folders/%s", parentFolderNumber)
		fID, err = r.client.GetFolder(ctx, teamSlug, parent)

		if errors.Is(err, ErrNotFound) {
			fID, err = r.client.CreateFolder(ctx, teamSlug, parent)
			if err != nil {
				return fmt.Errorf("create GCP folder for team %q in env %q: %w", teamSlug, env, err)
			}
			log.WithField("folder_id", fID).Info("created GCP folder")
		} else if err != nil {
			return fmt.Errorf("get GCP folder for team %q in env %q: %w", teamSlug, env, err)
		} else {
			log.WithField("folder_id", fID).Debug("resolved GCP folder id")
		}

		if _, err := teamFoldersClient.UpsertTeamFolder(ctx, &protoapi.UpsertGcpTeamFolderRequest{
			Folder: &protoapi.GcpTeamFolder{
				TeamSlug: teamSlug,
				Env:      env,
				FolderId: fID,
			},
		}); err != nil {
			return fmt.Errorf("store folder ID for env %q: %w", env, err)
		}
	}

	if err := r.client.TagFolder(ctx, fID, tagValueName); err != nil {
		return fmt.Errorf("tag folder %q in env %q: %w", fID, env, err)
	}

	if pamCfg.configured() {
		if r.pam == nil {
			return fmt.Errorf("PAM client is not configured")
		}
		if err := r.pam.EnsureEntitlement(ctx, fID, pamCfg.Role, pamCfg.Principals, pamCfg.MaxGrantDuration); err != nil {
			return fmt.Errorf("reconcile PAM entitlement for folder %q in env %q: %w", fID, env, err)
		}
	}

	return nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
