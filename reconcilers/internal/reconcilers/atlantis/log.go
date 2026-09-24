package atlantis

import (
	"github.com/sirupsen/logrus"
	"knative.dev/pkg/kmp"
)

func LogDiff[T any](a, b T, log logrus.FieldLogger) {
	diff, err := kmp.SafeDiff(a, b)
	if err != nil {
		log.Errorf("tried to generate diff, got error: %s", err)
	} else {
		log.Infof("diff(%T): %v", a, diff)
	}
}
