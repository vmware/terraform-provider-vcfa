// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package project

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func mapProjectToDataSourceModel(ctx context.Context, project vcfatypes.Project, bindings []vcfatypes.ProjectRoleBinding, model *vcfaProjectDataSourceModel, diags *diag.Diagnostics) {
	model.Name = types.StringValue(project.Name)
	model.Description = types.StringValue(project.Spec.Description)
	model.UID = types.StringValue(string(project.UID))
	model.Members = mapBindingsToMembers(ctx, bindings, nil, diags)
}
