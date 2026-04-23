package entraidclient

import (
	"context"
	"fmt"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

type Client struct {
	client *msgraphsdkgo.GraphServiceClient
}

func New(tenantId, clientId string) (*Client, error) {
	creds, err := azidentity.NewClientAssertionCredential(tenantId, clientId, func(ctx context.Context) (string, error) {
		creds, err := idtoken.NewCredentials(&idtoken.Options{Audience: "api://AzureADTokenExchange"})
		if err != nil {
			return "", err
		}
		token, err := creds.Token(ctx)
		if err != nil {
			return "", err
		}
		return token.Value, nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("exchange for azure credentials: %w", err)
	}

	client, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(creds, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, fmt.Errorf("create graph service client: %w", err)
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) AddUserToGroup(ctx context.Context, groupId, userId string) error {
	requestBody := models.NewReferenceCreate()
	odataId := fmt.Sprintf("https://graph.microsoft.com/v1.0/directoryObjects/%s", userId)
	requestBody.SetOdataId(&odataId)
	if err := c.client.Groups().ByGroupId(groupId).Members().Ref().Post(ctx, requestBody, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) RemoveUserFromGroup(ctx context.Context, groupId, userId string) error {
	if err := c.client.Groups().ByGroupId(groupId).Members().ByDirectoryObjectId(userId).Ref().Delete(ctx, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateGroup(ctx context.Context, groupName string) (models.Groupable, error) {
	requestBody := models.NewGroup()
	requestBody.SetDisplayName(&groupName)
	requestBody.SetSecurityEnabled(new(true))
	requestBody.SetMailEnabled(new(false))
	requestBody.SetMailNickname(&groupName)
	requestBody.SetDescription(new("source:dapla-api"))

	group, err := c.client.Groups().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (c *Client) GetGroup(ctx context.Context, groupId string) (models.Groupable, error) {
	return c.client.Groups().ByGroupId(groupId).Get(ctx, nil)
}

func (c *Client) GetTransitiveMembers(ctx context.Context, groupId string) ([]models.Userable, error) {
	var users []models.Userable
	req, err := c.client.Groups().ByGroupId(groupId).TransitiveMembers().GraphUser().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get entra id group members: %w", err)
	}

	pageIterator, err := msgraphcore.NewPageIterator[models.Userable](req, c.client.GetAdapter(), models.CreateUserCollectionResponseFromDiscriminatorValue)
	if err != nil {
		return nil, fmt.Errorf("create entra id users pageiterator: %w", err)
	}

	if err := pageIterator.Iterate(ctx, func(user models.Userable) bool {
		users = append(users, user)
		return true
	}); err != nil {
		return nil, fmt.Errorf("list all users group members: %w", err)
	}
	return users, nil
}

func (c *Client) AssignAppRoleToGroup(ctx context.Context, groupId string, resourceId *uuid.UUID, appRoleId *uuid.UUID) error {
	gcpSyncAssignment := models.NewAppRoleAssignment()
	gcpSyncAssignment.SetAppRoleId(appRoleId)
	gcpSyncAssignment.SetResourceId(resourceId)
	groupUuid, err := uuid.Parse(groupId)
	if err != nil {
		return fmt.Errorf("parse group id: %w", err)
	}
	gcpSyncAssignment.SetPrincipalId(&groupUuid)
	if _, err := c.client.Groups().ByGroupId(groupId).AppRoleAssignments().Post(ctx, gcpSyncAssignment, nil); err != nil {
		return fmt.Errorf("create app role assignment: %w", err)
	}

	return nil
}

func (c *Client) GetAppRolesForGroup(ctx context.Context, groupId string) ([]models.AppRoleAssignmentable, error) {
	appRolesResponse, err := c.client.Groups().ByGroupId(groupId).AppRoleAssignments().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	pageIterator, _ := msgraphcore.NewPageIterator[models.AppRoleAssignmentable](appRolesResponse, c.client.GetAdapter(), models.CreateAppRoleAssignmentCollectionResponseFromDiscriminatorValue)

	var appRoleAssignments []models.AppRoleAssignmentable
	if err := pageIterator.Iterate(ctx, func(apa models.AppRoleAssignmentable) bool {
		appRoleAssignments = append(appRoleAssignments, apa)
		return true
	}); err != nil {
		return nil, err
	}

	return appRoleAssignments, nil
}
