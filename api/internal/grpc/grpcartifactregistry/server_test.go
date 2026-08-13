//go:build integration_test

package grpcartifactregistry_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/database"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/grpc/grpcartifactregistry"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestArtifactRegistryServer_GetArtifactRegistryRepo(t *testing.T) {
	ctx := context.Background()
	log, _ := logrustest.NewNullLogger()

	container, dsn, err := startPostgresql(ctx, t, log)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Run("repo not found", func(t *testing.T) {
		pool := getConnection(ctx, t, container, dsn, log)

		resp, err := grpcartifactregistry.NewServer(pool).GetArtifactRegistryRepo(ctx, &protoapi.GetArtifactRegistryRepoRequest{
			TeamSlug: "team-not-found",
			Format:   "docker",
		})
		if resp != nil {
			t.Error("expected response to be nil")
		}
		if s, ok := status.FromError(err); !ok || s.Code() != codes.Internal {
			t.Errorf("expected status code %v, got %v", codes.Internal, err)
		}
	})

	t.Run("get repo", func(t *testing.T) {
		pool := getConnection(ctx, t, container, dsn, log)
		insertTeam(t, ctx, pool, "team")
		insertArtifactRegistryRepo(t, ctx, pool, "team", "docker", 123)
		insertGithubRepository(t, ctx, pool, "team", "awesome-repo")

		resp, err := grpcartifactregistry.NewServer(pool).GetArtifactRegistryRepo(ctx, &protoapi.GetArtifactRegistryRepoRequest{
			TeamSlug: "team",
			Format:   "docker",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Repo == nil {
			t.Fatal("expected repo to be non-nil")
		}
		if resp.Repo.TeamSlug != "team" {
			t.Errorf("expected team slug %q, got %q", "team", resp.Repo.TeamSlug)
		}
		if resp.Repo.Format != "docker" {
			t.Errorf("expected format %q, got %q", "docker", resp.Repo.Format)
		}
		if resp.Repo.SizeBytes != 123 {
			t.Errorf("expected size bytes %d, got %d", 123, resp.Repo.SizeBytes)
		}
	})
}

func TestArtifactRegistryServer_SetArtifactRegistryRepoSizeBytes(t *testing.T) {
	ctx := context.Background()
	log, _ := logrustest.NewNullLogger()

	container, dsn, err := startPostgresql(ctx, t, log)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	pool := getConnection(ctx, t, container, dsn, log)
	insertTeam(t, ctx, pool, "team")
	insertArtifactRegistryRepo(t, ctx, pool, "team", "docker", 123)

	resp, err := grpcartifactregistry.NewServer(pool).SetArtifactRegistryRepoSizeBytes(ctx, &protoapi.SetArtifactRegistryRepoSizeBytesRequest{
		TeamSlug:  "team",
		Format:    "docker",
		SizeBytes: 123,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response to be non-nil")
	}

	var sizeBytes int64
	if err := pool.QueryRow(ctx, "SELECT size_bytes FROM team_artifact_registry_repositories WHERE team_slug = $1 AND format = $2", "team", "docker").Scan(&sizeBytes); err != nil {
		t.Fatalf("failed to get updated repo: %v", err)
	}
	if sizeBytes != 123 {
		t.Errorf("expected size bytes %d, got %d", 123, sizeBytes)
	}
}

func TestArtifactRegistryServer_ListArtifactRegistryReposForTeam(t *testing.T) {
	ctx := context.Background()
	log, _ := logrustest.NewNullLogger()

	container, dsn, err := startPostgresql(ctx, t, log)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Run("no repos", func(t *testing.T) {
		pool := getConnection(ctx, t, container, dsn, log)

		resp, err := grpcartifactregistry.NewServer(pool).ListArtifactRegistryReposForTeam(ctx, &protoapi.ListArtifactRegistryReposForTeamRequest{
			TeamSlug: "team",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Nodes) != 0 {
			t.Errorf("expected 0 repos, got %d", len(resp.Nodes))
		}
		if resp.PageInfo.TotalCount != 0 {
			t.Errorf("expected total count 0, got %d", resp.PageInfo.TotalCount)
		}
	})

	t.Run("repos for team", func(t *testing.T) {
		pool := getConnection(ctx, t, container, dsn, log)
		teamSlug := "team"
		insertTeam(t, ctx, pool, teamSlug)
		insertTeam(t, ctx, pool, "other-team")
		insertGithubRepository(t, ctx, pool, teamSlug, "awesome-repo")
		insertGithubRepository(t, ctx, pool, "other-team", "awesome-repo")

		insertArtifactRegistryRepo(t, ctx, pool, teamSlug, "npm", 123)
		insertArtifactRegistryRepo(t, ctx, pool, teamSlug, "docker", 123)
		insertArtifactRegistryRepo(t, ctx, pool, teamSlug, "maven", 123)
		insertArtifactRegistryRepo(t, ctx, pool, "other-team", "docker", 123)

		resp, err := grpcartifactregistry.NewServer(pool).ListArtifactRegistryReposForTeam(ctx, &protoapi.ListArtifactRegistryReposForTeamRequest{
			TeamSlug: teamSlug,
			Limit:    3,
			Offset:   0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, format := range [3]string{"npm", "docker", "maven"} {
			if !slices.ContainsFunc(resp.Nodes, func(e *protoapi.ArtifactRegistryRepo) bool {
				return e.Format == format
			}) {
				t.Errorf("expected result to contain repo of format %q, but it did not, whole result: %q", format, resp.Nodes)
			}
		}

		for _, repo := range resp.Nodes {
			if repo.TeamSlug != teamSlug {
				t.Errorf("expected repo to belong to team %q, got %q", teamSlug, repo.TeamSlug)
			}
		}
	})
}

func TestArtifactRegistryServer_GetArtifactRegistryGithubAllowlist(t *testing.T) {
	ctx := context.Background()
	log, _ := logrustest.NewNullLogger()

	container, dsn, err := startPostgresql(ctx, t, log)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	pool := getConnection(ctx, t, container, dsn, log)
	insertTeam(t, ctx, pool, "team")
	insertGithubRepository(t, ctx, pool, "team", "awesome-repo")
	insertGithubRepository(t, ctx, pool, "team", "awesome-repo2")
	insertGithubRepository(t, ctx, pool, "team", "awesome-repo3")

	resp, err := grpcartifactregistry.NewServer(pool).GetArtifactRegistryGithubAllowlist(ctx, &protoapi.GetArtifactRegistryGithubAllowlistRequest{
		TeamSlug: "team",
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Nodes) != 3 {
		t.Fatalf("expected 3 github repository, got %d", len(resp.Nodes))
	}
	if resp.Nodes[0] != "awesome-repo" {
		t.Errorf("expected github repository %q, got %q", "awesome-repo", resp.Nodes[0])
	}
	if resp.PageInfo == nil {
		t.Error("expected page info to be non-nil")
	}
}

func insertTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, slug string) {
	t.Helper()

	sectionCode := "123"
	if _, err := pool.Exec(ctx, "INSERT INTO sections (code, name) VALUES ($1, 'Section') ON CONFLICT (code) DO NOTHING", sectionCode); err != nil {
		t.Fatalf("failed to insert team: %v", err)
	}
	if _, err := pool.Exec(ctx, "INSERT INTO teams (slug, display_name, section_code, is_managed) VALUES ($1, 'Team', $2, TRUE)", slug, sectionCode); err != nil {
		t.Fatalf("failed to insert team: %v", err)
	}
}

func insertArtifactRegistryRepo(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamSlug, format string, sizeBytes int64) {
	t.Helper()

	if _, err := pool.Exec(ctx, "INSERT INTO team_artifact_registry_repositories (team_slug, format, size_bytes) VALUES ($1, $2, $3)", teamSlug, format, sizeBytes); err != nil {
		t.Fatalf("failed to insert artifact registry repo: %v", err)
	}
}

func insertGithubRepository(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamSlug, githubRepository string) {
	t.Helper()

	if _, err := pool.Exec(ctx, "INSERT INTO team_artifact_registry_github_repositories (team_slug, github_repository) VALUES ($1, $2)", teamSlug, githubRepository); err != nil {
		t.Fatalf("failed to insert github repository: %v", err)
	}
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
