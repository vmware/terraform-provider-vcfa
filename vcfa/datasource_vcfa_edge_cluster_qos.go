package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaEdgeClusterQos() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaEdgeClusterQosRead,

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
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Ingress committed bandwidth in Mbps for %s", labelVcfaEdgeCluster),
			},
			"ingress_burst_size_bytes": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Ingress burst size bytes for %s", labelVcfaEdgeCluster),
			},
			"egress_committed_bandwidth_mbps": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Egress committed bandwidth in Mbps for %s", labelVcfaEdgeCluster),
			},
			"egress_burst_size_bytes": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Egress burst size bytes for %s", labelVcfaEdgeCluster),
			},
		},
	}
}

func datasourceVcfaEdgeClusterQosRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	c := dsReadConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:              labelVcfaEdgeClusterQos,
		stateStoreFunc:           setTmEdgeClusterQosData,
		overrideDefaultNameField: "edge_cluster_id", // pass the value of this field to getEntityFunc
		getEntityFunc:            vcdClient.GetTmEdgeClusterById,
	}
	return readDatasource(ctx, d, meta, c)
}
