//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func TestRefId(t *testing.T) {
	if got := refId(nil); got != "" {
		t.Fatalf("expected empty ID for nil reference, got '%s'", got)
	}
	if got := refId(&types.OpenApiReference{ID: "urn:test"}); got != "urn:test" {
		t.Fatalf("expected 'urn:test', got '%s'", got)
	}
}

func TestDerefOr(t *testing.T) {
	if got := derefOr[int](nil); got != 0 {
		t.Fatalf("expected 0 for nil int pointer, got %d", got)
	}
	if got := derefOr(addrOf(int64(42))); got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
	if got := derefOr[bool](nil); got {
		t.Fatalf("expected false for nil bool pointer")
	}
	if got := derefOr(addrOf(true)); !got {
		t.Fatalf("expected true")
	}
}

func TestConvertSliceOfStringsToContentLibraryStorageClasses(t *testing.T) {
	got := convertSliceOfStringsToContentLibraryStorageClasses([]string{"a", "b"})
	want := []types.ContentLibraryStorageClass{{ID: "a"}, {ID: "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	if got := convertSliceOfStringsToContentLibraryStorageClasses(nil); len(got) != 0 {
		t.Fatalf("expected an empty slice, got %v", got)
	}
}

// A Regional Networking Setting that egresses through a distributed connection has no Provider Gateway
// reference, so reading it must not fail
func TestRegionalNetworkingSettingDataNilProviderGateway(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceVcfaOrgRegionalNetworking().Schema, map[string]interface{}{})
	setting := &govcd.TmRegionalNetworkingSetting{TmRegionalNetworkingSetting: &types.TmRegionalNetworkingSetting{
		ID:                           "urn:setting",
		Name:                         "test",
		OrgRef:                       types.OpenApiReference{ID: "urn:org"},
		RegionRef:                    types.OpenApiReference{ID: "urn:region"},
		DistributedVlanConnectionRef: &types.OpenApiReference{ID: "urn:vlan"},
	}}

	if err := setTmRegionalNetworkingSettingData(nil, d, setting); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got := d.Get("provider_gateway_id").(string); got != "" {
		t.Fatalf("expected empty provider_gateway_id, got '%s'", got)
	}
	if got := d.Get("distributed_vlan_connection_id").(string); got != "urn:vlan" {
		t.Fatalf("expected distributed_vlan_connection_id 'urn:vlan', got '%s'", got)
	}
}

// Exactly one of provider_gateway_id, distributed_vlan_connection_id and distributed_vxlan_connection_id is required
func TestRegionalNetworkingEgressExactlyOne(t *testing.T) {
	base := func() map[string]interface{} {
		return map[string]interface{}{"name": "n", "org_id": "o", "region_id": "r"}
	}
	tests := []struct {
		name    string
		extra   map[string]interface{}
		wantErr bool
	}{
		{"none", nil, true},
		{"provider gateway", map[string]interface{}{"provider_gateway_id": "pg"}, false},
		{"vlan", map[string]interface{}{"distributed_vlan_connection_id": "vlan"}, false},
		{"vxlan", map[string]interface{}{"distributed_vxlan_connection_id": "vxlan"}, false},
		{"gateway and vlan", map[string]interface{}{"provider_gateway_id": "pg", "distributed_vlan_connection_id": "vlan"}, true},
		{"vlan and vxlan", map[string]interface{}{"distributed_vlan_connection_id": "vlan", "distributed_vxlan_connection_id": "vxlan"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := base()
			for k, v := range tt.extra {
				cfg[k] = v
			}
			diags := resourceVcfaOrgRegionalNetworking().Validate(terraform.NewResourceConfigRaw(cfg))
			if tt.wantErr != diags.HasError() {
				t.Fatalf("expected error: %t, got diagnostics: %v", tt.wantErr, diags)
			}
		})
	}
}
