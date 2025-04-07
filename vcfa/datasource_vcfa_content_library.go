// © Broadcom. All Rights Reserved.
// The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
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

func datasourceVcfaContentLibrary() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaContentLibraryRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The name of the %s", labelVcfaContentLibrary),
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The reference to the %s that the %s belongs to", labelVcfaOrg, labelVcfaContentLibrary),
			},
			"storage_class_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("A set of %s IDs used by this %s", labelVcfaStorageClass, labelVcfaContentLibrary),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"auto_attach": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf("For Tenant Content Libraries this field represents whether this %s should be "+
					"automatically attached to all current and future namespaces in the tenant organization", labelVcfaContentLibrary),
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The ISO-8601 timestamp representing when this %s was created", labelVcfaContentLibrary),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The description of the %s", labelVcfaContentLibrary),
			},
			"is_shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is shared with other %ss", labelVcfaContentLibrary, labelVcfaOrg),
			},
			"is_subscribed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is subscribed from an external published library", labelVcfaContentLibrary),
			},
			"library_type": {
				Type:     schema.TypeString,
				Computed: true,
				Description: fmt.Sprintf("The type of %s, can be either PROVIDER (%s that is scoped to a "+
					"provider) or TENANT (%s that is scoped to a tenant organization)", labelVcfaContentLibrary, labelVcfaContentLibrary, labelVcfaContentLibrary),
			},
			"subscription_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: fmt.Sprintf("A block representing subscription settings of a %s", labelVcfaContentLibrary),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subscription_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: fmt.Sprintf("Subscription url of this %s", labelVcfaContentLibrary),
						},
						"password": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Password to use to authenticate with the publisher",
						},
					},
				},
			},
			"version_number": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Version number of this %s", labelVcfaContentLibrary),
			},
		},
	}
}

func datasourceVcfaContentLibraryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := dsReadConfig[*govcd.ContentLibrary, types.ContentLibrary]{
		entityLabel: labelVcfaContentLibrary,
		getEntityFunc: func(name string) (*govcd.ContentLibrary, error) {
			tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
			if err != nil {
				return nil, err
			}
			return tmClient.GetContentLibraryByName(name, tenantContext)
		},
		stateStoreFunc: setContentLibraryData,
	}
	return readDatasource(ctx, d, meta, c)
}
