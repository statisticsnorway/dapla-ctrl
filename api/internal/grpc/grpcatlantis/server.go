package grpcatlantis

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/grpc/grpcatlantis/grpcatlantissql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "atlantis config not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "get atlantis config: %s", err)
	}

	return &protoapi.GetTeamAtlantisResponse{Config: toProtoTeamAtlantisConfig(row.TeamAtlantisConfig)}, nil
}

func (s *Server) SetTeamAtlantisWebhookSecret(ctx context.Context, req *protoapi.SetTeamAtlantisWebhookSecretRequest) (*protoapi.SetTeamAtlantisWebhookSecretResponse, error) {
	if err := s.querier.UpsertWebhookSecret(ctx, grpcatlantissql.UpsertWebhookSecretParams{
		TeamSlug:      slug.Slug(req.TeamSlug),
		WebhookSecret: &req.WebhookSecret,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "upsert webhook secret: %s", err)
	}

	return &protoapi.SetTeamAtlantisWebhookSecretResponse{}, nil
}

func toProtoTeamAtlantisConfig(config grpcatlantissql.TeamAtlantisConfig) *protoapi.AtlantisConfig {
	return &protoapi.AtlantisConfig{
		TeamSlug:      config.TeamSlug.String(),
		WebhookSecret: config.WebhookSecret,
	}
}
