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

func datasourceVcfaDistributedVlanConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaDistributedVlanConnectionRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaDistributedVlanConnection),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Region ID for this %s", labelVcfaDistributedVlanConnection),
			},
			"backing_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("ID for the matching %s in NSX", labelVcfaDistributedVlanConnection),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaDistributedVlanConnection),
			},
			"gateway_cidr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The gateway CIDR for the %s", labelVcfaDistributedVlanConnection),
			},
			"ip_space_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Reference to an IP Block that is used for the external connection for this %s", labelVcfaDistributedVlanConnection),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaDistributedVlanConnection),
			},
			"subnet_exclusive": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is exclusively for the gateway CIDR only", labelVcfaDistributedVlanConnection),
			},
			"vlan_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The VLAN ID for the external traffic",
			},
			"zone_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("The supervisor zones this %s spans", labelVcfaDistributedVlanConnection),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func datasourceVcfaDistributedVlanConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	getDistributedVlanConnectionByName := func(name string) (*govcd.TmDistributedVlanConnection, error) {
		return tmClient.GetTmDistributedVlanConnectionByNameAndRegionId(name, d.Get("region_id").(string))
	}

	c := dsReadConfig[*govcd.TmDistributedVlanConnection, types.TmDistributedVlanConnection]{
		entityLabel:    labelVcfaDistributedVlanConnection,
		getEntityFunc:  getDistributedVlanConnectionByName,
		stateStoreFunc: setDistributedVlanConnectionData,
	}
	return readDatasource(ctx, d, meta, c)
}
