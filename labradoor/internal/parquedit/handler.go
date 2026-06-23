package parquedit

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/config"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

const (
	cloudSQLClientRole       = "roles/cloudsql.client"
	cloudSQLInstanceUserRole = "roles/cloudsql.instanceUser"
)

type Client struct {
	db                 *pgxpool.Pool
	crm                CloudResourceManager
	sqlManager         SqlManager
	cloudSqlProject    string
	cloudSqlInstance   string
	cloudSqlUserSuffix string
}

func (c *Client) Close() {
	c.db.Close()
}

type CloudResourceManager interface {
	AddBindings(ctx context.Context, projectID, member string, roles ...string) error
	RemoveMember(ctx context.Context, projectID, member string, roles ...string) error
}

type SqlManager interface {
	AddUser(ctx context.Context, projectID, instance string, user *sqladmin.User) error
	RemoveUser(ctx context.Context, projectID, instance, user string) error
}

func New(ctx context.Context, config config.ParqueditConfig, crm CloudResourceManager, sqlClient SqlManager) (*Client, error) {
	pool, err := pgxpool.New(ctx, config.DatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to connect to DB: %w", err)
	}

	parquedit := &Client{
		db:                 pool,
		crm:                crm,
		sqlManager:         sqlClient,
		cloudSqlProject:    config.CloudSQLProject,
		cloudSqlInstance:   config.CloudSQLInstance,
		cloudSqlUserSuffix: config.CloudSqlUserSuffix,
	}

	return parquedit, nil
}

func (c *Client) EnableForTeam(w http.ResponseWriter, req *http.Request) {
	team := strings.ToLower(chi.URLParam(req, "team"))
	schema, err := toSchemaName(team)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log := config.LoggerFromCtx(req.Context())

	err = c.crm.AddBindings(req.Context(), c.cloudSqlProject, saDevelopersEmail(team, c.cloudSqlUserSuffix), cloudSQLClientRole, cloudSQLInstanceUserRole)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("bindings for cloudsql on project created")

	saMember := saDevelopersCloudSqlMember(team, c.cloudSqlUserSuffix)
	log.Info("Add user to sql instance", "user", saMember)
	err = c.sqlManager.AddUser(req.Context(), c.cloudSqlProject, c.cloudSqlInstance, &sqladmin.User{
		Name: saMember,
		Type: "CLOUD_IAM_SERVICE_ACCOUNT",
	})
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("Added user to sql instance")

	log.Info("create schema", "schema", schema)
	result, err := c.db.Exec(req.Context(), "CREATE SCHEMA IF NOT EXISTS "+schema)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("created schema", "schema", schema, "result", result.String())

	log.Info("grant on schema", "user", saMember)
	
	result, err = c.db.Exec(req.Context(), "GRANT CREATE, USAGE ON SCHEMA "+schema+" TO "+saMember)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("granted privileges on schema", "result", result.String())

	log.Info("grant on all tables in schema")
	result, err = c.db.Exec(req.Context(), "GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA "+schema+" TO "+saMember)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("granted privileges on all tables in schema", "result", result.String())

	log.Info("enabled parquedit for team")
	w.WriteHeader(http.StatusOK)
}

func (c *Client) DisableForTeam(w http.ResponseWriter, req *http.Request) {
	team := strings.ToLower(chi.URLParam(req, "team"))
	schema, err := toSchemaName(team)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log := config.LoggerFromCtx(req.Context())

	log.Info("drop schema for team", "schema", schema)
	result, err := c.db.Exec(req.Context(), "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("droped schema for team", "result", result.String())

	log.Info("remove bindings for cloudsql on project")
	err = c.crm.RemoveMember(req.Context(), c.cloudSqlProject, saDevelopersEmail(team, c.cloudSqlUserSuffix), cloudSQLClientRole, cloudSQLInstanceUserRole)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("removed bindings for cloudsql on project")

	log.Info("remove user from sql instance")
	saMember := saDevelopersCloudSqlMember(team, c.cloudSqlUserSuffix)
	err = c.sqlManager.RemoveUser(req.Context(), c.cloudSqlProject, c.cloudSqlInstance, saMember)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("removed user from sql instance")

	log.Info("disabled parquedit for team")
	w.WriteHeader(http.StatusOK)
}

func (c *Client) HasEnabled(w http.ResponseWriter, req *http.Request) {
	team := strings.ToLower(chi.URLParam(req, "team"))
	schema, err := toSchemaName(team)
	if err != nil {
		httplog.SetError(req.Context(), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var exists bool
	err = c.db.QueryRow(req.Context(), `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.schemata
			WHERE schema_name = $1
		)`, schema).Scan(&exists)
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

// schema names are prefixed with _team to be clear in the DB and avoid potential collisions
func toSchemaName(team string) (string, error) {
	if len(team) == 0 {
		return "", fmt.Errorf("team must not be empty")
	}

	schema := "team_" + strings.ReplaceAll(team, "-", "_")
	// https://www.postgresql.org/docs/18/sql-syntax-lexical.html#SQL-SYNTAX-IDENTIFIERS
	validSchemaName, _ := regexp.MatchString("^[a-z][a-z0-9_]{0,62}$", schema)
	if !validSchemaName {
		return "", fmt.Errorf("schema name %q is invalid", schema)
	}

	if strings.HasPrefix(schema, "pg_") || strings.EqualFold(schema, "public") || strings.EqualFold(schema, "information_schema") {
		return "", fmt.Errorf("schema name %q is reserved", schema)
	}
	return schema, nil
}

// the sa email to be used for binding in gcloud
func saDevelopersEmail(team, teamSuffix string) string {
	return "serviceAccount:" + team + teamSuffix
}

// the sa member name to be used for binding in cloudsql when using cloud iam service account
func saDevelopersCloudSqlMember(team, teamSuffix string) string {
	return strings.TrimSuffix(team+teamSuffix, ".gserviceaccount.com")
}
