package googlesqladmin

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

type FakeSqlManager struct {
	ConnectionString string
}

func NewFake(connectionString string) *FakeSqlManager {
	return &FakeSqlManager{
		ConnectionString: connectionString,
	}
}

func (f *FakeSqlManager) AddUser(ctx context.Context, projectID, instance string, user *sqladmin.User) error {
	conn, err := pgx.Connect(ctx, f.ConnectionString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Take into account that google uses iam users (with dash), we'll mimic this in this fake
	username := googleSaUsernameToDBFriendly(user.Name)
	slog.Info("got username " + user.Name + " but replacing with " + username + "in fake sqladmin")

	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)", username).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err = conn.Exec(ctx, "CREATE USER \""+username+"\"")
	}
	return err
}

func (f *FakeSqlManager) RemoveUser(ctx context.Context, projectID, instance, user string) error {
	conn, err := pgx.Connect(ctx, f.ConnectionString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	username := googleSaUsernameToDBFriendly(user)
	slog.Info("got username " + user + " but replacing with " + username + " in fake sqladmin")
	// Take into account that google uses iam users (with dash), we'll mimic this in this fake
	_, err = conn.Exec(ctx, "DROP USER IF EXISTS \""+username+"\"")
	return err
}

func googleSaUsernameToDBFriendly(str string) string {
	return strings.TrimPrefix(str, "serviceAccount:")
}
