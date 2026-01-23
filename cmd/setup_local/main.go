package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"
	"unicode"

	pubsub "cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/google/uuid"
	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/internal/activitylog"
	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/graph/model"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/group"
	"github.com/statisticsnorway/dapla-api/internal/logger"
	"github.com/statisticsnorway/dapla-api/internal/section"
	"github.com/statisticsnorway/dapla-api/internal/slug"
	"github.com/statisticsnorway/dapla-api/internal/team"
	"github.com/statisticsnorway/dapla-api/internal/user"
	"github.com/statisticsnorway/dapla-api/internal/usersync/usersyncer"
	"github.com/statisticsnorway/dapla-api/internal/usersync/usersyncsql"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/utils/ptr"
)

const (
	exitCodeSuccess = iota
	exitCodeConfigError
	exitCodeLoggerError
	exitCodeRunError
)

type seedConfig struct {
	DatabaseURL               string `env:"DATABASE_URL,default=postgres://api:api@localhost:3002/api?sslmode=disable"`
	Domain                    string `env:"TENANT_DOMAIN,default=example.com"`
	GoogleManagementProjectID string `env:"GOOGLE_MANAGEMENT_PROJECT_ID,default=nais-local-dev"`

	NumUsers          *int
	NumTeams          *int
	NumOwnersPerTeam  *int
	NumMembersPerTeam *int
	ForceSeed         *bool
	ProvisionPubSub   *bool
}

func newSeedConfig(ctx context.Context) (*seedConfig, error) {
	cfg := &seedConfig{}
	if err := envconfig.Process(ctx, cfg); err != nil {
		return nil, err
	}

	cfg.NumUsers = flag.Int("users", 1000, "number of users to insert")
	cfg.NumTeams = flag.Int("teams", 200, "number of teams to insert")
	cfg.NumOwnersPerTeam = flag.Int("owners", 3, "number of owners per team")
	cfg.NumMembersPerTeam = flag.Int("members", 10, "number of members per team")
	cfg.ForceSeed = flag.Bool("force", false, "seed regardless of existing database content")
	cfg.ProvisionPubSub = flag.Bool("provision_pub_sub", true, "set up pubsub credentials")
	flag.Parse()

	return cfg, nil
}

func main() {
	ctx := context.Background()
	log, err := logger.New("text", "INFO")
	if err != nil {
		fmt.Printf("log error: %s", err)
		os.Exit(exitCodeLoggerError)
	}

	cfg, err := newSeedConfig(ctx)
	if err != nil {
		log.WithError(err).Errorf("configuration error")
		os.Exit(exitCodeConfigError)
	}

	if err := run(ctx, cfg, log); err != nil {
		log.WithError(err).Errorf("fatal error in run()")
		os.Exit(exitCodeRunError)
	}

	os.Exit(exitCodeSuccess)
}

func run(ctx context.Context, cfg *seedConfig, log logrus.FieldLogger) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if *cfg.ProvisionPubSub {
		log.Infof("Provisioning pubsub")

		if err := os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:3004"); err != nil {
			return err
		}

		projectID := cfg.GoogleManagementProjectID
		client, err := pubsub.NewClient(ctx, projectID)
		if err != nil {
			return err
		}

		log.Infof("creating topic")

		topic := &pubsubpb.Topic{
			Name: fmt.Sprintf("projects/%s/topics/%s", projectID, "dapla-api"),
		}
		if _, err := client.TopicAdminClient.CreateTopic(ctx, topic); err != nil {
			if s, ok := status.FromError(err); !ok || s.Code() != codes.AlreadyExists {
				return err
			}
		}

		log.Infof("creating subscription")

		subscription := &pubsubpb.Subscription{
			Name:                     fmt.Sprintf("projects/%s/subscriptions/%s", projectID, "dapla-api-reconcilers-api-events"),
			Topic:                    topic.Name,
			MessageRetentionDuration: durationpb.New(1 * time.Hour),
		}
		if _, err := client.SubscriptionAdminClient.CreateSubscription(ctx, subscription); err != nil {
			if s, ok := status.FromError(err); !ok || s.Code() != codes.AlreadyExists {
				return err
			}
		}
	}

	firstNames, err := fileToSlice("data/first_names.txt")
	if err != nil {
		return err
	}
	numFirstNames := len(firstNames)

	lastNames, err := fileToSlice("data/last_names.txt")
	if err != nil {
		return err
	}
	numLastNames := len(lastNames)

	log.Infof("initializing database")

	pool, err := database.New(ctx, cfg.DatabaseURL, log)
	if err != nil {
		return err
	}
	defer pool.Close()

	ctx = database.NewLoaderContext(ctx, pool)
	ctx = activitylog.NewLoaderContext(ctx, pool)
	ctx = user.NewLoaderContext(ctx, pool)
	ctx = team.NewLoaderContext(ctx, pool)
	ctx = authz.NewLoaderContext(ctx, pool)
	ctx = group.NewLoaderContext(ctx, pool)
	ctx = section.NewLoaderContext(ctx, pool)

	emails := map[string]struct{}{}
	slugs := map[slug.Slug]struct{}{}

	if !*cfg.ForceSeed {
		if existingUsers, err := getAllUsers(ctx); err != nil {
			return fmt.Errorf("fetch existing users: %w", err)
		} else if len(existingUsers) != 0 {
			return fmt.Errorf("database already has users, abort")
		}

		if existingTeams, err := getAllTeams(ctx); err != nil {
			return fmt.Errorf("fetch existing teams: %w", err)
		} else if len(existingTeams) != 0 {
			return fmt.Errorf("database already has teams, abort")
		}
	} else {
		users, err := getAllUsers(ctx)
		if err != nil {
			return fmt.Errorf("fetch existing users: %w", err)
		}
		for _, u := range users {
			emails[u.Email] = struct{}{}
		}

		teams, err := getAllTeams(ctx)
		if err != nil {
			return fmt.Errorf("fetch existing teams: %w", err)
		}
		for _, t := range teams {
			slugs[t.Slug] = struct{}{}
		}
	}

	err = database.Transaction(ctx, func(ctx context.Context) error {
		const (
			adminName = "admin usersen"
			devName   = "dev usersen"
		)

		var err error
		var adminUser, devUser *user.User

		usersyncq := usersyncsql.New(database.TransactionFromContext(ctx))

		createUser := func(ctx context.Context, name, email, sectionCode string) (*user.User, error) {
			usu, err := usersyncq.Create(ctx, usersyncsql.CreateParams{
				Name:        name,
				Email:       email,
				ExternalID:  uuid.New().String(),
				SectionCode: &sectionCode,
			})
			if err != nil {
				return nil, fmt.Errorf("create user: %w", err)
			}

			usr, err := user.GetByEmail(ctx, usu.Email)
			if err != nil {
				return nil, fmt.Errorf("get user: %w", err)
			}

			return usr, nil
		}

		sections, err := getAllSectionCodes(ctx)
		if err != nil {
			return fmt.Errorf("get section codes: %w", err)
		}
		numSections := len(sections)

		adminUser, err = user.GetByEmail(ctx, nameToEmail(adminName, cfg.Domain))
		if err != nil {
			adminUser, err = createUser(ctx, adminName, nameToEmail(adminName, cfg.Domain), sections[rand.IntN(numSections)])
			if err != nil {
				return fmt.Errorf("create admin user: %w", err)
			}
		}

		if err := usersyncq.AssignGlobalAdmin(ctx, adminUser.UUID); err != nil {
			return fmt.Errorf("assign global admin role to admin user: %w", err)
		}
		actor := &authz.Actor{User: adminUser}

		devUser, err = user.GetByEmail(ctx, nameToEmail(devName, cfg.Domain))
		if err != nil {
			devUser, err = createUser(ctx, devName, nameToEmail(devName, cfg.Domain), sections[rand.IntN(numSections)])
			if err != nil {
				return fmt.Errorf("create dev user: %w", err)
			}
		}

		if err := usersyncer.AssignDefaultPermissionsToUser(ctx, usersyncq, devUser.UUID); err != nil {
			return fmt.Errorf("assign default permissions to dev user: %w", err)
		}

		users := []*user.User{devUser}
		for i := 1; i <= *cfg.NumUsers; i++ {
			firstName := firstNames[rand.IntN(numFirstNames)]
			lastName := lastNames[rand.IntN(numLastNames)]
			section := sections[rand.IntN(numSections)]
			name := firstName + " " + lastName
			email := nameToEmail(name, cfg.Domain)
			if _, exists := emails[email]; exists {
				continue
			}

			u, err := createUser(ctx, name, email, section)
			if err != nil {
				return fmt.Errorf("create user %q: %w", email, err)
			}

			if err = usersyncer.AssignDefaultPermissionsToUser(ctx, usersyncq, u.UUID); err != nil {
				return fmt.Errorf("assign default permissions to user %q: %w", u.Email, err)
			}

			log.Infof("%d/%d users created", i, *cfg.NumUsers)
			users = append(users, u)
			emails[email] = struct{}{}
		}

		var devteam *team.Team
		devteam, err = team.Get(ctx, "devteam")
		if err != nil {
			input := &team.CreateTeamInput{
				Slug:        "devteam",
				DisplayName: "Dev Team",
				Purpose:     "dev-purpose",
				SectionCode: "724",
				IsManaged:   ptr.To(true),
			}
			devteam, err = team.Create(ctx, input, actor)
			if err != nil {
				return fmt.Errorf("create devteam: %w", err)
			}
		}
		// devuser is first in array
		createGroupAndAddUsers(ctx, actor, devteam.Slug, "developers", nil, users[:1], 1)

		createGroupAndAddUsers(ctx, actor, devteam.Slug, "managers", nil, users[:1], 1)

		for i := 1; i <= *cfg.NumTeams; i++ {
			name := teamName()
			if _, exists := slugs[name]; exists {
				continue
			}

			input := &team.CreateTeamInput{
				Slug:        name,
				DisplayName: strings.ToTitle(name.String()),
				Purpose:     "some purpose",
				SectionCode: "724",
				IsManaged:   ptr.To(true),
			}
			_, err := team.Create(ctx, input, actor)
			if err != nil {
				return fmt.Errorf("create team %q: %w", name, err)
			}

			log.Infof("%d/%d teams created", i, *cfg.NumTeams)

			suffixes := []string{
				"wizards",
				"ninjas",
				"cowboys",
				"10x",
			}

			createGroupAndAddUsers(ctx, actor, name, "managers", nil, users, rand.IntN(2)+1)
			createGroupAndAddUsers(ctx, actor, name, "developers", nil, users, rand.IntN(3)+1)
			createGroupAndAddUsers(ctx, actor, name, "developers", &suffixes[rand.IntN(len(suffixes))], users, rand.IntN(2)+1)

			log.Infof("\tGroups created and users added")
			slugs[name] = struct{}{}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error during transaction: %w", err)
	}

	log.Infof("done")
	return nil
}

func createGroupAndAddUsers(ctx context.Context, actor *authz.Actor, team slug.Slug, teamCategory string, suffix *string, users []*user.User, membersToAdd int) {
	createdGroup, _ := group.Create(ctx, &group.CreateGroupInput{
		TeamSlug: team,
		Category: teamCategory,
		Suffix:   suffix,
	}, actor)

	i := 0
	if len(users)-membersToAdd > 0 {
		i = rand.IntN(len(users) - membersToAdd)
	}
	for index := range membersToAdd {
		user := users[i+index].UUID
		err := group.AddMember(ctx, group.AddGroupMemberInput{
			GroupName: createdGroup.Name,
			UserID:    user,
		}, actor)
		if err != nil {
			fmt.Printf("error adding user: %s", err)
		}
	}
}

func teamName() slug.Slug {
	letters := []byte("abcdefghijklmnopqrstuvwxyz")
	b := make([]byte, 10)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return slug.Slug(b)
}

func nameToEmail(name, domain string) string {
	name = strings.NewReplacer(" ", ".", "æ", "ae", "ø", "oe", "å", "aa").Replace(strings.ToLower(name))
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	name, _, _ = transform.String(t, name)
	return name + "@" + domain
}

func fileToSlice(path string) ([]string, error) {
	file, err := os.Open(path) // #nosec: G304
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

func getAllUsers(ctx context.Context) ([]*user.User, error) {
	first := 100
	allUsers := make([]*user.User, 0)
	orderBy := &user.UserOrder{
		Field:     user.UserOrderFieldName,
		Direction: model.OrderDirectionAsc,
	}
	var after *pagination.Cursor
	for {
		p, err := pagination.ParsePage(&first, after, nil, nil)
		if err != nil {
			return nil, err
		}
		conn, err := user.List(ctx, p, orderBy)
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, conn.Nodes()...)
		if !conn.PageInfo.HasNextPage {
			break
		}
		after = conn.PageInfo.EndCursor
	}

	return allUsers, nil
}

func getAllTeams(ctx context.Context) ([]*team.Team, error) {
	first := 100
	allTeams := make([]*team.Team, 0)
	orderBy := &team.TeamOrder{
		Field:     team.TeamOrderFieldSlug,
		Direction: model.OrderDirectionAsc,
	}
	var after *pagination.Cursor
	for {
		p, err := pagination.ParsePage(&first, after, nil, nil)
		if err != nil {
			return nil, err
		}
		conn, err := team.List(ctx, p, orderBy)
		if err != nil {
			return nil, err
		}
		allTeams = append(allTeams, conn.Nodes()...)
		if !conn.PageInfo.HasNextPage {
			break
		}
		after = conn.PageInfo.EndCursor
	}

	return allTeams, nil
}

func getAllSectionCodes(ctx context.Context) ([]string, error) {
	first := 100
	allCodes := make([]string, 0)
	orderBy := &section.SectionOrder{
		Field:     section.SectionOrderFieldCode,
		Direction: model.OrderDirectionAsc,
	}
	var after *pagination.Cursor
	for {
		p, err := pagination.ParsePage(&first, after, nil, nil)
		if err != nil {
			return nil, err
		}
		conn, err := section.List(ctx, p, orderBy)
		if err != nil {
			return nil, err
		}
		for _, node := range conn.Nodes() {
			allCodes = append(allCodes, node.Code)
		}
		if !conn.PageInfo.HasNextPage {
			break
		}
		after = conn.PageInfo.EndCursor
	}

	return allCodes, nil
}
