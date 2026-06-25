// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkscluster

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/common"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/helpers"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/kubernetes"
	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

// ── API → Terraform state ────────────────────────────────────────────────────
func mapVksClusterToResourceModel(ctx context.Context, cluster *vcfatypes.VksCluster, model *vcfaVksClusterResourceModel, diags *diag.Diagnostics) {
	// Metadata
	model.Metadata = helpers.ObjFrom(ctx, kubernetes.MetadataAttrTypes,
		kubernetes.MapMetadataToModel(ctx, cluster.ObjectMeta, diags), diags)

	// Availability gates
	if len(cluster.Spec.AvailabilityGates) > 0 {
		availabilityGates := make([]vksClusterAvailabilityGateModel, 0, len(cluster.Spec.AvailabilityGates))
		for _, ag := range cluster.Spec.AvailabilityGates {
			availabilityGates = append(availabilityGates, vksClusterAvailabilityGateModel{
				ConditionType: types.StringValue(ag.ConditionType),
				Polarity:      types.StringValue(string(ag.Polarity)),
			})
		}
		model.AvailabilityGates = helpers.SetFrom(ctx, types.ObjectType{AttrTypes: vksClusterAvailabilityGateAttrTypes}, availabilityGates, diags)
	} else {
		model.AvailabilityGates = types.SetNull(types.ObjectType{AttrTypes: vksClusterAvailabilityGateAttrTypes})
	}

	// Cluster network
	clusterNetwork := cluster.Spec.ClusterNetwork
	if clusterNetwork.ServiceDomain != "" || len(clusterNetwork.Pods.CIDRBlocks) > 0 || len(clusterNetwork.Services.CIDRBlocks) > 0 {
		serviceDomain := types.StringNull()
		if clusterNetwork.ServiceDomain != "" {
			serviceDomain = types.StringValue(clusterNetwork.ServiceDomain)
		}
		clusterNetworkModel := vksClusterNetworkModel{
			ServiceDomain: serviceDomain,
			Pods:          mapClusterNetworkRangesToModel(ctx, clusterNetwork.Pods.CIDRBlocks, diags),
			Services:      mapClusterNetworkRangesToModel(ctx, clusterNetwork.Services.CIDRBlocks, diags),
		}
		model.ClusterNetwork = helpers.ObjFrom(ctx, vksClusterNetworkAttrTypes, &clusterNetworkModel, diags)
	} else {
		model.ClusterNetwork = types.ObjectNull(vksClusterNetworkAttrTypes)
	}

	// Control Plane endpoint
	controlPlaneEndpoint := cluster.Spec.ControlPlaneEndpoint
	if controlPlaneEndpoint.Host != "" || controlPlaneEndpoint.Port != 0 {
		model.ControlPlaneEndpoint = helpers.ObjFrom(ctx, vksClusterApiEndpointAttrTypes,
			&vksClusterApiEndpointModel{
				Host: types.StringValue(controlPlaneEndpoint.Host),
				Port: types.Int32Value(controlPlaneEndpoint.Port),
			}, diags)
	} else {
		model.ControlPlaneEndpoint = types.ObjectNull(vksClusterApiEndpointAttrTypes)
	}

	// Topology
	if cluster.Spec.Topology.IsDefined() {
		mapClusterTopologyToModel(ctx, cluster.Spec.Topology, model, diags)
	} else {
		model.ClusterClass = types.ObjectNull(vksClusterClassRefAttrTypes)
		model.ControlPlane = types.ObjectNull(vksClusterControlPlaneTopologyAttrTypes)
		model.MachineDeployments = types.SetNull(types.ObjectType{AttrTypes: vksMachineDeploymentTopologyAttrTypes})
		model.Variables = types.SetNull(types.ObjectType{AttrTypes: vksClusterVariableAttrTypes})
		model.Version = types.StringNull()
	}

	// Status
	model.Status = mapClusterStatusToModel(ctx, cluster, diags)
}

// ── Terraform state → API object ─────────────────────────────────────────────
func mapResourceModelToVksCluster(ctx context.Context, model *vcfaVksClusterResourceModel, diags *diag.Diagnostics) *vcfatypes.VksCluster {
	vcfContext := common.ExtractVcfContext(ctx, model.Context, diags)
	cluster := &vcfatypes.VksCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: vcfatypes.VksClusterGroup + "/" + vcfatypes.VksClusterVersion,
			Kind:       vcfatypes.VksClusterKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        model.Name.ValueString(),
			Namespace:   vcfContext.Namespace.ValueString(),
			Labels:      helpers.ExtractStringMap(ctx, model.Labels, diags),
			Annotations: helpers.ExtractStringMap(ctx, model.Annotations, diags),
		},
	}

	// Cluster network
	if !model.ClusterNetwork.IsNull() && !model.ClusterNetwork.IsUnknown() {
		var cn vksClusterNetworkModel
		diags.Append(model.ClusterNetwork.As(ctx, &cn, basetypes.ObjectAsOptions{})...)
		cluster.Spec.ClusterNetwork = clusterv1.ClusterNetwork{}
		if !cn.ServiceDomain.IsNull() && !cn.ServiceDomain.IsUnknown() {
			cluster.Spec.ClusterNetwork.ServiceDomain = cn.ServiceDomain.ValueString()
		}
		if !cn.Pods.IsNull() && !cn.Pods.IsUnknown() {
			var pods vksClusterNetworkRangesModel
			diags.Append(cn.Pods.As(ctx, &pods, basetypes.ObjectAsOptions{})...)
			var prefixes []cidrtypes.IPPrefix
			diags.Append(pods.CIDRBlocks.ElementsAs(ctx, &prefixes, false)...)
			cidrs := make([]string, len(prefixes))
			for i, p := range prefixes {
				cidrs[i] = p.ValueString()
			}
			cluster.Spec.ClusterNetwork.Pods = clusterv1.NetworkRanges{CIDRBlocks: cidrs}
		}
		if !cn.Services.IsNull() && !cn.Services.IsUnknown() {
			var services vksClusterNetworkRangesModel
			diags.Append(cn.Services.As(ctx, &services, basetypes.ObjectAsOptions{})...)
			var prefixes []cidrtypes.IPPrefix
			diags.Append(services.CIDRBlocks.ElementsAs(ctx, &prefixes, false)...)
			cidrs := make([]string, len(prefixes))
			for i, p := range prefixes {
				cidrs[i] = p.ValueString()
			}
			cluster.Spec.ClusterNetwork.Services = clusterv1.NetworkRanges{CIDRBlocks: cidrs}
		}
	}

	topology := clusterv1.Topology{
		Version: model.Version.ValueString(),
	}

	if !model.ClusterClass.IsNull() && !model.ClusterClass.IsUnknown() {
		var cr vksClusterClassRefModel
		diags.Append(model.ClusterClass.As(ctx, &cr, basetypes.ObjectAsOptions{})...)
		topology.ClassRef = clusterv1.ClusterClassRef{
			Name:      cr.Name.ValueString(),
			Namespace: cr.Namespace.ValueString(),
		}
	}

	if !model.Variables.IsNull() && !model.Variables.IsUnknown() {
		var vars []vksClusterVariableModel
		diags.Append(model.Variables.ElementsAs(ctx, &vars, false)...)
		for _, v := range vars {
			cv, err := mapClusterVariableFromModel(v)
			if err != nil {
				diags.AddError("error encoding topology variable", err.Error())
				continue
			}
			topology.Variables = append(topology.Variables, cv)
		}
	}

	if !model.ControlPlane.IsNull() && !model.ControlPlane.IsUnknown() {
		topology.ControlPlane = mapControlPlaneTopologyFromModel(ctx, model.ControlPlane, diags)
	}

	if !model.MachineDeployments.IsNull() && !model.MachineDeployments.IsUnknown() {
		var mdModels []vksClusterMachineDeploymentTopologyModel
		diags.Append(model.MachineDeployments.ElementsAs(ctx, &mdModels, false)...)
		for _, md := range mdModels {
			topology.Workers.MachineDeployments = append(topology.Workers.MachineDeployments,
				mapMachineDeploymentTopologyFromModel(ctx, md, diags))
		}
	}

	cluster.Spec.Topology = topology

	return cluster
}

// ── Shared mapping helpers ─────────────────────────────────────────----------

func mapClusterNetworkRangesToModel(ctx context.Context, blocks []string, diags *diag.Diagnostics) types.Object {
	if len(blocks) == 0 {
		return types.ObjectNull(vksClusterNetworkRangesAttrTypes)
	}
	prefixes := make([]cidrtypes.IPPrefix, 0, len(blocks))
	for _, b := range blocks {
		prefixes = append(prefixes, cidrtypes.NewIPPrefixValue(b))
	}
	cidrBlocks, d := types.SetValueFrom(ctx, cidrtypes.IPPrefixType{}, prefixes)
	diags.Append(d...)
	return helpers.ObjFrom(ctx, vksClusterNetworkRangesAttrTypes, &vksClusterNetworkRangesModel{CIDRBlocks: cidrBlocks}, diags)
}

func mapMachineTaintsToModel(ctx context.Context, taints []vcfatypes.VksMachineTaint, diags *diag.Diagnostics) types.Set {
	if len(taints) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: vksClusterMachineTaintAttrTypes})
	}
	taintsModel := make([]vksClusterMachineTaintModel, 0, len(taints))
	for _, t := range taints {
		taintsModel = append(taintsModel, vksClusterMachineTaintModel{
			Key:         types.StringValue(t.Key),
			Value:       types.StringValue(t.Value),
			Effect:      types.StringValue(string(t.Effect)),
			Propagation: types.StringValue(string(t.Propagation)),
		})
	}
	return helpers.SetFrom(ctx, types.ObjectType{AttrTypes: vksClusterMachineTaintAttrTypes}, taintsModel, diags)
}

func mapMachineTaintsFromModel(taints []vksClusterMachineTaintModel) []clusterv1.MachineTaint {
	machineTaints := make([]clusterv1.MachineTaint, 0, len(taints))
	for _, t := range taints {
		machineTaints = append(machineTaints, clusterv1.MachineTaint{
			Key:         t.Key.ValueString(),
			Value:       t.Value.ValueString(),
			Effect:      corev1.TaintEffect(t.Effect.ValueString()),
			Propagation: clusterv1.MachineTaintPropagation(t.Propagation.ValueString()),
		})
	}
	return machineTaints
}

func mapMachineReadinessGatesToModel(ctx context.Context, gates []vcfatypes.VksMachineReadinessGate, diags *diag.Diagnostics) types.Set {
	if len(gates) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: vksClusterMachineReadinessGateAttrTypes})
	}
	readinessGatesModel := make([]vksClusterMachineReadinessGateModel, 0, len(gates))
	for _, g := range gates {
		readinessGatesModel = append(readinessGatesModel, vksClusterMachineReadinessGateModel{
			ConditionType: types.StringValue(g.ConditionType),
			Polarity:      types.StringValue(string(g.Polarity)),
		})
	}
	return helpers.SetFrom(ctx, types.ObjectType{AttrTypes: vksClusterMachineReadinessGateAttrTypes}, readinessGatesModel, diags)
}

func mapMachineReadinessGatesFromModel(gates []vksClusterMachineReadinessGateModel) []clusterv1.MachineReadinessGate {
	readinessGates := make([]clusterv1.MachineReadinessGate, 0, len(gates))
	for _, g := range gates {
		readinessGates = append(readinessGates, clusterv1.MachineReadinessGate{
			ConditionType: g.ConditionType.ValueString(),
			Polarity:      clusterv1.ConditionPolarity(g.Polarity.ValueString()),
		})
	}
	return readinessGates
}

func mapUnhealthyNodeConditionsToModel(ctx context.Context, nc []vcfatypes.VksUnhealthyNodeConditions, diags *diag.Diagnostics) types.Set {
	if len(nc) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: vksClusterUnhealthyNodeConditionAttrTypes})
	}
	unhealthyNodeConditionsModel := make([]vksClusterUnhealthyNodeConditionModel, 0, len(nc))
	for _, n := range nc {
		unhealthyNodeConditionsModel = append(unhealthyNodeConditionsModel, vksClusterUnhealthyNodeConditionModel{
			Type:           types.StringValue(string(n.Type)),
			Status:         types.StringValue(string(n.Status)),
			TimeoutSeconds: types.Int32PointerValue(n.TimeoutSeconds),
		})
	}
	return helpers.SetFrom(ctx, types.ObjectType{AttrTypes: vksClusterUnhealthyNodeConditionAttrTypes}, unhealthyNodeConditionsModel, diags)
}

func mapUnhealthyNodeConditionsFromModel(nc []vksClusterUnhealthyNodeConditionModel) []vcfatypes.VksUnhealthyNodeConditions {
	unhealthyNodeConditions := make([]vcfatypes.VksUnhealthyNodeConditions, 0, len(nc))
	for _, n := range nc {
		v := vcfatypes.VksUnhealthyNodeConditions{
			Type:   corev1.NodeConditionType(n.Type.ValueString()),
			Status: corev1.ConditionStatus(n.Status.ValueString()),
		}
		if !n.TimeoutSeconds.IsNull() && !n.TimeoutSeconds.IsUnknown() {
			v.TimeoutSeconds = n.TimeoutSeconds.ValueInt32Pointer()
		}
		unhealthyNodeConditions = append(unhealthyNodeConditions, v)
	}
	return unhealthyNodeConditions
}

func mapObjectMetaToModel(ctx context.Context, meta vcfatypes.VksObjectMeta, diags *diag.Diagnostics) types.Object {
	if len(meta.Labels) == 0 && len(meta.Annotations) == 0 {
		return types.ObjectNull(vksClusterObjectMetaAttrTypes)
	}

	// Store null (not empty map) when one side is absent so that removing
	// labels or annotations from config is reflected as null in state and
	// no spurious diff is generated.
	labelMap := types.MapNull(types.StringType)
	if len(meta.Labels) > 0 {
		var d diag.Diagnostics
		labelMap, d = types.MapValueFrom(ctx, types.StringType, meta.Labels)
		diags.Append(d...)
	}

	annotationMap := types.MapNull(types.StringType)
	if len(meta.Annotations) > 0 {
		var d diag.Diagnostics
		annotationMap, d = types.MapValueFrom(ctx, types.StringType, meta.Annotations)
		diags.Append(d...)
	}

	return helpers.ObjFrom(ctx, vksClusterObjectMetaAttrTypes, &vksClusterObjectMetaModel{
		Labels:      labelMap,
		Annotations: annotationMap,
	}, diags)
}

func mapObjectMetaFromModel(ctx context.Context, metaObj types.Object, diags *diag.Diagnostics) clusterv1.ObjectMeta {
	var m vksClusterObjectMetaModel
	diags.Append(metaObj.As(ctx, &m, basetypes.ObjectAsOptions{})...)
	objectMeta := clusterv1.ObjectMeta{}
	if !m.Labels.IsNull() && !m.Labels.IsUnknown() {
		var labels map[string]string
		diags.Append(m.Labels.ElementsAs(ctx, &labels, false)...)
		objectMeta.Labels = labels
	}
	if !m.Annotations.IsNull() && !m.Annotations.IsUnknown() {
		var annotations map[string]string
		diags.Append(m.Annotations.ElementsAs(ctx, &annotations, false)...)
		objectMeta.Annotations = annotations
	}
	return objectMeta
}

func mapVariableOverridesToModel(ctx context.Context, overrides []vcfatypes.VksClusterVariable, diags *diag.Diagnostics) types.Set {
	if len(overrides) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: vksClusterVariableAttrTypes})
	}
	variablesModel := make([]vksClusterVariableModel, 0, len(overrides))
	for _, v := range overrides {
		variablesModel = append(variablesModel, mapClusterVariableToModel(v))
	}
	return helpers.SetFrom(ctx, types.ObjectType{AttrTypes: vksClusterVariableAttrTypes}, variablesModel, diags)
}

func mapClusterVariableToModel(v vcfatypes.VksClusterVariable) vksClusterVariableModel {
	return vksClusterVariableModel{
		Name:  types.StringValue(v.Name),
		Value: types.StringValue(normalizeClusterVariableValue(v.Value.Raw)),
	}
}

func mapClusterVariableFromModel(v vksClusterVariableModel) (clusterv1.ClusterVariable, error) {
	raw := json.RawMessage(v.Value.ValueString())
	if !json.Valid(raw) {
		b, err := json.Marshal(v.Value.ValueString())
		if err != nil {
			return clusterv1.ClusterVariable{}, err
		}
		raw = b
	}
	return clusterv1.ClusterVariable{
		Name:  v.Name.ValueString(),
		Value: apiextensionsv1.JSON{Raw: raw},
	}, nil
}

func normalizeClusterVariableValue(raw []byte) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

// ── Topology mapping helpers ───────────────────────────────────────────────--

func mapClusterTopologyToModel(ctx context.Context, topology vcfatypes.VksClusterTopology, model *vcfaVksClusterResourceModel, diags *diag.Diagnostics) {
	model.ClusterClass = helpers.ObjFrom(ctx, vksClusterClassRefAttrTypes, &vksClusterClassRefModel{
		Name:      types.StringValue(topology.ClassRef.Name),
		Namespace: types.StringValue(topology.ClassRef.Namespace),
	}, diags)

	model.ControlPlane = mapControlPlaneTopologyToModel(ctx, topology.ControlPlane, diags)

	model.MachineDeployments = types.SetNull(types.ObjectType{AttrTypes: vksMachineDeploymentTopologyAttrTypes})
	if len(topology.Workers.MachineDeployments) > 0 {
		mdModels := make([]vksClusterMachineDeploymentTopologyModel, 0, len(topology.Workers.MachineDeployments))
		for _, md := range topology.Workers.MachineDeployments {
			mdModels = append(mdModels, mapMachineDeploymentTopologyToModel(ctx, md, diags))
		}
		model.MachineDeployments = helpers.SetFrom(ctx, types.ObjectType{AttrTypes: vksMachineDeploymentTopologyAttrTypes}, mdModels, diags)
	}

	model.Variables = types.SetNull(types.ObjectType{AttrTypes: vksClusterVariableAttrTypes})
	if len(topology.Variables) > 0 {
		varModels := make([]vksClusterVariableModel, 0, len(topology.Variables))
		for _, v := range topology.Variables {
			if isBackendInjectedVariable(v) {
				continue
			}
			varModels = append(varModels, mapClusterVariableToModel(v))
		}
		if len(varModels) > 0 {
			model.Variables = helpers.SetFrom(ctx, types.ObjectType{AttrTypes: vksClusterVariableAttrTypes}, varModels, diags)
		}
	}

	model.Version = types.StringValue(topology.Version)
}

// ── Control Plane mapping helpers ──────────────────────────────────────------

func mapControlPlaneTopologyToModel(ctx context.Context, cp vcfatypes.VksControlPlaneTopology, diags *diag.Diagnostics) types.Object {
	// mapOsImageToModel deletes the annotation from the map in-place, so it must
	// be called before mapObjectMetaToModel to prevent the annotation from leaking
	// into metadata.annotations in state.
	osImage := mapOsImageToModel(ctx, cp.Metadata.Annotations, diags)
	metadata := mapObjectMetaToModel(ctx, cp.Metadata, diags)

	rollout := types.ObjectNull(vksClusterControlPlaneTopologyRolloutAttrTypes)
	if !cp.Rollout.After.IsZero() {
		rollout = helpers.ObjFrom(ctx, vksClusterControlPlaneTopologyRolloutAttrTypes, &vksClusterControlPlaneTopologyRolloutModel{
			After: timetypes.NewRFC3339TimeValue(cp.Rollout.After.UTC()),
		}, diags)
	}

	healthCheck := mapControlPlaneHealthCheckToModel(ctx, cp.HealthCheck, diags)

	deletion := types.ObjectNull(vksClusterControlPlaneTopologyMachineDeletionAttrTypes)
	if cp.Deletion.NodeDrainTimeoutSeconds != nil || cp.Deletion.NodeVolumeDetachTimeoutSeconds != nil || cp.Deletion.NodeDeletionTimeoutSeconds != nil {
		deletion = helpers.ObjFrom(ctx, vksClusterControlPlaneTopologyMachineDeletionAttrTypes, &vksClusterControlPlaneTopologyMachineDeletionModel{
			NodeDrainTimeoutSeconds:        types.Int32PointerValue(cp.Deletion.NodeDrainTimeoutSeconds),
			NodeVolumeDetachTimeoutSeconds: types.Int32PointerValue(cp.Deletion.NodeVolumeDetachTimeoutSeconds),
			NodeDeletionTimeoutSeconds:     types.Int32PointerValue(cp.Deletion.NodeDeletionTimeoutSeconds),
		}, diags)
	}

	m := vksClusterControlPlaneTopologyModel{
		Metadata:          metadata,
		Replicas:          types.Int32PointerValue(cp.Replicas),
		Rollout:           rollout,
		HealthCheck:       healthCheck,
		Deletion:          deletion,
		Taints:            mapMachineTaintsToModel(ctx, cp.Taints, diags),
		ReadinessGates:    mapMachineReadinessGatesToModel(ctx, cp.ReadinessGates, diags),
		VariableOverrides: mapVariableOverridesToModel(ctx, cp.Variables.Overrides, diags),
		OsImage:           osImage,
	}
	return helpers.ObjFrom(ctx, vksClusterControlPlaneTopologyAttrTypes, &m, diags)
}

func mapControlPlaneTopologyFromModel(ctx context.Context, cpObj types.Object, diags *diag.Diagnostics) clusterv1.ControlPlaneTopology {
	var m vksClusterControlPlaneTopologyModel
	diags.Append(cpObj.As(ctx, &m, basetypes.ObjectAsOptions{})...)

	cpTopology := clusterv1.ControlPlaneTopology{}

	if !m.Metadata.IsNull() && !m.Metadata.IsUnknown() {
		cpTopology.Metadata = mapObjectMetaFromModel(ctx, m.Metadata, diags)
	}

	cpTopology.Metadata.Annotations = injectOsImageAnnotation(ctx, m.OsImage, cpTopology.Metadata.Annotations, diags)

	if !m.Replicas.IsNull() && !m.Replicas.IsUnknown() {
		cpTopology.Replicas = m.Replicas.ValueInt32Pointer()
	}

	if !m.Rollout.IsNull() && !m.Rollout.IsUnknown() {
		var rollout vksClusterControlPlaneTopologyRolloutModel
		diags.Append(m.Rollout.As(ctx, &rollout, basetypes.ObjectAsOptions{})...)
		if !rollout.After.IsNull() && !rollout.After.IsUnknown() {
			t, d := rollout.After.ValueRFC3339Time()
			diags.Append(d...)
			if !diags.HasError() {
				cpTopology.Rollout = clusterv1.ControlPlaneTopologyRolloutSpec{After: metav1.Time{Time: t}}
			}
		}
	}

	if !m.HealthCheck.IsNull() && !m.HealthCheck.IsUnknown() {
		cpTopology.HealthCheck = mapControlPlaneHealthCheckFromModel(ctx, m.HealthCheck, diags)
	}

	if !m.Deletion.IsNull() && !m.Deletion.IsUnknown() {
		var del vksClusterControlPlaneTopologyMachineDeletionModel
		diags.Append(m.Deletion.As(ctx, &del, basetypes.ObjectAsOptions{})...)
		cpTopology.Deletion = clusterv1.ControlPlaneTopologyMachineDeletionSpec{
			NodeDrainTimeoutSeconds:        del.NodeDrainTimeoutSeconds.ValueInt32Pointer(),
			NodeVolumeDetachTimeoutSeconds: del.NodeVolumeDetachTimeoutSeconds.ValueInt32Pointer(),
			NodeDeletionTimeoutSeconds:     del.NodeDeletionTimeoutSeconds.ValueInt32Pointer(),
		}
	}

	if !m.Taints.IsNull() && !m.Taints.IsUnknown() {
		var taints []vksClusterMachineTaintModel
		diags.Append(m.Taints.ElementsAs(ctx, &taints, false)...)
		cpTopology.Taints = mapMachineTaintsFromModel(taints)
	}

	if !m.ReadinessGates.IsNull() && !m.ReadinessGates.IsUnknown() {
		var gates []vksClusterMachineReadinessGateModel
		diags.Append(m.ReadinessGates.ElementsAs(ctx, &gates, false)...)
		cpTopology.ReadinessGates = mapMachineReadinessGatesFromModel(gates)
	}

	if !m.VariableOverrides.IsNull() && !m.VariableOverrides.IsUnknown() {
		var overrides []vksClusterVariableModel
		diags.Append(m.VariableOverrides.ElementsAs(ctx, &overrides, false)...)
		for _, v := range overrides {
			cv, err := mapClusterVariableFromModel(v)
			if err != nil {
				diags.AddError("error encoding control plane variable override", err.Error())
				continue
			}
			cpTopology.Variables.Overrides = append(cpTopology.Variables.Overrides, cv)
		}
	}

	return cpTopology
}

func mapControlPlaneHealthCheckToModel(ctx context.Context, hc vcfatypes.VksControlPlaneTopologyHealthCheck, diags *diag.Diagnostics) types.Object {
	hcNonZero := hc.Enabled != nil ||
		hc.Checks.NodeStartupTimeoutSeconds != nil ||
		len(hc.Checks.UnhealthyNodeConditions) > 0 ||
		hc.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo != nil ||
		hc.Remediation.TriggerIf.UnhealthyInRange != ""
	if !hcNonZero {
		return types.ObjectNull(vksClusterControlPlaneTopologyHealthCheckAttrTypes)
	}

	triggerIf := types.ObjectNull(vksClusterControlPlaneTopologyHealthCheckRemediationTriggerIfAttrTypes)
	if hc.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo != nil || hc.Remediation.TriggerIf.UnhealthyInRange != "" {
		lte := types.StringNull()
		if hc.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo != nil {
			lte = types.StringValue(hc.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo.String())
		}
		inRange := types.StringNull()
		if hc.Remediation.TriggerIf.UnhealthyInRange != "" {
			inRange = types.StringValue(hc.Remediation.TriggerIf.UnhealthyInRange)
		}
		triggerIf = helpers.ObjFrom(ctx, vksClusterControlPlaneTopologyHealthCheckRemediationTriggerIfAttrTypes,
			&vksClusterControlPlaneTopologyHealthCheckRemediationTriggerIfModel{
				UnhealthyLessThanOrEqualTo: lte,
				UnhealthyInRange:           inRange,
			}, diags)
	}

	remediation := types.ObjectNull(vksClusterControlPlaneTopologyHealthCheckRemediationAttrTypes)
	if !triggerIf.IsNull() {
		remediation = helpers.ObjFrom(ctx, vksClusterControlPlaneTopologyHealthCheckRemediationAttrTypes,
			&vksClusterControlPlaneTopologyHealthCheckRemediationModel{
				TriggerIf: triggerIf,
			}, diags)
	}

	checksNonZero := hc.Checks.NodeStartupTimeoutSeconds != nil ||
		len(hc.Checks.UnhealthyNodeConditions) > 0
	checks := types.ObjectNull(vksClusterControlPlaneTopologyHealthCheckChecksAttrTypes)
	if checksNonZero {
		checks = helpers.ObjFrom(ctx, vksClusterControlPlaneTopologyHealthCheckChecksAttrTypes,
			&vksClusterControlPlaneTopologyHealthCheckChecksModel{
				NodeStartupTimeoutSeconds: types.Int32PointerValue(hc.Checks.NodeStartupTimeoutSeconds),
				UnhealthyNodeConditions:   mapUnhealthyNodeConditionsToModel(ctx, hc.Checks.UnhealthyNodeConditions, diags),
			}, diags)
	}

	return helpers.ObjFrom(ctx, vksClusterControlPlaneTopologyHealthCheckAttrTypes,
		&vksClusterControlPlaneTopologyHealthCheckModel{
			Enabled:     types.BoolPointerValue(hc.Enabled),
			Checks:      checks,
			Remediation: remediation,
		}, diags)
}

func mapControlPlaneHealthCheckFromModel(ctx context.Context, hcObj types.Object, diags *diag.Diagnostics) clusterv1.ControlPlaneTopologyHealthCheck {
	var m vksClusterControlPlaneTopologyHealthCheckModel
	diags.Append(hcObj.As(ctx, &m, basetypes.ObjectAsOptions{})...)

	cpTopologyHealthCheck := clusterv1.ControlPlaneTopologyHealthCheck{
		Enabled: m.Enabled.ValueBoolPointer(),
	}

	if !m.Checks.IsNull() && !m.Checks.IsUnknown() {
		var checks vksClusterControlPlaneTopologyHealthCheckChecksModel
		diags.Append(m.Checks.As(ctx, &checks, basetypes.ObjectAsOptions{})...)
		var nodeNC []vksClusterUnhealthyNodeConditionModel
		if !checks.UnhealthyNodeConditions.IsNull() && !checks.UnhealthyNodeConditions.IsUnknown() {
			diags.Append(checks.UnhealthyNodeConditions.ElementsAs(ctx, &nodeNC, false)...)
		}
		cpTopologyHealthCheck.Checks = clusterv1.ControlPlaneTopologyHealthCheckChecks{
			NodeStartupTimeoutSeconds: checks.NodeStartupTimeoutSeconds.ValueInt32Pointer(),
			UnhealthyNodeConditions:   mapUnhealthyNodeConditionsFromModel(nodeNC),
		}
	}

	if !m.Remediation.IsNull() && !m.Remediation.IsUnknown() {
		var rem vksClusterControlPlaneTopologyHealthCheckRemediationModel
		diags.Append(m.Remediation.As(ctx, &rem, basetypes.ObjectAsOptions{})...)
		cpTopologyHealthCheck.Remediation = clusterv1.ControlPlaneTopologyHealthCheckRemediation{}
		if !rem.TriggerIf.IsNull() && !rem.TriggerIf.IsUnknown() {
			var tf vksClusterControlPlaneTopologyHealthCheckRemediationTriggerIfModel
			diags.Append(rem.TriggerIf.As(ctx, &tf, basetypes.ObjectAsOptions{})...)
			cpTopologyHealthCheck.Remediation.TriggerIf = clusterv1.ControlPlaneTopologyHealthCheckRemediationTriggerIf{
				UnhealthyInRange: tf.UnhealthyInRange.ValueString(),
			}
			if !tf.UnhealthyLessThanOrEqualTo.IsNull() && !tf.UnhealthyLessThanOrEqualTo.IsUnknown() && tf.UnhealthyLessThanOrEqualTo.ValueString() != "" {
				v := intstr.Parse(tf.UnhealthyLessThanOrEqualTo.ValueString())
				cpTopologyHealthCheck.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo = &v
			}
		}
	}

	return cpTopologyHealthCheck
}

// ── Machine Deployment mapping helpers ──────────────────────────────---------

func mapMachineDeploymentTopologyToModel(ctx context.Context, md vcfatypes.VksMachineDeploymentTopology, diags *diag.Diagnostics) vksClusterMachineDeploymentTopologyModel {
	// mapOsImageToModel and mapAutoscalerToModel both delete their annotation
	// keys from the map in-place, so they must be called before
	// mapObjectMetaToModel to prevent the annotations from leaking into
	// metadata.annotations in state.
	osImage := mapOsImageToModel(ctx, md.Metadata.Annotations, diags)
	autoscaler := mapAutoscalerToModel(ctx, md.Metadata.Annotations, diags)
	metadata := mapObjectMetaToModel(ctx, md.Metadata, diags)

	failureDomain := types.StringNull()
	if md.FailureDomain != "" {
		failureDomain = types.StringValue(md.FailureDomain)
	}

	healthCheck := mapMachineDeploymentHealthCheckToModel(ctx, md.HealthCheck, diags)

	deletion := types.ObjectNull(vksClusterMachineDeploymentTopologyDeletionAttrTypes)
	if md.Deletion.Order != "" || md.Deletion.NodeDrainTimeoutSeconds != nil || md.Deletion.NodeVolumeDetachTimeoutSeconds != nil || md.Deletion.NodeDeletionTimeoutSeconds != nil {
		order := types.StringNull()
		if md.Deletion.Order != "" {
			order = types.StringValue(string(md.Deletion.Order))
		}
		deletion = helpers.ObjFrom(ctx, vksClusterMachineDeploymentTopologyDeletionAttrTypes, &vksClusterMachineDeploymentTopologyDeletionModel{
			Order:                          order,
			NodeDrainTimeoutSeconds:        types.Int32PointerValue(md.Deletion.NodeDrainTimeoutSeconds),
			NodeVolumeDetachTimeoutSeconds: types.Int32PointerValue(md.Deletion.NodeVolumeDetachTimeoutSeconds),
			NodeDeletionTimeoutSeconds:     types.Int32PointerValue(md.Deletion.NodeDeletionTimeoutSeconds),
		}, diags)
	}

	rollout := mapMachineDeploymentRolloutToModel(ctx, md.Rollout, diags)

	return vksClusterMachineDeploymentTopologyModel{
		Metadata:          metadata,
		Class:             types.StringValue(md.Class),
		Name:              types.StringValue(md.Name),
		FailureDomain:     failureDomain,
		Replicas:          types.Int32PointerValue(md.Replicas),
		Autoscaler:        autoscaler,
		HealthCheck:       healthCheck,
		Deletion:          deletion,
		Taints:            mapMachineTaintsToModel(ctx, md.Taints, diags),
		MinReadySeconds:   types.Int32PointerValue(md.MinReadySeconds),
		ReadinessGates:    mapMachineReadinessGatesToModel(ctx, md.ReadinessGates, diags),
		Rollout:           rollout,
		VariableOverrides: mapVariableOverridesToModel(ctx, md.Variables.Overrides, diags),
		OsImage:           osImage,
	}
}

func mapMachineDeploymentTopologyFromModel(ctx context.Context, md vksClusterMachineDeploymentTopologyModel, diags *diag.Diagnostics) clusterv1.MachineDeploymentTopology {
	mdTopology := clusterv1.MachineDeploymentTopology{
		Class: md.Class.ValueString(),
		Name:  md.Name.ValueString(),
	}

	if !md.Metadata.IsNull() && !md.Metadata.IsUnknown() {
		mdTopology.Metadata = mapObjectMetaFromModel(ctx, md.Metadata, diags)
	}

	mdTopology.Metadata.Annotations = injectOsImageAnnotation(ctx, md.OsImage, mdTopology.Metadata.Annotations, diags)
	mdTopology.Metadata.Annotations = injectAutoscalerAnnotations(ctx, md.Autoscaler, mdTopology.Metadata.Annotations, diags)

	if !md.FailureDomain.IsNull() && !md.FailureDomain.IsUnknown() {
		mdTopology.FailureDomain = md.FailureDomain.ValueString()
	}

	if !md.Replicas.IsNull() && !md.Replicas.IsUnknown() {
		mdTopology.Replicas = md.Replicas.ValueInt32Pointer()
	}

	if !md.HealthCheck.IsNull() && !md.HealthCheck.IsUnknown() {
		mdTopology.HealthCheck = mapMachineDeploymentHealthCheckFromModel(ctx, md.HealthCheck, diags)
	}

	if !md.Deletion.IsNull() && !md.Deletion.IsUnknown() {
		var del vksClusterMachineDeploymentTopologyDeletionModel
		diags.Append(md.Deletion.As(ctx, &del, basetypes.ObjectAsOptions{})...)
		mdTopology.Deletion = clusterv1.MachineDeploymentTopologyMachineDeletionSpec{
			Order:                          clusterv1.MachineSetDeletionOrder(del.Order.ValueString()),
			NodeDrainTimeoutSeconds:        del.NodeDrainTimeoutSeconds.ValueInt32Pointer(),
			NodeVolumeDetachTimeoutSeconds: del.NodeVolumeDetachTimeoutSeconds.ValueInt32Pointer(),
			NodeDeletionTimeoutSeconds:     del.NodeDeletionTimeoutSeconds.ValueInt32Pointer(),
		}
	}

	if !md.Taints.IsNull() && !md.Taints.IsUnknown() {
		var taints []vksClusterMachineTaintModel
		diags.Append(md.Taints.ElementsAs(ctx, &taints, false)...)
		mdTopology.Taints = mapMachineTaintsFromModel(taints)
	}

	if !md.MinReadySeconds.IsNull() && !md.MinReadySeconds.IsUnknown() {
		mdTopology.MinReadySeconds = md.MinReadySeconds.ValueInt32Pointer()
	}

	if !md.ReadinessGates.IsNull() && !md.ReadinessGates.IsUnknown() {
		var gates []vksClusterMachineReadinessGateModel
		diags.Append(md.ReadinessGates.ElementsAs(ctx, &gates, false)...)
		mdTopology.ReadinessGates = mapMachineReadinessGatesFromModel(gates)
	}

	if !md.Rollout.IsNull() && !md.Rollout.IsUnknown() {
		mdTopology.Rollout = mapMachineDeploymentRolloutFromModel(ctx, md.Rollout, diags)
	}

	if !md.VariableOverrides.IsNull() && !md.VariableOverrides.IsUnknown() {
		var overrides []vksClusterVariableModel
		diags.Append(md.VariableOverrides.ElementsAs(ctx, &overrides, false)...)
		for _, v := range overrides {
			cv, err := mapClusterVariableFromModel(v)
			if err != nil {
				diags.AddError("error encoding machine deployment variable override", err.Error())
				continue
			}
			mdTopology.Variables.Overrides = append(mdTopology.Variables.Overrides, cv)
		}
	}

	return mdTopology
}

func mapMachineDeploymentHealthCheckToModel(ctx context.Context, hc clusterv1.MachineDeploymentTopologyHealthCheck, diags *diag.Diagnostics) types.Object {
	hcNonZero := hc.Enabled != nil ||
		hc.Checks.NodeStartupTimeoutSeconds != nil ||
		len(hc.Checks.UnhealthyNodeConditions) > 0 ||
		hc.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo != nil ||
		hc.Remediation.TriggerIf.UnhealthyInRange != ""
	if !hcNonZero {
		return types.ObjectNull(vksClusterMachineDeploymentTopologyHealthCheckAttrTypes)
	}

	triggerIf := types.ObjectNull(vksClusterMachineDeploymentTopologyHealthCheckRemediationTriggerIfAttrTypes)
	if hc.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo != nil || hc.Remediation.TriggerIf.UnhealthyInRange != "" {
		lte := types.StringNull()
		if hc.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo != nil {
			lte = types.StringValue(hc.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo.String())
		}
		inRange := types.StringNull()
		if hc.Remediation.TriggerIf.UnhealthyInRange != "" {
			inRange = types.StringValue(hc.Remediation.TriggerIf.UnhealthyInRange)
		}
		triggerIf = helpers.ObjFrom(ctx, vksClusterMachineDeploymentTopologyHealthCheckRemediationTriggerIfAttrTypes,
			&vksClusterMachineDeploymentTopologyHealthCheckRemediationTriggerIfModel{
				UnhealthyLessThanOrEqualTo: lte,
				UnhealthyInRange:           inRange,
			}, diags)
	}

	maxInFlight := types.StringNull()
	if hc.Remediation.MaxInFlight != nil {
		maxInFlight = types.StringValue(hc.Remediation.MaxInFlight.String())
	}
	remediation := types.ObjectNull(vksClusterMachineDeploymentTopologyHealthCheckRemediationAttrTypes)
	if !maxInFlight.IsNull() || !triggerIf.IsNull() {
		remediation = helpers.ObjFrom(ctx, vksClusterMachineDeploymentTopologyHealthCheckRemediationAttrTypes,
			&vksClusterMachineDeploymentTopologyHealthCheckRemediationModel{
				MaxInFlight: maxInFlight,
				TriggerIf:   triggerIf,
			}, diags)
	}

	checksNonZero := hc.Checks.NodeStartupTimeoutSeconds != nil ||
		len(hc.Checks.UnhealthyNodeConditions) > 0
	checks := types.ObjectNull(vksClusterMachineDeploymentTopologyHealthCheckChecksAttrTypes)
	if checksNonZero {
		checks = helpers.ObjFrom(ctx, vksClusterMachineDeploymentTopologyHealthCheckChecksAttrTypes,
			&vksClusterMachineDeploymentTopologyHealthCheckChecksModel{
				NodeStartupTimeoutSeconds: types.Int32PointerValue(hc.Checks.NodeStartupTimeoutSeconds),
				UnhealthyNodeConditions:   mapUnhealthyNodeConditionsToModel(ctx, hc.Checks.UnhealthyNodeConditions, diags),
			}, diags)
	}

	return helpers.ObjFrom(ctx, vksClusterMachineDeploymentTopologyHealthCheckAttrTypes,
		&vksClusterMachineDeploymentTopologyHealthCheckModel{
			Enabled:     types.BoolPointerValue(hc.Enabled),
			Checks:      checks,
			Remediation: remediation,
		}, diags)
}

func mapMachineDeploymentHealthCheckFromModel(ctx context.Context, hcObj types.Object, diags *diag.Diagnostics) clusterv1.MachineDeploymentTopologyHealthCheck {
	var m vksClusterMachineDeploymentTopologyHealthCheckModel
	diags.Append(hcObj.As(ctx, &m, basetypes.ObjectAsOptions{})...)

	healthCheck := clusterv1.MachineDeploymentTopologyHealthCheck{
		Enabled: m.Enabled.ValueBoolPointer(),
	}

	if !m.Checks.IsNull() && !m.Checks.IsUnknown() {
		var checks vksClusterMachineDeploymentTopologyHealthCheckChecksModel
		diags.Append(m.Checks.As(ctx, &checks, basetypes.ObjectAsOptions{})...)
		var nodeNC []vksClusterUnhealthyNodeConditionModel
		if !checks.UnhealthyNodeConditions.IsNull() && !checks.UnhealthyNodeConditions.IsUnknown() {
			diags.Append(checks.UnhealthyNodeConditions.ElementsAs(ctx, &nodeNC, false)...)
		}
		healthCheck.Checks = clusterv1.MachineDeploymentTopologyHealthCheckChecks{
			NodeStartupTimeoutSeconds: checks.NodeStartupTimeoutSeconds.ValueInt32Pointer(),
			UnhealthyNodeConditions:   mapUnhealthyNodeConditionsFromModel(nodeNC),
		}
	}

	if !m.Remediation.IsNull() && !m.Remediation.IsUnknown() {
		var rem vksClusterMachineDeploymentTopologyHealthCheckRemediationModel
		diags.Append(m.Remediation.As(ctx, &rem, basetypes.ObjectAsOptions{})...)
		healthCheck.Remediation = clusterv1.MachineDeploymentTopologyHealthCheckRemediation{}
		if !rem.MaxInFlight.IsNull() && !rem.MaxInFlight.IsUnknown() && rem.MaxInFlight.ValueString() != "" {
			v := intstr.Parse(rem.MaxInFlight.ValueString())
			healthCheck.Remediation.MaxInFlight = &v
		}
		if !rem.TriggerIf.IsNull() && !rem.TriggerIf.IsUnknown() {
			var tf vksClusterMachineDeploymentTopologyHealthCheckRemediationTriggerIfModel
			diags.Append(rem.TriggerIf.As(ctx, &tf, basetypes.ObjectAsOptions{})...)
			healthCheck.Remediation.TriggerIf = clusterv1.MachineDeploymentTopologyHealthCheckRemediationTriggerIf{
				UnhealthyInRange: tf.UnhealthyInRange.ValueString(),
			}
			if !tf.UnhealthyLessThanOrEqualTo.IsNull() && !tf.UnhealthyLessThanOrEqualTo.IsUnknown() && tf.UnhealthyLessThanOrEqualTo.ValueString() != "" {
				v := intstr.Parse(tf.UnhealthyLessThanOrEqualTo.ValueString())
				healthCheck.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo = &v
			}
		}
	}

	return healthCheck
}

func mapMachineDeploymentRolloutToModel(ctx context.Context, rollout clusterv1.MachineDeploymentTopologyRolloutSpec, diags *diag.Diagnostics) types.Object {
	if rollout.After.IsZero() && rollout.Strategy.Type == "" {
		return types.ObjectNull(vksClusterMachineDeploymentTopologyRolloutAttrTypes)
	}

	rolloutModel := vksClusterMachineDeploymentTopologyRolloutModel{
		After:    timetypes.NewRFC3339Null(),
		Strategy: types.ObjectNull(vksClusterMachineDeploymentTopologyRolloutStrategyAttrTypes),
	}
	if !rollout.After.IsZero() {
		rolloutModel.After = timetypes.NewRFC3339TimeValue(rollout.After.UTC())
	}
	if rollout.Strategy.Type != "" {
		stratModel := vksClusterMachineDeploymentTopologyRolloutStrategyModel{
			Type:          types.StringValue(string(rollout.Strategy.Type)),
			RollingUpdate: types.ObjectNull(vksClusterMachineDeploymentTopologyRolloutStrategyRollingUpdateAttrTypes),
		}
		ru := rollout.Strategy.RollingUpdate
		if ru.MaxUnavailable != nil || ru.MaxSurge != nil {
			ruModel := vksClusterMachineDeploymentTopologyRolloutStrategyRollingUpdateModel{
				MaxUnavailable: types.StringNull(),
				MaxSurge:       types.StringNull(),
			}
			if ru.MaxUnavailable != nil {
				ruModel.MaxUnavailable = types.StringValue(ru.MaxUnavailable.String())
			}
			if ru.MaxSurge != nil {
				ruModel.MaxSurge = types.StringValue(ru.MaxSurge.String())
			}
			stratModel.RollingUpdate = helpers.ObjFrom(ctx, vksClusterMachineDeploymentTopologyRolloutStrategyRollingUpdateAttrTypes, &ruModel, diags)
		}
		rolloutModel.Strategy = helpers.ObjFrom(ctx, vksClusterMachineDeploymentTopologyRolloutStrategyAttrTypes, &stratModel, diags)
	}
	return helpers.ObjFrom(ctx, vksClusterMachineDeploymentTopologyRolloutAttrTypes, &rolloutModel, diags)
}

func mapMachineDeploymentRolloutFromModel(ctx context.Context, rolloutObj types.Object, diags *diag.Diagnostics) clusterv1.MachineDeploymentTopologyRolloutSpec {
	var m vksClusterMachineDeploymentTopologyRolloutModel
	diags.Append(rolloutObj.As(ctx, &m, basetypes.ObjectAsOptions{})...)

	rolloutSpec := clusterv1.MachineDeploymentTopologyRolloutSpec{}

	if !m.After.IsNull() && !m.After.IsUnknown() {
		t, d := m.After.ValueRFC3339Time()
		diags.Append(d...)
		if !diags.HasError() {
			rolloutSpec.After = metav1.Time{Time: t}
		}
	}

	if !m.Strategy.IsNull() && !m.Strategy.IsUnknown() {
		var strat vksClusterMachineDeploymentTopologyRolloutStrategyModel
		diags.Append(m.Strategy.As(ctx, &strat, basetypes.ObjectAsOptions{})...)
		rolloutSpec.Strategy = clusterv1.MachineDeploymentTopologyRolloutStrategy{
			Type: clusterv1.MachineDeploymentRolloutStrategyType(strat.Type.ValueString()),
		}
		if !strat.RollingUpdate.IsNull() && !strat.RollingUpdate.IsUnknown() {
			var ru vksClusterMachineDeploymentTopologyRolloutStrategyRollingUpdateModel
			diags.Append(strat.RollingUpdate.As(ctx, &ru, basetypes.ObjectAsOptions{})...)
			ruAPI := clusterv1.MachineDeploymentTopologyRolloutStrategyRollingUpdate{}
			if !ru.MaxUnavailable.IsNull() && !ru.MaxUnavailable.IsUnknown() && ru.MaxUnavailable.ValueString() != "" {
				v := intstr.Parse(ru.MaxUnavailable.ValueString())
				ruAPI.MaxUnavailable = &v
			}
			if !ru.MaxSurge.IsNull() && !ru.MaxSurge.IsUnknown() && ru.MaxSurge.ValueString() != "" {
				v := intstr.Parse(ru.MaxSurge.ValueString())
				ruAPI.MaxSurge = &v
			}
			rolloutSpec.Strategy.RollingUpdate = ruAPI
		}
	}

	return rolloutSpec
}

// ── Status mapping helpers ───────────────────────────────────────────────────

func mapClusterStatusToModel(ctx context.Context, cluster *vcfatypes.VksCluster, diags *diag.Diagnostics) types.Object {
	conditions := helpers.SetFrom(ctx,
		types.ObjectType{AttrTypes: kubernetes.ConditionAttrTypes},
		kubernetes.MapConditionsToModel(ctx, cluster.Status.Conditions, diags),
		diags)

	initializationModel := vksClusterStatusInitializationModel{
		InfrastructureProvisioned: types.BoolValue(false),
		ControlPlaneInitialized:   types.BoolValue(false),
	}
	if cluster.Status.Initialization.InfrastructureProvisioned != nil {
		initializationModel.InfrastructureProvisioned = types.BoolPointerValue(cluster.Status.Initialization.InfrastructureProvisioned)
	}
	if cluster.Status.Initialization.ControlPlaneInitialized != nil {
		initializationModel.ControlPlaneInitialized = types.BoolPointerValue(cluster.Status.Initialization.ControlPlaneInitialized)
	}

	controlPlaneStatusModel := vksClusterStatusControlPlaneStatusModel{}
	if cp := cluster.Status.ControlPlane; cp != nil {
		controlPlaneStatusModel = vksClusterStatusControlPlaneStatusModel{
			DesiredReplicas:   types.Int32PointerValue(cp.DesiredReplicas),
			Replicas:          types.Int32PointerValue(cp.Replicas),
			UpToDateReplicas:  types.Int32PointerValue(cp.UpToDateReplicas),
			ReadyReplicas:     types.Int32PointerValue(cp.ReadyReplicas),
			AvailableReplicas: types.Int32PointerValue(cp.AvailableReplicas),
		}
	}

	workersStatusModel := vksClusterStatusWorkersStatusModel{}
	if w := cluster.Status.Workers; w != nil {
		workersStatusModel = vksClusterStatusWorkersStatusModel{
			DesiredReplicas:   types.Int32PointerValue(w.DesiredReplicas),
			Replicas:          types.Int32PointerValue(w.Replicas),
			UpToDateReplicas:  types.Int32PointerValue(w.UpToDateReplicas),
			ReadyReplicas:     types.Int32PointerValue(w.ReadyReplicas),
			AvailableReplicas: types.Int32PointerValue(w.AvailableReplicas),
		}
	}

	failureDomainModels := make([]vksClusterStatusFailureDomainModel, 0, len(cluster.Status.FailureDomains))
	for _, fd := range cluster.Status.FailureDomains {
		cpBool := types.BoolValue(false)
		if fd.ControlPlane != nil {
			cpBool = types.BoolValue(*fd.ControlPlane)
		}
		attrs := map[string]string{}
		for k, v := range fd.Attributes {
			attrs[k] = v
		}
		attrsMap, d := types.MapValueFrom(ctx, types.StringType, attrs)
		diags.Append(d...)
		failureDomainModels = append(failureDomainModels, vksClusterStatusFailureDomainModel{
			Name:         types.StringValue(fd.Name),
			ControlPlane: cpBool,
			Attributes:   attrsMap,
		})
	}
	failureDomains := helpers.SetFrom(ctx, types.ObjectType{AttrTypes: vksClusterStatusFailureDomainAttrTypes}, failureDomainModels, diags)

	statusModel := vksClusterStatusModel{
		Conditions:         conditions,
		Initialization:     helpers.ObjFrom(ctx, vksClusterStatusInitializationAttrTypes, &initializationModel, diags),
		ControlPlane:       helpers.ObjFrom(ctx, vksClusterStatusControlPlaneStatusAttrTypes, &controlPlaneStatusModel, diags),
		Workers:            helpers.ObjFrom(ctx, vksClusterStatusWorkersStatusAttrTypes, &workersStatusModel, diags),
		FailureDomains:     failureDomains,
		Phase:              types.StringValue(cluster.Status.Phase),
		ObservedGeneration: types.Int64Value(cluster.Status.ObservedGeneration),
	}
	return helpers.ObjFrom(ctx, vksClusterStatusAttrTypes, &statusModel, diags)
}

// ── OS image annotation helpers ──────────────────────────────────────────────

func mapOsImageToModel(ctx context.Context, annotations map[string]string, diags *diag.Diagnostics) types.Object {
	raw, ok := annotations[vcfatypes.OsImageAnnotationKey]
	if !ok {
		return types.ObjectNull(vksClusterOsImageAttrTypes)
	}

	delete(annotations, vcfatypes.OsImageAnnotationKey)

	name, version := parseOsImageAnnotationValue(raw)

	osImageVersion := types.StringNull()
	if version != "" {
		osImageVersion = types.StringValue(version)
	}

	return helpers.ObjFrom(ctx, vksClusterOsImageAttrTypes, &vksClusterOsImageModel{
		Name:    types.StringValue(name),
		Version: osImageVersion,
	}, diags)
}

func injectOsImageAnnotation(ctx context.Context, osImageObj types.Object, annotations map[string]string, diags *diag.Diagnostics) map[string]string {
	if osImageObj.IsNull() || osImageObj.IsUnknown() {
		return annotations
	}

	var m vksClusterOsImageModel
	diags.Append(osImageObj.As(ctx, &m, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return annotations
	}

	result := make(map[string]string, len(annotations)+1)
	for k, v := range annotations {
		result[k] = v
	}
	result[vcfatypes.OsImageAnnotationKey] = buildOsImageAnnotationValue(m)
	return result
}

// buildOsImageAnnotationValue builds the annotation value from the os_image model.
// Format: "os-name=<name>" or "os-name=<name>,os-version=<version>".
func buildOsImageAnnotationValue(m vksClusterOsImageModel) string {
	v := "os-name=" + m.Name.ValueString()
	if !m.Version.IsNull() && !m.Version.IsUnknown() && m.Version.ValueString() != "" {
		v += ",os-version=" + m.Version.ValueString()
	}
	return v
}

func parseOsImageAnnotationValue(annotation string) (name, version string) {
	for _, part := range strings.Split(annotation, ",") {
		if len(part) > 8 && part[:8] == "os-name=" {
			name = part[8:]
		} else if len(part) > 10 && part[:10] == "os-version=" {
			version = part[10:]
		}
	}
	return
}

// ── Autoscaler annotation helpers ────────────────────────────────────────────

// mapAutoscalerToModel reads the autoscaler annotation keys from the supplied
// annotations map, removes them in-place, and returns the corresponding
// types.Object. The in-place deletion must happen before mapObjectMetaToModel
// is called so the keys do not leak into metadata.annotations in state.
func mapAutoscalerToModel(ctx context.Context, annotations map[string]string, diags *diag.Diagnostics) types.Object {
	rawMin, hasMin := annotations[vcfatypes.AutoscalerMinSizeAnnotationKey]
	rawMax, hasMax := annotations[vcfatypes.AutoscalerMaxSizeAnnotationKey]

	if !hasMin && !hasMax {
		return types.ObjectNull(vksClusterAutoscalerAttrTypes)
	}

	delete(annotations, vcfatypes.AutoscalerMinSizeAnnotationKey)
	delete(annotations, vcfatypes.AutoscalerMaxSizeAnnotationKey)

	minSize := types.Int32Null()
	if hasMin {
		if n, err := strconv.ParseInt(rawMin, 10, 32); err == nil {
			minSize = types.Int32Value(int32(n))
		}
	}

	maxSize := types.Int32Null()
	if hasMax {
		if n, err := strconv.ParseInt(rawMax, 10, 32); err == nil {
			maxSize = types.Int32Value(int32(n))
		}
	}

	return helpers.ObjFrom(ctx, vksClusterAutoscalerAttrTypes, &vksClusterAutoscalerModel{
		MinSize: minSize,
		MaxSize: maxSize,
	}, diags)
}

func injectAutoscalerAnnotations(ctx context.Context, autoscalerObj types.Object, annotations map[string]string, diags *diag.Diagnostics) map[string]string {
	if autoscalerObj.IsNull() || autoscalerObj.IsUnknown() {
		return annotations
	}

	var m vksClusterAutoscalerModel
	diags.Append(autoscalerObj.As(ctx, &m, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return annotations
	}

	result := make(map[string]string, len(annotations)+2)
	for k, v := range annotations {
		result[k] = v
	}

	if !m.MinSize.IsNull() && !m.MinSize.IsUnknown() {
		result[vcfatypes.AutoscalerMinSizeAnnotationKey] = strconv.Itoa(int(m.MinSize.ValueInt32()))
	}
	if !m.MaxSize.IsNull() && !m.MaxSize.IsUnknown() {
		result[vcfatypes.AutoscalerMaxSizeAnnotationKey] = strconv.Itoa(int(m.MaxSize.ValueInt32()))
	}
	return result
}

// ── User-managed labels/annotations helpers ──────────────────────────────────

// filterToUserManagedKeys returns a types.Map whose keys are restricted to those
// that were present in priorState.  For each such key the value from the live API
// map (apiMap) is used so that actual server-side values are reflected.  Keys that
// the API no longer has are silently dropped (Terraform will show an add-diff on
// the next plan if they are still present in config).  When priorState is null/unknown
// or all keys are absent from the API the result is types.MapNull.
func filterToUserManagedKeys(ctx context.Context, apiMap map[string]string, priorState types.Map, diags *diag.Diagnostics) types.Map {
	if priorState.IsNull() || priorState.IsUnknown() || len(priorState.Elements()) == 0 {
		return types.MapNull(types.StringType)
	}
	var priorKeys map[string]string
	diags.Append(priorState.ElementsAs(ctx, &priorKeys, false)...)
	if diags.HasError() {
		return types.MapNull(types.StringType)
	}
	result := make(map[string]attr.Value, len(priorKeys))
	for k := range priorKeys {
		if v, ok := apiMap[k]; ok {
			result[k] = types.StringValue(v)
		}
	}
	if len(result) == 0 {
		return types.MapNull(types.StringType)
	}
	filtered, d := types.MapValue(types.StringType, result)
	diags.Append(d...)
	return filtered
}

// injectPerKeyMapDiffs post-processes a JSON merge-patch so that
// metadata.labels and metadata.annotations are expressed as per-key diffs
// rather than a single null (which would wipe all backend-managed keys).
func injectPerKeyMapDiffs(ctx context.Context, patchBytes []byte, state, plan vcfaVksClusterResourceModel, diags *diag.Diagnostics) ([]byte, error) {
	labelDiff := computePerKeyMapDiff(ctx, state.Labels, plan.Labels, diags)
	annotationDiff := computePerKeyMapDiff(ctx, state.Annotations, plan.Annotations, diags)

	if len(labelDiff) == 0 && len(annotationDiff) == 0 {
		return patchBytes, nil
	}

	var patchMap map[string]any
	if err := json.Unmarshal(patchBytes, &patchMap); err != nil {
		return patchBytes, err
	}

	// Ensure there is a "metadata" map in the patch.
	var metadata map[string]any
	if existing, ok := patchMap["metadata"]; ok {
		if m, ok := existing.(map[string]any); ok {
			metadata = m
		} else {
			metadata = map[string]any{}
		}
	} else {
		metadata = map[string]any{}
	}

	if len(labelDiff) > 0 {
		metadata["labels"] = labelDiff
	} else {
		// No user-managed label changes: ensure we don't accidentally send a
		// coarse {"labels": null} from the raw merge-patch.
		delete(metadata, "labels")
	}

	if len(annotationDiff) > 0 {
		metadata["annotations"] = annotationDiff
	} else {
		delete(metadata, "annotations")
	}

	if len(metadata) > 0 {
		patchMap["metadata"] = metadata
	} else {
		delete(patchMap, "metadata")
	}

	return json.Marshal(patchMap)
}

// computePerKeyMapDiff returns a map[string]any suitable for embedding in a JSON
// merge-patch (RFC 7396).  Keys that exist in oldMap but are absent from newMap
// are set to nil (which serialises as JSON null, signalling deletion to the API).
// Keys that are new or have changed values in newMap are set to their new string
// value.  Unchanged keys are omitted so the patch is minimal.
// This avoids sending {"labels": null} which would erase ALL labels – including
// backend-injected ones – when the user simply removes their last label.
func computePerKeyMapDiff(ctx context.Context, oldMap, newMap types.Map, diags *diag.Diagnostics) map[string]any {
	var oldKeys, newKeys map[string]string
	if !oldMap.IsNull() && !oldMap.IsUnknown() {
		diags.Append(oldMap.ElementsAs(ctx, &oldKeys, false)...)
	}
	if !newMap.IsNull() && !newMap.IsUnknown() {
		diags.Append(newMap.ElementsAs(ctx, &newKeys, false)...)
	}
	if len(oldKeys) == 0 && len(newKeys) == 0 {
		return nil
	}
	result := make(map[string]any)
	for k := range oldKeys {
		if _, ok := newKeys[k]; !ok {
			result[k] = nil // JSON null → remove from API
		}
	}
	for k, v := range newKeys {
		if oldV, ok := oldKeys[k]; !ok || oldV != v {
			result[k] = v
		}
	}
	return result
}

// isBackendInjectedVariable returns true when the variable was automatically injected
// by the backend platform and the user has not customised it with additional keys.
// A variable is considered backend-injected when its name is in backendInjectedVariableKeys
// AND every top-level key in its JSON value is a known injected key (no user-added keys).
// If the user adds any key not listed in backendInjectedVariableKeys, the variable is
// kept in state so that Terraform can track the user's customisations.
func isBackendInjectedVariable(v vcfatypes.VksClusterVariable) bool {
	var backendInjectedVariableKeys = map[string]map[string]struct{}{
		"kubernetes": {
			"certificateRotation": {},
		},
	}

	injectedKeys, ok := backendInjectedVariableKeys[v.Name]
	if !ok {
		return false
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(v.Value.Raw, &obj); err != nil {
		return false
	}
	if len(obj) == 0 {
		return false
	}
	for key := range obj {
		if _, isInjected := injectedKeys[key]; !isInjected {
			return false
		}
	}
	return true
}
