package team

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/database/notify"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/search"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team/teamsql"
)

func AddSearch(client search.Client, pool *pgxpool.Pool, notifier *notify.Notifier, log logrus.FieldLogger) {
	client.AddClient("TEAM", &teamSearch{
		db:       teamsql.New(pool),
		notifier: notifier,
		log:      log,
	})
}

type teamSearch struct {
	log      logrus.FieldLogger
	notifier *notify.Notifier
	db       teamsql.Querier
}

func (t *teamSearch) Convert(ctx context.Context, ids ...ident.Ident) ([]search.SearchNode, error) {
	slugs := make([]slug.Slug, 0, len(ids))
	for _, id := range ids {
		slug, err := parseTeamIdent(id)
		if err != nil {
			return nil, err
		}
		slugs = append(slugs, slug)
	}

	all, err := t.db.ListBySlugs(ctx, slugs)
	if err != nil {
		return nil, err
	}

	ret := make([]search.SearchNode, 0, len(all))
	for _, team := range all {
		ret = append(ret, toGraphTeam(team))
	}

	return ret, nil
}

func (t *teamSearch) ReIndex(ctx context.Context) []search.Document {
	all, err := t.db.ListAllForSearch(ctx)
	if err != nil {
		return nil
	}

	ret := make([]search.Document, 0, len(all))
	for _, team := range all {
		ret = append(ret, newSearchDocument(team.Slug, team.SectionCode))
	}

	return ret
}

func (t *teamSearch) Watch(ctx context.Context, indexer search.Indexer) error {
	go t.listen(ctx, indexer)
	return nil
}

func (t *teamSearch) listen(ctx context.Context, indexer search.Indexer) {
	ch := t.notifier.Listen("teams")

	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-ch:
			data := dataFromNotification(payload)
			if data.Slug == "" {
				continue
			}

			switch payload.Op {
			case notify.Insert, notify.Update:
				indexer.Upsert(newSearchDocument(data.Slug, data.SectionCode))
			case notify.Delete:
				indexer.Remove(newTeamIdent(data.Slug))
			default:
				t.log.WithField("op", payload.Op).Warn("unknown operation")
			}
		}
	}
}

type notificationData struct {
	Slug        slug.Slug `json:"slug"`
	SectionCode string    `json:"sectionCode"`
	IsManaged   bool      `json:"isManaged"`
}

func dataFromNotification(payload notify.Payload) notificationData {
	var slg slug.Slug
	var sectionCode string

	if sslug, ok := payload.Data["slug"].(string); ok {
		slg = slug.Slug(sslug)
	}

	if ssectionCode, ok := payload.Data["section_code"].(string); ok {
		sectionCode = ssectionCode
	}

	return notificationData{
		Slug:        slg,
		SectionCode: sectionCode,
	}
}

func newSearchDocument(teamSlug slug.Slug, sectionCode string) search.Document {
	sslug := teamSlug.String()
	return search.Document{
		ID:   newTeamIdent(teamSlug).String(),
		Name: sslug,
		Team: sslug,
		Kind: "TEAM",
		Fields: map[string]string{
			"sectionCode": sectionCode,
		},
	}
}
