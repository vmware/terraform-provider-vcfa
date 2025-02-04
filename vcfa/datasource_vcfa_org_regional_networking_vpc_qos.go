package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcfaOrgRegionalNetworkingVpcQos() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaOrgRegionalNetworkingVpcQosRead,

		Schema: map[string]*schema.Schema{
			"org_regional_networking_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("ID of %s", labelVcfaRegionalNetworkingSetting),
			},
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("ID of parent %s", labelVcfaEdgeCluster),
			},
			"ingress_committed_bandwidth_mbps": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Ingress committed bandwidth in Mbps for %s", labelVcfaOrgRegionalNetworkingVpcQos),
			},
			"ingress_burst_size_bytes": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Ingress burst size bytes for %s", labelVcfaOrgRegionalNetworkingVpcQos),
			},
			"egress_committed_bandwidth_mbps": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Egress committed bandwidth in Mbps for %s", labelVcfaOrgRegionalNetworkingVpcQos),
			},
			"egress_burst_size_bytes": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Ingress burst size bytes for %s", labelVcfaOrgRegionalNetworkingVpcQos),
			},
		},
	}
}

func datasourceVcfaOrgRegionalNetworkingVpcQosRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).tmClient
	rns, err := vcfaClient.GetTmRegionalNetworkingSettingById(d.Get("org_regional_networking_id").(string))
	if err != nil {
		return diag.Errorf("error looking up %s by ID: %s", labelVcfaOrgNetworking, err)
	}

	// ID is Org Regional Networking Setting ID
	d.SetId(rns.TmRegionalNetworkingSetting.ID)

	// fetch VPC profile
	vpcProfile, err := rns.GetDefaultVpcConnectivityProfile()
	if err != nil {
		return diag.Errorf("error retrieving %s: %s", labelVcfaOrgRegionalNetworkingVpcQos, err)
	}

	err = setTmOrgRegionalNetworkingVpcQosData(vcfaClient, d, vpcProfile)
	if err != nil {
		return diag.Errorf("error storing %s configuration to state: %s", labelVcfaOrgRegionalNetworkingVpcQos, err)
	}

	return nil
}
