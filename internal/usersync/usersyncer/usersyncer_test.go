package usersyncer_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/microsoft/kiota-abstractions-go/authentication"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/test"
	"github.com/statisticsnorway/dapla-api/internal/user"
	"github.com/statisticsnorway/dapla-api/internal/usersync/usersyncer"
	"github.com/statisticsnorway/dapla-api/internal/usersync/usersyncsql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	domain        = "example.com"
	allUsersGroup = "all-users"
	adminGroup    = "api-admins"
)

func TestSync(t *testing.T) {
	ctx := context.Background()
	log, _ := logrustest.NewNullLogger()

	container, dsn, err := startPostgresql(ctx, t, log)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	setup := func(t *testing.T) (context.Context, *pgxpool.Pool) {
		pool := getConnection(ctx, t, container, dsn, log)
		ctx = database.NewLoaderContext(ctx, pool)
		ctx = user.NewLoaderContext(ctx, pool)
		return ctx, pool
	}
	t.Run("No local users, no remote users", func(t *testing.T) {
		ctx, pool := setup(t)

		httpClient := test.NewTestHttpClient(
			func(req *http.Request) *http.Response {
				return test.Response("200 OK", `{"value":[]}`)
			},
			func(req *http.Request) *http.Response {
				return test.Response("200 OK", `{"value":[]}`)
			},
		)

		auth := authentication.NewBaseBearerTokenAuthenticationProvider(&fakeCred{})
		adapter, err := msgraphsdk.NewGraphRequestAdapterWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(
			auth, nil, nil, httpClient,
		)
		if err != nil {
			t.Fatal(err)
		}

		client := msgraphsdk.NewGraphServiceClient(adapter)

		err = usersyncer.
			New(pool, allUsersGroup, adminGroup, domain, client, log).
			Sync(ctx)
		if err != nil {
			t.Fatalf("failed to sync: %v", err)
		}

		p, _ := pagination.ParsePage(nil, nil, nil, nil)
		if users, err := user.List(ctx, p, nil); err != nil {
			t.Fatal(err)
		} else if total := len(users.Nodes()); total != 0 {
			t.Fatalf("expected 0 users, got %d", total)
		}
	})

	t.Run("Local users, no remote users", func(t *testing.T) {
		ctx, pool := setup(t)
		querier := usersyncsql.New(pool)

		user1, err := querier.Create(ctx, usersyncsql.CreateParams{
			Name:       "User 1",
			Email:      "user1@example.com",
			ExternalID: "123",
		})
		if err != nil {
			t.Fatal(err)
		}

		user2, err := querier.Create(ctx, usersyncsql.CreateParams{
			Name:       "User 2",
			Email:      "user2@example.com",
			ExternalID: "456",
		})
		if err != nil {
			t.Fatal(err)
		}

		if err := querier.AssignGlobalRole(ctx, usersyncsql.AssignGlobalRoleParams{
			UserID:   user1.ID,
			RoleName: "Team creator",
		}); err != nil {
			t.Fatal(err)
		}

		if err := querier.AssignGlobalAdmin(ctx, user2.ID); err != nil {
			t.Fatal(err)
		}

		httpClient := test.NewTestHttpClient(
			func(req *http.Request) *http.Response {
				return test.Response("200 OK", `{"value":[]}`)
			},
			func(req *http.Request) *http.Response {
				return test.Response("200 OK", `{"value":[]}`)
			},
		)

		auth := authentication.NewBaseBearerTokenAuthenticationProvider(&fakeCred{})
		adapter, err := msgraphsdk.NewGraphRequestAdapterWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(
			auth, nil, nil, httpClient,
		)
		if err != nil {
			t.Fatal(err)
		}

		client := msgraphsdk.NewGraphServiceClient(adapter)

		err = usersyncer.
			New(pool, allUsersGroup, adminGroup, domain, client, log).
			Sync(ctx)
		if err != nil {
			t.Fatal(err)
		}

		p, _ := pagination.ParsePage(nil, nil, nil, nil)
		if users, err := user.List(ctx, p, nil); err != nil {
			t.Fatal(err)
		} else if total := len(users.Nodes()); total != 0 {
			t.Fatalf("expected 0 users, got %d", total)
		}
	})

	t.Run("Create, update and delete users", func(t *testing.T) {
		ctx, pool := setup(t)
		querier := usersyncsql.New(pool)

		userWithIncorrectName, err := querier.Create(ctx, usersyncsql.CreateParams{
			Name:       "Incorrect Name",
			Email:      "user1@example.com",
			ExternalID: "1",
		})
		if err != nil {
			t.Fatal(err)
		}

		userWithIncorrectEmail, err := querier.Create(ctx, usersyncsql.CreateParams{
			Name:       "Some Name",
			Email:      "incorrect@example.com",
			ExternalID: "2",
		})
		if err != nil {
			t.Fatal(err)
		}

		userThatWillBeDeleted, err := querier.Create(ctx, usersyncsql.CreateParams{
			Name:       "Delete Me",
			Email:      "delete-me@example.com",
			ExternalID: "3",
		})
		if err != nil {
			t.Fatal(err)
		}

		userThatShouldLoseAdminRole, err := querier.Create(ctx, usersyncsql.CreateParams{
			Name:       "Should Lose Admin",
			Email:      "should-lose-admin@example.com",
			ExternalID: "4",
		})
		if err != nil {
			t.Fatal(err)
		}

		if err := querier.AssignGlobalAdmin(ctx, userThatShouldLoseAdminRole.ID); err != nil {
			t.Fatal(err)
		}

		httpClient := test.NewTestHttpClient(
			func(req *http.Request) *http.Response {
				return test.Response("200 OK", `{"value":[`+
					`{"@odata.type": "#microsoft.graph.user", "id": "1", "userPrincipalName":"user1@example.com","displayName":"Correct Name"},`+ // Will update name of local user
					`{"@odata.type": "#microsoft.graph.user", "id": "2", "userPrincipalName":"user2@example.com","displayName":"Some Name"},`+ // Will update euserPrincipalName of local user
					`{"@odata.type": "#microsoft.graph.user", "id": "4", "userPrincipalName":"should-lose-admin@example.com","displayName":"Should Lose Admin"},`+ // Will lose admin role
					`{"@odata.type": "#microsoft.graph.user", "id": "5", "userPrincipalName":"create-me@example.com","displayName":"Create Me"}]}`) // Will be created
			},
			func(req *http.Request) *http.Response {
				return test.Response("200 OK", `{"value":[`+
					`{"@odata.type": "#microsoft.graph.user", "id": "2", "userPrincipalName":"user2@example.com", "displayName":"Some Name"},`+ // Will be granted admin role
					`{"@odata.type": "#microsoft.graph.user", "id": "7", "userPrincipalName":"unknown-admin@example.com", "displayName": "Unknown Admin"}]}`) // Unknown user, will be logged
			},
		)

		auth := authentication.NewBaseBearerTokenAuthenticationProvider(&fakeCred{})
		adapter, err := msgraphsdk.NewGraphRequestAdapterWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(
			auth, nil, nil, httpClient,
		)
		if err != nil {
			t.Fatal(err)
		}

		client := msgraphsdk.NewGraphServiceClient(adapter)

		err = usersyncer.
			New(pool, allUsersGroup, adminGroup, domain, client, log).
			Sync(ctx)
		if err != nil {
			t.Fatal(err)
		}

		p, _ := pagination.ParsePage(nil, nil, nil, nil)
		if users, err := user.List(ctx, p, nil); err != nil {
			t.Fatal(err)
		} else if total := len(users.Nodes()); total != 4 {
			t.Fatalf("expected 3 users, got %d", total)
		}

		if u, err := user.Get(ctx, userWithIncorrectName.ID); err != nil {
			t.Fatal(err)
		} else if correctName := "Correct Name"; u.Name != correctName {
			t.Fatalf("expected name to be %q, got %q", correctName, u.Name)
		}

		if u, err := user.Get(ctx, userWithIncorrectEmail.ID); err != nil {
			t.Fatal(err)
		} else if correctEmail := "user2@example.com"; u.Email != correctEmail {
			t.Fatalf("expected email to be %q, got %q", correctEmail, u.Email)
		}

		if u, err := user.Get(ctx, userThatWillBeDeleted.ID); err == nil {
			t.Fatalf("expected user to be deleted, got %v", u)
		}

		if u, err := user.GetByEmail(ctx, "create-me@example.com"); err != nil {
			t.Fatal(err)
		} else if correctName := "Create Me"; u.Name != correctName {
			t.Fatalf("expected name to be %q, got %q", correctName, u.Name)
		}

		updatedUserThatShouldLoseAdmin, err := user.Get(ctx, userThatShouldLoseAdminRole.ID)
		if err != nil {
			t.Fatal(err)
		}
		if updatedUserThatShouldLoseAdmin.Admin {
			t.Fatalf("expected user to lose admin role, but still has it")
		}

		updatedUserWithIncorrectEmail, err := user.Get(ctx, userWithIncorrectEmail.ID)
		if err != nil {
			t.Fatal(err)
		}

		if !updatedUserWithIncorrectEmail.Admin {
			t.Fatalf("expected user to be granted admin role, but doesn't have it")
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

type fakeCred struct{}

var _ authentication.AccessTokenProvider = (*fakeCred)(nil)

func (b fakeCred) GetAuthorizationToken(context context.Context, url *url.URL, additionalAuthenticationContext map[string]any) (string, error) {
	return "HI", nil
}

func (b fakeCred) GetAllowedHostsValidator() *authentication.AllowedHostsValidator {
	return nil
}
