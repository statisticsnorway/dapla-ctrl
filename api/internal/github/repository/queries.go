package repository

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/activitylog"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/database"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/github/repository/repositorysql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
)

func getByIdent(_ context.Context, id ident.Ident) (*Repository, error) {
	ts, repo, err := parseIdent(id)
	if err != nil {
		return nil, err
	}

	return &Repository{
		TeamSlug: ts,
		Name:     repo,
	}, nil
}

func ListForTeam(ctx context.Context, teamSlug slug.Slug, page *pagination.Pagination, orderBy *RepositoryOrder, filter *TeamRepositoryFilter) (*RepositoryConnection, error) {
	if filter == nil {
		filter = &TeamRepositoryFilter{}
	}

	q := db(ctx)

	ret, err := q.ListForTeam(ctx, repositorysql.ListForTeamParams{
		TeamSlug: teamSlug,
		Offset:   page.Offset(),
		Limit:    page.Limit(),
		Search:   filter.Name,
		OrderBy:  orderBy.String(),
	})
	if err != nil {
		return nil, err
	}
	var total int64
	if len(ret) > 0 {
		total = ret[0].TotalCount
	}
	return pagination.NewConvertConnection(ret, page, total, func(from *repositorysql.ListForTeamRow) *Repository {
		return toGraphRepository(&from.TeamRepository)
	}), nil
}

func AddToTeam(ctx context.Context, input AddRepositoryToTeamInput) (*Repository, error) {
	var tr *repositorysql.TeamRepository
	err := database.Transaction(ctx, func(ctx context.Context) error {
		var err error
		tr, err = db(ctx).AddToTeam(ctx, repositorysql.AddToTeamParams{
			TeamSlug:         input.TeamSlug,
			GithubRepository: input.RepositoryName,
		})
		if err != nil {
			return err
		}

		return activitylog.Create(ctx, activitylog.CreateInput{
			Action:       activitylog.ActivityLogEntryActionAdded,
			Actor:        authz.ActorFromContext(ctx).User,
			ResourceType: activityLogEntryResourceTypeRepository,
			ResourceName: input.RepositoryName,
			TeamSlug:     new(input.TeamSlug),
		})
	})
	if err != nil {
		return nil, err
	}

	return toGraphRepository(tr), nil
}

func RemoveFromTeam(ctx context.Context, input RemoveRepositoryFromTeamInput) error {
	return database.Transaction(ctx, func(ctx context.Context) error {
		err := db(ctx).RemoveFromTeam(ctx, repositorysql.RemoveFromTeamParams{
			TeamSlug:         input.TeamSlug,
			GithubRepository: input.RepositoryName,
		})
		if err != nil {
			return err
		}

		return activitylog.Create(ctx, activitylog.CreateInput{
			Action:       activitylog.ActivityLogEntryActionRemoved,
			Actor:        authz.ActorFromContext(ctx).User,
			ResourceType: activityLogEntryResourceTypeRepository,
			ResourceName: input.RepositoryName,
			TeamSlug:     new(input.TeamSlug),
		})
	})
}

func GetByName(ctx context.Context, name string) ([]*Repository, error) {
	repos, err := db(ctx).GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	ret := make([]*Repository, 0, len(repos))
	for _, r := range repos {
		ret = append(ret, toGraphRepository(r))
	}
	return ret, nil
}
