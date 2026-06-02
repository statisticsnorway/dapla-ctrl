//go:build integration_test

package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/auth/middleware"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/serviceaccount"
)

func TestApiKeyAuthentication(t *testing.T) {
	ctx := context.Background()
	log, _ := test.NewNullLogger()

	container, dsn, err := startPostgresql(ctx, t, log)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	setup := func(t *testing.T) (context.Context, *pgxpool.Pool) {
		pool := getConnection(ctx, t, container, dsn, log)
		ctx = database.NewLoaderContext(ctx, pool)
		ctx = serviceaccount.NewLoaderContext(ctx, pool)
		ctx = authz.NewLoaderContext(ctx, pool)
		return ctx, pool
	}

	t.Run("No authorization header", func(t *testing.T) {
		ctx, _ := setup(t)

		apiKeyAuth := middleware.ApiKeyAuthentication()
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if actor := authz.ActorFromContext(r.Context()); actor != nil {
				t.Fatal("expected nil actor")
			}
		})
		req := getRequest(ctx)
		apiKeyAuth(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Unknown API key in header", func(t *testing.T) {
		ctx, _ := setup(t)

		apiKeyAuth := middleware.ApiKeyAuthentication()
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if actor := authz.ActorFromContext(r.Context()); actor != nil {
				t.Fatal("expected nil actor")
			}
		})
		req := getRequest(ctx)
		req.Header.Set("Authorization", "Bearer unknown")
		apiKeyAuth(next).ServeHTTP(responseWriter, req)
	})
	t.Run("Valid API key", func(t *testing.T) {
		ctx, pool := setup(t)

		stmt := `
			INSERT INTO service_accounts (name, description) VALUES
			('sa1', 'description'),
			('sa2', 'description')`
		if _, err = pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("failed to insert service accounts: %v", err)
		}

		token1, _ := serviceaccount.HashToken("key1")
		token2, _ := serviceaccount.HashToken("key2")
		stmt = `
			INSERT INTO service_account_tokens (token, service_account_id, name, description) VALUES
		   ($1, (SELECT id FROM service_accounts WHERE name = 'sa1'), 'name', 'description'),
		   ($2, (SELECT id FROM service_accounts WHERE name = 'sa2'), 'name', 'description')`
		if _, err = pool.Exec(ctx, stmt, token1, token2); err != nil {
			t.Fatalf("failed to insert service accounts: %v", err)
		}

		responseWriter := httptest.NewRecorder()
		next1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if actor := authz.ActorFromContext(r.Context()); actor == nil {
				t.Fatal("expected actor")
			} else if !actor.User.IsServiceAccount() {
				t.Fatal("expected service account")
			} else if expected := "sa1"; actor.User.Identity() != expected {
				t.Fatalf("expected %q, got %q", expected, actor.User.Identity())
			} else if len(actor.Roles) != 0 {
				t.Fatal("expected no role")
			}
		})
		next2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if actor := authz.ActorFromContext(r.Context()); actor == nil {
				t.Fatal("expected actor")
			} else if !actor.User.IsServiceAccount() {
				t.Fatal("expected service account")
			} else if expected := "sa2"; actor.User.Identity() != expected {
				t.Fatalf("expected %q, got %q", expected, actor.User.Identity())
			} else if len(actor.Roles) != 0 {
				t.Fatal("expected no role")
			}
		})

		req := getRequest(ctx)
		req.Header.Set("Authorization", "Bearer key1")
		middleware.ApiKeyAuthentication()(next1).ServeHTTP(responseWriter, req)

		req = getRequest(ctx)
		req.Header.Set("Authorization", "Bearer key2")
		middleware.ApiKeyAuthentication()(next2).ServeHTTP(responseWriter, req)
	})

	t.Run("Expired API key", func(t *testing.T) {
		ctx, pool := setup(t)

		stmt := `
			INSERT INTO service_accounts (name, description) VALUES
			('sa1', 'description')`
		if _, err = pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("failed to insert service accounts: %v", err)
		}

		token1, _ := serviceaccount.HashToken("key1")
		stmt = `
			INSERT INTO service_account_tokens (token, service_account_id, expires_at, name, description) VALUES
		   ($1, (SELECT id FROM service_accounts WHERE name = 'sa1'), '2021-01-01', 'token', 'description')`
		if _, err = pool.Exec(ctx, stmt, token1); err != nil {
			t.Fatalf("failed to insert service accounts: %v", err)
		}

		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if actor := authz.ActorFromContext(r.Context()); actor != nil {
				t.Fatal("expected nil actor")
			}
		})

		req := getRequest(ctx)
		req.Header.Set("Authorization", "Bearer key1")
		middleware.ApiKeyAuthentication()(next).ServeHTTP(responseWriter, req)
	})
}
