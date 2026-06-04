package authz

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/auth/authz/authzsql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
)

func ListRoles(ctx context.Context, page *pagination.Pagination) (*RoleConnection, error) {
	q := db(ctx)

	ret, err := q.ListRoles(ctx, authzsql.ListRolesParams{
		Offset: page.Offset(),
		Limit:  page.Limit(),
	})
	if err != nil {
		return nil, err
	}

	total, err := q.CountRoles(ctx)
	if err != nil {
		return nil, err
	}
	return pagination.NewConvertConnection(ret, page, total, toGraphRole), nil
}

func ListRolesForServiceAccount(ctx context.Context, serviceAccountID uuid.UUID, page *pagination.Pagination) (*RoleConnection, error) {
	q := db(ctx)

	ret, err := q.ListRolesForServiceAccount(ctx, authzsql.ListRolesForServiceAccountParams{
		Offset:           page.Offset(),
		Limit:            page.Limit(),
		ServiceAccountID: serviceAccountID,
	})
	if err != nil {
		return nil, err
	}

	total, err := q.CountRolesForServiceAccount(ctx, serviceAccountID)
	if err != nil {
		return nil, err
	}
	return pagination.NewConvertConnection(ret, page, total, toGraphRole), nil
}

func getRoleByIdent(ctx context.Context, id ident.Ident) (*Role, error) {
	name, err := parseRoleIdent(id)
	if err != nil {
		return nil, err
	}

	row, err := db(ctx).GetRoleByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return toGraphRole(row), nil
}

func ForUser(ctx context.Context, userID uuid.UUID) ([]*Role, error) {
	ur, err := fromContext(ctx).userRoles.Load(ctx, userID)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return []*Role{}, nil
	} else if err != nil {
		return nil, err
	}
	return ur.Roles, nil
}

func ForServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) ([]*Role, error) {
	sar, err := fromContext(ctx).serviceAccountRoles.Load(ctx, serviceAccountID)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return []*Role{}, nil
	} else if err != nil {
		return nil, err
	}
	return sar.Roles, nil
}

func AssignRoleToServiceAccount(ctx context.Context, serviceAccountID uuid.UUID, roleName string) error {
	return db(ctx).AssignRoleToServiceAccount(ctx, authzsql.AssignRoleToServiceAccountParams{
		ServiceAccountID: serviceAccountID,
		RoleName:         roleName,
	})
}

func RevokeRoleFromServiceAccount(ctx context.Context, serviceAccountID uuid.UUID, roleName string) error {
	return db(ctx).RevokeRoleFromServiceAccount(ctx, authzsql.RevokeRoleFromServiceAccountParams{
		ServiceAccountID: serviceAccountID,
		RoleName:         roleName,
	})
}

func MakeUserTeamMember(ctx context.Context, userID uuid.UUID, teamSlug slug.Slug) error {
	return ErrNotSupported
	// return db(ctx).AssignTeamRoleToUser(ctx, authzsql.AssignTeamRoleToUserParams{
	// 	UserID:         userID,
	// 	RoleName:       "Team member",
	// 	TargetTeamSlug: teamSlug,
	// })
}

func MakeUserTeamOwner(ctx context.Context, userID uuid.UUID, teamSlug slug.Slug) error {
	return ErrNotSupported
	// return db(ctx).AssignTeamRoleToUser(ctx, authzsql.AssignTeamRoleToUserParams{
	// 	UserID:         userID,
	// 	RoleName:       "Team owner",
	// 	TargetTeamSlug: teamSlug,
	// })
}

func GetRole(ctx context.Context, name string) (*Role, error) {
	row, err := db(ctx).GetRoleByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return toGraphRole(row), nil
}

func ServiceAccountHasRole(ctx context.Context, serviceAccountID uuid.UUID, roleName string) (bool, error) {
	return db(ctx).ServiceAccountHasRole(ctx, authzsql.ServiceAccountHasRoleParams{
		ServiceAccountID: serviceAccountID,
		RoleName:         roleName,
	})
}

func CanAssignRole(ctx context.Context, roleName string, targetTeamSlug *slug.Slug) (bool, error) {
	err := RequireGlobalAdmin(ctx)
	return err == nil, err
	// if actor.User.IsServiceAccount() {
	// 	return serviceAccountCanAssignRole(ctx, actor.User.GetID(), roleName, targetTeamSlug)
	// }

	// return userCanAssignRole(ctx, actor.User.GetID(), roleName, targetTeamSlug)
}

/* func userCanAssignRole(ctx context.Context, userID uuid.UUID, roleName string, targetTeamSlug *slug.Slug) (bool, error) {
	return db(ctx).UserCanAssignRole(ctx, authzsql.UserCanAssignRoleParams{
		UserID:         userID,
		RoleName:       roleName,
		TargetTeamSlug: targetTeamSlug,
	})
} */

/* func serviceAccountCanAssignRole(ctx context.Context, serviceAccountID uuid.UUID, roleName string, targetTeamSlug *slug.Slug) (bool, error) {
	return db(ctx).ServiceAccountCanAssignRole(ctx, authzsql.ServiceAccountCanAssignRoleParams{
		ServiceAccountID: serviceAccountID,
		RoleName:         roleName,
		TeamSlug:         targetTeamSlug,
	})
} */

func CanCreateServiceAccounts(ctx context.Context, teamSlug *slug.Slug) error {
	return requireAuthorization(ctx, "service_accounts:create", teamSlug)
}

func CanUpdateServiceAccounts(ctx context.Context, teamSlug *slug.Slug) error {
	return requireAuthorization(ctx, "service_accounts:update", teamSlug)
}

func CanDeleteServiceAccounts(ctx context.Context, teamSlug *slug.Slug) error {
	return requireAuthorization(ctx, "service_accounts:delete", teamSlug)
}

func CanCreateTeam(ctx context.Context) error {
	user := ActorFromContext(ctx).User
	if user.IsAdmin() {
		return nil
	}
	if user.IsServiceAccount() {
		return ErrUnauthorized
	}

	// TODO: UNCOMMENT WHEN WE WANT TO OPEN FOR TEAM CREATION
	// id := user.GetID()
	// authorized, err := db(ctx).IsSectionManager(ctx, &id)
	// if err != nil {
	// 	return err
	// }
	// if authorized {
	// 	return nil
	// }
	return ErrUnauthorized
}

func CanCreateGroup(ctx context.Context, teamSlug slug.Slug) error {
	return RequireGlobalAdmin(ctx)
}

func CanManageTeamMembers(ctx context.Context, teamSlug slug.Slug) error {
	return ErrNotSupported
}

func CanManageGroupMembers(ctx context.Context, teamSlug slug.Slug) error {
	if err := requireTeamAuthorization(ctx, teamSlug, "teams:members:admin"); err == nil {
		return nil
	}

	return CanManageTeam(ctx, teamSlug)
}

func CanUpdateTeamMetadata(ctx context.Context, teamSlug slug.Slug) error {
	return CanManageTeam(ctx, teamSlug)
}

func CanDeleteTeam(ctx context.Context, teamSlug slug.Slug) error {
	return ErrNotSupported // TODO: deletion not yet supported CanManageTeam(ctx, teamSlug)
}

func CanManageTeam(ctx context.Context, teamSlug slug.Slug) error {
	user := ActorFromContext(ctx).User
	if user.IsAdmin() {
		return nil
	}
	if user.IsServiceAccount() {
		return ErrUnauthorized
	}

	id := user.GetID()
	authorized, err := db(ctx).IsManagerForTeamSection(ctx, authzsql.IsManagerForTeamSectionParams{
		TeamSlug: teamSlug,
		UserID:   &id,
	})
	if err != nil {
		return err
	}
	if authorized {
		return nil
	}
	return ErrUnauthorized
}

func CanSendMessage(ctx context.Context) error {
	return requireGlobalAuthorization(ctx, "messages:send")
}

func RequireGlobalAdmin(ctx context.Context) error {
	if ActorFromContext(ctx).User.IsAdmin() {
		return nil
	}

	return ErrUnauthorized
}

func requireTeamAuthorization(ctx context.Context, teamSlug slug.Slug, authorizationName string) error {
	user := ActorFromContext(ctx).User
	var (
		hasAuthorization bool
		err              error
	)
	if user.IsServiceAccount() {
		hasAuthorization, err = db(ctx).ServiceAccountHasAuthorization(ctx, authzsql.ServiceAccountHasAuthorizationParams{
			ServiceAccountID:  user.GetID(),
			AuthorizationName: authorizationName,
		})
	} else {
		hasAuthorization, err = db(ctx).HasTeamAuthorization(ctx, authzsql.HasTeamAuthorizationParams{
			UserID:            user.GetID(),
			AuthorizationName: authorizationName,
			TeamSlug:          teamSlug,
		})
	}
	if err != nil {
		return err
	}

	if hasAuthorization {
		return nil
	}

	return newMissingAuthorizationError(authorizationName)
}

func requireGlobalAuthorization(ctx context.Context, authorizationName string) error {
	user := ActorFromContext(ctx).User
	var (
		authorized bool
		err        error
	)
	if user.IsServiceAccount() {
		authorized, err = db(ctx).ServiceAccountHasAuthorization(ctx, authzsql.ServiceAccountHasAuthorizationParams{
			ServiceAccountID:  user.GetID(),
			AuthorizationName: authorizationName,
		})
	} else {
		authorized, err = db(ctx).HasGlobalAuthorization(ctx, authzsql.HasGlobalAuthorizationParams{
			UserID:            user.GetID(),
			AuthorizationName: authorizationName,
		})
	}
	if err != nil {
		return err
	}

	if authorized {
		return nil
	}

	return newMissingAuthorizationError(authorizationName)
}

func requireAuthorization(ctx context.Context, authorizationName string, teamSlug *slug.Slug) error {
	if teamSlug == nil {
		return requireGlobalAuthorization(ctx, authorizationName)
	}

	return requireTeamAuthorization(ctx, *teamSlug, authorizationName)
}
