package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Fakes     FakesConfig
	Server    ServerConfig
	Router    RouterConfig
	Parquedit ParqueditConfig
}

type ServerConfig struct {
	ListenAddr string `env:"LISTEN_ADDR" envDefault:":8080"`
}

type FakesConfig struct {
	WithFakeCloudResourceManager   bool   `env:"WITH_FAKE_CLOUD_RESOURCE_MANAGER"`
	WithFakeSqlAdmin               bool   `env:"WITH_FAKE_SQL_ADMIN"`
	FakeSqlAdminDatabaseConnString string `env:"FAKE_SQL_ADMIN_DB_CONN_STRING"`
}

type RouterConfig struct {
	AuthToken string `env:"AUTH_TOKEN,required"`
}

type ParqueditConfig struct {
	DatabaseUrl        string `env:"PARQUEDIT_DATABASE_URL,required"`
	CloudSQLProject    string `env:"PARQUEDIT_CLOUDSQL_PROJECT"`
	CloudSQLInstance   string `env:"PARQUEDIT_CLOUDSQL_INSTANCE"`
	CloudSqlUserSuffix string `env:"PARQUEDIT_CLOUDSQL_USER_SUFFIX"` // e.g. "-developers@my-project-1d.iam.gserviceaccount.com"
}

func ParseConfig[T any]() (T, error) {
	result, err := env.ParseAs[T]()
	if err != nil {
		return *new(T), err
	}
	return result, nil
}
