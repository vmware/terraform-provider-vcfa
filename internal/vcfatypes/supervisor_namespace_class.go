// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfatypes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	LabelSupervisorNamespaceClass = "Supervisor Namespace Class"
)

const (
	SupervisorNamespaceClassAPI      = "infrastructure.cci.vmware.com"
	SupervisorNamespaceClassVersion  = "v1alpha2"
	SupervisorNamespaceClassKind     = "SupervisorNamespaceClass"
	SupervisorNamespaceClassResource = "supervisornamespaceclasses"

	SupervisorNamespaceClassConfigKind     = "SupervisorNamespaceClassConfig"
	SupervisorNamespaceClassConfigResource = "supervisornamespaceclassconfigs"

	// SupervisorNamespaceClassContentSourceLibrary is the only content source type supported.
	SupervisorNamespaceClassContentSourceLibrary = "ContentLibrary"
)

// GetSupervisorNamespaceClassGVR returns the GroupVersionResource for SupervisorNamespaceClass.
// It is cluster-scoped and served by the base CCI endpoint.
func GetSupervisorNamespaceClassGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    SupervisorNamespaceClassAPI,
		Version:  SupervisorNamespaceClassVersion,
		Resource: SupervisorNamespaceClassResource,
	}
}

// GetSupervisorNamespaceClassConfigGVR returns the GroupVersionResource for
// SupervisorNamespaceClassConfig, which is cluster-scoped and named after its class.
func GetSupervisorNamespaceClassConfigGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    SupervisorNamespaceClassAPI,
		Version:  SupervisorNamespaceClassVersion,
		Resource: SupervisorNamespaceClassConfigResource,
	}
}

// SupervisorNamespaceClass is the template (metadata) of a Supervisor Namespace.
type SupervisorNamespaceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SupervisorNamespaceClassSpec `json:"spec"`
}

type SupervisorNamespaceClassSpec struct {
	Description string `json:"description,omitempty"`
}

// SupervisorNamespaceClassConfig holds the resources and limits of the class with the same name.
type SupervisorNamespaceClassConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SupervisorNamespaceClassConfigSpec `json:"spec"`
}

type SupervisorNamespaceClassConfigSpec struct {
	StorageClasses []SupervisorNamespaceClassStorageClass `json:"storageClasses"`
	VMClasses      []SupervisorNamespaceClassVMClass      `json:"vmClasses"`
	ContentSources []SupervisorNamespaceClassContentSrc   `json:"contentSources,omitempty"`
	DefaultLimits  *SupervisorNamespaceClassLimits        `json:"defaultLimits,omitempty"`
	Zones          []SupervisorNamespaceClassZone         `json:"zones,omitempty"`
}

type SupervisorNamespaceClassStorageClass struct {
	Name  string `json:"name"`
	Limit string `json:"limit"`
}

type SupervisorNamespaceClassVMClass struct {
	Name string `json:"name"`
}

type SupervisorNamespaceClassContentSrc struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type SupervisorNamespaceClassLimits struct {
	CPULimit          string `json:"cpuLimit"`
	MemoryLimit       string `json:"memoryLimit"`
	CPUReservation    string `json:"cpuReservation"`
	MemoryReservation string `json:"memoryReservation"`
}

type SupervisorNamespaceClassZone struct {
	Name                string                                  `json:"name"`
	CPULimit            string                                  `json:"cpuLimit"`
	CPUReservation      string                                  `json:"cpuReservation"`
	MemoryLimit         string                                  `json:"memoryLimit"`
	MemoryReservation   string                                  `json:"memoryReservation"`
	VMClassReservations []SupervisorNamespaceClassVMReservation `json:"vmClassReservations,omitempty"`
}

type SupervisorNamespaceClassVMReservation struct {
	VMClassName string `json:"vmClassName"`
	Count       int32  `json:"count"`
}
