package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const labelVcfaRight = "Right"

func datasourceVcfaRight() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceRightRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaRight),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s description", labelVcfaRight),
			},
			"category_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("ID of the category for this %s", labelVcfaRight),
			},
			"bundle_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"right_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Type of the %s", labelVcfaRight),
			},
			"implied_rights": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("List of %ss that are implied with this one", labelVcfaRight),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: fmt.Sprintf("Name of the implied %s", labelVcfaRight),
						},
						"id": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: fmt.Sprintf("ID of the implied %s", labelVcfaRight),
						},
					},
				},
			},
		},
	}
}

func datasourceRightRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient

	rightName := d.Get("name").(string)

	right, err := vcdClient.Client.GetRightByName(rightName)
	if err != nil {
		return diag.Errorf("[%s read] error searching for right %s: %s", labelVcfaRight, rightName, err)
	}

	d.SetId(right.ID)
	dSet(d, "description", right.Description)
	dSet(d, "right_type", right.RightType)
	dSet(d, "category_id", right.Category)
	dSet(d, "bundle_key", right.BundleKey)
	var impliedRights []map[string]interface{}
	for _, ir := range right.ImpliedRights {
		impliedRights = append(impliedRights, map[string]interface{}{
			"name": ir.Name,
			"id":   ir.ID,
		})
	}
	err = d.Set("implied_rights", impliedRights)
	if err != nil {
		return diag.Errorf("[%s read] error setting implied rights for right %s: %s", labelVcfaRight, rightName, err)
	}
	return nil
}
