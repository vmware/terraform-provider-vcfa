// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisornamespaceclass

import "github.com/hashicorp/terraform-plugin-framework/types"

type vcfaSupervisorNamespaceClassDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	StorageClasses types.Set    `tfsdk:"storage_classes"`
	VMClasses      types.Set    `tfsdk:"vm_classes"`
	ContentSources types.Set    `tfsdk:"content_sources"`
	DefaultLimits  types.Object `tfsdk:"default_limits"`
	Zones          types.Set    `tfsdk:"zones"`
}
