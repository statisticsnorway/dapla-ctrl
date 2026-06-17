package googleresourcemanager

import (
	"context"
	"fmt"

	"google.golang.org/api/cloudresourcemanager/v1"
)

// Code forked from
// https://docs.cloud.google.com/iam/docs/write-policy-client-libraries

type GoogleCloudResourceManager struct {
	client *cloudresourcemanager.Service
}

func New(ctx context.Context) (*GoogleCloudResourceManager, error) {
	client, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return &GoogleCloudResourceManager{
		client: client,
	}, nil
}

// AddBinding adds the principal to one or more roles in the project's IAM policy.
func (c *GoogleCloudResourceManager) AddBindings(ctx context.Context, projectID, member string, roles ...string) error {
	if len(roles) == 0 {
		return nil
	}

	policy, err := c.getPolicy(ctx, projectID)
	if err != nil {
		return err
	}

	for _, role := range roles {
		// Find the policy binding for role. Only one binding can have the role.
		var binding *cloudresourcemanager.Binding
		for _, b := range policy.Bindings {
			if b.Role == role {
				binding = b
				break
			}
		}

		if binding != nil {
			// If the binding exists, adds the principal to the binding.
			binding.Members = append(binding.Members, member)
		} else {
			// If the binding does not exist, adds a new binding to the policy.
			binding = &cloudresourcemanager.Binding{
				Role:    role,
				Members: []string{member},
			}
			policy.Bindings = append(policy.Bindings, binding)
		}
	}

	return c.setPolicy(ctx, projectID, policy)
}

// RemoveMember removes the principal from the project's IAM policy
func (c *GoogleCloudResourceManager) RemoveMember(ctx context.Context, projectID, member string, roles ...string) error {
	if len(roles) == 0 {
		return nil
	}

	policy, err := c.getPolicy(ctx, projectID)
	if err != nil {
		return err
	}

	for _, role := range roles {
		// Find the policy binding for role. Only one binding can have the role.
		var binding *cloudresourcemanager.Binding
		var bindingIndex int
		for i, b := range policy.Bindings {
			if b.Role == role {
				binding = b
				bindingIndex = i
				break
			}
		}

		// Order doesn't matter for bindings or members, so to remove, move the last item
		// into the removed spot and shrink the slice.
		if len(binding.Members) == 1 {
			// If the principal is the only member in the binding, removes the binding
			last := len(policy.Bindings) - 1
			policy.Bindings[bindingIndex] = policy.Bindings[last]
			policy.Bindings = policy.Bindings[:last]
		} else {
			// If there is more than one member in the binding, removes the principal
			var memberIndex int
			for i, mm := range binding.Members {
				if mm == member {
					memberIndex = i
				}
			}
			last := len(policy.Bindings[bindingIndex].Members) - 1
			binding.Members[memberIndex] = binding.Members[last]
			binding.Members = binding.Members[:last]
		}
	}

	return c.setPolicy(ctx, projectID, policy)
}

// getPolicy gets the project's IAM policy
func (c *GoogleCloudResourceManager) getPolicy(ctx context.Context, projectID string) (*cloudresourcemanager.Policy, error) {
	request := new(cloudresourcemanager.GetIamPolicyRequest)
	policy, err := c.client.Projects.GetIamPolicy(projectID, request).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get iam policy on project: %w", err)
	}

	return policy, nil
}

// setPolicy sets the project's IAM policy
func (c *GoogleCloudResourceManager) setPolicy(ctx context.Context, projectID string, policy *cloudresourcemanager.Policy) error {
	request := new(cloudresourcemanager.SetIamPolicyRequest)
	request.Policy = policy
	policy, err := c.client.Projects.SetIamPolicy(projectID, request).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("set iam policy on project: %w", err)
	}
	return nil
}
