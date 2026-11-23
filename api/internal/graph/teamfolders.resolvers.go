package graph

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/teamfolders"
)

func (r *teamResolver) TeamFolders(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor) (*pagination.Connection[*teamfolders.TeamFolder], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return teamfolders.ListForTeam(ctx, obj.Slug, page)
}

func (r *teamFolderResolver) Team(ctx context.Context, obj *teamfolders.TeamFolder) (*team.Team, error) {
	return team.Get(ctx, obj.TeamSlug)
}

func (r *Resolver) TeamFolder() gengql.TeamFolderResolver { return &teamFolderResolver{r} }

type teamFolderResolver struct{ *Resolver }
