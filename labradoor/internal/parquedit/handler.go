package parquedit

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

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
	// TODO
	team := chi.URLParam(req, "team")
	schema := pgx.Identifier{team}.Sanitize()

	result, err := c.db.Exec(req.Context(), "CREATE SCHEMA IF NOT EXISTS "+schema)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	slog.Info(result.String())

	w.WriteHeader(200)
}

func (c *Client) DisableForTeam(w http.ResponseWriter, req *http.Request) {
	// TODO
	team := chi.URLParam(req, "team")
	schema := pgx.Identifier{team}.Sanitize()

	result, err := c.db.Exec(req.Context(), "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(500)
		return
	}
	slog.Info(result.String())
	w.WriteHeader(200)
}

func (c *Client) HasEnabled(w http.ResponseWriter, req *http.Request) {
	// TODO
	team := chi.URLParam(req, "team")

	var schema_name string
	err := c.db.QueryRow(req.Context(), "SELECT schema_name FROM information_schema.schemata WHERE schema_name = $1", team).Scan(&schema_name)
	if err == pgx.ErrNoRows {
		w.WriteHeader(404)
		return
	}
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}
