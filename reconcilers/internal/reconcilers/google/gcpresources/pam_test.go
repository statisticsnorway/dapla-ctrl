package gcpresources

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	privilegedaccessmanager "cloud.google.com/go/privilegedaccessmanager/apiv1"
	"cloud.google.com/go/privilegedaccessmanager/apiv1/privilegedaccessmanagerpb"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakePAMServer struct {
	privilegedaccessmanagerpb.UnimplementedPrivilegedAccessManagerServer
	entitlement    *privilegedaccessmanagerpb.Entitlement
	createRequests []*privilegedaccessmanagerpb.CreateEntitlementRequest
	updateRequests []*privilegedaccessmanagerpb.UpdateEntitlementRequest
	getErr         error
	createErr      error
	updateErr      error
	operationErr   *rpcstatus.Status
}

func (s *fakePAMServer) GetEntitlement(_ context.Context, req *privilegedaccessmanagerpb.GetEntitlementRequest) (*privilegedaccessmanagerpb.Entitlement, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.entitlement == nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	if req.Name != s.entitlement.Name {
		return nil, status.Error(codes.InvalidArgument, "unexpected entitlement name")
	}
	return proto.Clone(s.entitlement).(*privilegedaccessmanagerpb.Entitlement), nil
}

func (s *fakePAMServer) CreateEntitlement(_ context.Context, req *privilegedaccessmanagerpb.CreateEntitlementRequest) (*longrunningpb.Operation, error) {
	s.createRequests = append(s.createRequests, req)
	if s.createErr != nil {
		return nil, s.createErr
	}
	s.entitlement = proto.Clone(req.Entitlement).(*privilegedaccessmanagerpb.Entitlement)
	s.entitlement.Etag = "current-version"
	s.entitlement.CreateTime = timestamppb.Now()
	s.entitlement.State = privilegedaccessmanagerpb.Entitlement_AVAILABLE
	return s.operation()
}

func (s *fakePAMServer) UpdateEntitlement(_ context.Context, req *privilegedaccessmanagerpb.UpdateEntitlementRequest) (*longrunningpb.Operation, error) {
	s.updateRequests = append(s.updateRequests, req)
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	if req.Entitlement.Etag != s.entitlement.Etag {
		return nil, status.Error(codes.Aborted, "etag mismatch")
	}
	s.entitlement = proto.Clone(req.Entitlement).(*privilegedaccessmanagerpb.Entitlement)
	return s.operation()
}

func (s *fakePAMServer) operation() (*longrunningpb.Operation, error) {
	op := &longrunningpb.Operation{Name: "folders/123/locations/global/operations/test", Done: true}
	if s.operationErr != nil {
		op.Result = &longrunningpb.Operation_Error{Error: s.operationErr}
		return op, nil
	}
	response, err := anypb.New(s.entitlement)
	if err != nil {
		return nil, err
	}
	op.Result = &longrunningpb.Operation_Response{Response: response}
	return op, nil
}

func newFakePAMClient(t *testing.T, server *fakePAMServer) *GooglePrivilegedAccessManager {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	privilegedaccessmanagerpb.RegisterPrivilegedAccessManagerServer(grpcServer, server)
	go func() { _ = grpcServer.Serve(listener) }()
	t.Cleanup(grpcServer.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client, err := privilegedaccessmanager.NewClient(t.Context(), option.WithGRPCConn(conn))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return &GooglePrivilegedAccessManager{client: client}
}

func TestEnsureEntitlement_CreatesAndReconcilesPolicy(t *testing.T) {
	server := &fakePAMServer{}
	client := newFakePAMClient(t, server)
	principals := []string{"serviceAccount:reader@test.iam.gserviceaccount.com", "user:reader@example.com", "group:readers@example.com"}
	role := "roles/storage.objectViewer"
	condition := `resource.name.startsWith("projects/_/buckets/ssb-my-team-data-produkt-dev/")`
	if err := client.EnsureEntitlement(t.Context(), "123", role, condition, principals, time.Hour); err != nil {
		t.Fatal(err)
	}
	if len(server.createRequests) != 1 {
		t.Fatalf("got %d creates, want 1", len(server.createRequests))
	}
	request := server.createRequests[0]
	if request.Parent != "folders/123/locations/global" || request.EntitlementId != "ssb-bucket-read" {
		t.Fatalf("incorrect folder scope: %v", request)
	}
	entitlement := request.Entitlement
	access := entitlement.PrivilegedAccess.GetGcpIamAccess()
	if entitlement.Name != "folders/123/locations/global/entitlements/ssb-bucket-read" ||
		access.Resource != "//cloudresourcemanager.googleapis.com/folders/123" ||
		access.ResourceType != "cloudresourcemanager.googleapis.com/Folder" {
		t.Fatalf("incorrect entitlement scope: %v", entitlement)
	}
	if len(access.RoleBindings) != 1 || access.RoleBindings[0].Role != role ||
		access.RoleBindings[0].ConditionExpression != condition {
		t.Fatalf("incorrect roles or unexpected condition: %v", access.RoleBindings)
	}
	if len(entitlement.EligibleUsers) != 1 {
		t.Fatalf("incorrect eligible principals: %v", entitlement.EligibleUsers)
	}
	if diff := cmp.Diff(principals, entitlement.EligibleUsers[0].Principals); diff != "" {
		t.Fatalf("incorrect eligible principals (-want +got):\n%s", diff)
	}
	if entitlement.ApprovalWorkflow != nil || entitlement.RequesterJustificationConfig.GetNotMandatory() == nil ||
		entitlement.MaxRequestDuration.AsDuration() != time.Hour {
		t.Fatalf("incorrect grant policy: %v", entitlement)
	}

	if err := client.EnsureEntitlement(t.Context(), "123", role, condition, principals, time.Hour); err != nil {
		t.Fatal(err)
	}
	if len(server.createRequests) != 1 || len(server.updateRequests) != 0 {
		t.Fatal("unchanged entitlement must not be created or updated again")
	}

	server.entitlement.PrivilegedAccess.GetGcpIamAccess().RoleBindings[0].Role = "roles/storage.admin"
	server.entitlement.ApprovalWorkflow = &privilegedaccessmanagerpb.ApprovalWorkflow{
		ApprovalWorkflow: &privilegedaccessmanagerpb.ApprovalWorkflow_ManualApprovals{
			ManualApprovals: &privilegedaccessmanagerpb.ManualApprovals{},
		},
	}
	newPrincipals := []string{"group:other-readers@example.com"}
	newRole := "organizations/573742569423/roles/ssb.bucket.read"
	newCondition := `resource.name.startsWith("projects/_/buckets/ssb-my-team-data-produkt-prod/")`
	if err := client.EnsureEntitlement(t.Context(), "123", newRole, newCondition, newPrincipals, 2*time.Hour); err != nil {
		t.Fatal(err)
	}
	if len(server.updateRequests) != 1 {
		t.Fatalf("got %d updates, want 1", len(server.updateRequests))
	}
	update := server.updateRequests[0]
	if update.Entitlement.Etag != "current-version" {
		t.Fatal("update must use existing etag")
	}
	expected := proto.Clone(entitlement).(*privilegedaccessmanagerpb.Entitlement)
	expected.Etag = "current-version"
	expected.EligibleUsers[0].Principals = newPrincipals
	expected.PrivilegedAccess.GetGcpIamAccess().RoleBindings[0].Role = newRole
	expected.PrivilegedAccess.GetGcpIamAccess().RoleBindings[0].ConditionExpression = newCondition
	expected.MaxRequestDuration = durationpb.New(2 * time.Hour)
	if diff := cmp.Diff(expected, update.Entitlement, protocmp.Transform()); diff != "" {
		t.Fatalf("update policy mismatch (-want +got):\n%s", diff)
	}
	wantPaths := []string{"eligible_users", "privileged_access", "max_request_duration", "approval_workflow", "requester_justification_config"}
	if diff := cmp.Diff(wantPaths, update.UpdateMask.Paths); diff != "" {
		t.Fatalf("update mask mismatch (-want +got):\n%s", diff)
	}
}

func TestEnsureEntitlement_UpdatesOnlyCondition(t *testing.T) {
	server := &fakePAMServer{}
	client := newFakePAMClient(t, server)
	principals := []string{"user:reader@example.com"}
	for _, condition := range []string{"false", `resource.type == "storage.googleapis.com/Object"`, `resource.type == "storage.googleapis.com/Object"`} {
		if err := client.EnsureEntitlement(t.Context(), "123", "roles/storage.objectViewer", condition, principals, time.Hour); err != nil {
			t.Fatal(err)
		}
	}
	if len(server.createRequests) != 1 || len(server.updateRequests) != 1 {
		t.Fatalf("expected one create and one condition-only update, got %d creates and %d updates", len(server.createRequests), len(server.updateRequests))
	}
	if server.entitlement.PrivilegedAccess.GetGcpIamAccess().RoleBindings[0].ConditionExpression != `resource.type == "storage.googleapis.com/Object"` {
		t.Fatalf("condition was not updated: %v", server.entitlement)
	}
}

func TestEnsureEntitlement_RejectsEmptyCondition(t *testing.T) {
	server := &fakePAMServer{}
	client := newFakePAMClient(t, server)
	err := client.EnsureEntitlement(t.Context(), "123", "roles/storage.objectViewer", "  ", []string{"user:reader@example.com"}, time.Hour)
	if err == nil || !strings.Contains(err.Error(), "PAM IAM condition is required") {
		t.Fatalf("got %v, want missing condition error", err)
	}
	if len(server.createRequests) != 0 {
		t.Fatal("an unconditional entitlement must not be created")
	}
}

func TestEnsureEntitlement_PropagatesFailures(t *testing.T) {
	tests := map[string]struct {
		getErr       error
		createErr    error
		updateErr    error
		operationErr *rpcstatus.Status
		existing     bool
		want         string
	}{
		"get":              {getErr: status.Error(codes.PermissionDenied, "denied"), want: "get entitlement"},
		"create":           {createErr: status.Error(codes.PermissionDenied, "denied"), want: "create entitlement"},
		"create operation": {operationErr: &rpcstatus.Status{Code: int32(codes.Internal), Message: "failed"}, want: "wait for entitlement creation"},
		"update":           {existing: true, updateErr: status.Error(codes.Aborted, "conflict"), want: "update entitlement"},
		"update operation": {existing: true, operationErr: &rpcstatus.Status{Code: int32(codes.Internal), Message: "failed"}, want: "wait for entitlement update"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &fakePAMServer{getErr: tt.getErr, createErr: tt.createErr, updateErr: tt.updateErr, operationErr: tt.operationErr}
			if tt.existing {
				server.entitlement = &privilegedaccessmanagerpb.Entitlement{
					Name: "folders/123/locations/global/entitlements/ssb-bucket-read",
					Etag: "current-version",
				}
			}
			client := newFakePAMClient(t, server)
			err := client.EnsureEntitlement(t.Context(), "123", "roles/storage.objectViewer", "false", []string{"user:reader@example.com"}, time.Hour)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v, want error containing %q", err, tt.want)
			}
		})
	}
}
