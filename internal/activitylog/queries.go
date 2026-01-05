package activitylog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-api/internal/activitylog/activitylogsql"
	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/slug"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"k8s.io/utils/ptr"
)

type CreateInput struct {
	Action       ActivityLogEntryAction
	Actor        authz.AuthenticatedUser
	ResourceType ActivityLogEntryResourceType
	ResourceName string

	Data     any        // optional
	TeamSlug *slug.Slug // optional
}

func MarshalData(input CreateInput) ([]byte, error) {
	if input.Data == nil {
		return nil, nil
	}

	bytes, err := json.Marshal(input.Data)
	if err != nil {
		return nil, fmt.Errorf("marshaling data: %w", err)
	}

	return bytes, nil
}

func Create(ctx context.Context, input CreateInput) error {
	q := db(ctx)

	data, err := MarshalData(input)
	if err != nil {
		return err
	}

	return q.Create(ctx, activitylogsql.CreateParams{
		Action:       string(input.Action),
		Actor:        input.Actor.Identity(),
		Data:         data,
		ResourceName: input.ResourceName,
		ResourceType: string(input.ResourceType),
		TeamSlug:     input.TeamSlug,
	})
}

func Get(ctx context.Context, uid uuid.UUID) (ActivityLogEntry, error) {
	return fromContext(ctx).activityLogLoader.Load(ctx, uid)
}

func GetByIdent(ctx context.Context, id ident.Ident) (ActivityLogEntry, error) {
	uid, err := parseIdent(id)
	if err != nil {
		return nil, err
	}
	return Get(ctx, uid)
}

func ListForTeam(ctx context.Context, teamSlug slug.Slug, page *pagination.Pagination, filter *ActivityLogFilter) (*ActivityLogEntryConnection, error) {
	q := db(ctx)

	ret, err := q.ListForTeam(ctx, activitylogsql.ListForTeamParams{
		TeamSlug: ptr.To(teamSlug),
		Offset:   page.Offset(),
		Limit:    page.Limit(),
		Filter:   withFilters(filter),
	})
	if err != nil {
		return nil, err
	}

	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}
	return pagination.NewConvertConnectionWithError(ret, page, total, func(from *activitylogsql.ListForTeamRow) (ActivityLogEntry, error) {
		return toGraphActivityLogEntry(&from.ActivityLogEntry)
	})
}

func ListForResource(ctx context.Context, resourceType ActivityLogEntryResourceType, resourceName string, page *pagination.Pagination, filter *ActivityLogFilter) (*ActivityLogEntryConnection, error) {
	q := db(ctx)

	ret, err := q.ListForResource(ctx, activitylogsql.ListForResourceParams{
		ResourceType: string(resourceType),
		ResourceName: resourceName,
		Offset:       page.Offset(),
		Limit:        page.Limit(),
		Filter:       withFilters(filter),
	})
	if err != nil {
		return nil, err
	}

	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}
	return pagination.NewConvertConnectionWithError(ret, page, total, func(from *activitylogsql.ListForResourceRow) (ActivityLogEntry, error) {
		return toGraphActivityLogEntry(&from.ActivityLogEntry)
	})
}

func List(ctx context.Context, page *pagination.Pagination, filter *ActivityLogFilter) (*ActivityLogEntryConnection, error) {
	q := db(ctx)

	ret, err := q.List(ctx, activitylogsql.ListParams{
		Offset: page.Offset(),
		Limit:  page.Limit(),
		Filter: withFilters(filter),
	})
	if err != nil {
		return nil, err
	}

	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}
	return pagination.NewConvertConnectionWithError(ret, page, total, func(from *activitylogsql.ListRow) (ActivityLogEntry, error) {
		return toGraphActivityLogEntry(&from.ActivityLogEntry)
	})
}

func toGraphActivityLogEntry(row *activitylogsql.ActivityLogEntry) (ActivityLogEntry, error) {
	titler := cases.Title(language.English)
	entry := GenericActivityLogEntry{
		Action:       ActivityLogEntryAction(row.Action),
		Actor:        row.Actor,
		CreatedAt:    row.CreatedAt.Time,
		Message:      titler.String(row.Action) + " " + titler.String(row.ResourceType),
		ResourceType: ActivityLogEntryResourceType(row.ResourceType),
		ResourceName: row.ResourceName,
		TeamSlug:     row.TeamSlug,
		UUID:         row.ID,
		Data:         row.Data,
	}
	if row.TeamSlug == nil {
		entry.TeamSlug = ptr.To(slug.Slug(""))
	}

	transformer, ok := knownTransformers[ActivityLogEntryResourceType(row.ResourceType)]
	if !ok {
		return nil, fmt.Errorf("no transformer registered for activity log resource type: %q", row.ResourceType)
	}

	return transformer(entry)
}
