package artifactregistry

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/artifactregistry/artifactregistrysql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/database"
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
	internalQuerier *artifactregistrysql.Queries
}

func newLoaders(dbConn *pgxpool.Pool) *loaders {
	return &loaders{
		internalQuerier: artifactregistrysql.New(dbConn),
	}
}

func db(ctx context.Context) *artifactregistrysql.Queries {
	l := fromContext(ctx)

	if tx := database.TransactionFromContext(ctx); tx != nil {
		return l.internalQuerier.WithTx(tx)
	}

	return l.internalQuerier
}
