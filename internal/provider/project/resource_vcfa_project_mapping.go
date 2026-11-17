// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package project

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

// mapProjectToResourceModel populates the Project attributes of the model. Members are mapped
// separately, as they come from the Project's role bindings.
func mapProjectToResourceModel(project *vcfatypes.Project, model *vcfaProjectResourceModel) {
	model.ID = types.StringValue(project.Name)
	model.Name = types.StringValue(project.Name)
	model.Description = types.StringValue(project.Spec.Description)
	model.UID = types.StringValue(string(project.UID))
}

// extractMembers returns the members of a known set, or nil when the set is null or unknown.
func extractMembers(ctx context.Context, set types.Set, diags *diag.Diagnostics) []projectMemberModel {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var members []projectMemberModel
	diags.Append(set.ElementsAs(ctx, &members, false)...)
	return members
}

// memberBindingName returns the name of the role binding that grants the member its role.
func memberBindingName(member projectMemberModel) string {
	return vcfatypes.ProjectRoleBindingName(member.Kind.ValueString(), member.Name.ValueString())
}

// desiredBindings returns the role bindings that grant the members their role, keyed by name.
func desiredBindings(projectName string, members []projectMemberModel) map[string]vcfatypes.ProjectRoleBinding {
	bindings := make(map[string]vcfatypes.ProjectRoleBinding, len(members))
	for _, m := range members {
		b := vcfatypes.NewProjectRoleBinding(projectName, m.Kind.ValueString(), m.Name.ValueString(), m.Role.ValueString())
		bindings[b.Name] = b
	}
	return bindings
}

// mapBindingsToMembers converts the role bindings of a Project into a members set. The backend
// may echo a subject name in a different form than it was given (a group named `devs` can be
// returned as `devs@`), so the name of a prior member whose binding still exists is kept as it
// was configured.
func mapBindingsToMembers(ctx context.Context, bindings []vcfatypes.ProjectRoleBinding, prior []projectMemberModel, diags *diag.Diagnostics) types.Set {
	priorNames := make(map[string]string, len(prior))
	for _, m := range prior {
		priorNames[memberBindingName(m)] = m.Name.ValueString()
	}

	members := make([]projectMemberModel, 0, len(bindings))
	for _, b := range bindings {
		if len(b.Subjects) == 0 {
			continue
		}
		subject := b.Subjects[0]
		name, ok := priorNames[b.Name]
		if !ok {
			name = subject.Name
			if subject.Kind == vcfatypes.ProjectRoleBindingSubjectKindGroup {
				name = strings.TrimSuffix(name, "@")
			}
		}
		members = append(members, projectMemberModel{
			Kind: types.StringValue(subject.Kind),
			Name: types.StringValue(name),
			Role: types.StringValue(b.RoleRef.Name),
		})
	}
	sort.Slice(members, func(i, j int) bool { return memberBindingName(members[i]) < memberBindingName(members[j]) })

	return helpers.SetFrom(ctx, projectMemberObjectType, members, diags)
}

// bindingChanges computes the calls that turn the live role bindings into the desired ones: the
// names of the bindings to delete, and the bindings to update (role changes) and create. Each
// list is sorted by name so that the calls are made in a stable order.
func bindingChanges(live []vcfatypes.ProjectRoleBinding, desired map[string]vcfatypes.ProjectRoleBinding) (toDelete []string, toUpdate, toCreate []vcfatypes.ProjectRoleBinding) {
	liveByName := make(map[string]vcfatypes.ProjectRoleBinding, len(live))
	for _, b := range live {
		liveByName[b.Name] = b
		if _, ok := desired[b.Name]; !ok {
			toDelete = append(toDelete, b.Name)
		}
	}

	for name, d := range desired {
		l, ok := liveByName[name]
		switch {
		case !ok:
			toCreate = append(toCreate, d)
		case l.RoleRef.Name != d.RoleRef.Name:
			toUpdate = append(toUpdate, d)
		}
	}

	sort.Strings(toDelete)
	byName := func(s []vcfatypes.ProjectRoleBinding) func(i, j int) bool {
		return func(i, j int) bool { return s[i].Name < s[j].Name }
	}
	sort.Slice(toUpdate, byName(toUpdate))
	sort.Slice(toCreate, byName(toCreate))
	return toDelete, toUpdate, toCreate
}
