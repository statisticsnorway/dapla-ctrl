package grpcatlantis

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/grpc/grpcatlantis/grpcatlantissql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
)

type Server struct {
	querier grpcatlantissql.Querier
	protoapi.UnimplementedAtlantisServer
}

func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{
		querier: grpcatlantissql.New(pool),
	}
}

func (s *Server) GetTeamAtlantis(ctx context.Context, req *protoapi.GetTeamAtlantisRequest) (*protoapi.GetTeamAtlantisResponse, error) {
	row, err := s.querier.Get(ctx, slug.Slug(req.TeamSlug))
	if err != nil {
		return nil, err
	}

	return &protoapi.GetTeamAtlantisResponse{Config: toProtoTeamAtlantisConfig(row.TeamAtlantisConfig)}, nil
}

func toProtoTeamAtlantisConfig(config grpcatlantissql.TeamAtlantisConfig) *protoapi.AtlantisConfig {
	return &protoapi.AtlantisConfig{
		TeamSlug:      config.TeamSlug.String(),
		WebhookSecret: config.WebhookSecret,
	}
}
