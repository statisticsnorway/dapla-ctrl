package main

import (
	"context"

	// Auto configure settings when running in a container
	_ "go.uber.org/automaxprocs"

	"github.com/statisticsnorway/dapla-api/internal/cmd/api"
)

func main() {
	ctx := context.Background()
	api.Run(ctx)
}
