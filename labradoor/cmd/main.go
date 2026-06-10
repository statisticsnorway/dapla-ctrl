package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/api"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)
	slog.SetLogLoggerLevel(slog.LevelWarn)

	if err := loadEnvFile(); err != nil {
		slog.Error("error loading .env file", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	err := api.Run(ctx)
	if err != nil {
		slog.Error("error in Run()", "error", err)
		os.Exit(1)
	}
}

// loadEnvFile will load a .env file if it exists. This is useful for local development.
func loadEnvFile() error {
	if _, err := os.Stat(".env"); errors.Is(err, os.ErrNotExist) {
		slog.Info("no .env file found")
		return nil
	}

	if err := godotenv.Load(".env"); err != nil {
		return err
	}

	slog.Info("loaded .env file")
	return nil
}
