package atlantis

import (
	"fmt"
	"strings"
)

type FieldsValidationError map[string]string

func (e FieldsValidationError) Error() string {
	var errs []string
	for field, err := range e {
		errs = append(errs, fmt.Sprintf("%s=%q", field, err))
	}
	return fmt.Sprintf("these fields have errors: %s", strings.Join(errs, ", "))
}
