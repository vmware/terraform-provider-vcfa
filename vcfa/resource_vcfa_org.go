package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaOrg = "Organization"

func resourceVcfaOrg() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaOrgCreate,
		ReadContext:   resourceVcfaOrgRead,
		UpdateContext: resourceVcfaOrgUpdate,
		DeleteContext: resourceVcfaOrgDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The unique identifier in the full URL with which users log in to this %s", labelVcfaOrg),
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Appears in the Cloud application as a human-readable name of the %s", labelVcfaOrg),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: fmt.Sprintf("Defines if the %s enabled. Defaults to 'true'", labelVcfaOrg),
			},
			"is_classic_tenant": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true, // Cannot be changed once created
				Description: fmt.Sprintf("Defines whether the %s is a classic VRA-style tenant", labelVcfaOrg),
			},
			"managed_by_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s owner ID", labelVcfaOrg),
			},
			"managed_by_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s owner Name", labelVcfaOrg),
			},
			"org_vdc_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of VDCs belonging to the %s", labelVcfaOrg),
			},
			"catalog_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of catalog belonging to the %s", labelVcfaOrg),
			},
			"vapp_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of vApps belonging to the %s", labelVcfaOrg),
			},
			"running_vm_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of running VMs in the %s", labelVcfaOrg),
			},
			"user_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of users in the %s", labelVcfaOrg),
			},
			"disk_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of disks in the %s", labelVcfaOrg),
			},
			"can_publish": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Defines whether the %s can publish catalogs externally", labelVcfaOrg),
			},
			"directly_managed_org_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of directly managed %ss", labelVcfaOrg),
			},
		},
	}
}

func resourceVcfaOrgCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmOrg, types.TmOrg]{
		entityLabel:      labelVcfaOrg,
		getTypeFunc:      getOrgType,
		stateStoreFunc:   setOrgData,
		createFunc:       tmClient.CreateTmOrg,
		resourceReadFunc: resourceVcfaOrgRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaOrgUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmOrg, types.TmOrg]{
		entityLabel:      labelVcfaOrg,
		getTypeFunc:      getOrgType,
		getEntityFunc:    tmClient.GetTmOrgById,
		resourceReadFunc: resourceVcfaOrgRead,

		preUpdateHooks: []outerEntityHookInnerEntityType[*govcd.TmOrg, *types.TmOrg]{
			validateRenameOrgDisabled,    // 'name' can only be changed when 'is_enabled=false'
			resubmitIdAndManagedByFields, // TODO: TM: review if ID and ManagedBy should always be submitted on update
		},
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaOrgRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmOrg, types.TmOrg]{
		entityLabel:    labelVcfaOrg,
		getEntityFunc:  tmClient.GetTmOrgById,
		stateStoreFunc: setOrgData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaOrgDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	c := crudConfig[*govcd.TmOrg, types.TmOrg]{
		entityLabel:    labelVcfaOrg,
		getEntityFunc:  tmClient.GetTmOrgById,
		preDeleteHooks: []outerEntityHook[*govcd.TmOrg]{disableTmOrg}, // Org must be disabled before deletion
	}

	return deleteResource(ctx, d, meta, c)
}

// disableTmOrg disables Org which is useful before deletion as a non-disabled Org cannot be
// removed
func disableTmOrg(t *govcd.TmOrg) error {
	if t.TmOrg.IsEnabled {
		return t.Disable()
	}
	return nil
}

func resubmitIdAndManagedByFields(d *schema.ResourceData, o *govcd.TmOrg, i *types.TmOrg) error {
	// TODO: TM: review if ManagedBy should always be submitted
	i.ID = o.TmOrg.ID

	i.ManagedBy = o.TmOrg.ManagedBy // It is optional in docs, but API rejects update without it
	return nil
}

// validateRenameOrgDisabled is and update hook that checks Org can be renamed. It can be renamed if
// * it is going to be disabled with the same API call
// * if it was previously disabled and is being enabled together with new name
func validateRenameOrgDisabled(d *schema.ResourceData, oldCfg *govcd.TmOrg, newCfg *types.TmOrg) error {
	if d.HasChange("name") &&
		// this condition is a negative xor - it will be matched if Org is not transitioning from or to disabled state
		((!newCfg.IsEnabled && !oldCfg.TmOrg.IsEnabled) || newCfg.IsEnabled && oldCfg.TmOrg.IsEnabled) {
		return fmt.Errorf("%s must be disabled (is_enabled=false) to change name because it changes tenant login URL", labelVcfaOrg)
	}

	return nil
}

func resourceVcfaOrgImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	o, err := tmClient.GetTmOrgByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error getting Org: %s", err)
	}

	d.SetId(o.TmOrg.ID)
	return []*schema.ResourceData{d}, nil
}

func getOrgType(_ *VCDClient, d *schema.ResourceData) (*types.TmOrg, error) {
	t := &types.TmOrg{
		Name:            d.Get("name").(string),
		DisplayName:     d.Get("display_name").(string),
		Description:     d.Get("description").(string),
		IsEnabled:       d.Get("is_enabled").(bool),
		IsClassicTenant: d.Get("is_classic_tenant").(bool),
	}

	return t, nil
}

func setOrgData(_ *VCDClient, d *schema.ResourceData, org *govcd.TmOrg) error {
	if org == nil || org.TmOrg == nil {
		return fmt.Errorf("cannot save state for nil Org")
	}

	d.SetId(org.TmOrg.ID)
	dSet(d, "name", org.TmOrg.Name)
	dSet(d, "display_name", org.TmOrg.DisplayName)
	dSet(d, "description", org.TmOrg.Description)
	dSet(d, "is_enabled", org.TmOrg.IsEnabled)

	// Computed in resource
	var managedById string
	var managedByName string
	if org.TmOrg.ManagedBy != nil {
		managedById = org.TmOrg.ManagedBy.ID
		managedByName = org.TmOrg.ManagedBy.Name
	}
	dSet(d, "managed_by_id", managedById)
	dSet(d, "managed_by_name", managedByName)
	dSet(d, "org_vdc_count", org.TmOrg.OrgVdcCount)
	dSet(d, "catalog_count", org.TmOrg.CatalogCount)
	dSet(d, "vapp_count", org.TmOrg.VappCount)
	dSet(d, "running_vm_count", org.TmOrg.RunningVMCount)
	dSet(d, "user_count", org.TmOrg.UserCount)
	dSet(d, "disk_count", org.TmOrg.DiskCount)
	dSet(d, "can_publish", org.TmOrg.CanPublish)
	dSet(d, "directly_managed_org_count", org.TmOrg.DirectlyManagedOrgCount)
	dSet(d, "is_classic_tenant", org.TmOrg.IsClassicTenant)

	return nil
}
