package atlantis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func (r *reconciler) reconcileKnativeService(ctx context.Context, name, namespace string, repoAllowList []string, image, resources *string) error {
	if r.knativeServiceTemplate == nil {
		return errors.New("missing knative template")
	}
	services := r.knServices.Services(namespace)

	templatedKnativeService, err := r.buildKnativeService(name, repoAllowList, image, resources)
	if err != nil {
		return err
	}

	ksvc, err := services.Get(ctx, name, metav1.GetOptions{})
	// Create it if it does not already exist
	if apierrors.IsNotFound(err) {
		if _, err = services.Create(ctx, templatedKnativeService, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("could not create service: %w", err)
		}
		return nil
	} else if err != nil {
		return err
	}

	// If any fields are different, ignoring fields not set in the templated spec,
	// we need to update the service.
	// We ignore unset fields because of Knative's defaulting mechanics.
	// This could cause problems if we explicitly want to unset a field, but I don't see us doing that.
	if equality.Semantic.DeepDerivative(templatedKnativeService.Spec.ConfigurationSpec, ksvc.Spec.ConfigurationSpec) {
		return nil
	}

	ksvc.Spec.ConfigurationSpec = templatedKnativeService.Spec.ConfigurationSpec
	if _, err := services.Update(ctx, ksvc, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func parseResources(resources *string) (*v1.ResourceRequirements, error) {
	if resources == nil {
		return nil, nil
	}

	rr := new(v1.ResourceRequirements)
	if err := json.Unmarshal([]byte(*resources), rr); err != nil {
		return nil, err
	}
	return rr, nil
}

func (r *reconciler) buildKnativeService(name string, repoAllowList []string, customImage, resources *string) (*knv1.Service, error) {
	image := r.config.atlantisImage
	if customImage != nil {
		image = *customImage
	}
	// Template up a new Knative Atlantis service, in case we have changed
	// the template, or the Knative service itself has changed.
	// Is this slow? maybe, but this is more flexible and readable than doing all
	// the Go structs by hand. Maybe it should be a Helm chart..
	buf := new(bytes.Buffer)
	if err := r.knativeServiceTemplate.Execute(buf, map[string]string{
		"Name":          name,
		"RepoAllowList": strings.Join(repoAllowList, ","),
		"Image":         image,
	}); err != nil {
		return nil, err
	}

	var templatedKnativeService knv1.Service
	if err := yaml.Unmarshal(buf.Bytes(), &templatedKnativeService); err != nil {
		return nil, err
	}

	resourceRequirements, err := parseResources(resources)
	if err != nil {
		return nil, err
	}
	if resourceRequirements != nil {
		templatedKnativeService.Spec.Template.Spec.GetContainer().Resources = *resourceRequirements
	}

	return &templatedKnativeService, nil
}
