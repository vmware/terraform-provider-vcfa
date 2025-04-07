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

func datasourceVcfaOrgSettings() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaOrgSettingsRead,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s for %s", labelVcfaOrg, labelVcfaOrgSettings),
			},
			"can_create_subscribed_libraries": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether the %s can create content libraries that are subscribed to external sources", labelVcfaOrg),
			},
			"quarantine_content_library_items": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether to quarantine new %ss for file inspection", labelVcfaContentLibraryItem),
			},
		},
	}
}

func datasourceVcfaOrgSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	org, err := tmClient.GetTmOrgById(d.Get("org_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving %s: %s", labelVcfaOrg, err)
	}

	d.SetId(org.TmOrg.ID)
	dSet(d, "org_id", org.TmOrg.ID)

	orgSettings, err := org.GetSettings()
	if err != nil {
		return diag.Errorf("error retrieving %s for %s:%s", labelVcfaOrgSettings, labelVcfaOrg, err)
	}

	err = setOrgSettingsData(tmClient, d, orgSettings)
	if err != nil {
		return diag.Errorf("error storing read %s: %s", labelVcfaOrgSettings, err)
	}

	return nil
}
