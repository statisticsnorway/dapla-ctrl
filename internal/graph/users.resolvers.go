package graph

import (
	"context"

	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/graph/apierror"
	"github.com/statisticsnorway/dapla-api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/section"
	"github.com/statisticsnorway/dapla-api/internal/team"
	"github.com/statisticsnorway/dapla-api/internal/user"
)

func (r *queryResolver) Users(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *user.UserOrder) (*pagination.Connection[*user.User], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return user.List(ctx, page, orderBy)
}

func (r *queryResolver) User(ctx context.Context, email *string) (*user.User, error) {
	if email == nil {
		return nil, apierror.Errorf("email argument must be provided")
	}
	return user.GetByEmail(ctx, *email)
}

func (r *queryResolver) Me(ctx context.Context) (authz.AuthenticatedUser, error) {
	return authz.ActorFromContext(ctx).User, nil
}

func (r *userResolver) Teams(ctx context.Context, obj *user.User, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *team.UserTeamOrder) (*pagination.Connection[*team.TeamMember], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return team.ListForUser(ctx, obj.UUID, page, orderBy)
}

func (r *userResolver) IsSectionManager(ctx context.Context, obj *user.User) (bool, error) {
	return section.IsUserSectionManager(ctx, obj.UUID)
}

func (r *Resolver) User() gengql.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
