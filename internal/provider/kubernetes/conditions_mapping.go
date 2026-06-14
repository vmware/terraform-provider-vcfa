// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MapConditionsToModel(ctx context.Context, conditions []metav1.Condition, diags *diag.Diagnostics) ConditionsModel {
	result := make([]ConditionModel, 0, len(conditions))
	for _, c := range conditions {
		result = append(result, ConditionModel{
			Type:               types.StringValue(c.Type),
			Status:             types.StringValue(string(c.Status)),
			ObservedGeneration: types.Int64Value(c.ObservedGeneration),
			LastTransitionTime: types.StringValue(c.LastTransitionTime.Format("2006-01-02T15:04:05Z")),
			Reason:             types.StringValue(c.Reason),
			Message:            types.StringValue(c.Message),
		})
	}
	return result
}
