package message

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/graph/loader"
	"github.com/statisticsnorway/dapla-api/internal/message/messagesql"
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
	internalQuerier *messagesql.Queries
	messageLoader   *dataloadgen.Loader[uuid.UUID, *Message]
}

func newLoaders(dbConn *pgxpool.Pool) *loaders {
	db := messagesql.New(dbConn)
	messageLoader := &dataloader{db: db}

	return &loaders{
		internalQuerier: db,
		messageLoader:   dataloadgen.NewLoader(messageLoader.list, loader.DefaultDataLoaderOptions...),
	}
}

type dataloader struct {
	db messagesql.Querier
}

func (l dataloader) list(ctx context.Context, ids []uuid.UUID) ([]*Message, []error) {
	makeKey := func(obj *Message) uuid.UUID { return obj.UUID }
	return loader.LoadModels(ctx, ids, l.db.GetByIDs, toGraphMessage, makeKey)
}

func db(ctx context.Context) *messagesql.Queries {
	l := fromContext(ctx)

	if tx := database.TransactionFromContext(ctx); tx != nil {
		return l.internalQuerier.WithTx(tx)
	}

	return l.internalQuerier
}
