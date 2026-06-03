package groupserviceaccounts

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
)

type GroupServiceAccounts interface {
	GetOrCreate(ctx context.Context, group, projectId string) (*iam.ServiceAccount, error)
	UpdateDescription(ctx context.Context, name string, description string, projectId string) error
}

type GoogleServiceAccounts struct {
	client *iam.Service
}

func NewGoogleServiceAccounts(ctx context.Context) (*GoogleServiceAccounts, error) {
	client, err := iam.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return &GoogleServiceAccounts{
		client: client,
	}, nil
}

func (g *GoogleServiceAccounts) GetOrCreate(ctx context.Context, group, projectId string) (*iam.ServiceAccount, error) {
	saName := fmt.Sprintf("projects/-/serviceAccounts/%s@%s.iam.gserviceaccount.com", group, projectId)
	sa, err := g.client.Projects.ServiceAccounts.Get(saName).Context(ctx).Do()

	// TODO: replace with
	if gErr, ok := errors.AsType[*googleapi.Error](err); ok && gErr.Code == http.StatusNotFound {
		return g.createServiceAccount(ctx, group, projectId)
	} else if err != nil {
		return nil, fmt.Errorf("unexpected error getting service account: %w", err)
	}
	return sa, nil
}

func (g *GoogleServiceAccounts) createServiceAccount(ctx context.Context, groupName, projectId string) (*iam.ServiceAccount, error) {
	req := iam.CreateServiceAccountRequest{
		AccountId: groupName,
		ServiceAccount: &iam.ServiceAccount{
			Description: saDescription,
		},
	}

	sa, err := g.client.Projects.ServiceAccounts.Create(fmt.Sprintf("projects/%s", projectId), &req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating service account: %w", err)
	}

	return sa, nil
}

func (g *GoogleServiceAccounts) UpdateDescription(ctx context.Context, group, description, projectId string) error {
	req := iam.PatchServiceAccountRequest{
		ServiceAccount: &iam.ServiceAccount{
			Description: saDescription,
		},
		UpdateMask: "description",
	}

	saName := fmt.Sprintf("projects/-/serviceAccounts/%s@%s.iam.gserviceaccount.com", group, projectId)
	if _, err := g.client.Projects.ServiceAccounts.Patch(saName, &req).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unexpected error patching sa description: %w", err)
	}
	return nil
}
