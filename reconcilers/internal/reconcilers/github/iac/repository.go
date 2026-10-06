package iac

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v92/github"
)

type repositoriesReconcilerClient interface {
	Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
	Create(ctx context.Context, org string, repo *github.Repository) (*github.Repository, *github.Response, error)
	EnableVulnerabilityAlerts(ctx context.Context, owner, repository string) (*github.Response, error)
	UpdateBranchProtection(ctx context.Context, owner, repo, branch string, body *github.ProtectionRequest) (*github.Protection, *github.Response, error)
}

type repositoryReconciler struct {
	client repositoriesReconcilerClient
}

func (r *repositoryReconciler) getOrCreate(ctx context.Context, owner, repo string, daplaTeam string, isManaged bool) (bool, error) {
	_, _, err := r.client.Get(ctx, owner, repo)
	if err == nil {
		return false, nil
	}

	githubError, ok := errors.AsType[*github.ErrorResponse](err)
	if !ok || githubError.Response.StatusCode != http.StatusNotFound {
		return false, err
	}

	description := "IaC repo for " + daplaTeam
	managedTopic := "managed"
	if !isManaged {
		managedTopic = "self-managed"
	}
	_, _, err = r.client.Create(ctx, owner, &github.Repository{
		Name:        &repo,
		Description: &description,
		Visibility:  new("internal"),
		HasIssues:   new(true),
		Topics:      []string{"terraform", "dapla-team", "kuben", managedTopic},
		AutoInit:    new(true),
	})
	if err != nil {
		return false, err
	}

	// Let GitHub finish creating the repo, such that we avoid race condition later on
	_, err = r.waitForRepoVisible(ctx, owner, repo)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *repositoryReconciler) reconcileVulnerabilityAlerts(ctx context.Context, owner, repo string) error {
	_, err := r.client.EnableVulnerabilityAlerts(ctx, owner, repo)
	return err
}

func (r *repositoryReconciler) waitForRepoVisible(ctx context.Context, owner, repo string) (*github.Repository, error) {
	// 5 attempt with exponential backoff to max 4 seconds each -> total potential 11 seconds hold
	maxAttempts := 5
	backoff := 1 * time.Second
	waited := time.Duration(0)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context done while waiting for github repo %q to become visible after creation: %w", repo, err)
		}

		ghRepo, _, err := r.client.Get(ctx, owner, repo)
		if err == nil {
			return ghRepo, nil
		}

		if githubError, ok := errors.AsType[*github.ErrorResponse](err); ok {
			status := githubError.Response.StatusCode
			shouldRetry := status == http.StatusNotFound ||
				status == http.StatusTooManyRequests ||
				status >= http.StatusInternalServerError
			if !shouldRetry {
				return nil, fmt.Errorf("failed to verify repository %q after creation: %w", repo, err)
			}
		}

		if attempt == maxAttempts {
			break
		}

		// backoff
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context done while waiting for repository %q to become visible after creation: %w", repo, ctx.Err())
		case <-time.After(backoff):
		}

		waited += backoff
		backoff *= 2
		if backoff > 4*time.Second {
			backoff = 4 * time.Second
		}
	}

	return nil, fmt.Errorf("github repo %q was created but could not be verified as visible after %d attempts (waited ca %s )", repo, maxAttempts, waited)
}

func (r *repositoryReconciler) reconcileBranchProtection(ctx context.Context, owner, repo string, isManaged bool) error {
	if !isManaged {
		return nil
	}

	_, _, err := r.client.UpdateBranchProtection(ctx, owner, repo, defaultBranch, &github.ProtectionRequest{
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
	return err
}
