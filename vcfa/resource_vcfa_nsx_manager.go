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

func resourceVcfaNsxManagerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:      labelVcfaNsxManager,
		getTypeFunc:      getNsxManagerType,
		stateStoreFunc:   setNsxManagerData,
		createFunc:       tmClient.CreateNsxtManagerOpenApi,
		resourceReadFunc: resourceVcfaNsxManagerRead,
		preCreateHooks:   []schemaHook{autoTrustHostCertificate("url", "auto_trust_certificate")},
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaNsxManagerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:      labelVcfaNsxManager,
		getTypeFunc:      getNsxManagerType,
		getEntityFunc:    tmClient.GetNsxtManagerOpenApiById,
		resourceReadFunc: resourceVcfaNsxManagerRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaNsxManagerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:    labelVcfaNsxManager,
		getEntityFunc:  tmClient.GetNsxtManagerOpenApiById,
		stateStoreFunc: setNsxManagerData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaNsxManagerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	c := crudConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:   labelVcfaNsxManager,
		getEntityFunc: tmClient.GetNsxtManagerOpenApiById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaNsxManagerImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	nsxManager, err := tmClient.GetNsxtManagerOpenApiByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s '%s': %s", labelVcfaNsxManager, d.Id(), err)
	}
	d.SetId(nsxManager.NsxtManagerOpenApi.ID)
	return []*schema.ResourceData{d}, nil
}

func getNsxManagerType(_ *VCDClient, d *schema.ResourceData) (*types.NsxtManagerOpenApi, error) {
	t := &types.NsxtManagerOpenApi{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Username:    d.Get("username").(string),
		Password:    d.Get("password").(string),
		Url:         d.Get("url").(string),
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
	dSet(d, "active", n.Active)
	dSet(d, "cluster_id", n.ClusterId)
	dSet(d, "is_dedicated_for_classic_tenants", n.IsDedicatedForClassicTenants)
	dSet(d, "status", n.Status)
	dSet(d, "href", t.BuildHref())

	return nil
}
