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

	// Entra ID client secret
	EntraIdClientSecret string `env:"ENTRA_ID_CLIENT_SECRET"`

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

	// ClientSecret The client secret to use for the OAuth login flow.
	ClientSecret string `env:"OAUTH_CLIENT_SECRET"`

	// RedirectURL The URL that Google will redirect back to after performing authentication.
	RedirectURL string `env:"OAUTH_REDIRECT_URL"`

	// AdditionalScopes is a list of additional scopes to request in the OAuth login flow.
	AdditionalScopes []string `env:"OAUTH_ADDITIONAL_SCOPES"`
}

type Config struct {
	// Tenant is the active tenant
	Tenant string `env:"TENANT,default=dev-nais"`

	// TenantDomain The domain for the tenant.
	TenantDomain string `env:"TENANT_DOMAIN,default=example.com"`

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

	// GRPCListenAddress is host:port combination used by the GRPC server
	GRPCListenAddress string `env:"GRPC_LISTEN_ADDRESS,default=127.0.0.1:3001"`

	LeaseName      string `env:"LEASE_NAME,default=nais-api-lease"`
	LeaseNamespace string `env:"LEASE_NAMESPACE,default=nais-system"`

	// ReplaceEnvironmentNames is a map of cluster names to replace in the UI. Keys are cluster names used in
	// Kubernetes, for instance "prod", and the values are user-facing environment names, for instance "prod-gcp". This
	// configuration value is only used by the nav.no tenant.
	ReplaceEnvironmentNames map[string]string `env:"REPLACE_ENVIRONMENT_NAMES, noinit"`

	Usersync usersyncConfig
	OAuth    oAuthConfig

	Fakes    Fakes
}

type Fakes struct {
	WithInsecureUserHeader bool `env:"WITH_INSECURE_USER_HEADER"`
	WithFakeKubernetes     bool `env:"WITH_FAKE_KUBERNETES"`
	WithFakeCloudSQL       bool `env:"WITH_FAKE_CLOUD_SQL"`
	WithFakePrometheus     bool `env:"WITH_FAKE_PROMETHEUS"`
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
