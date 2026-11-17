// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

// importedPrivateKey marks a resource that was just imported, so that its first Read takes over
// the members of the Project instead of leaving them unmanaged.
const importedPrivateKey = "imported"

var (
	_ resource.Resource                   = (*vcfaProjectResource)(nil)
	_ resource.ResourceWithConfigure      = (*vcfaProjectResource)(nil)
	_ resource.ResourceWithImportState    = (*vcfaProjectResource)(nil)
	_ resource.ResourceWithValidateConfig = (*vcfaProjectResource)(nil)
)

type vcfaProjectResource struct {
	tmClient *vcfa.VCDClient
}

func NewVcfaProjectResource() resource.Resource {
	return &vcfaProjectResource{}
}

func (r *vcfaProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *vcfaProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig rejects members that would share a role binding: the backend keeps a single
// role per user or group, and `alice` and `alice@` name the same user.
func (r *vcfaProjectResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config vcfaProjectResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	seen := map[string]string{}
	for _, m := range extractMembers(ctx, config.Members, &resp.Diagnostics) {
		if m.Kind.IsUnknown() || m.Name.IsUnknown() || m.Kind.IsNull() || m.Name.IsNull() {
			continue
		}
		bindingName := memberBindingName(m)
		if other, ok := seen[bindingName]; ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("members"),
				"Duplicate project member",
				fmt.Sprintf("%s %q and %q are the same member. Each user or group can be assigned a single role in a %s",
					m.Kind.ValueString(), other, m.Name.ValueString(), vcfatypes.LabelProject),
			)
			continue
		}
		seen[bindingName] = m.Name.ValueString()
	}
}

func (r *vcfaProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vcfaProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	kubernetesClient, err := kubernetes.NewClient(r.tmClient, name, "")
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error creating %s %s", vcfatypes.LabelProject, name),
			fmt.Sprintf("error creating Kubernetes client for the CCI endpoint: %s", err),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(kubernetesClient.FlushWarnings()...) }()

	project := vcfatypes.NewProject(name, plan.Description.ValueString())
	var created vcfatypes.Project
	if err := kubernetesClient.CreateClusterScopedResource(ctx, vcfatypes.GetProjectGVR(), &project, &created, false); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error creating %s %s", vcfatypes.LabelProject, name),
			fmt.Sprintf("could not create %s %s: %s", vcfatypes.LabelProject, name, err.Error()),
		)
		return
	}
	mapProjectToResourceModel(&created, &plan)

	// The Project exists from here on, so it is saved to the state even if assigning the members fails.
	reconcileMembers(ctx, kubernetesClient, &plan, "creating", &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vcfaProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vcfaProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	kubernetesClient, err := kubernetes.NewClient(r.tmClient, name, "")
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s", vcfatypes.LabelProject, name),
			fmt.Sprintf("error creating Kubernetes client for the CCI endpoints: %s", err),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(kubernetesClient.FlushWarnings()...) }()

	var project vcfatypes.Project
	if err := kubernetesClient.ReadClusterScopedResource(ctx, name, vcfatypes.GetProjectGVR(), &project); err != nil {
		if apierrors.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s", vcfatypes.LabelProject, name),
			fmt.Sprintf("could not read %s %s: %s", vcfatypes.LabelProject, name, err.Error()),
		)
		return
	}
	mapProjectToResourceModel(&project, &state)

	imported, diags := req.Private.GetKey(ctx, importedPrivateKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.Members.IsNull() || len(imported) > 0 {
		prior := extractMembers(ctx, state.Members, &resp.Diagnostics)
		bindings, err := listProjectRoleBindings(ctx, kubernetesClient, name)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("error reading %s %s", vcfatypes.LabelProject, name), err.Error())
			return
		}
		state.Members = mapBindingsToMembers(ctx, bindings, prior, &resp.Diagnostics)
	}
	if len(imported) > 0 {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, importedPrivateKey, nil)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vcfaProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state vcfaProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan vcfaProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	kubernetesClient, err := kubernetes.NewClient(r.tmClient, name, "")
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error updating %s %s", vcfatypes.LabelProject, name),
			fmt.Sprintf("error creating Kubernetes client for the CCI endpoints: %s", err),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(kubernetesClient.FlushWarnings()...) }()

	if !plan.Description.Equal(state.Description) {
		// Only the description can be updated. The backend does not check resourceVersion.
		project := vcfatypes.NewProject(name, plan.Description.ValueString())
		var updated vcfatypes.Project
		if err := kubernetesClient.UpdateClusterScopedResource(ctx, vcfatypes.GetProjectGVR(), &project, &updated, false); err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error updating %s %s", vcfatypes.LabelProject, name),
				fmt.Sprintf("could not update %s %s: %s", vcfatypes.LabelProject, name, err.Error()),
			)
			return
		}
		mapProjectToResourceModel(&updated, &plan)
	} else {
		plan.UID = state.UID
	}

	reconcileMembers(ctx, kubernetesClient, &plan, "updating", &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vcfaProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vcfaProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	kubernetesClient, err := kubernetes.NewClient(r.tmClient, name, "")
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error deleting %s %s", vcfatypes.LabelProject, name),
			fmt.Sprintf("error creating Kubernetes client for the CCI endpoint: %s", err),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(kubernetesClient.FlushWarnings()...) }()

	// Deleting a Project deletes its role bindings, and is synchronous.
	if err := kubernetesClient.DeleteClusterScopedResource(ctx, name, vcfatypes.GetProjectGVR(), false); err != nil {
		if apierrors.IsNotFound(err) {
			return
		}
		detail := fmt.Sprintf("could not delete %s %s: %s", vcfatypes.LabelProject, name, err.Error())
		if apierrors.IsConflict(err) {
			detail += fmt.Sprintf("\n\nA %s can only be deleted once it holds no resources", vcfatypes.LabelProject)
		}
		resp.Diagnostics.AddError(fmt.Sprintf("error deleting %s %s", vcfatypes.LabelProject, name), detail)
	}
}

func (r *vcfaProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, importedPrivateKey, []byte(`true`))...)
}

// listProjectRoleBindings returns all the role bindings of a Project. The backend returns every
// item when the list is not limited. It answers with a 400, not a 404, when the Project does not
// exist, so callers read the Project first.
func listProjectRoleBindings(ctx context.Context, k8sClient *kubernetes.Client, projectName string) ([]vcfatypes.ProjectRoleBinding, error) {
	var list vcfatypes.ProjectRoleBindingList
	if err := k8sClient.ListNamespaceScopedResources(ctx, projectName, vcfatypes.GetProjectRoleBindingGVR(), "", &list); err != nil {
		return nil, fmt.Errorf("could not list %ss of %s %s: %w", vcfatypes.LabelProjectRoleBinding, vcfatypes.LabelProject, projectName, err)
	}
	return list.Items, nil
}

// reconcileMembers makes the role bindings of the Project match the planned members, then
// refreshes the members from the role bindings that exist, so that the state reflects what was
// applied even if some of the calls failed. Members are left alone when they are not set.
func reconcileMembers(ctx context.Context, kubernetesClient *kubernetes.Client, model *vcfaProjectResourceModel, operation string, diags *diag.Diagnostics) {
	if model.Members.IsNull() {
		return
	}

	name := model.Name.ValueString()
	planned := extractMembers(ctx, model.Members, diags)
	if diags.HasError() {
		return
	}

	live, err := listProjectRoleBindings(ctx, kubernetesClient, name)
	if err != nil {
		diags.AddError(fmt.Sprintf("error %s %s %s", operation, vcfatypes.LabelProject, name), err.Error())
		return
	}

	// Bindings are removed first, so that the members that remain are never left with more
	// access than planned if a later call fails.
	gvr := vcfatypes.GetProjectRoleBindingGVR()
	toDelete, toUpdate, toCreate := bindingChanges(live, desiredBindings(name, planned))
	var errs []error
	for _, bindingName := range toDelete {
		if err := kubernetesClient.DeleteNamespaceScopedResource(ctx, name, bindingName, gvr, false); err != nil && !apierrors.IsNotFound(err) {
			errs = append(errs, err)
		}
	}
	for _, b := range toUpdate {
		var updated vcfatypes.ProjectRoleBinding
		errs = append(errs, kubernetesClient.UpdateNamespaceScopedResource(ctx, gvr, name, &b, &updated, false))
	}
	for _, b := range toCreate {
		var created vcfatypes.ProjectRoleBinding
		errs = append(errs, kubernetesClient.CreateNamespaceScopedResource(ctx, gvr, name, &b, &created, false))
	}
	if err := errors.Join(errs...); err != nil {
		diags.AddError(fmt.Sprintf("error %s members of %s %s", operation, vcfatypes.LabelProject, name), err.Error())
	}

	refreshed, err := listProjectRoleBindings(ctx, kubernetesClient, name)
	if err != nil {
		diags.AddError(fmt.Sprintf("error %s %s %s", operation, vcfatypes.LabelProject, name), err.Error())
		return
	}

	model.Members = mapBindingsToMembers(ctx, refreshed, planned, diags)
}
