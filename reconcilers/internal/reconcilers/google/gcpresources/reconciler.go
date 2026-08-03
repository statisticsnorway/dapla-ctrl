package gcpresources

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const reconcilerName = "google:gcpresources"

// Config holds the environment-variable-based configuration for the GCP resources reconciler.
type Config struct {
	// TagKeyNamespacedName is the namespaced name of the GCP tag key used to tag team
	// folders (e.g. "organizations/123456/tagKeys/team" or "123456/team").
	TagKeyNamespacedName string
	// EnvParentFolders maps each environment name (dev, test, prod) to the parent
	// folder number under which team folders are created (e.g. {"dev": "45678"}).
	EnvParentFolders map[string]string
}

type reconciler struct {
	client GcpClient
	cfg    Config
}

type optFunc func(*reconciler)

func WithGcpClient(c GcpClient) optFunc {
	return func(r *reconciler) {
		r.client = c
	}
}

func New(ctx context.Context, cfg Config, opts ...optFunc) (reconcilers.Reconciler, error) {
	r := &reconciler{cfg: cfg}

	for _, opt := range opts {
		opt(r)
	}

	if r.client == nil {
		c, err := NewGoogleGcpClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("create GCP client: %w", err)
		}
		r.client = c
	}

	return r, nil
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
	gcpAPI := client.GcpTeamResources()

	tagValueName, err := r.client.GetOrCreateTagValue(ctx, r.cfg.TagKeyNamespacedName, teamSlug)
	if err != nil {
		return fmt.Errorf("get or create tag value for team %q: %w", teamSlug, err)
	}
	log.WithField("tag_value_name", tagValueName).Debug("resolved GCP tag value")

	for env, parentFolderNumber := range r.cfg.EnvParentFolders {
		if err := r.reconcileEnvFolder(ctx, gcpAPI, teamSlug, env, parentFolderNumber, tagValueName, log); err != nil {
			return err
		}
	}

	return nil
}

func (r *reconciler) reconcileEnvFolder(
	ctx context.Context,
	gcpAPI protoapi.GcpTeamResourcesClient,
	teamSlug, env, parentFolderNumber, tagValueName string,
	log logrus.FieldLogger,
) error {
	log = log.WithField("env", env)

	resp, err := gcpAPI.GetTeamFolder(ctx, &protoapi.GetGcpTeamFolderRequest{
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

		if _, err := gcpAPI.UpsertTeamFolder(ctx, &protoapi.UpsertGcpTeamFolderRequest{
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
	teamSlug := daplaTeam.Slug
	gcpAPI := client.GcpTeamResources()

	for env := range r.cfg.EnvParentFolders {
		log := log.WithField("env", env)
		resp, err := gcpAPI.GetTeamFolder(ctx, &protoapi.GetGcpTeamFolderRequest{
			TeamSlug: teamSlug,
			Env:      env,
		})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				continue
			}
			return fmt.Errorf("get folder for env %q: %w", env, err)
		}

		if err := r.client.DeleteFolder(ctx, resp.Folder.FolderId); err != nil {
			return fmt.Errorf("delete GCP folder %q for env %q: %w", resp.Folder.FolderId, env, err)
		}
		log.WithField("folder_id", resp.Folder.FolderId).Info("deleted GCP folder")
	}

	if _, err := gcpAPI.DeleteTeamFolders(ctx, &protoapi.DeleteGcpTeamFoldersRequest{
		TeamSlug: teamSlug,
	}); err != nil {
		return fmt.Errorf("delete stored team folders: %w", err)
	}

	return nil
}
