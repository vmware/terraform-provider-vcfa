// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisornamespaceclass

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

// mapSupervisorNamespaceClassToDataSourceModel populates the data source model from a class and its config, which
// map to the same attributes as in the resource.
func mapSupervisorNamespaceClassToDataSourceModel(ctx context.Context, class *vcfatypes.SupervisorNamespaceClass, config *vcfatypes.SupervisorNamespaceClassConfig, model *vcfaSupervisorNamespaceClassDataSourceModel, diags *diag.Diagnostics) {
	var m vcfaSupervisorNamespaceClassResourceModel
	mapSupervisorNamespaceClassToResourceModel(class, &m)
	mapSupervisorNamespaceClassConfigToResourceModel(ctx, config, &m, diags)

	model.ID = m.ID
	model.Name = m.Name
	model.Description = m.Description
	model.StorageClasses = m.StorageClasses
	model.VMClasses = m.VMClasses
	model.ContentSources = m.ContentSources
	model.DefaultLimits = m.DefaultLimits
	model.Zones = m.Zones
}
