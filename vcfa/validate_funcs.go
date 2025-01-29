package vcfa

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// validateCase checks if a string is of caseType "upper" or "lower"
func validateCase(caseType string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		switch caseType {
		case "upper":
			if strings.ToUpper(v) != v {
				es = append(es, fmt.Errorf(
					"expected string to be upper cased, got: %s", v))
			}
		case "lower":
			if strings.ToLower(v) != v {
				es = append(es, fmt.Errorf(
					"expected string to be lower cased, got: %s", v))
			}
		default:
			panic("unsupported validation type for validateCase() function")
		}
		return
	}
}

// IsIntAndAtLeast returns a SchemaValidateFunc which tests if the provided value string is convertable to int
// and is at least min (inclusive)
func IsIntAndAtLeast(min int) schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(func(i interface{}, k string) (warnings []string, errors []error) {
		value, err := strconv.Atoi(i.(string))
		if err != nil {
			errors = append(errors, fmt.Errorf("expected type of %s to be integer", k))
			return warnings, errors
		}

		if value < min {
			errors = append(errors, fmt.Errorf("expected %s to be at least (%d), got %d", k, min, value))
			return warnings, errors
		}

		return warnings, errors
	})
}
