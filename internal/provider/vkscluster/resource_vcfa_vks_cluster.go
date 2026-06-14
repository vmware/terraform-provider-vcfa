// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkscluster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	jsonpatch "gopkg.in/evanphx/json-patch.v4"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/common"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

const (
	vksClusterCreateDefaultTimeout  = 30 * time.Minute
	vksClusterUpdateDefaultTimeout  = 30 * time.Minute
	vksClusterDeleteDefaultTimeout  = 10 * time.Minute
	vksClusterPollInterval          = 5 * time.Second
	vksClusterConflictMaxRetries    = 5
	vksClusterConflictRetryInterval = 2 * time.Second
)

var (
	_ resource.Resource                   = (*vcfaVksClusterResource)(nil)
	_ resource.ResourceWithConfigure      = (*vcfaVksClusterResource)(nil)
	_ resource.ResourceWithImportState    = (*vcfaVksClusterResource)(nil)
	_ resource.ResourceWithValidateConfig = (*vcfaVksClusterResource)(nil)
	_ resource.ResourceWithModifyPlan     = (*vcfaVksClusterResource)(nil)
)

type vcfaVksClusterResource struct {
	tmClient *vcfa.VCDClient
}

func NewVcfaVksClusterResource() resource.Resource {
	return &vcfaVksClusterResource{}
}

func (r *vcfaVksClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vks_cluster"
}

func (r *vcfaVksClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	tmClient, err := helpers.GetTmClientFromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("error retrieving TM client from provider data", err.Error())
		return
	}
	r.tmClient = tmClient
}

func (r *vcfaVksClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vcfaVksClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vcfContext := common.ExtractVcfContext(ctx, plan.Context, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, vksClusterCreateDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	waitForAvailable, diags := r.extractWaitForAvailable(ctx, plan.WaitFor)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := vcfContext.Project.ValueString()
	namespace := vcfContext.Namespace.ValueString()
	name := plan.Name.ValueString()

	k8sClient, err := kubernetes.NewClient(r.tmClient, project, namespace)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error creating %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("error creating Kubernetes client for VCF context %s/%s: %s", project, namespace, err.Error()),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(k8sClient.FlushWarnings()...) }()

	clusterObj := mapResourceModelToVksCluster(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var created vcfatypes.VksCluster
	if err := k8sClient.CreateNamespaceScopedResource(ctx, vcfatypes.GetVksClusterGVR(), namespace, clusterObj, &created, false); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error creating %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("could not create %s %s in VCF context %s/%s: %s", vcfatypes.LabelVksCluster, name, project, namespace, err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", project, namespace, name))

	if waitForAvailable {
		if err := r.waitForClusterAvailable(ctx, k8sClient, project, namespace, name, createTimeout); err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s %s created but not yet available", vcfatypes.LabelVksCluster, name),
				fmt.Sprintf("%s %s in VCF context %s/%s was created but did not reach available state within the timeout: %s", vcfatypes.LabelVksCluster, name, project, namespace, err.Error()),
			)
		}
	}

	helpers.SanitizeUnknownForState(ctx, reflect.ValueOf(&plan).Elem())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vcfaVksClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vcfaVksClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vcfContext := common.ExtractVcfContext(ctx, state.Context, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	project := vcfContext.Project.ValueString()
	namespace := vcfContext.Namespace.ValueString()
	name := state.Name.ValueString()

	k8sClient, err := kubernetes.NewClient(r.tmClient, project, namespace)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("error creating Kubernetes client for VCF context %s/%s: %s", project, namespace, err.Error()),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(k8sClient.FlushWarnings()...) }()

	var cluster vcfatypes.VksCluster
	if err := k8sClient.ReadNamespaceScopedResource(ctx, namespace, name, vcfatypes.GetVksClusterGVR(), &cluster); err != nil {
		if apierrors.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("could not read %s %s in VCF context %s/%s: %s", vcfatypes.LabelVksCluster, name, project, namespace, err.Error()),
		)
		return
	}

	// Capture the user-managed keys before mapVksClusterToResourceModel overwrites
	// unrelated fields.  After the mapping we filter the live API labels/annotations
	// down to only the keys the user is tracking.
	priorLabels := state.Labels
	priorAnnotations := state.Annotations

	mapVksClusterToResourceModel(ctx, &cluster, &state, &resp.Diagnostics)
	state.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", project, namespace, name))

	// Restore only the user-managed subset of labels/annotations so that
	// backend-injected entries never appear as diffs in the plan.
	state.Labels = filterToUserManagedKeys(ctx, cluster.Labels, priorLabels, &resp.Diagnostics)
	state.Annotations = filterToUserManagedKeys(ctx, cluster.Annotations, priorAnnotations, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vcfaVksClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state vcfaVksClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan vcfaVksClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vcfContext := common.ExtractVcfContext(ctx, plan.Context, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, vksClusterUpdateDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	waitForAvailable, diags := r.extractWaitForAvailable(ctx, plan.WaitFor)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := vcfContext.Project.ValueString()
	namespace := vcfContext.Namespace.ValueString()
	name := plan.Name.ValueString()

	k8sClient, err := kubernetes.NewClient(r.tmClient, project, namespace)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error updating %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("error creating Kubernetes client for VCF context %s/%s: %s", project, namespace, err.Error()),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(k8sClient.FlushWarnings()...) }()

	patchBytes, diags := r.createMergePatch(ctx, state, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only send the patch request if there are changes to apply.
	if len(patchBytes) > 2 {
		var updatedCluster vcfatypes.VksCluster

		var patchMap map[string]any
		if err := json.Unmarshal(patchBytes, &patchMap); err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error updating %s %s", vcfatypes.LabelVksCluster, name),
				fmt.Sprintf("could not unmarshal patch: %s", err.Error()),
			)
			return
		}

		// Retry on conflict: the cluster may have been modified between the read and the
		// patch (e.g. by a controller updating status or metadata). On each conflict we
		// re-read the live resourceVersion and re-send the same content patch.
		var patchErr error
		for attempt := 1; attempt <= vksClusterConflictMaxRetries; attempt++ {
			// Read the live object to obtain the current resourceVersion for optimistic concurrency.
			var currentCluster vcfatypes.VksCluster
			if err := k8sClient.ReadNamespaceScopedResource(ctx, namespace, name, vcfatypes.GetVksClusterGVR(), &currentCluster); err != nil {
				resp.Diagnostics.AddError(
					fmt.Sprintf("error updating %s %s", vcfatypes.LabelVksCluster, name),
					fmt.Sprintf("could not read %s %s in VCF context %s/%s before update: %s", vcfatypes.LabelVksCluster, name, project, namespace, err.Error()),
				)
				return
			}

			// Inject metadata.resourceVersion into the patch for optimistic concurrency.
			meta, _ := patchMap["metadata"].(map[string]any)
			if meta == nil {
				meta = make(map[string]any)
			}
			meta["resourceVersion"] = currentCluster.ResourceVersion
			patchMap["metadata"] = meta

			finalPatch, err := json.Marshal(patchMap)
			if err != nil {
				resp.Diagnostics.AddError(
					fmt.Sprintf("error updating %s %s", vcfatypes.LabelVksCluster, name),
					fmt.Sprintf("could not marshal final patch: %s", err.Error()),
				)
				return
			}

			// Apply the patch
			patchErr = k8sClient.PatchNamespaceScopedResource(ctx, vcfatypes.GetVksClusterGVR(), namespace, name, k8stypes.MergePatchType, finalPatch, &updatedCluster, false)
			if patchErr == nil || !apierrors.IsConflict(patchErr) {
				break
			}

			log.Printf("[DEBUG] conflict patching %s %s in VCF context %s/%s (attempt %d/%d), retrying in %s...",
				vcfatypes.LabelVksCluster, name, project, namespace, attempt, vksClusterConflictMaxRetries, vksClusterConflictRetryInterval)
			select {
			case <-time.After(vksClusterConflictRetryInterval):
			case <-ctx.Done():
				resp.Diagnostics.AddError(
					fmt.Sprintf("error updating %s %s", vcfatypes.LabelVksCluster, name),
					fmt.Sprintf("context cancelled while retrying conflict patch for %s %s in VCF context %s/%s: %s", vcfatypes.LabelVksCluster, name, project, namespace, ctx.Err()),
				)
				return
			}
		}

		if patchErr != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error updating %s %s", vcfatypes.LabelVksCluster, name),
				fmt.Sprintf("could not patch %s %s in VCF context %s/%s: %s", vcfatypes.LabelVksCluster, name, project, namespace, patchErr.Error()),
			)
			return
		}

		// Capture planned values before mapping overwrites plan with the API response.
		// The provider stores what the user supplied so the post-apply state matches
		// the plan, preventing "inconsistent result after apply" errors.
		// Read will always reflect the API's actual values, so real backend-side
		// changes will surface on the next plan.
		planVersion := plan.Version
		planVariables := plan.Variables

		// Planned control-plane variable overrides.
		var planCPOverrides types.Set
		if !plan.ControlPlane.IsNull() && !plan.ControlPlane.IsUnknown() {
			var cp vksClusterControlPlaneTopologyModel
			if d := plan.ControlPlane.As(ctx, &cp, basetypes.ObjectAsOptions{}); !d.HasError() {
				planCPOverrides = cp.VariableOverrides
			}
		}

		// Planned per-MachineDeployment variable overrides, keyed by MD name.
		planMDOverrides := map[string]types.Set{}
		if !plan.MachineDeployments.IsNull() && !plan.MachineDeployments.IsUnknown() {
			var planMDs []vksClusterMachineDeploymentTopologyModel
			if d := plan.MachineDeployments.ElementsAs(ctx, &planMDs, false); !d.HasError() {
				for _, md := range planMDs {
					if !md.Name.IsNull() && !md.Name.IsUnknown() {
						planMDOverrides[md.Name.ValueString()] = md.VariableOverrides
					}
				}
			}
		}

		mapVksClusterToResourceModel(ctx, &updatedCluster, &plan, &resp.Diagnostics)

		plan.Version = planVersion

		// Restore fields whose API response may differ from the planned values
		// mid-reconciliation (e.g. backend-injected variables, resource_version,
		// in-progress status conditions). The next Read will refresh them.
		plan.Metadata = state.Metadata
		plan.Status = state.Status
		plan.Variables = planVariables

		// Restore control-plane variable overrides to planned values.
		if !plan.ControlPlane.IsNull() && !plan.ControlPlane.IsUnknown() {
			var cp vksClusterControlPlaneTopologyModel
			if d := plan.ControlPlane.As(ctx, &cp, basetypes.ObjectAsOptions{}); !d.HasError() {
				cp.VariableOverrides = planCPOverrides
				plan.ControlPlane = helpers.ObjFrom(ctx, vksClusterControlPlaneTopologyAttrTypes, &cp, &resp.Diagnostics)
			}
		}

		// Restore each MachineDeployment's variable overrides to planned values.
		if !plan.MachineDeployments.IsNull() && !plan.MachineDeployments.IsUnknown() {
			var mds []vksClusterMachineDeploymentTopologyModel
			if d := plan.MachineDeployments.ElementsAs(ctx, &mds, false); !d.HasError() {
				for i, md := range mds {
					if !md.Name.IsNull() && !md.Name.IsUnknown() {
						if planned, ok := planMDOverrides[md.Name.ValueString()]; ok {
							mds[i].VariableOverrides = planned
						}
					}
				}
				plan.MachineDeployments = helpers.SetFrom(ctx, types.ObjectType{AttrTypes: vksMachineDeploymentTopologyAttrTypes}, mds, &resp.Diagnostics)
			}
		}

		if waitForAvailable {
			if err := r.waitForClusterAvailable(ctx, k8sClient, project, namespace, name, updateTimeout); err != nil {
				resp.Diagnostics.AddError(
					fmt.Sprintf("%s %s updated but not yet available", vcfatypes.LabelVksCluster, name),
					fmt.Sprintf("%s %s in VCF context %s/%s was updated but did not reach available state within the timeout: %s", vcfatypes.LabelVksCluster, name, project, namespace, err.Error()),
				)
			}
		}
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", project, namespace, name))

	helpers.SanitizeUnknownForState(ctx, reflect.ValueOf(&plan).Elem())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vcfaVksClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vcfaVksClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vcfContext := common.ExtractVcfContext(ctx, state.Context, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, vksClusterDeleteDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	waitDeleted, diags := r.extractWaitForDeleted(ctx, state.WaitFor)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := vcfContext.Project.ValueString()
	namespace := vcfContext.Namespace.ValueString()
	name := state.Name.ValueString()

	k8sClient, err := kubernetes.NewClient(r.tmClient, project, namespace)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error deleting %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("error creating Kubernetes client for VCF context %s/%s: %s", project, namespace, err.Error()),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(k8sClient.FlushWarnings()...) }()

	if err := k8sClient.DeleteNamespaceScopedResource(ctx, namespace, name, vcfatypes.GetVksClusterGVR(), false); err != nil {
		if apierrors.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("error deleting %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("could not delete %s %s in VCF context %s/%s: %s", vcfatypes.LabelVksCluster, name, project, namespace, err.Error()),
		)
		return
	}

	if waitDeleted {
		if err := r.waitForClusterDeleted(ctx, k8sClient, project, namespace, name, deleteTimeout); err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s %s deletion still in progress", vcfatypes.LabelVksCluster, name),
				fmt.Sprintf("%s %s deletion in VCF context %s/%s was initiated but did not complete within the timeout: %s", vcfatypes.LabelVksCluster, name, project, namespace, err.Error()),
			)
		}
	}
}

func (r *vcfaVksClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"invalid import ID format",
			fmt.Sprintf("expected project:namespace:name, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("context").AtName("project"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("context").AtName("namespace"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[2])...)
}

func (r *vcfaVksClusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data vcfaVksClusterResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The `version` attribute accepts both the VKS Kubernetes Release `name` and `version`.
	// If the Kubernetes Release `name` is provided (e.g. `v1.34.1---vmware.1-vkr.4`), the
	// backend converts it to its canonical form (e.g. `v1.34.1+vmware.1`). If we flag the
	// attribute as Required, terraform will fail as it will detect an inconsistent result
	// after apply (the value stored at the state is different from the value in the plan).
	// For this reason, we flag the attribute as Optional and Computed, which will allow the
	// value to be set to the value stored at the state and we will use this custom validator to
	// ensure that the value is set.
	if data.Version.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("version"),
			"Missing required attribute",
			"version must be set. Provide a Kubernetes Release version (e.g. v1.34.1+vmware.1).",
		)
	}
}

func (r *vcfaVksClusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// No API calls are possible when the provider is not yet configured (e.g. terraform validate).
	if r.tmClient == nil {
		return
	}

	// Skip for destroy operations.
	if req.Plan.Raw.IsNull() || req.Config.Raw.IsNull() {
		return
	}

	var plan vcfaVksClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only run the dry-run validation when the user has explicitly opted in.
	if plan.DryRunValidation.IsNull() || plan.DryRunValidation.IsUnknown() || !plan.DryRunValidation.ValueBool() {
		return
	}

	// Skip early if the context or name are not yet known — this happens when they are
	// derived from another resource that has not been applied yet.
	if plan.Context.IsNull() || plan.Context.IsUnknown() {
		return
	}
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		return
	}

	vcfContext := common.ExtractVcfContext(ctx, plan.Context, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	project := vcfContext.Project.ValueString()
	namespace := vcfContext.Namespace.ValueString()
	name := plan.Name.ValueString()

	k8sClient, err := kubernetes.NewClient(r.tmClient, project, namespace)
	if err != nil {
		// Report as a warning so that a transient connectivity issue does not fail the plan.
		resp.Diagnostics.AddWarning(
			fmt.Sprintf("skipping dry-run validation for %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("could not create Kubernetes client for VCF context %s/%s: %s", project, namespace, err.Error()),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(k8sClient.FlushWarnings()...) }()

	var dummy vcfatypes.VksCluster

	// Read the live object to determine create-vs-update and to obtain the
	// resourceVersion needed for the Patch dry-run.
	var currentCluster vcfatypes.VksCluster
	readErr := k8sClient.ReadNamespaceScopedResource(ctx, namespace, name, vcfatypes.GetVksClusterGVR(), &currentCluster)
	if apierrors.IsNotFound(readErr) {
		// ── Create path: validate the full object with a dry-run Create ──────────
		clusterObj := mapResourceModelToVksCluster(ctx, &plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		if err := k8sClient.CreateNamespaceScopedResource(ctx, vcfatypes.GetVksClusterGVR(), namespace, clusterObj, &dummy, true); err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("dry-run validation failed for %s %s", vcfatypes.LabelVksCluster, name),
				fmt.Sprintf("the planned %s configuration was rejected by the backend: %s", vcfatypes.LabelVksCluster, err.Error()),
			)
		}
		return
	}

	if readErr != nil {
		// Transient error reading the live resource; skip validation rather than blocking the plan.
		return
	}

	// ── Update path ──────────────────────────────────────────────────────────
	// Build the patch from the live Kubernetes state to the plan. This is more
	// accurate than using req.State (which may be stale) and correctly reflects
	// what will actually be sent to the API during apply.
	var liveModel vcfaVksClusterResourceModel
	var mappingDiags diag.Diagnostics
	mapVksClusterToResourceModel(ctx, &currentCluster, &liveModel, &mappingDiags)
	if mappingDiags.HasError() {
		// Mapping the live object failed — skip dry-run rather than surfacing an
		// internal error that would block the plan.
		return
	}
	liveModel.Name = plan.Name
	liveModel.Context = plan.Context
	liveModel.Labels = filterToUserManagedKeys(ctx, currentCluster.Labels, plan.Labels, &mappingDiags)
	liveModel.Annotations = filterToUserManagedKeys(ctx, currentCluster.Annotations, plan.Annotations, &mappingDiags)
	if mappingDiags.HasError() {
		return
	}

	patchBytes, diags := r.createMergePatch(ctx, liveModel, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(patchBytes) <= 2 {
		return // Nothing changed; nothing to validate.
	}

	var patchMap map[string]any
	if err := json.Unmarshal(patchBytes, &patchMap); err != nil {
		return
	}

	meta, _ := patchMap["metadata"].(map[string]any)
	if meta == nil {
		meta = make(map[string]any)
	}
	meta["resourceVersion"] = currentCluster.ResourceVersion
	patchMap["metadata"] = meta

	finalPatch, err := json.Marshal(patchMap)
	if err != nil {
		return
	}

	// Retry on conflict: the cluster may have been modified between the read and the
	// patch (e.g. by a controller updating status or metadata). On each conflict we
	// re-read the live resourceVersion and re-send the same content patch.
	// After all retries, report a warning (not an error): a transient conflict
	// during plan is not a validation failure and should not block terraform apply.
	var dryRunErr error
	for attempt := 1; attempt <= vksClusterConflictMaxRetries; attempt++ {
		if attempt > 1 {
			var freshCluster vcfatypes.VksCluster
			if rerr := k8sClient.ReadNamespaceScopedResource(ctx, namespace, name, vcfatypes.GetVksClusterGVR(), &freshCluster); rerr != nil {
				break // Cannot refresh; skip further retries
			}
			meta["resourceVersion"] = freshCluster.ResourceVersion
			patchMap["metadata"] = meta
			finalPatch, err = json.Marshal(patchMap)
			if err != nil {
				break
			}
			select {
			case <-time.After(vksClusterConflictRetryInterval):
			case <-ctx.Done():
				return
			}
		}

		dryRunErr = k8sClient.PatchNamespaceScopedResource(ctx, vcfatypes.GetVksClusterGVR(), namespace, name, k8stypes.MergePatchType, finalPatch, &dummy, true)
		if dryRunErr == nil || !apierrors.IsConflict(dryRunErr) {
			break
		}

		log.Printf("[DEBUG] conflict dry-run patching %s %s in VCF context %s/%s (attempt %d/%d), retrying in %s...",
			vcfatypes.LabelVksCluster, name, project, namespace, attempt, vksClusterConflictMaxRetries, vksClusterConflictRetryInterval)
	}

	if dryRunErr == nil {
		return
	}

	if apierrors.IsConflict(dryRunErr) {
		// Still conflicting after retries — the cluster is under heavy churn.
		// Report as a warning so the plan can proceed; the apply will also
		// retry on conflict.
		resp.Diagnostics.AddWarning(
			fmt.Sprintf("skipping dry-run validation for %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("the %s is being actively modified by the backend; dry-run validation was skipped to avoid blocking the plan: %s", vcfatypes.LabelVksCluster, dryRunErr.Error()),
		)
		return
	}

	resp.Diagnostics.AddError(
		fmt.Sprintf("dry-run validation failed for %s %s", vcfatypes.LabelVksCluster, name),
		fmt.Sprintf("the planned %s updated configuration was rejected by the backend: %s", vcfatypes.LabelVksCluster, dryRunErr.Error()),
	)
}

func (r *vcfaVksClusterResource) createMergePatch(ctx context.Context, state vcfaVksClusterResourceModel, plan vcfaVksClusterResourceModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := plan.Name.ValueString()

	oldObj := mapResourceModelToVksCluster(ctx, &state, &diags)
	if diags.HasError() {
		return nil, diags
	}
	oldJSON, err := json.Marshal(oldObj)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error updating %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("could not marshal old state to JSON: %s", err.Error()),
		)
		return nil, diags
	}

	newObj := mapResourceModelToVksCluster(ctx, &plan, &diags)
	if diags.HasError() {
		return nil, diags
	}
	newJSON, err := json.Marshal(newObj)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error updating %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("could not marshal new plan to JSON: %s", err.Error()),
		)
		return nil, diags
	}

	// Compute the JSON Merge Patch (RFC 7396) from the diff between old and new.
	// Only fields that actually changed appear in the patch; unchanged fields are omitted,
	// which prevents overwriting server-managed spec fields (e.g. controlPlaneRef).
	patchBytes, err := jsonpatch.CreateMergePatch(oldJSON, newJSON)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error updating %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("could not compute merge patch: %s", err.Error()),
		)
		return patchBytes, diags
	}

	// RFC 7396 merge patch represents "remove this field" as {"field": null}.
	// When the user removes their last label the raw patch would contain
	// {"metadata": {"labels": null}}, which would erase ALL labels from the
	// cluster — including backend-injected ones.  Replace the coarse-grained
	// null with a precise per-key diff (removed keys → null, added/changed
	// keys → new value) so that only user-managed keys are affected.
	patchBytes, err = injectPerKeyMapDiffs(ctx, patchBytes, state, plan, &diags)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error updating %s %s", vcfatypes.LabelVksCluster, name),
			fmt.Sprintf("could not inject per-key label/annotation diffs: %s", err.Error()),
		)
		return patchBytes, diags
	}

	return patchBytes, diags
}

func (r *vcfaVksClusterResource) extractWaitForAvailable(ctx context.Context, waitForObj types.Object) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	if waitForObj.IsNull() || waitForObj.IsUnknown() {
		return false, diags
	}
	var wf vksClusterWaitForModel
	diags.Append(waitForObj.As(ctx, &wf, basetypes.ObjectAsOptions{})...)
	if wf.Available.IsNull() || wf.Available.IsUnknown() {
		return false, diags
	}
	return wf.Available.ValueBool(), diags
}

func (r *vcfaVksClusterResource) waitForClusterAvailable(ctx context.Context, k8sClient *kubernetes.Client, projectName string, namespace string, name string, timeout time.Duration) error {
	const (
		vksClusterStateAvailable    = "Available"
		vksClusterStateNotAvailable = "NotAvailable"
	)

	conf := &retry.StateChangeConf{
		Pending:      []string{vksClusterStateNotAvailable},
		Target:       []string{vksClusterStateAvailable},
		Timeout:      timeout,
		PollInterval: vksClusterPollInterval,
		Refresh: func() (any, string, error) {
			var cluster vcfatypes.VksCluster
			if err := k8sClient.ReadNamespaceScopedResource(ctx, namespace, name, vcfatypes.GetVksClusterGVR(), &cluster); err != nil {
				if apierrors.IsNotFound(err) {
					return nil, "", fmt.Errorf("%s %s in VCF context %s/%s not found while waiting to become available", vcfatypes.LabelVksCluster, name, projectName, namespace)
				}
				return nil, "", fmt.Errorf("error polling %s %s in VCF context %s/%s while waiting to become available: %w", vcfatypes.LabelVksCluster, name, projectName, namespace, err)
			}

			if kubernetes.IsConditionTrue(cluster.Status.Conditions, vcfatypes.VksConditionAvailable) {
				return &cluster, vksClusterStateAvailable, nil
			}
			condition := kubernetes.FindCondition(cluster.Status.Conditions, vcfatypes.VksConditionAvailable)
			if condition != nil {
				log.Printf("[DEBUG] waiting for %s %s in VCF context %s/%s to become %s (reason: %s - lastTransitionTime: %s - message: %s)", vcfatypes.LabelVksCluster, name, projectName, namespace, vcfatypes.VksConditionAvailable, condition.Reason, condition.LastTransitionTime, condition.Message)
			} else {
				log.Printf("[DEBUG] waiting for %s %s in VCF context %s/%s to become %s", vcfatypes.LabelVksCluster, name, projectName, namespace, vcfatypes.VksConditionAvailable)
			}
			return &cluster, vksClusterStateNotAvailable, nil
		},
	}

	if _, err := conf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for %s %s in VCF context %s/%s to be available: %w", vcfatypes.LabelVksCluster, name, projectName, namespace, err)
	}
	return nil
}

func (r *vcfaVksClusterResource) extractWaitForDeleted(ctx context.Context, waitForObj types.Object) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	if waitForObj.IsNull() || waitForObj.IsUnknown() {
		return false, diags
	}
	var wf vksClusterWaitForModel
	diags.Append(waitForObj.As(ctx, &wf, basetypes.ObjectAsOptions{})...)
	if wf.Deleted.IsNull() || wf.Deleted.IsUnknown() {
		return false, diags
	}
	return wf.Deleted.ValueBool(), diags
}

func (r *vcfaVksClusterResource) waitForClusterDeleted(ctx context.Context, k8sClient *kubernetes.Client, projectName string, namespace string, name string, deleteTimeout time.Duration) error {
	const (
		vksClusterStateExists  = "Exists"
		vksClusterStateDeleted = "Deleted"
	)

	conf := &retry.StateChangeConf{
		Pending:      []string{vksClusterStateExists},
		Target:       []string{vksClusterStateDeleted},
		Timeout:      deleteTimeout,
		PollInterval: vksClusterPollInterval,
		Refresh: func() (any, string, error) {
			var cluster vcfatypes.VksCluster
			if err := k8sClient.ReadNamespaceScopedResource(ctx, namespace, name, vcfatypes.GetVksClusterGVR(), &cluster); err != nil {
				if apierrors.IsNotFound(err) {
					return "", vksClusterStateDeleted, nil
				}
				return nil, "", fmt.Errorf("error polling %s %s in VCF context %s/%s while waiting to be deleted: %w", vcfatypes.LabelVksCluster, name, projectName, namespace, err)
			}
			log.Printf("[DEBUG] waiting for %s %s in VCF context %s/%s to be deleted (deletionTimestamp: %s - finalizers: %s)", vcfatypes.LabelVksCluster, name, projectName, namespace, cluster.DeletionTimestamp, cluster.Finalizers)
			return &cluster, vksClusterStateExists, nil
		},
	}

	if _, err := conf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for %s %s in VCF context %s/%s to be deleted: %w", vcfatypes.LabelVksCluster, name, projectName, namespace, err)
	}

	return nil
}
