// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisornamespaceclass

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

var (
	_ resource.Resource                = (*vcfaSupervisorNamespaceClassResource)(nil)
	_ resource.ResourceWithConfigure   = (*vcfaSupervisorNamespaceClassResource)(nil)
	_ resource.ResourceWithImportState = (*vcfaSupervisorNamespaceClassResource)(nil)
)

type vcfaSupervisorNamespaceClassResource struct {
	tmClient *vcfa.VCDClient
}

func NewVcfaSupervisorNamespaceClassResource() resource.Resource {
	return &vcfaSupervisorNamespaceClassResource{}
}

func (r *vcfaSupervisorNamespaceClassResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_supervisor_namespace_class"
}

func (r *vcfaSupervisorNamespaceClassResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *vcfaSupervisorNamespaceClassResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vcfaSupervisorNamespaceClassResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	kubernetesClient, err := kubernetes.NewClient(r.tmClient, "", "")
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error creating %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
			fmt.Sprintf("error creating Kubernetes client for the CCI endpoint: %s", err),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(kubernetesClient.FlushWarnings()...) }()

	config := buildSupervisorNamespaceClassConfig(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	class := buildSupervisorNamespaceClass(&plan)
	var created vcfatypes.SupervisorNamespaceClass
	if err := kubernetesClient.CreateClusterScopedResource(ctx, vcfatypes.GetSupervisorNamespaceClassGVR(), &class, &created, false); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error creating %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
			fmt.Sprintf("could not create %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error()),
		)
		return
	}

	var createdConfig vcfatypes.SupervisorNamespaceClassConfig
	if err := kubernetesClient.CreateClusterScopedResource(ctx, vcfatypes.GetSupervisorNamespaceClassConfigGVR(), &config, &createdConfig, false); err != nil {
		// A class without its config is unusable, so it is rolled back rather than left half-created.
		detail := fmt.Sprintf("could not create the configuration of %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error())
		if delErr := kubernetesClient.DeleteClusterScopedResource(ctx, name, vcfatypes.GetSupervisorNamespaceClassGVR(), false); delErr != nil && !apierrors.IsNotFound(delErr) {
			detail += fmt.Sprintf("\n\nThe %s could not be removed, delete it manually: %s", vcfatypes.LabelSupervisorNamespaceClass, delErr.Error())
		}
		resp.Diagnostics.AddError(fmt.Sprintf("error creating %s %s", vcfatypes.LabelSupervisorNamespaceClass, name), detail)
		return
	}

	mapSupervisorNamespaceClassToResourceModel(&created, &plan)
	mapSupervisorNamespaceClassConfigToResourceModel(ctx, &createdConfig, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vcfaSupervisorNamespaceClassResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vcfaSupervisorNamespaceClassResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	kubernetesClient, err := kubernetes.NewClient(r.tmClient, "", "")
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
			fmt.Sprintf("error creating Kubernetes client for the CCI endpoint: %s", err),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(kubernetesClient.FlushWarnings()...) }()

	var class vcfatypes.SupervisorNamespaceClass
	if err := kubernetesClient.ReadClusterScopedResource(ctx, name, vcfatypes.GetSupervisorNamespaceClassGVR(), &class); err != nil {
		if apierrors.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
			fmt.Sprintf("could not read %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error()),
		)
		return
	}
	mapSupervisorNamespaceClassToResourceModel(&class, &state)

	var config vcfatypes.SupervisorNamespaceClassConfig
	if err := kubernetesClient.ReadClusterScopedResource(ctx, name, vcfatypes.GetSupervisorNamespaceClassConfigGVR(), &config); err != nil {
		if !apierrors.IsNotFound(err) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error reading %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
				fmt.Sprintf("could not read the configuration of %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error()),
			)
			return
		}
		// The configuration was removed outside of Terraform: an empty one makes the next plan
		// show the difference, and the next apply recreates it.
		config = vcfatypes.SupervisorNamespaceClassConfig{}
	}
	mapSupervisorNamespaceClassConfigToResourceModel(ctx, &config, &state, &resp.Diagnostics)
	if len(config.Spec.StorageClasses) == 0 {
		state.StorageClasses = types.SetNull(storageClassObjectType)
		state.VMClasses = types.SetNull(types.StringType)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vcfaSupervisorNamespaceClassResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan vcfaSupervisorNamespaceClassResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	kubernetesClient, err := kubernetes.NewClient(r.tmClient, "", "")
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error updating %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
			fmt.Sprintf("error creating Kubernetes client for the CCI endpoint: %s", err),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(kubernetesClient.FlushWarnings()...) }()

	config := buildSupervisorNamespaceClassConfig(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) {
		var live, updated vcfatypes.SupervisorNamespaceClass
		if err := kubernetesClient.ReadClusterScopedResource(ctx, name, vcfatypes.GetSupervisorNamespaceClassGVR(), &live); err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error updating %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
				fmt.Sprintf("could not read %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error()),
			)
			return
		}
		class := buildSupervisorNamespaceClass(&plan)
		class.ResourceVersion = live.ResourceVersion
		if err := kubernetesClient.UpdateClusterScopedResource(ctx, vcfatypes.GetSupervisorNamespaceClassGVR(), &class, &updated, false); err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error updating %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
				fmt.Sprintf("could not update %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error()),
			)
			return
		}
		mapSupervisorNamespaceClassToResourceModel(&updated, &plan)
	}

	// The configuration is created again if it was removed outside of Terraform.
	var liveConfig, applied vcfatypes.SupervisorNamespaceClassConfig
	configGVR := vcfatypes.GetSupervisorNamespaceClassConfigGVR()
	err = kubernetesClient.ReadClusterScopedResource(ctx, name, configGVR, &liveConfig)
	switch {
	case apierrors.IsNotFound(err):
		err = kubernetesClient.CreateClusterScopedResource(ctx, configGVR, &config, &applied, false)
	case err == nil:
		config.ResourceVersion = liveConfig.ResourceVersion
		err = kubernetesClient.UpdateClusterScopedResource(ctx, configGVR, &config, &applied, false)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error updating %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
			fmt.Sprintf("could not update the configuration of %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error()),
		)
		return
	}
	mapSupervisorNamespaceClassConfigToResourceModel(ctx, &applied, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vcfaSupervisorNamespaceClassResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vcfaSupervisorNamespaceClassResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	kubernetesClient, err := kubernetes.NewClient(r.tmClient, "", "")
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error deleting %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
			fmt.Sprintf("error creating Kubernetes client for the CCI endpoint: %s", err),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(kubernetesClient.FlushWarnings()...) }()

	// The class goes first: if its deletion is rejected, nothing was removed. If only the
	// configuration deletion fails, the resource stays in the state and the next delete retries it.
	if err := kubernetesClient.DeleteClusterScopedResource(ctx, name, vcfatypes.GetSupervisorNamespaceClassGVR(), false); err != nil && !apierrors.IsNotFound(err) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error deleting %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
			fmt.Sprintf("could not delete %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error()),
		)
		return
	}
	if err := kubernetesClient.DeleteClusterScopedResource(ctx, name, vcfatypes.GetSupervisorNamespaceClassConfigGVR(), false); err != nil && !apierrors.IsNotFound(err) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error deleting %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
			fmt.Sprintf("could not delete the configuration of %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error()),
		)
	}
}

func (r *vcfaSupervisorNamespaceClassResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}
