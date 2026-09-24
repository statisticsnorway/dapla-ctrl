package atlantis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	_ "embed"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func (r *reconciler) reconcileKnativeService(ctx context.Context, name, namespace string, repoAllowList []string, config *protoapi.AtlantisConfig, log logrus.FieldLogger) error {
	if r.knativeServiceTemplate == nil {
		return errors.New("missing knative template")
	}
	services := r.knServices.Services(namespace)

	image := r.config.AtlantisImage
	if config.CustomImage != nil {
		image = *config.CustomImage
	}

	resources, err := parseResources(config.Resources)
	if err != nil {
		return fmt.Errorf("parse resources: %w", err)
	}

	templatedKnativeService, err := r.buildKnativeService(name, image, resources, repoAllowList)
	if err != nil {
		return fmt.Errorf("build knative service: %w", err)
	}

	ksvc, err := services.Get(ctx, name, metav1.GetOptions{})
	// Create it if it does not already exist
	if apierrors.IsNotFound(err) {
		if _, err = services.Create(ctx, templatedKnativeService, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("create service: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("get service: %w", err)
	}

	// If any fields are different, ignoring fields not set in the templated spec,
	// we need to update the service.
	// We ignore unset fields because of Knative's defaulting mechanics.
	// This could cause problems if we explicitly want to unset a field, but I don't see us doing that.
	if equality.Semantic.DeepDerivative(templatedKnativeService.Spec.ConfigurationSpec, ksvc.Spec.ConfigurationSpec) {
		return nil
	}

	if r.config.LogDiffs {
		LogDiff(templatedKnativeService.Spec.ConfigurationSpec, ksvc.Spec.ConfigurationSpec, log)
	}

	// Or...
	// ctx = apis.WithinCreate(ctx)
	// ksvc.Spec.ConfigurationSpec.SetDefaults(ctx)
	// The problem is that Knative sets some default fields based on the
	// config-defaults ConfigMap. Therefore we cannot directly compare our wanted
	// state and the live state as these defaults create a permadiff. We use DeepDerivative
	// to only compare those fields which we have set. Another method is using SetDefaults,
	// but this requires us to make it think it's in a Create event, or else it does not
	// set all of the required defaults. Hopefully we can find a better way of doing this later.

	ksvc.Spec.ConfigurationSpec = templatedKnativeService.Spec.ConfigurationSpec
	if _, err := services.Update(ctx, ksvc, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("update service: %w", err)
	}

	return nil
}

func parseResources(resources []byte) (*v1.ResourceRequirements, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	rr := new(v1.ResourceRequirements)
	if err := json.Unmarshal(resources, rr); err != nil {
		return nil, err
	}
	return rr, nil
}

func (r *reconciler) buildKnativeService(name, image string, resources *v1.ResourceRequirements, repoAllowList []string) (*knv1.Service, error) {
	// Template up a new Knative Atlantis service, in case we have changed
	// the template, or the Knative service itself has changed.
	// Is this slow? maybe, but this is more flexible and readable than doing all
	// the Go structs by hand. Maybe it should be a Helm chart..
	buf := new(bytes.Buffer)
	if err := r.knativeServiceTemplate.Execute(buf, map[string]string{
		"Name":          name,
		"RepoAllowList": strings.Join(repoAllowList, ","),
		"Image":         image,
		"BaseDomain":    r.config.AtlantisBaseDomain,
		"GithubAppId":   r.config.GithubAppId,
	}); err != nil {
		return nil, err
	}

	var templatedKnativeService knv1.Service
	if err := yaml.Unmarshal(buf.Bytes(), &templatedKnativeService); err != nil {
		return nil, err
	}

	if resources != nil {
		templatedKnativeService.Spec.Template.Spec.GetContainer().Resources = *resources
	}

	return &templatedKnativeService, nil
}
