// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vksclusterclass

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func mapVksClusterClassToModel(ctx context.Context, clusterClass *vcfatypes.VksClusterClass, model *vksClusterClassModel, diags *diag.Diagnostics) error {
	// Metadata
	model.Metadata = kubernetes.MapMetadataToModel(ctx, clusterClass.ObjectMeta, diags)

	// AvailabilityGates
	for _, ag := range clusterClass.Spec.AvailabilityGates {
		model.AvailabilityGates = append(model.AvailabilityGates, vksClusterClassAvailabilityGateModel{
			ConditionType: types.StringValue(ag.ConditionType),
			Polarity:      types.StringValue(string(ag.Polarity)),
		})
	}

	// Infrastructure
	infra := clusterClass.Spec.Infrastructure
	model.Infrastructure = &vksClusterClassInfrastructureModel{
		TemplateRef: mapClusterClassTemplateRefToModel(infra.TemplateRef),
		Naming:      mapClusterClassNamingToModel(infra.Naming.Template),
	}

	// ControlPlane
	cp := clusterClass.Spec.ControlPlane
	cpMetadata := mapObjectMetaToModel(ctx, cp.Metadata, diags)
	model.ControlPlane = &vksClusterClassControlPlaneModel{
		Metadata:    cpMetadata,
		TemplateRef: mapClusterClassTemplateRefToModel(cp.TemplateRef),
		MachineInfrastructure: &vksClusterClassControlPlaneMachineInfrastructureModel{
			TemplateRef: mapClusterClassTemplateRefToModel(cp.MachineInfrastructure.TemplateRef),
		},
		HealthCheck:    mapControlPlaneHealthCheckToModel(cp),
		Naming:         mapClusterClassNamingToModel(cp.Naming.Template),
		Deletion:       mapControlPlaneDeletionToModel(cp),
		Taints:         mapMachineTaintsToModel(cp.Taints),
		ReadinessGates: mapMachineReadinessGatesToModel(cp.ReadinessGates),
	}

	// Workers
	model.Workers = &vksClusterClassWorkersModel{}
	for _, md := range clusterClass.Spec.Workers.MachineDeployments {
		model.Workers.MachineDeployments = append(model.Workers.MachineDeployments, mapMachineDeploymentToModel(ctx, md, diags))
	}

	// Variables
	for _, v := range clusterClass.Spec.Variables {
		schemaJSON := ""
		if schemaBytes, err := json.Marshal(v.Schema.OpenAPIV3Schema); err == nil {
			schemaJSON = string(schemaBytes)
		}
		model.Variables = append(model.Variables, vksClusterClassVariableModel{
			Name:     types.StringValue(v.Name),
			Required: types.BoolPointerValue(v.Required),
			Schema:   &vksClusterClassVariableSchemaModel{OpenAPIV3Schema: types.StringValue(schemaJSON)},
		})
	}

	// Patches
	for _, p := range clusterClass.Spec.Patches {
		model.Patches = append(model.Patches, mapPatchToModel(ctx, p, diags))
	}

	// Upgrade
	model.Upgrade = &vksClusterClassUpgradeModel{
		External: &vksClusterClassUpgradeExternalModel{
			GenerateUpgradePlanExtension: types.StringValue(clusterClass.Spec.Upgrade.External.GenerateUpgradePlanExtension),
		},
	}

	// KubernetesVersions
	kubernetesVersions, d := types.ListValueFrom(ctx, types.StringType, clusterClass.Spec.KubernetesVersions)
	diags.Append(d...)
	model.KubernetesVersions = kubernetesVersions

	// Status
	model.Status = &vksClusterClassStatusModel{
		Conditions:         kubernetes.MapConditionsToModel(ctx, clusterClass.Status.Conditions, diags),
		ObservedGeneration: types.Int64Value(clusterClass.Status.ObservedGeneration),
	}
	for _, sv := range clusterClass.Status.Variables {
		svModel := vksClusterClassStatusVariableModel{
			Name:                types.StringValue(sv.Name),
			DefinitionsConflict: types.BoolPointerValue(sv.DefinitionsConflict),
		}
		for _, def := range sv.Definitions {
			schemaJSON := ""
			if schemaBytes, err := json.Marshal(def.Schema.OpenAPIV3Schema); err == nil {
				schemaJSON = string(schemaBytes)
			}
			svModel.Definitions = append(svModel.Definitions, vksClusterClassStatusVariableDefinitionModel{
				From:     types.StringValue(def.From),
				Required: types.BoolPointerValue(def.Required),
				Schema:   vksClusterClassStatusVariableDefinitionSchemaModel{OpenAPIV3Schema: types.StringValue(schemaJSON)},
			})
		}
		model.Status.Variables = append(model.Status.Variables, svModel)
	}

	return nil
}

// ── Shared mapping helpers ───────────────────────────────────────────────────

func mapClusterClassTemplateRefToModel(templateRef vcfatypes.VksClusterClassTemplateRef) *vksClusterClassTemplateReferenceModel {
	return &vksClusterClassTemplateReferenceModel{
		APIVersion: types.StringValue(templateRef.APIVersion),
		Kind:       types.StringValue(templateRef.Kind),
		Name:       types.StringValue(templateRef.Name),
	}
}

func mapClusterClassNamingToModel(template string) *vksClusterClassNamingModel {
	return &vksClusterClassNamingModel{Template: types.StringValue(template)}
}

func mapMachineTaintsToModel(taints []vcfatypes.VksMachineTaint) []vksClusterClassMachineTaintModel {
	result := make([]vksClusterClassMachineTaintModel, 0, len(taints))
	for _, t := range taints {
		result = append(result, vksClusterClassMachineTaintModel{
			Key:         types.StringValue(t.Key),
			Value:       types.StringValue(t.Value),
			Effect:      types.StringValue(string(t.Effect)),
			Propagation: types.StringValue(string(t.Propagation)),
		})
	}
	return result
}

func mapMachineReadinessGatesToModel(gates []vcfatypes.VksMachineReadinessGate) []vksClusterClassMachineReadinessGateModel {
	result := make([]vksClusterClassMachineReadinessGateModel, 0, len(gates))
	for _, g := range gates {
		result = append(result, vksClusterClassMachineReadinessGateModel{
			ConditionType: types.StringValue(g.ConditionType),
			Polarity:      types.StringValue(string(g.Polarity)),
		})
	}
	return result
}

func mapUnhealthyNodeConditionsToModel(nc []vcfatypes.VksUnhealthyNodeConditions) []vksClusterClassUnhealthyNodeConditionModel {
	result := make([]vksClusterClassUnhealthyNodeConditionModel, 0, len(nc))
	for _, n := range nc {
		result = append(result, vksClusterClassUnhealthyNodeConditionModel{
			Type:           types.StringValue(string(n.Type)),
			Status:         types.StringValue(string(n.Status)),
			TimeoutSeconds: types.Int32PointerValue(n.TimeoutSeconds),
		})
	}
	return result
}

func mapUnhealthyMachineConditionsToModel(mc []vcfatypes.VksUnhealthyMachineConditions) []vksClusterClassUnhealthyMachineConditionModel {
	result := make([]vksClusterClassUnhealthyMachineConditionModel, 0, len(mc))
	for _, m := range mc {
		result = append(result, vksClusterClassUnhealthyMachineConditionModel{
			Type:           types.StringValue(m.Type),
			Status:         types.StringValue(string(m.Status)),
			TimeoutSeconds: types.Int32PointerValue(m.TimeoutSeconds),
		})
	}
	return result
}

func mapObjectMetaToModel(ctx context.Context, meta vcfatypes.VksObjectMeta, diags *diag.Diagnostics) *vksClusterClassObjectMetaModel {
	m := &vksClusterClassObjectMetaModel{}

	labels := meta.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	labelMap, d := types.MapValueFrom(ctx, types.StringType, labels)
	diags.Append(d...)
	m.Labels = labelMap

	annotations := meta.Annotations
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotationMap, d := types.MapValueFrom(ctx, types.StringType, annotations)
	diags.Append(d...)
	m.Annotations = annotationMap

	return m
}

// ── ControlPlane mapping helpers ─────────────────────────────────────────────

func mapControlPlaneHealthCheckToModel(cp vcfatypes.VksControlPlaneClass) *vksClusterClassControlPlaneHealthCheckModel {
	hc := cp.HealthCheck
	var lte string
	if hc.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo != nil {
		lte = hc.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo.String()
	}
	return &vksClusterClassControlPlaneHealthCheckModel{
		Checks: &vksClusterClassControlPlaneHealthCheckChecksModel{
			NodeStartupTimeoutSeconds:  types.Int32PointerValue(hc.Checks.NodeStartupTimeoutSeconds),
			UnhealthyNodeConditions:    mapUnhealthyNodeConditionsToModel(hc.Checks.UnhealthyNodeConditions),
			UnhealthyMachineConditions: mapUnhealthyMachineConditionsToModel(hc.Checks.UnhealthyMachineConditions),
		},
		Remediation: &vksClusterClassControlPlaneHealthCheckRemediationModel{
			TriggerIf: &vksClusterClassControlPlaneHealthCheckRemediationTriggerIfModel{
				UnhealthyLessThanOrEqualTo: types.StringValue(lte),
				UnhealthyInRange:           types.StringValue(hc.Remediation.TriggerIf.UnhealthyInRange),
			},
			TemplateRef: &vksClusterClassTemplateReferenceModel{
				APIVersion: types.StringValue(hc.Remediation.TemplateRef.APIVersion),
				Kind:       types.StringValue(hc.Remediation.TemplateRef.Kind),
				Name:       types.StringValue(hc.Remediation.TemplateRef.Name),
			},
		},
	}
}

func mapControlPlaneDeletionToModel(cp vcfatypes.VksControlPlaneClass) *vksClusterClassControlPlaneDeletionModel {
	del := cp.Deletion
	return &vksClusterClassControlPlaneDeletionModel{
		NodeDrainTimeoutSeconds:        types.Int32PointerValue(del.NodeDrainTimeoutSeconds),
		NodeVolumeDetachTimeoutSeconds: types.Int32PointerValue(del.NodeVolumeDetachTimeoutSeconds),
		NodeDeletionTimeoutSeconds:     types.Int32PointerValue(del.NodeDeletionTimeoutSeconds),
	}
}

// ── MachineDeployment maping helpers ─────────────────────────────────────────

func mapMachineDeploymentToModel(ctx context.Context, md vcfatypes.VksMachineDeploymentClass, diags *diag.Diagnostics) vksClusterClassMachineDeploymentModel {
	mdMetadata := mapObjectMetaToModel(ctx, md.Metadata, diags)

	var lte string
	if md.HealthCheck.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo != nil {
		lte = md.HealthCheck.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo.String()
	}
	var maxInFlight string
	if md.HealthCheck.Remediation.MaxInFlight != nil {
		maxInFlight = md.HealthCheck.Remediation.MaxInFlight.String()
	}

	var ruModel *vksClusterClassMachineDeploymentRolloutStrategyRollingUpdateModel
	if ru := md.Rollout.Strategy.RollingUpdate; ru.MaxUnavailable != nil || ru.MaxSurge != nil {
		ruModel = &vksClusterClassMachineDeploymentRolloutStrategyRollingUpdateModel{}
		if ru.MaxUnavailable != nil {
			ruModel.MaxUnavailable = types.StringValue(ru.MaxUnavailable.String())
		}
		if ru.MaxSurge != nil {
			ruModel.MaxSurge = types.StringValue(ru.MaxSurge.String())
		}
	}

	return vksClusterClassMachineDeploymentModel{
		Metadata: mdMetadata,
		Class:    types.StringValue(md.Class),
		Bootstrap: &vksClusterClassMachineDeploymentBootstrapModel{
			TemplateRef: mapClusterClassTemplateRefToModel(md.Bootstrap.TemplateRef),
		},
		Infrastructure: &vksClusterClassMachineDeploymentInfrastructureModel{
			TemplateRef: mapClusterClassTemplateRefToModel(md.Infrastructure.TemplateRef),
		},
		HealthCheck: &vksClusterClassMachineDeploymentHealthCheckModel{
			Checks: &vksClusterClassMachineDeploymentHealthCheckChecksModel{
				NodeStartupTimeoutSeconds:  types.Int32PointerValue(md.HealthCheck.Checks.NodeStartupTimeoutSeconds),
				UnhealthyNodeConditions:    mapUnhealthyNodeConditionsToModel(md.HealthCheck.Checks.UnhealthyNodeConditions),
				UnhealthyMachineConditions: mapUnhealthyMachineConditionsToModel(md.HealthCheck.Checks.UnhealthyMachineConditions),
			},
			Remediation: &vksClusterClassMachineDeploymentHealthCheckRemediationModel{
				MaxInFlight: types.StringValue(maxInFlight),
				TriggerIf: &vksClusterClassMachineDeploymentHealthCheckRemediationTriggerIfModel{
					UnhealthyLessThanOrEqualTo: types.StringValue(lte),
					UnhealthyInRange:           types.StringValue(md.HealthCheck.Remediation.TriggerIf.UnhealthyInRange),
				},
				TemplateRef: &vksClusterClassTemplateReferenceModel{
					APIVersion: types.StringValue(md.HealthCheck.Remediation.TemplateRef.APIVersion),
					Kind:       types.StringValue(md.HealthCheck.Remediation.TemplateRef.Kind),
					Name:       types.StringValue(md.HealthCheck.Remediation.TemplateRef.Name),
				},
			},
		},
		FailureDomain: types.StringValue(md.FailureDomain),
		Naming:        mapClusterClassNamingToModel(md.Naming.Template),
		Deletion: &vksClusterClassMachineDeploymentDeletionModel{
			Order:                          types.StringValue(string(md.Deletion.Order)),
			NodeDrainTimeoutSeconds:        types.Int32PointerValue(md.Deletion.NodeDrainTimeoutSeconds),
			NodeVolumeDetachTimeoutSeconds: types.Int32PointerValue(md.Deletion.NodeVolumeDetachTimeoutSeconds),
			NodeDeletionTimeoutSeconds:     types.Int32PointerValue(md.Deletion.NodeDeletionTimeoutSeconds),
		},
		Taints:          mapMachineTaintsToModel(md.Taints),
		MinReadySeconds: types.Int32PointerValue(md.MinReadySeconds),
		ReadinessGates:  mapMachineReadinessGatesToModel(md.ReadinessGates),
		Rollout: &vksClusterClassMachineDeploymentRolloutModel{
			Strategy: &vksClusterClassMachineDeploymentRolloutStrategyModel{
				Type:          types.StringValue(string(md.Rollout.Strategy.Type)),
				RollingUpdate: ruModel,
			},
		},
	}
}

// ── Patch mapping helpers ────────────────────────────────────────────────────

func mapPatchToModel(ctx context.Context, p vcfatypes.VksClusterClassPatch, diags *diag.Diagnostics) vksClusterClassPatchModel {
	patchModel := vksClusterClassPatchModel{
		Name:        types.StringValue(p.Name),
		Description: types.StringValue(p.Description),
		EnabledIf:   types.StringValue(p.EnabledIf),
	}

	for _, def := range p.Definitions {
		patchModel.Definitions = append(patchModel.Definitions, mapPatchDefinitionToModel(ctx, def, diags))
	}

	if p.External != nil {
		settings, d := types.MapValueFrom(ctx, types.StringType, p.External.Settings)
		diags.Append(d...)
		patchModel.External = &vksClusterClassExternalPatchDefinitionModel{
			GeneratePatchesExtension:   types.StringValue(p.External.GeneratePatchesExtension),
			ValidateTopologyExtension:  types.StringValue(p.External.ValidateTopologyExtension),
			DiscoverVariablesExtension: types.StringValue(p.External.DiscoverVariablesExtension),
			Settings:                   settings,
		}
	}

	return patchModel
}

func mapPatchDefinitionToModel(ctx context.Context, def vcfatypes.VksPatchDefinition, diags *diag.Diagnostics) vksClusterClassPatchDefinitionModel {
	sel := def.Selector
	mr := sel.MatchResources
	matchModel := &vksClusterClassPatchSelectorMatchModel{
		ControlPlane:          types.BoolPointerValue(mr.ControlPlane),
		InfrastructureCluster: types.BoolPointerValue(mr.InfrastructureCluster),
	}
	if mr.MachineDeploymentClass != nil {
		mdcNames, d := types.SetValueFrom(ctx, types.StringType, mr.MachineDeploymentClass.Names)
		diags.Append(d...)
		matchModel.MachineDeploymentClass = &vksClusterClassPatchSelectorMatchMachineDeploymentClassModel{Names: mdcNames}
	}

	defModel := vksClusterClassPatchDefinitionModel{
		Selector: &vksClusterClassPatchSelectorModel{
			APIVersion:     types.StringValue(sel.APIVersion),
			Kind:           types.StringValue(sel.Kind),
			MatchResources: matchModel,
		},
	}

	for _, jp := range def.JSONPatches {
		jpModel := vksClusterClassJsonPatchModel{
			Op:   types.StringValue(jp.Op),
			Path: types.StringValue(jp.Path),
		}
		if jp.Value != nil {
			jpModel.Value = types.StringValue(string(jp.Value.Raw))
		} else {
			jpModel.Value = types.StringNull()
		}
		if jp.ValueFrom != nil {
			jpModel.ValueFrom = &vksClusterClassJsonPatchValueModel{
				Variable: types.StringValue(jp.ValueFrom.Variable),
				Template: types.StringValue(jp.ValueFrom.Template),
			}
		}
		defModel.JSONPatches = append(defModel.JSONPatches, jpModel)
	}

	return defModel
}
