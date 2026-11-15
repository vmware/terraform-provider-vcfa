// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfatypes

const (
	// KubernetesSubpath is the base CCI Kubernetes API endpoint on the VCFA host, used to
	// reach the SupervisorNamespace CRD before any specific Supervisor Namespace exists
	// (i.e. scoped to a Project rather than to a Supervisor Namespace's own endpoint).
	KubernetesSubpath = "%s://%s/cci/kubernetes"
)
