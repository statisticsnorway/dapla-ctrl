package atlantis

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmp"
	"knative.dev/serving/pkg/client/clientset/versioned/fake"
)

const testTemplate = `
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: {{ .Name }}
spec:
  template:
    spec:
      containers:
      - env:
        - name: ATLANTIS_REPO_ALLOWLIST
          value: {{ .RepoAllowList }}
        - name: ATLANTIS_ATLANTIS_URL
          value: https://{{ .Name }}.{{ .BaseDomain }}
        image: {{ .Image }}
`

func TestReconcileKnativeService(t *testing.T) {

	teamName := "test"
	atlantisName := "atlantis-" + teamName
	namespace := "default"
	repoAllowList := []string{"my-repo", "my-repo-2"}

	fakeServing := fake.NewSimpleClientset().ServingV1()

	tpl, err := template.New("").Parse(testTemplate)
	if err != nil {
		t.Fatal(err)
	}

	r := &reconciler{
		knServices: fakeServing,

		atlantisImage:      "atlantis:v0",
		atlantisBaseDomain: "ssb.no",

		knativeServiceTemplate: tpl,
	}

	t.Run("create if not exists", func(t *testing.T) {
		if err := r.reconcileKnativeService(t.Context(), atlantisName, namespace, repoAllowList); err != nil {
			t.Fatal(err)
		}

		knsvc, err := fakeServing.Services(namespace).Get(t.Context(), atlantisName, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		container := knsvc.Spec.Template.Spec.Containers[0]

		if container.Image != r.atlantisImage {
			t.Errorf("incorrect image %q, expected %q", container.Image, r.atlantisImage)
		}

		if !slices.ContainsFunc(container.Env, func(e corev1.EnvVar) bool {
			if e.Name != "ATLANTIS_REPO_ALLOWLIST" {
				return false
			}
			if expected := strings.Join(repoAllowList, ","); expected != e.Value {
				t.Errorf("incorrect repo allowlist, expected %q, got %q", expected, e.Value)
			}
			return true
		}) {
			t.Errorf("missing repo allowlist in env vars: %v", container.Env)
		}

		if !slices.ContainsFunc(container.Env, func(e corev1.EnvVar) bool {
			if e.Name != "ATLANTIS_ATLANTIS_URL" {
				return false
			}
			if expected := fmt.Sprintf("https://%s.%s", atlantisName, r.atlantisBaseDomain); expected != e.Value {
				t.Errorf("incorrect atlantis url, expected %q, got %q", expected, e.Value)
			}
			return true
		}) {
			t.Errorf("missing repo allowlist in env vars: %v", container.Env)
		}

	})

	t.Run("image updated on config change", func(t *testing.T) {
		r.atlantisImage = "atlantis:v1"

		if err := r.reconcileKnativeService(t.Context(), atlantisName, namespace, repoAllowList); err != nil {
			t.Fatal(err)
		}

		knsvc, err := fakeServing.Services(namespace).Get(t.Context(), atlantisName, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		container := knsvc.Spec.Template.Spec.Containers[0]

		if container.Image != r.atlantisImage {
			t.Errorf("incorrect image %q, expected %q", container.Image, r.atlantisImage)
		}
	})

	t.Run("service is changed back if live version has been altered", func(t *testing.T) {
		knsvc, err := fakeServing.Services(namespace).Get(t.Context(), atlantisName, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		original := knsvc.DeepCopy()

		container := &(knsvc.Spec.Template.Spec.Containers[0])
		container.Image = "atlantis:my-dev-version"
		container.Env = nil

		if _, err := fakeServing.Services(namespace).Update(t.Context(), knsvc, metav1.UpdateOptions{}); err != nil {
			t.Fatal(err)
		}

		if err := r.reconcileKnativeService(t.Context(), atlantisName, namespace, repoAllowList); err != nil {
			t.Fatal(err)
		}

		knsvc, err = fakeServing.Services(namespace).Get(t.Context(), atlantisName, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		if diff, err := kmp.SafeDiff(knsvc.Spec.ConfigurationSpec, original.Spec.ConfigurationSpec); err != nil {
			t.Fatal(err)
		} else if diff != "" {
			t.Fatalf("diff between original and reconciled altered version: %v", diff)
		}

	})

	t.Run("won't run if missing template", func(t *testing.T) {
		r.knativeServiceTemplate = nil

		if err := r.reconcileKnativeService(t.Context(), atlantisName, namespace, repoAllowList); err == nil {
			t.Fatal("knative reconciler ran with nil template")
		} else if !strings.Contains(err.Error(), "missing knative template") {
			t.Fatalf("unknown error occured: %s", err)
		}
	})
}
