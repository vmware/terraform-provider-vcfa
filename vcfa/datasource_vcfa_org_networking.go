// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcfaOrgNetworking() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaOrgNetworkingRead,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s for %s", labelVcfaOrg, labelVcfaOrgNetworking),
			},
			"log_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("A globally unique identifier (max 8 char) for this %s in the logs of the backing network provider", labelVcfaOrg),
			},
			"networking_tenancy_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this Organization has tenancy for the network domain in the backing network provider",
			},
		},
	}
}

func datasourceVcfaOrgNetworkingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	org, err := tmClient.GetTmOrgById(d.Get("org_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving %s: %s", labelVcfaOrg, err)
	}

	d.SetId(org.TmOrg.ID)
	dSet(d, "org_id", org.TmOrg.ID)

	orgNetworkingSettings, err := org.GetOrgNetworkingSettings()
	if err != nil {
		return diag.Errorf("error retrieving %s for %s:%s", labelVcfaOrgNetworking, labelVcfaOrg, err)
	}

	err = setOrgNetworkingSettingsData(tmClient, d, orgNetworkingSettings)
	if err != nil {
		return diag.Errorf("error storing read %s: %s", labelVcfaOrgNetworking, err)
	}

	return nil
}
