// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternaluser

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	vcdtypes "github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelOrgExternalUser = "Org External User"

type vcfaOrgExternalUserResourceModel struct {
	ID           types.String `tfsdk:"id"`
	OrgID        types.String `tfsdk:"org_id"`
	ProviderType types.String `tfsdk:"provider_type"`
	RoleIDs      types.Set    `tfsdk:"role_ids"`
	Username     types.String `tfsdk:"username"`
}

// buildUser returns the user to create or update. nameInSource is the current name in source of
// an existing user, which the backend requires on update.
func buildUser(ctx context.Context, model *vcfaOrgExternalUserResourceModel, orgName, userID, nameInSource string, diags *diag.Diagnostics) *vcdtypes.OpenApiUser {
	var roleIDs []string
	diags.Append(model.RoleIDs.ElementsAs(ctx, &roleIDs, false)...)

	roles := make([]vcdtypes.OpenApiReference, 0, len(roleIDs))
	for _, id := range roleIDs {
		roles = append(roles, vcdtypes.OpenApiReference{ID: id})
	}

	locked := false
	return &vcdtypes.OpenApiUser{
		ID:             userID,
		NameInSource:   nameInSource,
		OrgEntityRef:   &vcdtypes.OpenApiReference{ID: model.OrgID.ValueString(), Name: orgName},
		Username:       model.Username.ValueString(),
		ProviderType:   model.ProviderType.ValueString(),
		RoleEntityRefs: roles,
		Locked:         &locked,
	}
}

// roleIDsFromUser returns the IDs of the roles of the user as a set.
func roleIDsFromUser(user *vcdtypes.OpenApiUser, diags *diag.Diagnostics) types.Set {
	ids := make([]string, 0, len(user.RoleEntityRefs))
	for _, ref := range user.RoleEntityRefs {
		ids = append(ids, ref.ID)
	}
	set, d := types.SetValueFrom(context.Background(), types.StringType, ids)
	diags.Append(d...)
	return set
}

func mapUserToResourceModel(user *vcdtypes.OpenApiUser, model *vcfaOrgExternalUserResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(user.ID)
	model.Username = types.StringValue(user.Username)
	model.ProviderType = types.StringValue(user.ProviderType)
	model.RoleIDs = roleIDsFromUser(user, diags)
	model.OrgID = types.StringValue("")
	if user.OrgEntityRef != nil {
		model.OrgID = types.StringValue(user.OrgEntityRef.ID)
	}
}
