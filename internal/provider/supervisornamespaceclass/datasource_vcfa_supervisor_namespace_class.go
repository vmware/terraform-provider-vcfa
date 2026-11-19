// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisornamespaceclass

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

var (
	_ datasource.DataSource              = (*vcfaSupervisorNamespaceClassDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*vcfaSupervisorNamespaceClassDataSource)(nil)
)

type vcfaSupervisorNamespaceClassDataSource struct {
	tmClient *vcfa.VCDClient
}

func NewVcfaSupervisorNamespaceClassDataSource() datasource.DataSource {
	return &vcfaSupervisorNamespaceClassDataSource{}
}

func (d *vcfaSupervisorNamespaceClassDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_supervisor_namespace_class"
}

func (d *vcfaSupervisorNamespaceClassDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vcfaSupervisorNamespaceClassDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vcfaSupervisorNamespaceClassDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	kubernetesClient, err := kubernetes.NewClient(d.tmClient, "", "")
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
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s %s not found", vcfatypes.LabelSupervisorNamespaceClass, name),
				fmt.Sprintf("could not find %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error()),
			)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
			fmt.Sprintf("could not read %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error()),
		)
		return
	}

	var config vcfatypes.SupervisorNamespaceClassConfig
	if err := kubernetesClient.ReadClusterScopedResource(ctx, name, vcfatypes.GetSupervisorNamespaceClassConfigGVR(), &config); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s", vcfatypes.LabelSupervisorNamespaceClass, name),
			fmt.Sprintf("could not read the configuration of %s %s: %s", vcfatypes.LabelSupervisorNamespaceClass, name, err.Error()),
		)
		return
	}

	mapSupervisorNamespaceClassToDataSourceModel(ctx, &class, &config, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
