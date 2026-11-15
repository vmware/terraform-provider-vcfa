//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package project

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func TestProjectRoleBindingName(t *testing.T) {
	tests := []struct {
		kind, name, want string
	}{
		{vcfatypes.ProjectRoleBindingSubjectKindUser, "alice", "cci:user:alice"},
		{vcfatypes.ProjectRoleBindingSubjectKindUser, "alice@vsphere.local", "cci:user:vsphere.local:alice"},
		{vcfatypes.ProjectRoleBindingSubjectKindUser, "alice@", "cci:user:alice"},
		{vcfatypes.ProjectRoleBindingSubjectKindUser, "a@b@corp.com", "cci:user:corp.com:a@b"},
		{vcfatypes.ProjectRoleBindingSubjectKindGroup, "devs", "cci:group:devs"},
		{vcfatypes.ProjectRoleBindingSubjectKindGroup, "devs@", "cci:group:devs"},
		{vcfatypes.ProjectRoleBindingSubjectKindGroup, "devs@corp.com", "cci:group:corp.com:devs"},
	}
	for _, tt := range tests {
		t.Run(tt.kind+"/"+tt.name, func(t *testing.T) {
			if got := vcfatypes.ProjectRoleBindingName(tt.kind, tt.name); got != tt.want {
				t.Errorf("ProjectRoleBindingName(%q, %q) = %q, want %q", tt.kind, tt.name, got, tt.want)
			}
		})
	}
}

func TestNewProjectRoleBindingGroupSubject(t *testing.T) {
	for in, want := range map[string]string{"devs": "devs@", "devs@": "devs@", "devs@corp.com": "devs@corp.com"} {
		got := vcfatypes.NewProjectRoleBinding("p", vcfatypes.ProjectRoleBindingSubjectKindGroup, in, vcfatypes.ProjectRoleView)
		if got.Subjects[0].Name != want {
			t.Errorf("group %q: subject %q, want %q", in, got.Subjects[0].Name, want)
		}
	}
}

func TestNewProjectRoleBinding(t *testing.T) {
	got := vcfatypes.NewProjectRoleBinding("my-project", vcfatypes.ProjectRoleBindingSubjectKindUser, "alice@vsphere.local", vcfatypes.ProjectRoleEdit)

	if got.Name != "cci:user:vsphere.local:alice" || got.Namespace != "my-project" {
		t.Errorf("unexpected metadata %s/%s", got.Namespace, got.Name)
	}
	if got.APIVersion != "authorization.cci.vmware.com/v1alpha1" || got.Kind != "ProjectRoleBinding" {
		t.Errorf("unexpected type %s %s", got.APIVersion, got.Kind)
	}
	wantSubjects := []vcfatypes.ProjectRoleBindingSubject{{Kind: "User", Name: "alice@vsphere.local"}}
	if !reflect.DeepEqual(got.Subjects, wantSubjects) {
		t.Errorf("subjects = %+v, want %+v", got.Subjects, wantSubjects)
	}
	wantRoleRef := vcfatypes.ProjectRoleBindingRoleRef{APIGroup: "authorization.cci.vmware.com", Kind: "ProjectRole", Name: "edit"}
	if got.RoleRef != wantRoleRef {
		t.Errorf("roleRef = %+v, want %+v", got.RoleRef, wantRoleRef)
	}
}

func member(kind, name, role string) projectMemberModel {
	return projectMemberModel{Kind: types.StringValue(kind), Name: types.StringValue(name), Role: types.StringValue(role)}
}

func binding(kind, name, role string) vcfatypes.ProjectRoleBinding {
	return vcfatypes.NewProjectRoleBinding("p", kind, name, role)
}

func TestBindingChanges(t *testing.T) {
	live := []vcfatypes.ProjectRoleBinding{
		binding("User", "alice", "admin"),    // kept
		binding("User", "bob", "view"),       // role changed
		binding("Group", "old-team", "edit"), // removed
	}
	desired := desiredBindings("p", []projectMemberModel{
		member("User", "alice", "admin"),
		member("User", "bob", "edit"),
		member("Group", "new-team", "edit_adv"),
	})

	toDelete, toUpdate, toCreate := bindingChanges(live, desired)

	if want := []string{"cci:group:old-team"}; !reflect.DeepEqual(toDelete, want) {
		t.Errorf("toDelete = %v, want %v", toDelete, want)
	}
	if len(toUpdate) != 1 || toUpdate[0].Name != "cci:user:bob" || toUpdate[0].RoleRef.Name != "edit" {
		t.Errorf("toUpdate = %+v, want bob as edit", toUpdate)
	}
	if len(toCreate) != 1 || toCreate[0].Name != "cci:group:new-team" || toCreate[0].RoleRef.Name != "edit_adv" {
		t.Errorf("toCreate = %+v, want new-team as edit_adv", toCreate)
	}

	t.Run("no members removes every binding", func(t *testing.T) {
		toDelete, toUpdate, toCreate := bindingChanges(live, desiredBindings("p", nil))
		if len(toDelete) != 3 || len(toUpdate) != 0 || len(toCreate) != 0 {
			t.Errorf("got delete=%v update=%v create=%v, want 3 deletes only", toDelete, toUpdate, toCreate)
		}
	})

	t.Run("matching bindings need no calls", func(t *testing.T) {
		toDelete, toUpdate, toCreate := bindingChanges(live[:1], desiredBindings("p", []projectMemberModel{member("User", "alice", "admin")}))
		if len(toDelete)+len(toUpdate)+len(toCreate) != 0 {
			t.Errorf("got delete=%v update=%v create=%v, want no calls", toDelete, toUpdate, toCreate)
		}
	})
}

func TestMapBindingsToMembers(t *testing.T) {
	ctx := context.Background()

	// The backend echoes a group's name with a trailing '@'.
	groupBinding := binding("Group", "devs", "edit")
	groupBinding.Subjects[0].Name = "devs@"
	bindings := []vcfatypes.ProjectRoleBinding{
		binding("User", "bob@corp.com", "view"),
		groupBinding,
		{ObjectMeta: binding("User", "ghost", "view").ObjectMeta}, // no subject: ignored
	}

	t.Run("without prior members", func(t *testing.T) {
		var diags diag.Diagnostics
		got := extractMembers(ctx, mapBindingsToMembers(ctx, bindings, nil, &diags), &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		want := []projectMemberModel{
			member("Group", "devs", "edit"),
			member("User", "bob@corp.com", "view"),
		}
		if !membersEqual(got, want) {
			t.Errorf("members = %v, want %v", got, want)
		}
	})

	t.Run("prior names are kept as configured", func(t *testing.T) {
		var diags diag.Diagnostics
		prior := []projectMemberModel{member("Group", "devs@", "view")}
		got := extractMembers(ctx, mapBindingsToMembers(ctx, bindings, prior, &diags), &diags)
		want := []projectMemberModel{
			member("Group", "devs@", "edit"),
			member("User", "bob@corp.com", "view"),
		}
		if !membersEqual(got, want) {
			t.Errorf("members = %v, want %v", got, want)
		}
	})

	t.Run("no bindings is an empty set, not null", func(t *testing.T) {
		var diags diag.Diagnostics
		got := mapBindingsToMembers(ctx, nil, nil, &diags)
		if got.IsNull() || len(got.Elements()) != 0 {
			t.Errorf("members = %v, want an empty set", got)
		}
	})
}

func membersEqual(a, b []projectMemberModel) bool {
	if len(a) != len(b) {
		return false
	}
	count := map[projectMemberModel]int{}
	for _, m := range a {
		count[m]++
	}
	for _, m := range b {
		count[m]--
	}
	for _, c := range count {
		if c != 0 {
			return false
		}
	}
	return true
}

func TestValidateConfigDuplicateMembers(t *testing.T) {
	ctx := context.Background()
	r := &vcfaProjectResource{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	memberType := objType.AttributeTypes["members"].(tftypes.Set).ElementType

	memberValue := func(kind, name, role string) tftypes.Value {
		return tftypes.NewValue(memberType, map[string]tftypes.Value{
			"kind": tftypes.NewValue(tftypes.String, kind),
			"name": tftypes.NewValue(tftypes.String, name),
			"role": tftypes.NewValue(tftypes.String, role),
		})
	}
	configWith := func(members ...tftypes.Value) tfsdk.Config {
		var membersValue tftypes.Value
		if members == nil {
			membersValue = tftypes.NewValue(objType.AttributeTypes["members"], nil)
		} else {
			membersValue = tftypes.NewValue(objType.AttributeTypes["members"], members)
		}
		return tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
				"id":          tftypes.NewValue(tftypes.String, nil),
				"name":        tftypes.NewValue(tftypes.String, "p"),
				"description": tftypes.NewValue(tftypes.String, nil),
				"uid":         tftypes.NewValue(tftypes.String, nil),
				"members":     membersValue,
			}),
		}
	}

	tests := []struct {
		name    string
		config  tfsdk.Config
		wantErr bool
	}{
		{"no members", configWith(), false},
		{"distinct members", configWith(memberValue("User", "alice", "admin"), memberValue("Group", "alice", "view")), false},
		{"same user with two roles", configWith(memberValue("User", "alice", "admin"), memberValue("User", "alice", "view")), true},
		{"same user written two ways", configWith(memberValue("User", "alice", "admin"), memberValue("User", "alice@", "admin")), true},
		{"unknown name is not validated yet", configWith(memberValue("User", "alice", "admin"), tftypes.NewValue(memberType, map[string]tftypes.Value{
			"kind": tftypes.NewValue(tftypes.String, "User"),
			"name": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"role": tftypes.NewValue(tftypes.String, "view"),
		})), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &resource.ValidateConfigResponse{}
			r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tt.config}, resp)
			if got := resp.Diagnostics.HasError(); got != tt.wantErr {
				t.Errorf("HasError() = %t, want %t (%v)", got, tt.wantErr, resp.Diagnostics)
			}
		})
	}
}
