// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisornamespaceclass

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── Resource Top-level model ─────────────────────────────────────────────────

type vcfaSupervisorNamespaceClassResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	StorageClasses types.Set    `tfsdk:"storage_classes"`
	VMClasses      types.Set    `tfsdk:"vm_classes"`
	ContentSources types.Set    `tfsdk:"content_sources"`
	DefaultLimits  types.Object `tfsdk:"default_limits"`
	Zones          types.Set    `tfsdk:"zones"`
}

// ── Nested models, shared with the data source ───────────────────────────────

type storageClassModel struct {
	Name  types.String `tfsdk:"name"`
	Limit types.String `tfsdk:"limit"`
}

type contentSourceModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

type limitsModel struct {
	CPULimit          types.String `tfsdk:"cpu_limit"`
	MemoryLimit       types.String `tfsdk:"memory_limit"`
	CPUReservation    types.String `tfsdk:"cpu_reservation"`
	MemoryReservation types.String `tfsdk:"memory_reservation"`
}

type zoneModel struct {
	Name                types.String `tfsdk:"name"`
	CPULimit            types.String `tfsdk:"cpu_limit"`
	CPUReservation      types.String `tfsdk:"cpu_reservation"`
	MemoryLimit         types.String `tfsdk:"memory_limit"`
	MemoryReservation   types.String `tfsdk:"memory_reservation"`
	VMClassReservations types.Set    `tfsdk:"vm_class_reservations"`
}

type vmClassReservationModel struct {
	VMClassName types.String `tfsdk:"vm_class_name"`
	Count       types.Int32  `tfsdk:"count"`
}

var (
	storageClassAttrTypes = map[string]attr.Type{
		"name": types.StringType, "limit": types.StringType,
	}
	contentSourceAttrTypes = map[string]attr.Type{
		"name": types.StringType, "type": types.StringType,
	}
	limitsAttrTypes = map[string]attr.Type{
		"cpu_limit":          types.StringType,
		"memory_limit":       types.StringType,
		"cpu_reservation":    types.StringType,
		"memory_reservation": types.StringType,
	}
	vmClassReservationAttrTypes = map[string]attr.Type{
		"vm_class_name": types.StringType, "count": types.Int32Type,
	}
	zoneAttrTypes = map[string]attr.Type{
		"name":               types.StringType,
		"cpu_limit":          types.StringType,
		"cpu_reservation":    types.StringType,
		"memory_limit":       types.StringType,
		"memory_reservation": types.StringType,
		"vm_class_reservations": types.SetType{
			ElemType: types.ObjectType{AttrTypes: vmClassReservationAttrTypes},
		},
	}

	storageClassObjectType  = types.ObjectType{AttrTypes: storageClassAttrTypes}
	contentSourceObjectType = types.ObjectType{AttrTypes: contentSourceAttrTypes}
	zoneObjectType          = types.ObjectType{AttrTypes: zoneAttrTypes}
	vmReservationObjectType = types.ObjectType{AttrTypes: vmClassReservationAttrTypes}
)
