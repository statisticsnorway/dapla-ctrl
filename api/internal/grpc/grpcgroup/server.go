package grpcgroup

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-api/internal/grpc/grpcgroup/grpcgroupsql"
	"github.com/statisticsnorway/dapla-api/internal/grpc/grpcpagination"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/protoapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	querier grpcgroupsql.Querier
	protoapi.UnimplementedGroupsServer
}

func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{
		querier: grpcgroupsql.New(pool),
	}
}

func (t *Server) Members(ctx context.Context, req *protoapi.ListGroupMembersRequest) (*protoapi.ListGroupMembersResponse, error) {
	limit, offset := grpcpagination.Pagination(req)
	users, err := t.querier.ListMembers(ctx, grpcgroupsql.ListMembersParams{
		GroupName: req.Name,
		Offset:    offset,
		Limit:     limit,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list group members: %s", err)
	}

	total, err := t.querier.CountMembers(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get group member count: %s", err)
	}

	resp := &protoapi.ListGroupMembersResponse{
		PageInfo: grpcpagination.PageInfo(req, int(total)),
		Nodes:    make([]*protoapi.GroupMember, len(users)),
	}
	for i, user := range users {
		resp.Nodes[i] = toProtoGroupMember(&user.User)
	}

	return resp, nil
}

func (t *Server) Get(ctx context.Context, req *protoapi.GetGroupRequest) (*protoapi.GetGroupResponse, error) {
	group, err := t.querier.Get(ctx, req.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "group not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get group: %s", err)
	}

	return &protoapi.GetGroupResponse{
		Group: toProtoGroup(group),
	}, nil
}

func (t *Server) SetExternalId(ctx context.Context, req *protoapi.SetExternalIdRequest) (*protoapi.SetExternalIdResponse, error) {
	_, err := t.querier.Get(ctx, req.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "group not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get group: %s", err)
	}

	if err = t.querier.UpdateExternalId(ctx, grpcgroupsql.UpdateExternalIdParams{
		Name:       req.Name,
		ExternalID: &req.ExternalId,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update external id: %s", err)
	}

	return &protoapi.SetExternalIdResponse{}, nil
}

func (t *Server) AddMember(ctx context.Context, req *protoapi.AddMemberRequest) (*protoapi.AddMemberResponse, error) {
	_, err := t.querier.Get(ctx, req.Groupname)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "group not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get group: %s", err)
	}

	user, err := t.querier.GetUserByExternalId(ctx, req.UserExternalId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "user not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %s", err)
	}

	if err = t.querier.AddMember(ctx, grpcgroupsql.AddMemberParams{
		GroupName: req.Groupname,
		UserID:    user.ID,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add member to group")
	}

	// Not the end of the world if it fails
	_ = t.querier.RefreshTeamMembers(ctx)

	return &protoapi.AddMemberResponse{}, nil
}

func (t *Server) RemoveMember(ctx context.Context, req *protoapi.RemoveMemberRequest) (*protoapi.RemoveMemberResponse, error) {
	_, err := t.querier.Get(ctx, req.Groupname)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "group not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get group: %s", err)
	}

	user, err := t.querier.GetUserByExternalId(ctx, req.UserExternalId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "user not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %s", err)
	}

	if err = t.querier.RemoveMember(ctx, grpcgroupsql.RemoveMemberParams{
		GroupName: req.Groupname,
		UserID:    user.ID,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove member from group")
	}

	// Not the end of the world if it fails
	_ = t.querier.RefreshTeamMembers(ctx)

	return &protoapi.RemoveMemberResponse{}, nil
}

func toProtoGroupMember(u *grpcgroupsql.User) *protoapi.GroupMember {
	return &protoapi.GroupMember{
		User: &protoapi.User{
			Id:         u.ID.String(),
			Name:       u.Name,
			Email:      u.Email,
			ExternalId: u.ExternalID,
		},
	}
}

func toProtoGroup(g *grpcgroupsql.GetRow) *protoapi.Group {
	return &protoapi.Group{
		TeamSlug:   g.TeamSlug.String(),
		Name:       g.Name,
		ExternalId: g.ExternalID,
	}
}
