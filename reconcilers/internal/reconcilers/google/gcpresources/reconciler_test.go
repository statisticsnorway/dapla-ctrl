package gcpresources_test

import (
	"context"
	"net"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
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
type FakeResourceManager struct {
	TagValues map[string]string // tagKey+teamSlug -> tagValueName
	Folders   map[string]string // "parent/displayName" -> folderID
	Tags      map[string]string // folderID -> tagValueName
	Deleted   []string          // deleted folderIDs
}
func NewFakeResourceManager() *FakeResourceManager {
	return &FakeResourceManager{
		TagValues: make(map[string]string),
		Folders:   make(map[string]string),
		Tags:      make(map[string]string),
	}
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

func TestReconcile_CreatesFoldersAndTagsThem(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	fakeGcp := gcpresources.NewFakeResourceManager()
	fakeSrv := newFakeGcpTeamResourcesServer()
	apiClient := startFakeGrpcServer(t, fakeSrv)

	cfg := gcpresources.Config{
		TagKeyNamespacedName: "321123/team",
		EnvParentFolders: map[string]string{
			"dev":  "11111",
			"test": "22222",
			"prod": "33333",
		},
	}

	r, err := gcpresources.New(ctx, cfg, gcpresources.WithResourceManager(fakeGcp))
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
		TagKeyNamespacedName: "321123/team",
		EnvParentFolders:     map[string]string{"dev": "11111"},
	}

	r, _ := gcpresources.New(ctx, cfg, gcpresources.WithResourceManager(fakeGcp))
	team := &protoapi.Team{Slug: "my-team"}

	if err := r.Reconcile(ctx, apiClient, team, log); err != nil {
		t.Fatalf("first Reconcile: %v", err)
	}
	if err := r.Reconcile(ctx, apiClient, team, log); err != nil {
		t.Fatalf("second Reconcile: %v", err)
	}
}
