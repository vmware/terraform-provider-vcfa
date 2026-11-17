// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package project

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── Resource Top-level model ─────────────────────────────────────────────────

type vcfaProjectResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	UID         types.String `tfsdk:"uid"`

	// Members is authoritative when set: role bindings not listed are removed. When null, the
	// role bindings of the Project are not managed.
	Members types.Set `tfsdk:"members"`
}

// ── Members ──────────────────────────────────────────────────────────────────

type projectMemberModel struct {
	Kind types.String `tfsdk:"kind"`
	Name types.String `tfsdk:"name"`
	Role types.String `tfsdk:"role"`
}

var projectMemberAttrTypes = map[string]attr.Type{
	"kind": types.StringType,
	"name": types.StringType,
	"role": types.StringType,
}

var projectMemberObjectType = types.ObjectType{AttrTypes: projectMemberAttrTypes}
