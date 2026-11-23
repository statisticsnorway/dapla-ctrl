package teamfolders

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/database"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/teamfolders/teamfolderssql"
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
	internalQuerier *teamfolderssql.Queries
}

func newLoaders(dbConn *pgxpool.Pool) *loaders {
	return &loaders{
		internalQuerier: teamfolderssql.New(dbConn),
	}
}

func db(ctx context.Context) *teamfolderssql.Queries {
	l := fromContext(ctx)

	if tx := database.TransactionFromContext(ctx); tx != nil {
		return l.internalQuerier.WithTx(tx)
	}

	return l.internalQuerier
}
