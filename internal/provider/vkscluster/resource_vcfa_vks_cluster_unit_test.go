//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkscluster

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func TestParseOsImageAnnotationValue(t *testing.T) {
	tests := []struct {
		annotation  string
		wantName    string
		wantVersion string
	}{
		{"os-name=ubuntu", "ubuntu", ""},
		{"os-name=ubuntu,os-version=24.04", "ubuntu", "24.04"},
		{"os-version=24.04,os-name=ubuntu", "ubuntu", "24.04"},
		{"os-name=ubuntu, os-version=24.04", "ubuntu", "24.04"},
		{"os-name=photon,os-version=5", "photon", "5"},
	}
	for _, tt := range tests {
		name, version := parseOsImageAnnotationValue(tt.annotation)
		if name != tt.wantName || version != tt.wantVersion {
			t.Errorf("parseOsImageAnnotationValue(%q) = (%q, %q), want (%q, %q)",
				tt.annotation, name, version, tt.wantName, tt.wantVersion)
		}
	}
}

func TestOsImageAnnotationRoundTrip(t *testing.T) {
	m := vksClusterOsImageModel{Name: types.StringValue("ubuntu"), Version: types.StringValue("24.04")}
	name, version := parseOsImageAnnotationValue(buildOsImageAnnotationValue(m))
	if name != "ubuntu" || version != "24.04" {
		t.Errorf("round trip = (%q, %q), want (\"ubuntu\", \"24.04\")", name, version)
	}
}

func TestWarnOnClusterClassRebase(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics
	planned := helpers.ObjFrom(ctx, vksClusterClassRefAttrTypes, &vksClusterClassRefModel{
		Name:      types.StringValue("builtin-generic-v3.6.0"),
		Namespace: types.StringValue("vmware-system-vks-public"),
	}, &diags)

	cluster := func(className string) *vcfatypes.VksCluster {
		c := &vcfatypes.VksCluster{}
		c.Spec.Topology.ClassRef.Name = className
		c.Spec.Topology.Version = "v1.36.2+vmware.2"
		return c
	}

	tests := []struct {
		name        string
		planned     types.Object
		cluster     *vcfatypes.VksCluster
		wantRebased bool
	}{
		{"same class", planned, cluster("builtin-generic-v3.6.0"), false},
		{"rebased by backend", planned, cluster("builtin-generic-v3.7.0"), true},
		{"no topology", planned, &vcfatypes.VksCluster{}, false},
		{"unknown planned class", types.ObjectUnknown(vksClusterClassRefAttrTypes), cluster("builtin-generic-v3.7.0"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d diag.Diagnostics
			got := warnOnClusterClassRebase(ctx, tt.planned, tt.cluster, "test", &d)
			if got != tt.wantRebased {
				t.Errorf("warnOnClusterClassRebase() = %v, want %v", got, tt.wantRebased)
			}
			if got != (d.WarningsCount() == 1) || d.HasError() {
				t.Errorf("unexpected diagnostics: %v", d)
			}
		})
	}
}
