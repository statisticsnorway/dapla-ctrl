package iac

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v92/github"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
)

const (
	reconcilerName = "github:iac"

	configTeamAllowlistKey = "teamAllowlist"
)

type pullRequestsClient interface {
	Create(ctx context.Context, owner, repo string, body github.CreatePullRequest) (*github.PullRequest, *github.Response, error)
}

type teamsClient interface {
	AddTeamRepoBySlug(ctx context.Context, org, slug, owner, repo string, body *github.TeamAddTeamRepoOptions) (*github.Response, error)
}

type reconciler struct {
	teamAllowlist []string
	org           string

	pullRequests pullRequestsClient
	teams        teamsClient
	repository   *repositoryReconciler
	repoContent  *repoContentReconciler
	webhooks     *webhookReconciler
}

type optFunc func(*reconciler)

func New(ctx context.Context, org string, appId, installationId int64, privateKeyFile string, opts ...optFunc) (*reconciler, error) {
	r := &reconciler{
		org: org,
	}

	for _, opt := range opts {
		opt(r)
	}

	tr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appId, installationId, privateKeyFile)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	client, err := github.NewClient(github.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	r.pullRequests = client.PullRequests
	r.teams = client.Teams
	r.repository = &repositoryReconciler{client: client.Repositories}
	r.repoContent = &repoContentReconciler{client: client.Git}
	r.webhooks = &webhookReconciler{client: client.Repositories}

	return r, nil
}

func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "GitHub iac Repo",
		Description: "Create GitHub dapla team iac repositories",
		MemberAware: true,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         configTeamAllowlistKey,
				DisplayName: "Team whitelist",
				Description: "Comma-separated list of teams to create iac repos for. Empty list means create for all teams.",
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

	// empty allowlist allows all
	if len(r.teamAllowlist) > 0 && !slices.Contains(r.teamAllowlist, daplaTeam.Slug) {
		return nil
	}
	teamName := daplaTeam.Slug
	isManaged := daplaTeam.IsManaged

	owner := r.org
	repo := teamName + "-iac"
	created, err := r.repository.getOrCreate(ctx, owner, repo, teamName, isManaged)
	if err != nil {
		return err
	}

	if created {
		// Note: If this step fails we must init the repo our self
		if err := r.repoContent.initIacRepoContent(ctx, owner, repo, teamName, isManaged); err != nil {
			return fmt.Errorf("initialize content of repo %s: %w", repo, err)
		}
		err = r.createInitialPR(ctx, repo)
		if err != nil {
			return fmt.Errorf("create pull request of repo %s: %w", repo, err)
		}
	}

	err = r.reconcileAtlantisWebhook(ctx, client, repo, teamName)
	if err != nil {
		return err
	}

	err = r.repository.reconcileVulnerabilityAlerts(ctx, owner, repo)
	if err != nil {
		return err
	}

	err = r.reconcileGithubRepoPermissions(ctx, repo, teamName, isManaged)
	if err != nil {
		return err
	}

	err = r.repository.reconcileBranchProtection(ctx, owner, repo, isManaged)
	if err != nil {
		return err
	}

	return nil
}

func (r *reconciler) reconcileAtlantisWebhook(ctx context.Context, client *apiclient.APIClient, repo, teamName string) error {
	atlantisUrl, err := getAtlantisUrl(ctx, client.Atlantis(), teamName)
	if err != nil {
		return err
	}
	atlantisSecret, err := getAtlantisWebhookSecret(ctx, client.Atlantis(), teamName)
	if err != nil {
		return err
	}

	return r.webhooks.upsertAtlantisWebhook(ctx, r.org, repo, atlantisUrl, atlantisSecret)
}

func (r *reconciler) reconcileGithubRepoPermissions(ctx context.Context, repoName, teamName string, isManaged bool) error {
	teamPermission := "push"
	if !isManaged {
		teamPermission = "admin"
	}

	for team, permission := range map[string]string{
		"dapla-skyinfra-developers": "admin",
		"dapla-platform-developers": "push",
		teamName + "-developers":    teamPermission,
	} {
		_, err := r.teams.AddTeamRepoBySlug(ctx, r.org, team, r.org, repoName, &github.TeamAddTeamRepoOptions{
			Permission: permission,
		})
		if err != nil {
			return fmt.Errorf("failed to add team repo permission, repo: %s, github team: %s, err: %w", repoName, team, err)
		}
	}
	return nil
}

func (r *reconciler) createInitialPR(ctx context.Context, repoName string) error {
	_, _, err := r.pullRequests.Create(ctx, r.org, repoName, github.CreatePullRequest{
		Title: new("Initial setup"),
		Head:  initBranch,
		Base:  defaultBranch,
		Body:  new("Create base repo structure and initial IaC code"),
	})
	return err
}

func (r *reconciler) updateConfig(ctx context.Context, client *apiclient.APIClient) error {
	config, err := client.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return fmt.Errorf("get reconciler config: %w", err)
	}

	for _, c := range config.Nodes {
		switch c.Key {
		case configTeamAllowlistKey:
			if c.Value == "" {
				r.teamAllowlist = nil
				break
			}
			whitelist := strings.Split(c.Value, ",")
			if !slices.Equal(r.teamAllowlist, whitelist) {
				r.teamAllowlist = whitelist
			}
		default:
			return fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	return nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	// Delete is not implemented in any reconcilers yet.
	return nil
}
