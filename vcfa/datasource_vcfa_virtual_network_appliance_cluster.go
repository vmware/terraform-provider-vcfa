// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaVirtualNetworkApplianceCluster = "Virtual Network Appliance Cluster"

func datasourceVcfaVirtualNetworkApplianceCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaVirtualNetworkApplianceClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaVirtualNetworkApplianceCluster),
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf("Region ID of %s. If set, it must match the Region of the found %s", labelVcfaVirtualNetworkApplianceCluster, labelVcfaVirtualNetworkApplianceCluster),
			},
			"sync_before_read": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: fmt.Sprintf("Will trigger SYNC operation before looking for a given %s", labelVcfaVirtualNetworkApplianceCluster),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaVirtualNetworkApplianceCluster),
			},
			"node_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of Virtual Network Appliances in %s", labelVcfaVirtualNetworkApplianceCluster),
			},
			"org_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Number of Organizations using %s", labelVcfaVirtualNetworkApplianceCluster),
			},
			"health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Health status of %s", labelVcfaVirtualNetworkApplianceCluster),
			},
			"service_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Service type of %s", labelVcfaVirtualNetworkApplianceCluster),
			},
		},
	}
}

func datasourceVcfaVirtualNetworkApplianceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	regionId := d.Get("region_id").(string)
	getByName := func(name string) (*govcd.TmVirtualNetworkApplianceCluster, error) {
		cluster, err := tmClient.GetTmVirtualNetworkApplianceClusterByName(name)
		if err != nil {
			return nil, err
		}
		if foundRegionId := refId(cluster.TmVirtualNetworkApplianceCluster.RegionRef); regionId != "" && foundRegionId != regionId {
			return nil, fmt.Errorf("%s '%s' belongs to Region '%s', not to the requested Region '%s'",
				labelVcfaVirtualNetworkApplianceCluster, name, foundRegionId, regionId)
		}
		return cluster, nil
	}

	c := dsReadConfig[*govcd.TmVirtualNetworkApplianceCluster, types.TmVirtualNetworkApplianceCluster]{
		entityLabel:    labelVcfaVirtualNetworkApplianceCluster,
		getEntityFunc:  getByName,
		stateStoreFunc: setTmVirtualNetworkApplianceClusterData,
		preReadHooks:   []schemaHook{syncTmVirtualNetworkApplianceClustersBeforeReadHook},
	}
	return readDatasource(ctx, d, meta, c)
}

func setTmVirtualNetworkApplianceClusterData(_ *VCDClient, d *schema.ResourceData, t *govcd.TmVirtualNetworkApplianceCluster) error {
	if t == nil || t.TmVirtualNetworkApplianceCluster == nil {
		return fmt.Errorf("empty %s received", labelVcfaVirtualNetworkApplianceCluster)
	}
	v := t.TmVirtualNetworkApplianceCluster

	d.SetId(v.ID)
	dSet(d, "name", v.Name)
	dSet(d, "description", v.Description)
	dSet(d, "region_id", refId(v.RegionRef))
	dSet(d, "node_count", v.NodeCount)
	dSet(d, "org_count", v.OrgCount)
	dSet(d, "health_status", v.HealthStatus)
	dSet(d, "service_type", v.ServiceType)

	return nil
}

func syncTmVirtualNetworkApplianceClustersBeforeReadHook(tmClient *VCDClient, d *schema.ResourceData) error {
	if d.Get("sync_before_read").(bool) {
		err := tmClient.SyncTmVirtualNetworkApplianceClusters()
		if err != nil {
			return fmt.Errorf("error syncing %s before lookup: %s", labelVcfaVirtualNetworkApplianceCluster, err)
		}
	}
	return nil
}
