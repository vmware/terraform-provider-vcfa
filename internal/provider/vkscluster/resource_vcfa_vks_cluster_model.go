// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkscluster

import (
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
)

// ── Resource Top-level model ─────────────────────────────────────────────────

type vcfaVksClusterResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Context types.Object `tfsdk:"context"`
	Name    types.String `tfsdk:"name"`

	// Validation controls
	DryRunValidation types.Bool `tfsdk:"dry_run_validation"`

	// Wait controls
	WaitFor types.Object `tfsdk:"wait_for"`

	// Timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`

	// Metadata
	Metadata types.Object `tfsdk:"metadata"`

	// User-managed labels and annotations on the cluster's ObjectMeta.
	// Only the keys the user specifies are tracked; backend-injected entries are ignored.
	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`

	// Spec fields
	AvailabilityGates    types.Set    `tfsdk:"availability_gates"`
	ClusterClass         types.Object `tfsdk:"cluster_class"`
	ClusterNetwork       types.Object `tfsdk:"cluster_network"`
	ControlPlane         types.Object `tfsdk:"control_plane"`
	ControlPlaneEndpoint types.Object `tfsdk:"control_plane_endpoint"`
	MachineDeployments   types.Set    `tfsdk:"machine_deployments"`
	Variables            types.Set    `tfsdk:"variables"`
	Version              types.String `tfsdk:"version"`

	// Status
	Status types.Object `tfsdk:"status"`
}

// ── Wait controls ─-----────-----─────────────────────────────────────────────

type vksClusterWaitForModel struct {
	Available types.Bool `tfsdk:"available"`
	Deleted   types.Bool `tfsdk:"deleted"`
}

var vksClusterWaitForAttrTypes = map[string]attr.Type{
	"available": types.BoolType,
	"deleted":   types.BoolType,
}

// ── Availability gates ───────────────────────────────────────────────────────

type vksClusterAvailabilityGateModel struct {
	ConditionType types.String `tfsdk:"condition_type"`
	Polarity      types.String `tfsdk:"polarity"`
}

var vksClusterAvailabilityGateAttrTypes = map[string]attr.Type{
	"condition_type": types.StringType,
	"polarity":       types.StringType,
}

// ── Cluster class reference ──────────────────────────────────────────────────

type vksClusterClassRefModel struct {
	Name      types.String `tfsdk:"name"`
	Namespace types.String `tfsdk:"namespace"`
}

var vksClusterClassRefAttrTypes = map[string]attr.Type{
	"name":      types.StringType,
	"namespace": types.StringType,
}

// ── Cluster network ──────────────────────────────────────────────────────────

type vksClusterNetworkModel struct {
	ServiceDomain types.String `tfsdk:"service_domain"`
	Pods          types.Object `tfsdk:"pods"`
	Services      types.Object `tfsdk:"services"`
}

var vksClusterNetworkAttrTypes = map[string]attr.Type{
	"service_domain": types.StringType,
	"pods": types.ObjectType{
		AttrTypes: vksClusterNetworkRangesAttrTypes,
	},
	"services": types.ObjectType{
		AttrTypes: vksClusterNetworkRangesAttrTypes,
	},
}

type vksClusterNetworkRangesModel struct {
	CIDRBlocks types.Set `tfsdk:"cidr_blocks"`
}

var vksClusterNetworkRangesAttrTypes = map[string]attr.Type{
	"cidr_blocks": types.SetType{
		ElemType: cidrtypes.IPPrefixType{},
	},
}

// ── Control plane topology ───────────────────────────────────────────────────

type vksClusterControlPlaneTopologyModel struct {
	Metadata          types.Object `tfsdk:"metadata"`
	OsImage           types.Object `tfsdk:"os_image"`
	Replicas          types.Int32  `tfsdk:"replicas"`
	Rollout           types.Object `tfsdk:"rollout"`
	HealthCheck       types.Object `tfsdk:"health_check"`
	Deletion          types.Object `tfsdk:"deletion"`
	Taints            types.Set    `tfsdk:"taints"`
	ReadinessGates    types.Set    `tfsdk:"readiness_gates"`
	VariableOverrides types.Set    `tfsdk:"variable_overrides"`
}

var vksClusterControlPlaneTopologyAttrTypes = map[string]attr.Type{
	"metadata": types.ObjectType{
		AttrTypes: vksClusterObjectMetaAttrTypes,
	},
	"replicas": types.Int32Type,
	"rollout": types.ObjectType{
		AttrTypes: vksClusterControlPlaneTopologyRolloutAttrTypes,
	},
	"health_check": types.ObjectType{
		AttrTypes: vksClusterControlPlaneTopologyHealthCheckAttrTypes,
	},
	"deletion": types.ObjectType{
		AttrTypes: vksClusterControlPlaneTopologyMachineDeletionAttrTypes,
	},
	"taints": types.SetType{ElemType: types.ObjectType{
		AttrTypes: vksClusterMachineTaintAttrTypes,
	}},
	"readiness_gates": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: vksClusterMachineReadinessGateAttrTypes,
		},
	},
	"variable_overrides": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: vksClusterVariableAttrTypes,
		},
	},
	"os_image": types.ObjectType{
		AttrTypes: vksClusterOsImageAttrTypes,
	},
}

type vksClusterControlPlaneTopologyRolloutModel struct {
	After timetypes.RFC3339 `tfsdk:"after"`
}

var vksClusterControlPlaneTopologyRolloutAttrTypes = map[string]attr.Type{
	"after": timetypes.RFC3339Type{},
}

type vksClusterControlPlaneTopologyHealthCheckModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Checks      types.Object `tfsdk:"checks"`
	Remediation types.Object `tfsdk:"remediation"`
}

var vksClusterControlPlaneTopologyHealthCheckAttrTypes = map[string]attr.Type{
	"enabled": types.BoolType,
	"checks": types.ObjectType{
		AttrTypes: vksClusterControlPlaneTopologyHealthCheckChecksAttrTypes,
	},
	"remediation": types.ObjectType{
		AttrTypes: vksClusterControlPlaneTopologyHealthCheckRemediationAttrTypes,
	},
}

type vksClusterControlPlaneTopologyHealthCheckChecksModel struct {
	NodeStartupTimeoutSeconds types.Int32 `tfsdk:"node_startup_timeout_seconds"`
	UnhealthyNodeConditions   types.Set   `tfsdk:"unhealthy_node_conditions"`
}

var vksClusterControlPlaneTopologyHealthCheckChecksAttrTypes = map[string]attr.Type{
	"node_startup_timeout_seconds": types.Int32Type,
	"unhealthy_node_conditions": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: vksClusterUnhealthyNodeConditionAttrTypes,
		},
	},
}

type vksClusterControlPlaneTopologyHealthCheckRemediationModel struct {
	TriggerIf types.Object `tfsdk:"trigger_if"`
}

var vksClusterControlPlaneTopologyHealthCheckRemediationAttrTypes = map[string]attr.Type{
	"trigger_if": types.ObjectType{
		AttrTypes: vksClusterControlPlaneTopologyHealthCheckRemediationTriggerIfAttrTypes,
	},
}

type vksClusterControlPlaneTopologyHealthCheckRemediationTriggerIfModel struct {
	UnhealthyLessThanOrEqualTo types.String `tfsdk:"unhealthy_less_than_or_equal_to"`
	UnhealthyInRange           types.String `tfsdk:"unhealthy_in_range"`
}

var vksClusterControlPlaneTopologyHealthCheckRemediationTriggerIfAttrTypes = map[string]attr.Type{
	"unhealthy_less_than_or_equal_to": types.StringType,
	"unhealthy_in_range":              types.StringType,
}

type vksClusterControlPlaneTopologyMachineDeletionModel struct {
	NodeDrainTimeoutSeconds        types.Int32 `tfsdk:"node_drain_timeout_seconds"`
	NodeVolumeDetachTimeoutSeconds types.Int32 `tfsdk:"node_volume_detach_timeout_seconds"`
	NodeDeletionTimeoutSeconds     types.Int32 `tfsdk:"node_deletion_timeout_seconds"`
}

var vksClusterControlPlaneTopologyMachineDeletionAttrTypes = map[string]attr.Type{
	"node_drain_timeout_seconds":         types.Int32Type,
	"node_volume_detach_timeout_seconds": types.Int32Type,
	"node_deletion_timeout_seconds":      types.Int32Type,
}

// ── Control Plane endpoint ───────────────────────────────────────────────────

type vksClusterApiEndpointModel struct {
	Host types.String `tfsdk:"host"`
	Port types.Int32  `tfsdk:"port"`
}

var vksClusterApiEndpointAttrTypes = map[string]attr.Type{
	"host": types.StringType,
	"port": types.Int32Type,
}

// ── Machine deployment topology ──────────────────────────────────────────────

type vksClusterMachineDeploymentTopologyModel struct {
	Metadata          types.Object `tfsdk:"metadata"`
	Class             types.String `tfsdk:"class"`
	Name              types.String `tfsdk:"name"`
	FailureDomain     types.String `tfsdk:"failure_domain"`
	Replicas          types.Int32  `tfsdk:"replicas"`
	Autoscaler        types.Object `tfsdk:"autoscaler"`
	HealthCheck       types.Object `tfsdk:"health_check"`
	Deletion          types.Object `tfsdk:"deletion"`
	Taints            types.Set    `tfsdk:"taints"`
	MinReadySeconds   types.Int32  `tfsdk:"min_ready_seconds"`
	ReadinessGates    types.Set    `tfsdk:"readiness_gates"`
	Rollout           types.Object `tfsdk:"rollout"`
	VariableOverrides types.Set    `tfsdk:"variable_overrides"`
	OsImage           types.Object `tfsdk:"os_image"`
}

var vksMachineDeploymentTopologyAttrTypes = map[string]attr.Type{
	"metadata": types.ObjectType{
		AttrTypes: vksClusterObjectMetaAttrTypes,
	},
	"class":          types.StringType,
	"name":           types.StringType,
	"failure_domain": types.StringType,
	"replicas":       types.Int32Type,
	"autoscaler": types.ObjectType{
		AttrTypes: vksClusterAutoscalerAttrTypes,
	},
	"health_check": types.ObjectType{
		AttrTypes: vksClusterMachineDeploymentTopologyHealthCheckAttrTypes,
	},
	"deletion": types.ObjectType{
		AttrTypes: vksClusterMachineDeploymentTopologyDeletionAttrTypes,
	},
	"taints": types.SetType{ElemType: types.ObjectType{
		AttrTypes: vksClusterMachineTaintAttrTypes,
	}},
	"min_ready_seconds": types.Int32Type,
	"readiness_gates": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: vksClusterMachineReadinessGateAttrTypes,
		},
	},
	"rollout": types.ObjectType{
		AttrTypes: vksClusterMachineDeploymentTopologyRolloutAttrTypes,
	},
	"variable_overrides": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: vksClusterVariableAttrTypes,
		},
	},
	"os_image": types.ObjectType{
		AttrTypes: vksClusterOsImageAttrTypes,
	},
}

type vksClusterMachineDeploymentTopologyHealthCheckModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Checks      types.Object `tfsdk:"checks"`
	Remediation types.Object `tfsdk:"remediation"`
}

var vksClusterMachineDeploymentTopologyHealthCheckAttrTypes = map[string]attr.Type{
	"enabled": types.BoolType,
	"checks": types.ObjectType{
		AttrTypes: vksClusterMachineDeploymentTopologyHealthCheckChecksAttrTypes,
	},
	"remediation": types.ObjectType{
		AttrTypes: vksClusterMachineDeploymentTopologyHealthCheckRemediationAttrTypes,
	},
}

type vksClusterMachineDeploymentTopologyHealthCheckChecksModel struct {
	NodeStartupTimeoutSeconds types.Int32 `tfsdk:"node_startup_timeout_seconds"`
	UnhealthyNodeConditions   types.Set   `tfsdk:"unhealthy_node_conditions"`
}

var vksClusterMachineDeploymentTopologyHealthCheckChecksAttrTypes = map[string]attr.Type{
	"node_startup_timeout_seconds": types.Int32Type,
	"unhealthy_node_conditions": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: vksClusterUnhealthyNodeConditionAttrTypes,
		},
	},
}

type vksClusterMachineDeploymentTopologyHealthCheckRemediationModel struct {
	MaxInFlight types.String `tfsdk:"max_in_flight"`
	TriggerIf   types.Object `tfsdk:"trigger_if"`
}

var vksClusterMachineDeploymentTopologyHealthCheckRemediationAttrTypes = map[string]attr.Type{
	"max_in_flight": types.StringType,
	"trigger_if": types.ObjectType{
		AttrTypes: vksClusterMachineDeploymentTopologyHealthCheckRemediationTriggerIfAttrTypes,
	},
}

type vksClusterMachineDeploymentTopologyHealthCheckRemediationTriggerIfModel struct {
	UnhealthyLessThanOrEqualTo types.String `tfsdk:"unhealthy_less_than_or_equal_to"`
	UnhealthyInRange           types.String `tfsdk:"unhealthy_in_range"`
}

var vksClusterMachineDeploymentTopologyHealthCheckRemediationTriggerIfAttrTypes = map[string]attr.Type{
	"unhealthy_less_than_or_equal_to": types.StringType,
	"unhealthy_in_range":              types.StringType,
}

type vksClusterMachineDeploymentTopologyDeletionModel struct {
	Order                          types.String `tfsdk:"order"`
	NodeDrainTimeoutSeconds        types.Int32  `tfsdk:"node_drain_timeout_seconds"`
	NodeVolumeDetachTimeoutSeconds types.Int32  `tfsdk:"node_volume_detach_timeout_seconds"`
	NodeDeletionTimeoutSeconds     types.Int32  `tfsdk:"node_deletion_timeout_seconds"`
}

var vksClusterMachineDeploymentTopologyDeletionAttrTypes = map[string]attr.Type{
	"order":                              types.StringType,
	"node_drain_timeout_seconds":         types.Int32Type,
	"node_volume_detach_timeout_seconds": types.Int32Type,
	"node_deletion_timeout_seconds":      types.Int32Type,
}

type vksClusterMachineDeploymentTopologyRolloutModel struct {
	After    timetypes.RFC3339 `tfsdk:"after"`
	Strategy types.Object      `tfsdk:"strategy"`
}

var vksClusterMachineDeploymentTopologyRolloutAttrTypes = map[string]attr.Type{
	"after": timetypes.RFC3339Type{},
	"strategy": types.ObjectType{
		AttrTypes: vksClusterMachineDeploymentTopologyRolloutStrategyAttrTypes,
	},
}

type vksClusterMachineDeploymentTopologyRolloutStrategyModel struct {
	Type          types.String `tfsdk:"type"`
	RollingUpdate types.Object `tfsdk:"rolling_update"`
}

var vksClusterMachineDeploymentTopologyRolloutStrategyAttrTypes = map[string]attr.Type{
	"type": types.StringType,
	"rolling_update": types.ObjectType{
		AttrTypes: vksClusterMachineDeploymentTopologyRolloutStrategyRollingUpdateAttrTypes,
	},
}

type vksClusterMachineDeploymentTopologyRolloutStrategyRollingUpdateModel struct {
	MaxUnavailable types.String `tfsdk:"max_unavailable"`
	MaxSurge       types.String `tfsdk:"max_surge"`
}

var vksClusterMachineDeploymentTopologyRolloutStrategyRollingUpdateAttrTypes = map[string]attr.Type{
	"max_unavailable": types.StringType,
	"max_surge":       types.StringType,
}

// ── Variables ────────────────────────────────────────────────────────────────

type vksClusterVariableModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

var vksClusterVariableAttrTypes = map[string]attr.Type{
	"name":  types.StringType,
	"value": types.StringType,
}

// ── Autoscaler ────────────────────────────────────────────────────────────────

type vksClusterAutoscalerModel struct {
	MinSize types.Int32 `tfsdk:"min_size"`
	MaxSize types.Int32 `tfsdk:"max_size"`
}

var vksClusterAutoscalerAttrTypes = map[string]attr.Type{
	"min_size": types.Int32Type,
	"max_size": types.Int32Type,
}

// ── OS image ─────────────────────────────────────────────────────────────────

type vksClusterOsImageModel struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

var vksClusterOsImageAttrTypes = map[string]attr.Type{
	"name":    types.StringType,
	"version": types.StringType,
}

// ── Status ───────────────────────────────────────────────────────────────────

type vksClusterStatusModel struct {
	Conditions         types.Set    `tfsdk:"conditions"`
	Initialization     types.Object `tfsdk:"initialization"`
	ControlPlane       types.Object `tfsdk:"control_plane"`
	Workers            types.Object `tfsdk:"workers"`
	FailureDomains     types.Set    `tfsdk:"failure_domains"`
	Phase              types.String `tfsdk:"phase"`
	ObservedGeneration types.Int64  `tfsdk:"observed_generation"`
}

var vksClusterStatusAttrTypes = map[string]attr.Type{
	"conditions": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: kubernetes.ConditionAttrTypes,
		},
	},
	"initialization": types.ObjectType{
		AttrTypes: vksClusterStatusInitializationAttrTypes,
	},
	"control_plane": types.ObjectType{
		AttrTypes: vksClusterStatusControlPlaneStatusAttrTypes,
	},
	"workers": types.ObjectType{
		AttrTypes: vksClusterStatusWorkersStatusAttrTypes,
	},
	"failure_domains": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: vksClusterStatusFailureDomainAttrTypes,
		},
	},
	"phase":               types.StringType,
	"observed_generation": types.Int64Type,
}

type vksClusterStatusInitializationModel struct {
	InfrastructureProvisioned types.Bool `tfsdk:"infrastructure_provisioned"`
	ControlPlaneInitialized   types.Bool `tfsdk:"control_plane_initialized"`
}

var vksClusterStatusInitializationAttrTypes = map[string]attr.Type{
	"infrastructure_provisioned": types.BoolType,
	"control_plane_initialized":  types.BoolType,
}

type vksClusterStatusControlPlaneStatusModel struct {
	DesiredReplicas   types.Int32 `tfsdk:"desired_replicas"`
	Replicas          types.Int32 `tfsdk:"replicas"`
	UpToDateReplicas  types.Int32 `tfsdk:"up_to_date_replicas"`
	ReadyReplicas     types.Int32 `tfsdk:"ready_replicas"`
	AvailableReplicas types.Int32 `tfsdk:"available_replicas"`
}

var vksClusterStatusControlPlaneStatusAttrTypes = map[string]attr.Type{
	"desired_replicas":    types.Int32Type,
	"replicas":            types.Int32Type,
	"up_to_date_replicas": types.Int32Type,
	"ready_replicas":      types.Int32Type,
	"available_replicas":  types.Int32Type,
}

type vksClusterStatusWorkersStatusModel struct {
	DesiredReplicas   types.Int32 `tfsdk:"desired_replicas"`
	Replicas          types.Int32 `tfsdk:"replicas"`
	UpToDateReplicas  types.Int32 `tfsdk:"up_to_date_replicas"`
	ReadyReplicas     types.Int32 `tfsdk:"ready_replicas"`
	AvailableReplicas types.Int32 `tfsdk:"available_replicas"`
}

var vksClusterStatusWorkersStatusAttrTypes = map[string]attr.Type{
	"desired_replicas":    types.Int32Type,
	"replicas":            types.Int32Type,
	"up_to_date_replicas": types.Int32Type,
	"ready_replicas":      types.Int32Type,
	"available_replicas":  types.Int32Type,
}

type vksClusterStatusFailureDomainModel struct {
	Name         types.String `tfsdk:"name"`
	ControlPlane types.Bool   `tfsdk:"control_plane"`
	Attributes   types.Map    `tfsdk:"attributes"`
}

var vksClusterStatusFailureDomainAttrTypes = map[string]attr.Type{
	"name":          types.StringType,
	"control_plane": types.BoolType,
	"attributes": types.MapType{
		ElemType: types.StringType,
	},
}

// ── Common ───────────────────────────────────────────────────────────────────

type vksClusterMachineTaintModel struct {
	Key         types.String `tfsdk:"key"`
	Value       types.String `tfsdk:"value"`
	Effect      types.String `tfsdk:"effect"`
	Propagation types.String `tfsdk:"propagation"`
}

var vksClusterMachineTaintAttrTypes = map[string]attr.Type{
	"key":         types.StringType,
	"value":       types.StringType,
	"effect":      types.StringType,
	"propagation": types.StringType,
}

type vksClusterMachineReadinessGateModel struct {
	ConditionType types.String `tfsdk:"condition_type"`
	Polarity      types.String `tfsdk:"polarity"`
}

var vksClusterMachineReadinessGateAttrTypes = map[string]attr.Type{
	"condition_type": types.StringType,
	"polarity":       types.StringType,
}

type vksClusterUnhealthyNodeConditionModel struct {
	Type           types.String `tfsdk:"type"`
	Status         types.String `tfsdk:"status"`
	TimeoutSeconds types.Int32  `tfsdk:"timeout_seconds"`
}

var vksClusterUnhealthyNodeConditionAttrTypes = map[string]attr.Type{
	"type":            types.StringType,
	"status":          types.StringType,
	"timeout_seconds": types.Int32Type,
}

type vksClusterObjectMetaModel struct {
	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
}

var vksClusterObjectMetaAttrTypes = map[string]attr.Type{
	"labels": types.MapType{
		ElemType: types.StringType,
	},
	"annotations": types.MapType{
		ElemType: types.StringType,
	},
}
