package graph

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/search"
)

func (r *queryResolver) Search(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, filter search.SearchFilter) (*pagination.Connection[search.SearchNode], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return search.Search(ctx, page, filter)
}
