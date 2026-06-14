// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkscluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/common"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func (d *vcfaVksClusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	vksClusterObjectMetaAttrs := schema.SingleNestedAttribute{
		Computed:    true,
		Description: "Standard Kubernetes object metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
		Attributes: map[string]schema.Attribute{
			"labels": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Labels merged with the corresponding ClusterClass metadata at runtime",
			},
			"annotations": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Annotations merged with the corresponding ClusterClass metadata at runtime",
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Data source for reading a %s", vcfatypes.LabelVksCluster),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Internal identifier of the %s", vcfatypes.LabelVksCluster),
			},

			// Required lookup attributes
			"context": common.VcfContextDataSourceSchema,
			"name": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", vcfatypes.LabelVksCluster),
			},

			// Metadata attributes
			"metadata": kubernetes.MetadataDataSourceSchema,

			// Spec attributes
			"availability_gates": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Additional conditions evaluated when determining the cluster's Available condition",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"condition_type": schema.StringAttribute{
							Computed:    true,
							Description: "Condition type in the Cluster's condition list used as an availability gate",
						},
						"polarity": schema.StringAttribute{
							Computed:    true,
							Description: "Polarity of the condition: Positive (true = healthy) or Negative (false = healthy)",
						},
					},
				},
			},
			"cluster_class": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Reference to the ClusterClass",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "Name of the ClusterClass",
					},
					"namespace": schema.StringAttribute{
						Computed:    true,
						Description: "Namespace of the ClusterClass (defaults to the cluster namespace)",
					},
				},
			},
			"cluster_network": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Cluster-wide network configuration including pod and service CIDR blocks",
				Attributes: map[string]schema.Attribute{
					"service_domain": schema.StringAttribute{
						Computed:    true,
						Description: "Service domain for the cluster (default: cluster.local)",
					},
					"pods": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Pod network CIDR configuration",
						Attributes: map[string]schema.Attribute{
							"cidr_blocks": schema.SetAttribute{
								Computed:    true,
								ElementType: cidrtypes.IPPrefixType{},
								Description: "Set of CIDR blocks allocated for pod IP addresses",
							},
						},
					},
					"services": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Service network CIDR configuration",
						Attributes: map[string]schema.Attribute{
							"cidr_blocks": schema.SetAttribute{
								Computed:    true,
								ElementType: cidrtypes.IPPrefixType{},
								Description: "Set of CIDR blocks allocated for Service VIPs",
							},
						},
					},
				},
			},
			"control_plane": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Topology configuration for the control plane",
				Attributes: map[string]schema.Attribute{
					"metadata": vksClusterObjectMetaAttrs,
					"replicas": schema.Int32Attribute{
						Computed:    true,
						Description: "Desired number of control plane nodes",
					},
					"os_image": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "OS image selection for control plane machines, derived from the run.tanzu.vmware.com/resolve-os-image annotation",
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed:    true,
								Description: "OS image name (e.g. \"ubuntu\")",
							},
							"version": schema.StringAttribute{
								Computed:    true,
								Description: "OS image version (e.g. \"22.04\")",
							},
						},
					},
					"rollout": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Rolling update configuration for the control plane",
						Attributes: map[string]schema.Attribute{
							"after": schema.StringAttribute{
								Computed:    true,
								CustomType:  timetypes.RFC3339Type{},
								Description: "RFC3339 timestamp after which a rollout is triggered even with no spec changes",
							},
						},
					},
					"health_check": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Health check configuration for control plane machines",
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Computed:    true,
								Description: "Whether a MachineHealthCheck should be created for the control plane machines",
							},
							"checks": schema.SingleNestedAttribute{
								Computed:    true,
								Description: "Criteria used to evaluate if a Machine is healthy",
								Attributes: map[string]schema.Attribute{
									"node_startup_timeout_seconds": schema.Int32Attribute{
										Computed:    true,
										Description: "Maximum seconds before a Machine is considered unhealthy if its Node does not appear (0 = disabled, default 10 minutes)",
									},
									"unhealthy_node_conditions": schema.SetNestedAttribute{
										Computed:    true,
										Description: "Node conditions that cause a machine to be considered unhealthy",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													Computed:    true,
													Description: "Node condition type",
												},
												"status": schema.StringAttribute{
													Computed:    true,
													Description: "Condition status (True, False, or Unknown)",
												},
												"timeout_seconds": schema.Int32Attribute{
													Computed:    true,
													Description: "Duration (seconds) the node must be in this state before being deemed unhealthy",
												},
											},
										},
									},
								},
							},
							"remediation": schema.SingleNestedAttribute{
								Computed:    true,
								Description: "Remediation configuration when a Machine is unhealthy",
								Attributes: map[string]schema.Attribute{
									"trigger_if": schema.SingleNestedAttribute{
										Computed:    true,
										Description: "Conditions under which remediation is triggered",
										Attributes: map[string]schema.Attribute{
											"unhealthy_less_than_or_equal_to": schema.StringAttribute{
												Computed:    true,
												Description: "Trigger remediation only when unhealthy machine count is ≤ this value (absolute number or percentage, e.g. '5' or '20%')",
											},
											"unhealthy_in_range": schema.StringAttribute{
												Computed:    true,
												Description: "Trigger remediation only when unhealthy count falls within this range, e.g. '[3-5]'",
											},
										},
									},
								},
							},
						},
					},
					"deletion": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Machine deletion configuration for the control plane",
						Attributes: map[string]schema.Attribute{
							"node_drain_timeout_seconds": schema.Int32Attribute{
								Computed:    true,
								Description: "Maximum seconds to spend draining a node (0 = unlimited)",
							},
							"node_volume_detach_timeout_seconds": schema.Int32Attribute{
								Computed:    true,
								Description: "Maximum seconds to wait for all volumes to detach (0 = unlimited)",
							},
							"node_deletion_timeout_seconds": schema.Int32Attribute{
								Computed:    true,
								Description: "Seconds the controller tries to delete the Node before giving up (0 = retry indefinitely, default 10)",
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
									Description: "Taint effect (NoSchedule, PreferNoSchedule, or NoExecute)",
								},
								"propagation": schema.StringAttribute{
									Computed:    true,
									Description: "Taint propagation (Always or OnInitialization)",
								},
							},
						},
					},
					"readiness_gates": schema.SetNestedAttribute{
						Computed:    true,
						Description: "Additional conditions included when evaluating Machine Ready on control plane nodes",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"condition_type": schema.StringAttribute{
									Computed:    true,
									Description: "Condition type",
								},
								"polarity": schema.StringAttribute{
									Computed:    true,
									Description: "Polarity of the condition: Positive (true = healthy) or Negative (false = healthy)",
								},
							},
						},
					},
					"variable_overrides": schema.SetNestedAttribute{
						Computed:    true,
						Description: "Variable overrides for the control plane",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "Variable name",
								},
								"value": schema.StringAttribute{
									Computed:    true,
									Description: "Variable value serialised as a JSON string",
								},
							},
						},
					},
				},
			},
			"control_plane_endpoint": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The externally reachable API server endpoint for the cluster",
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						Computed:    true,
						Description: "Hostname or IP address of the Kubernetes API server",
					},
					"port": schema.Int32Attribute{
						Computed:    true,
						Description: "TCP port of the Kubernetes API server",
					},
				},
			},
			"machine_deployments": schema.SetNestedAttribute{
				Computed:    true,
				Description: "MachineDeployment topology entries",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"metadata": vksClusterObjectMetaAttrs,
						"class": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the MachineDeploymentClass defined in the ClusterClass",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Unique identifier for this MachineDeployment within the cluster topology",
						},
						"failure_domain": schema.StringAttribute{
							Computed:    true,
							Description: "Failure domain for the machines in this deployment",
						},
						"replicas": schema.Int32Attribute{
							Computed:    true,
							Description: "Desired number of worker nodes in this deployment",
						},
						"os_image": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "OS image selection for this MachineDeployment's machines, derived from the run.tanzu.vmware.com/resolve-os-image annotation",
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "OS image name (e.g. \"ubuntu\")",
								},
								"version": schema.StringAttribute{
									Computed:    true,
									Description: "OS image version (e.g. \"22.04\")",
								},
							},
						},
						"autoscaler": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Cluster Autoscaler bounds for this MachineDeployment, derived from the cluster-api-autoscaler-node-group-min/max-size annotations",
							Attributes: map[string]schema.Attribute{
								"min_size": schema.Int32Attribute{
									Computed:    true,
									Description: "Minimum number of nodes the autoscaler can scale down to",
								},
								"max_size": schema.Int32Attribute{
									Computed:    true,
									Description: "Maximum number of nodes the autoscaler can scale up to",
								},
							},
						},
						"min_ready_seconds": schema.Int32Attribute{
							Computed:    true,
							Description: "Minimum seconds a Machine must be ready before it is considered available (0 = immediate)",
						},
						"health_check": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Health check configuration for the MachineDeployment; overrides ClusterClass settings when set",
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Computed:    true,
									Description: "Whether a MachineHealthCheck should be created for the MachineDeployment machines",
								},
								"checks": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Criteria used to evaluate if a Machine is healthy",
									Attributes: map[string]schema.Attribute{
										"node_startup_timeout_seconds": schema.Int32Attribute{
											Computed:    true,
											Description: "Maximum seconds before a Machine is considered unhealthy if its Node does not appear (0 = disabled, default 10 minutes)",
										},
										"unhealthy_node_conditions": schema.SetNestedAttribute{
											Computed:    true,
											Description: "Node conditions that cause a machine to be considered unhealthy",
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"type": schema.StringAttribute{
														Computed:    true,
														Description: "Node condition type",
													},
													"status": schema.StringAttribute{
														Computed:    true,
														Description: "Condition status (True, False, or Unknown)",
													},
													"timeout_seconds": schema.Int32Attribute{
														Computed:    true,
														Description: "Duration (seconds) the node must be in this state before being deemed unhealthy",
													},
												},
											},
										},
									},
								},
								"remediation": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Remediation configuration when a Machine is unhealthy",
									Attributes: map[string]schema.Attribute{
										"max_in_flight": schema.StringAttribute{
											Computed:    true,
											Description: "Maximum concurrent remediations (absolute number or percentage, e.g. '5' or '20%')",
										},
										"trigger_if": schema.SingleNestedAttribute{
											Computed:    true,
											Description: "Conditions under which remediation is triggered",
											Attributes: map[string]schema.Attribute{
												"unhealthy_less_than_or_equal_to": schema.StringAttribute{
													Computed:    true,
													Description: "Trigger remediation only when unhealthy machine count is ≤ this value (absolute number or percentage, e.g. '5' or '20%')",
												},
												"unhealthy_in_range": schema.StringAttribute{
													Computed:    true,
													Description: "Trigger remediation only when unhealthy count falls within this range, e.g. '[3-5]'",
												},
											},
										},
									},
								},
							},
						},
						"deletion": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Machine deletion configuration for this MachineDeployment",
							Attributes: map[string]schema.Attribute{
								"order": schema.StringAttribute{
									Computed:    true,
									Description: "Order in which Machines are deleted when downscaling: Random, Newest, or Oldest (default: Random)",
								},
								"node_drain_timeout_seconds": schema.Int32Attribute{
									Computed:    true,
									Description: "Maximum seconds to spend draining a node (0 = unlimited)",
								},
								"node_volume_detach_timeout_seconds": schema.Int32Attribute{
									Computed:    true,
									Description: "Maximum seconds to wait for all volumes to detach (0 = unlimited)",
								},
								"node_deletion_timeout_seconds": schema.Int32Attribute{
									Computed:    true,
									Description: "Seconds the controller tries to delete the Node before giving up (0 = retry indefinitely, default 10)",
								},
							},
						},
						"taints": schema.SetNestedAttribute{
							Computed:    true,
							Description: "Node taints managed by Cluster API on this MachineDeployment's nodes",
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
										Description: "Taint effect (NoSchedule, PreferNoSchedule, or NoExecute)",
									},
									"propagation": schema.StringAttribute{
										Computed:    true,
										Description: "Taint propagation (Always or OnInitialization)",
									},
								},
							},
						},
						"readiness_gates": schema.SetNestedAttribute{
							Computed:    true,
							Description: "Additional conditions included when evaluating Machine Ready on this MachineDeployment",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"condition_type": schema.StringAttribute{
										Computed:    true,
										Description: "Condition type",
									},
									"polarity": schema.StringAttribute{
										Computed:    true,
										Description: "Polarity of the condition: Positive (true = healthy) or Negative (false = healthy)",
									},
								},
							},
						},
						"rollout": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Rolling update configuration for this MachineDeployment",
							Attributes: map[string]schema.Attribute{
								"after": schema.StringAttribute{
									Computed:    true,
									CustomType:  timetypes.RFC3339Type{},
									Description: "RFC3339 timestamp after which a rollout is triggered even with no spec changes",
								},
								"strategy": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "Rollout strategy; defaults to RollingUpdate",
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Computed:    true,
											Description: "Strategy type: RollingUpdate or OnDelete",
										},
										"rolling_update": schema.SingleNestedAttribute{
											Computed:    true,
											Description: "Rolling update config; present only when type is RollingUpdate",
											Attributes: map[string]schema.Attribute{
												"max_unavailable": schema.StringAttribute{
													Computed:    true,
													Description: "Maximum unavailable machines during update (absolute number or percentage, e.g. '5' or '10%')",
												},
												"max_surge": schema.StringAttribute{
													Computed:    true,
													Description: "Maximum machines that can be scheduled above the desired count (absolute or percentage)",
												},
											},
										},
									},
								},
							},
						},
						"variable_overrides": schema.SetNestedAttribute{
							Computed:    true,
							Description: "Variable overrides for this MachineDeployment",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:    true,
										Description: "Variable name",
									},
									"value": schema.StringAttribute{
										Computed:    true,
										Description: "Variable value serialised as a JSON string",
									},
								},
							},
						},
					},
				},
			},
			"variables": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Cluster-level variable values passed to ClusterClass patches",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Variable name",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "Variable value serialised as a JSON string",
						},
					},
				},
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Desired Kubernetes Release version for the cluster (e.g. v1.34.1+vmware.1)",
			},

			// Status attributes
			"status": schema.SingleNestedAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Observed state of the %s", vcfatypes.LabelVksCluster),
				Attributes: map[string]schema.Attribute{
					"conditions": kubernetes.ConditionsDataSourceSchema,
					"initialization": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "One-time initialisation milestones",
						Attributes: map[string]schema.Attribute{
							"infrastructure_provisioned": schema.BoolAttribute{
								Computed:    true,
								Description: "True once the infrastructure has been fully provisioned",
							},
							"control_plane_initialized": schema.BoolAttribute{
								Computed:    true,
								Description: "True once the control plane is functional enough to accept requests",
							},
						},
					},
					"control_plane": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Replica counts for the control plane",
						Attributes: map[string]schema.Attribute{
							"desired_replicas": schema.Int32Attribute{
								Computed:    true,
								Description: "Total desired control plane machines",
							},
							"replicas": schema.Int32Attribute{
								Computed:    true,
								Description: "Total control plane machines including those being provisioned or deleted",
							},
							"up_to_date_replicas": schema.Int32Attribute{
								Computed:    true,
								Description: "Control plane machines running the latest spec",
							},
							"ready_replicas": schema.Int32Attribute{
								Computed:    true,
								Description: "Control plane machines in the Ready state",
							},
							"available_replicas": schema.Int32Attribute{
								Computed:    true,
								Description: "Control plane machines that have been ready for at least minReadySeconds",
							},
						},
					},
					"workers": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Aggregate replica counts across all worker MachineDeployments",
						Attributes: map[string]schema.Attribute{
							"desired_replicas": schema.Int32Attribute{
								Computed:    true,
								Description: "Total desired worker machines",
							},
							"replicas": schema.Int32Attribute{
								Computed:    true,
								Description: "Total worker machines including those being provisioned or deleted",
							},
							"up_to_date_replicas": schema.Int32Attribute{
								Computed:    true,
								Description: "Worker machines running the latest spec",
							},
							"ready_replicas": schema.Int32Attribute{
								Computed:    true,
								Description: "Worker machines in the Ready state",
							},
							"available_replicas": schema.Int32Attribute{
								Computed:    true,
								Description: "Worker machines that have been ready for at least minReadySeconds",
							},
						},
					},
					"failure_domains": schema.SetNestedAttribute{
						Computed:    true,
						Description: "Failure domains discovered from the infrastructure provider and available for scheduling",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "Name of the failure domain",
								},
								"control_plane": schema.BoolAttribute{
									Computed:    true,
									Description: "Whether this failure domain is suitable for control plane machines",
								},
								"attributes": schema.MapAttribute{
									Computed:    true,
									ElementType: types.StringType,
									Description: "Map of free-form key-value attributes provided by the infrastructure provider",
								},
							},
						},
					},
					"phase": schema.StringAttribute{
						Computed:    true,
						Description: "Current lifecycle phase of the cluster: Pending, Provisioning, Provisioned, Deleting, Failed, or Unknown",
					},
					"observed_generation": schema.Int64Attribute{
						Computed:    true,
						Description: "Most recent generation of the Cluster spec observed by the controller",
					},
				},
			},
		},
	}
}
