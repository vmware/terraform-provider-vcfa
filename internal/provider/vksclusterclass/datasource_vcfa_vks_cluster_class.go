// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vksclusterclass

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/common"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

var (
	_ datasource.DataSource              = (*vcfaVksClusterClassDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*vcfaVksClusterClassDataSource)(nil)
)

type vcfaVksClusterClassDataSource struct {
	tmClient *vcfa.VCDClient
}

func NewVcfaVksClusterClassDataSource() datasource.DataSource {
	return &vcfaVksClusterClassDataSource{}
}

func (d *vcfaVksClusterClassDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vks_cluster_class"
}

func (d *vcfaVksClusterClassDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vcfaVksClusterClassDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vksClusterClassModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vcfContext := common.ExtractVcfContext(ctx, data.Context, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	project := vcfContext.Project.ValueString()
	namespace := vcfContext.Namespace.ValueString()
	name := data.Name.ValueString()
	system := data.System.ValueBool()

	kubernetesClient, err := kubernetes.NewClient(d.tmClient, project, namespace)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s", vcfatypes.LabelVksClusterClass, name),
			fmt.Sprintf("error creating Kubernetes client for VCF context %s/%s: %s", project, namespace, err.Error()),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(kubernetesClient.FlushWarnings()...) }()

	clusterClassNamespace := namespace
	if system {
		clusterClassNamespace = vcfatypes.VksClusterClassSystemNamespace
	}
	var clusterClass vcfatypes.VksClusterClass
	if err := kubernetesClient.ReadNamespaceScopedResource(ctx, clusterClassNamespace, name, vcfatypes.GetVksClusterClassGVR(), &clusterClass); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s %s/%s", vcfatypes.LabelVksClusterClass, clusterClassNamespace, name),
			fmt.Sprintf("Could not read %s %s/%s in VCF context %s/%s: %s", vcfatypes.LabelVksClusterClass, clusterClassNamespace, name, project, namespace, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s:%s:%s", project, namespace, clusterClassNamespace, name))
	mapVksClusterClassToModel(ctx, &clusterClass, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
