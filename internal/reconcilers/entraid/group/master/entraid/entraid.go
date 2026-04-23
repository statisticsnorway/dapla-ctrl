package entraid

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers/entraid/group/master"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/protoapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Master struct {
	client *apiclient.APIClient
}

func New(client *apiclient.APIClient) *Master {
	return &Master{
		client: client,
	}
}

func (m *Master) Name() string {
	return "entraid"
}

func (m *Master) RemoveUsers(ctx context.Context, group master.Group, localOnlyUsers, remoteOnlyUsers []master.User, log logrus.FieldLogger) error {
	var errs []error
	for _, user := range localOnlyUsers {
		_, err := m.client.Groups().RemoveMember(ctx, &protoapi.RemoveMemberRequest{
			Groupname:      group.Name,
			UserExternalId: user.ExternalId,
		})
		if status.Code(err) == codes.NotFound {
			log.Warn("user not found in database", "user", user.Email)
		} else if err != nil {
			errs = append(errs, fmt.Errorf("remove user %q from group %q: %w", user.Email, group.Name, err))
		}
	}
	return errors.Join(errs...)
}

func (m *Master) AddUsers(ctx context.Context, group master.Group, localOnlyUsers, remoteOnlyUsers []master.User, log logrus.FieldLogger) error {
	var errs []error
	for _, user := range remoteOnlyUsers {
		_, err := m.client.Groups().AddMember(ctx, &protoapi.AddMemberRequest{
			Groupname:      group.Name,
			UserExternalId: user.ExternalId,
		})
		if status.Code(err) == codes.NotFound {
			log.Warn("user not found in database", "user", user.Email)
		} else if err != nil {
			errs = append(errs, fmt.Errorf("add user %q to group %q: %w", user.Email, group.Name, err))
		}
	}
	return errors.Join(errs...)
}
