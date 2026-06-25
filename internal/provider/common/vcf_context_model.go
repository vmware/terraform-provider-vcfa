// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// VcfContextModel holds the VCF Automation project and namespace that scope
// a Kubernetes resource lookup. Resources embed this as a types.Object field
// tagged `tfsdk:"context"`.
type VcfContextModel struct {
	Project   types.String `tfsdk:"project"`
	Namespace types.String `tfsdk:"namespace"`
}

// VcfContextAttrTypes is the attr.Type map that corresponds to VcfContextModel.
// Use this when constructing types.ObjectType or calling helpers.ObjFrom for
// the context object.
var VcfContextAttrTypes = map[string]attr.Type{
	"project":   types.StringType,
	"namespace": types.StringType,
}
