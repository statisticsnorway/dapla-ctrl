package main

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/cmd/reconciler"
)

func main() {
	reconciler.Run(context.Background())
}
