package artifactregistry

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/artifactregistry/artifactregistrysql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/validate"
)

//mgo:gen model
//mgo:impl node paginated
type ArtifactRegistryRepository struct {
	// Team this repository belongs to.
	TeamSlug slug.Slug `json:"teamSlug"`
	// Format of the repository.
	Format ArtifactRegistryFormat `json:"format"`
}

func (r ArtifactRegistryRepository) ID() ident.Ident {
	return newARIdent(r.TeamSlug, r.Format.String())
}

//mgo:gen enum DOCKER PYTHON MAVEN NPM GO
type ArtifactRegistryFormat string

//mgo:gen model
//mgo:impl node paginated
type ArtifactRegistryAllowedGithubRepos struct {
	// Name of the repository, with the organization prefix.
	Name string `json:"name"`
	// Team this repository is connected to.
	TeamSlug slug.Slug `json:"teamSlug"`
}

func (r ArtifactRegistryAllowedGithubRepos) ID() ident.Ident {
	return newGHIdent(r.TeamSlug, r.Name)
}

func toGraphArtifactRegistryAllowedGithubRepos(r *artifactregistrysql.TeamArtifactRegistryGhReposAllowList) *ArtifactRegistryAllowedGithubRepos {
	return &ArtifactRegistryAllowedGithubRepos{
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
	Repository *ArtifactRegistryAllowedGithubRepos `json:"repository,omitempty"`
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

type CreateArtifactRegistryRepositoryInput struct {
	// Slug of the team
	TeamSlug slug.Slug `json:"teamSlug"`
	// Format of the repo (DOCKER, MAVEN, etc.)
	Format ArtifactRegistryFormat
}

func (i CreateArtifactRegistryRepositoryInput) Validate(ctx context.Context) error {
	verr := validate.New()

	// check if team exists
	if exists, err := db(ctx).TeamExists(ctx, i.TeamSlug); err != nil {
		return err
	} else if !exists {
		verr.Add("teamSlug", "Team with the given slug does not exists.")
	}

	if !i.Format.IsValid() {
		verr.Add("format", "Invalid or unsupported format.")
	}

	return verr.NilIfEmpty()
}

type CreateArtifactRegistryRepositoryPayload struct {
	// Repository that was created
	Repository *ArtifactRegistryRepository `json:"repository,omitempty"`
}

func toGraphArtifactRegistryRepo(r *artifactregistrysql.TeamArtifactRegistryRepository) *ArtifactRegistryRepository {
	return &ArtifactRegistryRepository{
		TeamSlug: r.TeamSlug,
		Format:   ArtifactRegistryFormat(r.Format),
	}
}
