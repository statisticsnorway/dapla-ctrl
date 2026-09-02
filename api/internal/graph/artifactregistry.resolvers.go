package graph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/artifactregistry"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team"
)

func (r *artifactRegistryAllowedGithubReposResolver) Team(ctx context.Context, obj *artifactregistry.ArtifactRegistryAllowedGithubRepos) (*team.Team, error) {
	return team.Get(ctx, obj.TeamSlug)
}

func (r *artifactRegistryRepositoryResolver) ID(ctx context.Context, obj *artifactregistry.ArtifactRegistryRepository) (*ident.Ident, error) {
	panic(fmt.Errorf("not implemented: ID - id"))
}

func (r *artifactRegistryRepositoryResolver) Team(ctx context.Context, obj *artifactregistry.ArtifactRegistryRepository) (*team.Team, error) {
	panic(fmt.Errorf("not implemented: Team - team"))
}

func (r *mutationResolver) GrantGithubRepoAccessToTeamArtifactRegistry(ctx context.Context, input artifactregistry.GrantGithubRepoAccessToTeamArtifactRegistryInput) (*artifactregistry.GrantGithubRepoAccessToTeamArtifactRegistryPayload, error) {
	actor := authz.ActorFromContext(ctx)
	isAdmin := actor.User.IsAdmin()
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

	correlationID := uuid.New()
	r.triggerTeamUpdatedEvent(ctx, input.TeamSlug, correlationID)

	return &artifactregistry.GrantGithubRepoAccessToTeamArtifactRegistryPayload{
		Repository: argr,
	}, nil
}

func (r *mutationResolver) RevokeGithubRepoAccessFromTeamArtifactRegistry(ctx context.Context, input artifactregistry.RevokeGithubRepoAccessFromTeamArtifactRegistryInput) (*artifactregistry.RevokeGithubRepoAccessFromTeamArtifactRegistryPayload, error) {
	actor := authz.ActorFromContext(ctx)
	isAdmin := actor.User.IsAdmin()
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

	correlationID := uuid.New()
	r.triggerTeamUpdatedEvent(ctx, input.TeamSlug, correlationID)

	return &artifactregistry.RevokeGithubRepoAccessFromTeamArtifactRegistryPayload{
		Success: new(true),
	}, nil
}

func (r *mutationResolver) CreateArtifactRegistryRepository(ctx context.Context, input artifactregistry.CreateArtifactRegistryRepositoryInput) (*artifactregistry.CreateArtifactRegistryRepositoryPayload, error) {
	panic(fmt.Errorf("not implemented: CreateArtifactRegistryRepository - createArtifactRegistryRepository"))
}

func (r *teamResolver) ArtifactRegistryAllowedGithubRepos(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor) (*pagination.Connection[*artifactregistry.ArtifactRegistryAllowedGithubRepos], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return artifactregistry.ListGithubReposForTeam(ctx, obj.Slug, page)
}

func (r *teamResolver) ArtifactRegistryRepositories(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor) (*pagination.Connection[*artifactregistry.ArtifactRegistryRepository], error) {
	panic(fmt.Errorf("not implemented: ArtifactRegistryRepositories - artifactRegistryRepositories"))
}

func (r *Resolver) ArtifactRegistryAllowedGithubRepos() gengql.ArtifactRegistryAllowedGithubReposResolver {
	return &artifactRegistryAllowedGithubReposResolver{r}
}

func (r *Resolver) ArtifactRegistryRepository() gengql.ArtifactRegistryRepositoryResolver {
	return &artifactRegistryRepositoryResolver{r}
}

type (
	artifactRegistryAllowedGithubReposResolver struct{ *Resolver }
	artifactRegistryRepositoryResolver         struct{ *Resolver }
)
