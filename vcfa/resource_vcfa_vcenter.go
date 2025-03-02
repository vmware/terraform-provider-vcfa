package vcfa

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

const labelVcfaVirtualCenter = "vCenter Server"
const extraSleepAfterListenerConnected = 3 * time.Second

// vCenter task is sometimes unreliable and trying to refresh it immediately after it becomes
// connected causes a "BUSY_ENTITY" error (which has a few different messages)
var maximumVcenterRetryTime = 120 * time.Second                                         // The maximum time a single operation will be retried before giving up
var vCenterEntityBusyRegexp = regexp.MustCompile(`(is currently busy|400|BUSY_ENTITY)`) // Regexp to match entity busy error

func resourceVcfaVcenter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaVcenterCreate,
		ReadContext:   resourceVcfaVcenterRead,
		UpdateContext: resourceVcfaVcenterUpdate,
		DeleteContext: resourceVcfaVcenterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaVcenterImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaVirtualCenter),
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("URL including port of %s", labelVcfaVirtualCenter),
			},
			"auto_trust_certificate": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Defines if the %s certificate should automatically be trusted", labelVcfaVirtualCenter),
			},
			"nsx_manager_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("%s ID that this %s uses", labelVcfaNsxManager, labelVcfaVirtualCenter),
			},
			"refresh_vcenter_on_create": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: fmt.Sprintf("Defines if the %s should be refreshed after creation", labelVcfaVirtualCenter),
			},
			"refresh_policies_on_create": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: fmt.Sprintf("Defines if the %s should refresh Policies after creation", labelVcfaVirtualCenter),
			},
			"refresh_vcenter_on_read": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: fmt.Sprintf("Defines if the %s should be refreshed on every read operation", labelVcfaVirtualCenter),
			},
			"refresh_policies_on_read": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: fmt.Sprintf("Defines if the %s should refresh Policies on every read operation", labelVcfaVirtualCenter),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Username of %s", labelVcfaVirtualCenter),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: fmt.Sprintf("Password of %s", labelVcfaVirtualCenter),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: fmt.Sprintf("Should the %s be enabled", labelVcfaVirtualCenter),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaVirtualCenter),
			},
			"has_proxy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s has proxy defined", labelVcfaVirtualCenter),
			},
			"is_connected": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s is connected", labelVcfaVirtualCenter),
			},
			"mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelVcfaVirtualCenter),
			},
			"connection_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Listener state of %s", labelVcfaVirtualCenter),
			},
			"cluster_health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelVcfaVirtualCenter),
			},
			"vcenter_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Version of %s", labelVcfaVirtualCenter),
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s UUID", labelVcfaVirtualCenter),
			},
			"vcenter_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s hostname", labelVcfaVirtualCenter),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter status",
			},
		},
	}
}

func getVcenterType(_ *VCDClient, d *schema.ResourceData) (*types.VSphereVirtualCenter, error) {
	t := &types.VSphereVirtualCenter{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Url:         d.Get("url").(string),
		Username:    d.Get("username").(string),
		Password:    d.Get("password").(string),
		IsEnabled:   d.Get("is_enabled").(bool),
	}

	return t, nil
}

func setVcenterData(_ *VCDClient, d *schema.ResourceData, v *govcd.VCenter) error {
	if v == nil || v.VSphereVCenter == nil {
		return fmt.Errorf("nil object for %s", labelVcfaVirtualCenter)
	}

	dSet(d, "name", v.VSphereVCenter.Name)
	dSet(d, "description", v.VSphereVCenter.Description)
	dSet(d, "url", v.VSphereVCenter.Url)
	dSet(d, "username", v.VSphereVCenter.Username)
	// dSet(d, "password", v.VSphereVCenter.Password) // password is never returned,
	dSet(d, "is_enabled", v.VSphereVCenter.IsEnabled)

	dSet(d, "has_proxy", v.VSphereVCenter.HasProxy)
	dSet(d, "is_connected", v.VSphereVCenter.IsConnected)
	dSet(d, "mode", v.VSphereVCenter.Mode)
	dSet(d, "connection_status", v.VSphereVCenter.ListenerState)
	dSet(d, "cluster_health_status", v.VSphereVCenter.ClusterHealthStatus)
	dSet(d, "vcenter_version", v.VSphereVCenter.VcVersion)
	dSet(d, "uuid", v.VSphereVCenter.Uuid)
	host, err := url.Parse(v.VSphereVCenter.Url)
	if err != nil {
		return fmt.Errorf("error parsing URL for storing 'vcenter_host': %s", err)
	}
	dSet(d, "vcenter_host", host.Host)

	// Status is a derivative value that was present in XML Query API, but is no longer maintained
	// The value was derived from multiple fields based on a complex logic. Instead, evaluating if
	// vCenter is ready for operations, would be to rely on `is_enabled`, `is_connected` and
	// optionally `cluster_health_status` fields.
	//
	// The `status` is a rough approximation of this value
	dSet(d, "status", "NOT_READY")
	if v.VSphereVCenter.IsConnected && v.VSphereVCenter.ListenerState == "CONNECTED" {
		dSet(d, "status", "READY")
	}

	d.SetId(v.VSphereVCenter.VcId)

	return nil
}

func resourceVcfaVcenterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	shouldRefresh := d.Get("refresh_vcenter_on_create").(bool)
	shouldRefreshPolicies := d.Get("refresh_policies_on_create").(bool)
	shouldWaitForListenerStatus := true
	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:    labelVcfaVirtualCenter,
		getTypeFunc:    getVcenterType,
		stateStoreFunc: setVcenterData,
		// createAsyncFunc:  tmClient.CreateVcenterAsync,
		createFunc: tmClient.CreateVcenter,
		// getEntityFunc:    tmClient.GetVCenterById,
		resourceReadFunc: resourceVcfaVcenterRead,
		// certificate should be trusted for the vCenter to work
		preCreateHooks: []schemaHook{autoTrustHostCertificate("url", "auto_trust_certificate")},
		postCreateHooks: []outerEntityHook[*govcd.VCenter]{
			// TODO: TM ensure that the vCenter listener state is "CONNECTED"  before triggering
			// refresh as it will fail otherwise. At the moment it has a delay before it becomes
			// CONNECTED after creation task succeeds. It should not be needed once vCenter creation
			// task ensures that the listener is connected.
			shouldWaitForListenerStatusConnected(shouldWaitForListenerStatus),

			refreshVcenter(shouldRefresh),               // vCenter read can optionally trigger "refresh" operation
			refreshVcenterPolicy(shouldRefreshPolicies), // vCenter read can optionally trigger "refresh policies" operation
		},
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaVcenterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// return immediately if only flags are updated
	if !d.HasChangesExcept("refresh_vcenter_on_read", "refresh_policies_on_read") {
		return nil
	}

	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:      labelVcfaVirtualCenter,
		getTypeFunc:      getVcenterType,
		getEntityFunc:    tmClient.GetVCenterById,
		resourceReadFunc: resourceVcfaVcenterRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaVcenterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	// prefetch vCenter so that vc.VSphereVCenter.IsEnabled and vc.VSphereVCenter.IsConnected flags
	// can be verified and avoid triggering refreshes if VC is disconnected
	vc, err := tmClient.GetVCenterById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
		}
		return diag.Errorf("error retrieving vCenter by Id: %s", err)
	}

	// TODO: TM: remove this block and use the commented one within crudConfig below.
	// Retrieval endpoints by Name and by ID return differently formated url (the by Id one returns
	// URL with port http://host:443, while the one by name - doesn't). Using the same getByName to
	// match format everywhere
	fakeGetById := func(_ string) (*govcd.VCenter, error) {
		return tmClient.GetVCenterByName(vc.VSphereVCenter.Name)
	}

	shouldRefresh := d.Get("refresh_vcenter_on_read").(bool)
	shouldRefreshPolicies := d.Get("refresh_policies_on_read").(bool)
	shouldWaitForListenerStatus := true

	// There is no way to detect if a resource is 'tainted' ('d.State().Tainted' is not reliable),
	// but if a resource is not connected and is not enabled - there is no point in refreshing
	// anything
	// It will help in the case when invalid configuration is supplied and a creation task fails as
	// a tainted resource has to be read before releasing it
	if !vc.VSphereVCenter.IsEnabled && !vc.VSphereVCenter.IsConnected {
		shouldRefresh = false
		shouldRefreshPolicies = false
		shouldWaitForListenerStatus = false
	}
	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel: labelVcfaVirtualCenter,
		// getEntityFunc:  tmClient.GetVCenterById,// TODO: TM: use this function
		getEntityFunc:  fakeGetById, // TODO: TM: remove this function
		stateStoreFunc: setVcenterData,
		readHooks: []outerEntityHook[*govcd.VCenter]{
			// TODO: TM ensure that the vCenter listener state is "CONNECTED"  before triggering
			// refresh as it will fail otherwise. At the moment it has a delay before it becomes
			// CONNECTED after creation task succeeds. It should not be needed once vCenter creation
			// task ensures that the listener is connected.
			shouldWaitForListenerStatusConnected(shouldWaitForListenerStatus),

			refreshVcenter(shouldRefresh),               // vCenter read can optionally trigger "refresh" operation
			refreshVcenterPolicy(shouldRefreshPolicies), // vCenter read can optionally trigger "refresh policies" operation
		},
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaVcenterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:    labelVcfaVirtualCenter,
		getEntityFunc:  tmClient.GetVCenterById,
		preDeleteHooks: []outerEntityHook[*govcd.VCenter]{disableVcenter}, // vCenter must be disabled before deletion
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaVcenterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	v, err := tmClient.GetVCenterByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by name: %s", labelVcfaVirtualCenter, err)
	}

	d.SetId(v.VSphereVCenter.VcId)
	return []*schema.ResourceData{d}, nil
}

// disableVcenter disables vCenter which is usefull before deletion as a non-disabled vCenter cannot
// be removed
func disableVcenter(v *govcd.VCenter) error {
	if v.VSphereVCenter.IsEnabled {
		return v.Disable()
	}
	return nil
}

// refreshVcenter triggers refresh on vCenter which is useful for reloading some of the vCenter
// components like Supervisors
func refreshVcenter(execute bool) outerEntityHook[*govcd.VCenter] {
	return func(v *govcd.VCenter) error {
		if execute {
			err := runWithRetry(v.RefreshVcenter, vCenterEntityBusyRegexp, maximumVcenterRetryTime)
			if err != nil {
				return fmt.Errorf("error refreshing vCenter: %s", err)
			}
		}
		return nil
	}
}

// refreshVcenterPolicy triggers refresh on vCenter which is useful for reloading some of the
// vCenter components like Supervisors
func refreshVcenterPolicy(execute bool) outerEntityHook[*govcd.VCenter] {
	return func(v *govcd.VCenter) error {
		if execute {
			err := runWithRetry(v.RefreshStorageProfiles, vCenterEntityBusyRegexp, maximumVcenterRetryTime)
			if err != nil {
				return fmt.Errorf("error refreshing vCenter Storage Policies: %s", err)
			}
		}
		return nil
	}
}

// TODO: TM: should not be required because a successful vCenter creation task should work
func shouldWaitForListenerStatusConnected(shouldWait bool) func(v *govcd.VCenter) error {
	return func(v *govcd.VCenter) error {
		if !shouldWait {
			return nil
		}
		for c := 0; c < 20; c++ {
			err := v.Refresh()
			if err != nil {
				return fmt.Errorf("error refreshing vCenter: %s", err)
			}

			if v.VSphereVCenter.ListenerState == "CONNECTED" {
				// TODO: TM: put an extra sleep to be sure the entity is released
				time.Sleep(extraSleepAfterListenerConnected)

				return nil
			}

			time.Sleep(2 * time.Second)
		}

		return fmt.Errorf("failed waiting for listener state to become 'CONNECTED', got '%s'", v.VSphereVCenter.ListenerState)
	}
}

// autoTrustHostCertificate can automatically add host certificate to trusted ones
// * urlSchemaFieldName - Terraform schema field (TypeString) name that contains URL of entity
// * trustSchemaFieldName - Terraform schema field (TypeBool) name that defines if the certificate should be trusted
// Note. It will not add new entry if the certificate is already trusted
func autoTrustHostCertificate(urlSchemaFieldName, trustSchemaFieldName string) schemaHook {
	return func(tmClient *VCDClient, d *schema.ResourceData) error {
		shouldExecute := d.Get(trustSchemaFieldName).(bool)
		if !shouldExecute {
			util.Logger.Printf("[DEBUG] Skipping certificate trust execution as '%s' is false", trustSchemaFieldName)
			return nil
		}
		schemaUrl := d.Get(urlSchemaFieldName).(string)
		parsedUrl, err := url.Parse(schemaUrl)
		if err != nil {
			return fmt.Errorf("error parsing provided url '%s': %s", schemaUrl, err)
		}

		_, err = tmClient.AutoTrustHttpsCertificate(parsedUrl, nil)
		if err != nil {
			return fmt.Errorf("error trusting '%s' certificate: %s", schemaUrl, err)
		}

		return nil
	}
}

func runWithRetry(runOperation func() error, errRegexp *regexp.Regexp, duration time.Duration) error {
	startTime := time.Now()
	endTime := startTime.Add(duration)
	util.Logger.Printf("[DEBUG] runWithRetry - running with retry for %f seconds if error contains '%s' ", duration.Seconds(), errRegexp)
	count := 1
	for {
		err := runOperation()
		util.Logger.Printf("[DEBUG] runWithRetry - ran attempt %d, got error: %s ", count, err)
		// Operation had no error - it succeeded
		if err == nil {
			util.Logger.Printf("[DEBUG] runWithRetry - no error occurred after attempt %d, got error: %s ", count, err)
			return nil
		}
		// If there is an error, but it doesn't contain the retryIfErrContains value - exit it
		if !errRegexp.MatchString(err.Error()) {
			util.Logger.Printf("[DEBUG] runWithRetry - returning error after attempt %d, got error: %s ", count, err)
			return err
		}

		// If time limit is exceeded - return error containing statistics and original error
		if time.Now().After(endTime) {
			util.Logger.Printf("[DEBUG] runWithRetry - exceeded time after attempt %d, got error: %s ", count, err)
			return fmt.Errorf("error attempting to wait until error does not contain '%s' after %f seconds: %s", errRegexp, duration.Seconds(), err)
		}

		// Sleep and continue
		util.Logger.Printf("[DEBUG] runWithRetry - sleeping after attempt %d, will retry", count)
		// Sleep 2 seconds and attempt once more if the timeout is not excdeeded
		time.Sleep(2 * time.Second)
		count++
	}
}
