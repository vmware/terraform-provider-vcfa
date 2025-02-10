package vcfa

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// TODO: TM:
// - Data source
// - Test
// - Documentation

const labelVcfaOrgRegionQuotaStoragePolicy = "Region Quota Storage Policy"

func resourceVcfaOrgRegionQuotaStoragePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaOrgRegionQuotaStoragePolicyCreate,
		ReadContext:   resourceVcfaOrgRegionQuotaStoragePolicyRead,
		UpdateContext: resourceVcfaOrgRegionQuotaStoragePolicyUpdate,
		DeleteContext: resourceVcfaOrgRegionQuotaStoragePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgRegionQuotaStoragePolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"org_region_quota_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s ID", labelVcfaOrgRegionQuota),
			},
			"region_storage_policy_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("The parent %s for this %s", labelVcfaRegionStoragePolicy, labelVcfaOrgRegionQuotaStoragePolicy),
			},
			"storage_limit_mib": {
				Type:             schema.TypeInt,
				Required:         true,
				Description:      "Maximum allowed storage allocation in mebibytes. Minimum value: 0",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The name of the %s. It follows RFC 1123 Label Names to conform with Kubernetes standards", labelVcfaOrgRegionQuotaStoragePolicy),
			},
			"storage_used_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Amount of storage used in mebibytes",
			},
		},
	}
}

func resourceVcfaOrgRegionQuotaStoragePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	regionQuotaId := d.Get("org_region_quota_id").(string)
	// TODO: TM: Check if the lock is really needed (it is a different endpoint and these entities have an UUID on their own)
	vcfa.kvLock(regionQuotaId)
	defer vcfa.kvUnlock(regionQuotaId)

	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.RegionQuotaStoragePolicy, types.VirtualDatacenterStoragePolicy]{
		entityLabel:    labelVcfaOrgRegionQuotaStoragePolicy,
		getTypeFunc:    getOrgRegionQuotaStoragePolicyType,
		stateStoreFunc: setOrgRegionQuotaStoragePolicyData,
		createFunc: func(config *types.VirtualDatacenterStoragePolicy) (*govcd.RegionQuotaStoragePolicy, error) {
			rq, err := tmClient.GetRegionQuotaById(regionQuotaId)
			if err != nil {
				return nil, err
			}
			sps, err := rq.CreateStoragePolicies(&types.VirtualDatacenterStoragePolicies{Values: []types.VirtualDatacenterStoragePolicy{*config}})
			if err != nil {
				return nil, err
			}
			if len(sps) != 1 {
				return nil, fmt.Errorf("expected 1 %s after creation, received %d", labelVcfaOrgRegionQuotaStoragePolicy, len(sps))
			}
			return sps[0], nil
		},
		resourceReadFunc: resourceVcfaOrgRegionQuotaStoragePolicyRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionQuotaStoragePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	regionQuotaId := d.Get("org_region_quota_id").(string)
	// TODO: TM: Check if the lock is really needed (it is a different endpoint and these entities have an UUID on their own)
	vcfa.kvLock(regionQuotaId)
	defer vcfa.kvUnlock(regionQuotaId)

	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.RegionQuotaStoragePolicy, types.VirtualDatacenterStoragePolicy]{
		entityLabel:      labelVcfaOrgRegionQuotaStoragePolicy,
		getTypeFunc:      getOrgRegionQuotaStoragePolicyType,
		getEntityFunc:    tmClient.GetRegionQuotaStoragePolicyById,
		resourceReadFunc: resourceVcfaOrgRegionQuotaStoragePolicyRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionQuotaStoragePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.RegionQuotaStoragePolicy, types.VirtualDatacenterStoragePolicy]{
		entityLabel:    labelVcfaOrgRegionQuotaStoragePolicy,
		getEntityFunc:  tmClient.GetRegionQuotaStoragePolicyById,
		stateStoreFunc: setOrgRegionQuotaStoragePolicyData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionQuotaStoragePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	regionQuotaId := d.Get("org_region_quota_id").(string)
	// TODO: TM: Check if the lock is really needed (it is a different endpoint and these entities have an UUID on their own)
	vcfa.kvLock(regionQuotaId)
	defer vcfa.kvUnlock(regionQuotaId)

	tmClient := meta.(ClientContainer).tmClient

	c := crudConfig[*govcd.RegionQuotaStoragePolicy, types.VirtualDatacenterStoragePolicy]{
		entityLabel:   labelVcfaOrgRegionQuotaStoragePolicy,
		getEntityFunc: tmClient.GetRegionQuotaStoragePolicyById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionQuotaStoragePolicyImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	idSlice := strings.Split(d.Id(), ImportSeparator)
	if len(idSlice) != 3 {
		return nil, fmt.Errorf("expected import ID to be <org name>%s<region name>%s<policy name>", ImportSeparator, ImportSeparator)
	}

	rq, err := tmClient.GetRegionQuotaByName(fmt.Sprintf("%s_%s", idSlice[0], idSlice[1]))
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s: %s", labelVcfaOrgRegionQuota, err)
	}

	policy, err := rq.GetStoragePolicyByName(idSlice[2])
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s: %s", labelVcfaOrgRegionQuotaStoragePolicy, err)
	}

	d.SetId(policy.VirtualDatacenterStoragePolicy.ID)
	return []*schema.ResourceData{d}, nil
}

func getOrgRegionQuotaStoragePolicyType(_ *VCDClient, d *schema.ResourceData) (*types.VirtualDatacenterStoragePolicy, error) {
	t := &types.VirtualDatacenterStoragePolicy{
		RegionStoragePolicy: types.OpenApiReference{
			ID: d.Get("region_storage_policy_id").(string),
		},
		StorageLimitMiB: int64(d.Get("storage_limit_mib").(int)),
		VirtualDatacenter: types.OpenApiReference{
			ID: d.Get("org_region_quota_id").(string),
		},
	}
	return t, nil
}

func setOrgRegionQuotaStoragePolicyData(_ *VCDClient, d *schema.ResourceData, rqSp *govcd.RegionQuotaStoragePolicy) error {
	if rqSp == nil || rqSp.VirtualDatacenterStoragePolicy == nil {
		return fmt.Errorf("provided %s is nil", labelVcfaOrgRegionQuotaStoragePolicy)
	}

	d.SetId(rqSp.VirtualDatacenterStoragePolicy.ID)
	dSet(d, "region_storage_policy_id", rqSp.VirtualDatacenterStoragePolicy.RegionStoragePolicy.ID) // Can't be nil
	dSet(d, "org_region_quota_id", rqSp.VirtualDatacenterStoragePolicy.VirtualDatacenter.ID)        // Can't be nil
	dSet(d, "storage_limit_mib", rqSp.VirtualDatacenterStoragePolicy.StorageLimitMiB)
	dSet(d, "storage_used_mib", rqSp.VirtualDatacenterStoragePolicy.StorageUsedMiB)
	dSet(d, "name", rqSp.VirtualDatacenterStoragePolicy.Name)
	return nil
}
