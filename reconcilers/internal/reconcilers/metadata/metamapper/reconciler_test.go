package metamapper_test

import (
	"bytes"
	"context"
	"errors"
	"net"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers/metadata/metamapper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const testPAMConditionTemplate = `resource.name == "projects/_/buckets/ssb-{{.TeamSlug}}-data-produkt/objects/example"`

type fakeGcpTeamResourcesServer struct {
	protoapi.UnimplementedGcpTeamResourcesServer
	protoapi.UnimplementedReconcilersServer
	folders        map[string]*protoapi.GcpTeamFolder
	config         map[string]string
	configErr      error
	getErr         error
	getRequests    []*protoapi.GetGcpTeamFolderRequest
	configRequests []*protoapi.ConfigReconcilerRequest
}

func newFakeGcpTeamResourcesServer() *fakeGcpTeamResourcesServer {
	return &fakeGcpTeamResourcesServer{
		folders: make(map[string]*protoapi.GcpTeamFolder),
		config: map[string]string{
			"test_entitlement_id":     "test-bucket-read",
			"test_role":               "roles/storage.objectViewer",
			"test_principals":         "user:reader@example.com",
			"test_max_grant_duration": "1h",
			"test_condition_template": testPAMConditionTemplate,
		},
	}
}

func (s *fakeGcpTeamResourcesServer) Config(_ context.Context, req *protoapi.ConfigReconcilerRequest) (*protoapi.ConfigReconcilerResponse, error) {
	s.configRequests = append(s.configRequests, req)
	if req.ReconcilerName != "metadata:metamapper:pam" {
		return nil, status.Errorf(codes.InvalidArgument, "unexpected reconciler %q", req.ReconcilerName)
	}
	if s.configErr != nil {
		return nil, s.configErr
	}
	resp := &protoapi.ConfigReconcilerResponse{}
	for key, value := range s.config {
		resp.Nodes = append(resp.Nodes, &protoapi.ReconcilerConfig{Key: key, Value: value})
	}
	return resp, nil
}

func (s *fakeGcpTeamResourcesServer) GetTeamFolder(_ context.Context, req *protoapi.GetGcpTeamFolderRequest) (*protoapi.GetGcpTeamFolderResponse, error) {
	s.getRequests = append(s.getRequests, req)
	if s.getErr != nil {
		return nil, s.getErr
	}
	folder, ok := s.folders[req.TeamSlug+"/"+req.Env]
	if !ok {
		return nil, status.Error(codes.NotFound, "folder not found")
	}
	return &protoapi.GetGcpTeamFolderResponse{Folder: folder}, nil
}

func startFakeGrpcServer(t *testing.T, srv *fakeGcpTeamResourcesServer) *apiclient.APIClient {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	protoapi.RegisterGcpTeamResourcesServer(s, srv)
	protoapi.RegisterReconcilersServer(s, srv)
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)
	client, err := apiclient.New(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

type entitlementCall struct {
	folderID            string
	entitlementID       string
	role                string
	conditionExpression string
	principals          []string
	duration            time.Duration
}

type permissionsCall struct {
	folderID   string
	principals []string
}

type fakePrivilegedAccessManager struct {
	calls            []entitlementCall
	permissionsCalls []permissionsCall
	err              error
}

func (f *fakePrivilegedAccessManager) EnsureEntitlement(_ context.Context, folderID, entitlementID, role, conditionExpression string, principals []string, duration time.Duration) error {
	f.calls = append(f.calls, entitlementCall{folderID, entitlementID, role, conditionExpression, principals, duration})
	return f.err
}

func (f *fakePrivilegedAccessManager) EnsurePermissions(_ context.Context, folderID string, principals []string) error {
	f.permissionsCalls = append(f.permissionsCalls, permissionsCall{folderID, principals})
	return nil
}

func TestNew_ValidatesEnvironments(t *testing.T) {
	for name, tt := range map[string]struct {
		environments []string
		wantErr      string
	}{
		"missing": {wantErr: "at least one PAM environment must be configured"},
		"empty":   {environments: []string{"TEST", " "}, wantErr: "PAM environments must not contain empty entries"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := metamapper.New(t.Context(), metamapper.Config{Environments: tt.environments}, nil)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestReconcile_IndependentEnvironmentConfiguration(t *testing.T) {
	server := newFakeGcpTeamResourcesServer()
	server.config["test_team_allowlist"] = " other-team, my-team "
	server.config["prod_entitlement_id"] = " prod-bucket-read "
	server.config["prod_role"] = " roles/storage.objectAdmin "
	server.config["prod_principals"] = " group:prod@example.com, user:admin@example.com "
	server.config["prod_max_grant_duration"] = " 2h "
	server.config["prod_condition_template"] = `resource.name.startsWith("{{.TeamSlug}}-prod")`
	server.config["prod_team_allowlist"] = "my-team"
	for env, id := range map[string]string{"TEST": "100", "PROD": "200"} {
		server.folders["my-team/"+env] = &protoapi.GcpTeamFolder{FolderId: id}
	}
	pam := &fakePrivilegedAccessManager{}
	cfg := metamapper.Config{Environments: []string{"TEST", "PROD"}}
	r, err := metamapper.New(t.Context(), cfg, nil, metamapper.WithPrivilegedAccessManager(pam))
	if err != nil {
		t.Fatal(err)
	}
	cfg.Environments[0] = "other"
	var keys []string
	for _, spec := range r.Configuration().Config {
		keys = append(keys, spec.Key)
	}
	if !slices.Equal(keys, []string{
		"test_entitlement_id", "test_role", "test_principals", "test_max_grant_duration", "test_condition_template", "test_team_allowlist",
		"prod_entitlement_id", "prod_role", "prod_principals", "prod_max_grant_duration", "prod_condition_template", "prod_team_allowlist",
	}) {
		t.Fatalf("unexpected API configuration keys: %v", keys)
	}
	client := startFakeGrpcServer(t, server)
	team := &protoapi.Team{Slug: "my-team", IsManaged: true}
	log := logrus.New()
	if err := r.Reconcile(t.Context(), client, team, log); err != nil {
		t.Fatal(err)
	}
	if len(pam.calls) != 2 || len(server.configRequests) != 1 {
		t.Fatalf("expected two entitlements from one config snapshot, got %+v", pam.calls)
	}
	test, prod := pam.calls[0], pam.calls[1]
	if test.folderID != "100" || test.entitlementID != "test-bucket-read" || test.role != "roles/storage.objectViewer" || test.duration != time.Hour ||
		!slices.Equal(test.principals, []string{"user:reader@example.com"}) ||
		test.conditionExpression != `resource.name == "projects/_/buckets/ssb-my-team-data-produkt/objects/example"` {
		t.Fatalf("unexpected TEST policy: %+v", test)
	}
	if prod.folderID != "200" || prod.entitlementID != "prod-bucket-read" || prod.role != "roles/storage.objectAdmin" || prod.duration != 2*time.Hour ||
		!slices.Equal(prod.principals, []string{"group:prod@example.com", "user:admin@example.com"}) ||
		prod.conditionExpression != `resource.name.startsWith("my-team-prod")` {
		t.Fatalf("unexpected PROD policy: %+v", prod)
	}
	server.config["test_team_allowlist"] = "another-team"
	server.config["prod_role"] = "roles/storage.objectViewer"
	server.config["prod_entitlement_id"] = "updated-prod-reader"
	if err := r.Reconcile(t.Context(), client, team, log); err != nil {
		t.Fatal(err)
	}
	if len(pam.calls) != 3 || pam.calls[2].folderID != "200" || pam.calls[2].entitlementID != "updated-prod-reader" || pam.calls[2].role != "roles/storage.objectViewer" {
		t.Fatalf("per-environment changes were not applied: %+v", pam.calls)
	}
	server.config["test_team_allowlist"] = ""
	delete(server.config, "prod_condition_template")
	var output bytes.Buffer
	log.SetOutput(&output)
	if err := r.Reconcile(t.Context(), client, team, log); err != nil {
		t.Fatal(err)
	}
	if len(pam.calls) != 4 || pam.calls[3].folderID != "100" {
		t.Fatalf("incomplete PROD must not disable TEST: calls=%+v logs=%s", pam.calls, output.String())
	}
	delete(server.config, "test_condition_template")
	if err := r.Reconcile(t.Context(), client, team, log); err != nil {
		t.Fatal(err)
	}
	if len(pam.calls) != 4 {
		t.Fatal("cleared policy must not reuse stale configuration")
	}
	server.config["prod_condition_template"] = `resource.name.startsWith("{{.TeamSlug}}-prod")`
	if err := r.Reconcile(t.Context(), client, team, log); err != nil {
		t.Fatal(err)
	}
	if len(pam.calls) != 5 || pam.calls[4].folderID != "200" {
		t.Fatal("PROD must run independently while TEST is unconfigured")
	}
	server.config["prod_condition_template"] = "{{"
	if err := r.Reconcile(t.Context(), client, team, log); err == nil || !strings.Contains(err.Error(), "env \"PROD\": parse PAM IAM condition template") {
		t.Fatalf("malformed updated PROD template must return an environment-specific error, got %v", err)
	}
	server.configErr = status.Error(codes.Unavailable, "configuration unavailable")
	if err := r.Reconcile(t.Context(), client, team, log); status.Code(err) != codes.Unavailable {
		t.Fatalf("configuration API error must be propagated, got %v", err)
	}
	if len(pam.calls) != 5 {
		t.Fatal("invalid or unavailable configuration must not reuse a previous policy")
	}
}

func TestReconcile_EnvironmentPolicyValidation(t *testing.T) {
	for name, tt := range map[string]struct {
		key, value, wantErr string
		remove              bool
		wantSkip            bool
	}{
		"missing ID":        {key: "test_entitlement_id", remove: true, wantSkip: true},
		"blank ID":          {key: "test_entitlement_id", value: " ", wantSkip: true},
		"invalid ID":        {key: "test_entitlement_id", value: "invalid/id", wantErr: "PAM entitlement ID must be"},
		"missing role":      {key: "test_role", remove: true, wantSkip: true},
		"missing users":     {key: "test_principals", remove: true, wantSkip: true},
		"missing duration":  {key: "test_max_grant_duration", remove: true, wantSkip: true},
		"missing condition": {key: "test_condition_template", remove: true, wantSkip: true},
		"blank role":        {key: "test_role", value: " ", wantSkip: true},
		"blank users":       {key: "test_principals", value: " ", wantSkip: true},
		"blank condition":   {key: "test_condition_template", value: " ", wantSkip: true},
		"empty user entry":  {key: "test_principals", value: "user:a@example.com,", wantErr: "principals must not contain empty entries"},
		"invalid duration":  {key: "test_max_grant_duration", value: "one hour", wantErr: "invalid duration"},
		"too short":         {key: "test_max_grant_duration", value: "29m59s", wantErr: "between 30m and 168h"},
		"negative":          {key: "test_max_grant_duration", value: "-1h", wantErr: "between 30m and 168h"},
		"too long":          {key: "test_max_grant_duration", value: "168h1s", wantErr: "between 30m and 168h"},
		"minimum":           {key: "test_max_grant_duration", value: "30m"},
		"maximum":           {key: "test_max_grant_duration", value: "168h"},
		"group":             {key: "test_principals", value: "group:readers@example.com"},
		"federated":         {key: "test_principals", value: "principal://iam.googleapis.com/locations/global/workforcePools/example/subject/reader"},
		"malformed":         {key: "test_condition_template", value: "{{", wantErr: "parse PAM IAM condition template"},
		"unknown variable":  {key: "test_condition_template", value: "{{.Unknown}}", wantErr: "render PAM IAM condition"},
		"empty expression":  {key: "test_condition_template", value: `{{if eq .TeamSlug "absent"}}true{{end}}`, wantErr: "empty expression"},
		"unknown key":       {key: "test_unknown", value: "x", wantErr: "unknown config key"},
		"old key":           {key: "role", value: "x", wantErr: "unknown config key"},
	} {
		t.Run(name, func(t *testing.T) {
			server := newFakeGcpTeamResourcesServer()
			server.config[tt.key] = tt.value
			if tt.remove {
				delete(server.config, tt.key)
			}
			server.folders["my-team/TEST"] = &protoapi.GcpTeamFolder{FolderId: "100"}
			pam := &fakePrivilegedAccessManager{}
			r, err := metamapper.New(t.Context(), metamapper.Config{Environments: []string{"TEST"}}, nil, metamapper.WithPrivilegedAccessManager(pam))
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			log := logrus.New()
			log.SetOutput(&output)
			err = r.Reconcile(t.Context(), startFakeGrpcServer(t, server), &protoapi.Team{Slug: "my-team", IsManaged: true}, log)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) || len(pam.calls) != 0 {
					t.Fatalf("got %v and %d PAM calls, want %q", err, len(pam.calls), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantSkip {
				if len(pam.calls) != 0 || len(server.getRequests) != 0 {
					t.Fatalf("incomplete configuration must skip")
				}
			} else if len(pam.calls) != 1 {
				t.Fatalf("got %d PAM calls, want 1", len(pam.calls))
			}
		})
	}
}

func TestReconcile_TeamAllowlist(t *testing.T) {
	for _, tt := range []struct {
		value            string
		allowed, invalid bool
	}{
		{value: "", allowed: true},
		{value: " \t ", allowed: true},
		{value: "other, my-team ", allowed: true},
		{value: "my-team-extra,other-my-team"},
		{value: "My-Team"},
		{value: "*"},
		{value: "my-team,", invalid: true},
		{value: ",", invalid: true},
	} {
		t.Run(tt.value, func(t *testing.T) {
			server := newFakeGcpTeamResourcesServer()
			server.config["test_team_allowlist"] = tt.value
			server.folders["my-team/TEST"] = &protoapi.GcpTeamFolder{FolderId: "100"}
			pam := &fakePrivilegedAccessManager{}
			r, err := metamapper.New(t.Context(), metamapper.Config{Environments: []string{"TEST"}}, nil, metamapper.WithPrivilegedAccessManager(pam))
			if err != nil {
				t.Fatal(err)
			}
			err = r.Reconcile(t.Context(), startFakeGrpcServer(t, server), &protoapi.Team{Slug: "my-team", IsManaged: true}, logrus.New())
			if tt.invalid {
				if err == nil || !strings.Contains(err.Error(), "team allowlist must not contain empty entries") {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if tt.allowed {
				if len(pam.calls) != 1 {
					t.Fatal("allowed team must receive entitlement")
				}
			} else if len(pam.calls) != 0 || len(server.getRequests) != 0 {
				t.Fatal("excluded team must not query folders or PAM")
			}
		})
	}
}

func TestReconcile_FailuresAndManagedGating(t *testing.T) {
	for name, tt := range map[string]struct {
		configErr, folderErr, pamErr error
		emptyFolder, unmanaged       bool
		wantErr                      string
	}{
		"config API":     {configErr: status.Error(codes.Unavailable, "unavailable"), wantErr: "get reconciler config"},
		"missing folder": {folderErr: status.Error(codes.NotFound, "missing"), wantErr: "get stored folder"},
		"empty folder":   {emptyFolder: true, wantErr: "is empty"},
		"PAM":            {pamErr: errors.New("PAM unavailable"), wantErr: "PAM unavailable"},
		"self managed":   {unmanaged: true},
	} {
		t.Run(name, func(t *testing.T) {
			server := newFakeGcpTeamResourcesServer()
			server.configErr, server.getErr = tt.configErr, tt.folderErr
			if !tt.emptyFolder {
				server.folders["my-team/TEST"] = &protoapi.GcpTeamFolder{FolderId: "100"}
			} else {
				server.folders["my-team/TEST"] = nil
			}
			pam := &fakePrivilegedAccessManager{err: tt.pamErr}
			r, err := metamapper.New(t.Context(), metamapper.Config{Environments: []string{"TEST"}}, nil, metamapper.WithPrivilegedAccessManager(pam))
			if err != nil {
				t.Fatal(err)
			}
			err = r.Reconcile(t.Context(), startFakeGrpcServer(t, server), &protoapi.Team{Slug: "my-team", IsManaged: !tt.unmanaged}, logrus.New())
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("got %v, want %q", err, tt.wantErr)
			}
			if tt.unmanaged && (err != nil || len(server.configRequests) != 0 || len(server.getRequests) != 0 || len(pam.calls) != 0) {
				t.Fatal("self-managed team must be skipped entirely")
			}
			if len(pam.permissionsCalls) != 0 {
				t.Fatalf("permissions must only be ensured after a successful entitlement, got %+v", pam.permissionsCalls)
			}
		})
	}
}

func TestReconcile_EnsuresPermissionsAfterEntitlement(t *testing.T) {
	server := newFakeGcpTeamResourcesServer()
	server.folders["my-team/TEST"] = &protoapi.GcpTeamFolder{FolderId: "100"}
	pam := &fakePrivilegedAccessManager{}
	r, err := metamapper.New(t.Context(), metamapper.Config{Environments: []string{"TEST"}}, nil, metamapper.WithPrivilegedAccessManager(pam))
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Reconcile(t.Context(), startFakeGrpcServer(t, server), &protoapi.Team{Slug: "my-team", IsManaged: true}, logrus.New()); err != nil {
		t.Fatal(err)
	}
	if len(pam.permissionsCalls) != 1 || pam.permissionsCalls[0].folderID != "100" || !slices.Equal(pam.permissionsCalls[0].principals, []string{"user:reader@example.com"}) {
		t.Fatalf("unexpected permissions calls: %+v", pam.permissionsCalls)
	}
}
