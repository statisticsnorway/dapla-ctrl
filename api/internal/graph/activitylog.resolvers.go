package graph

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/activitylog"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/group"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/reconciler"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team"
)

func (r *groupResolver) ActivityLog(ctx context.Context, obj *group.Group, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, filter *activitylog.ActivityLogFilter) (*pagination.Connection[activitylog.ActivityLogEntry], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}
	return activitylog.ListForResource(ctx, group.ActivityLogEntryResourceTypeGroup, obj.Name, page, filter)
}

func (r *queryResolver) ActivityLog(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, filter *activitylog.ActivityLogFilter) (*pagination.Connection[activitylog.ActivityLogEntry], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}
	return activitylog.List(ctx, page, filter)
}

func (r *reconcilerResolver) ActivityLog(ctx context.Context, obj *reconciler.Reconciler, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, filter *activitylog.ActivityLogFilter) (*pagination.Connection[activitylog.ActivityLogEntry], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return activitylog.ListForResource(ctx, reconciler.ActivityLogEntryResourceTypeReconciler, obj.Name, page, filter)
}

func (r *teamResolver) ActivityLog(ctx context.Context, obj *team.Team, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, filter *activitylog.ActivityLogFilter) (*pagination.Connection[activitylog.ActivityLogEntry], error) {
	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}

	return activitylog.ListForTeam(ctx, obj.Slug, page, filter)
}
