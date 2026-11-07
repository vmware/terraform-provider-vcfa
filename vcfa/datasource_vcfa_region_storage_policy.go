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
				Description: fmt.Sprintf("Storage capacity in mebibytes for this %s", labelVcfaRegionStoragePolicy),
				Deprecated:  "Use `storage_capacity_mib` instead. The value is expressed in mebibytes",
			},
			"storage_consumed_mb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Consumed storage in mebibytes for this %s", labelVcfaRegionStoragePolicy),
				Deprecated:  "Use `storage_consumed_mib` instead. The value is expressed in mebibytes",
			},
			"storage_capacity_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Storage capacity in mebibytes for this %s", labelVcfaRegionStoragePolicy),
			},
			"storage_consumed_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Consumed storage in mebibytes for this %s", labelVcfaRegionStoragePolicy),
			},
			"kubernetes_compliant_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Kubernetes-compliant name of the %s", labelVcfaRegionStoragePolicy),
			},
			"is_restricted": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether the %s is restricted to system-owned namespaces", labelVcfaRegionStoragePolicy),
			},
		},
	}
}

func datasourceVcfaRegionStoragePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	region, err := tmClient.GetRegionById(d.Get("region_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving %s with ID '%s: %s", labelVcfaRegion, d.Get("region_id").(string), err)
	}

	c := dsReadConfig[*govcd.RegionStoragePolicy, types.RegionStoragePolicy]{
		entityLabel: labelVcfaRegionStoragePolicy,
		getEntityFunc: func(name string) (*govcd.RegionStoragePolicy, error) {
			return region.GetStoragePolicyByName(name)
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
	dSet(d, "storage_capacity_mb", rsp.RegionStoragePolicy.StorageCapacityMiB)
	dSet(d, "storage_consumed_mb", rsp.RegionStoragePolicy.StorageConsumedMiB)
	dSet(d, "storage_capacity_mib", rsp.RegionStoragePolicy.StorageCapacityMiB)
	dSet(d, "storage_consumed_mib", rsp.RegionStoragePolicy.StorageConsumedMiB)
	dSet(d, "kubernetes_compliant_name", rsp.RegionStoragePolicy.KubernetesCompliantName)
	dSet(d, "is_restricted", derefOr(rsp.RegionStoragePolicy.IsRestricted))
	dSet(d, "status", rsp.RegionStoragePolicy.Status)

	d.SetId(rsp.RegionStoragePolicy.ID)

	return nil
}
