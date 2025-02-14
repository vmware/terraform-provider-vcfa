package vcfa

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaSupervisor() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaSupervisorRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Supervisor",
			},
			"vcenter_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Parent vCenter ID",
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Parent Region ID",
			},
		},
	}
}

func datasourceVcfaSupervisorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	s, err := tmClient.GetSupervisorByNameAndVcenterId(d.Get("name").(string), d.Get("vcenter_id").(string))
	if err != nil {
		return diag.Errorf("error getting Supervisor: %s", err)
	}

	err = setSupervisorData(d, s.Supervisor)
	if err != nil {
		return diag.Errorf("error storing Supervisor data: %s", err)
	}

	d.SetId(s.Supervisor.SupervisorID)

	return nil
}

func setSupervisorData(d *schema.ResourceData, s *types.Supervisor) error {
	vCenterId := ""
	if s.VirtualCenter != nil {
		vCenterId = s.VirtualCenter.ID
	}
	dSet(d, "vcenter_id", vCenterId)

	regionId := ""
	if s.Region != nil {
		regionId = s.Region.ID
	}

	dSet(d, "region_id", regionId)

	return nil
}
