package config_test

import (
	"context"
	"testing"

	"github.com/statisticsnorway/dapla-api-reconcilers/internal/cmd/reconciler/config"

	"github.com/sethvargo/go-envconfig"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	lookuper := envconfig.MapLookuper(map[string]string{
		"GCP_CLUSTERS": `{"name":{"project_id":"some-id","teams_folder_id":"123456789"}}`,
	})
	_, err := config.NewConfig(ctx, lookuper)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

}
