package main

import (
	"context"

	"github.com/statisticsnorway/dapla-api/internal/cmd/api"
)

func main() {
	ctx := context.Background()
	api.Run(ctx)
}
