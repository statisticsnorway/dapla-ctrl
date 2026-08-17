package artifactregistry

import (
	"github.com/statisticsnorway/dapla-ctrl/api/internal/artifactregistry/artifactregistrysql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
)

//mgo:gen model
//mgo:gen order NAME
//mgo:impl node paginated
type ArtifactRegistryGithubRepository struct {
	// Name of the repository, with the organization prefix.
	Name string `json:"name"`
	// Team this repository is connected to.
	TeamSlug slug.Slug `json:"teamSlug"`
}

func (r ArtifactRegistryGithubRepository) ID() ident.Ident {
	return newIdent(r.TeamSlug, r.Name)
}

func toGraphArtifactRegistryGithubRepository(r *artifactregistrysql.TeamArtifactRegistryGhReposAllowList) *ArtifactRegistryGithubRepository {
	return &ArtifactRegistryGithubRepository{
		TeamSlug: r.TeamSlug,
		Name:     r.RepositoryName,
	}
}

type GrantGithubRepoAccessToTeamArtifactRegistryInput struct {
	// Slug of the team.
	TeamSlug slug.Slug `json:"teamSlug"`
	// Name of the Github Repository which will be granted access. Without the org prefix, for instance 'repo'.
	RepositoryName string `json:"repositoryName"`
}

type GrantGithubRepoAccessToTeamArtifactRegistryPayload struct {
	// Repository that was granted access to the team artifact registry.
	Repository *ArtifactRegistryGithubRepository `json:"repository,omitempty"`
}

type RevokeGithubRepoAccessFromTeamArtifactRegistryInput struct {
	// Slug of the team.
	TeamSlug slug.Slug `json:"teamSlug"`
	// Name of the Github Repository where access should be revoked. Without the org prefix, for instance 'repo'.
	RepositoryName string `json:"repositoryName"`
}

type RevokeGithubRepoAccessFromTeamArtifactRegistryPayload struct {
	// Whether or not the repository was removed from the team.
	Success *bool `json:"success,omitempty"`
}
