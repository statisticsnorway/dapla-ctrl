package messagesender

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/internal/leaderelection"
	"github.com/statisticsnorway/dapla-api/internal/message/messagesql"
)

const (
	unsentStatus  = "PENDING"
	sentStatus    = "PUBLISHED"
	successStatus = "SUCCESSFUL"
	failedStatus  = "FAILED"
)

type MessageSender struct {
	pool                        *pgxpool.Pool
	querier                     *messagesql.Queries
	postmanIncomingTopic        string // Outgoing for dapla-api, incoming for SUP Postman
	postmanOutgoingSubscription string // Incoming for dapla-api, outgoing for SUP Postman
	postmanProjectId            string
	postmanPublishingAppName    string
	pubsubClient                *pubsub.Client

	log logrus.FieldLogger
}

type TopicEntry struct {
	Data     *MessageRequest  `json:"data"`
	Metadata *MessageMetadata `json:"metadata,omitempty"`
}

type MessageMetadata struct{}

type MessageRequest struct {
	Id             uuid.UUID     `json:"id"`
	MessageChannel string        `json:"messageChannel"`
	EmailRequest   *EmailRequest `json:"emailRequest,omitempty"`
	SMSRequest     *SMSRequest   `json:"smsRequest,omitempty"`
}

type EmailRequest struct {
	Subject         string `json:"subject"`
	Message         string `json:"message"`
	Recipient       string `json:"receiverEmailAddress"`
	FromType        string `json:"fromType"`
	FromDisplayName string `json:"fromDisplayName"`
	IncludeLogo     bool   `json:"includeLogo"`
}

type SMSRequest struct {
	Message     string `json:"message"`
	PhoneNumber string `json:"phoneNumber"`
}

type MessageResult struct {
	Id        uuid.UUID
	Result    string
	Timestamp string
}

func New(pool *pgxpool.Pool, postmanOutgoingSubscription, postmanIncomingTopic, postmanProjectId, postmanPublishingAppName string, pubsubClient *pubsub.Client, log logrus.FieldLogger) *MessageSender {
	return &MessageSender{
		pool:                        pool,
		querier:                     messagesql.New(pool),
		postmanOutgoingSubscription: postmanOutgoingSubscription,
		postmanIncomingTopic:        postmanIncomingTopic,
		postmanProjectId:            postmanProjectId,
		postmanPublishingAppName:    postmanPublishingAppName,
		pubsubClient:                pubsubClient,
		log:                         log,
	}
}

func NewFromConfig(ctx context.Context, pool *pgxpool.Pool, postmanOutgoingSubscription, postmanIncomingTopic, postmanProjectId, postmanPublishingAppName string, log logrus.FieldLogger) (*MessageSender, error) {
	pubsubClient, err := pubsub.NewClient(ctx, postmanProjectId)
	if err != nil {
		return nil, err
	}

	return New(pool, postmanOutgoingSubscription, postmanIncomingTopic, postmanProjectId, postmanPublishingAppName, pubsubClient, log), nil
}

func (m *MessageSender) Close() error {
	m.log.Debug("closing pubsub")
	return m.pubsubClient.Close()
}

func (m *MessageSender) SendMessages(ctx context.Context) error {
	unsentMessages, _ := m.querier.GetByStatus(ctx, unsentStatus)

	if len(unsentMessages) == 0 {
		return nil
	}

	publisher := m.pubsubClient.Publisher(m.postmanIncomingTopic)

	for _, rawMessage := range unsentMessages {
		recipient, err := m.querier.GetUserByID(ctx, rawMessage.Recipient)
		if err != nil {
			m.log.Errorf("could not get recipient: %s", err)
		}
		message := TopicEntry{
			Data: &MessageRequest{
				Id:             rawMessage.ID,
				MessageChannel: "EMAIL",
				EmailRequest: &EmailRequest{
					Subject:         rawMessage.Subject,
					Message:         rawMessage.Message,
					Recipient:       recipient.Email,
					FromType:        "NO_REPLY",
					FromDisplayName: "Dapla Ctrl",
					IncludeLogo:     false,
				},
			},
		}
		byteMessage, err := json.Marshal(message)
		if err != nil {
			m.log.Errorf("could not marshal message: %s", message)
			return err
		}
		pubsubMessage := pubsub.Message{
			Data: byteMessage,
			Attributes: map[string]string{
				"PUBLISHER_APP_NAME": m.postmanPublishingAppName,
			},
		}

		err = Send(ctx, m.pool, *m.querier, publisher, pubsubMessage, message.Data.Id, m.log)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MessageSender) ReceiveMessages(ctx context.Context, leaderElectionEnabled bool) error {
	subscriber := m.pubsubClient.Subscriber(m.postmanOutgoingSubscription)

	err := subscriber.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		if !leaderElectionEnabled && !leaderelection.IsLeader(m.log) {
			m.log.Debug("not leader, skipping message recieving")
			return
		}
		var messageResult MessageResult
		if err := json.Unmarshal(msg.Data, &messageResult); err != nil {
			msg.Nack()
			m.log.Errorf("could not unmarshal message result %v", err)
			return
		}

		if err := UpdateMessageStatus(ctx, m.pool, *m.querier, messageResult.Id, messageResult.Result, m.log); err != nil {
			msg.Nack()
			m.log.Errorf("could not update message status %v", err)
			return
		}
		msg.Ack()
		m.log.Debugf("received message with id %v", messageResult.Id)
	})
	if err != nil && err != context.Canceled {
		m.log.Errorf("could not receive message %v", err)
	}

	return nil
}

func Send(ctx context.Context, pool *pgxpool.Pool, querier messagesql.Queries, publisher *pubsub.Publisher, message pubsub.Message, messageId uuid.UUID, log logrus.FieldLogger) error {
	result := publisher.Publish(ctx, &message)
	_, err := result.Get(ctx)
	if err != nil {
		log.Errorf("could not get pubsub result: %v", err)
	}
	log.Debugf("published message with id %v", messageId)
	err = UpdateMessageStatus(ctx, pool, querier, messageId, sentStatus, log)
	if err != nil {
		return err
	}

	return nil
}

func UpdateMessageStatus(ctx context.Context, pool *pgxpool.Pool, querier messagesql.Queries, messageId uuid.UUID, status string, log logrus.FieldLogger) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err == nil {
			return
		} else if !errors.Is(err, pgx.ErrTxClosed) {
			log.WithError(err).Errorf("rollback transaction")
		}
	}()

	txQuerier := querier.WithTx(tx)
	_, err = txQuerier.UpdateStatus(ctx, messagesql.UpdateStatusParams{
		Status: status,
		ID:     messageId,
	})
	if err != nil {
		log.Errorf("could not update message status of id: %v", messageId)
		return err
	}

	return tx.Commit(ctx)
}
