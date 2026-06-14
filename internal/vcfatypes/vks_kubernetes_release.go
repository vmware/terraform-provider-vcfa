// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfatypes

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// KubernetesRelease objects represent Kubernetes releases available via Kubernetes Service, which can be used to create
// KubernetesCluster instances. KRs are immutable to end-users. They are created and managed by Kubernetes Service to
// provide discovery of Kubernetes releases to Kubernetes Service users.
type KubernetesRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubernetesReleaseSpec   `json:"spec,omitempty"`
	Status KubernetesReleaseStatus `json:"status,omitempty"`
}

// KubernetesReleaseSpec defines the desired state of KubernetesRelease
type KubernetesReleaseSpec struct {
	// Version is the fully qualified Semantic Versioning conformant version of the KubernetesRelease.
	// Version MUST be unique across all KubernetesRelease objects.
	Version string `json:"version"`

	// Kubernetes is Kubernetes
	Kubernetes KubernetesSpec `json:"kubernetes"`

	// OSImages lists references to all OSImage objects shipped with this KubernetesRelease.
	OSImages []corev1.LocalObjectReference `json:"osImages,omitempty"`

	// BootstrapPackages lists references to all bootstrap packages shipped with this KubernetesRelease.
	BootstrapPackages []corev1.LocalObjectReference `json:"bootstrapPackages,omitempty"`
}

// KubernetesSpec specifies the details about the Kubernetes distribution shipped by this KubernetesRelease.
type KubernetesSpec struct {
	// Version is Semantic Versioning conformant version of the Kubernetes build shipped by this KubernetesRelease.
	// The same Kubernetes build MAY be shipped by multiple KubernetesReleases.
	Version string `json:"version"`

	// ImageRepository specifies container image registry to pull images from.
	ImageRepository string `json:"imageRepository,omitempty"`

	// Etcd specifies the container image repository and tag for etcd.
	// +optional
	Etcd *ContainerImageInfo `json:"etcd"`

	// Pause specifies the container image repository and tag for pause.
	// +optional
	Pause *ContainerImageInfo `json:"pause,omitempty"`

	// CoreDNS specifies the container image repository and tag for coredns.
	// +optional
	CoreDNS *ContainerImageInfo `json:"coredns"`

	// KubeVIP specifies the container image repository and tag for kube-vip.
	// +optional
	KubeVIP *ContainerImageInfo `json:"kube-vip,omitempty"`
}

// ContainerImageInfo allows to customize the image used for components that are not
// originated from the Kubernetes/Kubernetes release process (such as etcd and coredns).
type ContainerImageInfo struct {
	// ImageRepository sets the container registry to pull images from.
	// if not set, defaults to the ImageRepository defined in KubernetesSpec.
	// +optional
	ImageRepository string `json:"imageRepository,omitempty"`

	// ImageTag specifies a tag for the image.
	ImageTag string `json:"imageTag,omitempty"`
}

// KubernetesReleaseStatus defines the observed state of KubernetesRelease
type KubernetesReleaseStatus struct {
	Conditions []clusterv1.Condition `json:"conditions,omitempty"` //nolint:staticcheck
}

// Constants for KubernetesRelease resource types and versions
const (
	VksKubernetesReleaseGroup    = "kubernetes.vmware.com"
	VksKubernetesReleaseVersion  = "v1alpha1"
	VksKubernetesReleaseKind     = "KubernetesRelease"
	VksKubernetesReleaseResource = "kubernetesreleases"
)

// Label for logging and error messages
const LabelVksKubernetesRelease = "VKS Kubernetes Release"

// GetVksKubernetesReleaseGVR returns the GroupVersionResource for KubernetesRelease
func GetVksKubernetesReleaseGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    VksKubernetesReleaseGroup,
		Version:  VksKubernetesReleaseVersion,
		Resource: VksKubernetesReleaseResource,
	}
}
