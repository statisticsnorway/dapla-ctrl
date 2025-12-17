package reconciler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers/entraid/gcpsyncer"
	entraidreconciler "github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers/entraid/group"

	"github.com/joho/godotenv"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/cmd/reconciler/config"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/logger"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers"

	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	if fileLoaded, err := loadEnvFile(); err != nil {
		log.WithError(err).Errorf("error when loading .env file")
		os.Exit(exitCodeEnvFileError)
	} else if fileLoaded {
		log.Infof("loaded .env file")
	}

	cfg, err := config.NewConfig(ctx, envconfig.OsLookuper())
	if err != nil {
		log.WithError(err).Errorf("error when processing configuration")
		os.Exit(exitCodeConfigError)
	}

	appLogger, err := logger.New(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		log.WithError(err).Errorf("error when creating application logger")
		os.Exit(exitCodeLoggerError)
	}

	err = run(ctx, cfg, appLogger)
	if err != nil {
		appLogger.WithError(err).Errorf("error in run()")
		os.Exit(exitCodeRunError)
	}

	os.Exit(exitCodeSuccess)
}

func run(ctx context.Context, cfg *config.Config, log logrus.FieldLogger) error {
	ctx, signalStop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer signalStop()

	start := time.Now()

	_, promRegistry, err := newMeterProvider()
	if err != nil {
		return fmt.Errorf("error when creating meter provider: %w", err)
	}
	log.WithField("duration", time.Since(start).String()).Debug("Created meter provider")

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		defer log.Debug("Done running main http server goroutine")
		return runHttpServer(ctx, cfg.ListenAddress, promRegistry, log)
	})

	opts := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
	if cfg.GRPC.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	client, err := apiclient.New(cfg.GRPC.Target, opts...)
	if err != nil {
		return fmt.Errorf("error when creating API client: %w", err)
	}
	log.WithField("duration", time.Since(start).String()).Debug("Created API client")

	reconcilerManager := reconcilers.NewManager(ctx, client, cfg.ReconcilersToEnable, cfg.PubSub.SubscriptionID, cfg.PubSub.ProjectID, log)
	log.WithField("duration", time.Since(start).String()).Debug("Created reconciler manager")

	// Init reconcilers
	gcpSyncer := gcpsyncer.New(client)

	entraIdGroupReconciler := entraidreconciler.New(ctx, gcpSyncer)
	log.WithField("duration", time.Since(start).String()).Debug("Created Entra ID group reconciler")

	// The reconcilers will be run in the order they are added to the manager
	reconcilerManager.AddReconciler(entraIdGroupReconciler)
	reconcilerManager.AddReconciler(gcpSyncer)

	log.WithField("duration", time.Since(start).String()).Debug("Added reconcilers to manager")

	if err := reconcilerManager.RegisterReconcilersWithAPI(ctx); err != nil {
		return fmt.Errorf("error when registering reconcilers with API: %w", err)
	}
	log.WithField("duration", time.Since(start).String()).Debug("Registered reconcilers with API")

	wg.Go(func() error {
		defer log.Debug("Done running gcpsyncer")
		gcpSyncer.Run(ctx)
		return nil
	})

	for i := range 10 {
		wg.Go(func() error {
			defer log.Debugf("Done running reconciler %v", i)
			reconcilerManager.Run(ctx)
			return nil
		})
	}

	wg.Go(func() error {
		defer log.Debug("Done listening for pubsub events")
		reconcilerManager.ListenForEvents(ctx)
		return nil
	})

	if err = reconcilerManager.SyncAllTeams(ctx, time.Minute*30); err != nil {
		return fmt.Errorf("error when syncing all teams: %w", err)
	}
	log.WithField("duration", time.Since(start).String()).Debug("Synced all teams")

	reconcilerManager.Close()
	return wg.Wait()
}

func loadEnvFile() (fileLoaded bool, err error) {
	if _, err = os.Stat(".env"); errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err = godotenv.Load(".env"); err != nil {
		return false, err
	}

	return true, nil
}
