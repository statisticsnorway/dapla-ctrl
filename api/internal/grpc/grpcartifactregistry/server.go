package grpcartifactregistry

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/grpc/grpcartifactregistry/grpcartifactregistrysql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/grpc/grpcpagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	querier grpcartifactregistrysql.Querier
	protoapi.UnimplementedArtifactRegistryServer
}

func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{
		querier: grpcartifactregistrysql.New(pool),
	}
}

func (t *Server) GetArtifactRegistryRepo(ctx context.Context, req *protoapi.GetArtifactRegistryRepoRequest) (*protoapi.GetArtifactRegistryRepoResponse, error) {
	repo, err := t.querier.Get(ctx, grpcartifactregistrysql.GetParams{
		TeamSlug: slug.Slug(req.TeamSlug),
		Format:   req.Format,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get repo: %s", err)
	}

	resp := &protoapi.GetArtifactRegistryRepoResponse{
		Repo: &protoapi.ArtifactRegistryRepo{
			TeamSlug: repo.TeamArtifactRegistryRepository.TeamSlug.String(),
			Format:   repo.TeamArtifactRegistryRepository.Format,
		},
	}
	return resp, nil
}

func (t *Server) ListArtifactRegistryReposForTeam(ctx context.Context, req *protoapi.ListArtifactRegistryReposForTeamRequest) (*protoapi.ListArtifactRegistryReposForTeamResponse, error) {
	limit, offset := grpcpagination.Pagination(req)
	repos, err := t.querier.List(ctx, grpcartifactregistrysql.ListParams{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list repos for team: %s", err)
	}

	total, err := t.querier.CountTeamRepos(ctx, slug.Slug(req.TeamSlug))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get teams count: %s", err)
	}

	resp := &protoapi.ListArtifactRegistryReposForTeamResponse{
		PageInfo: grpcpagination.PageInfo(req, int(total)),
		Nodes:    make([]*protoapi.ArtifactRegistryRepo, len(repos)),
	}
	for i, repo := range repos {
		resp.Nodes[i] = toProtoRepo(&repo.TeamArtifactRegistryRepository)
	}

	return resp, nil
}

func toProtoRepo(repo *grpcartifactregistrysql.TeamArtifactRegistryRepository) *protoapi.ArtifactRegistryRepo {
	r := &protoapi.ArtifactRegistryRepo{
		TeamSlug: repo.TeamSlug.String(),
		Format:   repo.Format,
	}

	return r
}
