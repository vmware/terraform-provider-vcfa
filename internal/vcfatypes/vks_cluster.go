// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfatypes

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// VksCluster is an alias for the ClusterAPI v1beta2 Cluster type
// This provides full access to all ClusterAPI v1beta2 Cluster CRD fields
type VksCluster = clusterv1.Cluster

// VksClusterTopology is an alias for the ClusterAPI v1beta2 Topology type
type VksClusterTopology = clusterv1.Topology

// VksControlPlaneTopology is an alias for the ClusterAPI v1beta2 ControlPlaneTopology type
type VksControlPlaneTopology = clusterv1.ControlPlaneTopology

// VksMachineDeploymentTopology is an alias for the ClusterAPI v1beta2 MachineDeploymentTopology type
type VksMachineDeploymentTopology = clusterv1.MachineDeploymentTopology

// VksClusterVariable is an alias for the ClusterAPI v1beta2 ClusterVariable type
type VksClusterVariable = clusterv1.ClusterVariable

// VksCondition is an alias for the ClusterAPI v1beta2 Condition type
type VksCondition = clusterv1.Condition //nolint:staticcheck

// VksControlPlaneTopologyHealthCheck is an alias for the ClusterAPI v1beta2 ControlPlaneTopologyHealthCheck type
type VksControlPlaneTopologyHealthCheck = clusterv1.ControlPlaneTopologyHealthCheck

// VksMachineTaint is an alias for the ClusterAPI v1beta2 MachineTaint type
type VksMachineTaint = clusterv1.MachineTaint

// VksMachineReadinessGate is an alias for the ClusterAPI v1beta2 MachineReadinessGate type
type VksMachineReadinessGate = clusterv1.MachineReadinessGate

// VksUnhealthyNodeConditions is an alias for the ClusterAPI v1beta2 UnhealthyNodeCondition type
type VksUnhealthyNodeConditions = clusterv1.UnhealthyNodeCondition

// VksUnhealthyMachineConditions is an alias for the ClusterAPI v1beta2 UnhealthyMachineCondition type
type VksUnhealthyMachineConditions = clusterv1.UnhealthyMachineCondition

// VksObjectMeta is an alias for the ClusterAPI v1beta2 ObjectMeta type
type VksObjectMeta = clusterv1.ObjectMeta

const (
	// OsImageAnnotationKey is the Tanzu annotation used to select the OS image for
	// a control plane or machine deployment topology entry.
	OsImageAnnotationKey = "run.tanzu.vmware.com/resolve-os-image"

	// AutoscalerMinSizeAnnotationKey is the Cluster API annotation that sets the
	// minimum node count for the cluster-autoscaler on a MachineDeployment.
	AutoscalerMinSizeAnnotationKey = "cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size"

	// AutoscalerMaxSizeAnnotationKey is the Cluster API annotation that sets the
	// maximum node count for the cluster-autoscaler on a MachineDeployment.
	AutoscalerMaxSizeAnnotationKey = "cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size"

	// V1Beta2 condition types (using metav1.Condition)
	VksConditionAvailable = "Available"
)

// Constants for ClusterAPI resource types and versions
const (
	VksClusterGroup    = "cluster.x-k8s.io"
	VksClusterVersion  = "v1beta2"
	VksClusterKind     = "Cluster"
	VksClusterResource = "clusters"
)

// Label for logging and error messages
const LabelVksCluster = "VKS Cluster"

// getVksClusterGVR returns the GroupVersionResource for ClusterAPI v1beta2 Cluster
func GetVksClusterGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    VksClusterGroup,
		Version:  VksClusterVersion,
		Resource: VksClusterResource,
	}
}
