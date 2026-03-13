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

func datasourceVcfaProviderGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaProviderGatewayRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaProviderGateway),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent %s of %s", labelVcfaRegion, labelVcfaProviderGateway),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaProviderGateway),
			},
			"tier0_gateway_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Parent %s of %s", labelVcfaTier0Gateway, labelVcfaProviderGateway),
			},
			"ip_space_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("A set of %s IDs used in this %s", labelVcfaIpSpace, labelVcfaProviderGateway),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"gateway_connection_backing_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the associated Gateway Connection in NSX, if any",
			},
			"inbound_remote_networks": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("The total span of IP addresses to which the %s has access", labelVcfaProviderGateway),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"allow_advertising_private_ip_blocks": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Allows the %s to advertise their own private IP Blocks", labelVcfaProviderGateway),
			},
			"nat_config_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether the Outbound NAT is enabled for the %s", labelVcfaProviderGateway),
			},
			"nat_config_ip_space_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s used to configure Outbound NAT", labelVcfaIpSpace),
			},
			"nat_config_logging": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether logging is enabled for the Outbound NAT configuration",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaProviderGateway),
			},
		},
	}
}

func datasourceVcfaProviderGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	getProviderGateway := func(name string) (*govcd.TmProviderGateway, error) {
		return tmClient.GetTmProviderGatewayByNameAndRegionId(name, d.Get("region_id").(string))
	}
	c := dsReadConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:    labelVcfaProviderGateway,
		getEntityFunc:  getProviderGateway,
		stateStoreFunc: setProviderGatewayData,
	}
	return readDatasource(ctx, d, meta, c)
}
