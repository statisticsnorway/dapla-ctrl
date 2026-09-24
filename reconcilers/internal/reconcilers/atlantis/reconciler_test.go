package atlantis

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestReconcileKubernetesWebhookSecret(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	fakeClient := fake.NewClientset()

	r := &reconciler{
		k8sClient: fakeClient,
	}

	teamName := "test"
	atlantisName := "atlantis-" + teamName
	namespace := "default"
	webhookSecret := "testing"
	webhookSecretNew := "not-testing"

	t.Run("kubernetes secret created if not exists", func(t *testing.T) {
		if err := r.reconcileKubernetesWebhookSecret(t.Context(), atlantisName, namespace, webhookSecret, log); err != nil {
			t.Fatal(err)
		}

		secret, err := fakeClient.CoreV1().Secrets(namespace).Get(t.Context(), atlantisName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		wantedData := map[string][]byte{
			webhookSecretKey: []byte(webhookSecret),
		}

		if diff := cmp.Diff(wantedData, secret.Data); diff != "" {
			t.Errorf("secret data differs from wanted:\n %s", diff)
		}
	})

	t.Run("kubernetes secret overriden if webhook secret changed", func(t *testing.T) {
		// Check that it already exists
		_, err := fakeClient.CoreV1().Secrets(namespace).Get(t.Context(), atlantisName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		if err := r.reconcileKubernetesWebhookSecret(t.Context(), atlantisName, namespace, webhookSecretNew, log); err != nil {
			t.Fatal(err)
		}

		secret, err := fakeClient.CoreV1().Secrets(namespace).Get(t.Context(), atlantisName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		wantedData := map[string][]byte{
			webhookSecretKey: []byte(webhookSecretNew),
		}

		if diff := cmp.Diff(wantedData, secret.Data); diff != "" {
			t.Errorf("secret data differs from wanted:\n %s", diff)
		}
	})
}

func TestReconcileKubernetesReposConfig(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	fakeClient := fake.NewClientset()

	r := &reconciler{
		k8sClient: fakeClient,
	}

	teamName := "test"
	atlantisName := "atlantis-" + teamName
	namespace := "default"
	reposConfig := "testing"
	reposConfigNew := "not-testing"

	t.Run("kubernetes repos configmap created if not exists", func(t *testing.T) {
		if err := r.reconcileKubernetesReposConfig(t.Context(), atlantisName, namespace, reposConfig, log); err != nil {
			t.Fatal(err)
		}

		cm, err := fakeClient.CoreV1().ConfigMaps(namespace).Get(t.Context(), atlantisName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		wantedData := map[string]string{
			reposYamlKey: reposConfig,
		}

		if diff := cmp.Diff(wantedData, cm.Data); diff != "" {
			t.Errorf("configmap data differs from wanted:\n %s", diff)
		}
	})

	t.Run("kubernetes repos configmap overriden if repos.yaml changed", func(t *testing.T) {
		// Check that it already exists
		_, err := fakeClient.CoreV1().ConfigMaps(namespace).Get(t.Context(), atlantisName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		if err := r.reconcileKubernetesReposConfig(t.Context(), atlantisName, namespace, reposConfigNew, log); err != nil {
			t.Fatal(err)
		}

		cm, err := fakeClient.CoreV1().ConfigMaps(namespace).Get(t.Context(), atlantisName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		wantedData := map[string]string{
			reposYamlKey: reposConfigNew,
		}

		if diff := cmp.Diff(wantedData, cm.Data); diff != "" {
			t.Errorf("configmap data differs from wanted:\n %s", diff)
		}
	})
}

func TestReconcileKubernetesServiceAccount(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	fakeClient := fake.NewClientset()

	teamName := "test"
	atlantisName := "atlantis-" + teamName
	namespace := "default"

	projectId := "atlantis-test"

	r := &reconciler{
		k8sClient: fakeClient,
		config: reconcilerConfig{
			atlantisProject: projectId,
		},
	}

	wantedAnnotations := map[string]string{
		wiAnnotationKey: fmt.Sprintf("%s@%s.iam.gserviceaccount.com", atlantisName, projectId),
	}

	t.Run("create if not exists", func(t *testing.T) {
		if err := r.reconcileKubernetesServiceAccount(t.Context(), atlantisName, namespace, log); err != nil {
			t.Fatal(err)
		}

		sa, err := fakeClient.CoreV1().ServiceAccounts(namespace).Get(t.Context(), atlantisName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		if annotationDiff := cmp.Diff(sa.Annotations, wantedAnnotations); annotationDiff != "" {
			t.Errorf("annotations does not match: %s", annotationDiff)
		}
	})

	t.Run("do nothing if webhook secret hasn't changed", func(t *testing.T) {
		if err := r.reconcileKubernetesServiceAccount(t.Context(), atlantisName, namespace, log); err != nil {
			t.Fatal(err)
		}
		if slices.ContainsFunc(fakeClient.Actions(), func(a ktesting.Action) bool {
			return strings.EqualFold(a.GetVerb(), "update")
		}) {
			t.Fatal("update called")
		}
	})

	t.Run("annotations are overwritten if they differ from wanted", func(t *testing.T) {
		// Check that it already exists
		sa, err := fakeClient.CoreV1().ServiceAccounts(namespace).Get(t.Context(), atlantisName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		sa.Annotations = map[string]string{
			"lorem": "ipsum",
			"dolor": "sit",
		}

		_, err = fakeClient.CoreV1().ServiceAccounts(namespace).Update(t.Context(), sa, v1.UpdateOptions{})
		if err != nil {
			t.Fatal(err)
		}

		if err := r.reconcileKubernetesServiceAccount(t.Context(), atlantisName, namespace, log); err != nil {
			t.Fatal(err)
		}

		sa, err = fakeClient.CoreV1().ServiceAccounts(namespace).Get(t.Context(), atlantisName, v1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		if annotationDiff := cmp.Diff(sa.Annotations, wantedAnnotations); annotationDiff != "" {
			t.Errorf("annotations does not match: %s", annotationDiff)
		}
	})
}

func TestReconcileKubernetesVolume(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	fakeClient := fake.NewClientset()

	r := &reconciler{
		k8sClient: fakeClient,
	}

	teamName := "test"
	atlantisName := "atlantis-" + teamName
	namespace := "default"

	diskSizeDefault := resource.MustParse("10Gi")

	t.Run("create if not exists", func(t *testing.T) {
		if err := r.reconcileKubernetesVolume(t.Context(), atlantisName, namespace, diskSizeDefault, log); err != nil {
			t.Fatal(err)
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
