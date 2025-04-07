// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"
	"fmt"

	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const labelVcfaRegionStoragePolicy = "Region Storage Policy"

func datasourceVcfaRegionStoragePolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaRegionStoragePolicyRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s name", labelVcfaRegionStoragePolicy),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The Region that this %s belongs to", labelVcfaRegionStoragePolicy),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of the %s", labelVcfaRegionStoragePolicy),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The creation status of the %s. Can be [NOT_READY, READY]", labelVcfaRegionStoragePolicy),
			},
			"storage_capacity_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Storage capacity in megabytes for this %s", labelVcfaRegionStoragePolicy),
			},
			"storage_consumed_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Consumed storage in megabytes for this %s", labelVcfaRegionStoragePolicy),
			},
		},
	}
}

func datasourceVcfaRegionStoragePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := dsReadConfig[*govcd.RegionStoragePolicy, types.RegionStoragePolicy]{
		entityLabel: labelVcfaRegionStoragePolicy,
		getEntityFunc: func(name string) (*govcd.RegionStoragePolicy, error) {
			return tmClient.GetRegionStoragePolicyByName(name)
		},
		stateStoreFunc: setRegionStoragePolicyData,
	}
	return readDatasource(ctx, d, meta, c)
}

func setRegionStoragePolicyData(_ *VCDClient, d *schema.ResourceData, rsp *govcd.RegionStoragePolicy) error {
	if rsp == nil || rsp.RegionStoragePolicy == nil {
		return fmt.Errorf("provided %s is nil", labelVcfaRegionStoragePolicy)
	}

	dSet(d, "name", rsp.RegionStoragePolicy.Name)
	dSet(d, "description", rsp.RegionStoragePolicy.Description)
	regionId := ""
	if rsp.RegionStoragePolicy.Region != nil {
		regionId = rsp.RegionStoragePolicy.Region.ID
	}
	dSet(d, "region_id", regionId)
	dSet(d, "storage_capacity_mb", rsp.RegionStoragePolicy.StorageCapacityMB)
	dSet(d, "storage_consumed_mb", rsp.RegionStoragePolicy.StorageConsumedMB)
	dSet(d, "status", rsp.RegionStoragePolicy.Status)

	d.SetId(rsp.RegionStoragePolicy.ID)

	return nil
}
