// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkscluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/common"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/validators"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func (r *vcfaVksClusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	vksClusterObjectMetaAttrs := map[string]schema.Attribute{
		"labels": schema.MapAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "Labels merged with the corresponding ClusterClass metadata at runtime",
			Validators: []validator.Map{
				mapvalidator.SizeAtLeast(1),
			},
		},
		"annotations": schema.MapAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "Annotations merged with the corresponding ClusterClass metadata at runtime",
			Validators: []validator.Map{
				mapvalidator.SizeAtLeast(1),
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Resource for managing a %s.", vcfatypes.LabelVksCluster),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Internal identifier of the %s", vcfatypes.LabelVksCluster),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Required attributes
			"context": common.VcfContextResourceSchema,
			"name": schema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Name of the %s (must be RFC 1123 DNS subdomain compliant)", vcfatypes.LabelVksCluster),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Validation attributes
			"dry_run_validation": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When true, a dry-run Create or Update is sent to the backend during `terraform plan` and `terraform apply` to validate the cluster configuration before committing any changes. Backend validation errors are surfaced as plan errors. Defaults to false.",
			},

			// Wait attributes
			"wait_for": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Controls whether certain operations block until the cluster reaches a certain state",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"available": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "When true, Create and Update operations blocks until the cluster's Available condition is True. Set to false (default) to return immediately after the API call.",
					},
					"deleted": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "When true, Delete operation blocks until the cluster is fully removed. Set to false (default) to return immediately after the delete API call.",
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),

			// Metadata attributes
			"metadata": kubernetes.MetadataResourceSchema,

			// User-managed cluster metadata. Only the keys explicitly set here are tracked in
			// Terraform state; any additional labels/annotations injected by the backend are
			// silently ignored and will never appear in the plan diff.
			"labels": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "User-managed labels to set on the cluster's ObjectMeta. Keys not present here are not tracked, so backend-injected labels are never shown as a diff.",
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
				},
			},
			"annotations": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "User-managed annotations to set on the cluster's ObjectMeta. Keys not present here are not tracked, so backend-injected annotations are never shown as a diff.",
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
				},
			},

			// Spec attributes
			"availability_gates": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Additional conditions evaluated when determining the Cluster Available condition.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
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
				Required:    true,
				Description: "Reference to the ClusterClass",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Required:    true,
						Description: "Name of the ClusterClass (1–253 characters; DNS subdomain format)",
						Validators: []validator.String{
							stringvalidator.LengthBetween(1, 253),
							stringvalidator.RegexMatches(kubernetes.ReDNSSubdomain, "must be a valid DNS subdomain"),
						},
					},
					"namespace": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Namespace of the ClusterClass (1–63 characters; DNS label format)",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.LengthBetween(1, 63),
							stringvalidator.RegexMatches(kubernetes.ReDNSLabel, "must be a valid DNS label"),
						},
					},
				},
			},
			"cluster_network": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Cluster-wide network configuration including pod and service CIDR blocks. Immutable after creation.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"service_domain": schema.StringAttribute{
						Optional:    true,
						Description: "Service domain for the cluster (default: cluster.local) (1–253 characters)",
						Validators: []validator.String{
							stringvalidator.LengthBetween(1, 253),
						},
					},
					"pods": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Pod network CIDR configuration",
						Attributes: map[string]schema.Attribute{
							"cidr_blocks": schema.SetAttribute{
								Required:    true,
								ElementType: cidrtypes.IPPrefixType{},
								Description: "Set of CIDR blocks allocated for pod IP addresses (1–100 entries; e.g. \"192.168.0.0/16\")",
								Validators: []validator.Set{
									setvalidator.SizeBetween(1, 100),
									setvalidator.ValueStringsAre(validators.IsValidCIDR()),
								},
							},
						},
					},
					"services": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Service network CIDR configuration",
						Attributes: map[string]schema.Attribute{
							"cidr_blocks": schema.SetAttribute{
								Required:    true,
								ElementType: cidrtypes.IPPrefixType{},
								Description: "Set of CIDR blocks allocated for Service VIPs (1–100 entries; e.g. \"10.96.0.0/12\")",
								Validators: []validator.Set{
									setvalidator.SizeBetween(1, 100),
									setvalidator.ValueStringsAre(validators.IsValidCIDR()),
								},
							},
						},
					},
				},
			},
			"control_plane": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Topology configuration for the control plane",
				Validators: []validator.Object{
					validators.VksOsImageAnnotationConflict(),
				},
				Attributes: map[string]schema.Attribute{
					"metadata": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Metadata merged with the ClusterClass control plane metadata at runtime",
						Validators: []validator.Object{
							validators.ObjectNotEmpty(),
						},
						Attributes: vksClusterObjectMetaAttrs,
					},
					"replicas": schema.Int32Attribute{
						Required:    true,
						Description: "Desired number of control plane nodes (must be 1, 3, or 5)",
						Validators: []validator.Int32{
							int32validator.OneOf(1, 3, 5),
						},
					},
					"rollout": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Rolling update configuration for the control plane",
						Attributes: map[string]schema.Attribute{
							"after": schema.StringAttribute{
								Required:    true,
								CustomType:  timetypes.RFC3339Type{},
								Description: "RFC3339 timestamp after which a rollout is triggered even with no spec changes",
							},
						},
					},
					"health_check": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Health check configuration for control plane machines",
						Validators: []validator.Object{
							validators.ObjectNotEmpty(),
						},
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Optional:    true,
								Description: "Whether a MachineHealthCheck should be created for the control plane machines",
							},
							"checks": schema.SingleNestedAttribute{
								Optional:    true,
								Description: "Criteria used to evaluate if a Machine is healthy",
								Validators: []validator.Object{
									validators.ObjectNotEmpty(),
								},
								Attributes: map[string]schema.Attribute{
									"node_startup_timeout_seconds": schema.Int32Attribute{
										Optional:    true,
										Description: "Maximum seconds before a Machine is considered unhealthy if its Node does not appear (0 = disabled, default 10 minutes). When non-zero the value must be at least 30.",
										Validators: []validator.Int32{
											int32validator.Any(
												int32validator.OneOf(0),
												int32validator.AtLeast(30),
											),
										},
									},
									"unhealthy_node_conditions": schema.SetNestedAttribute{
										Optional:    true,
										Description: "Node conditions that cause a machine to be considered unhealthy (1–100 entries when specified; logical OR)",
										Validators: []validator.Set{
											setvalidator.SizeBetween(1, 100),
										},
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													Required:    true,
													Description: "Node condition type",
												},
												"status": schema.StringAttribute{
													Required:    true,
													Description: "Condition status (True, False, or Unknown)",
												},
												"timeout_seconds": schema.Int32Attribute{
													Required:    true,
													Description: "Duration (seconds) the node must be in this state before being deemed unhealthy",
													Validators: []validator.Int32{
														int32validator.AtLeast(0),
													},
												},
											},
										},
									},
								},
							},
							"remediation": schema.SingleNestedAttribute{
								Optional:    true,
								Description: "Remediation configuration when a Machine is unhealthy",
								Validators: []validator.Object{
									validators.ObjectNotEmpty(),
								},
								Attributes: map[string]schema.Attribute{
									"trigger_if": schema.SingleNestedAttribute{
										Optional:    true,
										Description: "Conditions under which remediation is triggered",
										Validators: []validator.Object{
											validators.ObjectNotEmpty(),
										},
										Attributes: map[string]schema.Attribute{
											"unhealthy_less_than_or_equal_to": schema.StringAttribute{
												Optional:    true,
												Description: "Trigger remediation only when unhealthy machine count is ≤ this value (absolute number or percentage, e.g. '5' or '20%')",
											},
											"unhealthy_in_range": schema.StringAttribute{
												Optional:    true,
												Description: "Trigger remediation only when unhealthy count falls within this range, e.g. '[3-5]' (1–32 characters)",
												Validators: []validator.String{
													stringvalidator.LengthBetween(1, 32),
													stringvalidator.RegexMatches(kubernetes.ReUnhealthyInRange, "must match the pattern [min-max], e.g. [3-5]"),
												},
											},
										},
									},
								},
							},
						},
					},
					"deletion": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Machine deletion configuration for the control plane",
						Validators: []validator.Object{
							validators.ObjectNotEmpty(),
						},
						Attributes: map[string]schema.Attribute{
							"node_drain_timeout_seconds": schema.Int32Attribute{
								Optional:    true,
								Description: "Maximum seconds to spend draining a node (0 = unlimited)",
								Validators: []validator.Int32{
									int32validator.AtLeast(0),
								},
							},
							"node_volume_detach_timeout_seconds": schema.Int32Attribute{
								Optional:    true,
								Description: "Maximum seconds to wait for all volumes to detach (0 = unlimited)",
								Validators: []validator.Int32{
									int32validator.AtLeast(0),
								},
							},
							"node_deletion_timeout_seconds": schema.Int32Attribute{
								Optional:    true,
								Description: "Seconds the controller tries to delete the Node before giving up (0 = retry indefinitely, default 10)",
								Validators: []validator.Int32{
									int32validator.AtLeast(0),
								},
							},
						},
					},
					"taints": schema.SetNestedAttribute{
						Computed:    true,
						Description: "Node taints on control plane nodes.",
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
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
						Description: "Additional conditions included when evaluating Machine Ready on control plane nodes.",
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"condition_type": schema.StringAttribute{
									Computed:    true,
									Description: "Condition type",
								},
								"polarity": schema.StringAttribute{
									Computed:    true,
									Description: "Condition polarity (Positive or Negative)",
								},
							},
						},
					},
					"variable_overrides": schema.SetNestedAttribute{
						Optional:    true,
						Description: "Variable overrides for the control plane (1–1000 entries when specified)",
						Validators: []validator.Set{
							setvalidator.SizeBetween(1, 1000),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:    true,
									Description: "Variable name (1–256 characters)",
									Validators: []validator.String{
										stringvalidator.LengthBetween(1, 256),
									},
								},
								"value": schema.StringAttribute{
									Required:    true,
									Description: "Variable value serialised as a JSON string",
								},
							},
						},
					},
					"os_image": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "OS image selection for control plane machines. When set, injects the annotation \"run.tanzu.vmware.com/resolve-os-image\" into the control plane metadata. Conflicts with specifying that annotation directly in \"metadata.annotations\".",
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Required:    true,
								Description: "OS image name (e.g. \"ubuntu\")",
							},
							"version": schema.StringAttribute{
								Optional:    true,
								Description: "OS image version (e.g. \"22.04\")",
							},
						},
					},
				},
			},
			"control_plane_endpoint": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The externally reachable API server endpoint for the cluster.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
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
				Optional:    true,
				Computed:    true,
				Description: "Set of MachineDeployment topology entries (1–2000 entries when specified)",
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 2000),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						validators.VksMachineDeploymentHasScaling(),
						validators.VksOsImageAnnotationConflict(),
						validators.VksAutoscalerAnnotationConflict(),
					},
					Attributes: map[string]schema.Attribute{
						"metadata": schema.SingleNestedAttribute{
							Optional:    true,
							Description: "Metadata merged with the ClusterClass MachineDeployment metadata at runtime",
							Validators: []validator.Object{
								validators.ObjectNotEmpty(),
							},
							Attributes: vksClusterObjectMetaAttrs,
						},
						"class": schema.StringAttribute{
							Required:    true,
							Description: "Name of the MachineDeploymentClass defined in the ClusterClass (1–256 characters)",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 256),
							},
						},
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Unique identifier for this MachineDeployment within the cluster topology (1–63 characters; DNS subdomain format)",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 63),
								stringvalidator.RegexMatches(kubernetes.ReDNSSubdomain, "must be a valid DNS subdomain"),
							},
						},
						"failure_domain": schema.StringAttribute{
							Optional:    true,
							Description: "Failure domain for the machines in this deployment (1–256 characters)",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 256),
							},
						},
						"replicas": schema.Int32Attribute{
							Optional:    true,
							Description: "Desired number of worker nodes in this deployment. Mutually exclusive with \"autoscaler\".",
						},
						"autoscaler": schema.SingleNestedAttribute{
							Optional:    true,
							Description: "Cluster Autoscaler configuration for this MachineDeployment. When set, injects the \"cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size\" and \"cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size\" annotations into the MachineDeployment metadata. Conflicts with specifying those annotations directly in \"metadata.annotations\". Mutually exclusive with \"replicas\".",
							Attributes: map[string]schema.Attribute{
								"min_size": schema.Int32Attribute{
									Optional:    true,
									Description: "Minimum number of nodes the autoscaler can scale down to",
								},
								"max_size": schema.Int32Attribute{
									Optional:    true,
									Description: "Maximum number of nodes the autoscaler can scale up to",
								},
							},
						},
						"min_ready_seconds": schema.Int32Attribute{
							Optional:    true,
							Description: "Minimum seconds a Machine must be ready before it is considered available (0 = immediate, Minimum=0)",
							Validators: []validator.Int32{
								int32validator.AtLeast(0),
							},
						},
						"health_check": schema.SingleNestedAttribute{
							Optional:    true,
							Description: "Health check configuration for the MachineDeployment; overrides ClusterClass settings when set",
							Validators: []validator.Object{
								validators.ObjectNotEmpty(),
							},
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Optional:    true,
									Description: "Whether a MachineHealthCheck should be created for the MachineDeployment machines",
								},
								"checks": schema.SingleNestedAttribute{
									Optional:    true,
									Description: "Criteria used to evaluate if a Machine is healthy",
									Validators: []validator.Object{
										validators.ObjectNotEmpty(),
									},
									Attributes: map[string]schema.Attribute{
										"node_startup_timeout_seconds": schema.Int32Attribute{
											Optional:    true,
											Description: "Maximum seconds before a Machine is considered unhealthy if its Node does not appear (0 = disabled, default 10 minutes). When non-zero the value must be at least 30.",
											Validators: []validator.Int32{
												int32validator.Any(
													int32validator.OneOf(0),
													int32validator.AtLeast(30),
												),
											},
										},
										"unhealthy_node_conditions": schema.SetNestedAttribute{
											Optional:    true,
											Description: "Node conditions that cause a machine to be considered unhealthy (1–100 entries when specified; logical OR)",
											Validators: []validator.Set{
												setvalidator.SizeBetween(1, 100),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"type": schema.StringAttribute{
														Required:    true,
														Description: "Node condition type",
													},
													"status": schema.StringAttribute{
														Required:    true,
														Description: "Condition status (True, False, or Unknown)",
													},
													"timeout_seconds": schema.Int32Attribute{
														Required:    true,
														Description: "Duration (seconds) the node must be in this state before being deemed unhealthy",
														Validators: []validator.Int32{
															int32validator.AtLeast(0),
														},
													},
												},
											},
										},
									},
								},
								"remediation": schema.SingleNestedAttribute{
									Optional:    true,
									Description: "Remediation configuration when a Machine is unhealthy",
									Validators: []validator.Object{
										validators.ObjectNotEmpty(),
									},
									Attributes: map[string]schema.Attribute{
										"max_in_flight": schema.StringAttribute{
											Optional:    true,
											Description: "Maximum concurrent remediations (absolute number or percentage, e.g. '5' or '20%')",
										},
										"trigger_if": schema.SingleNestedAttribute{
											Optional:    true,
											Description: "Conditions under which remediation is triggered",
											Validators: []validator.Object{
												validators.ObjectNotEmpty(),
											},
											Attributes: map[string]schema.Attribute{
												"unhealthy_less_than_or_equal_to": schema.StringAttribute{
													Optional:    true,
													Description: "Trigger remediation only when unhealthy machine count is ≤ this value (absolute number or percentage, e.g. '5' or '20%')",
												},
												"unhealthy_in_range": schema.StringAttribute{
													Optional:    true,
													Description: "Trigger remediation only when unhealthy count falls within this range, e.g. '[3-5]' (1–32 characters)",
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 32),
														stringvalidator.RegexMatches(kubernetes.ReUnhealthyInRange, "must match the pattern [min-max], e.g. [3-5]"),
													},
												},
											},
										},
									},
								},
							},
						},
						"deletion": schema.SingleNestedAttribute{
							Optional:    true,
							Description: "Machine deletion configuration for this MachineDeployment",
							Validators: []validator.Object{
								validators.ObjectNotEmpty(),
							},
							Attributes: map[string]schema.Attribute{
								"order": schema.StringAttribute{
									Optional:    true,
									Description: "Order in which Machines are deleted when downscaling: Random, Newest, or Oldest (default: Random)",
									Validators: []validator.String{
										stringvalidator.OneOf("Random", "Newest", "Oldest"),
									},
								},
								"node_drain_timeout_seconds": schema.Int32Attribute{
									Optional:    true,
									Description: "Maximum seconds to spend draining a node (0 = unlimited)",
									Validators: []validator.Int32{
										int32validator.AtLeast(0),
									},
								},
								"node_volume_detach_timeout_seconds": schema.Int32Attribute{
									Optional:    true,
									Description: "Maximum seconds to wait for all volumes to detach (0 = unlimited)",
									Validators: []validator.Int32{
										int32validator.AtLeast(0),
									},
								},
								"node_deletion_timeout_seconds": schema.Int32Attribute{
									Optional:    true,
									Description: "Seconds the controller tries to delete the Node before giving up (0 = retry indefinitely, default 10)",
									Validators: []validator.Int32{
										int32validator.AtLeast(0),
									},
								},
							},
						},
						"taints": schema.SetNestedAttribute{
							Computed:    true,
							Description: "Node taints on this MachineDeployment's nodes.",
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
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
							Description: "Additional conditions included when evaluating Machine Ready on this MachineDeployment.",
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
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
							Optional:    true,
							Description: "Rolling update configuration for this MachineDeployment",
							Attributes: map[string]schema.Attribute{
								"after": schema.StringAttribute{
									Required:    true,
									CustomType:  timetypes.RFC3339Type{},
									Description: "RFC3339 timestamp after which a rollout is triggered even with no spec changes",
								},
								"strategy": schema.SingleNestedAttribute{
									Optional:    true,
									Description: "Rollout strategy; defaults to RollingUpdate",
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Required:    true,
											Description: "Strategy type: RollingUpdate or OnDelete",
											Validators: []validator.String{
												stringvalidator.OneOf("RollingUpdate", "OnDelete"),
											},
										},
										"rolling_update": schema.SingleNestedAttribute{
											Optional:    true,
											Description: "Rolling update config; present only when type is RollingUpdate",
											Validators: []validator.Object{
												validators.ObjectNotEmpty(),
											},
											Attributes: map[string]schema.Attribute{
												"max_unavailable": schema.StringAttribute{
													Optional:    true,
													Description: "Maximum unavailable machines during update (absolute number or percentage, e.g. '5' or '10%')",
												},
												"max_surge": schema.StringAttribute{
													Optional:    true,
													Description: "Maximum machines that can be scheduled above the desired count (absolute or percentage)",
												},
											},
										},
									},
								},
							},
						},
						"variable_overrides": schema.SetNestedAttribute{
							Optional:    true,
							Description: "Variable overrides for this MachineDeployment (1–1000 entries when specified)",
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 1000),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required:    true,
										Description: "Variable name (1–256 characters)",
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 256),
										},
									},
									"value": schema.StringAttribute{
										Required:    true,
										Description: "Variable value serialised as a JSON string",
									},
								},
							},
						},
						"os_image": schema.SingleNestedAttribute{
							Optional:    true,
							Description: "OS image selection for this MachineDeployment's machines. When set, injects the annotation \"run.tanzu.vmware.com/resolve-os-image\" into the MachineDeployment metadata. Conflicts with specifying that annotation directly in \"metadata.annotations\".",
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:    true,
									Description: "OS image name (e.g. \"ubuntu\")",
								},
								"version": schema.StringAttribute{
									Optional:    true,
									Description: "OS image version (e.g. \"22.04\")",
								},
							},
						},
					},
				},
			},
			"variables": schema.SetNestedAttribute{
				Required:    true,
				Description: "Cluster-level variable values passed to ClusterClass patches (1–1000 entries when specified)",
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 1000),
					validators.VksClusterVariablesHaveRequiredNames("vmClass", "storageClass"),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Variable name (1–256 characters)",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 256),
							},
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "Variable value serialised as a JSON string"},
					},
				},
			},
			"version": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Desired Kubernetes Release version for the cluster (e.g. v1.34.1+vmware.1)",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},

			// Status attributes
			"status": schema.SingleNestedAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Observed state of the %s", vcfatypes.LabelVksCluster),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"conditions": kubernetes.ConditionsResourceSchema,
					"initialization": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "One-time initialisation milestones",
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
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
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
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
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
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
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
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
