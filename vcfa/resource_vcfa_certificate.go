package vcfa

import (
	"context"
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/govcd"

	"github.com/vmware/go-vcloud-director/v3/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcfaCertificate() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcfaCertificateRead,
		CreateContext: resourceVcfaCertificateCreate,
		UpdateContext: resourceVcfaCertificateUpdate,
		DeleteContext: resourceVcfaCertificateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaCertificateImport,
		},
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("The ID of %s to use", labelVcfaOrg),
			},
			"alias": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Alias of the Certificate",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Certificate description",
			},
			"certificate": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Certificate content",
			},
			"private_key": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "Certificate private key",
			},
			"private_key_passphrase": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "Certificate private passphrase",
			},
		},
	}
}

// resourceVcfaCertificateCreate covers Create functionality for resource
func resourceVcfaCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient

	org, err := vcdClient.GetTmOrgById(d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	certificateConfig := getCertificateConfigurationType(d)
	var createdCertificate *govcd.Certificate
	if isSysOrg(org) {
		createdCertificate, err = vcdClient.Client.AddCertificateToLibrary(certificateConfig)
	} else {
		// TODO: TM: Implement these methods in TmOrg
		var adminOrg *govcd.AdminOrg
		adminOrg, err = vcdClient.GetAdminOrgById(org.TmOrg.ID)
		if err != nil {
			return diag.FromErr(err)
		}
		createdCertificate, err = adminOrg.AddCertificateToLibrary(certificateConfig)
	}
	if err != nil {
		return diag.Errorf("error adding certificate library item: %s", err)
	}

	d.SetId(createdCertificate.CertificateLibrary.Id)
	return resourceVcfaCertificateRead(ctx, d, meta)
}

func isSysOrg(adminOrg *govcd.TmOrg) bool {
	return strings.EqualFold(adminOrg.TmOrg.Name, "system")
}

// resourceVcfaCertificateUpdate covers Update functionality for resource
func resourceVcfaCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	certificate, err := getCertificateType(vcdClient, d.Get("org_id").(string), d.Id())
	if err != nil {
		return diag.Errorf("[certificate library update] : %s", err)
	}

	certificateConfig := getCertificateConfigurationType(d)
	certificate.CertificateLibrary.Alias = certificateConfig.Alias
	certificate.CertificateLibrary.Description = certificateConfig.Description
	_, err = certificate.Update()
	if err != nil {
		return diag.Errorf("[certificate library update] : %s", err)
	}

	return resourceVcfaCertificateRead(ctx, d, meta)
}

func getCertificateConfigurationType(d *schema.ResourceData) *types.CertificateLibraryItem {
	return &types.CertificateLibraryItem{
		Alias:                d.Get("alias").(string),
		Description:          d.Get("description").(string),
		Certificate:          d.Get("certificate").(string),
		PrivateKey:           d.Get("private_key").(string),
		PrivateKeyPassphrase: d.Get("private_key_passphrase").(string),
	}
}

func resourceVcfaCertificateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient

	certificate, err := getCertificateType(vcdClient, d.Get("org_id").(string), d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[certificate library read] : %s", err)
	}

	setCertificateConfigurationData(certificate.CertificateLibrary, d)

	return nil
}

func setCertificateConfigurationData(config *types.CertificateLibraryItem, d *schema.ResourceData) {
	dSet(d, "alias", config.Alias)
	dSet(d, "description", config.Description)
	dSet(d, "certificate", config.Certificate)
}

func getCertificateType(vcdClient *VCDClient, orgId, certLibId string) (*govcd.Certificate, error) {
	org, err := vcdClient.GetTmOrgById(orgId)
	if err != nil {
		return nil, err
	}
	var certificate *govcd.Certificate
	if isSysOrg(org) {
		certificate, err = vcdClient.Client.GetCertificateFromLibraryById(certLibId)
	} else {
		// TODO: TM: Implement these methods in TmOrg
		var adminOrg *govcd.AdminOrg
		adminOrg, err = vcdClient.GetAdminOrgById(org.TmOrg.ID)
		if err != nil {
			return nil, err
		}
		certificate, err = adminOrg.GetCertificateFromLibraryById(certLibId)
	}
	if err != nil {
		return nil, err
	}
	return certificate, nil
}

func resourceVcfaCertificateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	certificateToDelete, err := getCertificateType(vcdClient, d.Get("org_id").(string), d.Id())
	if err != nil {
		return diag.Errorf("[certificate library delete] error fetching certificate library: %s", err)
	}
	err = certificateToDelete.Delete()
	if err != nil {
		diag.Errorf("[certificate library delete] error deleting certificate library %s", err)
	}
	return nil
}

func resourceVcfaCertificateImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org-name%scertificate-name", ImportSeparator)
	}
	orgName, certificateName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(MetaContainer).VcfaClient
	org, err := vcdClient.GetTmOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("[certificate import] error retrieving org %s: %s", orgName, err)
	}

	var certificate *govcd.Certificate
	if isSysOrg(org) {
		certificate, err = vcdClient.Client.GetCertificateFromLibraryByName(certificateName)
	} else {
		// TODO: TM: Implement these methods in TmOrg
		var adminOrg *govcd.AdminOrg
		adminOrg, err = vcdClient.GetAdminOrgById(org.TmOrg.ID)
		if err != nil {
			return nil, err
		}
		certificate, err = adminOrg.GetCertificateFromLibraryByName(certificateName)
	}
	if err != nil {
		return nil, fmt.Errorf("error importing certificate library item: %s", err)
	}

	d.SetId(certificate.CertificateLibrary.Id)
	dSet(d, "org_id", org.TmOrg.ID)
	setCertificateConfigurationData(certificate.CertificateLibrary, d)

	return []*schema.ResourceData{d}, nil
}
