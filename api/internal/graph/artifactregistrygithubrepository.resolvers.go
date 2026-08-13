package graph

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/artifactregistry"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team"
)

func (r *artifactRegistryGithubRepositoryResolver) Team(ctx context.Context, obj *artifactregistry.ArtifactRegistryGithubRepository) (*team.Team, error) {
	return team.Get(ctx, obj.TeamSlug)
}

func (r *mutationResolver) AddArtifactRegistryGithubRepositoryToTeam(ctx context.Context, input artifactregistry.AddArtifactRegistryGithubRepositoryToTeamInput) (*artifactregistry.AddArtifactRegistryGithubRepositoryToTeamPayload, error) {
	actor := authz.ActorFromContext(ctx)
	isAdmin := actor.User.IsAdmin()
	if isAdmin {
		return nil, nil
	}
	isMember, err := team.UserIsMember(ctx, input.TeamSlug, actor.User.GetID())
	if err != nil {
		return nil, err
	}

	// By intention that section manager can not add
	canAdd := isAdmin || isMember
	if !canAdd {
		return nil, authz.ErrUnauthorized
	}

	team, err := team.Get(ctx, input.TeamSlug)
	if err != nil {
		return nil, err
	}
	if isMember && team.IsManaged {
		// Admin can bypass if the team is managed
		return nil, authz.ErrNotSupported
	}

	argr, err := artifactregistry.AddGithubRepositoryToTeam(ctx, input, actor)
	if err != nil {
		return nil, err
	}

	// TODO: Trigger event for reconcile

	return &artifactregistry.AddArtifactRegistryGithubRepositoryToTeamPayload{
		Repository: argr,
	}, nil
}

func (r *mutationResolver) RemoveArtifactRegistryGithubRepositoryFromTeam(ctx context.Context, input artifactregistry.RemoveArtifactRegistryGithubRepositoryFromTeamInput) (*artifactregistry.RemoveArtifactRegistryGithubRepositoryFromTeamPayload, error) {
	actor := authz.ActorFromContext(ctx)
	isAdmin := actor.User.IsAdmin()
	if isAdmin {
		return nil, nil
	}
	isMember, err := team.UserIsMember(ctx, input.TeamSlug, actor.User.GetID())
	if err != nil {
		return nil, err
	}

	canRemove := isAdmin || isMember
	if !canRemove {
		return nil, authz.ErrUnauthorized
	}

	team, err := team.Get(ctx, input.TeamSlug)
	if err != nil {
		return nil, err
	}
	if isMember && team.IsManaged {
		// Admin can bypass if the team is managed
		return nil, authz.ErrNotSupported
	}

	err = artifactregistry.RemoveGithubRepositoryFromTeam(ctx, input, actor)
	if err != nil {
		return nil, err
	}

	// TODO: Trigger event for reconcile

	return &artifactregistry.RemoveArtifactRegistryGithubRepositoryFromTeamPayload{
		Success: new(true),
	}, nil
}

func (r *teamResolver) ArtifactRegistryGithubRepository(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor) (*pagination.Connection[*artifactregistry.ArtifactRegistryGithubRepository], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return artifactregistry.ListForTeam(ctx, obj.Slug, page)
}

func (r *Resolver) ArtifactRegistryGithubRepository() gengql.ArtifactRegistryGithubRepositoryResolver {
	return &artifactRegistryGithubRepositoryResolver{r}
}

type artifactRegistryGithubRepositoryResolver struct{ *Resolver }
