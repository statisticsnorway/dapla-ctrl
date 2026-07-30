package artifactregistry

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/artifactregistry/artifactregistrysql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
)

func getByIdent(_ context.Context, id ident.Ident) (*Repository, error) {
	ts, format, err := parseIdent(id)
	if err != nil {
		return nil, err
	}

	return &Repository{
		TeamSlug: ts,
		Format:   format,
	}, nil
}

func ListForTeam(ctx context.Context, teamSlug slug.Slug, page *pagination.Pagination, orderBy *RepositoryOrder, filter *TeamRepositoryFilter) (*RepositoryConnection, error) {
	if filter == nil {
		filter = &TeamRepositoryFilter{}
	}

	q := db(ctx)

	ret, err := q.ListForTeam(ctx, artifactregistrysql.ListForTeamParams{
		TeamSlug: teamSlug,
		Offset:   page.Offset(),
		Limit:    page.Limit(),
	})
	if err != nil {
		return nil, err
	}
	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}
	return pagination.NewConvertConnection(ret, page, total, func(from *artifactregistrysql.ListForTeamRow) *Repository {
		return toGraphRepository(&from.TeamArtifactRegistryRepository)
	}), nil
}
