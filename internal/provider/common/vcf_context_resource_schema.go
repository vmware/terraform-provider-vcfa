// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// VcfContextResourceSchema is the reusable schema block for the `context`
// attribute on resources that need a VCF project and namespace to manage a
// Kubernetes resource.
var VcfContextResourceSchema = schema.SingleNestedAttribute{
	Required:    true,
	Description: "VCF Automation context required to manage this resource",
	PlanModifiers: []planmodifier.Object{
		objectplanmodifier.RequiresReplace(),
	},
	Attributes: map[string]schema.Attribute{
		"project": schema.StringAttribute{
			Required:    true,
			Description: "Name of the Project where the resource is located",
		},
		"namespace": schema.StringAttribute{
			Required:    true,
			Description: "Name of the Namespace where the resource is located",
		},
	},
}
