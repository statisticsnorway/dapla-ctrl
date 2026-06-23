//go:build integration_test

package parquedit

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/config"
	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/googleresourcemanager"
	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/googlesqladmin"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestManageTeamSchema(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	grm := googleresourcemanager.NewFake()
	connectionString := startPostgres(ctx, t)
	client, err := New(ctx, config.ParqueditConfig{DatabaseUrl: connectionString}, grm, googlesqladmin.NewFake(connectionString))
	if err != nil {
		t.Fatalf("failed to create parquedit client: %v", err)
	}
	t.Cleanup(client.Close)

	handler := setupTestRoutes(client)

	team := "binde-strek" // Test with a underdash team, since we otherwise have to do special logic related to test. In cloudsql dash is allowed
	schema := "team_binde_strek"

	t.Run("check status for non enabled team", func(t *testing.T) {
		assertStatus(t, handler, http.MethodGet, team, http.StatusNotFound)
		assertSchemaExists(ctx, t, client, schema, false)
	})

	t.Run("enable for team", func(t *testing.T) {
		assertStatus(t, handler, http.MethodPut, team, http.StatusOK)
		assertSchemaExists(ctx, t, client, schema, true)
		assertStatus(t, handler, http.MethodGet, team, http.StatusOK)
	})

	t.Run("enable is idempotent", func(t *testing.T) {
		assertStatus(t, handler, http.MethodPut, team, http.StatusOK)
		assertSchemaExists(ctx, t, client, schema, true)
	})

	t.Run("disable for team, and it is idempotent", func(t *testing.T) {
		assertStatus(t, handler, http.MethodDelete, team, http.StatusOK)
		assertSchemaExists(ctx, t, client, schema, false)
		assertStatus(t, handler, http.MethodGet, team, http.StatusNotFound)

		assertStatus(t, handler, http.MethodDelete, team, http.StatusOK)
		assertSchemaExists(ctx, t, client, schema, false)
	})
}

func startPostgres(ctx context.Context, t *testing.T) string {
	t.Helper()

	container, err := postgres.Run(
		ctx,
		"docker.io/postgres:18-alpine",
		postgres.WithDatabase("labradoor"),
		postgres.WithUsername("labradoor"),
		postgres.WithPassword("labradoor"),
		postgres.WithSQLDriver("pgx"),
		postgres.BasicWaitStrategies(),
	)
	defer testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connectionString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	return connectionString
}

func setupTestRoutes(client *Client) http.Handler {
	router := chi.NewRouter()
	router.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{}))
	router.Route("/parquedit/{team}", func(r chi.Router) {
		r.Get("/", client.HasEnabled)
		r.Put("/", client.EnableForTeam)
		r.Delete("/", client.DisableForTeam)
	})
	return router
}

func assertStatus(t *testing.T, handler http.Handler, method, team string, want int) {
	t.Helper()

	req := httptest.NewRequest(method, "/parquedit/"+team, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != want {
		t.Fatalf("%s %s status = %d, want %d", method, req.URL.Path, rec.Code, want)
	}
}

func assertSchemaExists(ctx context.Context, t *testing.T, client *Client, schema string, want bool) {
	t.Helper()

	var exists bool
	err := client.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.schemata
			WHERE schema_name = $1
		)`, schema).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check schema %q: %v", schema, err)
	}
	if exists != want {
		t.Fatalf("schema %q exists = %v, want %v", schema, exists, want)
	}
}
