// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func (d *vcfaProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Data source for reading a %s and the users and groups that have access to it", vcfatypes.LabelProject),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Internal identifier of the %s, which is its name", vcfatypes.LabelProject),
			},

			// Required lookup attributes
			"name": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", vcfatypes.LabelProject),
			},

			// Spec attributes
			"description": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Description of the %s", vcfatypes.LabelProject),
			},
			"uid": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Unique identifier of the %s", vcfatypes.LabelProject),
			},

			// Access attributes
			"members": schema.SetNestedAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Users and groups that have access to the %s, and their role", vcfatypes.LabelProject),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.StringAttribute{
							Computed:    true,
							Description: "Kind of member: `User` or `Group`",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the user or group",
						},
						"role": schema.StringAttribute{
							Computed:    true,
							Description: "Role of the member in the Project: " + strings.Join(projectMemberRoles, ","),
						},
					},
				},
			},
		},
	}
}
