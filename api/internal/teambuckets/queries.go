package teambuckets

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/teambuckets/teambucketssql"
)

func List(ctx context.Context, page *pagination.Pagination, orderBy *TeamBucketOrder, filter *TeamBucketFilter) (*TeamBucketConnection, error) {
	q := db(ctx)

	ret, err := q.List(ctx, teambucketssql.ListParams{
		Offset:  page.Offset(),
		Limit:   page.Limit(),
		OrderBy: orderBy.String(),
		Kinds:   filter.KindFilter(),
		Envs:    filter.EnvFilter(),
	})
	if err != nil {
		return nil, err
	}

	total := 0
	if len(ret) > 0 {
		total = int(ret[0].TotalCount)
	}
	return pagination.NewConvertConnection(ret, page, total, func(from *teambucketssql.ListRow) *TeamBucket {
		return &TeamBucket{
			Name:     from.Name,
			Env:      from.Env,
			Kind:     from.Kind,
			TeamSlug: from.TeamSlug,
		}
	}), nil
}

func ListForTeam(ctx context.Context, teamSlug slug.Slug, page *pagination.Pagination, orderBy *TeamBucketOrder, filter *TeamBucketFilter) (*TeamBucketConnection, error) {
	q := db(ctx)

	ret, err := q.ListForTeam(ctx, teambucketssql.ListForTeamParams{
		TeamSlug: teamSlug,
		Offset:   page.Offset(),
		Limit:    page.Limit(),
		OrderBy:  orderBy.String(),
		Kinds:    filter.KindFilter(),
		Envs:     filter.EnvFilter(),
	})
	if err != nil {
		return nil, err
	}

	total := 0
	if len(ret) > 0 {
		total = int(ret[0].TotalCount)
	}
	return pagination.NewConvertConnection(ret, page, total, func(from *teambucketssql.ListForTeamRow) *TeamBucket {
		return &TeamBucket{
			Name:     from.Name,
			Env:      from.Env,
			Kind:     from.Kind,
			TeamSlug: from.TeamSlug,
		}
	}), nil
}

func Get(ctx context.Context, name string) (*TeamBucket, error) {
	section, err := fromContext(ctx).bucketsLoader.Load(ctx, name)
	if err != nil {
		return nil, handleError(err)
	}
	return section, nil
}

func GetByIdent(ctx context.Context, ident ident.Ident) (*TeamBucket, error) {
	sectionCode, err := parseIdent(ident)
	if err != nil {
		return nil, err
	}
	return Get(ctx, sectionCode)
}
