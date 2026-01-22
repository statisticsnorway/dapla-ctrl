package group

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/statisticsnorway/dapla-api/internal/activitylog"
	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/graph/apierror"
	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/group/groupsql"
	"github.com/statisticsnorway/dapla-api/internal/slug"
	"k8s.io/utils/ptr"
)

func GenerateName(input *CreateGroupInput) string {
	s, _ := strings.CutSuffix(
		fmt.Sprintf("%s-%s-%s", input.TeamSlug, input.Category, *input.Suffix),
		"-",
	)
	return s
}

func Create(ctx context.Context, input *CreateGroupInput, actor *authz.Actor) (*Group, error) {
	if err := input.Validate(ctx); err != nil {
		return nil, err
	}

	var group *groupsql.Group
	err := database.Transaction(ctx, func(ctx context.Context) error {
		var err error
		group, err = db(ctx).Create(ctx, groupsql.CreateParams{
			Name:     GenerateName(input),
			TeamSlug: input.TeamSlug,
			Category: input.Category,
			Suffix:   *input.Suffix,
		})
		if err != nil {
			return err
		}

		return activitylog.Create(ctx, activitylog.CreateInput{
			Action:       activitylog.ActivityLogEntryActionCreated,
			Actor:        actor.User,
			ResourceType: ActivityLogEntryResourceTypeGroup,
			ResourceName: group.Name,
			TeamSlug:     ptr.To(input.TeamSlug),
		})
	})
	if err != nil {
		return nil, err
	}

	return toGraphGroup(group), nil
}

func Get(ctx context.Context, name string) (*Group, error) {
	g, err := fromContext(ctx).groupLoader.Load(ctx, name)
	if err != nil {
		return nil, handleError(err)
	}
	return g, nil
}

func GetByIdent(ctx context.Context, id ident.Ident) (*Group, error) {
	groupName, err := parseIdent(id)
	if err != nil {
		return nil, err
	}
	return Get(ctx, groupName)
}

func List(ctx context.Context, page *pagination.Pagination, orderBy *GroupOrder, filter *GroupFilter) (*GroupConnection, error) {
	q := db(ctx)

	ret, err := q.List(ctx, groupsql.ListParams{
		Offset:  page.Offset(),
		Limit:   page.Limit(),
		OrderBy: orderBy.String(),
		Filter:  filter.CategoryFilter(),
	})
	if err != nil {
		return nil, err
	}

	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}

	return pagination.NewConvertConnection(ret, page, total, func(from *groupsql.ListRow) *Group {
		return toGraphGroup(&from.Group)
	}), nil
}

func ListForUser(ctx context.Context, userID uuid.UUID, page *pagination.Pagination, orderBy *UserGroupOrder, filter *GroupFilter) (*GroupMemberConnection, error) {
	q := db(ctx)

	ret, err := q.ListForUser(ctx, groupsql.ListForUserParams{
		UserID:  userID,
		Offset:  page.Offset(),
		Limit:   page.Limit(),
		OrderBy: orderBy.String(),
		Filter:  filter.CategoryFilter(),
	})
	if err != nil {
		return nil, err
	}

	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}
	return pagination.NewConvertConnection(ret, page, total, toGraphUserGroup), nil
}

func ListForTeamMember(ctx context.Context, teamSlug slug.Slug, userID uuid.UUID) ([]*Group, error) {
	q := db(ctx)

	ret, err := q.ListForTeamMember(ctx, groupsql.ListForTeamMemberParams{
		UserID:   userID,
		TeamSlug: teamSlug,
	})
	if err != nil {
		return nil, err
	}

	var groups []*Group
	for _, g := range ret {
		groups = append(groups, toGraphGroup(&g.Group))
	}

	return groups, nil
}

func GetMemberByEmail(ctx context.Context, groupName string, email string) (*GroupMember, error) {
	q := db(ctx)

	m, err := q.GetMemberByEmail(ctx, groupsql.GetMemberByEmailParams{
		GroupName: groupName,
		Email:     email,
	})
	if err != nil {
		return nil, err
	}
	return &GroupMember{
		GroupName: groupName,
		UserID:    m.ID,
	}, nil
}

func ListMembers(ctx context.Context, groupName string, page *pagination.Pagination, orderBy *GroupMemberOrder) (*GroupMemberConnection, error) {
	q := db(ctx)

	ret, err := q.ListMembers(ctx, groupsql.ListMembersParams{
		GroupName: groupName,
		Offset:    page.Offset(),
		Limit:     page.Limit(),
		OrderBy:   orderBy.String(),
	})
	if err != nil {
		return nil, err
	}

	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}
	return pagination.NewConvertConnection(ret, page, total, toGraphGroupMember), nil
}

func AddMember(ctx context.Context, input AddGroupMemberInput, actor *authz.Actor) error {
	_, err := db(ctx).GetMember(ctx, groupsql.GetMemberParams{
		GroupName: input.GroupName,
		UserID:    input.UserID,
	})
	if !errors.Is(err, pgx.ErrNoRows) {
		return apierror.Errorf("User is already a member of the group.")
	}

	return database.Transaction(ctx, func(ctx context.Context) error {
		params := groupsql.AddMemberParams{
			UserID:    input.UserID,
			GroupName: input.GroupName,
		}
		if err := db(ctx).AddMember(ctx, params); err != nil {
			return err
		}
		group, err := db(ctx).Get(ctx, input.GroupName)
		if err != nil {
			return err
		}

		// Not the end of the world if it fails
		_ = db(ctx).RefreshTeamMembers(ctx)

		return activitylog.Create(ctx, activitylog.CreateInput{
			Action:       activitylog.ActivityLogEntryActionAdded,
			Actor:        actor.User,
			ResourceType: ActivityLogEntryResourceTypeGroup,
			ResourceName: input.GroupName,
			TeamSlug:     &group.TeamSlug,
			Data: &GroupMemberAddedActivityLogEntryData{
				UserUUID:  input.UserID,
				UserEmail: input.UserEmail,
			},
		})
	})
}

func RemoveMember(ctx context.Context, input RemoveGroupMemberInput, actor *authz.Actor) error {
	_, err := db(ctx).GetMember(ctx, groupsql.GetMemberParams{
		GroupName: input.GroupName,
		UserID:    input.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apierror.Errorf("User is not a member of the group.")
	} else if err != nil {
		return err
	}

	return database.Transaction(ctx, func(ctx context.Context) error {
		params := groupsql.RemoveMemberParams{
			UserID:    input.UserID,
			GroupName: input.GroupName,
		}
		if err := db(ctx).RemoveMember(ctx, params); err != nil {
			return err
		}
		group, err := db(ctx).Get(ctx, input.GroupName)
		if err != nil {
			return err
		}
		// Not the end of the world if it fails
		_ = db(ctx).RefreshTeamMembers(ctx)

		return activitylog.Create(ctx, activitylog.CreateInput{
			Action:       activitylog.ActivityLogEntryActionRemoved,
			Actor:        actor.User,
			ResourceType: ActivityLogEntryResourceTypeGroup,
			ResourceName: input.GroupName,
			TeamSlug:     &group.TeamSlug,
			Data: &GroupMemberRemovedActivityLogEntryData{
				UserUUID:  input.UserID,
				UserEmail: input.UserEmail,
			},
		})
	})
}

// Exists checks if an active team with the given slug exists.
func GroupExists(ctx context.Context, teamSlug slug.Slug, category, suffix string) (bool, error) {
	return db(ctx).GroupExists(ctx, groupsql.GroupExistsParams{
		TeamSlug: teamSlug,
		Category: category,
		Suffix:   suffix,
	})
}

func ListByTeamSlug(ctx context.Context, teamSlug slug.Slug, page *pagination.Pagination, order *GroupOrder, filter *GroupFilter) (*GroupConnection, error) {
	ret, err := db(ctx).ListByTeamSlug(ctx, groupsql.ListByTeamSlugParams{
		TeamSlug: teamSlug,
		OrderBy:  order.String(),
		Filter:   filter.CategoryFilter(),
	})
	if err != nil {
		return nil, err
	}

	p := pagination.Slice(ret, page)
	return pagination.NewConvertConnection(p, page, len(ret), toGraphGroup), nil
}
