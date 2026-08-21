package ai

import (
	"context"
	"net"
	"slices"
	"strings"
	"testing"

	"cloud.google.com/go/iam/apiv1/iampb"
	"cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type fakeGoogleServer struct {
	resourcemanagerpb.UnimplementedProjectsServer
	monitoringpb.UnimplementedNotificationChannelServiceServer

	projects           []*resourcemanagerpb.Project
	policy             *iampb.Policy
	setPolicyCalls     int
	channels           []*monitoringpb.NotificationChannel
	createChannelCalls int
	deleteChannelCalls int
}

func (s *fakeGoogleServer) SearchProjects(context.Context, *resourcemanagerpb.SearchProjectsRequest) (*resourcemanagerpb.SearchProjectsResponse, error) {
	return &resourcemanagerpb.SearchProjectsResponse{Projects: s.projects}, nil
}

func (s *fakeGoogleServer) GetIamPolicy(context.Context, *iampb.GetIamPolicyRequest) (*iampb.Policy, error) {
	return proto.Clone(s.policy).(*iampb.Policy), nil
}

func (s *fakeGoogleServer) SetIamPolicy(_ context.Context, req *iampb.SetIamPolicyRequest) (*iampb.Policy, error) {
	s.setPolicyCalls++
	s.policy = proto.Clone(req.Policy).(*iampb.Policy)
	return req.Policy, nil
}

func (s *fakeGoogleServer) ListNotificationChannels(_ context.Context, req *monitoringpb.ListNotificationChannelsRequest) (*monitoringpb.ListNotificationChannelsResponse, error) {
	channels := make([]*monitoringpb.NotificationChannel, 0, len(s.channels))
	for _, channel := range s.channels {
		if strings.Contains(req.Filter, `display_name = "`) && !strings.Contains(req.Filter, `display_name = "`+channel.DisplayName+`"`) {
			continue
		}
		if strings.Contains(req.Filter, `type = "`) && !strings.Contains(req.Filter, `type = "`+channel.Type+`"`) {
			continue
		}
		if strings.Contains(req.Filter, "labels.email_address") && !strings.Contains(req.Filter, `labels.email_address = "`+channel.Labels[aiBudgetNotificationLabel]+`"`) {
			continue
		}
		channels = append(channels, proto.Clone(channel).(*monitoringpb.NotificationChannel))
	}
	return &monitoringpb.ListNotificationChannelsResponse{NotificationChannels: channels}, nil
}

func (s *fakeGoogleServer) CreateNotificationChannel(_ context.Context, req *monitoringpb.CreateNotificationChannelRequest) (*monitoringpb.NotificationChannel, error) {
	s.createChannelCalls++
	channel := proto.Clone(req.NotificationChannel).(*monitoringpb.NotificationChannel)
	channel.Name = req.Name + "/notificationChannels/" + channel.Labels[aiBudgetNotificationLabel]
	s.channels = append(s.channels, channel)
	return channel, nil
}

func (s *fakeGoogleServer) DeleteNotificationChannel(_ context.Context, req *monitoringpb.DeleteNotificationChannelRequest) (*emptypb.Empty, error) {
	s.deleteChannelCalls++
	s.channels = slices.DeleteFunc(s.channels, func(channel *monitoringpb.NotificationChannel) bool {
		return channel.Name == req.Name
	})
	return &emptypb.Empty{}, nil
}

func fakeGoogleClients(t *testing.T, server *fakeGoogleServer) (*resourcemanager.ProjectsClient, *monitoring.NotificationChannelClient) {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	resourcemanagerpb.RegisterProjectsServer(grpcServer, server)
	monitoringpb.RegisterNotificationChannelServiceServer(grpcServer, server)
	go func() { _ = grpcServer.Serve(listener) }()
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	ctx := context.Background()
	projects, err := resourcemanager.NewProjectsClient(ctx, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatal(err)
	}
	channels, err := monitoring.NewNotificationChannelClient(ctx, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatal(err)
	}
	return projects, channels
}
