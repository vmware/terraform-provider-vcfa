// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaOrgRegionQuota() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaOrgRegionQuotaRead,

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
				Description: fmt.Sprintf("Description of the %s", labelVcfaOrgRegionQuota),
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaOrgRegionQuota),
			},
			"supervisor_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("A set of Supervisor IDs that back this %s", labelVcfaOrgRegionQuota),
			},
			"zone_resource_allocations": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        orgRegionQuotaDsZoneResourceAllocation,
				Description: "A set of Region Zones and their resource allocations",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s status", labelVcfaOrgRegionQuota),
			},
			"region_vm_class_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("A set of %s IDs assigned to this %s", labelVcfaRegionVmClass, labelVcfaOrgRegionQuota),
			},
			"region_storage_policy": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        orgRegionQuotaDsRegionStoragePolicy,
				Description: fmt.Sprintf("A set of %s assigned to this %s", labelVcfaRegionStoragePolicy, labelVcfaOrgRegionQuota),
			},
		},
	}
}

var orgRegionQuotaDsZoneResourceAllocation = &schema.Resource{
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

var orgRegionQuotaDsRegionStoragePolicy = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"region_storage_policy_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("The ID of the %s for this %s", labelVcfaRegionStoragePolicy, labelVcfaOrgRegionQuota),
		},
		"storage_limit_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Maximum allowed storage allocation in mebibytes",
		},
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("The ID of the %s Storage Policy", labelVcfaOrgRegionQuota),
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The name of the Storage Policy. It follows RFC 1123 Label Names to conform with Kubernetes standards",
		},
		"storage_used_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Amount of storage used in mebibytes",
		},
	},
}

func datasourceVcfaOrgRegionQuotaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	getByNameAndOrgId := func(_ string) (*govcd.RegionQuota, error) {
		region, err := tmClient.GetRegionById(d.Get("region_id").(string))
		if err != nil {
			return nil, err
		}
		org, err := tmClient.GetOrgById(d.Get("org_id").(string))
		if err != nil {
			return nil, err
		}
		return tmClient.GetRegionQuotaByName(fmt.Sprintf("%s_%s", org.Org.Name, region.Region.Name))
	}

	c := dsReadConfig[*govcd.RegionQuota, types.TmVdc]{
		entityLabel:   labelVcfaOrgRegionQuota,
		getEntityFunc: getByNameAndOrgId,
		stateStoreFunc: func(tmClient *VCDClient, d *schema.ResourceData, outerType *govcd.RegionQuota) error {
			err := setOrgRegionQuotaData(tmClient, d, outerType)
			if err != nil {
				return err
			}
			err = saveVmClassesInState(tmClient, d, outerType.TmVdc.ID)
			if err != nil {
				return err
			}
			return saveRegionStoragePoliciesInState(d, outerType, "datasource")
		},
	}
	return readDatasource(ctx, d, meta, c)
}
