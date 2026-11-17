//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisornamespaceclass

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func TestSupervisorNamespaceClassConfigRoundTrip(t *testing.T) {
	ctx := context.Background()
	want := vcfatypes.SupervisorNamespaceClassConfigSpec{
		StorageClasses: []vcfatypes.SupervisorNamespaceClassStorageClass{{Name: "gold", Limit: "100Gi"}},
		VMClasses:      []vcfatypes.SupervisorNamespaceClassVMClass{{Name: "best-effort-small"}},
		ContentSources: []vcfatypes.SupervisorNamespaceClassContentSrc{{Name: "lib", Type: "ContentLibrary"}},
		DefaultLimits:  &vcfatypes.SupervisorNamespaceClassLimits{CPULimit: "10", MemoryLimit: "64Gi", CPUReservation: "4", MemoryReservation: "32Gi"},
		Zones: []vcfatypes.SupervisorNamespaceClassZone{{
			Name: "zone-1", CPULimit: "5", CPUReservation: "2", MemoryLimit: "8Gi", MemoryReservation: "4Gi",
			VMClassReservations: []vcfatypes.SupervisorNamespaceClassVMReservation{{VMClassName: "best-effort-small", Count: 2}},
		}},
	}

	var diags diag.Diagnostics
	var model vcfaSupervisorNamespaceClassResourceModel
	mapSupervisorNamespaceClassConfigToResourceModel(ctx, &vcfatypes.SupervisorNamespaceClassConfig{Spec: want}, &model, &diags)
	model.Name = types.StringValue("my-class")
	got := buildSupervisorNamespaceClassConfig(ctx, &model, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !reflect.DeepEqual(got.Spec, want) {
		t.Errorf("spec = %+v, want %+v", got.Spec, want)
	}
	if got.Name != "my-class" || got.Kind != "SupervisorNamespaceClassConfig" || got.APIVersion != "infrastructure.cci.vmware.com/v1alpha2" {
		t.Errorf("unexpected metadata %s %s %s", got.APIVersion, got.Kind, got.Name)
	}
}

func TestMapSupervisorNamespaceClassConfigEmptyOptionals(t *testing.T) {
	var diags diag.Diagnostics
	var model vcfaSupervisorNamespaceClassResourceModel
	mapSupervisorNamespaceClassConfigToResourceModel(context.Background(), &vcfatypes.SupervisorNamespaceClassConfig{
		Spec: vcfatypes.SupervisorNamespaceClassConfigSpec{
			StorageClasses: []vcfatypes.SupervisorNamespaceClassStorageClass{{Name: "gold", Limit: "1Gi"}},
			VMClasses:      []vcfatypes.SupervisorNamespaceClassVMClass{{Name: "small"}},
		},
	}, &model, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !model.ContentSources.IsNull() || !model.DefaultLimits.IsNull() || !model.Zones.IsNull() {
		t.Errorf("absent optional attributes must be null: %v %v %v", model.ContentSources, model.DefaultLimits, model.Zones)
	}
}
