// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vksclusterkubeconfig

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/common"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func (d *vcfaVksClusterKubeconfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Data source for reading a %s", vcfatypes.LabelVksClusterKubeconfig),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Internal identifier of the %s", vcfatypes.LabelVksClusterKubeconfig),
			},

			// Required lookup attributes
			"context": common.VcfContextDataSourceSchema,
			"name": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", vcfatypes.LabelVksCluster),
			},

			// Computed attributes
			"host": schema.StringAttribute{
				Computed:    true,
				Description: "Kubernetes API server URL extracted from the kubeconfig",
			},
			"insecure_skip_tls_verify": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether TLS verification is disabled for the Kubernetes API server",
			},
			"kube_config_raw": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: fmt.Sprintf("Raw kubeconfig YAML content of the %s", vcfatypes.LabelVksCluster),
			},
			"context_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the current context in the kubeconfig",
			},
			"user": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the user entry in the kubeconfig",
			},
			"token": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Bearer token for authenticating to the Kubernetes API server (empty for certificate-based auth)",
			},
			"certificate_authority_data": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Base64-encoded PEM certificate authority data for the cluster",
			},
			"client_certificate_data": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Base64-encoded PEM client certificate data for authenticating to the cluster",
			},
			"client_key_data": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Base64-encoded PEM client key data for authenticating to the cluster",
			},
		},
	}
}
