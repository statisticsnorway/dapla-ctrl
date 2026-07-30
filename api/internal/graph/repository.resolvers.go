package graph

import (
	"context"
	"fmt"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/artifactregistry"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team"
)

func (r *mutationResolver) AddRepositoryToTeam(ctx context.Context, input artifactregistry.AddRepositoryToTeamInput) (*artifactregistry.AddRepositoryToTeamPayload, error) {
	panic(fmt.Errorf("not implemented: AddRepositoryToTeam - addRepositoryToTeam"))
}

func (r *mutationResolver) RemoveRepositoryFromTeam(ctx context.Context, input artifactregistry.RemoveRepositoryFromTeamInput) (*artifactregistry.RemoveRepositoryFromTeamPayload, error) {
	panic(fmt.Errorf("not implemented: RemoveRepositoryFromTeam - removeRepositoryFromTeam"))
}

func (r *repositoryResolver) Name(ctx context.Context, obj *artifactregistry.Repository) (string, error) {
	panic(fmt.Errorf("not implemented: Name - name"))
}

func (r *repositoryResolver) Team(ctx context.Context, obj *artifactregistry.Repository) (*team.Team, error) {
	panic(fmt.Errorf("not implemented: Team - team"))
}

func (r *teamResolver) Repositories(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *artifactregistry.RepositoryOrder, filter *artifactregistry.TeamRepositoryFilter) (*pagination.Connection[*artifactregistry.Repository], error) {
	panic(fmt.Errorf("not implemented: Repositories - repositories"))
}

func (r *Resolver) Repository() gengql.RepositoryResolver { return &repositoryResolver{r} }

type repositoryResolver struct{ *Resolver }
