package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/config"
	"golang.org/x/sync/errgroup"
)

func runHTTPServer(ctx context.Context, cfg config.ServerConfig, router *chi.Mux) error {
	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		slog.Info(" HTTP server shutting down...")
		if err := server.Shutdown(ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Info("HTTP server shutdown failed", "error", err)
			return err
		}
		return nil
	})

	wg.Go(func() error {
		slog.Info("HTTP server accepting requests on " + cfg.ListenAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Info("unexpected error from HTTP server", "error", err)
			return err
		}
		slog.Info("HTTP server finished, terminating...")
		return nil
	})
	return wg.Wait()
}
