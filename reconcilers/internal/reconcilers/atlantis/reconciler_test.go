package atlantis

import (
	"context"
	"net"
	"testing"

	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetOrGenerateWebhookSecret(t *testing.T) {
	atlantisServer := newFakeAtlantisServer()
	client := startFakeGrpcServer(t, atlantisServer)
	teamName := "blabla"
	t.Run("create secret if not exists", func(t *testing.T) {
		secret, err := getOrGenerateWebhookSecret(t.Context(), client, teamName)
		if err != nil {
			t.Fatal(err)
		}
		if stored := atlantisServer.webhookSecrets[teamName]; stored != secret {
			t.Fatalf("%q != %q", stored, secret)
		}
	})

	t.Run("existing secret is not overridden", func(t *testing.T) {
		before := atlantisServer.webhookSecrets[teamName]
		after, err := getOrGenerateWebhookSecret(t.Context(), client, teamName)
		if err != nil {
			t.Fatal(err)
		}
		if before != after {
			t.Fatalf("%q != %q", before, after)
		}
	})
}

func TestReconcileKubernetesSecret(t *testing.T) {
	fakeClient := fake.NewClientset()

	r := &reconciler{
		k8sClient: fakeClient,
	}

	teamName := "test"
	webhookSecret := "testing"
	webhookSecretAlt := "not-testing"

	t.Run("kubernetes secret created if not exists", func(t *testing.T) {
		if err := r.reconcileKubernetesSecret(t.Context(), teamName, webhookSecret); err != nil {
			t.Fatal(err)
		}

		secret, err := fakeClient.CoreV1().Secrets("default").Get(t.Context(), "atlantis-"+teamName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		storedSecret := secret.Data[webhookSecretKey]
		if string(storedSecret) != webhookSecret {
			t.Fatalf("stored %q != wanted %q, %v", storedSecret, webhookSecret, secret)
		}
	})

	t.Run("kubernetes secret overriden if webhook secret changed", func(t *testing.T) {
		// Check that it already exists
		_, err := fakeClient.CoreV1().Secrets("default").Get(t.Context(), "atlantis-"+teamName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		if err := r.reconcileKubernetesSecret(t.Context(), teamName, webhookSecretAlt); err != nil {
			t.Fatal(err)
		}

		secret, err := fakeClient.CoreV1().Secrets("default").Get(t.Context(), "atlantis-"+teamName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		storedSecret := secret.Data[webhookSecretKey]
		if string(storedSecret) != webhookSecretAlt {
			t.Fatalf("stored %q != wanted %q, %v", storedSecret, webhookSecretAlt, secret)
		}
	})

}

type fakeAtlantisServer struct {
	protoapi.UnimplementedAtlantisServer
	webhookSecrets map[string]string
}

func newFakeAtlantisServer() *fakeAtlantisServer {
	return &fakeAtlantisServer{
		webhookSecrets: make(map[string]string),
	}
}

func (s *fakeAtlantisServer) GetTeamAtlantis(ctx context.Context, req *protoapi.GetTeamAtlantisRequest) (*protoapi.GetTeamAtlantisResponse, error) {
	secret, ok := s.webhookSecrets[req.TeamSlug]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "team atlantis not found")
	}
	return &protoapi.GetTeamAtlantisResponse{
		Config: &protoapi.AtlantisConfig{
			TeamSlug:      req.TeamSlug,
			WebhookSecret: &secret,
		},
	}, nil
}

func (s *fakeAtlantisServer) SetTeamAtlantisWebhookSecret(ctx context.Context, req *protoapi.SetTeamAtlantisWebhookSecretRequest) (*protoapi.SetTeamAtlantisWebhookSecretResponse, error) {
	s.webhookSecrets[req.TeamSlug] = req.WebhookSecret
	return nil, nil
}

func startFakeGrpcServer(t *testing.T, srv *fakeAtlantisServer) *apiclient.APIClient {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	protoapi.RegisterAtlantisServer(s, srv)
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)

	client, err := apiclient.New(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("create api client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return client
}
