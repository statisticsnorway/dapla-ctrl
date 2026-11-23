package teamfolders

import (
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/teamfolders/teamfolderssql"
)

//mgo:gen model
//mgo:impl node paginated
type TeamFolder struct {
	TeamSlug slug.Slug `json:"slug"`
	Env      string    `json:"env"`
	FolderID string    `json:"folderId"`
}

func (f TeamFolder) ID() ident.Ident {
	return NewIdent(f.TeamSlug, f.Env)
}

func (f TeamFolder) Name() string {
	return f.TeamSlug.String()
}

func toGraphFolder(f *teamfolderssql.GcpTeamFolder) *TeamFolder {
	return &TeamFolder{
		TeamSlug: f.TeamSlug,
		Env:      f.Env,
		FolderID: f.FolderID,
	}
}
