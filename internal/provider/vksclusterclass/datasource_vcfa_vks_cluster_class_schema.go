// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vksclusterclass

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/common"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func (d *vcfaVksClusterClassDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	vksClusterClassTemplateReferenceAttrs := map[string]schema.Attribute{
		"api_version": schema.StringAttribute{
			Computed:    true,
			Description: "API version of the referenced resource",
		},
		"kind": schema.StringAttribute{
			Computed:    true,
			Description: "Kind of the referenced resource",
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: "Name of the referenced resource",
		},
	}

	vksClusterClassNamingAttrs := map[string]schema.Attribute{
		"template": schema.StringAttribute{
			Computed:    true,
			Description: "Go template used to generate the object name"},
	}

	vksClusterClassObjectMetaAttrs := schema.SingleNestedAttribute{
		Computed:    true,
		Description: "Standard Kubernetes object metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
		Attributes: map[string]schema.Attribute{
			"labels": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Map of string keys and values used to organize and categorize the object",
			},
			"annotations": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Unstructured key-value pairs set by external tools to store and retrieve arbitrary metadata",
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Data source for reading a %s", vcfatypes.LabelVksClusterClass),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Internal identifier of the %s", vcfatypes.LabelVksClusterClass),
			},

			// Required lookup attributes
			"context": common.VcfContextDataSourceSchema,
			"name": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", vcfatypes.LabelVksClusterClass),
			},
			"system": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf("Whether the %s is system-wide or not", vcfatypes.LabelVksClusterClass),
			},

			// Metadata attributes
			"metadata": kubernetes.MetadataDataSourceSchema,

			// Spec attributes
			"availability_gates": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Additional conditions to include when evaluating Cluster Available conditions",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"condition_type": schema.StringAttribute{
							Computed:    true,
							Description: "Condition type in the Cluster's condition list to use as an availability gate",
						},
						"polarity": schema.StringAttribute{
							Computed:    true,
							Description: "Polarity of the condition: Positive (true = healthy) or Negative (false = healthy)",
						},
					},
				},
			},
			"infrastructure": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Infrastructure class: provider-specific infrastructure cluster template",
				Attributes: map[string]schema.Attribute{
					"template_ref": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Reference to the infrastructure cluster template",
						Attributes:  vksClusterClassTemplateReferenceAttrs,
					},
					"naming": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Naming strategy for the infrastructure object",
						Attributes:  vksClusterClassNamingAttrs,
					},
				},
			},
			"control_plane": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Control plane class: control plane template and optional machine infrastructure",
				Attributes: map[string]schema.Attribute{
					"metadata": vksClusterClassObjectMetaAttrs,
					"template_ref": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Reference to the control plane template",
						Attributes:  vksClusterClassTemplateReferenceAttrs,
					},
					"machine_infrastructure": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Infrastructure template for individual control plane machines (machine-based providers only)",
						Attributes: map[string]schema.Attribute{
							"template_ref": schema.SingleNestedAttribute{
								Computed:    true,
								Description: "Reference to the machine infrastructure template",
								Attributes:  vksClusterClassTemplateReferenceAttrs,
							},
						},
					},
					"health_check": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "MachineHealthCheck configuration for control plane machines",
						Attributes: map[string]schema.Attribute{
							"checks": schema.SingleNestedAttribute{
								Computed:    true,
								Description: "Conditions that mark a worker machine as unhealthy",
								Attributes: map[string]schema.Attribute{
									"node_startup_timeout_seconds": schema.Int32Attribute{
										Computed:    true,
										Description: "Max seconds before a node must have a ProviderID; 0 disables the check (default 600)",
									},
									"unhealthy_node_conditions": schema.SetNestedAttribute{
										Computed:    true,
										Description: "Node conditions that mark a machine unhealthy (logical OR)",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													Computed:    true,
													Description: "Node condition type",
												},
												"status": schema.StringAttribute{
													Computed:    true,
													Description: "Required condition status (True/False/Unknown)",
												},
												"timeout_seconds": schema.Int32Attribute{
													Computed:    true,
													Description: "Duration the condition must persist before the machine is considered unhealthy",
												},
											},
										},
									},
									"unhealthy_machine_conditions": schema.SetNestedAttribute{
										Computed:    true,
										Description: "Machine conditions that mark a machine unhealthy (logical OR)",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													Computed:    true,
													Description: "Machine condition type",
												},
												"status": schema.StringAttribute{
													Computed:    true,
													Description: "Required condition status (True/False/Unknown)",
												},
												"timeout_seconds": schema.Int32Attribute{
													Computed:    true,
													Description: "Duration the condition must persist before the machine is considered unhealthy",
												},
											},
										},
									},
								},
							},
							"remediation": schema.SingleNestedAttribute{
								Computed:    true,
								Description: "How unhealthy worker machines are remediated",
								Attributes: map[string]schema.Attribute{
									"trigger_if": schema.SingleNestedAttribute{
										Computed:    true,
										Description: "Conditions under which remediation is triggered",
										Attributes: map[string]schema.Attribute{
											"unhealthy_less_than_or_equal_to": schema.StringAttribute{
												Computed:    true,
												Description: "Remediation triggered only when unhealthy count ≤ this value (int or percentage string, e.g. '3' or '20%')",
											},
											"unhealthy_in_range": schema.StringAttribute{
												Computed:    true,
												Description: "Remediation triggered only when unhealthy count is within this range, e.g. '[3-5]'",
											},
										},
									},
									"template_ref": schema.SingleNestedAttribute{
										Computed:    true,
										Description: "External remediation template (when set, delegates remediation to an external controller)",
										Attributes:  vksClusterClassTemplateReferenceAttrs,
									},
								},
							},
						},
					},
					"naming": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Naming strategy for the control plane object",
						Attributes:  vksClusterClassNamingAttrs,
					},
					"deletion": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Machine deletion configuration for control plane nodes",
						Attributes: map[string]schema.Attribute{
							"node_drain_timeout_seconds": schema.Int32Attribute{
								Computed:    true,
								Description: "Max seconds to spend draining a node before deletion (0 = unlimited)",
							},
							"node_volume_detach_timeout_seconds": schema.Int32Attribute{
								Computed:    true,
								Description: "Max seconds to wait for volumes to detach (0 = unlimited)",
							},
							"node_deletion_timeout_seconds": schema.Int32Attribute{
								Computed:    true,
								Description: "How long to retry node deletion (0 = indefinite, default 10s)",
							},
						},
					},
					"taints": schema.SetNestedAttribute{
						Computed:    true,
						Description: "Node taints managed by Cluster API on control plane nodes",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"key": schema.StringAttribute{
									Computed:    true,
									Description: "Taint key",
								},
								"value": schema.StringAttribute{
									Computed:    true,
									Description: "Taint value",
								},
								"effect": schema.StringAttribute{
									Computed:    true,
									Description: "Taint effect (NoSchedule, PreferNoSchedule, NoExecute)",
								},
								"propagation": schema.StringAttribute{
									Computed:    true,
									Description: "How the taint is propagated to nodes: Always or OnInitialization",
								},
							},
						},
					},
					"readiness_gates": schema.SetNestedAttribute{
						Computed:    true,
						Description: "Additional Machine conditions used to evaluate readiness",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"condition_type": schema.StringAttribute{
									Computed:    true,
									Description: "Machine condition type used as readiness gate",
								},
								"polarity": schema.StringAttribute{
									Computed:    true,
									Description: "Polarity of the condition (Positive or Negative)",
								},
							},
						},
					},
				},
			},
			"workers": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Workers configuration for the ClusterClass",
				Attributes: map[string]schema.Attribute{
					"machine_deployments": schema.SetNestedAttribute{
						Computed:    true,
						Description: "List of machine deployment classes",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"metadata": vksClusterClassObjectMetaAttrs,
								"class": schema.StringAttribute{
									Computed:    true,
									Description: "Unique class name, referenceable from a Cluster topology",
								},
								"bootstrap": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Bootstrap template",
									Attributes: map[string]schema.Attribute{
										"template_ref": schema.SingleNestedAttribute{
											Computed:    true,
											Description: "Reference to the bootstrap template",
											Attributes:  vksClusterClassTemplateReferenceAttrs,
										},
									},
								},
								"infrastructure": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Infrastructure template",
									Attributes: map[string]schema.Attribute{
										"template_ref": schema.SingleNestedAttribute{
											Computed:    true,
											Description: "Reference to the infrastructure template",
											Attributes:  vksClusterClassTemplateReferenceAttrs,
										},
									},
								},
								"health_check": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "MachineHealthCheck configuration for worker machines",
									Attributes: map[string]schema.Attribute{
										"checks": schema.SingleNestedAttribute{
											Computed:    true,
											Description: "Conditions that mark a control plane machine as unhealthy",
											Attributes: map[string]schema.Attribute{
												"node_startup_timeout_seconds": schema.Int32Attribute{
													Computed:    true,
													Description: "Max seconds before a node must have a ProviderID; 0 disables the check (default 600)",
												},
												"unhealthy_node_conditions": schema.SetNestedAttribute{
													Computed:    true,
													Description: "Node conditions that mark a machine unhealthy (logical OR)",
													NestedObject: schema.NestedAttributeObject{
														Attributes: map[string]schema.Attribute{
															"type": schema.StringAttribute{
																Computed:    true,
																Description: "Node condition type",
															},
															"status": schema.StringAttribute{
																Computed:    true,
																Description: "Required condition status (True/False/Unknown)",
															},
															"timeout_seconds": schema.Int32Attribute{
																Computed:    true,
																Description: "Duration the condition must persist before the machine is considered unhealthy",
															},
														},
													},
												},
												"unhealthy_machine_conditions": schema.SetNestedAttribute{
													Computed:    true,
													Description: "Machine conditions that mark a machine unhealthy (logical OR)",
													NestedObject: schema.NestedAttributeObject{
														Attributes: map[string]schema.Attribute{
															"type": schema.StringAttribute{
																Computed:    true,
																Description: "Machine condition type",
															},
															"status": schema.StringAttribute{
																Computed:    true,
																Description: "Required condition status (True/False/Unknown)",
															},
															"timeout_seconds": schema.Int32Attribute{
																Computed:    true,
																Description: "Duration the condition must persist before the machine is considered unhealthy",
															},
														},
													},
												},
											},
										},
										"remediation": schema.SingleNestedAttribute{
											Computed:    true,
											Description: "How unhealthy control plane machines are remediated",
											Attributes: map[string]schema.Attribute{
												"max_in_flight": schema.StringAttribute{
													Computed:    true,
													Description: "Maximum number or percentage of machines that can undergo remediation simultaneously (integer or percentage string, e.g. '3' or '20%')",
												},
												"trigger_if": schema.SingleNestedAttribute{
													Computed:    true,
													Description: "Conditions under which remediation is triggered",
													Attributes: map[string]schema.Attribute{
														"unhealthy_less_than_or_equal_to": schema.StringAttribute{
															Computed:    true,
															Description: "Remediation triggered only when unhealthy count ≤ this value (int or percentage string, e.g. '3' or '20%')",
														},
														"unhealthy_in_range": schema.StringAttribute{
															Computed:    true,
															Description: "Remediation triggered only when unhealthy count is within this range, e.g. '[3-5]'",
														},
													},
												},
												"template_ref": schema.SingleNestedAttribute{
													Computed:    true,
													Description: "External remediation template (when set, delegates remediation to an external controller)",
													Attributes:  vksClusterClassTemplateReferenceAttrs,
												},
											},
										},
									},
								},
								"failure_domain": schema.StringAttribute{
									Computed:    true,
									Description: "Default failure domain for the MachineDeployment",
								},
								"naming": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Naming strategy for the MachineDeployment object",
									Attributes:  vksClusterClassNamingAttrs,
								},
								"deletion": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Machine deletion configuration",
									Attributes: map[string]schema.Attribute{
										"order": schema.StringAttribute{
											Computed:    true,
											Description: "Order in which Machines are deleted when downscaling (Random, Newest, Oldest)",
										},
										"node_drain_timeout_seconds": schema.Int32Attribute{
											Computed:    true,
											Description: "Total time spent draining a node (0 = unlimited)",
										},
										"node_volume_detach_timeout_seconds": schema.Int32Attribute{
											Computed:    true,
											Description: "Total time waiting for all volumes to detach (0 = unlimited)",
										},
										"node_deletion_timeout_seconds": schema.Int32Attribute{
											Computed:    true,
											Description: "How long to attempt deleting the Node after Machine is marked for deletion",
										},
									},
								},
								"taints": schema.SetNestedAttribute{
									Computed:    true,
									Description: "Node taints managed by Cluster API",
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"key": schema.StringAttribute{
												Computed:    true,
												Description: "Taint key",
											},
											"value": schema.StringAttribute{
												Computed:    true,
												Description: "Taint value",
											},
											"effect": schema.StringAttribute{
												Computed:    true,
												Description: "Taint effect (NoSchedule, PreferNoSchedule, NoExecute)",
											},
											"propagation": schema.StringAttribute{
												Computed:    true,
												Description: "How the taint is propagated to nodes: Always or OnInitialization",
											},
										},
									},
								},
								"min_ready_seconds": schema.Int32Attribute{
									Computed:    true,
									Description: "Min seconds a new machine must be ready before being considered available (default 0)",
								},
								"readiness_gates": schema.SetNestedAttribute{
									Computed:    true,
									Description: "Additional Machine conditions used to evaluate readiness",
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"condition_type": schema.StringAttribute{
												Computed:    true,
												Description: "Machine condition type used as a readiness gate",
											},
											"polarity": schema.StringAttribute{
												Computed:    true,
												Description: "Polarity of the condition: Positive (true = ready) or Negative (false = ready)",
											},
										},
									},
								},
								"rollout": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Rollout configuration for the MachineDeployment",
									Attributes: map[string]schema.Attribute{
										"strategy": schema.SingleNestedAttribute{
											Computed:    true,
											Description: "Rollout strategy",
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													Computed:    true,
													Description: "Type of rollout strategy: RollingUpdate or OnDelete",
												},
												"rolling_update": schema.SingleNestedAttribute{
													Computed:    true,
													Description: "Configuration for the RollingUpdate strategy",
													Attributes: map[string]schema.Attribute{
														"max_unavailable": schema.StringAttribute{
															Computed:    true,
															Description: "Maximum number or percentage of unavailable machines during a rolling update",
														},
														"max_surge": schema.StringAttribute{
															Computed:    true,
															Description: "Maximum number or percentage of machines that can be scheduled above the desired number during a rolling update",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"variables": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Variables that can be configured in the Cluster topology and used in patches",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the variable",
						},
						"required": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the variable is required",
						},
						"schema": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "OpenAPI v3 schema for the variable",
							Attributes: map[string]schema.Attribute{
								"open_api_v3_schema": schema.StringAttribute{
									Computed:    true,
									Description: "OpenAPI v3 schema serialised as JSON",
								},
							},
						},
					},
				},
			},
			"patches": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Patches applied to customise referenced templates",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the patch",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Human-readable description of the patch",
						},
						"enabled_if": schema.StringAttribute{
							Computed:    true,
							Description: "Go template expression that enables this patch when it evaluates to 'true'",
						},
						"definitions": schema.SetNestedAttribute{
							Computed:    true,
							Description: "Inline patch definitions applied in order (mutually exclusive with external)",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"selector": schema.SingleNestedAttribute{
										Computed:    true,
										Description: "Selects the templates this patch definition applies to",
										Attributes: map[string]schema.Attribute{
											"api_version": schema.StringAttribute{
												Computed:    true,
												Description: "Filters templates by API version",
											},
											"kind": schema.StringAttribute{
												Computed:    true,
												Description: "Filters templates by kind",
											},
											"match_resources": schema.SingleNestedAttribute{
												Computed:    true,
												Description: "Selects templates based on where they are referenced (results are ORed)",
												Attributes: map[string]schema.Attribute{
													"control_plane": schema.BoolAttribute{
														Computed:    true,
														Description: "Selects templates referenced in spec.controlPlane",
													},
													"infrastructure_cluster": schema.BoolAttribute{
														Computed:    true,
														Description: "Selects templates referenced in spec.infrastructure",
													},
													"machine_deployment_class": schema.SingleNestedAttribute{
														Computed:    true,
														Description: "Selects templates in specific MachineDeploymentClasses",
														Attributes: map[string]schema.Attribute{
															"names": schema.SetAttribute{
																Computed:    true,
																ElementType: types.StringType,
																Description: "Class names to match",
															},
														},
													},
												},
											},
										},
									},
									"json_patches": schema.SetNestedAttribute{
										Computed:    true,
										Description: "JSON patches applied to the matching templates in order",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"op": schema.StringAttribute{
													Computed:    true,
													Description: "JSON patch operation: add, replace, or remove",
												},
												"path": schema.StringAttribute{
													Computed:    true,
													Description: "JSON patch path (must start with /spec/)",
												},
												"value": schema.StringAttribute{
													Computed:    true,
													Description: "Literal value for the patch serialised as JSON (mutually exclusive with value_from)",
												},
												"value_from": schema.SingleNestedAttribute{
													Computed:    true,
													Description: "Dynamic value for the patch (mutually exclusive with value)",
													Attributes: map[string]schema.Attribute{
														"variable": schema.StringAttribute{
															Computed:    true,
															Description: "Variable whose value is used (from spec.variables or builtins)",
														},
														"template": schema.StringAttribute{
															Computed:    true,
															Description: "Go template evaluated to produce the value",
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"external": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "External patch definition delegated to a runtime extension (mutually exclusive with definitions)",
							Attributes: map[string]schema.Attribute{
								"generate_patches_extension": schema.StringAttribute{
									Computed:    true,
									Description: "Name of the runtime extension called to generate patches",
								},
								"validate_topology_extension": schema.StringAttribute{
									Computed:    true,
									Description: "Name of the runtime extension called to validate the topology",
								},
								"discover_variables_extension": schema.StringAttribute{
									Computed:    true,
									Description: "Name of the runtime extension called to discover variables",
								},
								"settings": schema.MapAttribute{
									Computed:    true,
									ElementType: types.StringType,
									Description: "Key-value pairs passed to the extension (override ExtensionConfig settings)",
								},
							},
						},
					},
				},
			},
			"upgrade": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Upgrade configuration for clusters using this ClusterClass",
				Attributes: map[string]schema.Attribute{
					"external": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "External runtime extensions for upgrade operations",
						Attributes: map[string]schema.Attribute{
							"generate_upgrade_plan_extension": schema.StringAttribute{
								Computed:    true,
								Description: "Name of the runtime extension called to generate the upgrade plan",
							},
						},
					},
				},
			},
			"kubernetes_versions": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Ordered list of Kubernetes versions supported by this ClusterClass (oldest to newest)",
			},
			"status": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Observed state of the ClusterClass as reported by the controller",
				Attributes: map[string]schema.Attribute{
					"conditions": kubernetes.ConditionsDataSourceSchema,
					"variables": schema.SetNestedAttribute{
						Computed:    true,
						Description: "Variables as resolved and observed by the controller (includes variables discovered from runtime extensions)",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "Name of the variable",
								},
								"definitions_conflict": schema.BoolAttribute{
									Computed:    true,
									Description: "Whether there are conflicting definitions for this variable name",
								},
								"definitions": schema.SetNestedAttribute{
									Computed:    true,
									Description: "All definitions of this variable",
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"from": schema.StringAttribute{
												Computed:    true,
												Description: "Origin of the definition: 'inline' for variables defined in the ClusterClass, or the patch name for variables discovered via runtime extensions",
											},
											"required": schema.BoolAttribute{
												Computed:    true,
												Description: "Whether the variable is required",
											},
											"schema": schema.SingleNestedAttribute{
												Computed:    true,
												Description: "OpenAPI v3 schema for the variable",
												Attributes: map[string]schema.Attribute{
													"open_api_v3_schema": schema.StringAttribute{
														Computed:    true,
														Description: "OpenAPI v3 schema serialised as JSON",
													},
												},
											},
										},
									},
								},
							},
						},
					},
					"observed_generation": schema.Int64Attribute{
						Computed:    true,
						Description: "Most recent generation observed by the controller",
					},
				},
			},
		},
	}
}
