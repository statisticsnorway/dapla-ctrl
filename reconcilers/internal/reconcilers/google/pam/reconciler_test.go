package pam_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers/google/gcpresources"
	googlepam "github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers/google/pam"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// fakeGcpTeamResourcesServer implements protoapi.GcpTeamResourcesServer in memory.
type fakeGcpTeamResourcesServer struct {
	protoapi.UnimplementedGcpTeamResourcesServer
	protoapi.UnimplementedReconcilersServer
	folders        map[string]*protoapi.GcpTeamFolder // "teamSlug/env" -> folder
	getErr         error
	getRequests    []*protoapi.GetGcpTeamFolderRequest
	upsertCalls    int
	config         map[string]string
	configErr      error
	configRequests []*protoapi.ConfigReconcilerRequest
}

func newFakeGcpTeamResourcesServer() *fakeGcpTeamResourcesServer {
	return &fakeGcpTeamResourcesServer{
		folders: make(map[string]*protoapi.GcpTeamFolder),
		config: map[string]string{
			"role":               "roles/storage.objectViewer",
			"principals":         "user:reader@example.com",
			"max_grant_duration": "1h",
			"condition_template": testPAMConditionTemplate,
		},
	}
}

func (s *fakeGcpTeamResourcesServer) Config(_ context.Context, req *protoapi.ConfigReconcilerRequest) (*protoapi.ConfigReconcilerResponse, error) {
	s.configRequests = append(s.configRequests, req)
	if req.ReconcilerName != "google:pam" {
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

func (s *fakeGcpTeamResourcesServer) UpsertTeamFolder(_ context.Context, req *protoapi.UpsertGcpTeamFolderRequest) (*protoapi.UpsertGcpTeamFolderResponse, error) {
	s.upsertCalls++
	key := req.Folder.TeamSlug + "/" + req.Folder.Env
	s.folders[key] = req.Folder
	return &protoapi.UpsertGcpTeamFolderResponse{}, nil
}

func (s *fakeGcpTeamResourcesServer) GetTeamFolder(_ context.Context, req *protoapi.GetGcpTeamFolderRequest) (*protoapi.GetGcpTeamFolderResponse, error) {
	s.getRequests = append(s.getRequests, req)
	if s.getErr != nil {
		return nil, s.getErr
	}
	key := req.TeamSlug + "/" + req.Env
	f, ok := s.folders[key]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "folder not found")
	}
	return &protoapi.GetGcpTeamFolderResponse{Folder: f}, nil
}

// startFakeGrpcServer starts an in-process gRPC server and returns a connected APIClient.
func startFakeGrpcServer(t *testing.T, srv *fakeGcpTeamResourcesServer) *apiclient.APIClient {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	protoapi.RegisterGcpTeamResourcesServer(s, srv)
	protoapi.RegisterReconcilersServer(s, srv)
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)

	client, err := apiclient.New(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("create api client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return client
}

type entitlementCall struct {
	folderID            string
	role                string
	conditionExpression string
	principals          []string
	duration            time.Duration
}

type fakePrivilegedAccessManager struct {
	calls []entitlementCall
	err   error
}

func (f *fakePrivilegedAccessManager) EnsureEntitlement(_ context.Context, folderID, role, conditionExpression string, principals []string, duration time.Duration) error {
	f.calls = append(f.calls, entitlementCall{folderID, role, conditionExpression, principals, duration})
	return f.err
}

const testPAMConditionTemplate = `resource.name == "projects/_/buckets/ssb-{{.TeamSlug}}-data-produkt-{{.Env}}/objects/example"`

func TestNew_ValidatesEnvironments(t *testing.T) {
	for name, tt := range map[string]struct {
		environments []string
		wantErr      string
	}{
		"missing": {wantErr: "at least one PAM environment must be configured"},
		"empty":   {environments: []string{"dev", " "}, wantErr: "PAM environments must not contain empty entries"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := googlepam.New(t.Context(), googlepam.Config{Environments: tt.environments}, nil)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestReconcile_StoredFolderErrors(t *testing.T) {
	for name, tt := range map[string]struct {
		stored   bool
		folder   *protoapi.GcpTeamFolder
		getErr   error
		wantErr  string
		wantCode codes.Code
	}{
		"missing": {
			wantErr:  `get stored folder for team "my-team" in env "dev"`,
			wantCode: codes.NotFound,
		},
		"API unavailable": {
			getErr:   status.Error(codes.Unavailable, "API unavailable"),
			wantErr:  `get stored folder for team "my-team" in env "dev"`,
			wantCode: codes.Unavailable,
		},
		"empty folder ID": {
			stored:   true,
			folder:   &protoapi.GcpTeamFolder{TeamSlug: "my-team", Env: "dev"},
			wantErr:  `stored folder ID for team "my-team" in env "dev" is empty`,
			wantCode: codes.Unknown,
		},
		"missing folder in response": {
			stored:   true,
			wantErr:  `stored folder ID for team "my-team" in env "dev" is empty`,
			wantCode: codes.Unknown,
		},
	} {
		t.Run(name, func(t *testing.T) {
			server := newFakeGcpTeamResourcesServer()
			server.getErr = tt.getErr
			if tt.stored {
				server.folders["my-team/dev"] = tt.folder
			}
			pam := &fakePrivilegedAccessManager{}
			r, err := googlepam.New(t.Context(), googlepam.Config{
				Environments: []string{"dev"},
			}, nil, googlepam.WithPrivilegedAccessManager(pam))
			if err != nil {
				t.Fatal(err)
			}
			err = r.Reconcile(t.Context(), startFakeGrpcServer(t, server), &protoapi.Team{Slug: "my-team", IsManaged: true}, logrus.New())
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) || status.Code(err) != tt.wantCode {
				t.Fatalf("got %v, want %v error containing %q", err, tt.wantCode, tt.wantErr)
			}
			if len(pam.calls) != 0 || server.upsertCalls != 0 {
				t.Fatal("invalid or missing stored folders must not create folders or entitlements")
			}
		})
	}
}

func TestReconcile_AfterFolderReconciliation(t *testing.T) {
	for _, configured := range []bool{true, false} {
		t.Run(fmt.Sprintf("configured=%t", configured), func(t *testing.T) {
			server := newFakeGcpTeamResourcesServer()
			client := startFakeGrpcServer(t, server)
			resources := gcpresources.NewFakeResourceManager()
			folderReconciler, err := gcpresources.New(t.Context(), gcpresources.Config{
				TagKeyNamespacedName: "tagKeys/123456",
				EnvParentFolders:     map[string]string{"dev": "111"},
			}, nil, gcpresources.WithResourceManager(resources))
			if err != nil {
				t.Fatal(err)
			}
			cfg := googlepam.Config{
				Environments: []string{"dev"},
			}
			if !configured {
				server.config["role"] = ""
			}
			pam := &fakePrivilegedAccessManager{}
			pamReconciler, err := googlepam.New(t.Context(), cfg, nil, googlepam.WithPrivilegedAccessManager(pam))
			if err != nil {
				t.Fatal(err)
			}
			team := &protoapi.Team{Slug: "my-team", IsManaged: true}
			if err := folderReconciler.Reconcile(t.Context(), client, team, logrus.New()); err != nil {
				t.Fatal(err)
			}
			if err := pamReconciler.Reconcile(t.Context(), client, team, logrus.New()); err != nil {
				t.Fatal(err)
			}
			folder := server.folders["my-team/dev"]
			if folder == nil || resources.Tags[folder.FolderId] == "" || server.upsertCalls != 1 {
				t.Fatal("folder and tag reconciliation must remain independent of PAM configuration")
			}
			if configured {
				if len(pam.calls) != 1 || pam.calls[0].folderID != folder.FolderId {
					t.Fatalf("PAM must use the stored folder ID, got %+v", pam.calls)
				}
			} else if len(pam.calls) != 0 {
				t.Fatal("incomplete PAM configuration must not create entitlements")
			}
		})
	}
}

func TestReconcile_PAMOnlyForManagedTeams(t *testing.T) {
	for _, managed := range []bool{true, false} {
		name := "self-managed"
		if managed {
			name = "managed"
		}
		t.Run(name, func(t *testing.T) {
			server := newFakeGcpTeamResourcesServer()
			for env, id := range map[string]string{"dev": "100", "test": "200", "prod": "300"} {
				server.folders["my-team/"+env] = &protoapi.GcpTeamFolder{TeamSlug: "my-team", Env: env, FolderId: id}
			}
			pam := &fakePrivilegedAccessManager{}
			server.config = map[string]string{
				"principals":         " serviceAccount:reader@test.iam.gserviceaccount.com, group:readers@example.com ",
				"role":               " roles/storage.objectViewer ",
				"max_grant_duration": "2h",
				"condition_template": testPAMConditionTemplate,
				"team_allowlist":     "my-team",
			}
			r, err := googlepam.New(t.Context(), googlepam.Config{
				Environments: []string{"dev", "test", "prod"},
			}, nil, googlepam.WithPrivilegedAccessManager(pam))
			if err != nil {
				t.Fatal(err)
			}

			err = r.Reconcile(t.Context(), startFakeGrpcServer(t, server), &protoapi.Team{Slug: "my-team", IsManaged: managed}, logrus.New())
			if err != nil {
				t.Fatal(err)
			}
			if server.upsertCalls != 0 {
				t.Fatal("PAM must not create or store folders")
			}
			if !managed {
				if len(pam.calls) != 0 || len(server.getRequests) != 0 || len(server.configRequests) != 0 {
					t.Fatal("self-managed team must not use PAM or query configuration or folders")
				}
				return
			}
			if len(pam.calls) != 3 {
				t.Fatalf("got %d PAM calls, want 3", len(pam.calls))
			}
			if len(server.getRequests) != 3 {
				t.Fatalf("got %d folder requests, want 3", len(server.getRequests))
			}
			if len(server.configRequests) != 1 {
				t.Fatalf("got %d configuration requests, want one per team", len(server.configRequests))
			}
			folders := make(map[string]bool)
			for _, call := range pam.calls {
				folders[call.folderID] = true
				if !slices.Equal(call.principals, []string{"serviceAccount:reader@test.iam.gserviceaccount.com", "group:readers@example.com"}) ||
					call.role != "roles/storage.objectViewer" || call.duration != 2*time.Hour {
					t.Errorf("unexpected entitlement: %+v", call)
				}
				env := map[string]string{"100": "dev", "200": "test", "300": "prod"}[call.folderID]
				expectedCondition := fmt.Sprintf(`resource.name == "projects/_/buckets/ssb-my-team-data-produkt-%s/objects/example"`, env)
				if call.conditionExpression != expectedCondition {
					t.Errorf("incorrect condition for folder %s: %s", call.folderID, call.conditionExpression)
				}
			}
			for _, id := range []string{"100", "200", "300"} {
				if !folders[id] {
					t.Errorf("missing entitlement for stored folder %s", id)
				}
			}
		})
	}
}

func TestReconcile_TeamAllowlist(t *testing.T) {
	for name, tt := range map[string]struct {
		value       string
		omit        bool
		wantAllowed bool
		wantErr     bool
	}{
		"unset allows all":          {omit: true, wantAllowed: true},
		"empty allows all":          {wantAllowed: true},
		"whitespace allows all":     {value: " \t ", wantAllowed: true},
		"exact match":               {value: "my-team", wantAllowed: true},
		"trimmed multiple entries":  {value: " another-team, my-team ", wantAllowed: true},
		"not listed":                {value: "another-team"},
		"prefix and substring":      {value: "my-team-extra,other-my-team"},
		"case sensitive":            {value: "My-Team"},
		"wildcard is not supported": {value: "*"},
		"trailing comma":            {value: "my-team,", wantErr: true},
		"blank entry":               {value: "my-team, ,another-team", wantErr: true},
		"comma only":                {value: ",", wantErr: true},
	} {
		t.Run(name, func(t *testing.T) {
			server := newFakeGcpTeamResourcesServer()
			server.folders["my-team/dev"] = &protoapi.GcpTeamFolder{TeamSlug: "my-team", Env: "dev", FolderId: "100"}
			if !tt.omit {
				server.config["team_allowlist"] = tt.value
			}
			pam := &fakePrivilegedAccessManager{}
			r, err := googlepam.New(t.Context(), googlepam.Config{
				Environments: []string{"dev"},
			}, nil, googlepam.WithPrivilegedAccessManager(pam))
			if err != nil {
				t.Fatal(err)
			}
			err = r.Reconcile(t.Context(), startFakeGrpcServer(t, server), &protoapi.Team{Slug: "my-team", IsManaged: true}, logrus.New())
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "PAM team allowlist must not contain empty entries") {
					t.Fatalf("got %v, want invalid allowlist error", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if tt.wantAllowed {
				if len(pam.calls) != 1 || pam.calls[0].folderID != "100" {
					t.Fatalf("allowed team must receive its folder entitlement, got %+v", pam.calls)
				}
			} else if len(pam.calls) != 0 || len(server.getRequests) != 0 {
				t.Fatal("excluded teams and invalid allowlists must not query folders or reconcile entitlements")
			}
		})
	}
}

func TestReconcile_TeamAllowlistChangesWithoutRestart(t *testing.T) {
	server := newFakeGcpTeamResourcesServer()
	server.folders["my-team/dev"] = &protoapi.GcpTeamFolder{TeamSlug: "my-team", Env: "dev", FolderId: "100"}
	client := startFakeGrpcServer(t, server)
	pam := &fakePrivilegedAccessManager{}
	r, err := googlepam.New(t.Context(), googlepam.Config{
		Environments: []string{"dev"},
	}, nil, googlepam.WithPrivilegedAccessManager(pam))
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct {
		allowlist string
		wantCalls int
	}{
		{allowlist: "another-team", wantCalls: 0},
		{allowlist: "my-team", wantCalls: 1},
		{allowlist: "another-team", wantCalls: 1},
		{allowlist: "", wantCalls: 2},
	} {
		server.config["team_allowlist"] = step.allowlist
		if err := r.Reconcile(t.Context(), client, &protoapi.Team{Slug: "my-team", IsManaged: true}, logrus.New()); err != nil {
			t.Fatal(err)
		}
		if len(pam.calls) != step.wantCalls {
			t.Fatalf("allowlist %q: got %d PAM calls, want %d", step.allowlist, len(pam.calls), step.wantCalls)
		}
	}
}

func TestReconcile_EnvironmentAllowlist(t *testing.T) {
	for name, tt := range map[string]struct {
		value       string
		omit        bool
		excludeTeam bool
		wantEnvs    []string
		wantErr     bool
	}{
		"unset allows all":          {omit: true, wantEnvs: []string{"dev", "test", "prod"}},
		"empty allows all":          {wantEnvs: []string{"dev", "test", "prod"}},
		"whitespace allows all":     {value: " \t ", wantEnvs: []string{"dev", "test", "prod"}},
		"exact match":               {value: "dev", wantEnvs: []string{"dev"}},
		"trimmed multiple entries":  {value: " test, dev ", wantEnvs: []string{"dev", "test"}},
		"unconfigured environment":  {value: "qa"},
		"prefix and substring":      {value: "dev-extra,pre-dev"},
		"case sensitive":            {value: "Dev"},
		"wildcard is not supported": {value: "*"},
		"team must also be allowed": {value: "dev", excludeTeam: true},
		"trailing comma":            {value: "dev,", wantErr: true},
		"blank entry":               {value: "dev, ,test", wantErr: true},
		"comma only":                {value: ",", wantErr: true},
	} {
		t.Run(name, func(t *testing.T) {
			server := newFakeGcpTeamResourcesServer()
			folders := map[string]string{"dev": "100", "test": "200", "prod": "300"}
			for env, id := range folders {
				server.folders["my-team/"+env] = &protoapi.GcpTeamFolder{TeamSlug: "my-team", Env: env, FolderId: id}
			}
			server.config["team_allowlist"] = "my-team"
			if tt.excludeTeam {
				server.config["team_allowlist"] = "another-team"
			}
			if !tt.omit {
				server.config["environment_allowlist"] = tt.value
			}
			pam := &fakePrivilegedAccessManager{}
			r, err := googlepam.New(t.Context(), googlepam.Config{
				Environments: []string{"dev", "test", "prod"},
			}, nil, googlepam.WithPrivilegedAccessManager(pam))
			if err != nil {
				t.Fatal(err)
			}
			err = r.Reconcile(t.Context(), startFakeGrpcServer(t, server), &protoapi.Team{Slug: "my-team", IsManaged: true}, logrus.New())
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "PAM environment allowlist must not contain empty entries") {
					t.Fatalf("got %v, want invalid allowlist error", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			var gotEnvs []string
			for _, req := range server.getRequests {
				gotEnvs = append(gotEnvs, req.Env)
			}
			if !slices.Equal(gotEnvs, tt.wantEnvs) {
				t.Fatalf("queried environments %v, want %v", gotEnvs, tt.wantEnvs)
			}
			if len(pam.calls) != len(tt.wantEnvs) {
				t.Fatalf("got %d PAM calls, want %d", len(pam.calls), len(tt.wantEnvs))
			}
			for i, env := range tt.wantEnvs {
				if pam.calls[i].folderID != folders[env] {
					t.Errorf("environment %q: entitlement used folder %q, want %q", env, pam.calls[i].folderID, folders[env])
				}
			}
		})
	}
}

func TestReconcile_EnvironmentAllowlistChangesWithoutRestart(t *testing.T) {
	server := newFakeGcpTeamResourcesServer()
	for env, id := range map[string]string{"dev": "100", "test": "200"} {
		server.folders["my-team/"+env] = &protoapi.GcpTeamFolder{TeamSlug: "my-team", Env: env, FolderId: id}
	}
	client := startFakeGrpcServer(t, server)
	pam := &fakePrivilegedAccessManager{}
	r, err := googlepam.New(t.Context(), googlepam.Config{
		Environments: []string{"dev", "test"},
	}, nil, googlepam.WithPrivilegedAccessManager(pam))
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct {
		allowlist   string
		omit        bool
		wantFolders []string
	}{
		{allowlist: "dev", wantFolders: []string{"100"}},
		{allowlist: "test", wantFolders: []string{"200"}},
		{allowlist: "", wantFolders: []string{"100", "200"}},
		{allowlist: "qa"},
		{omit: true, wantFolders: []string{"100", "200"}},
	} {
		server.config["environment_allowlist"] = step.allowlist
		if step.omit {
			delete(server.config, "environment_allowlist")
		}
		before := len(pam.calls)
		if err := r.Reconcile(t.Context(), client, &protoapi.Team{Slug: "my-team", IsManaged: true}, logrus.New()); err != nil {
			t.Fatal(err)
		}
		var gotFolders []string
		for _, call := range pam.calls[before:] {
			gotFolders = append(gotFolders, call.folderID)
		}
		if !slices.Equal(gotFolders, step.wantFolders) {
			t.Fatalf("allowlist %q (omitted=%t): reconciled folders %v, want %v", step.allowlist, step.omit, gotFolders, step.wantFolders)
		}
	}
}

func TestReconcile_PAMConfigurationFromAPI(t *testing.T) {
	tests := map[string]struct {
		principals        string
		role              string
		omitRole          bool
		omitCondition     bool
		conditionTemplate string
		duration          string
		pamErr            error
		configErr         error
		unknownKey        string
		wantErr           string
		wantSkip          bool
	}{
		"missing all":               {wantSkip: true},
		"blank principals":          {principals: "  ", role: "roles/storage.objectViewer", duration: "1h", wantSkip: true},
		"missing principals":        {role: "roles/storage.objectViewer", duration: "1h", wantSkip: true},
		"missing duration":          {principals: "serviceAccount:reader@test.iam.gserviceaccount.com", role: "roles/storage.objectViewer", wantSkip: true},
		"missing role":              {principals: "user:reader@example.com", duration: "1h", omitRole: true, wantSkip: true},
		"missing condition":         {principals: "user:reader@example.com", role: "roles/storage.objectViewer", duration: "1h", omitCondition: true, wantSkip: true},
		"blank condition":           {principals: "user:reader@example.com", role: "roles/storage.objectViewer", duration: "1h", conditionTemplate: "  ", wantSkip: true},
		"invalid template":          {principals: "user:reader@example.com", role: "roles/storage.objectViewer", duration: "1h", conditionTemplate: "{{", wantErr: "parse PAM IAM condition template"},
		"unknown template variable": {principals: "user:reader@example.com", role: "roles/storage.objectViewer", duration: "1h", conditionTemplate: "{{.Unknown}}", wantErr: "render PAM IAM condition"},
		"empty rendered condition":  {principals: "user:reader@example.com", role: "roles/storage.objectViewer", duration: "1h", conditionTemplate: `{{if eq .Env "absent"}}true{{end}}`, wantErr: "rendered to an empty expression"},
		"empty role":                {principals: "user:reader@example.com", duration: "1h", wantSkip: true},
		"whitespace role":           {principals: "user:reader@example.com", role: "  ", duration: "1h", wantSkip: true},
		"group principal":           {principals: "group:readers@example.com", role: "roles/storage.objectViewer", duration: "1h"},
		"multiple principals":       {principals: "user:reader@example.com,group:readers@example.com", role: "roles/storage.objectViewer", duration: "1h"},
		"federated principal":       {principals: "principal://iam.googleapis.com/locations/global/workforcePools/example/subject/reader", role: "roles/storage.objectViewer", duration: "1h"},
		"empty entry":               {principals: "user:reader@example.com, ,group:readers@example.com", duration: "1h", wantErr: "principals must not contain empty entries"},
		"trailing comma":            {principals: "user:reader@example.com,", duration: "1h", wantErr: "principals must not contain empty entries"},
		"configured role":           {principals: "user:reader@example.com", role: "roles/storage.objectViewer", duration: "1h"},
		"invalid duration":          {duration: "one hour", wantErr: "invalid duration"},
		"negative duration":         {duration: "-1h", wantErr: "must be between 30m and 168h"},
		"below minimum":             {duration: "29m59s", wantErr: "must be between 30m and 168h"},
		"above maximum":             {duration: "168h1s", wantErr: "must be between 30m and 168h"},
		"minimum":                   {principals: "serviceAccount:reader@test.iam.gserviceaccount.com", role: "roles/storage.objectViewer", duration: "30m"},
		"maximum":                   {principals: "serviceAccount:reader@test.iam.gserviceaccount.com", role: "roles/storage.objectViewer", duration: "168h"},
		"padded duration":           {principals: "user:reader@example.com", role: "roles/storage.objectViewer", duration: " 1h "},
		"API error":                 {configErr: status.Error(codes.Unavailable, "configuration unavailable"), wantErr: "get reconciler config"},
		"unknown key":               {unknownKey: "unknown", wantErr: `unknown config key "unknown"`},
		"PAM error":                 {principals: "serviceAccount:reader@test.iam.gserviceaccount.com", role: "roles/storage.objectViewer", duration: "1h", pamErr: errors.New("PAM unavailable"), wantErr: `folder "100" in env "dev": PAM unavailable`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := newFakeGcpTeamResourcesServer()
			server.folders["my-team/dev"] = &protoapi.GcpTeamFolder{TeamSlug: "my-team", Env: "dev", FolderId: "100"}
			values := map[string]string{
				"principals":         tt.principals,
				"max_grant_duration": tt.duration,
			}
			if !tt.omitRole {
				values["role"] = tt.role
			}
			if !tt.omitCondition {
				conditionTemplate := tt.conditionTemplate
				if conditionTemplate == "" {
					conditionTemplate = testPAMConditionTemplate
				}
				values["condition_template"] = conditionTemplate
			}
			if tt.unknownKey != "" {
				values[tt.unknownKey] = "value"
			}
			server.config = values
			server.configErr = tt.configErr
			pam := &fakePrivilegedAccessManager{err: tt.pamErr}
			r, err := googlepam.New(t.Context(), googlepam.Config{
				Environments: []string{"dev"},
			}, nil, googlepam.WithPrivilegedAccessManager(pam))
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			log := logrus.New()
			log.SetOutput(&output)
			err = r.Reconcile(t.Context(), startFakeGrpcServer(t, server), &protoapi.Team{Slug: "my-team", IsManaged: true}, log)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("got %v, want error containing %q", err, tt.wantErr)
				}
				if tt.pamErr == nil && len(pam.calls) != 0 {
					t.Fatal("invalid or unavailable configuration must not reach PAM")
				}
				if tt.configErr != nil && (status.Code(err) != codes.Unavailable || len(server.getRequests) != 0) {
					t.Fatal("API configuration errors must be propagated before querying folders")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantSkip {
				if len(pam.calls) != 0 || len(server.getRequests) != 0 || !strings.Contains(output.String(), "level=warning") {
					t.Fatalf("PAM should be skipped with a warning, got calls=%v logs=%s", pam.calls, output.String())
				}
			} else if len(pam.calls) != 1 {
				t.Fatalf("got %d PAM calls, want 1", len(pam.calls))
			} else {
				if pam.calls[0].role != tt.role || !slices.Equal(pam.calls[0].principals, strings.Split(tt.principals, ",")) {
					t.Fatalf("unexpected entitlement: %+v", pam.calls[0])
				}
				if pam.calls[0].conditionExpression != `resource.name == "projects/_/buckets/ssb-my-team-data-produkt-dev/objects/example"` {
					t.Fatalf("unexpected condition: %s", pam.calls[0].conditionExpression)
				}
			}
		})
	}
}

func TestReconcile_PAMConfigurationChangesWithoutRestart(t *testing.T) {
	cfg := googlepam.Config{
		Environments: []string{"dev"},
	}
	pam := &fakePrivilegedAccessManager{}
	r, err := googlepam.New(t.Context(), cfg, nil, googlepam.WithPrivilegedAccessManager(pam))
	if err != nil {
		t.Fatal(err)
	}
	cfg.Environments[0] = "prod"
	server := newFakeGcpTeamResourcesServer()
	server.folders["my-team/dev"] = &protoapi.GcpTeamFolder{TeamSlug: "my-team", Env: "dev", FolderId: "100"}
	client := startFakeGrpcServer(t, server)
	team := &protoapi.Team{Slug: "my-team", IsManaged: true}
	if err := r.Reconcile(t.Context(), client, team, logrus.New()); err != nil {
		t.Fatal(err)
	}
	server.config["role"] = "roles/storage.objectAdmin"
	server.config["principals"] = "group:readers@example.com"
	server.config["max_grant_duration"] = "2h"
	server.config["condition_template"] = `resource.name == "projects/_/buckets/{{.TeamSlug}}-{{.Env}}"`
	if err := r.Reconcile(t.Context(), client, team, logrus.New()); err != nil {
		t.Fatal(err)
	}
	if len(pam.calls) != 2 {
		t.Fatalf("got %d PAM calls, want 2", len(pam.calls))
	}
	first, updated := pam.calls[0], pam.calls[1]
	if first.duration != time.Hour || first.role != "roles/storage.objectViewer" ||
		!slices.Equal(first.principals, []string{"user:reader@example.com"}) ||
		first.conditionExpression != `resource.name == "projects/_/buckets/ssb-my-team-data-produkt-dev/objects/example"` {
		t.Fatalf("unexpected initial configuration: %+v", first)
	}
	if updated.duration != 2*time.Hour || updated.role != "roles/storage.objectAdmin" ||
		!slices.Equal(updated.principals, []string{"group:readers@example.com"}) ||
		updated.conditionExpression != `resource.name == "projects/_/buckets/my-team-dev"` {
		t.Fatalf("configuration changes were not applied: %+v", updated)
	}
	server.config["condition_template"] = "{{"
	if err := r.Reconcile(t.Context(), client, team, logrus.New()); err == nil || !strings.Contains(err.Error(), "parse PAM IAM condition template") {
		t.Fatalf("invalid configuration must not reuse the previous policy, got %v", err)
	}
	server.config = nil
	var output bytes.Buffer
	log := logrus.New()
	log.SetOutput(&output)
	if err := r.Reconcile(t.Context(), client, team, log); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "level=warning") {
		t.Fatal("cleared configuration must be skipped with a warning")
	}
	server.configErr = status.Error(codes.Unavailable, "configuration unavailable")
	if err := r.Reconcile(t.Context(), client, team, log); status.Code(err) != codes.Unavailable {
		t.Fatalf("configuration API failure must be propagated, got %v", err)
	}
	if len(pam.calls) != 2 || len(server.getRequests) != 2 || len(server.configRequests) != 5 {
		t.Fatal("configuration must be refreshed on each run without reusing stale policies")
	}
}
