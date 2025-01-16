package vcfa

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcfaDeleteme() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaCatalogItemCreate,
		DeleteContext: resourceVcfaCatalogItemDelete,
		ReadContext:   resourceVcfaCatalogItemRead,
		UpdateContext: resourceVcfaCatalogItemUpdate,
		Schema: map[string]*schema.Schema{
			"foo": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceVcfaCatalogItemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("foo")
	return resourceVcfaCatalogItemRead(ctx, d, meta)
}

func resourceVcfaCatalogItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcfaCatalogItemDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcfaCatalogItemUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
