package artifactregistry

import (
	"context"
	"strings"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/activitylog"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/artifactregistry/artifactregistrysql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/database"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/apierror"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
)

func getByIdent(_ context.Context, id ident.Ident) (*ArtifactRegistryGithubRepoAccess, error) {
	ts, githubRepositoryName, err := parseIdent(id)
	if err != nil {
		return nil, err
	}

	return &ArtifactRegistryGithubRepoAccess{
		TeamSlug: ts,
		Name:     githubRepositoryName,
	}, nil
}

func AddGithubRepositoryToTeam(ctx context.Context, input GrantGithubRepoAccessToTeamArtifactRegistryInput, actor *authz.Actor) (*ArtifactRegistryGithubRepoAccess, error) {
	containsOrg := strings.Contains(input.RepositoryName, "/")
	if containsOrg {
		return nil, apierror.Errorf("Repository name should not contain organisation. E.g. `myrepo` (instead of `statisticsnorway/myrepo`)")
	}

	q := db(ctx)
	var repo *artifactregistrysql.TeamArtifactRegistryGhReposAllowList
	err := database.Transaction(ctx, func(ctx context.Context) error {
		var err error
		repo, err = q.AddGithubRepositoryToTeam(ctx, artifactregistrysql.AddGithubRepositoryToTeamParams{
			TeamSlug:       input.TeamSlug,
			RepositoryName: input.RepositoryName,
		})
		if err != nil {
			return err
		}

		return activitylog.Create(ctx, activitylog.CreateInput{
			Action:       activitylog.ActivityLogEntryActionAdded,
			Actor:        actor.User,
			ResourceType: activityLogEntryResourceTypeArtifactRegistryGithubRepoAccess,
			ResourceName: input.RepositoryName,
			TeamSlug:     new(input.TeamSlug),
		})
	})
	if err != nil {
		return nil, err
	}

	return toGraphArtifactRegistryGithubRepoAccess(repo), nil
}

func RemoveGithubRepositoryFromTeam(ctx context.Context, input RevokeGithubRepoAccessFromTeamArtifactRegistryInput, actor *authz.Actor) error {
	q := db(ctx)
	return database.Transaction(ctx, func(ctx context.Context) error {
		err := q.RemoveGithubRepositoryFromTeam(ctx, artifactregistrysql.RemoveGithubRepositoryFromTeamParams{
			TeamSlug:       input.TeamSlug,
			RepositoryName: input.RepositoryName,
		})
		if err != nil {
			return err
		}

		return activitylog.Create(ctx, activitylog.CreateInput{
			Action:       activitylog.ActivityLogEntryActionRemoved,
			Actor:        actor.User,
			ResourceType: activityLogEntryResourceTypeArtifactRegistryGithubRepoAccess,
			ResourceName: input.RepositoryName,
			TeamSlug:     new(input.TeamSlug),
		})
	})
}

func ListForTeam(ctx context.Context, teamSlug slug.Slug, page *pagination.Pagination) (*ArtifactRegistryGithubRepoAccessConnection, error) {
	q := db(ctx)

	ret, err := q.ListGithubReposForTeam(ctx, artifactregistrysql.ListGithubReposForTeamParams{
		TeamSlug: teamSlug,
		Offset:   page.Offset(),
		Limit:    page.Limit(),
	})
	if err != nil {
		return nil, err
	}
	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}
	return pagination.NewConvertConnection(ret, page, total, func(from *artifactregistrysql.ListGithubReposForTeamRow) *ArtifactRegistryGithubRepoAccess {
		return toGraphArtifactRegistryGithubRepoAccess(&from.TeamArtifactRegistryGhReposAllowList)
	}), nil
}
