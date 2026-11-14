// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternalgroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	vcdtypes "github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelOrgExternalGroup = "Org External Group"

type vcfaOrgExternalGroupResourceModel struct {
	ID           types.String `tfsdk:"id"`
	OrgID        types.String `tfsdk:"org_id"`
	ProviderType types.String `tfsdk:"provider_type"`
	RoleIDs      types.Set    `tfsdk:"role_ids"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
}

// buildGroup returns the group to create or update. nameInSource is the current name in source of
// an existing group, which the backend requires on update.
func buildGroup(ctx context.Context, model *vcfaOrgExternalGroupResourceModel, orgName, groupID, nameInSource string, diags *diag.Diagnostics) *vcdtypes.OpenApiGroup {
	var roleIDs []string
	diags.Append(model.RoleIDs.ElementsAs(ctx, &roleIDs, false)...)

	roles := make([]vcdtypes.OpenApiReference, 0, len(roleIDs))
	for _, id := range roleIDs {
		roles = append(roles, vcdtypes.OpenApiReference{ID: id})
	}

	return &vcdtypes.OpenApiGroup{
		ID:             groupID,
		NameInSource:   nameInSource,
		OrgEntityRef:   &vcdtypes.OpenApiReference{ID: model.OrgID.ValueString(), Name: orgName},
		Name:           model.Name.ValueString(),
		ProviderType:   model.ProviderType.ValueString(),
		RoleEntityRefs: roles,
	}
}

// roleIDsFromGroup returns the IDs of the roles of the group as a set.
func roleIDsFromGroup(group *vcdtypes.OpenApiGroup, diags *diag.Diagnostics) types.Set {
	ids := make([]string, 0, len(group.RoleEntityRefs))
	for _, ref := range group.RoleEntityRefs {
		ids = append(ids, ref.ID)
	}
	set, d := types.SetValueFrom(context.Background(), types.StringType, ids)
	diags.Append(d...)
	return set
}

func mapGroupToResourceModel(group *vcdtypes.OpenApiGroup, model *vcfaOrgExternalGroupResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(group.ID)
	model.Name = types.StringValue(group.Name)
	model.ProviderType = types.StringValue(group.ProviderType)
	model.Description = types.StringValue(group.Description)
	model.RoleIDs = roleIDsFromGroup(group, diags)
	model.OrgID = types.StringValue("")
	if group.OrgEntityRef != nil {
		model.OrgID = types.StringValue(group.OrgEntityRef.ID)
	}
}
