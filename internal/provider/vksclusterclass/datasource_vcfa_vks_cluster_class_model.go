// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vksclusterclass

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
)

// ── Top-level model ──────────────────────────────────────────────────────────

type vksClusterClassModel struct {
	ID      types.String `tfsdk:"id"`
	Context types.Object `tfsdk:"context"`
	Name    types.String `tfsdk:"name"`
	System  types.Bool   `tfsdk:"system"`

	// Metadata attributes
	Metadata *kubernetes.MetadataModel `tfsdk:"metadata"`

	// Spec attributes
	AvailabilityGates  []vksClusterClassAvailabilityGateModel `tfsdk:"availability_gates"`
	Infrastructure     *vksClusterClassInfrastructureModel    `tfsdk:"infrastructure"`
	ControlPlane       *vksClusterClassControlPlaneModel      `tfsdk:"control_plane"`
	Workers            *vksClusterClassWorkersModel           `tfsdk:"workers"`
	Variables          []vksClusterClassVariableModel         `tfsdk:"variables"`
	Patches            []vksClusterClassPatchModel            `tfsdk:"patches"`
	Upgrade            *vksClusterClassUpgradeModel           `tfsdk:"upgrade"`
	KubernetesVersions types.List                             `tfsdk:"kubernetes_versions"`

	// Status attributes
	Status *vksClusterClassStatusModel `tfsdk:"status"`
}

// ── Shared ─-----────-----────────────────────────────────────────────────────

type vksClusterClassTemplateReferenceModel struct {
	APIVersion types.String `tfsdk:"api_version"`
	Kind       types.String `tfsdk:"kind"`
	Name       types.String `tfsdk:"name"`
}

type vksClusterClassNamingModel struct {
	Template types.String `tfsdk:"template"`
}

type vksClusterClassMachineTaintModel struct {
	Key         types.String `tfsdk:"key"`
	Value       types.String `tfsdk:"value"`
	Effect      types.String `tfsdk:"effect"`
	Propagation types.String `tfsdk:"propagation"`
}

type vksClusterClassMachineReadinessGateModel struct {
	ConditionType types.String `tfsdk:"condition_type"`
	Polarity      types.String `tfsdk:"polarity"`
}

type vksClusterClassUnhealthyNodeConditionModel struct {
	Type           types.String `tfsdk:"type"`
	Status         types.String `tfsdk:"status"`
	TimeoutSeconds types.Int32  `tfsdk:"timeout_seconds"`
}

type vksClusterClassUnhealthyMachineConditionModel struct {
	Type           types.String `tfsdk:"type"`
	Status         types.String `tfsdk:"status"`
	TimeoutSeconds types.Int32  `tfsdk:"timeout_seconds"`
}

type vksClusterClassObjectMetaModel struct {
	Labels      types.Map `tfsdk:"labels"`
	Annotations types.Map `tfsdk:"annotations"`
}

// ── Availability gates ───────────────────────────────────────────────────────

type vksClusterClassAvailabilityGateModel struct {
	ConditionType types.String `tfsdk:"condition_type"`
	Polarity      types.String `tfsdk:"polarity"`
}

// ── Infrastructure ───────────────────────────────────────────────────────────

type vksClusterClassInfrastructureModel struct {
	TemplateRef *vksClusterClassTemplateReferenceModel `tfsdk:"template_ref"`
	Naming      *vksClusterClassNamingModel            `tfsdk:"naming"`
}

// ── ControlPlane ─────────────────────────────────────────────────────────────

type vksClusterClassControlPlaneModel struct {
	Metadata              *vksClusterClassObjectMetaModel                        `tfsdk:"metadata"`
	TemplateRef           *vksClusterClassTemplateReferenceModel                 `tfsdk:"template_ref"`
	MachineInfrastructure *vksClusterClassControlPlaneMachineInfrastructureModel `tfsdk:"machine_infrastructure"`
	HealthCheck           *vksClusterClassControlPlaneHealthCheckModel           `tfsdk:"health_check"`
	Naming                *vksClusterClassNamingModel                            `tfsdk:"naming"`
	Deletion              *vksClusterClassControlPlaneDeletionModel              `tfsdk:"deletion"`
	Taints                []vksClusterClassMachineTaintModel                     `tfsdk:"taints"`
	ReadinessGates        []vksClusterClassMachineReadinessGateModel             `tfsdk:"readiness_gates"`
}

type vksClusterClassControlPlaneMachineInfrastructureModel struct {
	TemplateRef *vksClusterClassTemplateReferenceModel `tfsdk:"template_ref"`
}

type vksClusterClassControlPlaneHealthCheckModel struct {
	Checks      *vksClusterClassControlPlaneHealthCheckChecksModel      `tfsdk:"checks"`
	Remediation *vksClusterClassControlPlaneHealthCheckRemediationModel `tfsdk:"remediation"`
}

type vksClusterClassControlPlaneHealthCheckChecksModel struct {
	NodeStartupTimeoutSeconds  types.Int32                                     `tfsdk:"node_startup_timeout_seconds"`
	UnhealthyNodeConditions    []vksClusterClassUnhealthyNodeConditionModel    `tfsdk:"unhealthy_node_conditions"`
	UnhealthyMachineConditions []vksClusterClassUnhealthyMachineConditionModel `tfsdk:"unhealthy_machine_conditions"`
}

type vksClusterClassControlPlaneHealthCheckRemediationModel struct {
	TriggerIf   *vksClusterClassControlPlaneHealthCheckRemediationTriggerIfModel `tfsdk:"trigger_if"`
	TemplateRef *vksClusterClassTemplateReferenceModel                           `tfsdk:"template_ref"`
}

type vksClusterClassControlPlaneHealthCheckRemediationTriggerIfModel struct {
	UnhealthyLessThanOrEqualTo types.String `tfsdk:"unhealthy_less_than_or_equal_to"`
	UnhealthyInRange           types.String `tfsdk:"unhealthy_in_range"`
}

type vksClusterClassControlPlaneDeletionModel struct {
	NodeDrainTimeoutSeconds        types.Int32 `tfsdk:"node_drain_timeout_seconds"`
	NodeVolumeDetachTimeoutSeconds types.Int32 `tfsdk:"node_volume_detach_timeout_seconds"`
	NodeDeletionTimeoutSeconds     types.Int32 `tfsdk:"node_deletion_timeout_seconds"`
}

// ── Workers ──────────────────────────────────────────────────────────────────

type vksClusterClassWorkersModel struct {
	MachineDeployments []vksClusterClassMachineDeploymentModel `tfsdk:"machine_deployments"`
}

// ── MachineDeployment ─-----──────────────────────────────────────────────────

type vksClusterClassMachineDeploymentModel struct {
	Metadata        *vksClusterClassObjectMetaModel                      `tfsdk:"metadata"`
	Class           types.String                                         `tfsdk:"class"`
	Bootstrap       *vksClusterClassMachineDeploymentBootstrapModel      `tfsdk:"bootstrap"`
	Infrastructure  *vksClusterClassMachineDeploymentInfrastructureModel `tfsdk:"infrastructure"`
	HealthCheck     *vksClusterClassMachineDeploymentHealthCheckModel    `tfsdk:"health_check"`
	FailureDomain   types.String                                         `tfsdk:"failure_domain"`
	Naming          *vksClusterClassNamingModel                          `tfsdk:"naming"`
	Deletion        *vksClusterClassMachineDeploymentDeletionModel       `tfsdk:"deletion"`
	Taints          []vksClusterClassMachineTaintModel                   `tfsdk:"taints"`
	MinReadySeconds types.Int32                                          `tfsdk:"min_ready_seconds"`
	ReadinessGates  []vksClusterClassMachineReadinessGateModel           `tfsdk:"readiness_gates"`
	Rollout         *vksClusterClassMachineDeploymentRolloutModel        `tfsdk:"rollout"`
}

type vksClusterClassMachineDeploymentBootstrapModel struct {
	TemplateRef *vksClusterClassTemplateReferenceModel `tfsdk:"template_ref"`
}

type vksClusterClassMachineDeploymentInfrastructureModel struct {
	TemplateRef *vksClusterClassTemplateReferenceModel `tfsdk:"template_ref"`
}

type vksClusterClassMachineDeploymentHealthCheckModel struct {
	Checks      *vksClusterClassMachineDeploymentHealthCheckChecksModel      `tfsdk:"checks"`
	Remediation *vksClusterClassMachineDeploymentHealthCheckRemediationModel `tfsdk:"remediation"`
}

type vksClusterClassMachineDeploymentHealthCheckChecksModel struct {
	NodeStartupTimeoutSeconds  types.Int32                                     `tfsdk:"node_startup_timeout_seconds"`
	UnhealthyNodeConditions    []vksClusterClassUnhealthyNodeConditionModel    `tfsdk:"unhealthy_node_conditions"`
	UnhealthyMachineConditions []vksClusterClassUnhealthyMachineConditionModel `tfsdk:"unhealthy_machine_conditions"`
}

type vksClusterClassMachineDeploymentHealthCheckRemediationModel struct {
	MaxInFlight types.String                                                          `tfsdk:"max_in_flight"`
	TriggerIf   *vksClusterClassMachineDeploymentHealthCheckRemediationTriggerIfModel `tfsdk:"trigger_if"`
	TemplateRef *vksClusterClassTemplateReferenceModel                                `tfsdk:"template_ref"`
}

type vksClusterClassMachineDeploymentHealthCheckRemediationTriggerIfModel struct {
	UnhealthyLessThanOrEqualTo types.String `tfsdk:"unhealthy_less_than_or_equal_to"`
	UnhealthyInRange           types.String `tfsdk:"unhealthy_in_range"`
}

type vksClusterClassMachineDeploymentDeletionModel struct {
	Order                          types.String `tfsdk:"order"`
	NodeDrainTimeoutSeconds        types.Int32  `tfsdk:"node_drain_timeout_seconds"`
	NodeVolumeDetachTimeoutSeconds types.Int32  `tfsdk:"node_volume_detach_timeout_seconds"`
	NodeDeletionTimeoutSeconds     types.Int32  `tfsdk:"node_deletion_timeout_seconds"`
}

type vksClusterClassMachineDeploymentRolloutModel struct {
	Strategy *vksClusterClassMachineDeploymentRolloutStrategyModel `tfsdk:"strategy"`
}

type vksClusterClassMachineDeploymentRolloutStrategyModel struct {
	Type          types.String                                                       `tfsdk:"type"`
	RollingUpdate *vksClusterClassMachineDeploymentRolloutStrategyRollingUpdateModel `tfsdk:"rolling_update"`
}

type vksClusterClassMachineDeploymentRolloutStrategyRollingUpdateModel struct {
	MaxUnavailable types.String `tfsdk:"max_unavailable"`
	MaxSurge       types.String `tfsdk:"max_surge"`
}

// ── Variables ────────────────────────────────────────────────────────────────

type vksClusterClassVariableModel struct {
	Name     types.String                        `tfsdk:"name"`
	Required types.Bool                          `tfsdk:"required"`
	Schema   *vksClusterClassVariableSchemaModel `tfsdk:"schema"`
}

type vksClusterClassVariableSchemaModel struct {
	OpenAPIV3Schema types.String `tfsdk:"open_api_v3_schema"`
}

// ── Patches ──────────────────────────────────────────────────────────────────

type vksClusterClassPatchModel struct {
	Name        types.String                                 `tfsdk:"name"`
	Description types.String                                 `tfsdk:"description"`
	EnabledIf   types.String                                 `tfsdk:"enabled_if"`
	Definitions []vksClusterClassPatchDefinitionModel        `tfsdk:"definitions"`
	External    *vksClusterClassExternalPatchDefinitionModel `tfsdk:"external"`
}

type vksClusterClassPatchDefinitionModel struct {
	Selector    *vksClusterClassPatchSelectorModel `tfsdk:"selector"`
	JSONPatches []vksClusterClassJsonPatchModel    `tfsdk:"json_patches"`
}

type vksClusterClassPatchSelectorModel struct {
	APIVersion     types.String                            `tfsdk:"api_version"`
	Kind           types.String                            `tfsdk:"kind"`
	MatchResources *vksClusterClassPatchSelectorMatchModel `tfsdk:"match_resources"`
}

type vksClusterClassPatchSelectorMatchModel struct {
	ControlPlane           types.Bool                                                    `tfsdk:"control_plane"`
	InfrastructureCluster  types.Bool                                                    `tfsdk:"infrastructure_cluster"`
	MachineDeploymentClass *vksClusterClassPatchSelectorMatchMachineDeploymentClassModel `tfsdk:"machine_deployment_class"`
}

type vksClusterClassPatchSelectorMatchMachineDeploymentClassModel struct {
	Names types.Set `tfsdk:"names"`
}

type vksClusterClassJsonPatchModel struct {
	Op        types.String                        `tfsdk:"op"`
	Path      types.String                        `tfsdk:"path"`
	Value     types.String                        `tfsdk:"value"`
	ValueFrom *vksClusterClassJsonPatchValueModel `tfsdk:"value_from"`
}

type vksClusterClassJsonPatchValueModel struct {
	Variable types.String `tfsdk:"variable"`
	Template types.String `tfsdk:"template"`
}

type vksClusterClassExternalPatchDefinitionModel struct {
	GeneratePatchesExtension   types.String `tfsdk:"generate_patches_extension"`
	ValidateTopologyExtension  types.String `tfsdk:"validate_topology_extension"`
	DiscoverVariablesExtension types.String `tfsdk:"discover_variables_extension"`
	Settings                   types.Map    `tfsdk:"settings"`
}

// ── Upgrade ──────────────────────────────────────────────────────────────────

type vksClusterClassUpgradeModel struct {
	External *vksClusterClassUpgradeExternalModel `tfsdk:"external"`
}

type vksClusterClassUpgradeExternalModel struct {
	GenerateUpgradePlanExtension types.String `tfsdk:"generate_upgrade_plan_extension"`
}

// ── Status ───────────────────────────────────────────────────────────────────

type vksClusterClassStatusModel struct {
	Conditions         []kubernetes.ConditionModel          `tfsdk:"conditions"`
	Variables          []vksClusterClassStatusVariableModel `tfsdk:"variables"`
	ObservedGeneration types.Int64                          `tfsdk:"observed_generation"`
}

// ── Status Variables ─────────────────────────────────────────────────────────

type vksClusterClassStatusVariableModel struct {
	Name                types.String                                   `tfsdk:"name"`
	DefinitionsConflict types.Bool                                     `tfsdk:"definitions_conflict"`
	Definitions         []vksClusterClassStatusVariableDefinitionModel `tfsdk:"definitions"`
}

type vksClusterClassStatusVariableDefinitionModel struct {
	From     types.String                                       `tfsdk:"from"`
	Required types.Bool                                         `tfsdk:"required"`
	Schema   vksClusterClassStatusVariableDefinitionSchemaModel `tfsdk:"schema"`
}

type vksClusterClassStatusVariableDefinitionSchemaModel struct {
	OpenAPIV3Schema types.String `tfsdk:"open_api_v3_schema"`
}
