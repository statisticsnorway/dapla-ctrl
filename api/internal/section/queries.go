package section

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/section/sectionsql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team"
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

func GetByManagerId(ctx context.Context, userId uuid.UUID) (*Section, error) {
	ss, err := db(ctx).GetByManagerId(ctx, &userId)
	if err != nil {
		return nil, handleError(err)
	}
	return toGraphSection(ss), nil
}

func IsUserSectionManager(ctx context.Context, userId uuid.UUID) (bool, error) {
	_, err := GetByManagerId(ctx, userId)
	if errors.As(err, &ErrNotFound{}) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func ListTeams(
	ctx context.Context,
	sectionCode string,
	page *pagination.Pagination,
	orderBy *team.TeamOrder,
) (*team.TeamConnection, error) {
	q := db(ctx)

	ret, err := q.ListTeamsForSection(ctx, sectionsql.ListTeamsForSectionParams{
		SectionCode: sectionCode,
		Offset:      page.Offset(),
		Limit:       page.Limit(),
		OrderBy:     orderBy.String(),
	})
	if err != nil {
		return nil, err
	}

	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	} else if page.Offset() > 0 {
		// An empty page has no rows, so we need to count the total separately
		total, err = q.CountTeamsForSection(ctx, sectionCode)
		if err != nil {
			return nil, err
		}
	}

	return pagination.NewConvertConnection(
		ret,
		page,
		total,
		func(from *sectionsql.ListTeamsForSectionRow) *team.Team {
			t := &team.Team{
				Slug:             from.Slug,
				DisplayName:      from.DisplayName,
				SectionCode:      from.SectionCode,
				IsManaged:        from.IsManaged,
				HasManualEditing: from.HasManualEditing,
			}

			if from.LastSuccessfulSync.Valid {
				t.LastSuccessfulSync = &from.LastSuccessfulSync.Time
			}
			if from.DeleteKeyConfirmedAt.Valid {
				t.DeleteKeyConfirmedAt = &from.DeleteKeyConfirmedAt.Time
			}

			return t
		},
	), nil
}
