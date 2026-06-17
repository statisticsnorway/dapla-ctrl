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
	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/googleresourcemanager"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

const (
	cloudSQLClientRole       = "roles/cloudsql.client"
	cloudSQLInstanceUserRole = "roles/cloudsql.instanceUser"
)

type ParqueditConfig struct {
	DatabaseUrl         string `env:"PARQUEDIT_DATABASE_URL,required"`
	CloudSQLProject     string `env:"PARQUEDIT_CLOUDSQL_PROJECT"`
	CloudSQLInstance    string `env:"PARQUEDIT_CLOUDSQL_INSTANCE"`
	DaplaGroupSaProject string `env:"PARQUEDIT_DAPLA_GROUP_SA_PROJECT"`
}

type Client struct {
	db                  *pgxpool.Pool
	gcrm                *googleresourcemanager.GoogleCloudResourceManager
	sqladmin            *sqladmin.Service
	cloudSqlProject     string
	cloudSqlInstance    string
	daplaGroupSaProject string
}

type enableForTeamRequest struct {
	Project string `json:"project"`
	User    string `json:"user"`
}

func (c *Client) Close() {
	c.db.Close()
}

func New(ctx context.Context, config ParqueditConfig) (*Client, error) {
	pool, err := pgxpool.New(ctx, config.DatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to connect to DB: %w", err)
	}

	crm, err := googleresourcemanager.New(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to create google cloud resource manager client: %w", err)
	}

	sqladminService, err := sqladmin.NewService(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to create google sqladmin client: %w", err)
	}

	parquedit := &Client{
		db:                  pool,
		gcrm:                crm,
		sqladmin:            sqladminService,
		cloudSqlProject:     config.CloudSQLProject,
		cloudSqlInstance:    config.CloudSQLInstance,
		daplaGroupSaProject: config.DaplaGroupSaProject,
	}

	return parquedit, nil
}

func (c *Client) EnableForTeam(w http.ResponseWriter, req *http.Request) {
	team := teamNameWithPrefix(req)
	err := validateSchemaName(team)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.gcrm.AddBindings(req.Context(), c.cloudSqlProject, saDevelopersEmail(team, c.daplaGroupSaProject), cloudSQLClientRole, cloudSQLInstanceUserRole)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Info("bindings for cloudsql on project created")

	saMember := saDevelopersCloudSqlMember(team, c.daplaGroupSaProject)
	op, err := c.sqladmin.Users.Insert(c.cloudSqlProject, c.cloudSqlInstance, &sqladmin.User{
		Name: saMember,
		Type: "CLOUD_IAM_SERVICE_ACCOUNT",
	}).Context(req.Context()).Do()
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Info("user inserted in sql instance", "identifier", op.Name)

	schema := pgx.Identifier{team}.Sanitize()

	result, err := c.db.Exec(req.Context(), "CREATE SCHEMA IF NOT EXISTS "+schema)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Info("created schema", "result", result.String())

	// TODO: double check with ffunk if grant all is correct
	result, err = c.db.Exec(req.Context(), "GRANT ALL ON ALL TABLES IN SCHEMA "+schema+" TO "+saMember)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Info("granted privileges on schema", "result", result.String())

	slog.Info("enabled parquedit for team")
	w.WriteHeader(http.StatusOK)
}

func (c *Client) DisableForTeam(w http.ResponseWriter, req *http.Request) {
	// TODO
	team := teamNameWithPrefix(req)
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
	slog.Info("disabled parquedit for team", "result", result.String())
	w.WriteHeader(http.StatusOK)
}

func (c *Client) HasEnabled(w http.ResponseWriter, req *http.Request) {
	// TODO
	team := teamNameWithPrefix(req)
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

func teamNameWithPrefix(req *http.Request) string {
	teamNameWithPotentialDash := strings.ToLower(chi.URLParam(req, "team"))
	return "team_" + strings.ReplaceAll(teamNameWithPotentialDash, "-", "_")
}

// the sa email to be used for binding in gcloud
func saDevelopersEmail(team, project string) string {
	return team + "-developers@" + project + ".iam.gserviceaccount.com"
}

// the sa member name to be used for binding in cloudsql when using cloud iam service account
func saDevelopersCloudSqlMember(team, project string) string {
	return team + "-developers@" + project + ".iam"
}
