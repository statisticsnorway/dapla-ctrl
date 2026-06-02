//go:build integration_test

package integrationtests

import (
	"context"
	"testing"

	"github.com/nais/tester/lua"
	"github.com/statisticsnorway/dapla-api/internal/integration"
)

func TestIntegration(t *testing.T) {
	ctx := context.Background()
	mgr, cleanup, err := integration.TestRunner(ctx, false)
	if err != nil {
		t.Fatal(err)
	}

	defer cleanup()

	if err := mgr.Run(ctx, ".", lua.NewTestReporter(t)); err != nil {
		t.Fatal(err)
	}
}
