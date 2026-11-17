// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package project

import "github.com/hashicorp/terraform-plugin-framework/types"

type vcfaProjectDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	UID         types.String `tfsdk:"uid"`
	Members     types.Set    `tfsdk:"members"`
}
