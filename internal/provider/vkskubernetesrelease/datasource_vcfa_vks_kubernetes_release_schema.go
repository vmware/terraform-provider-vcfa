// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkskubernetesrelease

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/common"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func (d *vcfaVksKubernetesReleaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	containerImageInfoAttrs := map[string]schema.Attribute{
		"image_repository": schema.StringAttribute{
			Computed:    true,
			Description: "Container image registry to pull images from",
		},
		"image_tag": schema.StringAttribute{
			Computed:    true,
			Description: "Container image tag",
		},
	}

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Data source for reading a %s", vcfatypes.LabelVksKubernetesRelease),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Internal identifier of the %s", vcfatypes.LabelVksKubernetesRelease),
			},

			// Required lookup attributes
			"context": common.VcfContextDataSourceSchema,
			"name": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", vcfatypes.LabelVksKubernetesRelease),
			},

			// Metadata attributes
			"metadata": kubernetes.MetadataDataSourceSchema,

			// Spec attributes
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified Semantic Versioning conformant version of the KubernetesRelease",
			},
			"kubernetes": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Details about the Kubernetes distribution shipped by this release",
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Computed:    true,
						Description: "Semantic versioning conformant version of the Kubernetes build",
					},
					"image_repository": schema.StringAttribute{
						Computed:    true,
						Description: "Container image registry to pull Kubernetes component images from",
					},
					"etcd": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Container image repository and tag for etcd",
						Attributes:  containerImageInfoAttrs,
					},
					"pause": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Container image repository and tag for pause",
						Attributes:  containerImageInfoAttrs,
					},
					"coredns": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Container image repository and tag for CoreDNS",
						Attributes:  containerImageInfoAttrs,
					},
					"kube_vip": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Container image repository and tag for kube-vip",
						Attributes:  containerImageInfoAttrs,
					},
				},
			},
			"os_images": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Names of OSImage objects shipped with this release",
			},
			"bootstrap_packages": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Names of bootstrap packages shipped with this release",
			},

			// Status attributes
			"status": schema.SingleNestedAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Observed state of the %s", vcfatypes.LabelVksKubernetesRelease),
				Attributes: map[string]schema.Attribute{
					"conditions": kubernetes.ConditionsDataSourceSchema,
				},
			},
		},
	}
}
