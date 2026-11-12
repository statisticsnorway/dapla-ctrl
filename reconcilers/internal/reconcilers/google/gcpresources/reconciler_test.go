package gcpresources_test

import (
	"bytes"
	"context"
	"errors"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// fakeGcpTeamResourcesServer implements protoapi.GcpTeamResourcesServer in memory.
type fakeGcpTeamResourcesServer struct {
	protoapi.UnimplementedGcpTeamResourcesServer
	folders map[string]*protoapi.GcpTeamFolder // "teamSlug/env" -> folder
}

func newFakeGcpTeamResourcesServer() *fakeGcpTeamResourcesServer {
	return &fakeGcpTeamResourcesServer{
		folders: make(map[string]*protoapi.GcpTeamFolder),
	}
}

func (s *fakeGcpTeamResourcesServer) UpsertTeamFolder(_ context.Context, req *protoapi.UpsertGcpTeamFolderRequest) (*protoapi.UpsertGcpTeamFolderResponse, error) {
	key := req.Folder.TeamSlug + "/" + req.Folder.Env
	s.folders[key] = req.Folder
	return &protoapi.UpsertGcpTeamFolderResponse{}, nil
}

func (s *fakeGcpTeamResourcesServer) GetTeamFolder(_ context.Context, req *protoapi.GetGcpTeamFolderRequest) (*protoapi.GetGcpTeamFolderResponse, error) {
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

func TestValidateConfig(t *testing.T) {
	tests := map[string]struct {
		cfg     gcpresources.Config
		wantErr string
	}{
		"missing tag key": {
			cfg: gcpresources.Config{
				EnvParentFolders: map[string]string{"dev": "11111"},
			},
			wantErr: "tag key name is required",
		},
		"invalid tag key format": {
			cfg: gcpresources.Config{
				TagKeyNamespacedName: "123456/team",
				EnvParentFolders:     map[string]string{"dev": "11111"},
			},
			wantErr: "tag key name must be in format tagKeys/{id}",
		},
		"missing env folders": {
			cfg: gcpresources.Config{
				TagKeyNamespacedName: "tagKeys/123456",
			},
			wantErr: "at least one environment parent folder must be configured",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := gcpresources.New(context.Background(), tt.cfg, nil)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestReconcile_CreatesFoldersAndTagsThem(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	fakeGcp := gcpresources.NewFakeResourceManager()
	fakeSrv := newFakeGcpTeamResourcesServer()
	apiClient := startFakeGrpcServer(t, fakeSrv)

	cfg := gcpresources.Config{
		TagKeyNamespacedName: "tagKeys/123456",
		EnvParentFolders: map[string]string{
			"dev":  "11111",
			"test": "22222",
			"prod": "33333",
		},
	}

	r, err := gcpresources.New(ctx, cfg, nil, gcpresources.WithResourceManager(fakeGcp))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	team := &protoapi.Team{Slug: "my-team"}
	if err := r.Reconcile(ctx, apiClient, team, log); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	for _, env := range []string{"dev", "test", "prod"} {
		resp, err := fakeSrv.GetTeamFolder(ctx, &protoapi.GetGcpTeamFolderRequest{
			TeamSlug: "my-team",
			Env:      env,
		})
		if err != nil {
			t.Errorf("env %q: folder not stored: %v", env, err)
			continue
		}
		t.Logf("env %q: folder_id=%s", env, resp.Folder.FolderId)

		if tag, ok := fakeGcp.Tags[resp.Folder.FolderId]; !ok {
			t.Errorf("env %q: folder %q was not tagged", env, resp.Folder.FolderId)
		} else {
			t.Logf("env %q: tagged with %s", env, tag)
		}
	}
}

func TestReconcile_IdempotentOnSecondRun(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()

	fakeGcp := gcpresources.NewFakeResourceManager()
	fakeSrv := newFakeGcpTeamResourcesServer()
	apiClient := startFakeGrpcServer(t, fakeSrv)

	cfg := gcpresources.Config{
		TagKeyNamespacedName: "tagKeys/123456",
		EnvParentFolders:     map[string]string{"dev": "11111"},
	}

	r, _ := gcpresources.New(ctx, cfg, nil, gcpresources.WithResourceManager(fakeGcp))
	team := &protoapi.Team{Slug: "my-team"}

	if err := r.Reconcile(ctx, apiClient, team, log); err != nil {
		t.Fatalf("first Reconcile: %v", err)
	}
	if err := r.Reconcile(ctx, apiClient, team, log); err != nil {
		t.Fatalf("second Reconcile: %v", err)
	}
}

type entitlementCall struct {
	folderID   string
	role       string
	principals []string
	duration   time.Duration
}

type fakePrivilegedAccessManager struct {
	calls []entitlementCall
	err   error
}

func (f *fakePrivilegedAccessManager) EnsureEntitlement(_ context.Context, folderID, role string, principals []string, duration time.Duration) error {
	f.calls = append(f.calls, entitlementCall{folderID, role, principals, duration})
	return f.err
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
			resources := gcpresources.NewFakeResourceManager()
			pam := &fakePrivilegedAccessManager{}
			r, err := gcpresources.New(t.Context(), gcpresources.Config{
				TagKeyNamespacedName: "tagKeys/123456",
				EnvParentFolders:     map[string]string{"dev": "111", "test": "222", "prod": "333"},
				PAM: gcpresources.PAMConfig{
					Principals:       []string{" serviceAccount:reader@test.iam.gserviceaccount.com", " group:readers@example.com "},
					Role:             " roles/storage.objectViewer ",
					MaxGrantDuration: 2 * time.Hour,
				},
			}, nil, gcpresources.WithResourceManager(resources), gcpresources.WithPrivilegedAccessManager(pam))
			if err != nil {
				t.Fatal(err)
			}

			err = r.Reconcile(t.Context(), startFakeGrpcServer(t, server), &protoapi.Team{Slug: "my-team", IsManaged: managed}, logrus.New())
			if err != nil {
				t.Fatal(err)
			}
			if len(resources.Tags) != 3 {
				t.Fatalf("tagged %d folders, want 3", len(resources.Tags))
			}
			if !managed {
				if len(pam.calls) != 0 {
					t.Fatal("self-managed team must not use PAM")
				}
				return
			}
			if len(pam.calls) != 3 {
				t.Fatalf("got %d PAM calls, want 3", len(pam.calls))
			}
			folders := make(map[string]bool)
			for _, call := range pam.calls {
				folders[call.folderID] = true
				if !slices.Equal(call.principals, []string{"serviceAccount:reader@test.iam.gserviceaccount.com", "group:readers@example.com"}) ||
					call.role != "roles/storage.objectViewer" || call.duration != 2*time.Hour {
					t.Errorf("unexpected entitlement: %+v", call)
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
		principals string
		role       string
		omitRole   bool
		duration   string
		pamErr     error
		wantErr    string
		wantSkip   bool
	}{
		"missing all":         {wantSkip: true},
		"missing principals":  {role: "roles/storage.objectViewer", duration: "1h", wantSkip: true},
		"missing duration":    {principals: "serviceAccount:reader@test.iam.gserviceaccount.com", role: "roles/storage.objectViewer", wantSkip: true},
		"missing role":        {principals: "user:reader@example.com", duration: "1h", omitRole: true, wantSkip: true},
		"empty role":          {principals: "user:reader@example.com", duration: "1h", wantSkip: true},
		"whitespace role":     {principals: "user:reader@example.com", role: "  ", duration: "1h", wantSkip: true},
		"group principal":     {principals: "group:readers@example.com", role: "roles/storage.objectViewer", duration: "1h"},
		"multiple principals": {principals: "user:reader@example.com,group:readers@example.com", role: "roles/storage.objectViewer", duration: "1h"},
		"federated principal": {principals: "principal://iam.googleapis.com/locations/global/workforcePools/example/subject/reader", role: "roles/storage.objectViewer", duration: "1h"},
		"empty entry":         {principals: "user:reader@example.com, ,group:readers@example.com", duration: "1h", wantErr: "principals must not contain empty entries"},
		"trailing comma":      {principals: "user:reader@example.com,", duration: "1h", wantErr: "principals must not contain empty entries"},
		"configured role":     {principals: "user:reader@example.com", role: "roles/storage.objectViewer", duration: "1h"},
		"invalid duration":    {duration: "one hour", wantErr: "invalid duration"},
		"negative duration":   {duration: "-1h", wantErr: "must be between 30m and 168h"},
		"below minimum":       {duration: "29m59s", wantErr: "must be between 30m and 168h"},
		"above maximum":       {duration: "168h1s", wantErr: "must be between 30m and 168h"},
		"minimum":             {principals: "serviceAccount:reader@test.iam.gserviceaccount.com", role: "roles/storage.objectViewer", duration: "30m"},
		"maximum":             {principals: "serviceAccount:reader@test.iam.gserviceaccount.com", role: "roles/storage.objectViewer", duration: "168h"},
		"PAM error":           {principals: "serviceAccount:reader@test.iam.gserviceaccount.com", role: "roles/storage.objectViewer", duration: "1h", pamErr: errors.New("PAM unavailable"), wantErr: `folder "fake-folder-my-team" in env "dev": PAM unavailable`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := newFakeGcpTeamResourcesServer()
			env := map[string]string{
				"PAM_MANAGED_TEAM_FOLDER_PRINCIPALS":         tt.principals,
				"PAM_MANAGED_TEAM_FOLDER_MAX_GRANT_DURATION": tt.duration,
			}
			if !tt.omitRole {
				env["PAM_MANAGED_TEAM_FOLDER_ROLE"] = tt.role
			}
			cfg, err := reconcilerconfig.NewConfig(t.Context(), envconfig.MapLookuper(env))
			if err != nil {
				if tt.wantErr == "" || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("unexpected environment config error: %v", err)
				}
				return
			}
			resources := gcpresources.NewFakeResourceManager()
			pam := &fakePrivilegedAccessManager{err: tt.pamErr}
			r, err := gcpresources.New(t.Context(), gcpresources.Config{
				TagKeyNamespacedName: "tagKeys/123456",
				EnvParentFolders:     map[string]string{"dev": "111"},
				PAM: gcpresources.PAMConfig{
					Role:             cfg.PAMManagedTeamFolder.Role,
					Principals:       cfg.PAMManagedTeamFolder.Principals,
					MaxGrantDuration: cfg.PAMManagedTeamFolder.MaxGrantDuration,
				},
			}, nil, gcpresources.WithResourceManager(resources), gcpresources.WithPrivilegedAccessManager(pam))
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
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(resources.Tags) != 1 {
				t.Fatal("folder and tag reconciliation must continue")
			}
			if tt.wantSkip {
				if len(pam.calls) != 0 || !strings.Contains(output.String(), "level=warning") {
					t.Fatalf("PAM should be skipped with a warning, got calls=%v logs=%s", pam.calls, output.String())
				}
			} else if len(pam.calls) != 1 {
				t.Fatalf("got %d PAM calls, want 1", len(pam.calls))
			} else {
				if pam.calls[0].role != tt.role || !slices.Equal(pam.calls[0].principals, strings.Split(tt.principals, ",")) {
					t.Fatalf("unexpected entitlement: %+v", pam.calls[0])
				}
			}
		})
	}
}

func TestReconcile_PAMConfigurationIsSnapshot(t *testing.T) {
	cfg := gcpresources.Config{
		TagKeyNamespacedName: "tagKeys/123456",
		EnvParentFolders:     map[string]string{"dev": "111"},
		PAM: gcpresources.PAMConfig{
			Principals:       []string{"user:reader@example.com"},
			Role:             "roles/storage.objectViewer",
			MaxGrantDuration: time.Hour,
		},
	}
	pam := &fakePrivilegedAccessManager{}
	r, err := gcpresources.New(t.Context(), cfg, nil, gcpresources.WithResourceManager(gcpresources.NewFakeResourceManager()), gcpresources.WithPrivilegedAccessManager(pam))
	if err != nil {
		t.Fatal(err)
	}
	cfg.PAM.Role = ""
	cfg.PAM.Principals[0] = "user:other@example.com"
	cfg.PAM.MaxGrantDuration = 2 * time.Hour
	client := startFakeGrpcServer(t, newFakeGcpTeamResourcesServer())
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
	}
}
