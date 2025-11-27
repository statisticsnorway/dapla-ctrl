package graph

import (
	"context"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/group"
	"github.com/statisticsnorway/dapla-api/internal/user"
)

func (r *groupResolver) Members(ctx context.Context, obj *group.Group, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *group.GroupMemberOrder) (*pagination.Connection[*group.GroupMember], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return group.ListMembers(ctx, obj.Name, page, orderBy)
}

func (r *groupMemberResolver) Group(ctx context.Context, obj *group.GroupMember) (*group.Group, error) {
	return group.Get(ctx, obj.GroupName)
}

func (r *groupMemberResolver) User(ctx context.Context, obj *group.GroupMember) (*user.User, error) {
	return user.Get(ctx, obj.UserID)
}

func (r *mutationResolver) CreateGroup(ctx context.Context, input group.CreateGroupInput) (*group.CreateGroupPayload, error) {
	actor := authz.ActorFromContext(ctx)

	if err := authz.CanCreateGroup(ctx, input.TeamSlug); err != nil {
		return nil, err
	}

	g, err := group.Create(ctx, &input, actor)
	if err != nil {
		return nil, err
	}

	correlationId := uuid.New()
	r.triggerTeamUpdatedEvent(ctx, input.TeamSlug, correlationId)

	return &group.CreateGroupPayload{
		Group: g,
	}, nil
}

func (r *mutationResolver) AddGroupMember(ctx context.Context, input group.AddGroupMemberInput) (*group.AddGroupMemberPayload, error) {
	actor := authz.ActorFromContext(ctx)

	g, err := group.Get(ctx, input.GroupName)
	if err != nil {
		return nil, err
	}

	if err := authz.CanManageGroupMembers(ctx, g.TeamSlug); err != nil {
		return nil, err
	}

	u, err := user.GetByEmail(ctx, input.UserEmail)
	if err != nil {
		return nil, err
	}

	input.UserID = u.UUID
	if err := group.AddMember(ctx, input, actor); err != nil {
		return nil, err
	}

	correlationId := uuid.New()
	r.triggerTeamUpdatedEvent(ctx, g.TeamSlug, correlationId)

	return &group.AddGroupMemberPayload{
		Member: &group.GroupMember{
			GroupName: input.GroupName,
			UserID:    input.UserID,
		},
	}, nil
}

func (r *mutationResolver) RemoveGroupMember(ctx context.Context, input group.RemoveGroupMemberInput) (*group.RemoveGroupMemberPayload, error) {
	actor := authz.ActorFromContext(ctx)

	g, err := group.Get(ctx, input.GroupName)
	if err != nil {
		return nil, err
	}

	if err := authz.CanManageGroupMembers(ctx, g.TeamSlug); err != nil {
		return nil, err
	}

	user, err := user.GetByEmail(ctx, input.UserEmail)
	if err != nil {
		return nil, err
	}
	input.UserID = user.UUID

	if err := group.RemoveMember(ctx, input, actor); err != nil {
		return nil, err
	}

	correlationId := uuid.New()
	r.triggerTeamUpdatedEvent(ctx, g.TeamSlug, correlationId)

	return &group.RemoveGroupMemberPayload{
		GroupName: input.GroupName,
		UserID:    input.UserID,
	}, nil
}

func (r *queryResolver) Groups(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *group.GroupOrder) (*pagination.Connection[*group.Group], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return group.List(ctx, page, orderBy)
}

func (r *queryResolver) Group(ctx context.Context, name string) (*group.Group, error) {
	return group.Get(ctx, name)
}

func (r *removeGroupMemberPayloadResolver) User(ctx context.Context, obj *group.RemoveGroupMemberPayload) (*user.User, error) {
	return user.Get(ctx, obj.UserID)
}

func (r *removeGroupMemberPayloadResolver) Group(ctx context.Context, obj *group.RemoveGroupMemberPayload) (*group.Group, error) {
	return group.Get(ctx, obj.GroupName)
}

func (r *Resolver) Group() gengql.GroupResolver { return &groupResolver{r} }

func (r *Resolver) GroupMember() gengql.GroupMemberResolver { return &groupMemberResolver{r} }

func (r *Resolver) RemoveGroupMemberPayload() gengql.RemoveGroupMemberPayloadResolver {
	return &removeGroupMemberPayloadResolver{r}
}

type (
	groupResolver                    struct{ *Resolver }
	groupMemberResolver              struct{ *Resolver }
	removeGroupMemberPayloadResolver struct{ *Resolver }
)
