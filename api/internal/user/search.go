package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/database/notify"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/search"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/user/usersql"
)

func AddSearch(client search.Client, pool *pgxpool.Pool, notifier *notify.Notifier, log logrus.FieldLogger) {
	client.AddClient("USER", &userSearch{
		db:       usersql.New(pool),
		notifier: notifier,
		log:      log,
	})
}

type userSearch struct {
	log      logrus.FieldLogger
	notifier *notify.Notifier
	db       *usersql.Queries
}

func (u *userSearch) Convert(ctx context.Context, ids ...ident.Ident) ([]search.SearchNode, error) {
	userIDs := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		userID, err := parseIdent(id)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	all, err := u.db.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	ret := make([]search.SearchNode, 0, len(all))
	for _, user := range all {
		ret = append(ret, toGraphUser(user))
	}

	return ret, nil
}

func (u *userSearch) ReIndex(ctx context.Context) []search.Document {
	all, err := u.db.ListAllForSearch(ctx)
	if err != nil {
		return nil
	}

	ret := make([]search.Document, 0, len(all))
	for _, user := range all {
		ret = append(ret, newSearchDocument(&user.User))
	}

	return ret
}

func (u *userSearch) Watch(ctx context.Context, indexer search.Indexer) error {
	go u.listen(ctx, indexer)
	return nil
}

func (u *userSearch) listen(ctx context.Context, indexer search.Indexer) {
	ch := u.notifier.Listen("users")

	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-ch:
			data := dataFromNotification(payload)
			if data.ID == uuid.Nil {
				continue
			}

			switch payload.Op {
			case notify.Insert, notify.Update:
				indexer.Upsert(newSearchDocument(&usersql.User{
					ID:    data.ID,
					Name:  data.Name,
					Email: data.Email,
				}))
			case notify.Delete:
				indexer.Remove(NewIdent(data.ID))
			default:
				u.log.WithField("op", payload.Op).Warn("unknown operation")
			}
		}
	}
}

type notificationData struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

func dataFromNotification(payload notify.Payload) notificationData {
	var id uuid.UUID
	var name, email string

	if uid, ok := payload.Data["id"].(string); ok {
		id, _ = uuid.Parse(uid)
	}

	if n, ok := payload.Data["name"].(string); ok {
		name = n
	}

	if e, ok := payload.Data["email"].(string); ok {
		email = e
	}

	return notificationData{
		ID:    id,
		Name:  name,
		Email: email,
	}
}

func newSearchDocument(user *usersql.User) search.Document {
	return search.Document{
		ID:     NewIdent(user.ID).String(),
		Name:   user.Name,
		Kind:   "USER",
		Fields: map[string]string{"email": user.Email},
	}
}
