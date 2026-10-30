package graph

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/section"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/user"
)

func (r *queryResolver) Sections(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *section.SectionOrder) (*pagination.Connection[*section.Section], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}
	return section.List(ctx, page, orderBy)
}

func (r *queryResolver) Section(ctx context.Context, code string) (*section.Section, error) {
	return section.Get(ctx, code)
}

func (r *sectionResolver) Manager(ctx context.Context, obj *section.Section) (*user.User, error) {
	if obj.ManagerId == nil {
		return nil, nil
	}
	return user.Get(ctx, *obj.ManagerId)
}

func (r *sectionResolver) Teams(ctx context.Context, obj *section.Section, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, orderBy *team.TeamOrder) (*pagination.Connection[*team.Team], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}
	return section.ListTeams(ctx, obj.Code, page, orderBy)
}

func (r *Resolver) Section() gengql.SectionResolver { return &sectionResolver{r} }

type sectionResolver struct{ *Resolver }
