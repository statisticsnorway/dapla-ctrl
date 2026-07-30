package artifactregistry

import (
	"github.com/statisticsnorway/dapla-ctrl/api/internal/artifactregistry/artifactregistrysql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
)

//mgo:gen model
//mgo:gen order NAME
//mgo:impl node paginated
type Repository struct {
	Format    string    `json:"format"`
	SizeBytes int64     `json:"sizeBytes"`
	TeamSlug  slug.Slug `json:"-"`
}

func (r Repository) ID() ident.Ident {
	return newIdent(r.TeamSlug, r.Format)
}

func toGraphRepository(r *artifactregistrysql.TeamArtifactRegistryRepository) *Repository {
	return &Repository{
		TeamSlug:  r.TeamSlug,
		Format:    r.Format,
		SizeBytes: r.SizeBytes,
	}
}

type AddRepositoryToTeamInput struct {
	TeamSlug       slug.Slug `json:"teamSlug"`
	RepositoryName string    `json:"repositoryName"`
}

type AddRepositoryToTeamPayload struct {
	Repository *Repository `json:"repository"`
}

type RemoveRepositoryFromTeamInput struct {
	TeamSlug       slug.Slug `json:"teamSlug"`
	RepositoryName string    `json:"repositoryName"`
}

type RemoveRepositoryFromTeamPayload struct {
	Success bool `json:"success"`
}

type TeamRepositoryFilter struct {
	Name *string `json:"name"`
}
