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

func toGraphArtifactRegistryGithubRepository(r *artifactregistrysql.TeamArtifactRegistryGithubRepository) *ArtifactRegistryGithubRepository {
	return &ArtifactRegistryGithubRepository{
		TeamSlug: r.TeamSlug,
		Name:     r.GithubRepository,
	}
}

type AddArtifactRegistryGithubRepositoryToTeamInput struct {
	// Slug of the team to add the repository to.
	TeamSlug slug.Slug `json:"teamSlug"`
	// Name of the repository, without the org prefix, for instance 'repo'.
	RepositoryName string `json:"repositoryName"`
}

type AddArtifactRegistryGithubRepositoryToTeamPayload struct {
	// Repository that was added to the team.
	Repository *ArtifactRegistryGithubRepository `json:"repository,omitempty"`
}

type RemoveArtifactRegistryGithubRepositoryFromTeamInput struct {
	// Slug of the team to remove the repository from.
	TeamSlug slug.Slug `json:"teamSlug"`
	// Name of the repository, without the org prefix, for instance 'repo'.
	RepositoryName string `json:"repositoryName"`
}

type RemoveArtifactRegistryGithubRepositoryFromTeamPayload struct {
	// Whether or not the repository was removed from the team.
	Success *bool `json:"success,omitempty"`
}
