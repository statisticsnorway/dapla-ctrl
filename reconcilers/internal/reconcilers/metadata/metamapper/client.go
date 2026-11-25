package metamapper

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"cloud.google.com/go/iam/apiv1/iampb"
	privilegedaccessmanager "cloud.google.com/go/privilegedaccessmanager/apiv1"
	"cloud.google.com/go/privilegedaccessmanager/apiv1/privilegedaccessmanagerpb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var entitlementIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{3,62}$`)

func validateEntitlementID(id string) error {
	if !entitlementIDPattern.MatchString(id) {
		return fmt.Errorf("PAM entitlement ID must be 4-63 lowercase letters, digits or hyphens, starting with a letter")
	}
	return nil
}

const pamViewerRole = "roles/privilegedaccessmanager.viewer"

type PrivilegedAccessManager interface {
	EnsureEntitlement(ctx context.Context, folderID, entitlementID, role, conditionExpression string, principals []string, maxGrantDuration time.Duration) error
	EnsurePermissions(ctx context.Context, folderID string, principals []string) error
}

type GooglePrivilegedAccessManager struct {
	client  *privilegedaccessmanager.Client
	folders *resourcemanager.FoldersClient
}

// In order for principals to be able to create grants against the entitlements created in this reconciler,
// they need access to view entitlements and grants.
func (g *GooglePrivilegedAccessManager) EnsurePermissions(ctx context.Context, folderID string, principals []string) error {
	if g.folders == nil {
		return fmt.Errorf("folders client is not configured")
	}
	return google.EnsureRolesBindingFunc(ctx, g.folders, "folders/"+folderID,
		[]string{pamViewerRole},
		func(b *iampb.Binding) (modified bool) {
			for _, principal := range principals {
				if slices.Contains(b.Members, principal) {
					continue
				}
				b.Members = append(b.Members, principal)
				modified = true
			}
			return modified
		})
}

func (g *GooglePrivilegedAccessManager) EnsureEntitlement(ctx context.Context, folderID, entitlementID, role, conditionExpression string, principals []string, maxGrantDuration time.Duration) error {
	if err := validateEntitlementID(entitlementID); err != nil {
		return err
	}
	if strings.TrimSpace(conditionExpression) == "" {
		return fmt.Errorf("PAM IAM condition is required")
	}
	if g.client == nil {
		return fmt.Errorf("PAM client is not configured")
	}

	parent := fmt.Sprintf("folders/%s/locations/global", folderID)
	desired := &privilegedaccessmanagerpb.Entitlement{
		Name: parent + "/entitlements/" + entitlementID,
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
			EntitlementId: entitlementID,
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
