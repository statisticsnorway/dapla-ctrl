package group

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/internal/database/notify"
	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-api/internal/group/groupsql"
	"github.com/statisticsnorway/dapla-api/internal/search"
	"github.com/statisticsnorway/dapla-api/internal/slug"
)

func AddSearch(client search.Client, pool *pgxpool.Pool, notifier *notify.Notifier, log logrus.FieldLogger) {
	client.AddClient("GROUP", &groupSearch{
		db:       groupsql.New(pool),
		notifier: notifier,
		log:      log,
	})
}

type groupSearch struct {
	log      logrus.FieldLogger
	notifier *notify.Notifier
	db       groupsql.Querier
}

func (g *groupSearch) Convert(ctx context.Context, ids ...ident.Ident) ([]search.SearchNode, error) {
	groupNames := make([]string, 0, len(ids))
	for _, id := range ids {
		groupName, err := parseIdent(id)
		if err != nil {
			return nil, err
		}
		groupNames = append(groupNames, groupName)
	}

	all, err := g.db.GetByNames(ctx, groupNames)
	if err != nil {
		return nil, err
	}

	ret := make([]search.SearchNode, 0, len(all))
	for _, group := range all {
		ret = append(ret, toGraphGroup(group))
	}

	return ret, nil
}

func (g *groupSearch) ReIndex(ctx context.Context) []search.Document {
	all, err := g.db.ListAllForSearch(ctx)
	if err != nil {
		return nil
	}

	ret := make([]search.Document, 0, len(all))
	for _, group := range all {
		ret = append(ret, newSearchDocument(group.Name, group.TeamSlug))
	}

	return ret
}

func (g *groupSearch) Watch(ctx context.Context, indexer search.Indexer) error {
	go g.listen(ctx, indexer)
	return nil
}

func (g *groupSearch) listen(ctx context.Context, indexer search.Indexer) {
	ch := g.notifier.Listen("groups")

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
	var groupName string
	var slg slug.Slug

	if sslug, ok := payload.Data["team_slug"].(string); ok {
		slg = slug.Slug(sslug)
	}

	if sname, ok := payload.Data["name"].(string); ok {
		groupName = sname
	}

	return notificationData{
		Name:     groupName,
		TeamSlug: slg,
	}
}

func newSearchDocument(groupName string, teamSlug slug.Slug) search.Document {
	return search.Document{
		ID:     NewIdent(groupName).String(),
		Name:   groupName,
		Team:   teamSlug.String(),
		Kind:   "GROUP",
		Fields: map[string]string{},
	}
}
