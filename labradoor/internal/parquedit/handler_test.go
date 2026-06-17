package parquedit

import (
	"strings"
	"testing"
)

func TestValidateSchemaName(t *testing.T) {
	tests := []struct {
		name    string
		schema  string
		wantErr bool
	}{
		{
			name:    "simple name",
			schema:  "team",
			wantErr: false,
		},
		{
			name:    "name with underscore",
			schema:  "team_understrek",
			wantErr: false,
		},
		{
			name:    "name with dash",
			schema:  "team-bindestrek",
			wantErr: true,
		},
		{
			name:    "empty name",
			schema:  "",
			wantErr: true,
		},
		{
			name:    "starts with number",
			schema:  "1team",
			wantErr: true,
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
			name:   "long team name",
			schema: "a" + strings.Repeat("b", 62),
		},
		{
			name:    "too long",
			schema:  "a" + strings.Repeat("b", 63),
			wantErr: true,
		},
		{
			name:    "public schema is reserved",
			schema:  "public",
			wantErr: true,
		},
		{
			name:    "information schema is reserved",
			schema:  "information_schema",
			wantErr: true,
		},
		{
			name:    "pg prefix is reserved",
			schema:  "pg_team",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchemaName(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateSchemaName(%q) error = %v, wantErr %v", tt.schema, err, tt.wantErr)
			}
		})
	}
}

