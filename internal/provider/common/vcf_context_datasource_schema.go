// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// VcfContextDataSourceSchema is the reusable schema block for the `context`
// attribute on datasources that need a VCF project and namespace to locate a
// Kubernetes resource.
var VcfContextDataSourceSchema = schema.SingleNestedAttribute{
	Required:    true,
	Description: "VCF Automation context required to look up this resource",
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
