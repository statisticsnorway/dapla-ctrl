package parquedit

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ParqueditConfig struct {
	DatabaseUrl string `env:"PARQUEDIT_DATABASE_URL,required"`
}

type Client struct {
	db *pgxpool.Pool
}

func (c *Client) Close() {
	c.db.Close()
}

func New(ctx context.Context, config ParqueditConfig) (*Client, error) {
	pool, err := pgxpool.New(context.Background(), config.DatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to connect to DB: %w", err)
	}

	parquedit := &Client{
		db: pool,
	}

	return parquedit, nil
}

func (c *Client) EnableForTeam(w http.ResponseWriter, req *http.Request) {
	// TODO, should we set AUTHORIZATION on the schema
	team := teamNameFromRequest(req)
	err := validateSchemaName(team)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	schema := pgx.Identifier{team}.Sanitize()

	result, err := c.db.Exec(req.Context(), "CREATE SCHEMA IF NOT EXISTS "+schema)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Info("enabled parquedit for team", "team", team, "result", result.String())

	w.WriteHeader(http.StatusOK)
}

func teamNameFromRequest(req *http.Request) string {
	teamNameWithPotentialDash := strings.ToLower(chi.URLParam(req, "team"))
	return strings.ReplaceAll(teamNameWithPotentialDash, "-", "_")
}

func (c *Client) DisableForTeam(w http.ResponseWriter, req *http.Request) {
	// TODO
	team := teamNameFromRequest(req)
	err := validateSchemaName(team)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	schema := pgx.Identifier{team}.Sanitize()

	result, err := c.db.Exec(req.Context(), "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Info("disabled parquedit for team", "team", team, "result", result.String())
	w.WriteHeader(http.StatusOK)
}

func (c *Client) HasEnabled(w http.ResponseWriter, req *http.Request) {
	// TODO
	team := teamNameFromRequest(req)
	if err := validateSchemaName(team); err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var exists bool
	err := c.db.QueryRow(req.Context(), `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.schemata
			WHERE schema_name = $1
		)`, team).Scan(&exists)

	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if exists {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func validateSchemaName(schema string) error {
	// https://www.postgresql.org/docs/18/sql-syntax-lexical.html#SQL-SYNTAX-IDENTIFIERS
	validSchemaName, _ := regexp.MatchString("^[a-z][a-z0-9_]{0,62}$", schema)
	if !validSchemaName {
		return fmt.Errorf("schema name %q is invalid", schema)
	}

	if strings.HasPrefix(schema, "pg_") || strings.EqualFold(schema, "public") || strings.EqualFold(schema, "information_schema") {
		return fmt.Errorf("schema name %q is reserved", schema)
	}
	return nil
}
