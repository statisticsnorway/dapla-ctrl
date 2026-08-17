package serviceaccounts

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
)

type Client struct {
	client *iam.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	client, err := iam.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (g *Client) GetOrCreate(ctx context.Context, name, description, projectId string) (*iam.ServiceAccount, error) {
	saName := fmt.Sprintf("projects/%s/serviceAccounts/%s@%s.iam.gserviceaccount.com", projectId, name, projectId)
	sa, err := g.client.Projects.ServiceAccounts.Get(saName).Context(ctx).Do()

	if gErr, ok := errors.AsType[*googleapi.Error](err); ok && gErr.Code == http.StatusNotFound {
		return g.createServiceAccount(ctx, name, description, projectId)
	} else if err != nil {
		return nil, fmt.Errorf("unexpected error getting service account: %w", err)
	}

	return sa, nil
}

func (g *Client) createServiceAccount(ctx context.Context, name, description, projectId string) (*iam.ServiceAccount, error) {
	req := iam.CreateServiceAccountRequest{
		AccountId: name,
		ServiceAccount: &iam.ServiceAccount{
			Description: description,
		},
	}

	sa, err := g.client.Projects.ServiceAccounts.Create(fmt.Sprintf("projects/%s", projectId), &req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating service account: %w", err)
	}

	return sa, nil
}

func (g *Client) UpdateDescription(ctx context.Context, saName, description string) error {
	req := iam.PatchServiceAccountRequest{
		ServiceAccount: &iam.ServiceAccount{
			Description: description,
		},
		UpdateMask: "description",
	}

	if _, err := g.client.Projects.ServiceAccounts.Patch(saName, &req).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unexpected error patching sa description: %w", err)
	}
	return nil
}

func (g *Client) EnsureRoleBindingFunc(ctx context.Context, saName, role string, modifyBinding func(b *iam.Binding) (modified bool)) error {
	policy, err := g.client.Projects.ServiceAccounts.GetIamPolicy(saName).Context(ctx).Do()
	if err != nil {
		return err
	}

	bindingIndex := slices.IndexFunc(policy.Bindings, func(b *iam.Binding) bool {
		return b.Role == role && b.Condition == nil
	})
	if bindingIndex == -1 {
		policy.Bindings = append(policy.Bindings, &iam.Binding{
			Role: role,
		})
		bindingIndex = len(policy.Bindings) - 1
	}

	binding := policy.Bindings[bindingIndex]

	if modified := modifyBinding(binding); !modified {
		return nil
	}

	_, err = g.client.Projects.ServiceAccounts.SetIamPolicy(saName, &iam.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	return err
}
