package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaOrgSettings = "Organization Settings"

func resourceVcfaOrgSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaOrgSettingsCreateUpdate,
		ReadContext:   resourceVcfaOrgSettingsRead,
		UpdateContext: resourceVcfaOrgSettingsCreateUpdate,
		DeleteContext: resourceVcfaOrgSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgSettingsImport, // The same as importing the Org
		},

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s for %s", labelVcfaOrg, labelVcfaOrgSettings),
			},
			"can_create_subscribed_libraries": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: fmt.Sprintf("Whether the %s can create content libraries that are subscribed to external sources", labelVcfaOrg),
			},
			"quarantine_content_library_items": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: fmt.Sprintf("Whether to quarantine new %ss for file inspection", labelVcfaContentLibraryItem),
			},
		},
	}
}

func resourceVcfaOrgSettingsCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Lock the Organization to prevent side effects
	orgId := d.Get("org_id").(string)
	vcfa.kvLock(orgId)
	defer vcfa.kvUnlock(orgId)

	tmClient := meta.(ClientContainer).tmClient

	org, err := tmClient.GetTmOrgById(orgId)
	if err != nil {
		return diag.Errorf("error retrieving %s '%s' : %s", labelVcfaOrg, orgId, err)
	}

	d.SetId(org.TmOrg.ID)
	dSet(d, "org_id", org.TmOrg.ID)

	orgNetworkingSettings, err := getOrgSettingsType(tmClient, d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = org.UpdateSettings(orgNetworkingSettings)
	if err != nil {
		return diag.Errorf("error updating %s for %s:%s", labelVcfaOrgSettings, labelVcfaOrg, err)
	}

	return resourceVcfaOrgSettingsRead(ctx, d, meta)
}

func resourceVcfaOrgSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	org, err := tmClient.GetTmOrgById(d.Get("org_id").(string))
	if err != nil {
		if govcd.ContainsNotFound(err) { // Org no longer present, removing from state
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving %s: %s", labelVcfaOrg, err)
	}

	d.SetId(org.TmOrg.ID)
	dSet(d, "org_id", org.TmOrg.ID)

	orgSettings, err := org.GetSettings()
	if err != nil {
		return diag.Errorf("error retrieving %s for %s: %s", labelVcfaOrgSettings, labelVcfaOrg, err)
	}

	err = setOrgSettingsData(tmClient, d, orgSettings)
	if err != nil {
		return diag.Errorf("error storing read %s: %s", labelVcfaOrgSettings, err)
	}

	return nil
}

func resourceVcfaOrgSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Lock the Organization to prevent side effects
	orgId := d.Get("org_id").(string)
	vcfa.kvLock(orgId)
	defer vcfa.kvUnlock(orgId)

	tmClient := meta.(ClientContainer).tmClient
	org, err := tmClient.GetTmOrgById(orgId)
	if err != nil {
		return diag.Errorf("error retrieving %s '%s': %s", labelVcfaOrg, orgId, err)
	}

	// reset settings
	resetSettings := &types.TmOrgSettings{
		CanCreateSubscribedLibraries:  addrOf(false),
		QuarantineContentLibraryItems: addrOf(false),
	}

	_, err = org.UpdateSettings(resetSettings)
	if err != nil {
		return diag.Errorf("error removing Org Settings: %s", err)
	}

	return nil
}

func resourceVcfaOrgSettingsImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	o, err := tmClient.GetTmOrgByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error getting Org: %s", err)
	}

	dSet(d, "org_id", o.TmOrg.ID)
	d.SetId(o.TmOrg.ID)
	return []*schema.ResourceData{d}, nil
}

func getOrgSettingsType(_ *VCDClient, d *schema.ResourceData) (*types.TmOrgSettings, error) {
	t := &types.TmOrgSettings{
		CanCreateSubscribedLibraries:  addrOf(d.Get("can_create_subscribed_libraries").(bool)),
		QuarantineContentLibraryItems: addrOf(d.Get("quarantine_content_library_items").(bool)),
	}

	return t, nil
}

func setOrgSettingsData(_ *VCDClient, d *schema.ResourceData, orgSettings *types.TmOrgSettings) error {
	if orgSettings == nil {
		return fmt.Errorf("organization settings cannot be nil")
	}
	if orgSettings.CanCreateSubscribedLibraries != nil {
		dSet(d, "can_create_subscribed_libraries", *orgSettings.CanCreateSubscribedLibraries)
	}
	if orgSettings.QuarantineContentLibraryItems != nil {
		dSet(d, "quarantine_content_library_items", *orgSettings.QuarantineContentLibraryItems)
	}
	return nil
}
