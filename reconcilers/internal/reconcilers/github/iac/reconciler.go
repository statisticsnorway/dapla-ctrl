package iac

import (
	"context"
	"errors"
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
	configRepoPrefixKey    = "repoPrefix"
)

type ghClient struct {
	Repositories *github.RepositoriesService
	PullRequests *github.PullRequestsService
	Teams        *github.TeamsService
}

type reconciler struct {
	teamAllowlist []string
	repoPrefix    string
	org           string
	ghClient      *ghClient
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

	r.ghClient = &ghClient{
		Repositories: client.Repositories,
		PullRequests: client.PullRequests,
		Teams:        client.Teams,
	}

	return r, nil
}

func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "GitHub Team",
		Description: "Create GitHub teams and sync them with Entra ID",
		MemberAware: true,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         configTeamAllowlistKey,
				DisplayName: "Team whitelist",
				Description: "Comma-separated list of teams to create iac repos for. Empty list means create for all teams.",
				Secret:      false,
			},
			{
				Key:         configRepoPrefixKey,
				DisplayName: "Repo prefix",
				Description: "Prefix to add to GitHub iac repos",
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

	repoName := r.repoPrefix + daplaTeam.Slug + "-iac"
	repo, created, err := r.getOrCreateRepository(ctx, r.org, repoName, daplaTeam)
	if err != nil {
		return err
	}

	if created {
		// get atlantis name
		atlantisUrl := "https://${var.atlantis_name}.atlantis.ssb.no/events" // TODO:
		secret := ""                                                         // TODO, fetch from api
		r.ghClient.Repositories.CreateHook(ctx, r.org, *repo.Name, &github.Hook{
			Config: &github.HookConfig{
				ContentType: new("json"),
				URL:         &atlantisUrl,
				Secret:      &secret,
			},
			Events: []string{
				"check_run",
				"create",
				"delete",
				"issue_comment",
				"issues",
				"pull_request",
				"pull_request_review",
				"pull_request_review_comment",
				"push",
			},
			Active: new(true),
		})

		// TODO: Commit render projects
		// TODO: create branch and commit iac repo tempalte
	}

	_, err = r.ghClient.Repositories.EnableVulnerabilityAlerts(ctx, r.org, *repo.Name)
	if err != nil {
		return err
	}

	// Give access to github repoL:
	// "dapla-skyinfra-developers" = "admin"
	// "dapla-platform-developers" = "push"
	// The team it self = push if managed, admin if not managed.

	ghTeamPermission := "push"
	if !daplaTeam.IsManaged {
		ghTeamPermission = "admin"
	}

	for slug, permission := range map[string]string{
		"dapla-skyinfra-developers":    "admin",
		"dapla-platform-developers":    "push",
		daplaTeam.Slug + "-developers": ghTeamPermission,
	} {
		_, err = r.ghClient.Teams.AddTeamRepoBySlug(ctx, r.org, slug, r.org, *repo.Name, &github.TeamAddTeamRepoOptions{
			Permission: permission,
		})
		if err != nil {
			return err
		}
	}

	// enforce branch protections rule
	if daplaTeam.IsManaged {
		_, _, err = r.ghClient.Repositories.UpdateBranchProtection(ctx, r.org, *repo.Name, "main", &github.ProtectionRequest{
			EnforceAdmins: true,
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				DismissStaleReviews:          true,
				RequireCodeOwnerReviews:      true,
				RequiredApprovingReviewCount: 1,
			},
			RequiredStatusChecks: &github.RequiredStatusChecks{
				Contexts: &[]string{"atlantis/apply"},
				Strict:   true,
			},
		})
		if err != nil {
			return err
		}
	}

	// TODO update webhook secret if it has changed (check obfuscated secret or some other field).

	return nil
}

func (r *reconciler) getOrCreateRepository(ctx context.Context, owner, repoName string, daplaTeam *protoapi.Team) (*github.Repository, bool, error) {
	repo, _, err := r.ghClient.Repositories.Get(ctx, owner, repoName)
	if err == nil {
		return repo, false, nil
	}

	githubError, ok := errors.AsType[*github.ErrorResponse](err)
	if !ok || githubError.Response.StatusCode != http.StatusNotFound {
		return nil, false, err
	}

	description := "IaC repo for " + daplaTeam.GetSlug()
	managedTopic := "managed"
	if !daplaTeam.IsManaged {
		managedTopic = "self-managed"
	}
	_, _, err = r.ghClient.Repositories.Create(ctx, owner, &github.Repository{
		Name:        &repoName,
		Description: &description,
		Visibility:  new("internal"),
		HasIssues:   new(true),
		Topics:      []string{"terraform", "dapla-team", "kuben", managedTopic},
		AutoInit:    new(true),
	})
	if err != nil {
		return nil, false, err
	}

	// Let GitHub finish creating the repo, such that we avoid race condition later on
	repo, err = r.waitForRepoVisible(ctx, owner, repoName)
	if err != nil {
		return nil, false, err
	}

	return repo, true, nil
}

func (r *reconciler) waitForRepoVisible(ctx context.Context, owner, repoName string) (*github.Repository, error) {
	// 5 apptemts with exponential backoff caped at 4 seconds -> total potential 15 seconds hold
	maxApptempts := 5
	backoff := 1 * time.Second
	waited := time.Duration(0)

	for attempt := 1; attempt <= maxApptempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context done while waiting for github repo %q to become visible after creation: %w", repoName, err)
		}

		repo, _, err := r.ghClient.Repositories.Get(ctx, owner, repoName)
		if err == nil {
			return repo, nil
		}

		if githubError, ok := errors.AsType[*github.ErrorResponse](err); ok {
			status := githubError.Response.StatusCode
			shouldRetry := status == http.StatusNotFound ||
				status == http.StatusTooManyRequests ||
				status >= http.StatusInternalServerError
			if !shouldRetry {
				return nil, fmt.Errorf("failed to verify repository %q after creation: %w", repoName, err)
			}
		}

		if attempt == maxApptempts {
			break
		}

		// backoff
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context done while waiting for repository %q to become visible after creation: %w", repoName, ctx.Err())
		case <-time.After(backoff):
		}

		waited += backoff
		backoff *= 2
		if backoff > 4 * time.Second {
			backoff = 4 * time.Second
		}
	}

	return nil, fmt.Errorf("github repo %q was created but could not be verified as visible after %d attempts (waited ca %s )", repoName, maxApptempts, waited)
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
		case configRepoPrefixKey:
			if r.repoPrefix != c.Value {
				r.repoPrefix = c.Value
			}
		default:
			return fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	return nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	// TODO: Archive github-repo
	return nil
}
