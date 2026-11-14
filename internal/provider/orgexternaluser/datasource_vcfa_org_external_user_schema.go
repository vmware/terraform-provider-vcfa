// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternaluser

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *vcfaOrgExternalUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Data source for reading an %s", labelOrgExternalUser),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("ID of the %s", labelOrgExternalUser),
			},
			"org_id": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Parent Organization ID for the %s", labelOrgExternalUser),
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Username of the %s", labelOrgExternalUser),
			},
			"provider_type": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Provider type of the %s", labelOrgExternalUser),
			},
			"role_ids": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: fmt.Sprintf("Roles of the %s", labelOrgExternalUser),
			},
		},
	}
}
