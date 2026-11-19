// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisornamespaceclass

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

// Patterns of the quantities, as defined by the CRD.
var (
	reCPUQuantity    = regexp.MustCompile(`(\d+)([MG])?`)
	reMemoryQuantity = regexp.MustCompile(`(\d+)([MGT]i)?`)
)

func (r *vcfaSupervisorNamespaceClassResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Resource for managing a %s, combining its SupervisorNamespaceClass and SupervisorNamespaceClassConfig", vcfatypes.LabelSupervisorNamespaceClass),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   fmt.Sprintf("Internal identifier of the %s, which is its name", vcfatypes.LabelSupervisorNamespaceClass),
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:      true,
				Description:   fmt.Sprintf("Name of the %s. It must be a lowercase DNS label of at most 63 characters", vcfatypes.LabelSupervisorNamespaceClass),
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(kubernetes.ReDNSLabel, "must consist of lowercase alphanumeric characters or '-', and must start and end with an alphanumeric character"),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: fmt.Sprintf("Description of the %s, of at most 2000 characters", vcfatypes.LabelSupervisorNamespaceClass),
				Validators:  []validator.String{stringvalidator.LengthAtMost(2000)},
			},
			"storage_classes": schema.SetNestedAttribute{
				Required:    true,
				Description: "Storage classes available in the Supervisor Namespaces created from this class, and their default limit. At least one is required",
				Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the storage class",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 253),
							},
						},
						"limit": schema.StringAttribute{
							Required:    true,
							Description: "Default storage limit of the storage class, e.g. `100Gi`",
							Validators: []validator.String{
								stringvalidator.RegexMatches(reMemoryQuantity, "must match the pattern "+reMemoryQuantity.String()),
							},
						},
					},
				},
			},
			"vm_classes": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Names of the VM classes available in the Supervisor Namespaces created from this class. At least one is required",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre([]validator.String{stringvalidator.LengthBetween(1, 253)}...),
				},
			},
			"content_sources": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Content sources available in the Supervisor Namespaces created from this class",
				Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the content source",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 253),
							},
						},
						"type": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(vcfatypes.SupervisorNamespaceClassContentSourceLibrary),
							Description: "Type of the content source. Only `ContentLibrary` is supported",
							Validators:  []validator.String{stringvalidator.OneOf(vcfatypes.SupervisorNamespaceClassContentSourceLibrary)},
						},
					},
				},
			},
			"default_limits": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Default resource limits per zone. Exactly one of `default_limits` and `zones` is required",
				Attributes: map[string]schema.Attribute{
					"cpu_limit": schema.StringAttribute{
						Required:    true,
						Description: "CPU limit per zone, e.g. `10` or `5000m`",
						Validators: []validator.String{
							stringvalidator.RegexMatches(reCPUQuantity, "must match the pattern "+reCPUQuantity.String()),
						},
					},
					"memory_limit": schema.StringAttribute{
						Required:    true,
						Description: "Memory limit per zone, e.g. `64Gi`",
						Validators: []validator.String{
							stringvalidator.RegexMatches(reMemoryQuantity, "must match the pattern "+reMemoryQuantity.String()),
						},
					},
					"cpu_reservation": schema.StringAttribute{
						Required:    true,
						Description: "CPU reservation per zone, e.g. `4` or `2000m`",
						Validators: []validator.String{
							stringvalidator.RegexMatches(reCPUQuantity, "must match the pattern "+reCPUQuantity.String()),
						},
					},
					"memory_reservation": schema.StringAttribute{
						Required:    true,
						Description: "Memory reservation per zone, e.g. `32Gi`",
						Validators: []validator.String{
							stringvalidator.RegexMatches(reMemoryQuantity, "must match the pattern "+reMemoryQuantity.String()),
						},
					},
				},
				Validators: []validator.Object{objectvalidator.ExactlyOneOf(path.MatchRoot("zones"))},
			},
			"zones": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Zones with their own limits. Exactly one of `default_limits` and `zones` is required",
				Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the zone these limits apply to",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 253),
							},
						},
						"cpu_limit": schema.StringAttribute{
							Required:    true,
							Description: "CPU limit of the zone, e.g. `10` or `5000m`",
							Validators: []validator.String{
								stringvalidator.RegexMatches(reCPUQuantity, "must match the pattern "+reCPUQuantity.String()),
							},
						},
						"memory_limit": schema.StringAttribute{
							Required:    true,
							Description: "Memory limit of the zone, e.g. `64Gi`",
							Validators: []validator.String{
								stringvalidator.RegexMatches(reMemoryQuantity, "must match the pattern "+reMemoryQuantity.String()),
							},
						},
						"cpu_reservation": schema.StringAttribute{
							Required:    true,
							Description: "CPU reservation of the zone, e.g. `4` or `2000m`",
							Validators: []validator.String{
								stringvalidator.RegexMatches(reCPUQuantity, "must match the pattern "+reCPUQuantity.String()),
							},
						},
						"memory_reservation": schema.StringAttribute{
							Required:    true,
							Description: "Memory reservation of the zone, e.g. `32Gi`",
							Validators: []validator.String{
								stringvalidator.RegexMatches(reMemoryQuantity, "must match the pattern "+reMemoryQuantity.String()),
							},
						},
						"vm_class_reservations": schema.SetNestedAttribute{
							Optional:    true,
							Description: "VM classes to reserve in the zone by default",
							Validators:  []validator.Set{setvalidator.SizeAtLeast(1)},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"vm_class_name": schema.StringAttribute{
										Required:    true,
										Description: "Name of the VM class to reserve instances of",
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 253),
										},
									},
									"count": schema.Int32Attribute{
										Required:    true,
										Description: "Number of instances to reserve. It must be at least 1",
										Validators:  []validator.Int32{int32validator.AtLeast(1)},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
