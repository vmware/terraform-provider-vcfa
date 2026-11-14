// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternalgroup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var providerTypes = []string{"LDAP", "OAUTH", "SAML"}

func (r *vcfaOrgExternalGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Resource for managing an %s, a group sourced from an external identity provider such as LDAP, OAUTH or SAML", labelOrgExternalGroup),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("ID of the %s", labelOrgExternalGroup),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Parent Organization ID for the %s", labelOrgExternalGroup),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider_type": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Provider type of the %s: `LDAP`, `OAUTH` or `SAML`", labelOrgExternalGroup),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(providerTypes...),
				},
			},
			"role_ids": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: fmt.Sprintf("Roles to use for the %s", labelOrgExternalGroup),
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Name of the %s, as known to the external identity provider", labelOrgExternalGroup),
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Description of the %s", labelOrgExternalGroup),
			},
		},
	}
}
