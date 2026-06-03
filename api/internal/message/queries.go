package message

import (
	"context"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/database"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/message/messagesql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/user"
)

func Create(ctx context.Context, input *SendMessageInput, actor *authz.Actor) (*Message, error) {
	if err := input.Validate(ctx); err != nil {
		return nil, err
	}

	recipientUser, err := user.GetByEmail(ctx, input.Recipient)
	if err != nil {
		return nil, err
	}

	var message *messagesql.Message
	err = database.Transaction(ctx, func(ctx context.Context) error {
		var err error
		message, err = db(ctx).Create(ctx, messagesql.CreateParams{
			Actor:     actor.User.Identity(),
			Recipient: recipientUser.UUID,
			Subject:   input.Subject,
			Message:   input.Message,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return toGraphMessage(message), nil
}

func Get(ctx context.Context, id uuid.UUID) (*Message, error) {
	t, err := fromContext(ctx).messageLoader.Load(ctx, id)
	if err != nil {
		return nil, handleError(err)
	}
	return t, nil
}

func GetByIdent(ctx context.Context, id ident.Ident) (*Message, error) {
	uuid, err := parseMessageIdent(id)
	if err != nil {
		return nil, err
	}
	return Get(ctx, uuid)
}

func List(ctx context.Context, page *pagination.Pagination, filter *MessageFilter) (*MessageConnection, error) {
	q := db(ctx)

	param := messagesql.ListParams{
		Offset: page.Offset(),
		Limit:  page.Limit(),
	}

	if filter != nil {
		if filter.Actor != nil {
			param.Actor = filter.Actor
		}

		if filter.Status != nil {
			param.Status = filter.Status
		}
		if filter.Recipient != nil {
			recipientUser, err := q.GetUserByEmail(ctx, *filter.Recipient)
			if err != nil {
				return nil, err
			}
			param.Recipient = &recipientUser.ID
		}
	}

	ret, err := q.List(ctx, param)
	if err != nil {
		return nil, err
	}

	total := 0
	if len(ret) > 0 {
		total = int(ret[0].TotalCount)
	}

	return pagination.NewConvertConnection(ret, page, total, func(from *messagesql.ListRow) *Message {
		return toGraphMessage(&from.Message)
	}), nil
}
