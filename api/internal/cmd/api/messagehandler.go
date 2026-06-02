package api

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/internal/leaderelection"
	"github.com/statisticsnorway/dapla-api/internal/message/messagesender"
	"golang.org/x/sync/errgroup"
)

const (
	sendMessageInterval = time.Second * 15
	sendMessageTimeout  = time.Minute
)

func runMessageHandling(ctx context.Context, pool *pgxpool.Pool, cfg *Config, log logrus.FieldLogger) error {
	if !cfg.Postman.Enabled {
		log.Warningf("mail sending is not enabled")
		return nil
	}

	if cfg.Postman.OutgoingSubscription == "" || cfg.Postman.IncomingTopic == "" {
		log.Errorf("env POSTMAN_OUTGOING_SUBSCRIPTION or POSTMAN_INCOMING_TOPIC is not set")
		return errors.New("missing envvar")
	}

	wg, ctx := errgroup.WithContext(ctx)

	sm, err := messagesender.NewFromConfig(ctx, pool, cfg.Postman.OutgoingSubscription, cfg.Postman.IncomingTopic, cfg.Postman.ProjectId, cfg.Postman.PublishingAppName, log)
	if err != nil {
		log.WithError(err).Errorf("unable to set up messagesender")
		return err
	}
	defer sm.Close()

	wg.Go(func() error {
		return sender(ctx, sm, cfg, log)
	})

	wg.Go(func() error {
		return sm.ReceiveMessages(ctx, cfg.LeaderElectionEnabled)
	})

	ch := make(chan error)
	go func() {
		ch <- wg.Wait()
	}()

	<-ctx.Done()

	select {
	case <-time.After(10 * time.Second):
		log.Warn("timed out waiting for graceful shutdown of messagehandler")
	case err := <-ch:
		return err
	}

	return nil
}

func sender(ctx context.Context, sm *messagesender.MessageSender, cfg *Config, log logrus.FieldLogger) error {
	for {
		if !cfg.LeaderElectionEnabled && !leaderelection.IsLeader(log) {
			log.Debug("not leader, skipping message sending")
			return nil
		}
		func() {
			ctx, cancel := context.WithTimeout(ctx, sendMessageTimeout)
			defer cancel()

			if err := sm.SendMessages(ctx); err != nil {
				log.WithError(err).Errorf("could not send messages")
			}
		}()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sendMessageInterval):
		}
	}
}
