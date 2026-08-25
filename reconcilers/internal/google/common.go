package google

import (
	"context"
	"slices"

	"cloud.google.com/go/iam/apiv1/iampb"
	"github.com/googleapis/gax-go/v2"
)

type IamPolicyClient interface {
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
	SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
}

func EnsureRoleBindingFunc(ctx context.Context, g IamPolicyClient, resourceName, role string, modifyBinding func(b *iampb.Binding) (modified bool)) error {
	return EnsureRolesBindingFunc(ctx, g, resourceName, []string{role}, modifyBinding)
}

func EnsureRolesBindingFunc(ctx context.Context, g IamPolicyClient, resourceName string, roles []string, modifyBinding func(b *iampb.Binding) (modified bool)) error {
	policy, err := g.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{Resource: resourceName})
	if err != nil {
		return err
	}

	modified := false
	for _, role := range roles {
		bindingIndex := slices.IndexFunc(policy.Bindings, func(b *iampb.Binding) bool {
			return b.Role == role && b.Condition == nil
		})
		if bindingIndex == -1 {
			policy.Bindings = append(policy.Bindings, &iampb.Binding{
				Role: role,
			})
			bindingIndex = len(policy.Bindings) - 1
		}

		binding := policy.Bindings[bindingIndex]

		modified = modifyBinding(binding) || modified
	}

	if !modified {
		return nil
	}

	_, err = g.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: resourceName,
		Policy:   policy,
	})
	return err
}
