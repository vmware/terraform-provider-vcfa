// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package project

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

var projectMemberKinds = []string{
	vcfatypes.ProjectRoleBindingSubjectKindUser,
	vcfatypes.ProjectRoleBindingSubjectKindGroup,
}

var projectMemberRoles = []string{
	vcfatypes.ProjectRoleAdmin,
	vcfatypes.ProjectRoleEdit,
	vcfatypes.ProjectRoleEditAdv,
	vcfatypes.ProjectRoleView,
}

func (r *vcfaProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Resource for managing a %s and the users and groups that have access to it", vcfatypes.LabelProject),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Internal identifier of the %s, which is its name", vcfatypes.LabelProject),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Required attributes
			"name": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Name of the %s. It must be a lowercase DNS label of at most 63 characters", vcfatypes.LabelProject),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(kubernetes.ReDNSLabel, "must consist of lowercase alphanumeric characters or '-', and must start and end with an alphanumeric character"),
				},
			},

			// Spec attributes
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: fmt.Sprintf("Description of the %s", vcfatypes.LabelProject),
				Validators: []validator.String{
					stringvalidator.LengthAtMost(2000),
				},
			},
			"uid": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Unique identifier of the %s", vcfatypes.LabelProject),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Access attributes
			"members": schema.SetNestedAttribute{
				Optional: true,
				Description: fmt.Sprintf("Users and groups that have access to the %s, and their role. When set, the list is authoritative: "+
					"users and groups assigned to the %s by other means are unassigned. When not set, the members of the %s are not managed. "+
					"Each user or group can be assigned a single role", vcfatypes.LabelProject, vcfatypes.LabelProject, vcfatypes.LabelProject),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.StringAttribute{
							Required:    true,
							Description: "Kind of member: `User` or `Group`",
							Validators: []validator.String{
								stringvalidator.OneOf(projectMemberKinds...),
							},
						},
						"name": schema.StringAttribute{
							Required: true,
							Description: "Name of the user or group, as known to the identity provider of the Organization. " +
								"A user from a specific domain can be given as `<name>@<domain>`",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 253),
								stringvalidator.RegexMatches(regexp.MustCompile(`^[^:]+$`), "must not contain ':'"),
							},
						},
						"role": schema.StringAttribute{
							Required:    true,
							Description: "Role of the member in the Project: " + strings.Join(projectMemberRoles, ","),
							Validators: []validator.String{
								stringvalidator.OneOf(projectMemberRoles...),
							},
						},
					},
				},
			},
		},
	}
}
