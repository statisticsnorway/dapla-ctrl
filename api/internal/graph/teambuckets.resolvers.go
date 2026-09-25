package graph

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/teambuckets"
)

func (r *queryResolver) TeamBuckets(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *teambuckets.TeamBucketOrder, filter *teambuckets.TeamBucketFilter) (*pagination.Connection[*teambuckets.TeamBucket], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return teambuckets.List(ctx, page, orderBy, filter)
}

func (r *queryResolver) TeamBucket(ctx context.Context, name string) (*teambuckets.TeamBucket, error) {
	return teambuckets.Get(ctx, name)
}

func (r *teamResolver) TeamBuckets(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *teambuckets.TeamBucketOrder, filter *teambuckets.TeamBucketFilter) (*pagination.Connection[*teambuckets.TeamBucket], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return teambuckets.ListForTeam(ctx, obj.Slug, page, orderBy, filter)
}

func (r *teamBucketResolver) Team(ctx context.Context, obj *teambuckets.TeamBucket) (*team.Team, error) {
	return team.Get(ctx, obj.TeamSlug)
}

func (r *Resolver) TeamBucket() gengql.TeamBucketResolver { return &teamBucketResolver{r} }

type teamBucketResolver struct{ *Resolver }
