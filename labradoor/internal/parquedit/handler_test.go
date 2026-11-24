package parquedit

import (
	"strings"
	"testing"

	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/config"
)

func TestDataPathTemplate(t *testing.T) {
	for _, tt := range []struct {
		name     string
		template string
		wantErr  bool
	}{
		{name: "production", template: "gs://ssb-{team}-data-produkt-prod/.parquedit_data/"},
		{name: "test", template: "gs://ssb-{team}-data-produkt-test/.parquedit_data/"},
		{name: "empty", wantErr: true},
		{name: "missing team", template: "gs://bucket/.parquedit_data/", wantErr: true},
		{name: "unknown placeholder", template: "gs://ssb-{team}-{env}/.parquedit_data/", wantErr: true},
		{name: "wrong scheme", template: "https://ssb-{team}/.parquedit_data/", wantErr: true},
		{name: "missing bucket", template: "gs:///{team}/", wantErr: true},
		{name: "missing trailing slash", template: "gs://ssb-{team}/.parquedit_data", wantErr: true},
		{name: "credentials", template: "gs://user@ssb-{team}/.parquedit_data/", wantErr: true},
		{name: "port", template: "gs://ssb-{team}:443/.parquedit_data/", wantErr: true},
		{name: "query", template: "gs://ssb-{team}/.parquedit_data/?other=bucket", wantErr: true},
		{name: "fragment", template: "gs://ssb-{team}/.parquedit_data/#other", wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateDataPathTemplate(tt.template); (err != nil) != tt.wantErr {
				t.Fatalf("validateDataPathTemplate() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func TestInvalidDataPathFailsBeforeConnecting(t *testing.T) {
	_, err := New(t.Context(), config.ParqueditConfig{
		DatabaseUrl:      "invalid database URL",
		DataPathTemplate: "gs://bucket/missing-team/",
	}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "PARQUEDIT_STORAGE_PATH") {
		t.Fatalf("expected template validation before database connection, got %v", err)
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
