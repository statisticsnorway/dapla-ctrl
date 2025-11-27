package main

import (
	"context"

	"github.com/statisticsnorway/dapla-api-reconcilers/internal/cmd/reconciler"
)

func main() {
	reconciler.Run(context.Background())
}
