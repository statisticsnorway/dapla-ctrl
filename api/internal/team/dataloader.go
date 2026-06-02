package team

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/graph/loader"
	"github.com/statisticsnorway/dapla-api/internal/slug"
	"github.com/statisticsnorway/dapla-api/internal/team/teamsql"
	"github.com/vikstrous/dataloadgen"
)

type ctxKey int

const loadersKey ctxKey = iota

func NewLoaderContext(ctx context.Context, dbConn *pgxpool.Pool, log logrus.FieldLogger) context.Context {
	return context.WithValue(ctx, loadersKey, newLoaders(dbConn, log))
}

func fromContext(ctx context.Context) *loaders {
	return ctx.Value(loadersKey).(*loaders)
}

type loaders struct {
	internalQuerier *teamsql.Queries
	teamLoader      *dataloadgen.Loader[slug.Slug, *Team]
	log             logrus.FieldLogger
}

func newLoaders(dbConn *pgxpool.Pool, log logrus.FieldLogger) *loaders {
	db := teamsql.New(dbConn)
	teamLoader := &dataloader{db: db}

	return &loaders{
		internalQuerier: db,
		teamLoader:      dataloadgen.NewLoader(teamLoader.list, loader.DefaultDataLoaderOptions...),
		log:             log,
	}
}

type dataloader struct {
	db teamsql.Querier
}

func (l dataloader) list(ctx context.Context, slugs []slug.Slug) ([]*Team, []error) {
	makeKey := func(obj *Team) slug.Slug { return obj.Slug }
	return loader.LoadModels(ctx, slugs, l.db.ListBySlugs, toGraphTeam, makeKey)
}

func db(ctx context.Context) *teamsql.Queries {
	l := fromContext(ctx)

	if tx := database.TransactionFromContext(ctx); tx != nil {
		return l.internalQuerier.WithTx(tx)
	}

	return l.internalQuerier
}
