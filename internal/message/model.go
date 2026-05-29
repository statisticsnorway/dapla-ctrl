package message

import (
	"context"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/message/messagesql"
	"github.com/statisticsnorway/dapla-api/internal/validate"
)

type (
	MessageConnection = pagination.Connection[*Message]
	MessageEdge       = pagination.Edge[*Message]
)

type Message struct {
	UUID      uuid.UUID `json:"id"`
	Actor     string    `json:"actor"`
	Recipient uuid.UUID `json:"recipient"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
}

func (Message) IsNode() {}

func (m Message) ID() ident.Ident {
	return newMessageIdent(m.UUID)
}

func toGraphMessage(m *messagesql.Message) *Message {
	ret := &Message{
		UUID:      m.ID,
		Actor:     m.Actor,
		Recipient: m.Recipient,
		Subject:   m.Subject,
		Message:   m.Message,
		Status:    m.Status,
	}
	return ret
}

type MessageFilter struct {
	// Filter by message status, e.g PENDING, PUBLISHED, SUCCESSFUL and FAILED
	Status *string `json:"status,omitempty"`
	// Filter by message actor
	Actor *string `json:"actor,omitempty"`
	// Filter by message recipient
	Recipient *string `json:"recipient,omitempty"`
}

type SendMessageInput struct {
	// Recipient of the email
	//
	// Have to be a valid email address
	Recipient string `json:"recipient"`
	// Subject of the email
	Subject string `json:"subject"`
	// Message body of the email
	Message string `json:"message"`
}

func (i *SendMessageInput) Validate(ctx context.Context) error {
	verr := validate.New()

	if i.Subject == "" {
		verr.Add("Subject", "Subject can not be empty.")
	}

	if i.Message == "" {
		verr.Add("Message", "Message can not be empty.")
	}

	if exists, err := db(ctx).UserExists(ctx, i.Recipient); err != nil {
		return err
	} else if !exists {
		verr.Add("Recipient", "User does not exists.")
	}

	return verr.NilIfEmpty()
}

type SendMessagePayload struct {
	// Unique message id
	MessageID string `json:"messageId"`
}
