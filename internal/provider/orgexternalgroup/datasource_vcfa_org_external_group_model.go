// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternalgroup

import "github.com/hashicorp/terraform-plugin-framework/types"

type vcfaOrgExternalGroupDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	OrgID        types.String `tfsdk:"org_id"`
	ProviderType types.String `tfsdk:"provider_type"`
	RoleIDs      types.Set    `tfsdk:"role_ids"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
}
