// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MapMetadataToModel(ctx context.Context, meta metav1.ObjectMeta, diags *diag.Diagnostics) *MetadataModel {
	m := &MetadataModel{
		Name:                       types.StringValue(meta.Name),
		GenerateName:               types.StringValue(meta.GenerateName),
		Namespace:                  types.StringValue(meta.Namespace),
		UID:                        types.StringValue(string(meta.UID)),
		ResourceVersion:            types.StringValue(meta.ResourceVersion),
		Generation:                 types.Int64Value(meta.Generation),
		CreationTimestamp:          types.StringValue(meta.CreationTimestamp.UTC().Format(time.RFC3339)),
		DeletionGracePeriodSeconds: types.Int64PointerValue(meta.DeletionGracePeriodSeconds),
		OwnerReferences:            []MetadataOwnerReferenceModel{},
	}

	if meta.DeletionTimestamp != nil {
		m.DeletionTimestamp = types.StringValue(meta.DeletionTimestamp.UTC().Format(time.RFC3339))
	} else {
		m.DeletionTimestamp = types.StringNull()
	}

	labels := meta.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	labelMap, d := types.MapValueFrom(ctx, types.StringType, labels)
	diags.Append(d...)
	m.Labels = labelMap

	annotations := meta.Annotations
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotationMap, d := types.MapValueFrom(ctx, types.StringType, annotations)
	diags.Append(d...)
	m.Annotations = annotationMap

	finalizers := meta.Finalizers
	if finalizers == nil {
		finalizers = []string{}
	}
	finalizerSet, d := types.SetValueFrom(ctx, types.StringType, finalizers)
	diags.Append(d...)
	m.Finalizers = finalizerSet

	for _, ref := range meta.OwnerReferences {
		m.OwnerReferences = append(m.OwnerReferences, MetadataOwnerReferenceModel{
			APIVersion:         types.StringValue(ref.APIVersion),
			Kind:               types.StringValue(ref.Kind),
			Name:               types.StringValue(ref.Name),
			UID:                types.StringValue(string(ref.UID)),
			Controller:         types.BoolPointerValue(ref.Controller),
			BlockOwnerDeletion: types.BoolPointerValue(ref.BlockOwnerDeletion),
		})
	}

	return m
}
