package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaRegionVmClass = "Region VM Class"

func datasourceVcfaRegionVmClass() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaRegionVmClassRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The name of the %s", labelVcfaRegionVmClass),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The ID of the %s that owns the %s", labelVcfaRegion, labelVcfaRegionVmClass),
			},
			"cpu_reservation_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("CPU that a Virtual Machine reserves when this %s is applied", labelVcfaRegionVmClass),
			},
			"memory_reservation_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Memory in MiB that a Virtual Machine reserves when this %s is applied", labelVcfaRegionVmClass),
			},
			"cpu_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of CPUs that a Virtual Machine gets when this %s is applied", labelVcfaRegionVmClass),
			},
			"memory_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Memory in MiB that a Virtual Machine gets when this %s is applied", labelVcfaRegionVmClass),
			},
			"reserved": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s can be used to reserve number of its instances within a namespace", labelVcfaRegionVmClass),
			},
		},
	}
}

func datasourceVcfaRegionVmClassRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := dsReadConfig[*govcd.RegionVirtualMachineClass, types.RegionVirtualMachineClass]{
		entityLabel: labelVcfaRegionVmClass,
		getEntityFunc: func(name string) (*govcd.RegionVirtualMachineClass, error) {
			return tmClient.GetRegionVirtualMachineClassByNameAndRegionId(name, d.Get("region_id").(string))
		},
		stateStoreFunc: setVmClassData,
	}
	return readDatasource(ctx, d, meta, c)
}

func setVmClassData(_ *VCDClient, d *schema.ResourceData, o *govcd.RegionVirtualMachineClass) error {
	if o == nil || o.RegionVirtualMachineClass == nil {
		return fmt.Errorf("VM Class cannot be nil")
	}
	d.SetId(o.RegionVirtualMachineClass.ID)
	dSet(d, "name", o.RegionVirtualMachineClass.Name)
	region := ""
	if o.RegionVirtualMachineClass.Region != nil {
		region = o.RegionVirtualMachineClass.Region.ID
	}
	dSet(d, "region_id", region)
	dSet(d, "cpu_reservation_mhz", o.RegionVirtualMachineClass.CpuReservationMHz)
	dSet(d, "memory_reservation_mib", o.RegionVirtualMachineClass.MemoryReservationMiB)
	dSet(d, "cpu_count", o.RegionVirtualMachineClass.CpuCount)
	dSet(d, "memory_mib", o.RegionVirtualMachineClass.MemoryMiB)
	dSet(d, "reserved", o.RegionVirtualMachineClass.Reserved)

	return nil
}
