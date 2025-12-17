package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/internal/grpc/grpcgroup"
	"github.com/statisticsnorway/dapla-api/internal/grpc/grpcreconciler"
	"github.com/statisticsnorway/dapla-api/internal/grpc/grpcteam"
	"github.com/statisticsnorway/dapla-api/internal/grpc/grpcuser"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/protoapi"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/alts"
)

type GrpcConfig struct {
	ListenAddress           string
	ExpectedServiceAccounts []string
	WithInsecureAuth        bool
}

func Run(ctx context.Context, cfg *GrpcConfig, pool *pgxpool.Pool, log logrus.FieldLogger) error {
	log.Info("GRPC serving on ", cfg.ListenAddress)
	lis, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	opts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}

	if !cfg.WithInsecureAuth {
		altsTC := alts.NewServerCreds(alts.DefaultServerOptions())
		opts = append(opts, grpc.Creds(altsTC))

		interceptor := grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			err := alts.ClientAuthorizationCheck(ctx, cfg.ExpectedServiceAccounts)
			if err != nil {
				return nil, err
			}

			return handler(ctx, req)
		})
		opts = append(opts, interceptor)
	}

	s := grpc.NewServer(opts...)

	protoapi.RegisterGroupsServer(s, grpcgroup.NewServer(pool))
	protoapi.RegisterTeamsServer(s, grpcteam.NewServer(pool))
	protoapi.RegisterUsersServer(s, grpcuser.NewServer(pool))
	protoapi.RegisterReconcilersServer(s, grpcreconciler.NewServer(pool))

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.Serve(lis) })
	g.Go(func() error {
		<-ctx.Done()

		ch := make(chan struct{})
		go func() {
			s.GracefulStop()
			close(ch)
		}()

		select {
		case <-ch:
			// ok
		case <-time.After(5 * time.Second):
			// force shutdown
			s.Stop()
		}

		return nil
	})

	return g.Wait()
}
