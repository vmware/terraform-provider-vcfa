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

func datasourceVcfaSharedSubnet() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaSharedSubnetRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaSharedSubnet),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Region ID for this %s", labelVcfaSharedSubnet),
			},
			"backing_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID for the matching Subnet in NSX",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaSharedSubnet),
			},
			"gateway_cidr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The gateway CIDR for the %s", labelVcfaSharedSubnet),
			},
			"ip_space_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The IP Block that is automatically created for this %s", labelVcfaSharedSubnet),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaSharedSubnet),
			},
			"subnet_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Type of %s", labelVcfaSharedSubnet),
			},
			"vlan_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The VLAN ID if type is VLAN",
			},
		},
	}
}

func datasourceVcfaSharedSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	getSharedSubnetByName := func(name string) (*govcd.TmSharedSubnet, error) {
		return tmClient.GetTmSharedSubnetByNameAndRegionId(name, d.Get("region_id").(string))
	}

	c := dsReadConfig[*govcd.TmSharedSubnet, types.TmSharedSubnet]{
		entityLabel:    labelVcfaSharedSubnet,
		getEntityFunc:  getSharedSubnetByName,
		stateStoreFunc: setSharedSubnetData,
	}
	return readDatasource(ctx, d, meta, c)
}
