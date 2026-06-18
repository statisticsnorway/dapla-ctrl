package googlesqladmin

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

type FakeSqlManager struct {
	DatabaseUrl string
}

func NewFake(databaseUrl string) *FakeSqlManager {
	return &FakeSqlManager{
		DatabaseUrl: databaseUrl,
	}
}

func (f *FakeSqlManager) AddUser(ctx context.Context, projectID, instance string, user *sqladmin.User) error {
	conn, err := pgx.Connect(ctx, f.DatabaseUrl)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)", user.Name).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err = conn.Exec(ctx, "CREATE USER "+user.Name)
	}
	return err
}

func (f *FakeSqlManager) RemoveUser(ctx context.Context, projectID, instance, user string) error {
	conn, err := pgx.Connect(ctx, f.DatabaseUrl)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "DROP USER IF EXISTS "+user)
	return err
}
