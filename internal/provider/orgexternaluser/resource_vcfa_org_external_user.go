// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternaluser

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/vmware/go-vcloud-director/v3/govcd"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

var (
	_ resource.Resource                = (*vcfaOrgExternalUserResource)(nil)
	_ resource.ResourceWithConfigure   = (*vcfaOrgExternalUserResource)(nil)
	_ resource.ResourceWithImportState = (*vcfaOrgExternalUserResource)(nil)
)

type vcfaOrgExternalUserResource struct {
	tmClient *vcfa.VCDClient
}

func NewVcfaOrgExternalUserResource() resource.Resource {
	return &vcfaOrgExternalUserResource{}
}

func (r *vcfaOrgExternalUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_external_user"
}

func (r *vcfaOrgExternalUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	tmClient, err := helpers.GetTmClientFromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving TM client from provider data", err.Error())
		return
	}
	r.tmClient = tmClient
}

func (r *vcfaOrgExternalUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vcfaOrgExternalUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantCtx, err := helpers.GetTenantContext(r.tmClient.VCDClient, plan.OrgID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error getting Org of %s", labelOrgExternalUser), err.Error())
		return
	}

	config := buildUser(ctx, &plan, tenantCtx.OrgName, "", "", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.tmClient.CreateUser(config, tenantCtx)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error creating %s %s", labelOrgExternalUser, plan.Username.ValueString()), err.Error())
		return
	}

	mapUserToResourceModel(user.User, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vcfaOrgExternalUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vcfaOrgExternalUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantCtx, err := helpers.GetTenantContext(r.tmClient.VCDClient, state.OrgID.ValueString())
	if err != nil {
		if govcd.ContainsNotFound(err) { // The Org no longer exists, therefore the user is also gone
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("error getting Org of %s", labelOrgExternalUser), err.Error())
		return
	}

	user, err := r.tmClient.GetUserById(state.ID.ValueString(), tenantCtx)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("error reading %s %s", labelOrgExternalUser, state.Username.ValueString()), err.Error())
		return
	}

	mapUserToResourceModel(user.User, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vcfaOrgExternalUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vcfaOrgExternalUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantCtx, err := helpers.GetTenantContext(r.tmClient.VCDClient, plan.OrgID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error getting Org of %s", labelOrgExternalUser), err.Error())
		return
	}

	user, err := r.tmClient.GetUserById(state.ID.ValueString(), tenantCtx)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error reading %s %s", labelOrgExternalUser, state.Username.ValueString()), err.Error())
		return
	}

	// The backend requires the current name in source, which for an external user is not its username
	config := buildUser(ctx, &plan, tenantCtx.OrgName, state.ID.ValueString(), user.User.NameInSource, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := user.Update(config)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error updating %s %s", labelOrgExternalUser, state.Username.ValueString()), err.Error())
		return
	}

	mapUserToResourceModel(updated.User, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vcfaOrgExternalUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vcfaOrgExternalUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantCtx, err := helpers.GetTenantContext(r.tmClient.VCDClient, state.OrgID.ValueString())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("error getting Org of %s", labelOrgExternalUser), err.Error())
		return
	}

	user, err := r.tmClient.GetUserById(state.ID.ValueString(), tenantCtx)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("error reading %s %s", labelOrgExternalUser, state.Username.ValueString()), err.Error())
		return
	}

	if err := user.Delete(); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error deleting %s %s", labelOrgExternalUser, state.Username.ValueString()), err.Error())
	}
}

// ImportState imports the user using the ID `<org name><separator><user name>`.
func (r *vcfaOrgExternalUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idSlice := strings.Split(req.ID, vcfa.ImportSeparator)
	if len(idSlice) != 2 {
		resp.Diagnostics.AddError("invalid import ID", fmt.Sprintf("expected import ID to be <org name>%s<user name>", vcfa.ImportSeparator))
		return
	}

	org, err := r.tmClient.GetTmOrgByName(idSlice[0])
	if err != nil {
		resp.Diagnostics.AddError("error getting Org", err.Error())
		return
	}

	user, err := r.tmClient.GetUserByName(idSlice[1], &govcd.TenantContext{OrgId: org.TmOrg.ID, OrgName: org.TmOrg.Name})
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error retrieving %s", labelOrgExternalUser), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), user.User.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), org.TmOrg.ID)...)
}
