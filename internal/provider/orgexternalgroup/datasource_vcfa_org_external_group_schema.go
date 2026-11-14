// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternalgroup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *vcfaOrgExternalGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Data source for reading an %s", labelOrgExternalGroup),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("ID of the %s", labelOrgExternalGroup),
			},
			"org_id": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Parent Organization ID for the %s", labelOrgExternalGroup),
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelOrgExternalGroup),
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Description of the %s", labelOrgExternalGroup),
			},
			"provider_type": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Provider type of the %s", labelOrgExternalGroup),
			},
			"role_ids": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: fmt.Sprintf("Roles of the %s", labelOrgExternalGroup),
			},
		},
	}
}
