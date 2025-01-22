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

const labelVcfaEdgeClusterQos = "Edge Cluster QoS"

func resourceVcfaEdgeClusterQos() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaEdgeClusterQosCreate,
		ReadContext:   resourceVcfaEdgeClusterQosRead,
		UpdateContext: resourceVcfaEdgeClusterQosUpdate,
		DeleteContext: resourceVcfaEdgeClusterQosDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaEdgeClusterQosImport,
		},

		Schema: map[string]*schema.Schema{
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("ID of %s", labelVcfaEdgeCluster),
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Region ID of  %s", labelVcfaEdgeCluster),
			},
			"ingress_committed_bandwidth_mbps": {
				Type:        schema.TypeString, // string + validation due to usual problem of differentiation between 0 and empty value for TypeInt
				Optional:    true,
				Description: fmt.Sprintf("Ingress committed bandwidth in Mbps for %s", labelVcfaEdgeCluster),
				ValidateDiagFunc: validation.AnyDiag(
					validation.ToDiagFunc(validation.StringIsEmpty),
					validation.ToDiagFunc(IsIntAndAtLeast(1))),
				RequiredWith: []string{"ingress_burst_size_bytes"},
			},
			"ingress_burst_size_bytes": {
				Type:        schema.TypeString, // string + validation due to usual problem of differentiation between 0 and empty value for TypeInt
				Optional:    true,
				Description: fmt.Sprintf("Ingress burst size bytes for %s", labelVcfaEdgeCluster),
				ValidateDiagFunc: validation.AnyDiag(
					validation.ToDiagFunc(validation.StringIsEmpty),
					validation.ToDiagFunc(IsIntAndAtLeast(1))),
				RequiredWith: []string{"ingress_committed_bandwidth_mbps"},
			},
			"egress_committed_bandwidth_mbps": {
				Type:        schema.TypeString, // string + validation due to usual problem of differentiation between 0 and empty value for TypeInt
				Optional:    true,
				Description: fmt.Sprintf("Egress committed bandwidth in Mbps for %s", labelVcfaEdgeCluster),
				ValidateDiagFunc: validation.AnyDiag(
					validation.ToDiagFunc(validation.StringIsEmpty),
					validation.ToDiagFunc(IsIntAndAtLeast(1))),
				RequiredWith: []string{"egress_burst_size_bytes"},
			},
			"egress_burst_size_bytes": {
				Type:        schema.TypeString, // string + validation due to usual problem of differentiation between 0 and empty value for TypeInt
				Optional:    true,
				Description: fmt.Sprintf("Ingress burst size bytes for %s", labelVcfaEdgeCluster),
				ValidateDiagFunc: validation.AnyDiag(
					validation.ToDiagFunc(validation.StringIsEmpty),
					validation.ToDiagFunc(IsIntAndAtLeast(1))),
				RequiredWith: []string{"egress_committed_bandwidth_mbps"},
			},
		},
	}
}

func resourceVcfaEdgeClusterQosCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// The Edge Cluster is already existing that is handled by 'vcfa_edge_cluster' data source.
	// This is not a "real" entity creation, rather a lookup and update of existing one
	createQosConfigInEdgeCluster := func(config *types.TmEdgeCluster) (*govcd.TmEdgeCluster, error) {
		ec, err := vcdClient.GetTmEdgeClusterById(d.Get("edge_cluster_id").(string))
		if err != nil {
			return nil, fmt.Errorf("error looking up %s by ID: %s", labelVcfaEdgeCluster, err)
		}
		return ec.Update(config)
	}

	c := crudConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:      labelVcfaEdgeClusterQos,
		getTypeFunc:      getTmEdgeClusterQosType,
		stateStoreFunc:   setTmEdgeClusterQosData,
		createFunc:       createQosConfigInEdgeCluster,
		resourceReadFunc: resourceVcfaEdgeClusterQosRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaEdgeClusterQosUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:      labelVcfaEdgeClusterQos,
		getTypeFunc:      getTmEdgeClusterQosType,
		getEntityFunc:    vcdClient.GetTmEdgeClusterById,
		resourceReadFunc: resourceVcfaEdgeClusterQosRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaEdgeClusterQosRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:    labelVcfaEdgeClusterQos,
		getEntityFunc:  vcdClient.GetTmEdgeClusterById,
		stateStoreFunc: setTmEdgeClusterQosData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaEdgeClusterQosDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:   labelVcfaEdgeClusterQos,
		getEntityFunc: vcdClient.GetTmEdgeClusterById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaEdgeClusterQosImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as region-name.edge-cluster-name")
	}
	regionName, edgeClusterName := resourceURI[0], resourceURI[1]

	region, err := vcdClient.GetRegionByName(regionName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by name '%s': %s", labelVcfaRegion, regionName, err)
	}

	ec, err := vcdClient.GetTmEdgeClusterByNameAndRegionId(edgeClusterName, region.Region.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by Name '%s' in %s '%s': %s",
			labelVcfaEdgeClusterQos, edgeClusterName, labelVcfaRegion, regionName, err)
	}

	d.SetId(ec.TmEdgeCluster.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmEdgeClusterQosType(vcdClient *VCDClient, d *schema.ResourceData) (*types.TmEdgeCluster, error) {
	// Only the QoS configuration is updatable, everything else is read-only
	t := &types.TmEdgeCluster{DefaultQosConfig: types.TmEdgeClusterDefaultQosConfig{}}

	// Ingress setup
	// Only initialize IngressProfile type if at least one of the fields is set
	ingressCommittedBandwidthMbps := d.Get("ingress_committed_bandwidth_mbps").(string)
	ingressBurstSizeBytes := d.Get("ingress_burst_size_bytes").(string)
	if ingressCommittedBandwidthMbps != "" || ingressBurstSizeBytes != "" {
		t.DefaultQosConfig.IngressProfile = &types.TmEdgeClusterQosProfile{Type: "DEFAULT"}

		if ingressCommittedBandwidthMbps != "" {
			t.DefaultQosConfig.IngressProfile.CommittedBandwidthMbps = mustStrToInt(ingressCommittedBandwidthMbps)
		}

		if ingressBurstSizeBytes != "" {
			t.DefaultQosConfig.IngressProfile.BurstSizeBytes = mustStrToInt(ingressBurstSizeBytes)
		}
	}

	// Egress setup
	// Only initialize EgressProfile type if at least one of the fields is set
	egressCommittedBandwidthMbps := d.Get("egress_committed_bandwidth_mbps").(string)
	egressBurstSizeBytes := d.Get("egress_burst_size_bytes").(string)
	if egressCommittedBandwidthMbps != "" || egressBurstSizeBytes != "" {
		t.DefaultQosConfig.EgressProfile = &types.TmEdgeClusterQosProfile{Type: "DEFAULT"}

		if egressCommittedBandwidthMbps != "" {
			t.DefaultQosConfig.EgressProfile.CommittedBandwidthMbps = mustStrToInt(egressCommittedBandwidthMbps)
		}

		if egressBurstSizeBytes != "" {
			t.DefaultQosConfig.EgressProfile.BurstSizeBytes = mustStrToInt(egressBurstSizeBytes)
		}
	}

	return t, nil
}

func setTmEdgeClusterQosData(_ *VCDClient, d *schema.ResourceData, t *govcd.TmEdgeCluster) error {
	if t == nil || t.TmEdgeCluster == nil {
		return fmt.Errorf("empty %s received", labelVcfaEdgeCluster)
	}

	d.SetId(t.TmEdgeCluster.ID)
	dSet(d, "edge_cluster_id", t.TmEdgeCluster.ID)

	dSet(d, "region_id", "")
	if t.TmEdgeCluster.RegionRef != nil {
		dSet(d, "region_id", t.TmEdgeCluster.RegionRef.ID)
	}

	dSet(d, "ingress_committed_bandwidth_mbps", nil)
	dSet(d, "ingress_burst_size_bytes", nil)
	if t.TmEdgeCluster.DefaultQosConfig.IngressProfile != nil {
		strValue := strconv.Itoa(t.TmEdgeCluster.DefaultQosConfig.IngressProfile.BurstSizeBytes)
		dSet(d, "ingress_burst_size_bytes", strValue)

		strValueCommitted := strconv.Itoa(t.TmEdgeCluster.DefaultQosConfig.IngressProfile.CommittedBandwidthMbps)
		dSet(d, "ingress_committed_bandwidth_mbps", strValueCommitted)
	}

	dSet(d, "egress_committed_bandwidth_mbps", nil)
	dSet(d, "egress_burst_size_bytes", nil)
	if t.TmEdgeCluster.DefaultQosConfig.EgressProfile != nil {
		strValueCommitted := strconv.Itoa(t.TmEdgeCluster.DefaultQosConfig.EgressProfile.CommittedBandwidthMbps)
		dSet(d, "egress_committed_bandwidth_mbps", strValueCommitted)

		strValueBurstSize := strconv.Itoa(t.TmEdgeCluster.DefaultQosConfig.EgressProfile.BurstSizeBytes)
		dSet(d, "egress_burst_size_bytes", strValueBurstSize)
	}

	return nil
}
