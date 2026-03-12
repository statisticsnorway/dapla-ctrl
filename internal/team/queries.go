package team

import (
	"context"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-api/internal/activitylog"
	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/slug"
	"github.com/statisticsnorway/dapla-api/internal/team/teamsql"
	"github.com/statisticsnorway/dapla-api/internal/user"
	"k8s.io/utils/ptr"
)

func Create(ctx context.Context, input *CreateTeamInput, actor *authz.Actor) (*Team, error) {
	if err := input.Validate(ctx); err != nil {
		return nil, err
	}

	var team *teamsql.Team
	err := database.Transaction(ctx, func(ctx context.Context) error {
		var err error
		team, err = db(ctx).Create(ctx, teamsql.CreateParams{
			Slug:        input.Slug,
			DisplayName: input.DisplayName,
			SectionCode: input.SectionCode,
			IsManaged:   *input.IsManaged,
		})
		if err != nil {
			return err
		}

		return activitylog.Create(ctx, activitylog.CreateInput{
			Action:       activitylog.ActivityLogEntryActionCreated,
			Actor:        actor.User,
			ResourceType: activityLogEntryResourceTypeTeam,
			ResourceName: input.Slug.String(),
			TeamSlug:     ptr.To(input.Slug),
		})
	})
	if err != nil {
		return nil, err
	}

	return toGraphTeam(team), nil
}

func Update(ctx context.Context, input *UpdateTeamInput, actor *authz.Actor) (*Team, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	existingTeam, err := Get(ctx, input.Slug)
	if err != nil {
		return nil, err
	}

	if input.HasNoChanges() {
		return existingTeam, nil
	}

	if input.DisplayName == nil {
		input.DisplayName = &existingTeam.DisplayName
	}

	if input.SectionCode == nil {
		input.SectionCode = &existingTeam.SectionCode
	}

	var team *teamsql.Team
	err = database.Transaction(ctx, func(ctx context.Context) error {
		team, err = db(ctx).Update(ctx, teamsql.UpdateParams{
			DisplayName: input.DisplayName,
			Slug:        input.Slug,
			SectionCode: input.SectionCode,
		})
		if err != nil {
			return err
		}

		updatedFields := make([]*TeamUpdatedActivityLogEntryDataUpdatedField, 0)
		if input.DisplayName != nil && *input.DisplayName != existingTeam.DisplayName {
			updatedFields = append(updatedFields, &TeamUpdatedActivityLogEntryDataUpdatedField{
				Field:    "displayName",
				OldValue: &existingTeam.DisplayName,
				NewValue: input.DisplayName,
			})
		}
		if input.SectionCode != nil && *input.SectionCode != existingTeam.SectionCode {
			updatedFields = append(updatedFields, &TeamUpdatedActivityLogEntryDataUpdatedField{
				Field:    "sectionCode",
				OldValue: &existingTeam.SectionCode,
				NewValue: input.SectionCode,
			})
		}

		return activitylog.Create(ctx, activitylog.CreateInput{
			Action:       activitylog.ActivityLogEntryActionUpdated,
			Actor:        actor.User,
			ResourceType: activityLogEntryResourceTypeTeam,
			ResourceName: input.Slug.String(),
			TeamSlug:     ptr.To(input.Slug),
			Data: func(fields []*TeamUpdatedActivityLogEntryDataUpdatedField) *TeamUpdatedActivityLogEntryData {
				if len(fields) == 0 {
					return nil
				}

				return &TeamUpdatedActivityLogEntryData{
					UpdatedFields: fields,
				}
			}(updatedFields),
		})
	})
	if err != nil {
		return nil, err
	}

	return toGraphTeam(team), nil
}

func Get(ctx context.Context, slug slug.Slug) (*Team, error) {
	t, err := fromContext(ctx).teamLoader.Load(ctx, slug)
	if err != nil {
		return nil, handleError(err)
	}
	return t, nil
}

func GetByIdent(ctx context.Context, id ident.Ident) (*Team, error) {
	teamSlug, err := parseTeamIdent(id)
	if err != nil {
		return nil, err
	}
	return Get(ctx, teamSlug)
}

func List(ctx context.Context, page *pagination.Pagination, orderBy *TeamOrder) (*TeamConnection, error) {
	q := db(ctx)

	ret, err := q.List(ctx, teamsql.ListParams{
		Offset:  page.Offset(),
		Limit:   page.Limit(),
		OrderBy: orderBy.String(),
	})
	if err != nil {
		return nil, err
	}

	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}

	return pagination.NewConvertConnection(ret, page, total, func(from *teamsql.ListRow) *Team {
		return toGraphTeam(&from.Team)
	}), nil
}

func ListForUser(ctx context.Context, userID uuid.UUID, page *pagination.Pagination, orderBy *TeamOrder) (*TeamMemberConnection, error) {
	q := db(ctx)

	ret, err := q.ListForUser(ctx, teamsql.ListForUserParams{
		UserID:  userID,
		Offset:  page.Offset(),
		Limit:   page.Limit(),
		OrderBy: orderBy.String(),
	})
	if err != nil {
		return nil, err
	}

	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}
	return pagination.NewConvertConnection(ret, page, total, toGraphUserTeam), nil
}

func ListMembers(ctx context.Context, teamSlug slug.Slug, page *pagination.Pagination, orderBy *user.UserOrder) (*TeamMemberConnection, error) {
	q := db(ctx)

	ret, err := q.ListMembers(ctx, teamsql.ListMembersParams{
		TeamSlug: teamSlug,
		Offset:   page.Offset(),
		Limit:    page.Limit(),
		OrderBy:  orderBy.String(),
	})
	if err != nil {
		return nil, err
	}

	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}
	return pagination.NewConvertConnection(ret, page, total, toGraphTeamMember), nil
}

func GetDeleteKey(ctx context.Context, teamSlug slug.Slug, key uuid.UUID) (*TeamDeleteKey, error) {
	ret, err := db(ctx).GetDeleteKey(ctx, teamsql.GetDeleteKeyParams{
		Key:  key,
		Slug: teamSlug,
	})
	if err != nil {
		return nil, err
	}

	return toGraphTeamDeleteKey(ret), nil
}

func CreateDeleteKey(ctx context.Context, teamSlug slug.Slug, actor *authz.Actor) (*TeamDeleteKey, error) {
	var key *teamsql.TeamDeleteKey
	var err error
	err = database.Transaction(ctx, func(ctx context.Context) error {
		key, err = db(ctx).CreateDeleteKey(ctx, teamsql.CreateDeleteKeyParams{
			TeamSlug:  teamSlug,
			CreatedBy: actor.User.GetID(),
		})
		if err != nil {
			return err
		}

		return activitylog.Create(ctx, activitylog.CreateInput{
			Action:       activityLogEntryActionCreateDeleteKey,
			Actor:        actor.User,
			ResourceType: activityLogEntryResourceTypeTeam,
			ResourceName: teamSlug.String(),
			TeamSlug:     ptr.To(teamSlug),
		})
	})
	if err != nil {
		return nil, err
	}

	return toGraphTeamDeleteKey(key), nil
}

func ConfirmDeleteKey(ctx context.Context, teamSlug slug.Slug, deleteKey uuid.UUID, actor *authz.Actor) error {
	return database.Transaction(ctx, func(ctx context.Context) error {
		db := db(ctx)

		if err := db.ConfirmDeleteKey(ctx, deleteKey); err != nil {
			return err
		}

		if err := db.SetDeleteKeyConfirmedAt(ctx, teamSlug); err != nil {
			return err
		}

		return activitylog.Create(ctx, activitylog.CreateInput{
			Action:       activityLogEntryActionConfirmDeleteKey,
			Actor:        actor.User,
			ResourceType: activityLogEntryResourceTypeTeam,
			ResourceName: teamSlug.String(),
			TeamSlug:     ptr.To(teamSlug),
		})
	})
}

func UserIsOwner(ctx context.Context, teamSlug slug.Slug, userID uuid.UUID) (bool, error) {
	return db(ctx).UserIsOwner(ctx, teamsql.UserIsOwnerParams{
		UserID:   userID,
		TeamSlug: teamSlug,
	})
}

func UserIsMember(ctx context.Context, teamSlug slug.Slug, userID uuid.UUID) (bool, error) {
	return db(ctx).UserIsMember(ctx, teamsql.UserIsMemberParams{
		UserID:   userID,
		TeamSlug: teamSlug,
	})
}

func Count(ctx context.Context) (int64, error) {
	// This is only implemented for vulnerability ranking. This should soon be removed.
	count, err := db(ctx).List(ctx, teamsql.ListParams{
		Limit: 1,
	})
	if err != nil {
		return 0, err
	}
	if len(count) == 0 {
		return 0, nil
	}

	return count[0].TotalCount, nil
}

// Exists checks if an active team with the given slug exists.
func Exists(ctx context.Context, slug slug.Slug) (bool, error) {
	return db(ctx).Exists(ctx, slug)
}

func ListBySlugs(ctx context.Context, slugs []slug.Slug, page *pagination.Pagination) (*TeamConnection, error) {
	ret, err := db(ctx).ListBySlugs(ctx, slugs)
	if err != nil {
		return nil, err
	}

	p := pagination.Slice(ret, page)
	return pagination.NewConvertConnection(p, page, len(ret), toGraphTeam), nil
}

func ListAllSlugs(ctx context.Context) ([]slug.Slug, error) {
	return db(ctx).ListAllSlugs(ctx)
}
