package graph

import (
	"context"
	"strings"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/apierror"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/group"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/section"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/sharedbucketsstopgap"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/user"
)

func (r *queryResolver) Users(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *user.UserOrder) (*pagination.Connection[*user.User], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return user.List(ctx, page, orderBy)
}

func (r *queryResolver) TeamMembers(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *user.UserOrder) (*pagination.Connection[*user.User], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return user.ListUsersWithTeams(ctx, page, orderBy)
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

func (r *userResolver) FirstName(ctx context.Context, obj *user.User) (string, error) {
	parts := strings.SplitN(obj.Name, ", ", 2)
	if len(parts) == 1 {
		// Assuming people use firstName more than lastName,
		// returning firstName=name for an "invalid" name seems correct.
		return parts[0], nil
	}
	return parts[1], nil
}

func (r *userResolver) LastName(ctx context.Context, obj *user.User) (string, error) {
	parts := strings.SplitN(obj.Name, ", ", 2)
	if len(parts) == 1 {
		// Assuming people use firstName more than lastName,
		// returning lastName="" for an "invalid" name seems correct.
		return "", nil
	}
	return parts[0], nil
}

func (r *userResolver) Section(ctx context.Context, obj *user.User) (*section.Section, error) {
	if obj.SectionCode == nil {
		return nil, nil
	}
	return section.Get(ctx, *obj.SectionCode)
}

func (r *userResolver) Teams(ctx context.Context, obj *user.User, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *team.TeamOrder) (*pagination.Connection[*team.TeamMember], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return team.ListForUser(ctx, obj.UUID, page, orderBy)
}

func (r *userResolver) TeamMembers(ctx context.Context, obj *user.User, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *user.UserOrder) (*pagination.Connection[*user.User], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return user.ListTeamMembersForUser(ctx, obj.UUID, page, orderBy)
}

func (r *userResolver) Groups(ctx context.Context, obj *user.User, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *group.GroupOrder, filter *group.GroupFilter) (*pagination.Connection[*group.GroupMember], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return group.ListForUser(ctx, obj.UUID, page, orderBy, filter)
}

func (r *userResolver) SharedBucketsAccess(ctx context.Context, obj *user.User, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *sharedbucketsstopgap.SharedBucketOrder) (*pagination.Connection[*sharedbucketsstopgap.SharedBucketAccess], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return sharedbucketsstopgap.ListForUser(ctx, obj.UUID, page, orderBy)
}

func (r *userResolver) IsSectionManager(ctx context.Context, obj *user.User) (bool, error) {
	return section.IsUserSectionManager(ctx, obj.UUID)
}

func (r *Resolver) User() gengql.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
