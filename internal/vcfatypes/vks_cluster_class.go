// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfatypes

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// VksClusterClass is an alias for the ClusterAPI v1beta2 ClusterClass type
type VksClusterClass = clusterv1.ClusterClass

// VksClusterClassStatus is an alias for the ClusterAPI v1beta2 ClusterClassStatus type
type VksClusterClassStatus = clusterv1.ClusterClassStatus

// VksClusterClassTemplateRef is an alias for the ClusterAPI v1beta2 ClusterClassTemplateReference type
type VksClusterClassTemplateRef = clusterv1.ClusterClassTemplateReference

// VksControlPlaneClass is an alias for the ClusterAPI v1beta2 ControlPlaneClass type
type VksControlPlaneClass = clusterv1.ControlPlaneClass

// VksMachineDeploymentClass is an alias for the ClusterAPI v1beta2 MachineDeploymentClass type
type VksMachineDeploymentClass = clusterv1.MachineDeploymentClass

// VksClusterClassPatch is an alias for the ClusterAPI v1beta2 ClusterClassPatch type
type VksClusterClassPatch = clusterv1.ClusterClassPatch

// VksPatchDefinition is an alias for the ClusterAPI v1beta2 PatchDefinition type
type VksPatchDefinition = clusterv1.PatchDefinition

// Constants for ClusterClass resource types and versions
const (
	VksClusterClassGroup           = "cluster.x-k8s.io"
	VksClusterClassVersion         = "v1beta2"
	VksClusterClassKind            = "ClusterClass"
	VksClusterClassResource        = "clusterclasses"
	VksClusterClassSystemNamespace = "vmware-system-vks-public"
)

// Label for logging and error messages
const LabelVksClusterClass = "VKS ClusterClass"

// GetVksClusterClassGVR returns the GroupVersionResource for ClusterAPI v1beta2 ClusterClass
func GetVksClusterClassGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    VksClusterClassGroup,
		Version:  VksClusterClassVersion,
		Resource: VksClusterClassResource,
	}
}
