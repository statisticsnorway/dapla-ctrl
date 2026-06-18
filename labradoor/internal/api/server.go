package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/errgroup"
)

type ServerConfig struct {
	ListenAddr string `env:"LISTEN_ADDR" envDefault:":8080"`
}

func runHTTPServer(ctx context.Context, router *chi.Mux) error {
	serverConfig := parseConfigOrDie[ServerConfig]()
	server := &http.Server{
		Addr:              serverConfig.ListenAddr,
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
		slog.Info("HTTP server accepting requests on " + serverConfig.ListenAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Info("unexpected error from HTTP server", "error", err)
			return err
		}
		slog.Info("HTTP server finished, terminating...")
		return nil
	})
	return wg.Wait()
}

func parseConfigOrDie[T any]() T {
	result, err := env.ParseAs[T]()
	if err != nil {
		slog.Error("could not parse config", "error", err)
		os.Exit(1)
	}
	return result
}
