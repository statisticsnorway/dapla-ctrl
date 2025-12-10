package section

import (
	"context"

	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/section/sectionsql"
)

func Get(ctx context.Context, sectionCode string) (*Section, error) {
	section, err := fromContext(ctx).sectionLoader.Load(ctx, sectionCode)
	if err != nil {
		return nil, handleError(err)
	}
	return section, nil
}

func GetByIdent(ctx context.Context, ident ident.Ident) (*Section, error) {
	sectionCode, err := parseIdent(ident)
	if err != nil {
		return nil, err
	}
	return Get(ctx, sectionCode)
}

func List(ctx context.Context, page *pagination.Pagination, orderBy *SectionOrder) (*SectionConnection, error) {
	q := db(ctx)

	ret, err := q.List(ctx, sectionsql.ListParams{
		Offset:  page.Offset(),
		Limit:   page.Limit(),
		OrderBy: orderBy.String(),
	})
	if err != nil {
		return nil, err
	}

	total := 0
	if len(ret) > 0 {
		total = int(ret[0].TotalCount)
	}
	return pagination.NewConvertConnection(ret, page, total, func(from *sectionsql.ListRow) *Section {
		return toGraphSection(&from.Section)
	}), nil
}
