package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcfaRightsBundle() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaRightsBundleRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaRightsBundle),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s description", labelVcfaRightsBundle),
			},
			"bundle_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is read-only", labelVcfaRightsBundle),
			},
			"rights": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("Set of %ss assigned to this %s", labelVcfaRight, labelVcfaRightsBundle),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"publish_to_all_orgs": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("When true, publishes the %s to all %ss", labelVcfaRightsBundle, labelVcfaOrg),
			},
			"org_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("Set of %ss to which this %s is published", labelVcfaOrg, labelVcfaRightsBundle),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func datasourceVcfaRightsBundleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcfaRightsBundleRead(ctx, d, meta, "datasource", "read")
}
