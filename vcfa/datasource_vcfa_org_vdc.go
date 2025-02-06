package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaOrgVdc() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaOrgVdcRead,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Parent Organization ID",
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Parent Region ID",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of the %s", labelVcfaOrgVdc),
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaOrgVdc),
			},
			"supervisor_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("A set of Supervisor IDs that back this %s", labelVcfaOrgVdc),
			},
			"zone_resource_allocations": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        orgVdcDsZoneResourceAllocation,
				Description: "A set of Region Zones and their resource allocations",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s status", labelVcfaOrgVdc),
			},
			"region_vm_class_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("A set of %s IDs assigned to this %s", labelVcfaRegionVmClass, labelVcfaOrgVdc),
			},
		},
	}
}

var orgVdcDsZoneResourceAllocation = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"region_zone_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("%s Name", labelVcfaRegionZone),
		},
		"region_zone_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("%s ID", labelVcfaRegionZone),
		},
		"memory_limit_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Memory limit in MiB",
		},
		"memory_reservation_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Memory reservation in MiB",
		},
		"cpu_limit_mhz": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "CPU limit in MHz",
		},
		"cpu_reservation_mhz": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "CPU reservation in MHz",
		},
	},
}

func datasourceVcfaOrgVdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	getByNameAndOrgId := func(_ string) (*govcd.TmVdc, error) {
		region, err := tmClient.GetRegionById(d.Get("region_id").(string))
		if err != nil {
			return nil, err
		}
		org, err := tmClient.GetOrgById(d.Get("org_id").(string))
		if err != nil {
			return nil, err
		}
		return tmClient.GetTmVdcByName(fmt.Sprintf("%s_%s", org.Org.Name, region.Region.Name))
	}

	c := dsReadConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:   labelVcfaOrgVdc,
		getEntityFunc: getByNameAndOrgId,
		stateStoreFunc: func(tmClient *VCDClient, d *schema.ResourceData, outerType *govcd.TmVdc) error {
			err := setTmVdcData(tmClient, d, outerType)
			if err != nil {
				return err
			}
			return saveVmClassesInState(tmClient, d, outerType.TmVdc.ID)
		},
	}
	return readDatasource(ctx, d, meta, c)
}
