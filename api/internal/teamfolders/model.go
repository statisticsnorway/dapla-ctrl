package teamfolders

import (
	"strings"

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

type TeamFolderFilter struct {
	Envs []string `json:"envs,omitempty"`
}

func (f *TeamFolderFilter) EnvFilter() []string {
	var envFilter []string
	if f != nil {
		for _, c := range f.Envs {
			envFilter = append(envFilter, strings.ToLower(c))
		}
	}
	return envFilter
}
