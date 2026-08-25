package graph

import (
	"slices"
	"strings"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/apierror"
)

func validateTeamFeatureArgs(feature string, env string) error {
	validFeatures := []string{"ai"}
	validEnvs := []string{"prod", "test"}

	formatError := func(element string, value string, expectedValues []string) error {
		return apierror.Errorf("validateTeamFeatureArgs: Invalid value for %s %q, must be one of %q", element, value, strings.Join(expectedValues, ","))
	}

	if !slices.Contains(validFeatures, feature) {
		return formatError("feature", feature, validFeatures)
	}
	if !slices.Contains(validEnvs, env) {
		return formatError("env", env, validEnvs)
	}
	if feature == "ai" && env != "test" {
		return apierror.Errorf("validateTeamFeatureArgs: Invalid combination of values feature: %q and env: %q", feature, env)
	}
	return nil
}
