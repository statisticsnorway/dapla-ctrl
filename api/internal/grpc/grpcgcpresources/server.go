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
		Folder: &protoapi.GcpTeamFolder{
			TeamSlug: string(row.TeamSlug),
			Env:      row.Env,
			FolderId: row.FolderID,
		},
	}, nil
}

func (s *Server) DeleteTeamFolders(ctx context.Context, req *protoapi.DeleteGcpTeamFoldersRequest) (*protoapi.DeleteGcpTeamFoldersResponse, error) {
	if err := s.querier.DeleteTeamFolders(ctx, slug.Slug(req.TeamSlug)); err != nil {
		return nil, status.Errorf(codes.Internal, "delete team folders: %s", err)
	}
	return &protoapi.DeleteGcpTeamFoldersResponse{}, nil
}
