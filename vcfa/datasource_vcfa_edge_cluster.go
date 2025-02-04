package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaEdgeCluster = "Edge Cluster"
const labelVcfaEdgeClusterSync = "Edge Cluster Sync"

func datasourceVcfaEdgeCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaEdgeClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name %s", labelVcfaEdgeCluster),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Region ID of  %s", labelVcfaEdgeCluster),
			},
			"sync_before_read": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: fmt.Sprintf("Will trigger SYNC operation before looking for a given %s", labelVcfaEdgeCluster),
			},
			"node_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Node count in %s", labelVcfaEdgeCluster),
			},
			"org_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Org count %s", labelVcfaEdgeCluster),
			},
			"vpc_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("VPC count %s", labelVcfaEdgeCluster),
			},
			"average_cpu_usage_percentage": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: fmt.Sprintf("Average CPU Usage percentage of %s ", labelVcfaEdgeCluster),
			},
			"average_memory_usage_percentage": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: fmt.Sprintf("Average Memory Usage percentage of %s ", labelVcfaEdgeCluster),
			},
			"health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Health status of %s", labelVcfaEdgeCluster),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaEdgeCluster),
			},
			"deployment_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Deployment type of %s", labelVcfaEdgeCluster),
			},
		},
	}
}

func datasourceVcfaEdgeClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	regionId := d.Get("region_id").(string)
	getByName := func(name string) (*govcd.TmEdgeCluster, error) {
		return tmClient.GetTmEdgeClusterByNameAndRegionId(name, regionId)
	}

	c := dsReadConfig[*govcd.TmEdgeCluster, types.TmEdgeCluster]{
		entityLabel:    labelVcfaEdgeCluster,
		getEntityFunc:  getByName,
		stateStoreFunc: setTmEdgeClusterData,
		preReadHooks:   []schemaHook{syncTmEdgeClustersBeforeReadHook},
	}
	return readDatasource(ctx, d, meta, c)
}

func setTmEdgeClusterData(_ *VCDClient, d *schema.ResourceData, t *govcd.TmEdgeCluster) error {
	if t == nil || t.TmEdgeCluster == nil {
		return fmt.Errorf("empty %s received", labelVcfaEdgeCluster)
	}

	d.SetId(t.TmEdgeCluster.ID)
	dSet(d, "status", t.TmEdgeCluster.Status)
	dSet(d, "health_status", t.TmEdgeCluster.HealthStatus)

	dSet(d, "region_id", "")
	if t.TmEdgeCluster.RegionRef != nil {
		dSet(d, "region_id", t.TmEdgeCluster.RegionRef.ID)
	}
	dSet(d, "deployment_type", t.TmEdgeCluster.DeploymentType)
	dSet(d, "node_count", t.TmEdgeCluster.NodeCount)
	dSet(d, "org_count", t.TmEdgeCluster.OrgCount)
	dSet(d, "vpc_count", t.TmEdgeCluster.VpcCount)
	dSet(d, "average_cpu_usage_percentage", t.TmEdgeCluster.AvgCPUUsagePercentage)
	dSet(d, "average_memory_usage_percentage", t.TmEdgeCluster.AvgMemoryUsagePercentage)

	return nil
}

func syncTmEdgeClustersBeforeReadHook(tmClient *VCDClient, d *schema.ResourceData) error {
	if d.Get("sync_before_read").(bool) {
		err := tmClient.TmSyncEdgeClusters()
		if err != nil {
			return fmt.Errorf("error syncing %s before lookup: %s", labelVcfaEdgeClusterSync, err)
		}
	}
	return nil
}
