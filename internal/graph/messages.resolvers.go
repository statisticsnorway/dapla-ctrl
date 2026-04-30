package graph

import (
	"context"

	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/message"
)

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
		MessageID: m.Id.String(),
	}, nil
}
