// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkskubernetesrelease

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
)

// ── Top-level model ──────────────────────────────────────────────────────────

type vcfaVksKubernetesReleaseModel struct {
	ID      types.String `tfsdk:"id"`
	Context types.Object `tfsdk:"context"`
	Name    types.String `tfsdk:"name"`

	// Metadata attributes
	Metadata types.Object `tfsdk:"metadata"`

	// Spec attributes
	Version           types.String `tfsdk:"version"`
	Kubernetes        types.Object `tfsdk:"kubernetes"`
	OsImages          types.Set    `tfsdk:"os_images"`
	BootstrapPackages types.Set    `tfsdk:"bootstrap_packages"`

	// Status attributes
	Status types.Object `tfsdk:"status"`
}

// ── Status ───────────────────────────────────────────────────────────────────

type vksKubernetesReleaseStatusModel struct {
	Conditions types.Set `tfsdk:"conditions"`
}

var vksKubernetesReleaseStatusAttrTypes = map[string]attr.Type{
	"conditions": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: kubernetes.ConditionAttrTypes,
		},
	},
}

// ── Kubernetes ───────────────────────────────────────────────────────────────

type vksKubernetesSpecModel struct {
	Version         types.String `tfsdk:"version"`
	ImageRepository types.String `tfsdk:"image_repository"`
	Etcd            types.Object `tfsdk:"etcd"`
	Pause           types.Object `tfsdk:"pause"`
	CoreDNS         types.Object `tfsdk:"coredns"`
	KubeVIP         types.Object `tfsdk:"kube_vip"`
}

var vksKubernetesSpecAttrTypes = map[string]attr.Type{
	"version":          types.StringType,
	"image_repository": types.StringType,
	"etcd": types.ObjectType{
		AttrTypes: vksContainerImageInfoAttrTypes,
	},
	"pause": types.ObjectType{
		AttrTypes: vksContainerImageInfoAttrTypes,
	},
	"coredns": types.ObjectType{
		AttrTypes: vksContainerImageInfoAttrTypes,
	},
	"kube_vip": types.ObjectType{
		AttrTypes: vksContainerImageInfoAttrTypes,
	},
}

// ── Container Image Info ─────────────────────────────────────────────────────

type vksContainerImageInfoModel struct {
	ImageRepository types.String `tfsdk:"image_repository"`
	ImageTag        types.String `tfsdk:"image_tag"`
}

var vksContainerImageInfoAttrTypes = map[string]attr.Type{
	"image_repository": types.StringType,
	"image_tag":        types.StringType,
}
