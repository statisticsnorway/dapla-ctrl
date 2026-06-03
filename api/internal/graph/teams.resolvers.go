package graph

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/apierror"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/group"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/section"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/sharedbucketsstopgap"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/user"
)

func (r *addTeamAccessManagerPayloadResolver) Team(ctx context.Context, obj *team.AddTeamAccessManagerPayload) (*team.Team, error) {
	return team.Get(ctx, obj.TeamSlug)
}

func (r *addTeamAccessManagerPayloadResolver) User(ctx context.Context, obj *team.AddTeamAccessManagerPayload) (*user.User, error) {
	return user.Get(ctx, obj.UserId)
}

func (r *mutationResolver) CreateTeam(ctx context.Context, input team.CreateTeamInput) (*team.CreateTeamPayload, error) {
	actor := authz.ActorFromContext(ctx)

	if err := authz.CanCreateTeam(ctx); err != nil {
		return nil, err
	}

	isRegularUser := !actor.User.IsAdmin()

	if isRegularUser {
		s, err := section.GetByManagerId(ctx, actor.User.GetID())
		if errors.As(err, &section.ErrNotFound{}) {
			return nil, apierror.Errorf("You do not have permission to create teams.")
		} else if err != nil {
			return nil, err
		}
		// Default to the section the user is a manager for
		if input.SectionCode == "" {
			input.SectionCode = s.Code
		}
		// Managers can only create teams in their own section
		if input.SectionCode != s.Code {
			return nil, apierror.Errorf("You cannot create a team in section %s", input.SectionCode)
		}
	} else if input.SectionCode == "" {
		input.SectionCode = "724"
	}

	if _, err := section.Get(ctx, input.SectionCode); errors.As(err, &section.ErrNotFound{}) {
		return nil, apierror.Errorf("Section %s does not exist", input.SectionCode)
	} else if err != nil {
		return nil, err
	}

	if strings.HasPrefix(input.Slug.String(), "team") {
		return nil, &slug.ErrInvalidSlug{Message: "The name prefix 'team' is redundant. When you create a team, it is by definition a team. Try again with a different name, perhaps just removing the prefix?"}
	}

	if input.IsManaged == nil {
		input.IsManaged = new(true)
	}

	if isRegularUser && !*input.IsManaged {
		// Require that the actor is a section manager of a 7xx section
		s, err := section.GetByManagerId(ctx, actor.User.GetID())
		if errors.As(err, &section.ErrNotFound{}) {
			return nil, apierror.Errorf("You do not have permission to create self_managed teams.")
		} else if err != nil {
			return nil, err
		}
		if s.Code[0] != '7' {
			return nil, apierror.Errorf("Only Section managers in IT department may create self-managed teams.")
		}
	}

	t, err := team.Create(ctx, &input, actor)
	if err != nil {
		return nil, err
	}

	correlationID := uuid.New()
	r.triggerTeamCreatedEvent(ctx, input.Slug, correlationID)

	return &team.CreateTeamPayload{
		Team: t,
	}, nil
}

func (r *mutationResolver) UpdateTeam(ctx context.Context, input team.UpdateTeamInput) (*team.UpdateTeamPayload, error) {
	actor := authz.ActorFromContext(ctx)

	if err := authz.CanUpdateTeamMetadata(ctx, input.Slug); err != nil {
		return nil, err
	}

	t, err := team.Update(ctx, &input, actor)
	if err != nil {
		return nil, err
	}

	correlationID := uuid.New()
	r.triggerTeamUpdatedEvent(ctx, input.Slug, correlationID)

	return &team.UpdateTeamPayload{
		Team: t,
	}, nil
}

func (r *mutationResolver) AddTeamAccessManager(ctx context.Context, input team.AddTeamAccessManagerInput) (*team.AddTeamAccessManagerPayload, error) {
	actor := authz.ActorFromContext(ctx)

	if err := authz.CanManageTeam(ctx, input.TeamSlug); err != nil {
		return nil, err
	}

	u, err := user.GetByEmail(ctx, input.UserEmail)
	if err != nil {
		return nil, err
	}

	input.UserId = u.UUID
	if err := team.AddAccessManager(ctx, input, actor); err != nil {
		return nil, err
	}

	return &team.AddTeamAccessManagerPayload{
		UserId:   input.UserId,
		TeamSlug: input.TeamSlug,
	}, nil
}

func (r *mutationResolver) RemoveTeamAccessManager(ctx context.Context, input team.RemoveTeamAccessManagerInput) (*team.RemoveTeamAccessManagerPayload, error) {
	actor := authz.ActorFromContext(ctx)

	if err := authz.CanManageTeam(ctx, input.TeamSlug); err != nil {
		return nil, err
	}

	u, err := user.GetByEmail(ctx, input.UserEmail)
	if err != nil {
		return nil, err
	}

	input.UserId = u.UUID
	if err := team.RemoveAccessManager(ctx, input, actor); err != nil {
		return nil, err
	}
	return &team.RemoveTeamAccessManagerPayload{
		UserId:   input.UserId,
		TeamSlug: input.TeamSlug,
	}, nil
}

func (r *queryResolver) Teams(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *team.TeamOrder) (*pagination.Connection[*team.Team], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return team.List(ctx, page, orderBy)
}

func (r *queryResolver) Team(ctx context.Context, slug slug.Slug) (*team.Team, error) {
	return team.Get(ctx, slug)
}

func (r *removeTeamAccessManagerPayloadResolver) Team(ctx context.Context, obj *team.RemoveTeamAccessManagerPayload) (*team.Team, error) {
	return team.Get(ctx, obj.TeamSlug)
}

func (r *removeTeamAccessManagerPayloadResolver) User(ctx context.Context, obj *team.RemoveTeamAccessManagerPayload) (*user.User, error) {
	return user.Get(ctx, obj.UserId)
}

func (r *teamResolver) Section(ctx context.Context, obj *team.Team) (*section.Section, error) {
	return section.Get(ctx, obj.SectionCode)
}

func (r *teamResolver) Members(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *user.UserOrder) (*pagination.Connection[*team.TeamMember], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return team.ListMembers(ctx, obj.Slug, page, orderBy)
}

func (r *teamResolver) Groups(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *group.GroupOrder, filter *group.GroupFilter) (*pagination.Connection[*group.Group], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return group.ListByTeamSlug(ctx, obj.Slug, page, orderBy, filter)
}

func (r *teamResolver) SharedBuckets(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *sharedbucketsstopgap.SharedBucketOrder, filter *sharedbucketsstopgap.SharedBucketFilter) (*pagination.Connection[*sharedbucketsstopgap.SharedBucket], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}
	return sharedbucketsstopgap.ListForTeam(ctx, obj.Slug, page, orderBy, filter)
}

func (r *teamResolver) SharedBucketsAccess(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *sharedbucketsstopgap.SharedBucketOrder) (*pagination.Connection[*sharedbucketsstopgap.SharedBucket], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}
	return sharedbucketsstopgap.ListAccessToForTeam(ctx, obj.Slug, page, orderBy)
}

func (r *teamResolver) ViewerIsOwner(ctx context.Context, obj *team.Team) (bool, error) {
	return team.UserIsOwner(ctx, obj.Slug, authz.ActorFromContext(ctx).User.GetID())
}

func (r *teamResolver) ViewerIsMember(ctx context.Context, obj *team.Team) (bool, error) {
	return team.UserIsMember(ctx, obj.Slug, authz.ActorFromContext(ctx).User.GetID())
}

func (r *teamResolver) ViewerCanManageMembers(ctx context.Context, obj *team.Team) (bool, error) {
	if err := authz.CanManageGroupMembers(ctx, obj.Slug); errors.Is(err, authz.ErrUnauthorized) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (r *teamResolver) AccessManagers(ctx context.Context, obj *team.Team) ([]*team.TeamAccessManager, error) {
	return team.GetAccessManagers(ctx, obj.Slug)
}

func (r *teamAccessManagerResolver) Team(ctx context.Context, obj *team.TeamAccessManager) (*team.Team, error) {
	return team.Get(ctx, obj.TeamSlug)
}

func (r *teamAccessManagerResolver) User(ctx context.Context, obj *team.TeamAccessManager) (*user.User, error) {
	return user.Get(ctx, obj.UserID)
}

func (r *teamMemberResolver) Team(ctx context.Context, obj *team.TeamMember) (*team.Team, error) {
	return team.Get(ctx, obj.TeamSlug)
}

func (r *teamMemberResolver) User(ctx context.Context, obj *team.TeamMember) (*user.User, error) {
	return user.Get(ctx, obj.UserID)
}

func (r *teamMemberResolver) Groups(ctx context.Context, obj *team.TeamMember) ([]*group.Group, error) {
	return group.ListForTeamMember(ctx, obj.TeamSlug, obj.UserID)
}

func (r *teamRoleAssignedActivityLogEntryDataResolver) User(ctx context.Context, obj *team.TeamRoleAssignedActivityLogEntryData) (*user.User, error) {
	return user.Get(ctx, obj.UserId)
}

func (r *teamRoleRevokedActivityLogEntryDataResolver) User(ctx context.Context, obj *team.TeamRoleRevokedActivityLogEntryData) (*user.User, error) {
	return user.Get(ctx, obj.UserId)
}

func (r *Resolver) AddTeamAccessManagerPayload() gengql.AddTeamAccessManagerPayloadResolver {
	return &addTeamAccessManagerPayloadResolver{r}
}

func (r *Resolver) RemoveTeamAccessManagerPayload() gengql.RemoveTeamAccessManagerPayloadResolver {
	return &removeTeamAccessManagerPayloadResolver{r}
}

func (r *Resolver) Team() gengql.TeamResolver { return &teamResolver{r} }

func (r *Resolver) TeamAccessManager() gengql.TeamAccessManagerResolver {
	return &teamAccessManagerResolver{r}
}

func (r *Resolver) TeamMember() gengql.TeamMemberResolver { return &teamMemberResolver{r} }

func (r *Resolver) TeamRoleAssignedActivityLogEntryData() gengql.TeamRoleAssignedActivityLogEntryDataResolver {
	return &teamRoleAssignedActivityLogEntryDataResolver{r}
}

func (r *Resolver) TeamRoleRevokedActivityLogEntryData() gengql.TeamRoleRevokedActivityLogEntryDataResolver {
	return &teamRoleRevokedActivityLogEntryDataResolver{r}
}

type (
	addTeamAccessManagerPayloadResolver          struct{ *Resolver }
	removeTeamAccessManagerPayloadResolver       struct{ *Resolver }
	teamResolver                                 struct{ *Resolver }
	teamAccessManagerResolver                    struct{ *Resolver }
	teamMemberResolver                           struct{ *Resolver }
	teamRoleAssignedActivityLogEntryDataResolver struct{ *Resolver }
	teamRoleRevokedActivityLogEntryDataResolver  struct{ *Resolver }
)
