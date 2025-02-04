package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaLocalUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaLocalUserRead,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent %s ID for %s", labelVcfaOrg, labelLocalUser),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s User Name", labelLocalUser),
			},
			"role_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("%ss to use for %s", labelVcfaRole, labelLocalUser),
			},
		},
	}
}

func datasourceVcfaLocalUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(vcfaClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	getByNameFunc := func(username string) (*govcd.OpenApiUser, error) {
		return vcfaClient.GetUserByName(username, tenantContext)
	}

	c := dsReadConfig[*govcd.OpenApiUser, types.OpenApiUser]{
		entityLabel:              labelLocalUser,
		getEntityFunc:            getByNameFunc,
		stateStoreFunc:           setLocalUserData,
		overrideDefaultNameField: "username",
	}
	return readDatasource(ctx, d, meta, c)
}
