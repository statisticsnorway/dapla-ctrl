package atlantis

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

}
