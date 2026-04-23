package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/entraidclient"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers/entraid/group/master"
)

type Master struct {
	client *entraidclient.Client
}

func New(client *entraidclient.Client) *Master {
	return &Master{
		client: client,
	}
}

func (m *Master) Name() string {
	return "database"
}

func (m *Master) RemoveUsers(ctx context.Context, group master.Group, localOnlyUsers, remoteOnlyUsers []master.User, log logrus.FieldLogger) error {
	var errs []error
	for _, user := range remoteOnlyUsers {
		if err := m.client.RemoveUserFromGroup(ctx, group.ExternalId, user.ExternalId); err != nil {
			errs = append(errs, fmt.Errorf("remove user %q from group %q: %w", user.Email, group.Name, err))
		}
	}
	return errors.Join(errs...)
}

func (m *Master) AddUsers(ctx context.Context, group master.Group, localOnlyUsers, remoteOnlyUsers []master.User, log logrus.FieldLogger) error {
	var errs []error
	for _, user := range localOnlyUsers {
		if err := m.client.AddUserToGroup(ctx, group.ExternalId, user.ExternalId); err != nil {
			errs = append(errs, fmt.Errorf("add user %q to group %q: %w", user.Email, group.Name, err))
		}
	}
	return errors.Join(errs...)
}
