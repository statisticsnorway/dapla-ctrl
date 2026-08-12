package gcpresources

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const reconcilerName = "google:gcpresources"

type Config struct {
	TagKeyNamespacedName string
	EnvParentFolders     map[string]string
}

type reconciler struct {
	client ResourceManager
	cfg    Config
}

type optFunc func(*reconciler)

func WithResourceManager(c ResourceManager) optFunc {
	return func(r *reconciler) {
		r.client = c
	}
}

func New(ctx context.Context, cfg Config, opts ...optFunc) (reconcilers.Reconciler, error) {
	err := validateConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	r := &reconciler{cfg: cfg}

	for _, opt := range opts {
		opt(r)
	}

	if r.client == nil {
		c, err := NewGoogleResourceManager(ctx)
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
	return nil
}

func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "GCP Resources reconciler",
		Description: "Create team folders and attach new team tags",
		MemberAware: false,
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	teamSlug := daplaTeam.Slug
	teamFoldersClient := client.GcpTeamResources()

	tagValueName, err := r.client.GetOrCreateTagValue(ctx, r.cfg.TagKeyNamespacedName, teamSlug)
	if err != nil {
		return fmt.Errorf("get or create tag value for team %q: %w", teamSlug, err)
	}
	log.WithField("tag_value_name", tagValueName).Debug("resolved GCP tag value")

	for env, parentFolderNumber := range r.cfg.EnvParentFolders {
		if err := r.reconcileEnvFolder(ctx, teamFoldersClient, teamSlug, env, parentFolderNumber, tagValueName, log); err != nil {
			return err
		}
	}

	return nil
}

func (r *reconciler) reconcileEnvFolder(
	ctx context.Context,
	teamFoldersClient protoapi.GcpTeamResourcesClient,
	teamSlug, env, parentFolderNumber, tagValueName string,
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
		log.WithField("folder_id", fID).Debug("folder already stored")
	} else {
		parent := fmt.Sprintf("folders/%s", parentFolderNumber)
		fID, err = r.client.GetOrCreateFolder(ctx, teamSlug, parent)
		if err != nil {
			return fmt.Errorf("get or create GCP folder for team %q in env %q: %w", teamSlug, env, err)
		}
		log.WithField("folder_id", fID).Info("created GCP folder")

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

	return nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
