package graph

import (
	"context"
	"fmt"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/section"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/user"
)

func (r *entityResolver) FindSectionByCode(ctx context.Context, code string) (*section.Section, error) {
	return section.Get(ctx, code)
}

func (r *entityResolver) FindTeamBySlug(ctx context.Context, slug slug.Slug) (*team.Team, error) {
	return team.Get(ctx, slug)
}

func (r *entityResolver) FindUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return user.GetByEmail(ctx, email)
}

func (r *Resolver) Entity() gengql.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
