package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaNsxManager = "NSX Manager"

func resourceVcfaNsxManager() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaNsxManagerCreate,
		ReadContext:   resourceVcfaNsxManagerRead,
		UpdateContext: resourceVcfaNsxManagerUpdate,
		DeleteContext: resourceVcfaNsxManagerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaNsxManagerImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaNsxManager),
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaNsxManager),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Username for authenticating to %s", labelVcfaNsxManager),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: fmt.Sprintf("Password for authenticating to %s", labelVcfaNsxManager),
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("URL of %s", labelVcfaNsxManager),
			},
			"auto_trust_certificate": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Defines if the %s certificate should automatically be trusted", labelVcfaNsxManager),
			},
			"network_provider_scope": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Network Provider Scope for %s", labelVcfaNsxManager),
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

func resourceVcfaNsxManagerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:      labelVcfaNsxManager,
		getTypeFunc:      getNsxManagerType,
		stateStoreFunc:   setNsxManagerData,
		createFunc:       vcdClient.CreateNsxtManagerOpenApi,
		resourceReadFunc: resourceVcfaNsxManagerRead,
		preCreateHooks:   []schemaHook{autoTrustHostCertificate("url", "auto_trust_certificate")},
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaNsxManagerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:      labelVcfaNsxManager,
		getTypeFunc:      getNsxManagerType,
		getEntityFunc:    vcdClient.GetNsxtManagerOpenApiById,
		resourceReadFunc: resourceVcfaNsxManagerRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaNsxManagerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:    labelVcfaNsxManager,
		getEntityFunc:  vcdClient.GetNsxtManagerOpenApiById,
		stateStoreFunc: setNsxManagerData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaNsxManagerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient

	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:   labelVcfaNsxManager,
		getEntityFunc: vcdClient.GetNsxtManagerOpenApiById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaNsxManagerImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(MetaContainer).VcfaClient

	nsxManager, err := vcdClient.GetNsxtManagerOpenApiByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s '%s': %s", labelVcfaNsxManager, d.Id(), err)
	}
	d.SetId(nsxManager.NsxtManagerOpenApi.ID)
	return []*schema.ResourceData{d}, nil
}

func getNsxManagerType(_ *VCDClient, d *schema.ResourceData) (*types.NsxtManagerOpenApi, error) {
	t := &types.NsxtManagerOpenApi{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		Username:             d.Get("username").(string),
		Password:             d.Get("password").(string),
		Url:                  d.Get("url").(string),
		NetworkProviderScope: d.Get("network_provider_scope").(string),
	}

	return t, nil
}

func setNsxManagerData(_ *VCDClient, d *schema.ResourceData, t *govcd.NsxtManagerOpenApi) error {
	if t == nil || t.NsxtManagerOpenApi == nil {
		return fmt.Errorf("nil object for %s", labelVcfaNsxManager)
	}
	n := t.NsxtManagerOpenApi

	d.SetId(n.ID)
	dSet(d, "name", n.Name)
	dSet(d, "description", n.Description)
	dSet(d, "username", n.Username)
	// dSet(d, "password", n.Password) // real password is never returned
	dSet(d, "url", n.Url)
	dSet(d, "network_provider_scope", n.NetworkProviderScope)
	dSet(d, "status", n.Status)
	dSet(d, "href", t.BuildHref())

	return nil
}
