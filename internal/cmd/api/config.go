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
	AdminGroupPrefix string `env:"USERSYNC_ADMIN_GROUP_PREFIX,default=console-admins"`

	// Service account to impersonate during user sync
	ServiceAccount string `env:"USERSYNC_SERVICE_ACCOUNT"`

	// SubjectEmail The email address to impersonate during user sync. This is an email address of a user
	// with the necessary permissions to read the Google organization.
	SubjectEmail string `env:"USERSYNC_SUBJECT_EMAIL"`
}

// costConfig is the configuration for the cost service
type costConfig struct {
	ImportEnabled     bool   `env:"COST_DATA_IMPORT_ENABLED"`
	BigQueryProjectID string `env:"BIGQUERY_PROJECTID,default=*detect-project-id*"`
}

// vulnerabilitiesConfig is the configuration for the vulnerability manager using the v13s api
type vulnerabilitiesConfig struct {
	Endpoint       string `env:"VULNERABILITIES_ENDPOINT,default=fake"`
	ServiceAccount string `env:"VULNERABILITIES_SERVICE_ACCOUNT,default=service-account"`
}

// hookdConfig is the configuration for the hookd service
type hookdConfig struct {
	Endpoint string `env:"HOOKD_ENDPOINT,default=http://hookd"`
	PSK      string `env:"HOOKD_PSK"`
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

type unleashConfig struct {
	// BifrostApiEndpoint is the endpoint for the Bifrost API
	BifrostApiUrl string `env:"UNLEASH_BIFROST_API_URL,default=*fake*"`
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
	Unleash  unleashConfig
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
