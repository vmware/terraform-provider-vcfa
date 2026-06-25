// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var ConditionsDataSourceSchema = schema.SetNestedAttribute{
	Computed:    true,
	Description: "Current conditions of the resource",
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of condition",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the condition (True, False, Unknown)",
			},
			"observed_generation": schema.Int64Attribute{
				Computed:    true,
				Description: "Generation that was current when this condition was last updated",
			},
			"last_transition_time": schema.StringAttribute{
				Computed:    true,
				Description: "Last time the condition transitioned",
			},
			"reason": schema.StringAttribute{
				Computed:    true,
				Description: "Machine-readable reason for the condition",
			},
			"message": schema.StringAttribute{
				Computed:    true,
				Description: "Human-readable message for the condition",
			},
		},
	},
}

var ConditionsResourceSchema = schema.SetNestedAttribute{
	Computed:    true,
	Description: "Current conditions of the resource",
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of condition",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the condition (True, False, Unknown)",
			},
			"observed_generation": schema.Int64Attribute{
				Computed:    true,
				Description: "Generation that was current when this condition was last updated",
			},
			"last_transition_time": schema.StringAttribute{
				Computed:    true,
				Description: "Last time the condition transitioned",
			},
			"reason": schema.StringAttribute{
				Computed:    true,
				Description: "Machine-readable reason for the condition",
			},
			"message": schema.StringAttribute{
				Computed:    true,
				Description: "Human-readable message for the condition",
			},
		},
	},
}
