package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaNsxtManager = "NSX-T Manager"

func resourceVcfaNsxtManager() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaNsxtManagerCreate,
		ReadContext:   resourceVcfaNsxtManagerRead,
		UpdateContext: resourceVcfaNsxtManagerUpdate,
		DeleteContext: resourceVcfaNsxtManagerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaNsxtManagerImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaNsxtManager),
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaNsxtManager),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Username for authenticating to %s", labelVcfaNsxtManager),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: fmt.Sprintf("Password for authenticating to %s", labelVcfaNsxtManager),
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("URL of %s", labelVcfaNsxtManager),
			},
			"auto_trust_certificate": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Defines if the %s certificate should automatically be trusted", labelVcfaNsxtManager),
			},
			"network_provider_scope": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Network Provider Scope for %s", labelVcfaNsxtManager),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaNsxtManager),
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("HREF of %s", labelVcfaNsxtManager),
			},
		},
	}
}

func resourceVcfaNsxtManagerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:      labelVcfaNsxtManager,
		getTypeFunc:      getNsxtManagerType,
		stateStoreFunc:   setNsxtManagerData,
		createFunc:       vcdClient.CreateNsxtManagerOpenApi,
		resourceReadFunc: resourceVcfaNsxtManagerRead,
		preCreateHooks:   []schemaHook{autoTrustHostCertificate("url", "auto_trust_certificate")},
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaNsxtManagerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:      labelVcfaNsxtManager,
		getTypeFunc:      getNsxtManagerType,
		getEntityFunc:    vcdClient.GetNsxtManagerOpenApiById,
		resourceReadFunc: resourceVcfaNsxtManagerRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaNsxtManagerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:    labelVcfaNsxtManager,
		getEntityFunc:  vcdClient.GetNsxtManagerOpenApiById,
		stateStoreFunc: setNsxtManagerData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaNsxtManagerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:   labelVcfaNsxtManager,
		getEntityFunc: vcdClient.GetNsxtManagerOpenApiById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaNsxtManagerImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	nsxtManager, err := vcdClient.GetNsxtManagerOpenApiByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s '%s': %s", labelVcfaNsxtManager, d.Id(), err)
	}
	d.SetId(nsxtManager.NsxtManagerOpenApi.ID)
	return []*schema.ResourceData{d}, nil
}

func getNsxtManagerType(_ *VCDClient, d *schema.ResourceData) (*types.NsxtManagerOpenApi, error) {
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

func setNsxtManagerData(_ *VCDClient, d *schema.ResourceData, t *govcd.NsxtManagerOpenApi) error {
	if t == nil || t.NsxtManagerOpenApi == nil {
		return fmt.Errorf("nil object for %s", labelVcfaNsxtManager)
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
