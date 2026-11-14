// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternaluser

import "github.com/hashicorp/terraform-plugin-framework/types"

type vcfaOrgExternalUserDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	OrgID        types.String `tfsdk:"org_id"`
	ProviderType types.String `tfsdk:"provider_type"`
	RoleIDs      types.Set    `tfsdk:"role_ids"`
	Username     types.String `tfsdk:"username"`
}
