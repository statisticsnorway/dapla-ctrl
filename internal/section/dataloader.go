package section

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/graph/loader"
	"github.com/statisticsnorway/dapla-api/internal/section/sectionsql"
	"github.com/vikstrous/dataloadgen"
)

type ctxKey int

const loadersKey ctxKey = iota

func NewLoaderContext(ctx context.Context, dbConn *pgxpool.Pool) context.Context {
	return context.WithValue(ctx, loadersKey, newLoaders(dbConn))
}

func fromContext(ctx context.Context) *loaders {
	return ctx.Value(loadersKey).(*loaders)
}

type loaders struct {
	internalQuerier *sectionsql.Queries
	sectionLoader   *dataloadgen.Loader[string, *Section]
}

func newLoaders(dbConn *pgxpool.Pool) *loaders {
	db := sectionsql.New(dbConn)
	sectionLoader := &dataloader{db: db}

	return &loaders{
		internalQuerier: db,
		sectionLoader:   dataloadgen.NewLoader(sectionLoader.list, loader.DefaultDataLoaderOptions...),
	}
}

type dataloader struct {
	db *sectionsql.Queries
}

func (l dataloader) list(ctx context.Context, sectionCodes []string) ([]*Section, []error) {
	makeKey := func(obj *Section) string { return obj.Code }
	return loader.LoadModels(ctx, sectionCodes, l.db.GetByCodes, toGraphUser, makeKey)
}

func db(ctx context.Context) *sectionsql.Queries {
	l := fromContext(ctx)

	if tx := database.TransactionFromContext(ctx); tx != nil {
		return l.internalQuerier.WithTx(tx)
	}

	return l.internalQuerier
}
