package vcfa

import (
	"context"
	"fmt"
	semver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcfaTmVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaTmVersionRead,
		Schema: map[string]*schema.Schema{
			"condition": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "A condition to check against the VCFA Tenant Manager version",
				RequiredWith: []string{"fail_if_not_match"},
			},
			"fail_if_not_match": {
				Type:         schema.TypeBool,
				Optional:     true,
				Description:  "This data source fails if the VCFA Tenant Manager doesn't match the version constraint set in 'condition'",
				RequiredWith: []string{"condition"},
			},
			"matches_condition": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether VCFA Tenant Manager matches the condition or not",
			},
			"tm_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VCFA Tenant Manager version",
			},
			"tm_api_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The maximum supported VCFA Tenant Manager API version",
			},
		},
	}
}

func datasourceVcfaTmVersionRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcfaVersion, err := vcdClient.VCDClient.Client.GetVcdShortVersion()
	if err != nil {
		return diag.Errorf("could not get VCFA Tenant Manager version: %s", err)
	}
	apiVersion, err := vcdClient.VCDClient.Client.MaxSupportedVersion()
	if err != nil {
		return diag.Errorf("could not get VCFA Tenant Manager API version: %s", err)
	}

	dSet(d, "tm_version", vcfaVersion)
	dSet(d, "tm_api_version", apiVersion)

	if condition, ok := d.GetOk("condition"); ok {
		checkVer, err := semver.NewVersion(vcfaVersion)
		if err != nil {
			return diag.Errorf("unable to parse version '%s': %s", vcfaVersion, err)
		}
		constraints, err := semver.NewConstraint(condition.(string))
		if err != nil {
			return diag.Errorf("unable to parse given version constraint '%s' : %s", condition, err)
		}
		matchesCondition := constraints.Check(checkVer)
		dSet(d, "matches_condition", matchesCondition)
		if !matchesCondition && d.Get("fail_if_not_match").(bool) {
			return diag.Errorf("the VCFA Tenant Manager version '%s' doesn't match the version constraint '%s'", vcfaVersion, condition)
		}
	}

	// The ID is artificial, and we try to identify each data source instance unequivocally through its parameters.
	d.SetId(fmt.Sprintf("tm_version='%s',condition='%s',fail_if_not_match='%t'", vcfaVersion, d.Get("condition"), d.Get("fail_if_not_match")))
	return nil
}
