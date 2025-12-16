package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/internal/auth/authn"
	"github.com/statisticsnorway/dapla-api/internal/auth/middleware"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/database/notify"
	"github.com/statisticsnorway/dapla-api/internal/graph"
	"github.com/statisticsnorway/dapla-api/internal/graph/gengql"
	"github.com/statisticsnorway/dapla-api/internal/grpc"
	"github.com/statisticsnorway/dapla-api/internal/logger"
	"golang.org/x/sync/errgroup"
)

const (
	exitCodeSuccess = iota
	exitCodeLoggerError
	exitCodeRunError
	exitCodeConfigError
	exitCodeEnvFileError
)

func Run(ctx context.Context) {
	log := logrus.StandardLogger()

	if err := loadEnvFile(log); err != nil {
		log.WithError(err).Errorf("error loading .env file")
		os.Exit(exitCodeEnvFileError)
	}

	if _, ok := os.LookupEnv("WITH_FAKE_CLIENTS"); ok {
		log.Errorf("WITH_FAKE_CLIENTS should no longer be used. Update your .env file or environment variables.")
		log.Errorf("See .env.example for new environment variables.")
		os.Exit(1)
	}

	cfg, err := NewConfig(ctx, envconfig.OsLookuper())
	if err != nil {
		log.WithError(err).Errorf("error when processing configuration")
		os.Exit(exitCodeConfigError)
	}

	appLogger, err := logger.New(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		log.WithError(err).Errorf("error when creating application logger")
		os.Exit(exitCodeLoggerError)
	}

	cfg.Fakes.Inform(appLogger)

	err = run(ctx, cfg, appLogger)
	if err != nil {
		appLogger.WithError(err).Errorf("error in run()")
		os.Exit(exitCodeRunError)
	}

	os.Exit(exitCodeSuccess)
}

func run(ctx context.Context, cfg *Config, log logrus.FieldLogger) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx, signalStop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer signalStop()

	dbSettings := []database.OptFunc{}
	if cfg.WithSlowQueryLogger {
		dbSettings = append(dbSettings, database.WithSlowQueryLogger(10*time.Millisecond))
	}

	_, promReg, err := newMeterProvider(ctx)
	if err != nil {
		return fmt.Errorf("create metric meter: %w", err)
	}

	log.Info("connecting to database")
	pool, err := database.New(ctx, cfg.DatabaseConnectionString, log.WithField("subsystem", "database"), dbSettings...)
	if err != nil {
		return fmt.Errorf("setting up database: %w", err)
	}
	defer pool.Close()

	pubsubClient, err := pubsub.NewClient(ctx, cfg.GoogleManagementProjectID)
	if err != nil {
		return err
	}
	pubsubTopic := pubsubClient.Topic("dapla-api")

	graphHandler, err := graph.NewHandler(gengql.Config{
		Resolvers: graph.NewResolver(
			&graph.TopicWrapper{Topic: pubsubTopic},
			graph.WithLogger(log),
		),
		Complexity: gengql.NewComplexityRoot(),
	}, log)
	if err != nil {
		return fmt.Errorf("create graph handler: %w", err)
	}

	authHandler, err := setupAuthHandler(ctx, cfg.OAuth, log)
	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)

	// Notifier to use only one connection to the database for LISTEN/NOTIFY pattern
	notifier := notify.New(pool, log)
	go notifier.Run(ctx)

	var jwtMiddleware func(next http.Handler) http.Handler
	if !cfg.JWT.SkipMiddleware {
		jwtMiddleware, err = middleware.JWTAuthentication(ctx, cfg.JWT.Issuer, cfg.JWT.Audience, cfg.JWT.EmailClaim, log.WithField("subsystem", "jwt"))
		if err != nil {
			return fmt.Errorf("failed to create JWT authentication middleware: %w", err)
		}
	}

	// HTTP server
	wg.Go(func() error {
		return runHttpServer(
			ctx,
			cfg.Fakes,
			cfg.ListenAddress,
			pool,
			authHandler,
			jwtMiddleware,
			graphHandler,
			notifier,
			log,
		)
	})
	wg.Go(func() error {
		return runInternalHTTPServer(
			ctx,
			cfg.InternalListenAddress,
			promReg,
			log,
		)
	})

	wg.Go(func() error {
		if err := grpc.Run(ctx, cfg.GRPCListenAddress, pool, log); err != nil {
			log.WithError(err).Errorf("error in GRPC server")
			return err
		}
		return nil
	})

	wg.Go(func() error {
		return runUsersync(ctx, pool, cfg, log)
	})

	<-ctx.Done()
	signalStop()
	log.Infof("shutting down...")

	ch := make(chan error)
	go func() {
		ch <- wg.Wait()
	}()

	select {
	case <-time.After(10 * time.Second):
		log.Warn("timed out waiting for graceful shutdown")
	case err := <-ch:
		return err
	}

	return nil
}

// loadEnvFile will load a .env file if it exists. This is useful for local development.
func loadEnvFile(log logrus.FieldLogger) error {
	if _, err := os.Stat(".env"); errors.Is(err, os.ErrNotExist) {
		log.Infof("no .env file found")
		return nil
	}

	if err := godotenv.Load(".env"); err != nil {
		return err
	}

	log.Infof("loaded .env file")
	return nil
}

func setupAuthHandler(ctx context.Context, cfg oAuthConfig, log logrus.FieldLogger) (authn.Handler, error) {
	cf, err := authn.NewOIDC(ctx, cfg.Issuer, cfg.ClientID, cfg.RedirectURL, cfg.AdditionalScopes)
	if err != nil {
		return nil, err
	}
	return authn.New(cf, log), nil
}
