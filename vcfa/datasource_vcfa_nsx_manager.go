/*
 * // © Broadcom. All Rights Reserved.
 * // The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
 * // SPDX-License-Identifier: MPL-2.0
 */

package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaNsxManager() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaNsxManagerRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaNsxManager),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaNsxManager),
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Username for authenticating to %s", labelVcfaNsxManager),
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("URL of %s", labelVcfaNsxManager),
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Indicates whether the %s can or cannot be used to manage networking constructs within VCFA", labelVcfaNsxManager),
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Cluster ID of the %s. Each NSX installation has a single cluster. This is not a VCFA URN", labelVcfaNsxManager),
			},
			"is_dedicated_for_classic_tenants": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf("Whether this %s is dedicated for legacy VRA-style tenants only and unable to "+
					"participate in modern constructs such as Regions and Zones. Legacy VRA-style is deprecated and this field exists for "+
					"the purpose of VRA backwards compatibility only", labelVcfaNsxManager),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaNsxManager),
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("HREF of %s", labelVcfaNsxManager),
			},
		},
	}
}

func datasourceVcfaNsxManagerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := dsReadConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:    labelVcfaNsxManager,
		getEntityFunc:  tmClient.GetNsxtManagerOpenApiByName,
		stateStoreFunc: setNsxManagerData,
	}
	return readDatasource(ctx, d, meta, c)
}
