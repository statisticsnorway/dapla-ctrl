package iac

import (
	"context"
	"maps"
	"testing"

	"github.com/google/go-github/v92/github"
)

type fakeTeamsClient struct {
	err         error
	existing    []*github.Team // existing repo teams, as returned by ListTeams
	permissions map[string]string
}

func (f *fakeTeamsClient) ListTeams(_ context.Context, owner, repo string, _ *github.ListOptions) ([]*github.Team, *github.Response, error) {
	return f.existing, &github.Response{}, nil
}

func githubTeamWithPermissions(slug, highestPermission string) *github.Team {
	levels := []string{"pull", "triage", "push", "maintain", "admin"}
	permissions := map[string]bool{}
	for _, l := range levels {
		permissions[l] = true
		if l == highestPermission {
			break
		}
	}

	return &github.Team{Slug: &slug, Permissions: permissions}
}

func (f *fakeTeamsClient) AddTeamRepoBySlug(_ context.Context, org, slug, owner, repo string, body *github.TeamAddTeamRepoOptions) (*github.Response, error) {
	if f.permissions == nil {
		f.permissions = map[string]string{}
	}
	f.permissions[org+"/"+slug+" -> "+owner+"/"+repo] = body.Permission
	return nil, f.err
}

func TestReconcileGithubRepoPermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		isManaged bool
		existing  []*github.Team
		want      map[string]string
	}{
		{
			name:      "no update when permissions are up to date",
			isManaged: true,
			existing: []*github.Team{
				githubTeamWithPermissions("dapla-skyinfra-developers", "admin"),
				githubTeamWithPermissions("dapla-platform-developers", "push"),
				githubTeamWithPermissions("play-developers", "push"),
				githubTeamWithPermissions("play-data-admins", "push"),
				githubTeamWithPermissions("unrelated-team", "pull"),
			},
			want: nil,
		},
		{
			name:      "managed team get push permission",
			isManaged: true,
			want: map[string]string{
				"statisticsnorway/dapla-skyinfra-developers -> statisticsnorway/play-iac": "admin",
				"statisticsnorway/dapla-platform-developers -> statisticsnorway/play-iac": "push",
				"statisticsnorway/play-developers -> statisticsnorway/play-iac":           "push",
				"statisticsnorway/play-data-admins -> statisticsnorway/play-iac":          "push",
			},
		},
		{
			name:      "self-managed team get admin permission",
			isManaged: false,
			want: map[string]string{
				"statisticsnorway/dapla-skyinfra-developers -> statisticsnorway/play-iac": "admin",
				"statisticsnorway/dapla-platform-developers -> statisticsnorway/play-iac": "push",
				"statisticsnorway/play-developers -> statisticsnorway/play-iac":           "admin",
				"statisticsnorway/play-data-admins -> statisticsnorway/play-iac":          "admin",
			},
		},
		{
			name:      "only outdated permissions are written",
			isManaged: true,
			existing: []*github.Team{
				githubTeamWithPermissions("dapla-skyinfra-developers", "admin"),
				githubTeamWithPermissions("dapla-platform-developers", "push"),
				githubTeamWithPermissions("play-developers", "admin"),
			},
			want: map[string]string{
				"statisticsnorway/play-developers -> statisticsnorway/play-iac":  "push",
				"statisticsnorway/play-data-admins -> statisticsnorway/play-iac": "push",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			teams := &fakeTeamsClient{existing: tt.existing}
			r := &reconciler{org: "statisticsnorway", teams: teams, repoTeams: teams}

			if err := r.reconcileGithubRepoPermissions(t.Context(), "play-iac", "play", tt.isManaged); err != nil {
				t.Fatal(err)
			}
			if !maps.Equal(teams.permissions, tt.want) {
				t.Errorf("permissions = %v, want %v", teams.permissions, tt.want)
			}
		})
	}
}
