package grpcsharedbucketsstopgap

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/grpc/grpcsharedbucketsstopgap/grpcsharedbucketsstopgapsql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	querier grpcsharedbucketsstopgapsql.Querier
	protoapi.UnimplementedSharedBucketsStopgapServer
}

func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{
		querier: grpcsharedbucketsstopgapsql.New(pool),
	}
}

func (s *Server) Create(ctx context.Context, req *protoapi.CreateSharedBucketsStopgapRequest) (*protoapi.CreateSharedBucketsStopgapResponse, error) {
	if err := s.querier.Create(ctx, grpcsharedbucketsstopgapsql.CreateParams{
		Name:      bucketName(req.SharedBucketStopgap),
		ShortName: req.SharedBucketStopgap.ShortName,
		TeamSlug:  slug.Slug(req.SharedBucketStopgap.TeamSlug),
		Env:       req.SharedBucketStopgap.Env,
		Kind:      req.SharedBucketStopgap.Type,
	}); err != nil {
		return nil, err
	}

	return &protoapi.CreateSharedBucketsStopgapResponse{}, nil
}

func (s *Server) Get(ctx context.Context, req *protoapi.GetSharedBucketsStopgapRequest) (*protoapi.GetSharedBucketsStopgapResponse, error) {
	res, err := s.querier.Get(ctx, grpcsharedbucketsstopgapsql.GetParams{
		ShortName: req.SharedBucketStopgap.ShortName,
		TeamSlug:  slug.Slug(req.SharedBucketStopgap.TeamSlug),
		Kind:      req.SharedBucketStopgap.Type,
		Env:       req.SharedBucketStopgap.Env,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "bucket not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get bucket: %s", err)
	}
	return &protoapi.GetSharedBucketsStopgapResponse{
		SharedBucketStopgap: toProtoSharedBucketStopgap(&res.SharedBucketsStopgap),
	}, nil
}

func (s *Server) Groups(ctx context.Context, req *protoapi.ListSharedBucketsStopgapGroupsRequest) (*protoapi.ListSharedBucketsStopgapGroupsResponse, error) {
	res, err := s.querier.ListGroups(ctx, bucketName(req.SharedBucketStopgap))
	if err != nil {
		return nil, err
	}
	resp := &protoapi.ListSharedBucketsStopgapGroupsResponse{
		Nodes: make([]*protoapi.Group, len(res)),
	}
	for i, group := range res {
		resp.Nodes[i] = toProtoGroup(&group.Group)
	}
	return resp, nil
}

func (s *Server) AddGroup(ctx context.Context, req *protoapi.AddSharedBucketsStopgapGroupRequest) (*protoapi.AddSharedBucketsStopgapGroupResponse, error) {
	if err := s.querier.AddGroup(ctx, grpcsharedbucketsstopgapsql.AddGroupParams{
		GroupName:  req.Group,
		BucketName: bucketName(req.SharedBucketStopgap),
	}); errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "bucket not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get bucket: %s", err)
	}

	return &protoapi.AddSharedBucketsStopgapGroupResponse{}, nil
}

func (s *Server) RemoveGroup(ctx context.Context, req *protoapi.RemoveSharedBucketsStopgapGroupRequest) (*protoapi.RemoveSharedBucketsStopgapGroupResponse, error) {
	if err := s.querier.RemoveGroup(ctx, grpcsharedbucketsstopgapsql.RemoveGroupParams{
		GroupName:  req.Group,
		BucketName: bucketName(req.SharedBucketStopgap),
	}); errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "bucket not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get bucket: %s", err)
	}

	return &protoapi.RemoveSharedBucketsStopgapGroupResponse{}, nil
}

func toProtoSharedBucketStopgap(from *grpcsharedbucketsstopgapsql.SharedBucketsStopgap) *protoapi.SharedBucketStopgap {
	return &protoapi.SharedBucketStopgap{
		ShortName: from.ShortName,
		TeamSlug:  string(from.TeamSlug),
		Type:      from.Kind,
		Env:       from.Env,
	}
}

func toProtoGroup(from *grpcsharedbucketsstopgapsql.Group) *protoapi.Group {
	return &protoapi.Group{
		TeamSlug:   from.TeamSlug.String(),
		Name:       from.Name,
		ExternalId: from.ExternalID,
	}
}

func bucketName(sbs *protoapi.SharedBucketStopgap) string {
	return fmt.Sprintf("ssb-%s-data-delt-%s-%s", sbs.TeamSlug, sbs.ShortName, sbs.Env)
}
