// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vksclusterkubeconfig

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
	_ datasource.DataSource              = (*vcfaVksClusterKubeconfigDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*vcfaVksClusterKubeconfigDataSource)(nil)
)

type vcfaVksClusterKubeconfigDataSource struct {
	tmClient *vcfa.VCDClient
}

func NewVcfaVksClusterKubeconfigDataSource() datasource.DataSource {
	return &vcfaVksClusterKubeconfigDataSource{}
}

func (d *vcfaVksClusterKubeconfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vks_cluster_kubeconfig"
}

func (d *vcfaVksClusterKubeconfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vcfaVksClusterKubeconfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vcfaVksClusterKubeconfigModel
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
	clusterName := data.Name.ValueString()

	kubernetesClient, err := kubernetes.NewClient(d.tmClient, project, namespace)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s for %s %s", vcfatypes.LabelVksClusterKubeconfig, vcfatypes.LabelVksCluster, clusterName),
			fmt.Sprintf("error creating Kubernetes client for VCF context %s/%s: %s", project, namespace, err),
		)
		return
	}
	defer func() { resp.Diagnostics.Append(kubernetesClient.FlushWarnings()...) }()

	var cluster vcfatypes.VksCluster
	if err := kubernetesClient.ReadNamespaceScopedResource(ctx, namespace, clusterName, vcfatypes.GetVksClusterGVR(), &cluster); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s for %s %s", vcfatypes.LabelVksClusterKubeconfig, vcfatypes.LabelVksCluster, clusterName),
			fmt.Sprintf("could not read %s %s in VCF context %s/%s: %s", vcfatypes.LabelVksCluster, clusterName, project, namespace, err),
		)
		return
	}

	secretName := fmt.Sprintf("%s%s", clusterName, vcfatypes.VksClusterKubeconfigSecretSuffix)
	secret, err := kubernetesClient.ReadSecret(ctx, namespace, secretName)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s for %s %s", vcfatypes.LabelVksClusterKubeconfig, vcfatypes.LabelVksCluster, clusterName),
			fmt.Sprintf("could not read kubeconfig secret %s in VCF context %s/%s: %s", secretName, project, namespace, err),
		)
		return
	}

	if string(secret.Type) != vcfatypes.VksClusterKubeconfigSecretType {
		resp.Diagnostics.AddError(
			fmt.Sprintf("error reading %s for %s %s", vcfatypes.LabelVksClusterKubeconfig, vcfatypes.LabelVksCluster, clusterName),
			fmt.Sprintf("secret %s has unexpected type %s (expected %q)", secretName, secret.Type, vcfatypes.VksClusterKubeconfigSecretType),
		)
		return
	}

	data.ID = types.StringValue(secretName)
	mapVksClusterKubeconfigToModel(ctx, clusterName, secret, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
