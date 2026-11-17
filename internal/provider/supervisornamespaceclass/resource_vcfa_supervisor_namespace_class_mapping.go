// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisornamespaceclass

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func typeMeta(kind string) metav1.TypeMeta {
	return metav1.TypeMeta{Kind: kind, APIVersion: vcfatypes.SupervisorNamespaceClassAPI + "/" + vcfatypes.SupervisorNamespaceClassVersion}
}

// buildSupervisorNamespaceClass builds the SupervisorNamespaceClass to create or update.
func buildSupervisorNamespaceClass(model *vcfaSupervisorNamespaceClassResourceModel) vcfatypes.SupervisorNamespaceClass {
	return vcfatypes.SupervisorNamespaceClass{
		TypeMeta: typeMeta(vcfatypes.SupervisorNamespaceClassKind),
		ObjectMeta: metav1.ObjectMeta{
			Name: model.Name.ValueString(),
		},
		Spec: vcfatypes.SupervisorNamespaceClassSpec{
			Description: model.Description.ValueString(),
		},
	}
}

// buildSupervisorNamespaceClassConfig builds the SupervisorNamespaceClassConfig, named after its class, to create or update.
func buildSupervisorNamespaceClassConfig(ctx context.Context, model *vcfaSupervisorNamespaceClassResourceModel, diags *diag.Diagnostics) vcfatypes.SupervisorNamespaceClassConfig {
	spec := vcfatypes.SupervisorNamespaceClassConfigSpec{
		StorageClasses: []vcfatypes.SupervisorNamespaceClassStorageClass{},
		VMClasses:      []vcfatypes.SupervisorNamespaceClassVMClass{},
	}

	for _, s := range helpers.Elements[storageClassModel](ctx, model.StorageClasses, diags) {
		spec.StorageClasses = append(spec.StorageClasses, vcfatypes.SupervisorNamespaceClassStorageClass{
			Name:  s.Name.ValueString(),
			Limit: s.Limit.ValueString(),
		})
	}
	for _, name := range helpers.Elements[types.String](ctx, model.VMClasses, diags) {
		spec.VMClasses = append(spec.VMClasses, vcfatypes.SupervisorNamespaceClassVMClass{Name: name.ValueString()})
	}
	for _, c := range helpers.Elements[contentSourceModel](ctx, model.ContentSources, diags) {
		spec.ContentSources = append(spec.ContentSources, vcfatypes.SupervisorNamespaceClassContentSrc{
			Name: c.Name.ValueString(), Type: c.Type.ValueString(),
		})
	}
	if !model.DefaultLimits.IsNull() && !model.DefaultLimits.IsUnknown() {
		var l limitsModel
		diags.Append(model.DefaultLimits.As(ctx, &l, basetypes.ObjectAsOptions{})...)
		spec.DefaultLimits = &vcfatypes.SupervisorNamespaceClassLimits{
			CPULimit:          l.CPULimit.ValueString(),
			MemoryLimit:       l.MemoryLimit.ValueString(),
			CPUReservation:    l.CPUReservation.ValueString(),
			MemoryReservation: l.MemoryReservation.ValueString(),
		}
	}
	for _, z := range helpers.Elements[zoneModel](ctx, model.Zones, diags) {
		zone := vcfatypes.SupervisorNamespaceClassZone{
			Name:              z.Name.ValueString(),
			CPULimit:          z.CPULimit.ValueString(),
			CPUReservation:    z.CPUReservation.ValueString(),
			MemoryLimit:       z.MemoryLimit.ValueString(),
			MemoryReservation: z.MemoryReservation.ValueString(),
		}
		for _, r := range helpers.Elements[vmClassReservationModel](ctx, z.VMClassReservations, diags) {
			zone.VMClassReservations = append(zone.VMClassReservations, vcfatypes.SupervisorNamespaceClassVMReservation{
				VMClassName: r.VMClassName.ValueString(),
				Count:       r.Count.ValueInt32(),
			})
		}
		spec.Zones = append(spec.Zones, zone)
	}

	return vcfatypes.SupervisorNamespaceClassConfig{
		TypeMeta:   typeMeta(vcfatypes.SupervisorNamespaceClassConfigKind),
		ObjectMeta: metav1.ObjectMeta{Name: model.Name.ValueString()},
		Spec:       spec,
	}
}

// mapSupervisorNamespaceClassToResourceModel populates the class attributes of the model.
func mapSupervisorNamespaceClassToResourceModel(class *vcfatypes.SupervisorNamespaceClass, model *vcfaSupervisorNamespaceClassResourceModel) {
	model.ID = types.StringValue(class.Name)
	model.Name = types.StringValue(class.Name)
	model.Description = types.StringValue(class.Spec.Description)
}

// mapSupervisorNamespaceClassConfigToResourceModel populates the config attributes of the model.
func mapSupervisorNamespaceClassConfigToResourceModel(ctx context.Context, config *vcfatypes.SupervisorNamespaceClassConfig, model *vcfaSupervisorNamespaceClassResourceModel, diags *diag.Diagnostics) {
	spec := config.Spec

	storage := make([]storageClassModel, 0, len(spec.StorageClasses))
	for _, s := range spec.StorageClasses {
		storage = append(storage, storageClassModel{Name: types.StringValue(s.Name), Limit: types.StringValue(s.Limit)})
	}
	model.StorageClasses = helpers.SetFrom(ctx, storageClassObjectType, storage, diags)

	vmClasses := make([]string, 0, len(spec.VMClasses))
	for _, v := range spec.VMClasses {
		vmClasses = append(vmClasses, v.Name)
	}
	model.VMClasses = helpers.SetFrom(ctx, types.StringType, vmClasses, diags)

	if len(spec.ContentSources) == 0 {
		model.ContentSources = types.SetNull(contentSourceObjectType)
	} else {
		sources := make([]contentSourceModel, 0, len(spec.ContentSources))
		for _, c := range spec.ContentSources {
			sources = append(sources, contentSourceModel{Name: types.StringValue(c.Name), Type: types.StringValue(c.Type)})
		}
		model.ContentSources = helpers.SetFrom(ctx, contentSourceObjectType, sources, diags)
	}

	if spec.DefaultLimits == nil {
		model.DefaultLimits = types.ObjectNull(limitsAttrTypes)
	} else {
		model.DefaultLimits = helpers.ObjFrom(ctx, limitsAttrTypes, limitsModel{
			CPULimit:          types.StringValue(spec.DefaultLimits.CPULimit),
			MemoryLimit:       types.StringValue(spec.DefaultLimits.MemoryLimit),
			CPUReservation:    types.StringValue(spec.DefaultLimits.CPUReservation),
			MemoryReservation: types.StringValue(spec.DefaultLimits.MemoryReservation),
		}, diags)
	}

	if len(spec.Zones) == 0 {
		model.Zones = types.SetNull(zoneObjectType)
		return
	}
	zones := make([]zoneModel, 0, len(spec.Zones))
	for _, z := range spec.Zones {
		reservationsSet := types.SetNull(vmReservationObjectType)
		if len(z.VMClassReservations) > 0 {
			reservations := make([]vmClassReservationModel, 0, len(z.VMClassReservations))
			for _, r := range z.VMClassReservations {
				reservations = append(reservations, vmClassReservationModel{VMClassName: types.StringValue(r.VMClassName), Count: types.Int32Value(r.Count)})
			}
			sort.Slice(reservations, func(i, j int) bool {
				return reservations[i].VMClassName.ValueString() < reservations[j].VMClassName.ValueString()
			})
			reservationsSet = helpers.SetFrom(ctx, vmReservationObjectType, reservations, diags)
		}
		zones = append(zones, zoneModel{
			Name:                types.StringValue(z.Name),
			CPULimit:            types.StringValue(z.CPULimit),
			CPUReservation:      types.StringValue(z.CPUReservation),
			MemoryLimit:         types.StringValue(z.MemoryLimit),
			MemoryReservation:   types.StringValue(z.MemoryReservation),
			VMClassReservations: reservationsSet,
		})
	}
	model.Zones = helpers.SetFrom(ctx, zoneObjectType, zones, diags)
}
