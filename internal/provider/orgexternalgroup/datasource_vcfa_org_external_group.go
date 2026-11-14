// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternalgroup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

var (
	_ datasource.DataSource              = (*vcfaOrgExternalGroupDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*vcfaOrgExternalGroupDataSource)(nil)
)

type vcfaOrgExternalGroupDataSource struct {
	tmClient *vcfa.VCDClient
}

func NewVcfaOrgExternalGroupDataSource() datasource.DataSource {
	return &vcfaOrgExternalGroupDataSource{}
}

func (d *vcfaOrgExternalGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_external_group"
}

func (d *vcfaOrgExternalGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	tmClient, err := helpers.GetTmClientFromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving TM client from provider data", err.Error())
		return
	}
	d.tmClient = tmClient
}

func (d *vcfaOrgExternalGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vcfaOrgExternalGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantCtx, err := helpers.GetTenantContext(d.tmClient.VCDClient, data.OrgID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error getting Org of %s", labelOrgExternalGroup), err.Error())
		return
	}

	group, err := d.tmClient.GetGroupByName(data.Name.ValueString(), tenantCtx)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error reading %s %s", labelOrgExternalGroup, data.Name.ValueString()), err.Error())
		return
	}

	data.ID = types.StringValue(group.Group.ID)
	data.Name = types.StringValue(group.Group.Name)
	data.ProviderType = types.StringValue(group.Group.ProviderType)
	data.Description = types.StringValue(group.Group.Description)
	data.RoleIDs = roleIDsFromGroup(group.Group, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
