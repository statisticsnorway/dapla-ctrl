package pam

import (
	"context"
	"fmt"
	"strings"
	"time"

	privilegedaccessmanager "cloud.google.com/go/privilegedaccessmanager/apiv1"
	"cloud.google.com/go/privilegedaccessmanager/apiv1/privilegedaccessmanagerpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const pamEntitlementID = "ssb-bucket-read"

type PrivilegedAccessManager interface {
	EnsureEntitlement(ctx context.Context, folderID, role, conditionExpression string, principals []string, maxGrantDuration time.Duration) error
}

type GooglePrivilegedAccessManager struct {
	client *privilegedaccessmanager.Client
}

func (g *GooglePrivilegedAccessManager) EnsureEntitlement(ctx context.Context, folderID, role, conditionExpression string, principals []string, maxGrantDuration time.Duration) error {
	if strings.TrimSpace(conditionExpression) == "" {
		return fmt.Errorf("PAM IAM condition is required")
	}
	if g.client == nil {
		return fmt.Errorf("PAM client is not configured")
	}

	parent := fmt.Sprintf("folders/%s/locations/global", folderID)
	desired := &privilegedaccessmanagerpb.Entitlement{
		Name: parent + "/entitlements/" + pamEntitlementID,
		EligibleUsers: []*privilegedaccessmanagerpb.AccessControlEntry{
			{Principals: principals},
		},
		PrivilegedAccess: &privilegedaccessmanagerpb.PrivilegedAccess{
			AccessType: &privilegedaccessmanagerpb.PrivilegedAccess_GcpIamAccess_{
				GcpIamAccess: &privilegedaccessmanagerpb.PrivilegedAccess_GcpIamAccess{
					ResourceType: "cloudresourcemanager.googleapis.com/Folder",
					Resource:     "//cloudresourcemanager.googleapis.com/folders/" + folderID,
					RoleBindings: []*privilegedaccessmanagerpb.PrivilegedAccess_GcpIamAccess_RoleBinding{
						{Role: role, ConditionExpression: conditionExpression},
					},
				},
			},
		},
		MaxRequestDuration: durationpb.New(maxGrantDuration),
		RequesterJustificationConfig: &privilegedaccessmanagerpb.Entitlement_RequesterJustificationConfig{
			JustificationType: &privilegedaccessmanagerpb.Entitlement_RequesterJustificationConfig_NotMandatory_{
				NotMandatory: &privilegedaccessmanagerpb.Entitlement_RequesterJustificationConfig_NotMandatory{},
			},
		},
	}

	existing, err := g.client.GetEntitlement(ctx, &privilegedaccessmanagerpb.GetEntitlementRequest{Name: desired.Name})
	if status.Code(err) == codes.NotFound {
		op, err := g.client.CreateEntitlement(ctx, &privilegedaccessmanagerpb.CreateEntitlementRequest{
			Parent:        parent,
			EntitlementId: pamEntitlementID,
			Entitlement:   desired,
		})
		if err != nil {
			return fmt.Errorf("create entitlement %q: %w", desired.Name, err)
		}
		if _, err := op.Wait(ctx); err != nil {
			return fmt.Errorf("wait for entitlement creation %q: %w", desired.Name, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("get entitlement %q: %w", desired.Name, err)
	}

	if proto.Equal(&privilegedaccessmanagerpb.Entitlement{
		EligibleUsers:                existing.EligibleUsers,
		PrivilegedAccess:             existing.PrivilegedAccess,
		MaxRequestDuration:           existing.MaxRequestDuration,
		ApprovalWorkflow:             existing.ApprovalWorkflow,
		RequesterJustificationConfig: existing.RequesterJustificationConfig,
	}, &privilegedaccessmanagerpb.Entitlement{
		EligibleUsers:                desired.EligibleUsers,
		PrivilegedAccess:             desired.PrivilegedAccess,
		MaxRequestDuration:           desired.MaxRequestDuration,
		ApprovalWorkflow:             desired.ApprovalWorkflow,
		RequesterJustificationConfig: desired.RequesterJustificationConfig,
	}) {
		return nil
	}

	desired.Etag = existing.Etag
	op, err := g.client.UpdateEntitlement(ctx, &privilegedaccessmanagerpb.UpdateEntitlementRequest{
		Entitlement: desired,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
			"eligible_users", "privileged_access", "max_request_duration",
			"approval_workflow", "requester_justification_config",
		}},
	})
	if err != nil {
		return fmt.Errorf("update entitlement %q: %w", desired.Name, err)
	}
	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait for entitlement update %q: %w", desired.Name, err)
	}
	return nil
}
