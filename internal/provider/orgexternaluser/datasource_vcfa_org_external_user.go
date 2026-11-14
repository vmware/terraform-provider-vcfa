// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternaluser

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

var (
	_ datasource.DataSource              = (*vcfaOrgExternalUserDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*vcfaOrgExternalUserDataSource)(nil)
)

type vcfaOrgExternalUserDataSource struct {
	tmClient *vcfa.VCDClient
}

func NewVcfaOrgExternalUserDataSource() datasource.DataSource {
	return &vcfaOrgExternalUserDataSource{}
}

func (d *vcfaOrgExternalUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_external_user"
}

func (d *vcfaOrgExternalUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vcfaOrgExternalUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vcfaOrgExternalUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantCtx, err := helpers.GetTenantContext(d.tmClient.VCDClient, data.OrgID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error getting Org of %s", labelOrgExternalUser), err.Error())
		return
	}

	user, err := d.tmClient.GetUserByName(data.Username.ValueString(), tenantCtx)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error reading %s %s", labelOrgExternalUser, data.Username.ValueString()), err.Error())
		return
	}

	data.ID = types.StringValue(user.User.ID)
	data.Username = types.StringValue(user.User.Username)
	data.ProviderType = types.StringValue(user.User.ProviderType)
	data.RoleIDs = roleIDsFromUser(user.User, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
