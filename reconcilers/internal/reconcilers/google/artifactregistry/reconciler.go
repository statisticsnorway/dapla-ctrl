package artifactregistry

import (
	"context"
	"fmt"
	"slices"
	"strings"

	arapiv1 "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/iterator"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google/serviceaccounts"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers"
	"google.golang.org/api/iam/v1"
)

const (
	reconcilerName = "google:artifactregistry"

	configParentKey = "parent"

	saNamePrefix = "gh-actions-"
	wiUserRole   = "roles/iam.workloadIdentityUser"
)

type reconciler struct {
	config          Config
	arClient        *arapiv1.Client
	serviceAccounts *serviceaccounts.Client
}

type Config struct {
	// Project ID
	ProjectId string

	// Location
	Location string

	WorkloadIdentityPoolId string
}

type Repository struct {
	Team      string
	Format    string
	SizeBytes int64
}

type optFunc func(*reconciler)

func New(ctx context.Context, serviceAccounts *serviceaccounts.Client, opts ...optFunc) (reconcilers.Reconciler, error) {
	r := new(reconciler)

	r.serviceAccounts = serviceAccounts

	for _, opt := range opts {
		opt(r)
	}

	if r.arClient == nil {
		client, err := arapiv1.NewClient(ctx)
		if err != nil {
			return nil, err
		}
		r.arClient = client
	}

	return r, nil
}

func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "Artifact Registry",
		Description: "Create Artifact Registry repositories for Dapla teams.",
		MemberAware: true,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         configParentKey,
				DisplayName: "Artifact Registry parent ID",
				Description: "Parent string for AR repos. E.q. `projects/project-id/locations/europe-north1",
				Secret:      false,
			},
		},
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	if err := r.updateConfig(ctx, client); err != nil {
		return fmt.Errorf("error getting reconciler config: %w", err)
	}

	// Ensure GH Actions SA exists
	sa, err := r.reconcileServiceAccount(ctx, daplaTeam.Slug)
	if err != nil {
		return err
	}

	// Give workload identity role binding to the specifiec GitHub repos
	githubReposAllowlist, err := r.getLocalGithubRepos(ctx, client, daplaTeam.Slug)
	if err != nil {
		return err
	}

	if err := r.reconcileGithubRepoAllowlist(ctx, sa.Name, githubReposAllowlist); err != nil {
		return err
	}

	remoteRepos, err := r.getRemoteRepositories(ctx, daplaTeam.Slug)
	if err != nil {
		return err
	}

	localRepos, err := r.getLocalRepositories(ctx, client, daplaTeam.Slug)
	if err != nil {
		return err
	}

	localOnly, remoteOnly := localAndRemoteOnly(localRepos, remoteRepos)

	return nil
}

func (r *reconciler) getLocalGithubRepos(ctx context.Context, client *apiclient.APIClient, team string) ([]string, error) {
	repoIt := iterator.New(ctx, 100, func(limit, offset int64) (*protoapi.GetArtifactRegistryGithubAllowlistResponse, error) {
		return client.ArtifactRegistry().GetArtifactRegistryGithubAllowlist(ctx, &protoapi.GetArtifactRegistryGithubAllowlistRequest{
			TeamSlug: team,
			Limit:    limit,
			Offset:   offset,
		})
	})

	var repos []string
	for repoIt.Next() {
		repos = append(repos, repoIt.Value())
	}

	if err := repoIt.Err(); err != nil {
		return nil, err
	}

	return repos, nil
}

func (r *reconciler) reconcileServiceAccount(ctx context.Context, team string) (*iam.ServiceAccount, error) {
	saName := "gh-actions-" + team
	saDescription := "SA used by GitHub Actions to push images to " + team + " AR repo"
	sa, err := r.serviceAccounts.GetOrCreate(ctx,
		saName,
		saDescription,
		r.config.ProjectId,
	)
	if err != nil {
		return nil, err
	}

	if sa.Description != saDescription {
		if err := r.serviceAccounts.UpdateDescription(ctx,
			sa.Name,
			saDescription,
		); err != nil {
			return nil, err
		}
	}

	return sa, nil
}

func (r *reconciler) reconcileGithubRepoAllowlist(ctx context.Context, saName string, repos []string) error {
	principalSetPrefix := "principalSet://iam.googleapis.com/" + r.config.WorkloadIdentityPoolId + "/attribute.repository/statisticsnorway/"
	return r.serviceAccounts.EnsureRoleBindingFunc(ctx, saName, wiUserRole, func(b *iam.Binding) bool {
		// Keep a copy of the current members of the binding
		currentMembers := slices.Clone(b.Members)
		// Delete the GitHub repo members
		b.Members = slices.DeleteFunc(b.Members, func(m string) bool {
			return strings.HasPrefix(m, principalSetPrefix)
		})
		// Add the GitHub repo members we want
		for _, repo := range repos {
			b.Members = append(b.Members, principalSetPrefix+repo)
		}
		// If the length has changed, it must mean one or more repos were addedd/removed
		if len(currentMembers) != len(b.Members) {
			return true
		}
		// If the length is equal we need to check if any of the repos names have changed
		slices.Sort(currentMembers)
		slices.Sort(b.Members)
		return !slices.Equal(currentMembers, b.Members)
	})
}

func (r *reconciler) getLocalRepositories(ctx context.Context, client *apiclient.APIClient, team string) ([]Repository, error) {
	reposIt := iterator.New(ctx, 100, func(limit, offset int64) (*protoapi.ListArtifactRegistryReposForTeamResponse, error) {
		return client.ArtifactRegistry().ListArtifactRegistryReposForTeam(ctx, &protoapi.ListArtifactRegistryReposForTeamRequest{
			TeamSlug: team,
			Limit:    limit,
			Offset:   offset,
		})
	})

	var repos []Repository
	for reposIt.Next() {
		repos = append(repos, Repository{
			Team:      team,
			Format:    reposIt.Value().Format,
			SizeBytes: reposIt.Value().SizeBytes,
		})
	}

	if reposIt.Err() != nil {
		return nil, reposIt.Err()
	}

	return repos, nil

}

func (r *reconciler) getRemoteRepositories(ctx context.Context, team string) ([]Repository, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", r.config.ProjectId, r.config.Location)
	resp := r.arClient.ListRepositories(ctx, &artifactregistrypb.ListRepositoriesRequest{
		Parent: parent,
		Filter: fmt.Sprintf("%s/repositories/%s-*", parent, team),
	})

	var repos []Repository
	for repo, err := range resp.All() {
		if err != nil {
			return nil, err
		}

		repoNameParts := strings.Split(repo.Name, "/")
		repoName := repoNameParts[len(repoNameParts)-1]
		format := strings.TrimPrefix(repoName, team+"-")
		// Filter out e.g. play-team-b from play-team
		if strings.Contains(format, "-") {
			continue
		}

		repos = append(repos, Repository{
			Team:      team,
			Format:    format,
			SizeBytes: repo.SizeBytes,
		})
	}

	return repos, nil
}

func localAndRemoteOnly(localRepos, remoteRepos []Repository) (localOnly []Repository, remoteOnly []Repository) {
	localOnly = slices.DeleteFunc(slices.Clone(localRepos), func(local Repository) bool {
		return slices.ContainsFunc(remoteRepos, func(remote Repository) bool {
			return local.Format == remote.Format
		})
	})

	remoteOnly = slices.DeleteFunc(slices.Clone(remoteRepos), func(remote Repository) bool {
		return slices.ContainsFunc(localRepos, func(local Repository) bool {
			return local.Format == remote.Format
		})
	})

	return localOnly, remoteOnly
}

func (r *reconciler) updateConfig(ctx context.Context, client *apiclient.APIClient) error {
	config, err := client.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return fmt.Errorf("get reconciler config: %w", err)
	}

	gac := Config{}

	for _, c := range config.Nodes {
		switch c.Key {
		case configParentKey:
			gac.Parent = c.Value
		default:
			return fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	r.config = gac
	return nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
