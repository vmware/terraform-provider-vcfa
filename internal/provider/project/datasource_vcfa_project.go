// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

var (
	_ datasource.DataSource              = (*vcfaProjectDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*vcfaProjectDataSource)(nil)
)

type vcfaProjectDataSource struct {
	tmClient *vcfa.VCDClient
}

func NewVcfaProjectDataSource() datasource.DataSource {
	return &vcfaProjectDataSource{}
}

func (d *vcfaProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *vcfaProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	tmClient, err := helpers.GetTmClientFromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("error getting TM client", err.Error())
		return
	}
	d.tmClient = tmClient
}

func (d *vcfaProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vcfaProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	kubernetesClient, err := kubernetes.NewClient(d.tmClient, name, "")
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s", vcfatypes.LabelProject, name),
			fmt.Sprintf("error creating Kubernetes client for the CCI endpoint: %s", err),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(kubernetesClient.FlushWarnings()...) }()

	var project vcfatypes.Project
	if err := kubernetesClient.ReadClusterScopedResource(ctx, name, vcfatypes.GetProjectGVR(), &project); err != nil {
		if apierrors.IsNotFound(err) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s %s not found", vcfatypes.LabelProject, name),
				fmt.Sprintf("could not find %s %s: %s", vcfatypes.LabelProject, name, err.Error()),
			)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s", vcfatypes.LabelProject, name),
			fmt.Sprintf("could not read %s %s: %s", vcfatypes.LabelProject, name, err.Error()),
		)
		return
	}

	bindings, err := listProjectRoleBindings(ctx, kubernetesClient, name)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error reading %s %s", vcfatypes.LabelProject, name), err.Error())
		return
	}

	data.ID = types.StringValue(name)
	mapProjectToDataSourceModel(ctx, project, bindings, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
