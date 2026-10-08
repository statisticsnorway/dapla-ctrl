package atlantis

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
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
			AtlantisProject: projectId,
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

	t.Run("create if not exists", func(t *testing.T) {
		if err := r.reconcileKubernetesVolume(t.Context(), atlantisName, namespace, "10Gi", log); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid disk size should error", func(t *testing.T) {
		if err := r.reconcileKubernetesVolume(t.Context(), atlantisName, namespace, "ErrMe", log); err == nil {
			t.Fatal(err)
		}
	})
}
