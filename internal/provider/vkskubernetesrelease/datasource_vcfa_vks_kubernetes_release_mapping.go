// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkskubernetesrelease

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func mapVksKubernetesReleaseToModel(ctx context.Context, kr *vcfatypes.KubernetesRelease, model *vcfaVksKubernetesReleaseModel, diags *diag.Diagnostics) {
	metaModel := kubernetes.MapMetadataToModel(ctx, kr.ObjectMeta, diags)

	// Metadata attributes
	model.Metadata = helpers.ObjFrom(ctx, kubernetes.MetadataAttrTypes, metaModel, diags)

	// Spec attributes
	model.Version = types.StringValue(kr.Spec.Version)
	model.Kubernetes = mapVksKubernetesSpecToModel(ctx, kr.Spec.Kubernetes, diags)
	model.OsImages = mapLocalObjectReferencesToModel(ctx, kr.Spec.OSImages, diags)
	model.BootstrapPackages = mapLocalObjectReferencesToModel(ctx, kr.Spec.BootstrapPackages, diags)

	// Status attributes
	model.Status = helpers.ObjFrom(ctx, vksKubernetesReleaseStatusAttrTypes, &vksKubernetesReleaseStatusModel{
		Conditions: mapCAPIV1Beta1ConditionsToModel(ctx, kr.Status.Conditions, diags),
	}, diags)
}

func mapVksKubernetesSpecToModel(ctx context.Context, ks vcfatypes.KubernetesSpec, diags *diag.Diagnostics) types.Object {
	return helpers.ObjFrom(ctx, vksKubernetesSpecAttrTypes, &vksKubernetesSpecModel{
		Version:         types.StringValue(ks.Version),
		ImageRepository: types.StringValue(ks.ImageRepository),
		Etcd:            mapContainerImageInfoToModel(ctx, ks.Etcd, diags),
		Pause:           mapContainerImageInfoToModel(ctx, ks.Pause, diags),
		CoreDNS:         mapContainerImageInfoToModel(ctx, ks.CoreDNS, diags),
		KubeVIP:         mapContainerImageInfoToModel(ctx, ks.KubeVIP, diags),
	}, diags)
}

func mapContainerImageInfoToModel(ctx context.Context, info *vcfatypes.ContainerImageInfo, diags *diag.Diagnostics) types.Object {
	if info == nil {
		return types.ObjectNull(vksContainerImageInfoAttrTypes)
	}
	return helpers.ObjFrom(ctx, vksContainerImageInfoAttrTypes, &vksContainerImageInfoModel{
		ImageRepository: types.StringValue(info.ImageRepository),
		ImageTag:        types.StringValue(info.ImageTag),
	}, diags)
}

func mapLocalObjectReferencesToModel(ctx context.Context, refs []corev1.LocalObjectReference, diags *diag.Diagnostics) types.Set {
	if len(refs) == 0 {
		return types.SetValueMust(types.StringType, nil)
	}
	names := make([]string, 0, len(refs))
	for _, r := range refs {
		names = append(names, r.Name)
	}
	return helpers.SetFrom(ctx, types.StringType, names, diags)
}

// mapCAPIV1Beta1ConditionsToModel maps the deprecated CAPI v1beta1-style Condition slice
// (clusterv1.Condition) into the standard kubernetes.ConditionModel slice used for
// Terraform state. ObservedGeneration is not present in v1beta1 conditions and is always 0.
func mapCAPIV1Beta1ConditionsToModel(ctx context.Context, conditions []clusterv1.Condition, diags *diag.Diagnostics) types.Set { //nolint:staticcheck
	models := make([]kubernetes.ConditionModel, 0, len(conditions))
	for _, c := range conditions {
		models = append(models, kubernetes.ConditionModel{
			Type:               types.StringValue(string(c.Type)),
			Status:             types.StringValue(string(c.Status)),
			ObservedGeneration: types.Int64Value(0),
			LastTransitionTime: types.StringValue(c.LastTransitionTime.UTC().Format("2006-01-02T15:04:05Z")),
			Reason:             types.StringValue(c.Reason),
			Message:            types.StringValue(c.Message),
		})
	}
	return helpers.SetFrom(ctx, types.ObjectType{AttrTypes: kubernetes.ConditionAttrTypes}, models, diags)
}
