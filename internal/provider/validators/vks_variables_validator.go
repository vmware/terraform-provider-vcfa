// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vksVariablesRequiredNamesValidator is a validator.Set that checks each required
// variable name appears at least once in the set's "name" attribute.
type vksVariablesRequiredNamesValidator struct {
	required []string
}

// VksClusterVariablesHaveRequiredNames returns a validator.Set that fails unless
// the set contains at least one entry whose "name" attribute matches each of the
// provided names. Use it on the cluster-level "variables" attribute to enforce
// mandatory ClusterClass variable keys (e.g. "vmClass", "storageClass").
func VksClusterVariablesHaveRequiredNames(names ...string) validator.Set {
	return vksVariablesRequiredNamesValidator{required: names}
}

func (v vksVariablesRequiredNamesValidator) Description(_ context.Context) string {
	return fmt.Sprintf("set must contain at least one entry for each required variable name: %s",
		strings.Join(v.required, ", "))
}

func (v vksVariablesRequiredNamesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v vksVariablesRequiredNamesValidator) ValidateSet(_ context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	present := make(map[string]bool, len(req.ConfigValue.Elements()))
	for _, elem := range req.ConfigValue.Elements() {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}
		nameAttr, exists := obj.Attributes()["name"]
		if !exists {
			continue
		}
		nameVal, ok := nameAttr.(types.String)
		if !ok || nameVal.IsNull() || nameVal.IsUnknown() {
			continue
		}
		present[nameVal.ValueString()] = true
	}

	for _, name := range v.required {
		if !present[name] {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Missing required cluster variable",
				fmt.Sprintf("The \"variables\" set must include an entry with name %q. "+
					"This variable is required by the ClusterClass.", name),
			)
		}
	}
}
