package iac

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/google/go-github/v92/github"
)

type fakeRepositoriesClient struct {
	getErrs  []error
	getCalls int

	created []*github.Repository

	vulnRepos []string

	protectedBranch    string
	protectionRequests []*github.ProtectionRequest
}

func (f *fakeRepositoriesClient) Get(_ context.Context, _, repo string) (*github.Repository, *github.Response, error) {
	i := f.getCalls
	f.getCalls++
	// only return errs if there is any left in "getErrs"
	if i < len(f.getErrs) && f.getErrs[i] != nil {
		return nil, nil, f.getErrs[i]
	}
	return &github.Repository{Name: &repo}, nil, nil
}

func (f *fakeRepositoriesClient) Create(_ context.Context, _ string, repo *github.Repository) (*github.Repository, *github.Response, error) {
	f.created = append(f.created, repo)
	return repo, nil, nil
}

func (f *fakeRepositoriesClient) EnableVulnerabilityAlerts(_ context.Context, _, repo string) (*github.Response, error) {
	f.vulnRepos = append(f.vulnRepos, repo)
	return nil, nil
}

func (f *fakeRepositoriesClient) UpdateBranchProtection(_ context.Context, _, _, branch string, body *github.ProtectionRequest) (*github.Protection, *github.Response, error) {
	f.protectedBranch = branch
	f.protectionRequests = append(f.protectionRequests, body)
	return nil, nil, nil
}

func ghErr(status int) error {
	return &github.ErrorResponse{Response: &http.Response{StatusCode: status}}
}

func TestRepositoryGetOrCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		isManaged     bool
		getErrs       []error
		wantCreated   bool
		wantErr       bool
		wantGetCalls  int
		wantCreateReq bool
		wantTopic     string
	}{
		{
			name:         "existing repo is left alone",
			wantGetCalls: 1,
		},
		{
			name:          "managed repo is created",
			isManaged:     true,
			getErrs:       []error{ghErr(http.StatusNotFound)},
			wantCreated:   true,
			wantGetCalls:  2,
			wantCreateReq: true,
			wantTopic:     "managed",
		},
		{
			name:          "self-managed repo is created",
			isManaged:     false,
			getErrs:       []error{ghErr(http.StatusNotFound)},
			wantCreated:   true,
			wantGetCalls:  2,
			wantCreateReq: true,
			wantTopic:     "self-managed",
		},
		{
			name:         "non-404 github error does not create repo / fail fast",
			getErrs:      []error{ghErr(http.StatusForbidden)},
			wantErr:      true,
			wantGetCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := &fakeRepositoriesClient{getErrs: tt.getErrs}
			r := &repositoryReconciler{client: client}

			created, err := r.getOrCreate(t.Context(), "statisticsnorway", "play-iac", "play", tt.isManaged)

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if created != tt.wantCreated {
				t.Errorf("created = %v, want %v", created, tt.wantCreated)
			}
			if client.getCalls != tt.wantGetCalls {
				t.Errorf("get calls = %d, want %d", client.getCalls, tt.wantGetCalls)
			}
			if !tt.wantCreateReq {
				if len(client.created) != 0 {
					t.Errorf("expected no create call, got %d", len(client.created))
				}
				return
			}
			if len(client.created) != 1 {
				t.Fatalf("create calls = %d, want 1", len(client.created))
			}
			repo := client.created[0]
			if repo.GetName() != "play-iac" {
				t.Errorf("repo name = %q, want %q", repo.GetName(), "play-iac")
			}
			if repo.GetVisibility() != "internal" {
				t.Errorf("visibility = %q, want internal", repo.GetVisibility())
			}
			if !slices.Contains(repo.Topics, tt.wantTopic) {
				t.Errorf("topics = %v, want to contain %q", repo.Topics, tt.wantTopic)
			}
		})
	}
}

func TestWaitForRepoVisible(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		getErrs      []error
		wantErr      bool
		wantGetCalls int
	}{
		{
			name:         "visible on first attempt",
			wantGetCalls: 1,
		},
		{
			name:         "fails fast on non-retryable status",
			getErrs:      []error{ghErr(http.StatusForbidden)},
			wantErr:      true,
			wantGetCalls: 1,
		},
		{
			name:         "retries on 404, 429 and 5xx with exponential backoff",
			getErrs:      []error{ghErr(http.StatusNotFound), ghErr(http.StatusTooManyRequests), ghErr(http.StatusBadGateway)},
			wantGetCalls: 4,
		},
		{
			name: "gives up after 5 attempts",
			getErrs: []error{
				ghErr(http.StatusNotFound), ghErr(http.StatusNotFound), ghErr(http.StatusNotFound),
				ghErr(http.StatusNotFound), ghErr(http.StatusNotFound),
			},
			wantErr:      true,
			wantGetCalls: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// synctest to avoid actually waiting
			synctest.Test(t, func(t *testing.T) {
				client := &fakeRepositoriesClient{getErrs: tt.getErrs}
				r := &repositoryReconciler{client: client}

				repo, err := r.waitForRepoVisible(t.Context(), "statisticsnorway", "play-iac")

				if (err != nil) != tt.wantErr {
					t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
				}
				if err != nil && !strings.Contains(err.Error(), `"play-iac"`) {
					t.Errorf("error should name the repo, got: %v", err)
				}
				if err == nil && repo.GetName() != "play-iac" {
					t.Errorf("repo name = %q, want play-iac", repo.GetName())
				}
				if client.getCalls != tt.wantGetCalls {
					t.Errorf("Get calls = %d, want %d", client.getCalls, tt.wantGetCalls)
				}
			})
		})
	}
}

func TestWaitForRepoVisibleContext(t *testing.T) {
	t.Parallel()

	t.Run("already cancelled context makes no requests", func(t *testing.T) {
		t.Parallel()
		client := &fakeRepositoriesClient{}
		r := &repositoryReconciler{client: client}
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, err := r.waitForRepoVisible(ctx, "statisticsnorway", "hanging-iac")

		if !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want context.Canceled", err)
		}
		if client.getCalls != 0 {
			t.Errorf("Get calls = %d, want 0", client.getCalls)
		}
	})

	t.Run("reaching deadline stops retrying", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			client := &fakeRepositoriesClient{getErrs: slices.Repeat([]error{ghErr(http.StatusNotFound)}, 5)}
			r := &repositoryReconciler{client: client}
			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
			defer cancel()

			_, err := r.waitForRepoVisible(ctx, "statisticsnorway", "dead-iac")

			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("err = %v, want context.DeadlineExceeded", err)
			}
			if client.getCalls != 2 {
				t.Errorf("Get calls = %d, want 2", client.getCalls)
			}
		})
	})
}

func TestReconcileBranchProtection(t *testing.T) {
	t.Parallel()

	t.Run("self-managed repo is not protected", func(t *testing.T) {
		t.Parallel()
		client := &fakeRepositoriesClient{}
		r := &repositoryReconciler{client: client}

		if err := r.reconcileBranchProtection(t.Context(), "statisticsnorway", "play-iac", false); err != nil {
			t.Fatal(err)
		}
		if len(client.protectionRequests) != 0 {
			t.Errorf("expected no UpdateBranchProtection call, got %d", len(client.protectionRequests))
		}
	})

	t.Run("managed repo protects default branch and requires atlantis apply", func(t *testing.T) {
		t.Parallel()
		client := &fakeRepositoriesClient{}
		r := &repositoryReconciler{client: client}

		if err := r.reconcileBranchProtection(t.Context(), "statisticsnorway", "play-iac", true); err != nil {
			t.Fatal(err)
		}
		if client.protectedBranch != defaultBranch {
			t.Fatalf("protected branches = %v, want [%s]", client.protectedBranch, defaultBranch)
		}
		req := client.protectionRequests[0]
		if !req.EnforceAdmins {
			t.Error("EnforceAdmins = false, want true")
		}
		if reviews := req.RequiredPullRequestReviews; reviews == nil || !reviews.RequireCodeOwnerReviews || reviews.RequiredApprovingReviewCount != 1 {
			t.Errorf("unexpected required reviews: %+v", reviews)
		}
		if checks := req.RequiredStatusChecks; checks == nil || checks.Contexts == nil || !slices.Equal(*checks.Contexts, []string{"atlantis/apply"}) {
			t.Errorf("unexpected required status checks: %+v", checks)
		}
	})
}

func TestReconcileVulnerabilityAlerts(t *testing.T) {
	t.Parallel()

	client := &fakeRepositoriesClient{}
	r := &repositoryReconciler{client: client}
	if err := r.reconcileVulnerabilityAlerts(t.Context(), "statisticsnorway", "play-iac"); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(client.vulnRepos, []string{"play-iac"}) {
		t.Errorf("vulnerability alerts enabled for %v, want [play-iac]", client.vulnRepos)
	}
}
