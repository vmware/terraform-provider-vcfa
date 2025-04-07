/*
 * // © Broadcom. All Rights Reserved.
 * // The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
 * // SPDX-License-Identifier: MPL-2.0
 */

package vcfa

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
