//go:build integration_test

package grpcteam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/grpc/grpcteam"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/protoapi"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTeamsServer_Get(t *testing.T) {
	ctx := context.Background()
	log, _ := logrustest.NewNullLogger()

	container, dsn, err := startPostgresql(ctx, t, log)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Run("team not found", func(t *testing.T) {
		pool := getConnection(ctx, t, container, dsn, log)
		resp, err := grpcteam.NewServer(pool).Get(ctx, &protoapi.GetTeamRequest{Slug: "team-not-found"})
		if resp != nil {
			t.Error("expected response to be nil")
		}

		if s, ok := status.FromError(err); !ok || s.Code() != codes.NotFound {
			t.Errorf("expected status code %v, got %v", codes.NotFound, err)
		}
	})

	t.Run("get team", func(t *testing.T) {
		pool := getConnection(ctx, t, container, dsn, log)

		teamSlug := "team"
		purpose := "purpose"
		sectionCode := "724"

		stmt := `
			INSERT INTO teams (slug, display_name, purpose, section_code, is_managed) VALUES
			($1, 'Team', $2, $3, TRUE)`
		if _, err = pool.Exec(ctx, stmt, teamSlug, purpose, sectionCode); err != nil {
			t.Fatalf("failed to insert team: %v", err)
		}

		resp, err := grpcteam.NewServer(pool).Get(ctx, &protoapi.GetTeamRequest{Slug: teamSlug})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Team == nil {
			t.Error("expected response to be non-nil")
		}

		if resp.Team.Slug != teamSlug {
			t.Errorf("expected team slug to be %q, got %q", teamSlug, resp.Team.Slug)
		}

		if resp.Team.Purpose != purpose {
			t.Errorf("expected team purpose to be %q, got %q", purpose, resp.Team.Purpose)
		}
	})
}

func TestTeamsServer_Delete(t *testing.T) {
	ctx := context.Background()
	log, _ := logrustest.NewNullLogger()

	container, dsn, err := startPostgresql(ctx, t, log)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Run("missing slug", func(t *testing.T) {
		pool := getConnection(ctx, t, container, dsn, log)

		resp, err := grpcteam.NewServer(pool).Delete(ctx, &protoapi.DeleteTeamRequest{})
		if resp != nil {
			t.Error("expected response to be nil")
		}

		if s, ok := status.FromError(err); !ok || s.Code() != codes.InvalidArgument {
			t.Errorf("expected status code %v, got %v", codes.InvalidArgument, err)
		}
	})

	t.Run("delete team", func(t *testing.T) {
		pool := getConnection(ctx, t, container, dsn, log)

		teamSlug := "team-slug"
		sectionCode := "724"

		stmt := "INSERT INTO teams (slug, display_name, purpose, delete_key_confirmed_at, section_code, is_managed) VALUES ($1, 'Team', 'some purpose', NOW(), $2, TRUE)"
		if _, err := pool.Exec(ctx, stmt, teamSlug, sectionCode); err != nil {
			t.Fatalf("failed to insert team: %v", err)
		}

		resp, err := grpcteam.NewServer(pool).Delete(ctx, &protoapi.DeleteTeamRequest{Slug: teamSlug})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp == nil {
			t.Error("expected response to be non-nil")
		}

		count := 0
		stmt = "SELECT COUNT(*) FROM teams WHERE slug = $1"
		if err := pool.QueryRow(ctx, stmt, teamSlug).Scan(&count); err != nil {
			t.Fatalf("failed to count teams: %v", err)
		} else if count != 0 {
			t.Fatalf("expected team to be deleted")
		}
	})
}

func TestTeamsServer_ToBeReconciled(t *testing.T) {
	ctx := context.Background()
	log, _ := logrustest.NewNullLogger()

	container, dsn, err := startPostgresql(ctx, t, log)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Run("fetch teams", func(t *testing.T) {
		pool := getConnection(ctx, t, container, dsn, log)

		stmt := "INSERT INTO teams (slug, display_name, purpose, section_code, is_managed) VALUES ('teamone', 'Team One', 'some purpose', '724', TRUE)"
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("failed to insert team: %v", err)
		}

		stmt = "INSERT INTO teams (slug, display_name, purpose, section_code, is_managed) VALUES ('teamtwo', 'Team Two', 'some purpose', '724', TRUE)"
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("failed to insert team: %v", err)
		}

		resp, err := grpcteam.NewServer(pool).List(ctx, &protoapi.ListTeamsRequest{
			Limit:  2,
			Offset: 0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Nodes) != 2 {
			t.Errorf("expected 2 teams, got %v", resp.Nodes)
		}

		if expected := "teamone"; resp.Nodes[0].Slug != expected {
			t.Errorf("expected first team to be %q, got %q", expected, resp.Nodes[0].Slug)
		}

		if expected := "teamtwo"; resp.Nodes[1].Slug != expected {
			t.Errorf("expected first team to be %q, got %q", expected, resp.Nodes[1].Slug)
		}
	})
}

func startPostgresql(ctx context.Context, t *testing.T, log logrus.FieldLogger) (container *postgres.PostgresContainer, dsn string, err error) {
	container, err = postgres.Run(
		ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithSQLDriver("pgx"),
		postgres.BasicWaitStrategies(),
	)
	defer testcontainers.CleanupContainer(t, container)

	if err != nil {
		return nil, "", fmt.Errorf("failed to start container: %w", err)
	}

	dsn, err = container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get connection string: %w", err)
	}

	pool, err := database.NewPool(ctx, dsn, log, true)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create pool: %w", err)
	}
	pool.Close()

	if err := container.Snapshot(ctx); err != nil {
		return nil, "", fmt.Errorf("failed to snapshot: %w", err)
	}

	return container, dsn, nil
}

func getConnection(ctx context.Context, t *testing.T, container *postgres.PostgresContainer, dsn string, log logrus.FieldLogger) *pgxpool.Pool {
	pool, _ := database.NewPool(ctx, dsn, log, false)

	t.Cleanup(func() {
		pool.Close()
		if err := container.Restore(ctx); err != nil {
			t.Fatalf("failed to restore database: %v", err)
		}
	})

	return pool
}
