package sharedbucketsstopgap

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/database/notify"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/search"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/sharedbucketsstopgap/sharedbucketsstopgapsql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
)

func AddSearch(client search.Client, pool *pgxpool.Pool, notifier *notify.Notifier, log logrus.FieldLogger) {
	client.AddClient("SHAREDBUCKET", &sharedBucketsSearch{
		db:       sharedbucketsstopgapsql.New(pool),
		notifier: notifier,
		log:      log,
	})
}

type sharedBucketsSearch struct {
	log      logrus.FieldLogger
	notifier *notify.Notifier
	db       sharedbucketsstopgapsql.Querier
}

func (g *sharedBucketsSearch) Convert(ctx context.Context, ids ...ident.Ident) ([]search.SearchNode, error) {
	bucketNames := make([]string, 0, len(ids))
	for _, id := range ids {
		bucketName, err := parseIdent(id)
		if err != nil {
			return nil, err
		}
		bucketNames = append(bucketNames, bucketName)
	}

	all, err := g.db.GetByNames(ctx, bucketNames)
	if err != nil {
		return nil, err
	}

	ret := make([]search.SearchNode, 0, len(all))
	for _, bucket := range all {
		ret = append(ret, toGraphBucket(bucket))
	}

	return ret, nil
}

func (g *sharedBucketsSearch) ReIndex(ctx context.Context) []search.Document {
	all, err := g.db.ListAllForSearch(ctx)
	if err != nil {
		return nil
	}

	ret := make([]search.Document, 0, len(all))
	for _, bucket := range all {
		ret = append(ret, newSearchDocument(bucket.Name, bucket.TeamSlug))
	}

	return ret
}

func (g *sharedBucketsSearch) Watch(ctx context.Context, indexer search.Indexer) error {
	go g.listen(ctx, indexer)
	return nil
}

func (g *sharedBucketsSearch) listen(ctx context.Context, indexer search.Indexer) {
	ch := g.notifier.Listen("shared_buckets_stopgap")

	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-ch:
			data := dataFromNotification(payload)
			if data.TeamSlug == "" {
				continue
			}

			switch payload.Op {
			case notify.Insert, notify.Update:
				indexer.Upsert(newSearchDocument(data.Name, data.TeamSlug))
			case notify.Delete:
				indexer.Remove(NewIdent(data.Name))
			default:
				g.log.WithField("op", payload.Op).Warn("unknown operation")
			}
		}
	}
}

type notificationData struct {
	Name     string    `json:"name"`
	TeamSlug slug.Slug `json:"teamSlug"`
}

func dataFromNotification(payload notify.Payload) notificationData {
	var bucketName string
	var slg slug.Slug

	if sslug, ok := payload.Data["team_slug"].(string); ok {
		slg = slug.Slug(sslug)
	}

	if sname, ok := payload.Data["name"].(string); ok {
		bucketName = sname
	}

	return notificationData{
		Name:     bucketName,
		TeamSlug: slg,
	}
}

func newSearchDocument(bucketName string, teamSlug slug.Slug) search.Document {
	return search.Document{
		ID:     NewIdent(bucketName).String(),
		Name:   bucketName,
		Team:   teamSlug.String(),
		Kind:   "SHAREDBUCKET",
		Fields: map[string]string{},
	}
}
