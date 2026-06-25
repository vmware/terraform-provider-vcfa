// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MetadataModel struct {
	Name                       types.String                  `tfsdk:"name"`
	GenerateName               types.String                  `tfsdk:"generate_name"`
	Namespace                  types.String                  `tfsdk:"namespace"`
	UID                        types.String                  `tfsdk:"uid"`
	ResourceVersion            types.String                  `tfsdk:"resource_version"`
	Generation                 types.Int64                   `tfsdk:"generation"`
	CreationTimestamp          types.String                  `tfsdk:"creation_timestamp"`
	DeletionTimestamp          types.String                  `tfsdk:"deletion_timestamp"`
	DeletionGracePeriodSeconds types.Int64                   `tfsdk:"deletion_grace_period_seconds"`
	Labels                     types.Map                     `tfsdk:"labels"`
	Annotations                types.Map                     `tfsdk:"annotations"`
	OwnerReferences            []MetadataOwnerReferenceModel `tfsdk:"owner_references"`
	Finalizers                 types.Set                     `tfsdk:"finalizers"`
}

var MetadataAttrTypes = map[string]attr.Type{
	"name":                          types.StringType,
	"generate_name":                 types.StringType,
	"namespace":                     types.StringType,
	"uid":                           types.StringType,
	"resource_version":              types.StringType,
	"generation":                    types.Int64Type,
	"creation_timestamp":            types.StringType,
	"deletion_timestamp":            types.StringType,
	"deletion_grace_period_seconds": types.Int64Type,
	"labels": types.MapType{
		ElemType: types.StringType,
	},
	"annotations": types.MapType{
		ElemType: types.StringType,
	},
	"owner_references": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: MetadataOwnerReferenceAttrTypes,
		},
	},
	"finalizers": types.SetType{
		ElemType: types.StringType,
	},
}

type MetadataOwnerReferenceModel struct {
	APIVersion         types.String `tfsdk:"api_version"`
	Kind               types.String `tfsdk:"kind"`
	Name               types.String `tfsdk:"name"`
	UID                types.String `tfsdk:"uid"`
	Controller         types.Bool   `tfsdk:"controller"`
	BlockOwnerDeletion types.Bool   `tfsdk:"block_owner_deletion"`
}

var MetadataOwnerReferenceAttrTypes = map[string]attr.Type{
	"api_version":          types.StringType,
	"kind":                 types.StringType,
	"name":                 types.StringType,
	"uid":                  types.StringType,
	"controller":           types.BoolType,
	"block_owner_deletion": types.BoolType,
}
