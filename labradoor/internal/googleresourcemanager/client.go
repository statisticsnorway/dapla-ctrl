package googleresourcemanager

import (
	"context"
	"fmt"
	"slices"

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

// AddBindings adds the principal to one or more roles in the project's IAM policy.
func (c *GoogleCloudResourceManager) AddBindings(ctx context.Context, projectID, member string, roles ...string) error {
	if len(roles) == 0 {
		return nil
	}

	policy, err := c.getPolicy(ctx, projectID)
	if err != nil {
		return err
	}

	if addBindingsToPolicy(policy, member, roles...) {
		return c.setPolicy(ctx, projectID, policy)
	}

	return nil
}

func addBindingsToPolicy(policy *cloudresourcemanager.Policy, member string, roles ...string) bool {
	changed := false

	for _, role := range roles {
		// Find the policy binding for role. Only one binding can have the role.
		var binding *cloudresourcemanager.Binding
		for _, b := range policy.Bindings {
			if b.Role == role {
				binding = b
				break
			}
		}

		if binding == nil {
			// create binding and add member
			policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
				Role:    role,
				Members: []string{member},
			})
			changed = true
			continue
		}

		memberExists := slices.Contains(binding.Members, member)
		if memberExists {
			continue
		}

		binding.Members = append(binding.Members, member)
		changed = true
	}

	return changed
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

	if removeMemberFromPolicy(policy, member, roles...) {
		return c.setPolicy(ctx, projectID, policy)
	}

	return nil
}

func removeMemberFromPolicy(policy *cloudresourcemanager.Policy, member string, roles ...string) bool {
	changed := false

	for _, role := range roles {
		// Find the policy binding for role. Only one binding can have the role.
		bindingIndex := -1
		for i, b := range policy.Bindings {
			if b.Role == role {
				bindingIndex = i
				break
			}
		}
		if bindingIndex == -1 {
			continue
		}

		// Loop over and readd all members without the one we want to remove
		binding := policy.Bindings[bindingIndex]
		members := binding.Members[:0]
		removed := false
		for _, mm := range binding.Members {
			if mm == member {
				removed = true
				continue
			}
			members = append(members, mm)
		}
		if !removed {
			continue
		}

		changed = true
		if len(members) == 0 {
			// Remove the binding if there are no members assigned to it
			last := len(policy.Bindings) - 1
			policy.Bindings[bindingIndex] = policy.Bindings[last]
			policy.Bindings = policy.Bindings[:last]
		} else {
			binding.Members = members
		}
	}

	return changed
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
