package graph

import (
	"context"

	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
)

func (r *queryResolver) Roles(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor) (*pagination.Connection[*authz.Role], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return authz.ListRoles(ctx, page)
}
