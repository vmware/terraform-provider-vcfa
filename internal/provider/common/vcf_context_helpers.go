// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ExtractVcfContext decodes a types.Object into a VcfContextModel, appending
// any diagnostics to the provided diags. Use this in datasource and resource
// CRUD methods to avoid repeating the three-line decode pattern.
func ExtractVcfContext(ctx context.Context, ctxObj types.Object, diags *diag.Diagnostics) VcfContextModel {
	var model VcfContextModel
	diags.Append(ctxObj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	return model
}
