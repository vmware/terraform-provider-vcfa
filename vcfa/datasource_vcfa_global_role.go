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

func datasourceVcfaGlobalRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceGlobalRoleRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaGlobalRole),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s description", labelVcfaGlobalRole),
			},
			"bundle_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is read-only", labelVcfaGlobalRole),
			},
			"rights": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("List of %ss assigned to this %s", labelVcfaRight, labelVcfaGlobalRole),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"publish_to_all_orgs": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("When true, publishes the %s to all %ss", labelVcfaGlobalRole, labelVcfaOrg),
			},
			"org_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("List of IDs of %ss to which this %s is published", labelVcfaOrg, labelVcfaGlobalRole),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func datasourceGlobalRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericGlobalRoleRead(ctx, d, meta, "datasource", "read")
}
