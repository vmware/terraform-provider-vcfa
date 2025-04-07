// © Broadcom. All Rights Reserved.
// The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaOrgNetworking = "Organization Networking Settings"

func resourceVcfaOrgNetworking() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaOrgNetworkingCreateUpdate,
		ReadContext:   resourceVcfaOrgNetworkingRead,
		UpdateContext: resourceVcfaOrgNetworkingCreateUpdate,
		DeleteContext: resourceVcfaOrgNetworkingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgNetworkingImport, // The same as importing the Org
		},

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s for %s", labelVcfaOrg, labelVcfaOrgNetworking),
			},
			"log_name": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      fmt.Sprintf("A globally unique identifier (max 8 char) for this %s in the logs of the backing network provider", labelVcfaOrg),
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 8)),
			},
			"networking_tenancy_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s has tenancy for the network domain in the backing network provider", labelVcfaOrg),
			},
		},
	}
}

func resourceVcfaOrgNetworkingCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	org, err := tmClient.GetTmOrgById(d.Get("org_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving %s: %s", labelVcfaOrg, err)
	}

	d.SetId(org.TmOrg.ID)
	dSet(d, "org_id", org.TmOrg.ID)

	orgNetworkingSettings, err := getOrgNetworkingSettingsType(tmClient, d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = org.UpdateOrgNetworkingSettings(orgNetworkingSettings)
	if err != nil {
		return diag.Errorf("error updating %s for %s:%s", labelVcfaOrgNetworking, labelVcfaOrg, err)
	}

	return resourceVcfaOrgNetworkingRead(ctx, d, meta)
}

func resourceVcfaOrgNetworkingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	org, err := tmClient.GetTmOrgById(d.Get("org_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving %s: %s", labelVcfaOrg, err)
	}

	d.SetId(org.TmOrg.ID)
	dSet(d, "org_id", org.TmOrg.ID)

	orgNetworkingSettings, err := org.GetOrgNetworkingSettings()
	if err != nil {
		if govcd.ContainsNotFound(err) { // Org no longer present, removing from state
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving %s for %s:%s", labelVcfaOrgNetworking, labelVcfaOrg, err)
	}

	err = setOrgNetworkingSettingsData(tmClient, d, orgNetworkingSettings)
	if err != nil {
		return diag.Errorf("error storing read %s: %s", labelVcfaOrgNetworking, err)
	}

	return nil
}

func resourceVcfaOrgNetworkingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	org, err := tmClient.GetTmOrgById(d.Get("org_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving %s: %s", labelVcfaOrg, err)
	}

	// reset settings
	resetSettings := &types.TmOrgNetworkingSettings{
		OrgNameForLogs: "",
	}

	_, err = org.UpdateOrgNetworkingSettings(resetSettings)
	if err != nil {
		return diag.Errorf("error removing Org Network Settings: %s", err)
	}

	return nil
}

func resourceVcfaOrgNetworkingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	o, err := tmClient.GetTmOrgByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error getting Org: %s", err)
	}

	dSet(d, "org_id", o.TmOrg.ID)
	d.SetId(o.TmOrg.ID)
	return []*schema.ResourceData{d}, nil
}

func getOrgNetworkingSettingsType(_ *VCDClient, d *schema.ResourceData) (*types.TmOrgNetworkingSettings, error) {
	t := &types.TmOrgNetworkingSettings{
		OrgNameForLogs: d.Get("log_name").(string),
		// No setting for it in UI
		// NetworkingTenancyEnabled: addrOf(d.Get("networking_tenancy_enabled").(bool)),
	}

	return t, nil
}

func setOrgNetworkingSettingsData(_ *VCDClient, d *schema.ResourceData, orgNetConfig *types.TmOrgNetworkingSettings) error {
	dSet(d, "log_name", orgNetConfig.OrgNameForLogs)
	if orgNetConfig.NetworkingTenancyEnabled != nil {
		dSet(d, "networking_tenancy_enabled", *orgNetConfig.NetworkingTenancyEnabled)
	}
	return nil
}
