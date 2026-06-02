package database

import (
	"context"
	"errors"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api-reconcilers/internal/reconcilers/entraid/group/master"
)

type Master struct {
	client *msgraphsdk.GraphServiceClient
}

func New(client *msgraphsdk.GraphServiceClient) *Master {
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
		if err := m.client.Groups().ByGroupId(group.ExternalId).Members().ByDirectoryObjectId(user.ExternalId).Ref().Delete(ctx, nil); err != nil {
			errs = append(errs, fmt.Errorf("remove user %q from group %q: %w", user.Email, group.Name, err))
		}
	}
	return errors.Join(errs...)
}

func (m *Master) AddUsers(ctx context.Context, group master.Group, localOnlyUsers, remoteOnlyUsers []master.User, log logrus.FieldLogger) error {
	var errs []error
	for _, user := range localOnlyUsers {
		requestBody := models.NewReferenceCreate()
		odataId := fmt.Sprintf("https://graph.microsoft.com/v1.0/directoryObjects/%s", user.ExternalId)
		requestBody.SetOdataId(&odataId)
		if err := m.client.Groups().ByGroupId(group.ExternalId).Members().Ref().Post(ctx, requestBody, nil); err != nil {
			errs = append(errs, fmt.Errorf("add user %q to group %q: %w", user.Email, group.Name, err))
		}
	}
	return errors.Join(errs...)
}
