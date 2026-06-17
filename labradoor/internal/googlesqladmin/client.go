package googlesqladmin

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/api/googleapi"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

type GoogleSqlAdmin struct {
	client *sqladmin.Service
}

func New(client *sqladmin.Service) *GoogleSqlAdmin {
	return &GoogleSqlAdmin{
		client: client,
	}
}

func (c *GoogleSqlAdmin) AddUser(ctx context.Context, projectID, instance string, user *sqladmin.User) error {
	_, err := c.client.Users.Insert(projectID, instance, user).Context(ctx).Do()
	if googleapi.IsNotModified(err) {
		return nil
	}
	var gerr *googleapi.Error
	if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
		return nil
	}
	return err
}

func (c *GoogleSqlAdmin) RemoveUser(ctx context.Context, projectID, instance, user string) error {
	_, err := c.client.Users.Delete(projectID, instance).Name(user).Context(ctx).Do()
	if googleapi.IsNotModified(err) {
		return nil
	}
	var gerr *googleapi.Error
	if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
		return nil
	}
	return err
}
