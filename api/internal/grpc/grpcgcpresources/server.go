package grpcgcpresources

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/grpc/grpcgcpresources/grpcgcpresourcessql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	querier grpcgcpresourcessql.Querier
	protoapi.UnimplementedGcpTeamResourcesServer
}

func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{
		querier: grpcgcpresourcessql.New(pool),
	}
}

func (s *Server) UpsertTeamFolder(ctx context.Context, req *protoapi.UpsertGcpTeamFolderRequest) (*protoapi.UpsertGcpTeamFolderResponse, error) {
	if err := s.querier.UpsertTeamFolder(ctx, grpcgcpresourcessql.UpsertTeamFolderParams{
		TeamSlug: slug.Slug(req.Folder.TeamSlug),
		Env:      req.Folder.Env,
		FolderID: req.Folder.FolderId,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "upsert team folder: %s", err)
	}
	return &protoapi.UpsertGcpTeamFolderResponse{}, nil
}

func (s *Server) GetTeamFolder(ctx context.Context, req *protoapi.GetGcpTeamFolderRequest) (*protoapi.GetGcpTeamFolderResponse, error) {
	row, err := s.querier.GetTeamFolder(ctx, grpcgcpresourcessql.GetTeamFolderParams{
		TeamSlug: slug.Slug(req.TeamSlug),
		Env:      req.Env,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "team folder not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "get team folder: %s", err)
	}
	return &protoapi.GetGcpTeamFolderResponse{
		Folder: toProtoTeamFolder(&row.GcpTeamFolder),
	}, nil
}

func (s *Server) ListTeamFolders(ctx context.Context, req *protoapi.ListGcpTeamFoldersRequest) (*protoapi.ListGcpTeamFoldersResponse, error) {
	rows, err := s.querier.ListTeamFolders(ctx, slug.Slug(req.TeamSlug))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "team folders not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "get team folder: %s", err)
	}

	folders := make([]*protoapi.GcpTeamFolder, 0, len(rows))
	for _, row := range rows {
		folders = append(folders, toProtoTeamFolder(&row.GcpTeamFolder))
	}

	return &protoapi.ListGcpTeamFoldersResponse{
		Folders: folders,
	}, nil
}

func toProtoTeamFolder(f *grpcgcpresourcessql.GcpTeamFolder) *protoapi.GcpTeamFolder {
	return &protoapi.GcpTeamFolder{
		TeamSlug: string(f.TeamSlug),
		Env:      f.Env,
		FolderId: f.FolderID,
	}
}
