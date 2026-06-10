package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/api"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)
	slog.SetLogLoggerLevel(slog.LevelWarn)

	ctx := context.Background()

	err := api.Run(ctx)
	if err != nil {
		slog.Error("error in Run()", "error", err)
		os.Exit(1)
	}
}
