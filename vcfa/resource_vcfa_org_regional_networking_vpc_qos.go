package vcfa

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaOrgRegionalNetworkingVpcQos = "Regional Networking VPC Connectivity Profile QoS"

func resourceVcfaOrgRegionalNetworkingVpcQos() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaOrgRegionalNetworkingVpcQosCreateUpdate,
		ReadContext:   resourceVcfaOrgRegionalNetworkingVpcQosRead,
		UpdateContext: resourceVcfaOrgRegionalNetworkingVpcQosCreateUpdate,
		DeleteContext: resourceVcfaOrgRegionalNetworkingVpcQosDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgRegionalNetworkingVpcQosImport,
		},

		Schema: map[string]*schema.Schema{
			"org_regional_networking_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("ID of %s", labelVcfaRegionalNetworkingSetting),
			},
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("ID of parent %s", labelVcfaEdgeCluster),
			},
			"ingress_committed_bandwidth_mbps": {
				Type:        schema.TypeString, // string + validation due to usual problem of differentiation between 0 and empty value for TypeInt
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf("Ingress committed bandwidth in Mbps for %s", labelVcfaOrgRegionalNetworkingVpcQos),
				ValidateDiagFunc: validation.AnyDiag(
					validation.ToDiagFunc(validation.StringIsEmpty),
					IsIntAndAtLeast(-1), // -1 is unlimited
				),
				RequiredWith: []string{"ingress_burst_size_bytes"},
			},
			"ingress_burst_size_bytes": {
				Type:        schema.TypeString, // string + validation due to usual problem of differentiation between 0 and empty value for TypeInt
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf("Ingress burst size bytes for %s", labelVcfaOrgRegionalNetworkingVpcQos),
				ValidateDiagFunc: validation.AnyDiag(
					validation.ToDiagFunc(validation.StringIsEmpty),
					IsIntAndAtLeast(-1), // -1 is unlimited
				),
				RequiredWith: []string{"ingress_committed_bandwidth_mbps"},
			},
			"egress_committed_bandwidth_mbps": {
				Type:        schema.TypeString, // string + validation due to usual problem of differentiation between 0 and empty value for TypeInt
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf("Egress committed bandwidth in Mbps for %s", labelVcfaOrgRegionalNetworkingVpcQos),
				ValidateDiagFunc: validation.AnyDiag(
					validation.ToDiagFunc(validation.StringIsEmpty),
					IsIntAndAtLeast(-1), // -1 is unlimited
				),
				RequiredWith: []string{"egress_burst_size_bytes"},
			},
			"egress_burst_size_bytes": {
				Type:        schema.TypeString, // string + validation due to usual problem of differentiation between 0 and empty value for TypeInt
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf("Ingress burst size bytes for %s", labelVcfaOrgRegionalNetworkingVpcQos),
				ValidateDiagFunc: validation.AnyDiag(
					validation.ToDiagFunc(validation.StringIsEmpty),
					IsIntAndAtLeast(-1), // -1 is unlimited
				),
				RequiredWith: []string{"egress_committed_bandwidth_mbps"},
			},
		},
	}
}

func resourceVcfaOrgRegionalNetworkingVpcQosCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	rns, err := tmClient.GetTmRegionalNetworkingSettingById(d.Get("org_regional_networking_id").(string))
	if err != nil {
		return diag.Errorf("error looking up %s by ID: %s", labelVcfaOrgNetworking, err)
	}

	// ID is Org Regional Networking Setting ID
	d.SetId(rns.TmRegionalNetworkingSetting.ID)

	cfg, err := getTmOrgRegionalNetworkingVpcQosType(tmClient, rns, d, true)
	if err != nil {
		return diag.Errorf("error getting %s configuration: %s", labelVcfaOrgRegionalNetworkingVpcQos, err)
	}

	_, err = rns.UpdateDefaultVpcConnectivityProfile(cfg)
	if err != nil {
		return diag.Errorf("error setting %s configuration: %s", labelVcfaOrgRegionalNetworkingVpcQos, err)
	}

	return resourceVcfaOrgRegionalNetworkingVpcQosRead(ctx, d, meta)
}

func resourceVcfaOrgRegionalNetworkingVpcQosRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	rns, err := tmClient.GetTmRegionalNetworkingSettingById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error looking up %s by ID: %s", labelVcfaOrgNetworking, err)
	}

	// ID is Org Regional Networking Setting ID
	d.SetId(rns.TmRegionalNetworkingSetting.ID)

	// fetch VPC profile
	vpcProfile, err := rns.GetDefaultVpcConnectivityProfile()
	if err != nil {
		return diag.Errorf("error retrieving %s: %s", labelVcfaOrgRegionalNetworkingVpcQos, err)
	}

	err = setTmOrgRegionalNetworkingVpcQosData(tmClient, d, vpcProfile)
	if err != nil {
		return diag.Errorf("error storing %s configuration to state: %s", labelVcfaOrgRegionalNetworkingVpcQos, err)
	}

	return nil
}

func resourceVcfaOrgRegionalNetworkingVpcQosDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	rns, err := tmClient.GetTmRegionalNetworkingSettingById(d.Get("org_regional_networking_id").(string))
	if err != nil {
		return diag.Errorf("error looking up %s by ID: %s", labelVcfaOrgNetworking, err)
	}

	// ID is Org Regional Networking Setting ID
	d.SetId(rns.TmRegionalNetworkingSetting.ID)

	cfg, err := getTmOrgRegionalNetworkingVpcQosType(tmClient, rns, d, false)
	if err != nil {
		return diag.Errorf("error getting %s configuration: %s", labelVcfaOrgRegionalNetworkingVpcQos, err)
	}

	_, err = rns.UpdateDefaultVpcConnectivityProfile(cfg)
	if err != nil {
		return diag.Errorf("error removing custom %s configuration: %s", labelVcfaOrgRegionalNetworkingVpcQos, err)
	}

	d.SetId("")

	return nil
}

func resourceVcfaOrgRegionalNetworkingVpcQosImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	id := strings.Split(d.Id(), ImportSeparator)
	if len(id) != 2 {
		return nil, fmt.Errorf("ID syntax should be <%s name>%s<%s name>", labelVcfaOrg, ImportSeparator, labelVcfaRegionalNetworkingSetting)
	}

	org, err := tmClient.GetTmOrgByName(id[0])
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s '%s': %s", labelVcfaOrg, id[0], err)
	}

	rns, err := tmClient.GetTmRegionalNetworkingSettingByNameAndOrgId(id[1], org.TmOrg.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s '%s' within %s '%s': %s",
			labelVcfaRegionalNetworkingSetting, id[1], labelVcfaOrg, id[0], err)
	}

	d.SetId(rns.TmRegionalNetworkingSetting.ID)
	dSet(d, "org_regional_networking_id", rns.TmRegionalNetworkingSetting.ID)
	return []*schema.ResourceData{d}, nil
}

// flag `isCreatedUpdate` is used to separate Create/Update and Delete operations
func getTmOrgRegionalNetworkingVpcQosType(tmClient *VCDClient, rns *govcd.TmRegionalNetworkingSetting, d *schema.ResourceData, isCreatedOrUpdated bool) (*types.TmRegionalNetworkingVpcConnectivityProfile, error) {
	// The QoS config of Edge Cluster that is used for Region is used as main configuration. One
	// can override it per Org Regional Networking configuration

	// Starting with parent Edge Cluster QoS configuration
	existingVpcProfile, err := rns.GetDefaultVpcConnectivityProfile()
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s: %s", labelVcfaOrgRegionalNetworkingVpcQos, err)
	}

	// Fetch edge cluster
	if existingVpcProfile.ServiceEdgeClusterRef == nil {
		return nil, fmt.Errorf("could not find %s for %s with ID %s", labelVcfaEdgeCluster, labelVcfaOrgNetworking, rns.TmRegionalNetworkingSetting.ID)
	}

	// Fetch parent Edge Cluster QoS Config
	backingEdgeCluster, err := tmClient.GetTmEdgeClusterById(existingVpcProfile.ServiceEdgeClusterRef.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving parent %s for %s with ID %s", labelVcfaEdgeCluster, labelVcfaOrgNetworking, rns.TmRegionalNetworkingSetting.ID)
	}

	// Starting with QoS Config that is based on parent Edge Cluster
	newQosConfig := &types.VpcConnectivityProfileQosConfig{
		IngressProfile: &types.VpcConnectivityProfileQosProfile{
			Type:                   "DEFAULT",
			CommittedBandwidthMbps: backingEdgeCluster.TmEdgeCluster.DefaultQosConfig.IngressProfile.CommittedBandwidthMbps,
			BurstSizeBytes:         backingEdgeCluster.TmEdgeCluster.DefaultQosConfig.IngressProfile.BurstSizeBytes,
		},
		EgressProfile: &types.VpcConnectivityProfileQosProfile{
			Type:                   "DEFAULT",
			CommittedBandwidthMbps: backingEdgeCluster.TmEdgeCluster.DefaultQosConfig.EgressProfile.CommittedBandwidthMbps,
			BurstSizeBytes:         backingEdgeCluster.TmEdgeCluster.DefaultQosConfig.EgressProfile.BurstSizeBytes,
		},
	}

	// overriding fields to specified ones if it is Create or Update. Not changing any values for Delete
	if isCreatedOrUpdated {
		// Overriding just provided fields
		ingressCommittedBandwidthMbps := d.Get("ingress_committed_bandwidth_mbps").(string)
		ingressBurstSizeBytes := d.Get("ingress_burst_size_bytes").(string)

		if ingressCommittedBandwidthMbps != "" || ingressBurstSizeBytes != "" { // schema requires both to be set if one is set
			newQosConfig.IngressProfile = &types.VpcConnectivityProfileQosProfile{
				Type:                   "CUSTOM",
				CommittedBandwidthMbps: mustStrToInt(ingressCommittedBandwidthMbps), // schema validates that fields are ints
				BurstSizeBytes:         mustStrToInt(ingressBurstSizeBytes),
			}
		}

		egressCommittedBandwidthMbps := d.Get("egress_committed_bandwidth_mbps").(string)
		egressBurstSizeBytes := d.Get("egress_burst_size_bytes").(string)
		if egressCommittedBandwidthMbps != "" || egressBurstSizeBytes != "" { // schema requires both to be set if one is set
			newQosConfig.EgressProfile = &types.VpcConnectivityProfileQosProfile{
				Type:                   "CUSTOM",
				CommittedBandwidthMbps: mustStrToInt(egressCommittedBandwidthMbps), // schema validates that fields are ints
				BurstSizeBytes:         mustStrToInt(egressBurstSizeBytes),
			}
		}
	}

	// update existing VPC profile with new QoS config, leave other body as it was
	existingVpcProfile.QosConfig = newQosConfig

	return existingVpcProfile, nil
}

func setTmOrgRegionalNetworkingVpcQosData(_ *VCDClient, d *schema.ResourceData, t *types.TmRegionalNetworkingVpcConnectivityProfile) error {
	if t == nil {
		return fmt.Errorf("empty %s received", labelVcfaOrgRegionalNetworkingVpcQos)
	}

	if t.ServiceEdgeClusterRef != nil {
		dSet(d, "edge_cluster_id", t.ServiceEdgeClusterRef.ID)
	}

	dSet(d, "ingress_committed_bandwidth_mbps", nil)
	dSet(d, "ingress_burst_size_bytes", nil)
	if t.QosConfig != nil && t.QosConfig.IngressProfile != nil {
		strValue := strconv.Itoa(t.QosConfig.IngressProfile.BurstSizeBytes)
		dSet(d, "ingress_burst_size_bytes", strValue)

		strValueCommitted := strconv.Itoa(t.QosConfig.IngressProfile.CommittedBandwidthMbps)
		dSet(d, "ingress_committed_bandwidth_mbps", strValueCommitted)
	}

	dSet(d, "egress_committed_bandwidth_mbps", nil)
	dSet(d, "egress_burst_size_bytes", nil)
	if t.QosConfig != nil && t.QosConfig.EgressProfile != nil {
		strValueCommitted := strconv.Itoa(t.QosConfig.EgressProfile.CommittedBandwidthMbps)
		dSet(d, "egress_committed_bandwidth_mbps", strValueCommitted)

		strValueBurstSize := strconv.Itoa(t.QosConfig.EgressProfile.BurstSizeBytes)
		dSet(d, "egress_burst_size_bytes", strValueBurstSize)
	}

	return nil
}
