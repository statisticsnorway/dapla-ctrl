package sharedbucketsstopgap

import (
	"context"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/group"
	"github.com/statisticsnorway/dapla-api/internal/sharedbucketsstopgap/sharedbucketsstopgapsql"
	"github.com/statisticsnorway/dapla-api/internal/slug"
	"github.com/statisticsnorway/dapla-api/internal/team"
	"github.com/statisticsnorway/dapla-api/internal/user"
)

func Get(ctx context.Context, name string) (*SharedBucket, error) {
	section, err := fromContext(ctx).bucketsLoader.Load(ctx, name)
	if err != nil {
		return nil, handleError(err)
	}
	return section, nil
}

func GetByIdent(ctx context.Context, ident ident.Ident) (*SharedBucket, error) {
	sectionCode, err := parseIdent(ident)
	if err != nil {
		return nil, err
	}
	return Get(ctx, sectionCode)
}

func List(ctx context.Context, page *pagination.Pagination, orderBy *SharedBucketOrder) (*SharedBucketConnection, error) {
	q := db(ctx)

	ret, err := q.List(ctx, sharedbucketsstopgapsql.ListParams{
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
	return pagination.NewConvertConnection(ret, page, total, func(from *sharedbucketsstopgapsql.ListRow) *SharedBucket {
		return toGraphBucket(&from.SharedBucketsStopgap)
	}), nil
}

func ListGroups(ctx context.Context, name string, page *pagination.Pagination, orderBy *group.GroupOrder) (*group.GroupConnection, error) {
	q := db(ctx)

	ret, err := q.ListGroupsForBucket(ctx, sharedbucketsstopgapsql.ListGroupsForBucketParams{
		Name:    name,
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

	return pagination.NewConvertConnection(ret, page, total, func(from *sharedbucketsstopgapsql.ListGroupsForBucketRow) *group.Group {
		return &group.Group{
			Name:     from.Group.Name,
			Suffix:   from.Group.Suffix,
			Category: from.Group.Category,
			TeamSlug: from.Group.TeamSlug,
		}
	}), nil
}

func ListUsers(ctx context.Context, name string, page *pagination.Pagination, orderBy *user.UserOrder) (*team.TeamMemberConnection, error) {
	q := db(ctx)

	ret, err := q.ListUsersForBucket(ctx, sharedbucketsstopgapsql.ListUsersForBucketParams{
		Name:    name,
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

	return pagination.NewConvertConnection(ret, page, total, func(from *sharedbucketsstopgapsql.ListUsersForBucketRow) *team.TeamMember {
		return &team.TeamMember{
			TeamSlug: from.TeamSlug,
			UserID:   from.UserID,
		}
	}), nil
}

func ListUniqueUsers(ctx context.Context, name string, page *pagination.Pagination, orderBy *user.UserOrder) (*user.UserConnection, error) {
	q := db(ctx)

	ret, err := q.ListUniqueUsersForBucket(ctx, sharedbucketsstopgapsql.ListUniqueUsersForBucketParams{
		Name:    name,
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

	return pagination.NewConvertConnection(ret, page, total, func(from *sharedbucketsstopgapsql.ListUniqueUsersForBucketRow) *user.User {
		return &user.User{
			UUID: from.ID,
		}
	}), nil
}

func ListTeams(ctx context.Context, name string, page *pagination.Pagination, orderBy *team.TeamOrder) (*team.TeamConnection, error) {
	q := db(ctx)

	ret, err := q.ListTeamsForBucket(ctx, sharedbucketsstopgapsql.ListTeamsForBucketParams{
		Name:    name,
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
	return pagination.NewConvertConnection(ret, page, total, func(from *sharedbucketsstopgapsql.ListTeamsForBucketRow) *team.Team {
		return &team.Team{
			Slug:        from.Slug,
			SectionCode: from.SectionCode,
			IsManaged:   from.IsManaged,
			DisplayName: from.DisplayName,
		}
	}), nil
}

func ListForTeam(ctx context.Context, teamSlug slug.Slug, page *pagination.Pagination, orderBy *SharedBucketOrder) (*SharedBucketConnection, error) {
	q := db(ctx)

	ret, err := q.ListForTeam(ctx, sharedbucketsstopgapsql.ListForTeamParams{
		TeamSlug: teamSlug,
		Offset:   page.Offset(),
		Limit:    page.Limit(),
		OrderBy:  orderBy.String(),
	})
	if err != nil {
		return nil, err
	}

	total := 0
	if len(ret) > 0 {
		total = int(ret[0].TotalCount)
	}

	return pagination.NewConvertConnection(ret, page, total, func(from *sharedbucketsstopgapsql.ListForTeamRow) *SharedBucket {
		return toGraphBucket(&from.SharedBucketsStopgap)
	}), nil
}

func ListForUser(ctx context.Context, userId uuid.UUID, page *pagination.Pagination, orderBy *SharedBucketOrder) (*SharedBucketConnection, error) {
	q := db(ctx)

	ret, err := q.ListForUser(ctx, sharedbucketsstopgapsql.ListForUserParams{
		UserID:  userId,
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

	return pagination.NewConvertConnection(ret, page, total, func(from *sharedbucketsstopgapsql.ListForUserRow) *SharedBucket {
		return toGraphBucket(&from.SharedBucketsStopgap)
	}), nil
}

func ListAccessToForTeam(ctx context.Context, teamSlug slug.Slug, page *pagination.Pagination, orderBy *SharedBucketOrder) (*SharedBucketConnection, error) {
	q := db(ctx)

	ret, err := q.ListAccessToForTeam(ctx, sharedbucketsstopgapsql.ListAccessToForTeamParams{
		TeamSlug: teamSlug,
		Offset:   page.Offset(),
		Limit:    page.Limit(),
		OrderBy:  orderBy.String(),
	})
	if err != nil {
		return nil, err
	}

	total := 0
	if len(ret) > 0 {
		total = int(ret[0].TotalCount)
	}

	return pagination.NewConvertConnection(ret, page, total, func(from *sharedbucketsstopgapsql.ListAccessToForTeamRow) *SharedBucket {
		return toGraphBucket(&from.SharedBucketsStopgap)
	}), nil
}
