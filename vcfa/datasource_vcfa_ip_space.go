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

var dsIpSpaceIpBlockSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("ID of the IP Block within %s", labelVcfaIpSpace),
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("Name of the IP Block within %s", labelVcfaIpSpace),
		},
		"cidr": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("The CIDR that represents this IP Block within %s", labelVcfaIpSpace),
		},
	},
}

var dsIpSpaceIpRangeSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("ID of IP Range within %s", labelVcfaIpSpace),
		},
		"start_ip_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Starting IP address in the range",
		},
		"end_ip_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Ending IP address in the range",
		},
	},
}

func datasourceVcfaIpSpace() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaIpSpaceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaIpSpace),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Region ID for this %s", labelVcfaIpSpace),
			},
			"backing_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID for the matching IP Block in NSX",
			},
			"cidr_blocks": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("CIDR blocks of %s. Along with 'ip_address_ranges' typically defines the span of IP addresses used within a Data Center", labelVcfaIpSpace),
				Elem:        dsIpSpaceIpBlockSchema,
			},
			"default_quota_max_subnet_size": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Maximum subnet size represented as a prefix length (e.g. 24, 28) in %s", labelVcfaIpSpace),
			},
			"default_quota_max_cidr_count": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Maximum number of subnets that can be allocated from internal scope in this %s. ('-1' for unlimited)", labelVcfaIpSpace),
			},
			"default_quota_max_ip_count": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Maximum number of single floating IP addresses that can be allocated from internal scope in this %s. ('-1' for unlimited)", labelVcfaIpSpace),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaIpSpace),
			},
			"external_scope": {
				Type:        schema.TypeString,
				Computed:    true,
				Deprecated:  "Use 'inbound_remote_networks' in 'vcfa_provider_gateway' datasource instead",
				Description: "External scope in CIDR format",
			},
			"internal_scope": {
				Type:        schema.TypeSet,
				Computed:    true,
				Deprecated:  "Use 'cidr_blocks' instead",
				Description: fmt.Sprintf("Internal scope of %s. Along with 'ip_address_ranges' typically defines the span of IP addresses used within a Data Center", labelVcfaIpSpace),
				Elem:        dsIpSpaceIpBlockSchema,
			},
			"ip_address_ranges": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("IP address ranges of %s. Along with 'cidr_blocks' typically defines the span of IP addresses used within a Data Center", labelVcfaIpSpace),
				Elem:        dsIpSpaceIpRangeSchema,
			},
			"is_imported_ip_block": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if the IP Block is imported from an existing NSX IP Block",
			},
			"provider_visibility_only": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("If set to true, the %s details will be hidden from organizations", labelVcfaIpSpace),
			},
			"reserved_ip_address_ranges": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("IP addresses that will not be considered for IP allocation within %s", labelVcfaIpSpace),
				Elem:        dsIpSpaceIpRangeSchema,
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaIpSpace),
			},
			"subnet_exclusive": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this IP Block is exclusively for a single CIDR",
			},
		},
	}
}

func datasourceVcfaIpSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	getIpSpaceByName := func(name string) (*govcd.TmIpSpace, error) {
		return tmClient.GetTmIpSpaceByNameAndRegionId(name, d.Get("region_id").(string))
	}

	c := dsReadConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:    labelVcfaIpSpace,
		getEntityFunc:  getIpSpaceByName,
		stateStoreFunc: setIpSpaceData,
	}
	return readDatasource(ctx, d, meta, c)
}
