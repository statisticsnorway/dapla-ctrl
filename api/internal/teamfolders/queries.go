package teamfolders

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/teamfolders/teamfolderssql"
)

func ListForTeam(ctx context.Context, teamSlug slug.Slug, page *pagination.Pagination, filter *TeamFolderFilter) (*TeamFolderConnection, error) {
	q := db(ctx)

	ret, err := q.ListForTeam(ctx, teamfolderssql.ListForTeamParams{
		TeamSlug: teamSlug,
		Offset:   page.Offset(),
		Limit:    page.Limit(),
		Envs:     filter.EnvFilter(),
	})
	if err != nil {
		return nil, err
	}

	total := 0
	if len(ret) > 0 {
		total = int(ret[0].TotalCount)
	}
	return pagination.NewConvertConnection(ret, page, total, func(from *teamfolderssql.ListForTeamRow) *TeamFolder {
		return &TeamFolder{
			TeamSlug: from.TeamSlug,
			Env:      from.Env,
			FolderID: from.FolderID,
		}
	}), nil
}

func GetByIdent(ctx context.Context, ident ident.Ident) (*TeamFolder, error) {
	teamSlug, env, err := parseIdent(ident)
	if err != nil {
		return nil, err
	}

	f, err := db(ctx).Get(ctx, teamfolderssql.GetParams{
		TeamSlug: teamSlug,
		Env:      env,
	})
	if err != nil {
		return nil, handleError(err)
	}
	return toGraphFolder(f), nil
}
