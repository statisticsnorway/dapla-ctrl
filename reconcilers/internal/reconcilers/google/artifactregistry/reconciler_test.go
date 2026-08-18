package artifactregistry

import (
	"context"
	"reflect"
	"slices"
	"testing"

	arapiv1 "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"github.com/googleapis/gax-go/v2"
	"github.com/sirupsen/logrus"
)

type fakeArtifactRegistryClient struct {
	createErr      error
	deleteErr      error
	createRequests []*artifactregistrypb.CreateRepositoryRequest
	deleteRequests []*artifactregistrypb.DeleteRepositoryRequest
}

func (f *fakeArtifactRegistryClient) ListRepositories(context.Context, *artifactregistrypb.ListRepositoriesRequest, ...gax.CallOption) *arapiv1.RepositoryIterator {
	return nil
}

func (f *fakeArtifactRegistryClient) CreateRepository(_ context.Context, request *artifactregistrypb.CreateRepositoryRequest, _ ...gax.CallOption) (*arapiv1.CreateRepositoryOperation, error) {
	f.createRequests = append(f.createRequests, request)
	return nil, f.createErr
}

func (f *fakeArtifactRegistryClient) DeleteRepository(_ context.Context, request *artifactregistrypb.DeleteRepositoryRequest, _ ...gax.CallOption) (*arapiv1.DeleteRepositoryOperation, error) {
	f.deleteRequests = append(f.deleteRequests, request)
	return nil, f.deleteErr
}

func (f *fakeArtifactRegistryClient) GetIamPolicy(_ context.Context, request *iampb.GetIamPolicyRequest, _ ...gax.CallOption) (*iampb.Policy, error) {
	return &iampb.Policy{}, nil
}

func (f *fakeArtifactRegistryClient) SetIamPolicy(_ context.Context, request *iampb.SetIamPolicyRequest, _ ...gax.CallOption) (*iampb.Policy, error) {
	return &iampb.Policy{}, nil
}

func TestLocalAndRemoteOnly(t *testing.T) {
	tests := []struct {
		name       string
		local      []Repository
		remote     []Repository
		wantLocal  []Repository
		wantRemote []Repository
	}{
		{
			name: "returns repositories that exist only locally or remotely",
			local: []Repository{
				{Team: "play-obr", Format: "docker"},
				{Team: "play-obr", Format: "maven"},
			},
			remote: []Repository{
				{Team: "play-obr", Format: "maven"},
				{Team: "play-obr", Format: "npm"},
			},
			wantLocal:  []Repository{{Team: "play-obr", Format: "docker"}},
			wantRemote: []Repository{{Team: "play-obr", Format: "npm"}},
		},
		{
			name:       "returns empty slices for identical repositories in a different order",
			local:      []Repository{{Team: "play-obr", Format: "docker"}, {Team: "play-obr", Format: "maven"}},
			remote:     []Repository{{Team: "play-obr", Format: "maven"}, {Team: "play-obr", Format: "docker"}},
			wantLocal:  []Repository{},
			wantRemote: []Repository{},
		},
		{
			name:       "handles empty repository lists",
			local:      []Repository{},
			remote:     []Repository{},
			wantLocal:  []Repository{},
			wantRemote: []Repository{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalLocal := append([]Repository(nil), tt.local...)
			originalRemote := append([]Repository(nil), tt.remote...)

			gotLocal, gotRemote := diffRepositoriesByFormat(tt.local, tt.remote)

			if !reflect.DeepEqual(gotLocal, tt.wantLocal) {
				t.Errorf("local-only repositories = %#v, want %#v", gotLocal, tt.wantLocal)
			}
			if !reflect.DeepEqual(gotRemote, tt.wantRemote) {
				t.Errorf("remote-only repositories = %#v, want %#v", gotRemote, tt.wantRemote)
			}
			if !slices.EqualFunc(tt.local, originalLocal, func(a, b Repository) bool { return a == b }) || !slices.EqualFunc(tt.remote, originalRemote, func(a, b Repository) bool { return a == b }) {
				t.Error("localAndRemoteOnly modified its input slices")
			}
		})
	}
}

func TestRepositoryFromArtifactRegistry(t *testing.T) {
	tests := []struct {
		name string
		team string
		repo *artifactregistrypb.Repository
		want Repository
		ok   bool
	}{
		{
			name: "accepts a repository whose name matches the team and format",
			team: "play-team",
			repo: &artifactregistrypb.Repository{
				Name:      "projects/project/locations/europe-north1/repositories/play-team-docker",
				Format:    artifactregistrypb.Repository_DOCKER,
				SizeBytes: 42,
			},
			want: Repository{Team: "play-team", Format: "docker", SizeBytes: 42},
			ok:   true,
		},
		{
			name: "rejects a repository for a team whose name shares the requested prefix",
			team: "play",
			repo: &artifactregistrypb.Repository{
				Name:   "projects/project/locations/europe-north1/repositories/play-team-docker",
				Format: artifactregistrypb.Repository_DOCKER,
			},
			ok: false,
		},
		{
			name: "rejects a repository whose name and format do not agree",
			team: "play-team",
			repo: &artifactregistrypb.Repository{
				Name:   "projects/project/locations/europe-north1/repositories/play-team-maven",
				Format: artifactregistrypb.Repository_DOCKER,
			},
			ok: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := repositoryFromArtifactRegistry(tt.team, tt.repo)
			if ok != tt.ok {
				t.Fatalf("accepted repository = %t, want %t", ok, tt.ok)
			}
			if ok && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("repository = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDeleteArtifactRegistryRepositoryDryRun(t *testing.T) {
	client := &fakeArtifactRegistryClient{}
	r := &reconciler{arClient: client}

	dryRun := true
	err := r.deleteArtifactRegistryRepository(
		context.Background(),
		dryRun,
		"projects/artifact-registry-project/locations/europe-north1",
		[]Repository{{Team: "play-team", Format: "docker"}},
		logrus.New(),
	)
	if err != nil {
		t.Fatalf("dry-run deletion returned an error: %v", err)
	}
	if len(client.deleteRequests) != 0 {
		t.Errorf("delete requests = %d, want 0 in dry-run mode", len(client.deleteRequests))
	}
}

func TestConfigValidate(t *testing.T) {
	valid := Config{
		ProjectID:              "project",
		Location:               "europe-north1",
		WorkloadIdentityPoolId: "pool",
		DeleteDryRun:           "false",
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{name: "valid configuration", config: valid},
		{name: "missing project ID", config: Config{Location: valid.Location, WorkloadIdentityPoolId: valid.WorkloadIdentityPoolId, DeleteDryRun: valid.DeleteDryRun}, wantErr: true},
		{name: "missing location", config: Config{ProjectID: valid.ProjectID, WorkloadIdentityPoolId: valid.WorkloadIdentityPoolId, DeleteDryRun: valid.DeleteDryRun}, wantErr: true},
		{name: "missing workload identity pool ID", config: Config{ProjectID: valid.ProjectID, Location: valid.Location, DeleteDryRun: valid.DeleteDryRun}, wantErr: true},
		{name: "missing delete dry run", config: Config{ProjectID: valid.ProjectID, Location: valid.Location, WorkloadIdentityPoolId: valid.WorkloadIdentityPoolId}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, want error = %t", err, tt.wantErr)
			}
		})
	}
}
