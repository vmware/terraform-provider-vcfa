// © Broadcom. All Rights Reserved.
// The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcfaRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaRoleRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaRole),
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The ID of the %s of the %s", labelVcfaOrg, labelVcfaRole),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s description", labelVcfaRole),
			},
			"bundle_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is read-only", labelVcfaRole),
			},
			"rights": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("Set of %ss assigned to this %s", labelVcfaRight, labelVcfaRole),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func datasourceVcfaRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcfaRoleRead(ctx, d, meta, "datasource", "read")
}
