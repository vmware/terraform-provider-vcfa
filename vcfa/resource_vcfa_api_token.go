package vcfa

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

const labelVcfaApiToken = "API Token"

func resourceVcfaApiToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaApiTokenCreate,
		ReadContext:   resourceVcfaApiTokenRead,
		DeleteContext: resourceVcfaApiTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaApiTokenImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaApiToken),
			},
			"file_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Name of the file that the %s will be saved to", labelVcfaApiToken),
			},
			"allow_token_file": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
				Description: fmt.Sprintf("Set this to true if you understand the security risks of using"+
					" %s files and agree to creating them", labelVcfaApiToken),
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					value := i.(bool)
					if !value {
						return diag.Diagnostics{
							diag.Diagnostic{
								Severity: diag.Error,
								Summary:  "This field must be set to true",
								Detail: fmt.Sprintf("The %s file should be considered SENSITIVE INFORMATION. "+
									"If you acknowledge that, set 'allow_token_file' to 'true'.", labelVcfaApiToken),
								AttributePath: path,
							},
						}
					}
					return nil
				},
			},
		},
	}
}

func resourceVcfaApiTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient

	// System Admin can't create API tokens outside SysOrg,
	// just as Org admins can't create API tokens in other Orgs
	org := vcdClient.SysOrg
	if org == "" {
		org = vcdClient.Org
	}

	tokenName := d.Get("name").(string)
	token, err := vcdClient.CreateToken(org, tokenName)
	if err != nil {
		return diag.Errorf("[%s create] error creating %s: %s", labelVcfaApiToken, labelVcfaApiToken, err)
	}
	d.SetId(token.Token.ID)

	apiToken, err := token.GetInitialApiToken()
	if err != nil {
		return diag.Errorf("[%s create] error getting refresh token from %s: %s", labelVcfaApiToken, labelVcfaApiToken, err)
	}

	filename := d.Get("file_name").(string)

	err = govcd.SaveApiTokenToFile(filename, vcdClient.Client.UserAgent, apiToken)
	if err != nil {
		return diag.Errorf("[%s create] error saving %s to file: %s", labelVcfaApiToken, labelVcfaApiToken, err)
	}

	return resourceVcfaApiTokenRead(ctx, d, meta)
}

func resourceVcfaApiTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient

	token, err := vcdClient.GetTokenById(d.Id())
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		log.Printf("[DEBUG] %s no longer exists. Removing from tfstate", labelVcfaApiToken)
	}
	if err != nil {
		return diag.Errorf("[%s read] error getting %s: %s", labelVcfaApiToken, labelVcfaApiToken, err)
	}

	d.SetId(token.Token.ID)
	dSet(d, "name", token.Token.Name)

	return nil
}

func resourceVcfaApiTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient

	token, err := vcdClient.GetTokenById(d.Id())
	if err != nil {
		return diag.Errorf("[%s delete] error getting %s: %s", labelVcfaApiToken, labelVcfaApiToken, err)
	}

	err = token.Delete()
	if err != nil {
		return diag.Errorf("[%s delete] error deleting %s: %s", labelVcfaApiToken, labelVcfaApiToken, err)
	}

	return nil
}

func resourceVcfaApiTokenImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] %s import initiated", labelVcfaApiToken)

	vcdClient := meta.(MetaContainer).VcfaClient
	sessionInfo, err := vcdClient.Client.GetSessionInfo()
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("[%s import] error getting username: %s", labelVcfaApiToken, err)
	}

	token, err := vcdClient.GetTokenByNameAndUsername(d.Id(), sessionInfo.User.Name)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("[%s import] error getting %s by name: %s", labelVcfaApiToken, labelVcfaApiToken, err)
	}

	d.SetId(token.Token.ID)
	dSet(d, "name", token.Token.Name)

	return []*schema.ResourceData{d}, nil
}
