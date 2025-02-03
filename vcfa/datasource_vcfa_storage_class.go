package vcfa

import (
	"context"
	"fmt"

	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const labelVcfaStorageClass = "Storage Class"

func datasourceVcfaStorageClass() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaStorageClassRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s name", labelVcfaStorageClass),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The Region that this %s belongs to", labelVcfaStorageClass),
			},
			"storage_capacity_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("The total storage capacity of the %s in mebibytes", labelVcfaStorageClass),
			},
			"storage_consumed_mib": {
				Type:     schema.TypeInt,
				Computed: true,
				Description: fmt.Sprintf("For tenants, this represents the total storage given to all namespaces consuming from this %s in mebibytes. "+
					"For providers, this represents the total storage given to tenants from this %s in mebibytes.", labelVcfaStorageClass, labelVcfaStorageClass),
			},
			"zone_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("A set with all the IDs of the zones available to the %s", labelVcfaStorageClass),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func datasourceVcfaStorageClassRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	c := dsReadConfig[*govcd.StorageClass, types.StorageClass]{
		entityLabel: labelVcfaStorageClass,
		getEntityFunc: func(name string) (*govcd.StorageClass, error) {
			return vcdClient.GetStorageClassByName(name)
		},
		stateStoreFunc: setStorageClassData,
	}
	return readDatasource(ctx, d, meta, c)
}

func setStorageClassData(_ *VCDClient, d *schema.ResourceData, sc *govcd.StorageClass) error {
	if sc == nil || sc.StorageClass == nil {
		return fmt.Errorf("provided %s is nil", labelVcfaStorageClass)
	}

	dSet(d, "name", sc.StorageClass.Name)
	dSet(d, "storage_capacity_mib", sc.StorageClass.StorageCapacityMiB)
	dSet(d, "storage_consumed_mib", sc.StorageClass.StorageConsumedMiB)
	regionId := ""
	if sc.StorageClass.Region != nil {
		regionId = sc.StorageClass.Region.ID
	}
	dSet(d, "region_id", regionId)

	var zoneIds []string
	if len(sc.StorageClass.Zones) > 0 {
		zoneIds = extractIdsFromOpenApiReferences(sc.StorageClass.Zones)
	}
	err := d.Set("zone_ids", zoneIds)
	if err != nil {
		return err
	}

	d.SetId(sc.StorageClass.ID)

	return nil
}
