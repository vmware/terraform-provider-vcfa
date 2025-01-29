package vcfa

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

func datasourceVcfaLibraryCertificate() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdLibraryCertificateRead,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The ID of %s to use", labelVcfaOrg),
			},
			"alias": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"alias", "id"},
				Description:  "Alias of certificate",
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"alias", "id"},
				Description:  "Certificate ID",
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Certificate description",
			},
			"certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Certificate content",
			},
		},
	}
}

func datasourceVcdLibraryCertificateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetTmOrgById(d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// get by ID when it's available
	var certificate *govcd.Certificate
	if isSysOrg(org) {
		alias := d.Get("alias").(string)
		if alias != "" {
			certificate, err = vcdClient.Client.GetCertificateFromLibraryByName(alias)
		} else if d.Get("id").(string) != "" {
			certificate, err = vcdClient.Client.GetCertificateFromLibraryById(d.Get("id").(string))
		} else {
			return diag.Errorf("Id or Alias value is missing %s", err)
		}
	} else {
		alias := d.Get("alias").(string)
		// TODO: TM: Implement these methods in TmOrg
		var adminOrg *govcd.AdminOrg
		adminOrg, err = vcdClient.GetAdminOrgById(org.TmOrg.ID)
		if err != nil {
			return diag.FromErr(err)
		}
		if alias != "" {
			certificate, err = adminOrg.GetCertificateFromLibraryByName(alias)
		} else if d.Get("id").(string) != "" {
			certificate, err = adminOrg.GetCertificateFromLibraryById(d.Get("id").(string))
		} else {
			return diag.Errorf("Id or Alias value is missing %s", err)
		}
	}
	if err != nil {
		return diag.Errorf("[certificate library read] : %s", err)
	}

	d.SetId(certificate.CertificateLibrary.Id)
	setCertificateConfigurationData(certificate.CertificateLibrary, d)

	return nil
}
