package grpcteam

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/grpc/grpcpagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/grpc/grpcteam/grpcteamsql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	querier grpcteamsql.Querier
	protoapi.UnimplementedTeamsServer
}

func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{
		querier: grpcteamsql.New(pool),
	}
}

func (t *Server) Delete(ctx context.Context, req *protoapi.DeleteTeamRequest) (*protoapi.DeleteTeamResponse, error) {
	if req.Slug == "" {
		return nil, status.Errorf(codes.InvalidArgument, "slug is required")
	}

	if err := t.querier.Delete(ctx, slug.Slug(req.Slug)); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to delete team: %q", req.Slug)
	}

	return &protoapi.DeleteTeamResponse{}, nil
}

func (t *Server) Get(ctx context.Context, req *protoapi.GetTeamRequest) (*protoapi.GetTeamResponse, error) {
	team, err := t.querier.Get(ctx, slug.Slug(req.Slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "team not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get team")
	}

	return &protoapi.GetTeamResponse{
		Team: toProtoTeam(team),
	}, nil
}

func (t *Server) List(ctx context.Context, req *protoapi.ListTeamsRequest) (*protoapi.ListTeamsResponse, error) {
	limit, offset := grpcpagination.Pagination(req)
	teams, err := t.querier.List(ctx, grpcteamsql.ListParams{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list teams: %s", err)
	}

	total, err := t.querier.Count(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get teams count: %s", err)
	}

	resp := &protoapi.ListTeamsResponse{
		PageInfo: grpcpagination.PageInfo(req, int(total)),
		Nodes:    make([]*protoapi.Team, len(teams)),
	}
	for i, team := range teams {
		resp.Nodes[i] = toProtoTeam(team)
	}

	return resp, nil
}

func (t *Server) Groups(ctx context.Context, req *protoapi.ListTeamGroupsRequest) (*protoapi.ListTeamGroupsResponse, error) {
	limit, offset := grpcpagination.Pagination(req)
	groups, err := t.querier.ListGroups(ctx, grpcteamsql.ListGroupsParams{
		TeamSlug: slug.Slug(req.Slug),
		Offset:   offset,
		Limit:    limit,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list team groups: %s", err)
	}

	total, err := t.querier.CountGroups(ctx, slug.Slug(req.Slug))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get team group count: %s", err)
	}

	resp := &protoapi.ListTeamGroupsResponse{
		PageInfo: grpcpagination.PageInfo(req, int(total)),
		Nodes:    make([]*protoapi.TeamGroup, len(groups)),
	}
	for i, group := range groups {
		resp.Nodes[i] = toProtoTeamGroup(group)
	}

	return resp, nil
}

func toProtoTeam(team *grpcteamsql.Team) *protoapi.Team {
	t := &protoapi.Team{
		Slug:             team.Slug.String(),
		HasManualEditing: team.HasManualEditing,
	}

	if team.DeleteKeyConfirmedAt.Valid {
		t.DeleteKeyConfirmedAt = timestamppb.New(team.DeleteKeyConfirmedAt.Time)
	}

	return t
}

func toProtoTeamGroup(group *grpcteamsql.ListGroupsRow) *protoapi.TeamGroup {
	return &protoapi.TeamGroup{
		Group: &protoapi.Group{
			TeamSlug:   group.TeamSlug.String(),
			Name:       group.Name,
			ExternalId: group.ExternalID,
		},
	}
}
