//go:build integration_test

package usersyncer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/microsoft/kiota-abstractions-go/authentication"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/section"
	"github.com/statisticsnorway/dapla-api/internal/test"
	"github.com/statisticsnorway/dapla-api/internal/user"
	"github.com/statisticsnorway/dapla-api/internal/usersync/usersyncer"
	"github.com/statisticsnorway/dapla-api/internal/usersync/usersyncsql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
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
		ctx = section.NewLoaderContext(ctx, pool)
		return ctx, pool
	}

	sectionManagerRegex := regexp.MustCompile("^(Seksjonssjef|Forskningsleder)")

	t.Run("No local users, no remote users", func(t *testing.T) {
		ctx, pool := setup(t)

		httpClient := test.NewTestHttpClient(
			func(req *http.Request) *http.Response {
				// All users response
				return test.Response("200 OK", `{"value":[]}`)
			},
			func(req *http.Request) *http.Response {
				// Admin users response
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
			New(pool, allUsersGroup, adminGroup, client, sectionManagerRegex, log).
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
				// All users response
				return test.Response("200 OK", `{"value":[]}`)
			},
			func(req *http.Request) *http.Response {
				// Admin users response
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
			New(pool, allUsersGroup, adminGroup, client, sectionManagerRegex, log).
			Sync(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Usersync should have removed all local users, as there are no remote users
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

		userWithOutdatedSection, err := querier.Create(ctx, usersyncsql.CreateParams{
			Name:       "I Am No Longer In Section 723",
			Email:      "fix-my-section@example.com",
			ExternalID: "5",
		})
		if err != nil {
			t.Fatal(err)
		}

		if err := querier.AssignGlobalAdmin(ctx, userThatShouldLoseAdminRole.ID); err != nil {
			t.Fatal(err)
		}

		httpClient := test.NewTestHttpClient(
			func(req *http.Request) *http.Response {
				// All users response
				return test.Response("200 OK", generateEntraIdResponse(
					externalUser{Id: "1", Email: "user1@example.com", Name: "Correct Name"},                                                                 // Will update name of local user
					externalUser{Id: "2", Email: "user2@example.com", Name: "Some Name"},                                                                    // Will update euserPrincipalName of local user
					externalUser{Id: "4", Email: "should-lose-admin@example.com", Name: "Should Lose Admin"},                                                // Will lose admin role
					externalUser{Id: "5", Email: "fix-my-section@example.com", Name: "I Am No Longer In Section 723", Section: "O 724 Seksjon for Testing"}, // Will get their section updated
					externalUser{Id: "6", Email: "create-me@example.com", Name: "Create Me", Section: "O 724 Seksjon for Testing"},                          // Will be created
					externalUser{Id: "7", Email: "invalid-section@example.com", Name: "My Section Doesn't Exist", Section: "O 1337 Seksjon for fantasi"}),   // Will be created, with section NULL
				)
			},
			func(req *http.Request) *http.Response {
				// Admin users response
				return test.Response("200 OK", generateEntraIdResponse(
					externalUser{Id: "2", Email: "user2@example.com", Name: "Some Name"},              // Will be granted admin role
					externalUser{Id: "8", Email: "unknown-admin@example.com", Name: "Unknown Admin"}), // Unknown user, will be logged
				)
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
			New(pool, allUsersGroup, adminGroup, client, sectionManagerRegex, log).
			Sync(ctx)
		if err != nil {
			t.Fatal(err)
		}

		p, _ := pagination.ParsePage(nil, nil, nil, nil)
		if users, err := user.List(ctx, p, nil); err != nil {
			t.Fatal(err)
		} else if total := users.PageInfo.TotalCount; total != 6 {
			t.Errorf("expected 6 users, got %d", total)
		}

		if u, err := user.Get(ctx, userWithIncorrectName.ID); err != nil {
			t.Fatal(err)
		} else if correctName := "Correct Name"; u.Name != correctName {
			t.Errorf("expected name to be %q, got %q", correctName, u.Name)
		}

		if u, err := user.Get(ctx, userWithIncorrectEmail.ID); err != nil {
			t.Fatal(err)
		} else if correctEmail := "user2@example.com"; u.Email != correctEmail {
			t.Errorf("expected email to be %q, got %q", correctEmail, u.Email)
		}

		if u, err := user.Get(ctx, userThatWillBeDeleted.ID); err == nil {
			t.Errorf("expected user to be deleted, got %v", u)
		}

		u, err := user.GetByEmail(ctx, "create-me@example.com")
		if err != nil {
			t.Fatal(err)
		}
		if correctName := "Create Me"; u.Name != correctName {
			t.Errorf("expected name to be %q, got %q", correctName, u.Name)
		}
		if section := u.SectionCode; section == nil {
			t.Errorf("expected user %q to have section 724, got nil", u.Email)
		} else if *section != "724" {
			t.Errorf("expected user %q to have section 724, got %q", u.Email, *section)
		}

		if updatedUserThatShouldLoseAdmin, err := user.Get(ctx, userThatShouldLoseAdminRole.ID); err != nil {
			t.Fatal(err)
		} else if updatedUserThatShouldLoseAdmin.Admin {
			t.Errorf("expected user to lose admin role, but still has it")
		}

		if updatedUserWithIncorrectEmail, err := user.Get(ctx, userWithIncorrectEmail.ID); err != nil {
			t.Fatal(err)
		} else if !updatedUserWithIncorrectEmail.Admin {
			t.Errorf("expected user to be granted admin role, but doesn't have it")
		}

		if updatedUserWithOutdatedSection, err := user.Get(ctx, userWithOutdatedSection.ID); err != nil {
			t.Fatal(err)
		} else if updatedUserWithOutdatedSection.SectionCode == nil {
			t.Errorf("expected user %q to have section 724, got nil", userWithOutdatedSection.Email)
		} else if *updatedUserWithOutdatedSection.SectionCode != "724" {
			t.Errorf("expected user %q to have section 724, got %q", userWithOutdatedSection.Email, *userWithOutdatedSection.SectionCode)
		}

		if updatedUserWithInvalidSection, err := user.GetByEmail(ctx, "invalid-section@example.com"); err != nil {
			t.Fatal(err)
		} else if updatedUserWithInvalidSection.SectionCode != nil {
			t.Errorf("expected user %q to have no section, got %q", updatedUserWithInvalidSection.Email, *updatedUserWithInvalidSection.SectionCode)
		}
	})

	t.Run("create users and set as section managers", func(t *testing.T) {
		ctx, pool := setup(t)
		querier := usersyncsql.New(pool)

		type testSection struct {
			Code string
			Name string
		}
		fullName := func(s testSection) string {
			return fmt.Sprintf("O %s %s", s.Code, s.Name)
		}
		sectionA := testSection{
			Code: "666",
			Name: "Seksjon for seksjoner",
		}
		sectionB := testSection{
			Code: "667",
			Name: "Seksjon for flere seksjoner",
		}

		// Create two users, one who is currently manager in the API's database,
		// and a new manager to replace them.
		oldBossA := externalUser{Id: "1", Email: "goodbye@example.com", Name: "Pen Sjonist", Section: fullName(sectionA), JobTitle: "Pensjonist"} // Will be removed
		bossUserA := externalUser{Id: "2", Email: "eljefe@example.com", Name: "Das Boss", Section: fullName(sectionA), JobTitle: "Seksjonssjef"}  // Will be created

		// Create forskningsleder for 667
		bossUserB := externalUser{Id: "3", Email: "einstein@example.com", Name: "Ein Stein", Section: fullName(sectionB), JobTitle: "Forskningsleder"} // Will be created

		// Create the old boss in the database so we can use him as manager
		oldBossDbUser, err := querier.Create(ctx, usersyncsql.CreateParams{
			Name:       oldBossA.Name,
			Email:      oldBossA.Email,
			ExternalID: oldBossA.Id,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create a section with our old boss as manager
		if s, err := querier.CreateSection(ctx, usersyncsql.CreateSectionParams{
			Code:      sectionA.Code,
			Name:      sectionA.Name,
			ManagerID: &(oldBossDbUser.ID),
		}); err != nil {
			t.Fatal(err)
		} else if s.ManagerID == nil {
			t.Fatalf("manager_id is nil, expected %q", oldBossDbUser.ID)
		} else if *s.ManagerID != oldBossDbUser.ID {
			t.Fatalf("manager_id is %q, expected %q", *s.ManagerID, oldBossDbUser.ID)
		}

		// Create a section with no manager
		if s, err := querier.CreateSection(ctx, usersyncsql.CreateSectionParams{
			Code:      sectionB.Code,
			Name:      sectionB.Name,
			ManagerID: nil,
		}); err != nil {
			t.Fatal(err)
		} else if s.ManagerID != nil {
			t.Fatalf("manager_id is not nil: %q", *s.ManagerID)
		}

		httpClient := test.NewTestHttpClient(
			// All users response
			func(req *http.Request) *http.Response {
				return test.Response("200 OK", generateEntraIdResponse(oldBossA, bossUserA, bossUserB))
			},
			// Admin users response
			func(req *http.Request) *http.Response {
				return test.Response("200 OK", generateEntraIdResponse())
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

		// Run usersync, this should replace our old boss with the new one as manager of the section
		err = usersyncer.
			New(pool, allUsersGroup, adminGroup, client, sectionManagerRegex, log).
			Sync(ctx)
		if err != nil {
			t.Fatal(err)
		}

		newBossADbUser, err := user.GetByEmail(ctx, bossUserA.Email)
		if err != nil {
			t.Fatal(err)
		} else if newBossADbUser.ExternalID != bossUserA.Id || newBossADbUser.Name != bossUserA.Name {
			t.Fatalf("expected external_id=%q,name=%q, got external_id=%q,name=%q", bossUserA.Id, bossUserA.Name, newBossADbUser.ExternalID, newBossADbUser.Name)
		}

		newBossBDbUser, err := user.GetByEmail(ctx, bossUserB.Email)
		if err != nil {
			t.Fatal(err)
		} else if newBossBDbUser.ExternalID != bossUserB.Id || newBossBDbUser.Name != bossUserB.Name {
			t.Fatalf("expected external_id=%q,name=%q, got external_id=%q,name=%q", bossUserB.Id, bossUserB.Name, newBossBDbUser.ExternalID, newBossBDbUser.Name)
		}

		if s, err := section.Get(ctx, sectionA.Code); err != nil {
			t.Fatal(err)
		} else if s.ManagerId == nil {
			t.Fatal("section does not have manager set")
		} else if *s.ManagerId != newBossADbUser.UUID {
			t.Fatalf("expected manager_id=%q, got %q", newBossADbUser.UUID, *s.ManagerId)
		}

		if s, err := section.Get(ctx, sectionB.Code); err != nil {
			t.Fatal(err)
		} else if s.ManagerId == nil {
			t.Fatal("section does not have manager set")
		} else if *s.ManagerId != newBossBDbUser.UUID {
			t.Fatalf("expected manager_id=%q, got %q", newBossADbUser.UUID, *s.ManagerId)
		}
	})

	t.Run("demote section manager because job title is no longer seksjonssjef", func(t *testing.T) {
		ctx, pool := setup(t)
		querier := usersyncsql.New(pool)

		sectionCode := "666"
		sectionName := "Seksjon for seksjoner"
		sectionFullName := fmt.Sprintf("O %s %s", sectionCode, sectionName)

		// The user to demote (imagine they previously had the title "Seksjonssjef")
		demoted := externalUser{Id: "1", Email: "ex-jefe@example.com", Name: "No Longer Boss", Section: sectionFullName, JobTitle: "Ikke-seksjonssjef"} // Will be created

		dbUser, err := querier.Create(ctx, usersyncsql.CreateParams{
			Name:       demoted.Name,
			Email:      demoted.Email,
			ExternalID: demoted.Id,
		})
		if err != nil {
			t.Fatal(err)
		}

		// We create a section with our predefined user as manager
		if s, err := querier.CreateSection(ctx, usersyncsql.CreateSectionParams{
			Code:      sectionCode,
			Name:      sectionName,
			ManagerID: &(dbUser.ID),
		}); err != nil {
			t.Fatal(err)
		} else if s.ManagerID == nil {
			t.Fatalf("manager_id is nil, expected %q", dbUser.ID)
		} else if *s.ManagerID != dbUser.ID {
			t.Fatalf("manager_id is %q, expected %q", *s.ManagerID, dbUser.ID)
		}

		httpClient := test.NewTestHttpClient(
			// All users response
			func(req *http.Request) *http.Response {
				return test.Response("200 OK", generateEntraIdResponse(demoted))
			},
			// Admin users response
			func(req *http.Request) *http.Response {
				return test.Response("200 OK", generateEntraIdResponse())
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

		// Run usersync, this should demote the ex-manager
		err = usersyncer.
			New(pool, allUsersGroup, adminGroup, client, sectionManagerRegex, log).
			Sync(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Check that our section no longer has a manager
		if s, err := section.Get(ctx, sectionCode); err != nil {
			t.Fatal(err)
		} else if s.ManagerId != nil {
			t.Fatalf("section has manager set: %s", *s.ManagerId)
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

type externalUser struct {
	Id       string `json:"id"`
	Email    string `json:"userPrincipalName"`
	Name     string `json:"displayName"`
	Section  string `json:"department"`
	JobTitle string `json:"jobTitle"`
}

func generateEntraIdResponse(users ...externalUser) string {
	type eIdResponseNode struct {
		externalUser
		DataType string `json:"@odata.type"`
	}

	type eIdResponse struct {
		Value []eIdResponseNode `json:"value"`
	}

	res := eIdResponse{}
	for _, u := range users {
		res.Value = append(res.Value, eIdResponseNode{u, "#microsoft.graph.user"})
	}

	resBytes, _ := json.Marshal(res)
	return string(resBytes)
}
