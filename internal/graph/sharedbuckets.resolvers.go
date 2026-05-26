package graph

import (
	"context"

	"github.com/statisticsnorway/dapla-api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/group"
	"github.com/statisticsnorway/dapla-api/internal/sharedbucketsstopgap"
	"github.com/statisticsnorway/dapla-api/internal/team"
	"github.com/statisticsnorway/dapla-api/internal/user"
)

func (r *queryResolver) SharedBuckets(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *sharedbucketsstopgap.SharedBucketOrder, filter *sharedbucketsstopgap.SharedBucketFilter) (*pagination.Connection[*sharedbucketsstopgap.SharedBucket], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return sharedbucketsstopgap.List(ctx, page, orderBy, filter)
}

func (r *queryResolver) SharedBucket(ctx context.Context, name string) (*sharedbucketsstopgap.SharedBucket, error) {
	return sharedbucketsstopgap.Get(ctx, name)
}

func (r *sharedBucketResolver) Team(ctx context.Context, obj *sharedbucketsstopgap.SharedBucket) (*team.Team, error) {
	return team.Get(ctx, obj.TeamSlug)
}

func (r *sharedBucketResolver) Groups(ctx context.Context, obj *sharedbucketsstopgap.SharedBucket, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *group.GroupOrder, filter *group.GroupFilter) (*pagination.Connection[*group.Group], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return sharedbucketsstopgap.ListGroups(ctx, obj.Name, page, orderBy)
}

func (r *sharedBucketResolver) Users(ctx context.Context, obj *sharedbucketsstopgap.SharedBucket, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *user.UserOrder) (*pagination.Connection[*team.TeamMember], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return sharedbucketsstopgap.ListUsers(ctx, obj.Name, page, orderBy)
}

func (r *sharedBucketResolver) UniqueUsers(ctx context.Context, obj *sharedbucketsstopgap.SharedBucket, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *user.UserOrder) (*pagination.Connection[*user.User], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return sharedbucketsstopgap.ListUniqueUsers(ctx, obj.Name, page, orderBy)
}

func (r *sharedBucketResolver) Teams(ctx context.Context, obj *sharedbucketsstopgap.SharedBucket, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *team.TeamOrder) (*pagination.Connection[*team.Team], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return sharedbucketsstopgap.ListTeams(ctx, obj.Name, page, orderBy)
}

func (r *sharedBucketAccessResolver) Bucket(ctx context.Context, obj *sharedbucketsstopgap.SharedBucketAccess) (*sharedbucketsstopgap.SharedBucket, error) {
	return sharedbucketsstopgap.Get(ctx, obj.BucketName)
}

func (r *sharedBucketAccessResolver) Team(ctx context.Context, obj *sharedbucketsstopgap.SharedBucketAccess) (*team.Team, error) {
	return team.Get(ctx, obj.TeamSlug)
}

func (r *sharedBucketAccessResolver) Groups(ctx context.Context, obj *sharedbucketsstopgap.SharedBucketAccess) ([]*group.Group, error) {
	return group.GetByNames(ctx, obj.GroupNames)
}

func (r *Resolver) SharedBucket() gengql.SharedBucketResolver { return &sharedBucketResolver{r} }

func (r *Resolver) SharedBucketAccess() gengql.SharedBucketAccessResolver {
	return &sharedBucketAccessResolver{r}
}

type (
	sharedBucketResolver       struct{ *Resolver }
	sharedBucketAccessResolver struct{ *Resolver }
)
