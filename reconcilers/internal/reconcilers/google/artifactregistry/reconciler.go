package artifactregistry

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	arapiv1 "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"github.com/googleapis/gax-go/v2"
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

	configProjectIDKey              = "project_id"
	configLocationKey               = "location"
	configWorkloadIdentityPoolIdKey = "workload_identity_pool_id"
	configDeleteDryRunKey           = "delete_dry_run"

	saNamePrefix = "gh-actions-"
	wiUserRole   = "roles/iam.workloadIdentityUser"
	arWriterRole = "roles/artifactregistry.writer"
)

type ArtifactRegistryClient interface {
	ListRepositories(ctx context.Context, req *artifactregistrypb.ListRepositoriesRequest, opts ...gax.CallOption) *arapiv1.RepositoryIterator
	CreateRepository(ctx context.Context, req *artifactregistrypb.CreateRepositoryRequest, opts ...gax.CallOption) (*arapiv1.CreateRepositoryOperation, error)
	DeleteRepository(ctx context.Context, req *artifactregistrypb.DeleteRepositoryRequest, opts ...gax.CallOption) (*arapiv1.DeleteRepositoryOperation, error)
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
	SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
}
type ServiceAccounts interface{}

type reconciler struct {
	config          Config
	arClient        ArtifactRegistryClient
	serviceAccounts *serviceaccounts.Client
}

type Config struct {
	ProjectID              string
	Location               string
	WorkloadIdentityPoolId string
	DeleteDryRun           string
}

func (c Config) validate() error {
	if c.ProjectID == "" || c.Location == "" || c.WorkloadIdentityPoolId == "" || c.DeleteDryRun == "" {
		return errors.New("all configuration parameters for artifact registry reconciler must be set")
	}
	return nil
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
				Key:         configProjectIDKey,
				DisplayName: "Artifact Registry project ID",
				Description: "Project id of the project the AR repos will be placed. E.g. `my-project-id-22`",
				Secret:      false,
			},
			{
				Key:         configLocationKey,
				DisplayName: "Artifact Registry location",
				Description: "The location where AR repos will be created. E.g. `europe-north1`",
				Secret:      false,
			},
			{
				Key:         configWorkloadIdentityPoolIdKey,
				DisplayName: "Workload Identity pool id",
				Description: "The ID of the Workload Identity Pool that will be granted access to AR repositories. E.g. 'gh-actions'.",
				Secret:      false,
			},
			{
				Key:         configDeleteDryRunKey,
				DisplayName: "Artifact Registry repostiory deletion dry run",
				Description: "Should deletion of AR repositories be dry run. 'true' will result in dry run.",
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
	log.WithField("sa", sa).Debug("reconcilled service account")

	githubReposAllowlist, err := r.getLocalGithubRepos(ctx, client, daplaTeam.Slug)
	if err != nil {
		return err
	}

	// Give workload identity role binding to the specifiec GitHub repos
	if err := r.reconcileGithubRepoAllowlist(ctx, sa.Name, githubReposAllowlist); err != nil {
		return err
	}
	log.WithField("sa", sa).Debug("reconciled github allow list")

	parent := fmt.Sprintf("projects/%s/locations/%s", r.config.ProjectID, r.config.Location)
	log = log.WithField("parent", parent)

	remoteRepos, err := r.getRemoteArtifactRegistryRepositories(ctx, parent, daplaTeam.Slug)
	if err != nil {
		return err
	}
	log.Debugf("fetched %s remote repos from artifact registry", len(remoteRepos))

	localRepos, err := r.getLocalArtifactRegistryRepositories(ctx, client, daplaTeam.Slug)
	if err != nil {
		return err
	}
	log.Debugf("fetched %s local repos", len(remoteRepos))

	localOnly, remoteOnly := diffRepositoriesByFormat(localRepos, remoteRepos)

	createErrs := r.createArtifactRegistryRepositories(ctx, sa.Email, parent, localOnly, log)

	deleteDryRun, err := strconv.ParseBool(r.config.DeleteDryRun)
	if err != nil {
		log.Error("could not parse DeleteDryRun config - defaulting to dry run = true")
		deleteDryRun = true
	}
	deleteErrs := r.deleteArtifactRegistryRepository(ctx, deleteDryRun, parent, remoteOnly, log)

	// We want to try to reconcile all resources rather than fail fast. Thus joining errors later.
	return errors.Join(createErrs, deleteErrs)
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
		r.config.ProjectID,
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

func (r *reconciler) getLocalArtifactRegistryRepositories(ctx context.Context, client *apiclient.APIClient, team string) ([]Repository, error) {
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

func (r *reconciler) getRemoteArtifactRegistryRepositories(ctx context.Context, parent, team string) ([]Repository, error) {
	resp := r.arClient.ListRepositories(ctx, &artifactregistrypb.ListRepositoriesRequest{
		Parent: parent,
		Filter: "name=" + fmt.Sprintf("\"%s/repositories/%s-*\"", parent, team),
	})

	var repos []Repository
	for repo, err := range resp.All() {
		if err != nil {
			return nil, err
		}

		parsedRepo, belongsToTeam := repositoryFromArtifactRegistry(team, repo)
		if !belongsToTeam {
			continue
		}
		repos = append(repos, parsedRepo)
	}

	return repos, nil
}

func repositoryFromArtifactRegistry(team string, repo *artifactregistrypb.Repository) (Repository, bool) {
	format := strings.ToLower(repo.GetFormat().String())
	expectedRepoName := team + "-" + format

	repoNameParts := strings.Split(repo.GetName(), "/")
	repoName := repoNameParts[len(repoNameParts)-1]

	// Filter out e.g. play-team-b when reconciling play-team.
	if repoName != expectedRepoName {
		return Repository{}, false
	}

	return Repository{
		Team:      team,
		Format:    format,
		SizeBytes: repo.GetSizeBytes(),
	}, true
}

func diffRepositoriesByFormat(localRepos, remoteRepos []Repository) (localOnly []Repository, remoteOnly []Repository) {
	localOnly = slices.DeleteFunc(slices.Clone(localRepos), func(local Repository) bool {
		return slices.ContainsFunc(remoteRepos, func(remote Repository) bool {
			return strings.EqualFold(local.Format, remote.Format)
		})
	})

	remoteOnly = slices.DeleteFunc(slices.Clone(remoteRepos), func(remote Repository) bool {
		return slices.ContainsFunc(localRepos, func(local Repository) bool {
			return strings.EqualFold(local.Format, remote.Format)
		})
	})

	return localOnly, remoteOnly
}

func (r *reconciler) createArtifactRegistryRepositories(ctx context.Context, saEmail, parent string, repos []Repository, log logrus.FieldLogger) error {
	allErrors := make([]error, 0)
	for _, repo := range repos {
		repoId := strings.ToLower(repo.Team + "-" + repo.Format)
		format := artifactregistrypb.Repository_Format_value[strings.ToUpper(repo.Format)]
		log.WithField("repo", repoId).Info("Create artifact registry repository and assigning IAM")
		op, err := r.arClient.CreateRepository(ctx, &artifactregistrypb.CreateRepositoryRequest{
			Parent:       parent,
			RepositoryId: repoId,
			Repository: &artifactregistrypb.Repository{
				Format: artifactregistrypb.Repository_Format(format),
			},
		})
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}

		arRepo, err := op.Wait(ctx)
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}

		if err := assignArtifactRegistryWriterRole(ctx, r.arClient, arRepo.Name, saEmail); err != nil {
			allErrors = append(allErrors, err)
		}
	}
	return errors.Join(allErrors...)
}

func assignArtifactRegistryWriterRole(ctx context.Context, client ArtifactRegistryClient, repositoryName, saEmail string) error {
	policy, err := client.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: repositoryName,
	})
	if err != nil {
		return err
	}

	policy.Bindings = append(policy.Bindings, &iampb.Binding{
		Members: []string{"serviceAccount:" + saEmail},
		Role:    arWriterRole,
	})

	_, err = client.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: repositoryName,
		Policy:   policy,
	})
	return err
}

func (r *reconciler) deleteArtifactRegistryRepository(ctx context.Context, dryRun bool, parent string, repos []Repository, log logrus.FieldLogger) error {
	allErrors := make([]error, 0)
	for _, repo := range repos {
		repoId := strings.ToLower(repo.Team + "-" + repo.Format)
		log.WithFields(logrus.Fields{
			"repo":   repoId,
			"dryrun": dryRun,
		}).Info("delete artifact registry repository")
		if dryRun {
			continue
		}
		// Will also delete IAM related to repository
		op, err := r.arClient.DeleteRepository(ctx, &artifactregistrypb.DeleteRepositoryRequest{
			Name: parent + "/repositories/" + repoId,
		})
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}
		err = op.Wait(ctx)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}
	return errors.Join(allErrors...)
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
		case configProjectIDKey:
			gac.ProjectID = c.Value
		case configLocationKey:
			gac.Location = c.Value
		case configDeleteDryRunKey:
			gac.DeleteDryRun = c.Value
		case configWorkloadIdentityPoolIdKey:
			gac.WorkloadIdentityPoolId = c.Value
		default:
			return fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if err := gac.validate(); err != nil {
		return fmt.Errorf("validate reconciler config: %w", err)
	}

	r.config = gac
	return nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
