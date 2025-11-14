package graph

import (
	"context"

	"github.com/statisticsnorway/dapla-api/internal/activitylog"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/reconciler"
	"github.com/statisticsnorway/dapla-api/internal/team"
)

func (r *reconcilerResolver) ActivityLog(ctx context.Context, obj *reconciler.Reconciler, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor) (*pagination.Connection[activitylog.ActivityLogEntry], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return activitylog.ListForResource(ctx, reconciler.ActivityLogEntryResourceTypeReconciler, obj.Name, page)
}

func (r *teamResolver) ActivityLog(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor) (*pagination.Connection[activitylog.ActivityLogEntry], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return activitylog.ListForTeam(ctx, obj.Slug, page)
}
