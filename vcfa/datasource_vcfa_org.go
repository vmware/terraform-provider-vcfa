package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaOrg() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The unique identifier in the full URL with which users log in to this %s", labelVcfaOrg),
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Appears in the Cloud application as a human-readable name of the %s", labelVcfaOrg),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Defines if the %s enabled", labelVcfaOrg),
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
			"is_classic_tenant": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Defines whether the %s is a classic VRA-style tenant", labelVcfaOrg),
			},
		},
	}
}

func datasourceVcfaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	c := dsReadConfig[*govcd.TmOrg, types.TmOrg]{
		entityLabel:    labelVcfaOrg,
		getEntityFunc:  vcdClient.GetTmOrgByName,
		stateStoreFunc: setOrgData,
	}
	return readDatasource(ctx, d, meta, c)
}
