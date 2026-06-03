package sharedbucketsstopgap

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/database"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/loader"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/sharedbucketsstopgap/sharedbucketsstopgapsql"
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
	internalQuerier *sharedbucketsstopgapsql.Queries
	bucketsLoader   *dataloadgen.Loader[string, *SharedBucket]
}

func newLoaders(dbConn *pgxpool.Pool) *loaders {
	db := sharedbucketsstopgapsql.New(dbConn)
	sectionLoader := &dataloader{db: db}

	return &loaders{
		internalQuerier: db,
		bucketsLoader:   dataloadgen.NewLoader(sectionLoader.list, loader.DefaultDataLoaderOptions...),
	}
}

type dataloader struct {
	db *sharedbucketsstopgapsql.Queries
}

func (l dataloader) list(ctx context.Context, sectionCodes []string) ([]*SharedBucket, []error) {
	makeKey := func(obj *SharedBucket) string { return obj.Name }
	return loader.LoadModels(ctx, sectionCodes, l.db.GetByNames, toGraphBucket, makeKey)
}

func db(ctx context.Context) *sharedbucketsstopgapsql.Queries {
	l := fromContext(ctx)

	if tx := database.TransactionFromContext(ctx); tx != nil {
		return l.internalQuerier.WithTx(tx)
	}

	return l.internalQuerier
}
