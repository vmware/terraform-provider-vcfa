package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaRegionZone = "Region Zone"

func datasourceVcfaRegionZone() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceVcfaRegionZoneRead,

		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent Region ID for %s", labelVcfaRegionZone),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaRegionZone),
			},
			"memory_limit_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Memory limit in MiB",
			},
			"memory_reservation_used_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Memory reservation in MiB",
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
			"cpu_reservation_used_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU reservation in MHz",
			},
		},
	}
}

func resourceVcfaRegionZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).VcfaClient
	region, err := vcfaClient.GetRegionById(d.Get("region_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving %s: %s", labelVcfaRegion, err)
	}
	getRegionZone := func(name string) (*govcd.Zone, error) {
		return region.GetZoneByName(name)
	}

	c := dsReadConfig[*govcd.Zone, types.Zone]{
		entityLabel:    labelVcfaRegionZone,
		getEntityFunc:  getRegionZone,
		stateStoreFunc: setZoneData,
	}
	return readDatasource(ctx, d, meta, c)
}

func setZoneData(_ *VCDClient, d *schema.ResourceData, z *govcd.Zone) error {
	if z == nil || z.Zone == nil {
		return fmt.Errorf("nil %s", labelVcfaRegionZone)
	}
	d.SetId(z.Zone.ID)
	dSet(d, "memory_limit_mib", z.Zone.MemoryLimitMiB)
	dSet(d, "memory_reservation_used_mib", z.Zone.MemoryReservationUsedMiB)
	dSet(d, "memory_reservation_mib", z.Zone.MemoryReservationMiB)
	dSet(d, "cpu_limit_mhz", z.Zone.CPULimitMhz)
	dSet(d, "cpu_reservation_mhz", z.Zone.CPUReservationMhz)
	dSet(d, "cpu_reservation_used_mhz", z.Zone.CPUReservationUsedMhz)

	return nil
}
