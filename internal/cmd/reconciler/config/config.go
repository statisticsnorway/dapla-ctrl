package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type FeatureFlags struct {
	// AttachSharedVpc enables the shared vpc feature
	AttachSharedVpc bool `env:"FEATURE_ATTACH_SHARED_VPC"`
}

type Config struct {
	FeatureFlags FeatureFlags

	GRPC struct {
		// Target The target address for the gRPC server.
		Target string `env:"GRPC_TARGET,default=127.0.0.1:3001"`

		// InsecureGRPC bypasses authentication, use for development purposes only.
		Insecure bool `env:"INSECURE_GRPC"`
	}

	PubSub struct {
		// SubscriptionID The ID of the Pub/Sub subscription used to listen for events from the NAIS API.
		SubscriptionID string `env:"PUBSUB_SUBSCRIPTION_ID,default=nais-api-reconcilers-api-events"`

		// ProjectID The ID of the Pub/Sub project used to listen for events from the NAIS API. Defaults to GoogleManagementProjectID.
		ProjectID string `env:"PUBSUB_PROJECT_ID,default=$GOOGLE_MANAGEMENT_PROJECT_ID"`
	}

	// GoogleManagementProjectID The ID of the NAIS management project in the tenant organization in GCP.
	GoogleManagementProjectID string `env:"GOOGLE_MANAGEMENT_PROJECT_ID"`

	// ListenAddress The host:port combination used by the http server.
	ListenAddress string `env:"LISTEN_ADDRESS,default=127.0.0.1:3105"`

	// LogFormat Customize the log format. Can be "text" or "json".
	LogFormat string `env:"LOG_FORMAT,default=json"`

	// LogLevel The log level used in api-reconcilers
	LogLevel string `env:"LOG_LEVEL,default=info"`

	// TenantDomain The domain for the tenant.
	TenantDomain string `env:"TENANT_DOMAIN,default=example.com"`

	// TenantName The name of the tenant.
	TenantName string `env:"TENANT_NAME,default=example"`

	// ClusterAlias The cluster alias for legacy migration
	ClusterAlias map[string]string `env:"CLUSTER_ALIAS"`

	// Reconcilers to enable the first time it is registered (one time only) in the NAIS API.
	// If you later would like to enable/disable a reconciler, you can do so through the Console frontend.
	ReconcilersToEnable []string `env:"RECONCILERS_TO_ENABLE"`
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
