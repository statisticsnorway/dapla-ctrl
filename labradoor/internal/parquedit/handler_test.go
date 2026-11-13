package parquedit

import (
	"strings"
	"testing"

	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/config"
)

func TestValidateDataPathTemplate(t *testing.T) {
	for _, tt := range []struct {
		name     string
		template string
		wantErr  bool
	}{
		{name: "production", template: "gs://ssb-{team}-data-produkt-prod/.parquedit_data/"},
		{name: "test bucket", template: "gs://parquedit-test/{team}/.parquedit_data/"},
		{name: "test per-team bucket", template: "gs://ssb-{team}-data-produkt-test/.parquedit_data/"},
		{name: "multiple team placeholders", template: "gs://{team}-test/{team}/"},
		{name: "quoted path", template: "gs://parquedit-test/{team}/team's-data/"},
		{name: "empty", wantErr: true},
		{name: "missing team placeholder", template: "gs://parquedit-test/shared/", wantErr: true},
		{name: "unknown placeholder", template: "gs://parquedit-{env}/{team}/", wantErr: true},
		{name: "unmatched placeholder", template: "gs://parquedit-test/{team}/{other/", wantErr: true},
		{name: "wrong scheme", template: "https://parquedit-test/{team}/", wantErr: true},
		{name: "missing bucket", template: "gs:///{team}/", wantErr: true},
		{name: "missing path", template: "gs://{team}-test", wantErr: true},
		{name: "credentials", template: "gs://user:password@parquedit-test/{team}/", wantErr: true},
		{name: "port", template: "gs://parquedit-test:443/{team}/", wantErr: true},
		{name: "empty port", template: "gs://parquedit-test:/{team}/", wantErr: true},
		{name: "query", template: "gs://parquedit-test/{team}/?key=value/", wantErr: true},
		{name: "empty query", template: "gs://parquedit-test/{team}/?", wantErr: true},
		{name: "fragment", template: "gs://parquedit-test/{team}/#fragment/", wantErr: true},
		{name: "whitespace", template: "gs://parquedit-test/{team}/data path/", wantErr: true},
		{name: "invalid escaping", template: "gs://parquedit-test/{team}/%zz/", wantErr: true},
		{name: "missing trailing slash", template: "gs://parquedit-test/{team}/.parquedit_data", wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDataPathTemplate(tt.template)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateDataPathTemplate() = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewRejectsInvalidDataPathBeforeConnecting(t *testing.T) {
	client, err := New(t.Context(), config.ParqueditConfig{
		DatabaseUrl:      "invalid database URL",
		DataPathTemplate: "gs://test-bucket/without-team/",
	}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "PARQUEDIT_DATA_PATH_TEMPLATE") {
		t.Fatalf("New() error = %v, want data path configuration error", err)
	}
	if client != nil {
		t.Fatal("New() returned a client for invalid configuration")
	}
}

func TestDeveloperServiceAccountUsesEnvironmentSuffix(t *testing.T) {
	for _, tt := range []struct {
		name        string
		suffix      string
		wantBinding string
		wantSQLUser string
	}{
		{
			name:        "production",
			suffix:      "-developers@prod-project.iam.gserviceaccount.com",
			wantBinding: "serviceAccount:example-team-developers@prod-project.iam.gserviceaccount.com",
			wantSQLUser: "example-team-developers@prod-project.iam",
		},
		{
			name:        "test",
			suffix:      "-developers@test-project.iam.gserviceaccount.com",
			wantBinding: "serviceAccount:example-team-developers@test-project.iam.gserviceaccount.com",
			wantSQLUser: "example-team-developers@test-project.iam",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := saDevelopersEmail("example-team", tt.suffix); got != tt.wantBinding {
				t.Fatalf("IAM member = %q, want %q", got, tt.wantBinding)
			}
			if got := saDevelopersCloudSqlMember("example-team", tt.suffix); got != tt.wantSQLUser {
				t.Fatalf("Cloud SQL user = %q, want %q", got, tt.wantSQLUser)
			}
		})
	}
}

func TestToSchemaName(t *testing.T) {
	tests := []struct {
		name       string
		schema     string
		wantErr    bool
		wantSchema string
	}{
		{
			name:       "simple name",
			schema:     "team",
			wantErr:    false,
			wantSchema: "team_team",
		},
		{
			name:       "name with underscore",
			schema:     "team_understrek",
			wantErr:    false,
			wantSchema: "team_team_understrek",
		},
		{
			name:       "name with dash",
			schema:     "team-bindestrek",
			wantErr:    false,
			wantSchema: "team_team_bindestrek",
		},
		{
			name:    "empty name",
			schema:  "",
			wantErr: true,
		},
		{
			name:       "starts with number",
			schema:     "1team",
			wantErr:    false,
			wantSchema: "team_1team",
		},
		{
			name:    "contains dot",
			schema:  "team.name",
			wantErr: true,
		},
		{
			name:    "contains uppercase",
			schema:  "teAm",
			wantErr: true,
		},
		{
			name:       "long team name",
			schema:     "a" + strings.Repeat("b", 57),
			wantSchema: "team_a" + strings.Repeat("b", 57),
		},
		{
			name:    "too long",
			schema:  "a" + strings.Repeat("b", 58),
			wantErr: true,
		},
		{
			name:       "public schema is prefixed with team",
			schema:     "public",
			wantErr:    false,
			wantSchema: "team_public",
		},
		{
			name:       "information schema is prefixed with team",
			schema:     "information_schema",
			wantErr:    false,
			wantSchema: "team_information_schema",
		},
		{
			name:       "pg prefix is also prefixed with team",
			schema:     "pg_team",
			wantErr:    false,
			wantSchema: "team_pg_team",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := toSchemaName(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Fatalf("toSchemaName(%q) wantErr='%v', error='%v', ", tt.schema, tt.wantErr, err)
			}
			if schema != tt.wantSchema {
				t.Fatalf("toSchemaName(%q) wantSchema='%v', schema='%v', ", tt.schema, tt.wantSchema, schema)
			}
		})
	}
}
