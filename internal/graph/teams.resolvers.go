package graph

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/graph/apierror"
	"github.com/statisticsnorway/dapla-api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/group"
	"github.com/statisticsnorway/dapla-api/internal/section"
	"github.com/statisticsnorway/dapla-api/internal/slug"
	"github.com/statisticsnorway/dapla-api/internal/team"
	"github.com/statisticsnorway/dapla-api/internal/user"
)

func (r *mutationResolver) CreateTeam(ctx context.Context, input team.CreateTeamInput) (*team.CreateTeamPayload, error) {
	actor := authz.ActorFromContext(ctx)

	if err := authz.CanCreateTeam(ctx); err != nil {
		return nil, err
	}

	if !actor.User.IsAdmin() {
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

func (r *teamResolver) Section(ctx context.Context, obj *team.Team) (*section.Section, error) {
	return section.Get(ctx, obj.SectionCode)
}

func (r *teamResolver) Members(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *team.TeamMemberOrder) (*pagination.Connection[*team.TeamMember], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return team.ListMembers(ctx, obj.Slug, page, orderBy)
}

func (r *teamResolver) Groups(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *group.GroupOrder) (*pagination.Connection[*group.Group], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return group.ListByTeamSlug(ctx, obj.Slug, page, orderBy)
}

func (r *teamResolver) ViewerIsOwner(ctx context.Context, obj *team.Team) (bool, error) {
	return team.UserIsOwner(ctx, obj.Slug, authz.ActorFromContext(ctx).User.GetID())
}

func (r *teamResolver) ViewerIsMember(ctx context.Context, obj *team.Team) (bool, error) {
	return team.UserIsMember(ctx, obj.Slug, authz.ActorFromContext(ctx).User.GetID())
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

func (r *Resolver) Team() gengql.TeamResolver { return &teamResolver{r} }

func (r *Resolver) TeamMember() gengql.TeamMemberResolver { return &teamMemberResolver{r} }

type (
	teamResolver       struct{ *Resolver }
	teamMemberResolver struct{ *Resolver }
)
