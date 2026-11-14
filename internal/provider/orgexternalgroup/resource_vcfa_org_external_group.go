// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternalgroup

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
	_ resource.Resource                = (*vcfaOrgExternalGroupResource)(nil)
	_ resource.ResourceWithConfigure   = (*vcfaOrgExternalGroupResource)(nil)
	_ resource.ResourceWithImportState = (*vcfaOrgExternalGroupResource)(nil)
)

type vcfaOrgExternalGroupResource struct {
	tmClient *vcfa.VCDClient
}

func NewVcfaOrgExternalGroupResource() resource.Resource {
	return &vcfaOrgExternalGroupResource{}
}

func (r *vcfaOrgExternalGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_external_group"
}

func (r *vcfaOrgExternalGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *vcfaOrgExternalGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vcfaOrgExternalGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantCtx, err := helpers.GetTenantContext(r.tmClient.VCDClient, plan.OrgID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error getting Org of %s", labelOrgExternalGroup), err.Error())
		return
	}

	config := buildGroup(ctx, &plan, tenantCtx.OrgName, "", "", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.tmClient.CreateGroup(config, tenantCtx)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error creating %s %s", labelOrgExternalGroup, plan.Name.ValueString()), err.Error())
		return
	}

	mapGroupToResourceModel(group.Group, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vcfaOrgExternalGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vcfaOrgExternalGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantCtx, err := helpers.GetTenantContext(r.tmClient.VCDClient, state.OrgID.ValueString())
	if err != nil {
		if govcd.ContainsNotFound(err) { // The Org no longer exists, therefore the group is also gone
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("error getting Org of %s", labelOrgExternalGroup), err.Error())
		return
	}

	group, err := r.tmClient.GetGroupById(state.ID.ValueString(), tenantCtx)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("error reading %s %s", labelOrgExternalGroup, state.Name.ValueString()), err.Error())
		return
	}

	mapGroupToResourceModel(group.Group, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vcfaOrgExternalGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vcfaOrgExternalGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantCtx, err := helpers.GetTenantContext(r.tmClient.VCDClient, plan.OrgID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error getting Org of %s", labelOrgExternalGroup), err.Error())
		return
	}

	group, err := r.tmClient.GetGroupById(state.ID.ValueString(), tenantCtx)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error reading %s %s", labelOrgExternalGroup, state.Name.ValueString()), err.Error())
		return
	}

	// The backend requires the current name in source, which for an external group is not its name
	config := buildGroup(ctx, &plan, tenantCtx.OrgName, state.ID.ValueString(), group.Group.NameInSource, &resp.Diagnostics)
	config.Description = group.Group.Description
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := group.Update(config)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error updating %s %s", labelOrgExternalGroup, state.Name.ValueString()), err.Error())
		return
	}

	mapGroupToResourceModel(updated.Group, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vcfaOrgExternalGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vcfaOrgExternalGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantCtx, err := helpers.GetTenantContext(r.tmClient.VCDClient, state.OrgID.ValueString())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("error getting Org of %s", labelOrgExternalGroup), err.Error())
		return
	}

	group, err := r.tmClient.GetGroupById(state.ID.ValueString(), tenantCtx)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("error reading %s %s", labelOrgExternalGroup, state.Name.ValueString()), err.Error())
		return
	}

	if err := group.Delete(); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error deleting %s %s", labelOrgExternalGroup, state.Name.ValueString()), err.Error())
	}
}

// ImportState imports the group using the ID `<org name><separator><group name>`.
func (r *vcfaOrgExternalGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idSlice := strings.Split(req.ID, vcfa.ImportSeparator)
	if len(idSlice) != 2 {
		resp.Diagnostics.AddError("invalid import ID", fmt.Sprintf("expected import ID to be <org name>%s<group name>", vcfa.ImportSeparator))
		return
	}

	org, err := r.tmClient.GetTmOrgByName(idSlice[0])
	if err != nil {
		resp.Diagnostics.AddError("error getting Org", err.Error())
		return
	}

	group, err := r.tmClient.GetGroupByName(idSlice[1], &govcd.TenantContext{OrgId: org.TmOrg.ID, OrgName: org.TmOrg.Name})
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error retrieving %s", labelOrgExternalGroup), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), group.Group.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), org.TmOrg.ID)...)
}
