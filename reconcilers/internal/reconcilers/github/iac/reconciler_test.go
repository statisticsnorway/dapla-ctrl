package iac

import (
	"context"
	"maps"
	"testing"

	"github.com/google/go-github/v92/github"
)

type fakeTeamsClient struct {
	err         error
	permissions map[string]string
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
		want      map[string]string
	}{
		{
			name:      "managed team get push permission",
			isManaged: true,
			want: map[string]string{
				"statisticsnorway/dapla-skyinfra-developers -> statisticsnorway/play-iac": "admin",
				"statisticsnorway/dapla-platform-developers -> statisticsnorway/play-iac": "push",
				"statisticsnorway/play-developers -> statisticsnorway/play-iac":           "push",
			},
		},
		{
			name:      "self-managed team get admin permission",
			isManaged: false,
			want: map[string]string{
				"statisticsnorway/dapla-skyinfra-developers -> statisticsnorway/play-iac": "admin",
				"statisticsnorway/dapla-platform-developers -> statisticsnorway/play-iac": "push",
				"statisticsnorway/play-developers -> statisticsnorway/play-iac":           "admin",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			teams := &fakeTeamsClient{}
			r := &reconciler{org: "statisticsnorway", teams: teams}

			if err := r.reconcileGithubRepoPermissions(t.Context(), "play-iac", "play", tt.isManaged); err != nil {
				t.Fatal(err)
			}
			if !maps.Equal(teams.permissions, tt.want) {
				t.Errorf("permissions = %v, want %v", teams.permissions, tt.want)
			}
		})
	}
}
