package group

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/graph/loader"
	"github.com/statisticsnorway/dapla-api/internal/group/groupsql"
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
	internalQuerier *groupsql.Queries
	groupLoader     *dataloadgen.Loader[string, *Group]
	log             logrus.FieldLogger
}

func newLoaders(dbConn *pgxpool.Pool, log logrus.FieldLogger) *loaders {
	db := groupsql.New(dbConn)
	groupLoader := &dataloader{db: db}

	return &loaders{
		internalQuerier: db,
		groupLoader:     dataloadgen.NewLoader(groupLoader.list, loader.DefaultDataLoaderOptions...),
		log:             log,
	}
}

type dataloader struct {
	db *groupsql.Queries
}

func (l dataloader) list(ctx context.Context, groupNames []string) ([]*Group, []error) {
	makeKey := func(obj *Group) string { return obj.Name }
	return loader.LoadModels(ctx, groupNames, l.db.GetByNames, toGraphGroup, makeKey)
}

func db(ctx context.Context) *groupsql.Queries {
	l := fromContext(ctx)

	if tx := database.TransactionFromContext(ctx); tx != nil {
		return l.internalQuerier.WithTx(tx)
	}

	return l.internalQuerier
}
