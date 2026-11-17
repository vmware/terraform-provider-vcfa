// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisornamespaceclass

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func (d *vcfaSupervisorNamespaceClassDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Data source for reading a %s, combining its SupervisorNamespaceClass and SupervisorNamespaceClassConfig", vcfatypes.LabelSupervisorNamespaceClass),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Internal identifier of the %s, which is its name", vcfatypes.LabelSupervisorNamespaceClass),
			},

			// Required lookup attributes
			"name": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", vcfatypes.LabelSupervisorNamespaceClass),
			},

			// Spec attributes
			"description": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Description of the %s", vcfatypes.LabelSupervisorNamespaceClass),
			},

			// Config attributes
			"storage_classes": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Storage classes available in the Supervisor Namespaces created from this class, and their default limit",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the storage class",
						},
						"limit": schema.StringAttribute{
							Computed:    true,
							Description: "Default storage limit of the storage class",
						},
					},
				},
			},
			"vm_classes": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Names of the VM classes available in the Supervisor Namespaces created from this class",
			},
			"content_sources": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Content sources available in the Supervisor Namespaces created from this class",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the content source",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "Type of the content source",
						},
					},
				},
			},
			"default_limits": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Default resource limits per zone",
				Attributes: map[string]schema.Attribute{
					"cpu_limit": schema.StringAttribute{
						Computed:    true,
						Description: "CPU limit per zone",
					},
					"memory_limit": schema.StringAttribute{
						Computed:    true,
						Description: "Memory limit per zone",
					},
					"cpu_reservation": schema.StringAttribute{
						Computed:    true,
						Description: "CPU reservation per zone",
					},
					"memory_reservation": schema.StringAttribute{
						Computed:    true,
						Description: "Memory reservation per zone",
					},
				},
			},
			"zones": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Zones with their own limits",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the zone these limits apply to",
						},
						"cpu_limit": schema.StringAttribute{
							Computed:    true,
							Description: "CPU limit of the zone",
						},
						"memory_limit": schema.StringAttribute{
							Computed:    true,
							Description: "Memory limit of the zone",
						},
						"cpu_reservation": schema.StringAttribute{
							Computed:    true,
							Description: "CPU reservation of the zone",
						},
						"memory_reservation": schema.StringAttribute{
							Computed:    true,
							Description: "Memory reservation of the zone",
						},
						"vm_class_reservations": schema.SetNestedAttribute{
							Computed:    true,
							Description: "VM classes reserved in the zone by default",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"vm_class_name": schema.StringAttribute{
										Computed:    true,
										Description: "Name of the VM class",
									},
									"count": schema.Int32Attribute{
										Computed:    true,
										Description: "Number of instances reserved",
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
