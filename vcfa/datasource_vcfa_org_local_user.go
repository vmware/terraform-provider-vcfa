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

func datasourceVcfaOrgLocalUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaOrgLocalUserRead,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent %s ID for %s", labelVcfaOrg, labelOrgLocalUser),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s User Name", labelOrgLocalUser),
			},
			"role_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("%ss to use for %s", labelVcfaRole, labelOrgLocalUser),
			},
		},
	}
}

func datasourceVcfaOrgLocalUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	getByNameFunc := func(username string) (*govcd.OpenApiUser, error) {
		return tmClient.GetUserByName(username, tenantContext)
	}

	c := dsReadConfig[*govcd.OpenApiUser, types.OpenApiUser]{
		entityLabel:              labelOrgLocalUser,
		getEntityFunc:            getByNameFunc,
		stateStoreFunc:           setOrgLocalUserData,
		overrideDefaultNameField: "username",
	}
	return readDatasource(ctx, d, meta, c)
}
