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

	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	reconcilerconfig "github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/cmd/reconciler/config"
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
	folders     map[string]*protoapi.GcpTeamFolder // "teamSlug/env" -> folder
	getErr      error
	getRequests []*protoapi.GetGcpTeamFolderRequest
	upsertCalls int
}

func newFakeGcpTeamResourcesServer() *fakeGcpTeamResourcesServer {
	return &fakeGcpTeamResourcesServer{
		folders: make(map[string]*protoapi.GcpTeamFolder),
	}
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
				Environments:      []string{"dev"},
				Principals:        []string{"user:reader@example.com"},
				Role:              "roles/storage.objectViewer",
				MaxGrantDuration:  time.Hour,
				ConditionTemplate: testPAMConditionTemplate,
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
				Environments:      []string{"dev"},
				Principals:        []string{"user:reader@example.com"},
				Role:              "roles/storage.objectViewer",
				MaxGrantDuration:  time.Hour,
				ConditionTemplate: testPAMConditionTemplate,
			}
			if !configured {
				cfg.Role = ""
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
			r, err := googlepam.New(t.Context(), googlepam.Config{
				Environments:      []string{"dev", "test", "prod"},
				Principals:        []string{" serviceAccount:reader@test.iam.gserviceaccount.com", " group:readers@example.com "},
				Role:              " roles/storage.objectViewer ",
				MaxGrantDuration:  2 * time.Hour,
				ConditionTemplate: testPAMConditionTemplate,
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
				if len(pam.calls) != 0 || len(server.getRequests) != 0 {
					t.Fatal("self-managed team must not use PAM or query folders")
				}
				return
			}
			if len(pam.calls) != 3 {
				t.Fatalf("got %d PAM calls, want 3", len(pam.calls))
			}
			if len(server.getRequests) != 3 {
				t.Fatalf("got %d folder requests, want 3", len(server.getRequests))
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

func TestReconcile_PAMConfigurationFromEnvironment(t *testing.T) {
	tests := map[string]struct {
		principals        string
		role              string
		omitRole          bool
		omitCondition     bool
		conditionTemplate string
		duration          string
		pamErr            error
		wantErr           string
		wantSkip          bool
	}{
		"missing all":               {wantSkip: true},
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
		"PAM error":                 {principals: "serviceAccount:reader@test.iam.gserviceaccount.com", role: "roles/storage.objectViewer", duration: "1h", pamErr: errors.New("PAM unavailable"), wantErr: `folder "100" in env "dev": PAM unavailable`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := newFakeGcpTeamResourcesServer()
			server.folders["my-team/dev"] = &protoapi.GcpTeamFolder{TeamSlug: "my-team", Env: "dev", FolderId: "100"}
			env := map[string]string{
				"PAM_MANAGED_TEAM_FOLDER_PRINCIPALS":         tt.principals,
				"PAM_MANAGED_TEAM_FOLDER_MAX_GRANT_DURATION": tt.duration,
			}
			if !tt.omitRole {
				env["PAM_MANAGED_TEAM_FOLDER_ROLE"] = tt.role
			}
			if !tt.omitCondition {
				conditionTemplate := tt.conditionTemplate
				if conditionTemplate == "" {
					conditionTemplate = testPAMConditionTemplate
				}
				env["PAM_MANAGED_TEAM_FOLDER_CONDITION_TEMPLATE"] = conditionTemplate
			}
			cfg, err := reconcilerconfig.NewConfig(t.Context(), envconfig.MapLookuper(env))
			if err != nil {
				if tt.wantErr == "" || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("unexpected environment config error: %v", err)
				}
				return
			}
			pam := &fakePrivilegedAccessManager{err: tt.pamErr}
			r, err := googlepam.New(t.Context(), googlepam.Config{
				Environments:      []string{"dev"},
				Role:              cfg.PAMManagedTeamFolder.Role,
				Principals:        cfg.PAMManagedTeamFolder.Principals,
				MaxGrantDuration:  cfg.PAMManagedTeamFolder.MaxGrantDuration,
				ConditionTemplate: cfg.PAMManagedTeamFolder.ConditionTemplate,
			}, nil, googlepam.WithPrivilegedAccessManager(pam))
			if err != nil {
				if tt.wantErr == "" || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("unexpected reconciler config error: %v", err)
				}
				return
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
					t.Fatal("invalid condition must not reach PAM")
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

func TestReconcile_PAMConfigurationIsSnapshot(t *testing.T) {
	cfg := googlepam.Config{
		Environments:      []string{"dev"},
		Principals:        []string{"user:reader@example.com"},
		Role:              "roles/storage.objectViewer",
		MaxGrantDuration:  time.Hour,
		ConditionTemplate: testPAMConditionTemplate,
	}
	pam := &fakePrivilegedAccessManager{}
	r, err := googlepam.New(t.Context(), cfg, nil, googlepam.WithPrivilegedAccessManager(pam))
	if err != nil {
		t.Fatal(err)
	}
	cfg.Role = ""
	cfg.Principals[0] = "user:other@example.com"
	cfg.Environments[0] = "prod"
	cfg.MaxGrantDuration = 2 * time.Hour
	cfg.ConditionTemplate = "false"
	server := newFakeGcpTeamResourcesServer()
	server.folders["my-team/dev"] = &protoapi.GcpTeamFolder{TeamSlug: "my-team", Env: "dev", FolderId: "100"}
	client := startFakeGrpcServer(t, server)
	team := &protoapi.Team{Slug: "my-team", IsManaged: true}
	for range 2 {
		if err := r.Reconcile(t.Context(), client, team, logrus.New()); err != nil {
			t.Fatal(err)
		}
	}
	if len(pam.calls) != 2 {
		t.Fatalf("got %d PAM calls, want 2", len(pam.calls))
	}
	for _, call := range pam.calls {
		if call.duration != time.Hour || call.role != "roles/storage.objectViewer" ||
			!slices.Equal(call.principals, []string{"user:reader@example.com"}) {
			t.Fatalf("startup config was changed: %+v", call)
		}
		if call.conditionExpression != `resource.name == "projects/_/buckets/ssb-my-team-data-produkt-dev/objects/example"` {
			t.Fatalf("startup condition template was changed: %s", call.conditionExpression)
		}
	}
}
