package api

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/config"
	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/googleresourcemanager"
	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/googlesqladmin"
	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/parquedit"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/cloudresourcemanager/v1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

func Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx, signalStop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer signalStop()

	cfg, err := config.ParseConfig[config.Config]()
	if err != nil {
		return fmt.Errorf("could not parse config, %w", err)
	}

	var resourceManager parquedit.CloudResourceManager
	if cfg.Fakes.WithFakeCloudResourceManager {
		slog.Warn("running with fake cloud resource manager")
		resourceManager = googleresourcemanager.NewFake()
	} else {
		gcrm, err := cloudresourcemanager.NewService(ctx)
		if err != nil {
			return fmt.Errorf("unable to create google cloudresourcemanager client: %w", err)
		}
		resourceManager = googleresourcemanager.New(gcrm)
	}

	var sqlAdminClient parquedit.SqlManager
	if cfg.Fakes.WithFakeSqlAdmin {
		slog.Warn("running with fake sql admin")
		sqlAdminClient = googlesqladmin.NewFake(cfg.Fakes.FakeSqlAdminDatabaseConnString)
	} else {
		sqladminService, err := sqladmin.NewService(ctx)
		if err != nil {
			return fmt.Errorf("unable to create google sqladmin client: %w", err)
		}
		sqlAdminClient = googlesqladmin.New(sqladminService)
	}

	parqueditClient, err := parquedit.New(ctx, cfg.Parquedit, resourceManager, sqlAdminClient)
	if err != nil {
		return fmt.Errorf("could not configure Parquedit: %w", err)
	}
	defer parqueditClient.Close()

	router := SetupRoutes(cfg.Router, parqueditClient)

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return runHTTPServer(
			ctx,
			cfg.Server,
			router,
		)
	})

	<-ctx.Done()
	signalStop()
	slog.Info("shutting down...")

	ch := make(chan error)
	go func() {
		ch <- wg.Wait()
	}()

	select {
	case <-time.After(10 * time.Second):
		slog.Warn("timed out waiting for graceful shutdown")
	case err := <-ch:
		return err
	}

	return nil
}
