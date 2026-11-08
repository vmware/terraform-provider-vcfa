// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaSupervisor() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaSupervisorRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Supervisor",
			},
			"vcenter_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Parent vCenter ID",
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Parent Region ID",
			},
			"nsx_manager_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the NSX Manager associated with the Supervisor",
			},
			"load_balancer_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of load balancer used by the Supervisor",
			},
			"is_dedicated_for_classic_tenants": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the Supervisor is dedicated for classic tenants",
			},
		},
	}
}

func datasourceVcfaSupervisorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	s, err := tmClient.GetSupervisorByNameAndVcenterId(d.Get("name").(string), d.Get("vcenter_id").(string))
	if err != nil {
		return diag.Errorf("error getting Supervisor: %s", err)
	}

	err = setSupervisorData(d, s.Supervisor)
	if err != nil {
		return diag.Errorf("error storing Supervisor data: %s", err)
	}

	d.SetId(s.Supervisor.SupervisorID)

	return nil
}

func setSupervisorData(d *schema.ResourceData, s *types.Supervisor) error {
	vCenterId := ""
	if s.VirtualCenter != nil {
		vCenterId = s.VirtualCenter.ID
	}
	dSet(d, "vcenter_id", vCenterId)

	regionId := ""
	if s.Region != nil {
		regionId = s.Region.ID
	}

	dSet(d, "region_id", regionId)

	nsxManagerId := ""
	if s.NsxManager != nil {
		nsxManagerId = s.NsxManager.ID
	}
	dSet(d, "nsx_manager_id", nsxManagerId)
	dSet(d, "load_balancer_type", s.LoadBalancerType)
	dSet(d, "is_dedicated_for_classic_tenants", s.IsDedicatedForClassicTenants)

	return nil
}
