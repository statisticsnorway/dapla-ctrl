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
	for _, env := range []string{"prod", "test"} {
		t.Run(env, func(t *testing.T) {
			template := "gs://ssb-{team}-data-produkt-" + env + "/.parquedit_data/"
			client, err := New(ctx, config.ParqueditConfig{
				DatabaseUrl:      connectionString,
				DataPathTemplate: template,
			}, grm, googlesqladmin.NewFake(connectionString))
			if err != nil {
				t.Fatalf("failed to create parquedit client: %v", err)
			}
			t.Cleanup(client.Close)

			handler := setupTestRoutes(client)
			team := "binde-strek" // Cloud SQL permits dashes, unlike the local test database.
			schema := "team_binde_strek"
			wantDataPath := "gs://ssb-" + team + "-data-produkt-" + env + "/.parquedit_data/"

			t.Run("check status for non enabled team", func(t *testing.T) {
				assertStatus(t, handler, http.MethodGet, team, http.StatusNotFound)
				assertSchemaExists(ctx, t, client, schema, false)
			})

			t.Run("enable for team", func(t *testing.T) {
				assertStatus(t, handler, http.MethodPut, team, http.StatusOK)
				assertSchemaExists(ctx, t, client, schema, true)
				assertDataPath(ctx, t, client, schema, wantDataPath)
				assertStatus(t, handler, http.MethodGet, team, http.StatusOK)
			})

			t.Run("enable is idempotent", func(t *testing.T) {
				assertStatus(t, handler, http.MethodPut, team, http.StatusOK)
				assertSchemaExists(ctx, t, client, schema, true)
				assertDataPath(ctx, t, client, schema, wantDataPath)
			})

			t.Run("disable for team, and it is idempotent", func(t *testing.T) {
				assertStatus(t, handler, http.MethodDelete, team, http.StatusOK)
				assertSchemaExists(ctx, t, client, schema, false)
				assertStatus(t, handler, http.MethodGet, team, http.StatusNotFound)

				assertStatus(t, handler, http.MethodDelete, team, http.StatusOK)
				assertSchemaExists(ctx, t, client, schema, false)
			})
		})
	}
}

func assertDataPath(ctx context.Context, t *testing.T, client *Client, schema, want string) {
	t.Helper()
	var got string
	if err := client.db.QueryRow(ctx, "SELECT value FROM "+schema+".ducklake_metadata WHERE key = 'data_path'").Scan(&got); err != nil {
		t.Fatalf("failed to query data_path: %v", err)
	}
	if got != want {
		t.Fatalf("data_path = %q, want %q", got, want)
	}
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
