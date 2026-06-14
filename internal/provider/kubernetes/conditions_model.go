// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConditionsModel []ConditionModel

type ConditionModel struct {
	Type               types.String `tfsdk:"type"`
	Status             types.String `tfsdk:"status"`
	ObservedGeneration types.Int64  `tfsdk:"observed_generation"`
	LastTransitionTime types.String `tfsdk:"last_transition_time"`
	Reason             types.String `tfsdk:"reason"`
	Message            types.String `tfsdk:"message"`
}

var ConditionAttrTypes = map[string]attr.Type{
	"type":                 types.StringType,
	"status":               types.StringType,
	"observed_generation":  types.Int64Type,
	"last_transition_time": types.StringType,
	"reason":               types.StringType,
	"message":              types.StringType,
}
