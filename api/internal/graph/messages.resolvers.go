package graph

import (
	"context"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/message"
	"github.com/statisticsnorway/dapla-api/internal/user"
)

func (r *messageResolver) MessageID(ctx context.Context, obj *message.Message) (string, error) {
	return obj.UUID.String(), nil
}

func (r *messageResolver) Recipient(ctx context.Context, obj *message.Message) (*user.User, error) {
	return user.Get(ctx, obj.Recipient)
}

func (r *mutationResolver) SendMessage(ctx context.Context, input message.SendMessageInput) (*message.SendMessagePayload, error) {
	actor := authz.ActorFromContext(ctx)

	if err := authz.CanSendMessage(ctx); err != nil {
		return nil, err
	}

	m, err := message.Create(ctx, &input, actor)
	if err != nil {
		return nil, err
	}

	return &message.SendMessagePayload{
		MessageID: m.UUID.String(),
	}, nil
}

func (r *queryResolver) Messages(ctx context.Context, first *int, after *pagination.Cursor, last *int, before *pagination.Cursor, filter *message.MessageFilter) (*pagination.Connection[*message.Message], error) {
	if err := authz.RequireGlobalAdmin(ctx); err != nil {
		return nil, err
	}

	page, err := pagination.ParsePage(first, after, last, before)
	if err != nil {
		return nil, err
	}
	return message.List(ctx, page, filter)
}

func (r *queryResolver) Message(ctx context.Context, messageID string) (*message.Message, error) {
	if err := authz.RequireGlobalAdmin(ctx); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(messageID)
	if err != nil {
		return nil, err
	}
	return message.Get(ctx, id)
}

func (r *Resolver) Message() gengql.MessageResolver { return &messageResolver{r} }

type messageResolver struct{ *Resolver }
