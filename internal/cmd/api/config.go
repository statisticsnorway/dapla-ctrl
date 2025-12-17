package api

import (
	"context"
	"reflect"

	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
)

type usersyncConfig struct {
	// Enabled When set to true api will keep the user database in sync with the connected Google
	// organization. The Google organization will be treated as the master.
	Enabled bool `env:"USERSYNC_ENABLED"`

	// AdminGroupPrefix The prefix of the admin group email address.
	AdminGroup string `env:"USERSYNC_ADMIN_GROUP"`

	// Entra ID client ID to use for authenticating
	EntraIdClientId string `env:"ENTRA_ID_CLIENT_ID"`

	// Entra ID tenant ID
	EntraIdTenantId string `env:"ENTRA_ID_TENANT_ID"`

	// Entra ID group containing all valid users
	// Required when usersyncer is enabled
	AllUsersGroup string `env:"ENTRA_ID_ALL_USERS_GROUP"`
}

type oAuthConfig struct {
	// Issuer The issuer of the OAuth 2.0 client to use for the OAuth login flow.
	Issuer string `env:"OAUTH_ISSUER,default=https://accounts.google.com"`

	// ClientID The ID of the OAuth 2.0 client to use for the OAuth login flow.
	ClientID string `env:"OAUTH_CLIENT_ID"`

	// RedirectURL The URL that Google will redirect back to after performing authentication.
	RedirectURL string `env:"OAUTH_REDIRECT_URL"`

	// AdditionalScopes is a list of additional scopes to request in the OAuth login flow.
	AdditionalScopes []string `env:"OAUTH_ADDITIONAL_SCOPES"`
}

type grpcConfig struct {
	// ListenAddress is host:port combination used by the GRPC server
	ListenAddress string `env:"GRPC_LISTEN_ADDRESS,default=127.0.0.1:3001"`

	// Which client SAs we trust
	ExpectedServiceAccounts []string `env:"GRPC_EXPECTED_CLIENT_SA_LIST"`
}

type JWTConfig struct {
	// The issuer to trust for JWT Bearer tokens
	Issuer string `env:"JWT_ISSUER,default=https://auth-play.test.ssb.no/realms/ssb"`
	// Required token audience
	Audience string `env:"JWT_AUDIENCE,default=dapla-api"`
	// Which claim in the token to extract the user's email from
	EmailClaim string `env:"JWT_EMAIL_CLAIM,default=preferred_username"`
	// Set to true to disable the JWT middleware
	SkipMiddleware bool `env:"JWT_SKIP_MIDDLEWARE,default=false"`
}

type Config struct {
	// GoogleManagementProjectID The ID of the Nais management project in the tenant organization in GCP.
	GoogleManagementProjectID string `env:"GOOGLE_MANAGEMENT_PROJECT_ID"`

	// DatabaseConnectionString is the database DSN
	DatabaseConnectionString string `env:"DATABASE_URL,default=postgres://api:api@127.0.0.1:3002/api?sslmode=disable"`

	LogFormat string `env:"LOG_FORMAT,default=json"`
	LogLevel  string `env:"LOG_LEVEL,default=info"`

	WithSlowQueryLogger bool `env:"WITH_SLOW_QUERY_LOGGER"`

	// ListenAddress is host:port combination used by the http server
	ListenAddress         string `env:"LISTEN_ADDRESS,default=127.0.0.1:3000"`
	InternalListenAddress string `env:"INTERNAL_LISTEN_ADDRESS,default=127.0.0.1:3005"`

	LeaderElectionEnabled bool `env:"LEADER_ELECTION_ENABLED,default=true"`

	Usersync usersyncConfig
	OAuth    oAuthConfig
	JWT      JWTConfig
	Grpc     grpcConfig

	Fakes Fakes
}

type Fakes struct {
	WithInsecureAuth   bool `env:"WITH_INSECURE_AUTH"`
	WithFakeCloudSQL   bool `env:"WITH_FAKE_CLOUD_SQL"`
	WithFakePrometheus bool `env:"WITH_FAKE_PROMETHEUS"`
}

func (f Fakes) Inform(log logrus.FieldLogger) {
	v := reflect.ValueOf(f)
	for i := range v.NumField() {
		field := v.Type().Field(i)
		if v.Field(i).Bool() {
			log.Warnf("%s is true", field.Name)
		}
	}
}

// NewConfig creates a new configuration instance from environment variables
func NewConfig(ctx context.Context, lookuper envconfig.Lookuper) (*Config, error) {
	cfg := &Config{}
	err := envconfig.ProcessWith(ctx, &envconfig.Config{
		Target:   cfg,
		Lookuper: lookuper,
	})
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
