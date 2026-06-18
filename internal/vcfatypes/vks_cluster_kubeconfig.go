// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfatypes

// Label for logging and error messages
const LabelVksClusterKubeconfig = "VKS Cluster Kubeconfig"

// Kubeconfig secret constants
const (
	// KubeconfigSecretSuffix is the suffix appended to the cluster name to form the kubeconfig secret name.
	VksClusterKubeconfigSecretSuffix = "-kubeconfig" //nolint:gosec

	// KubeconfigSecretType is the Kubernetes secret type used for CAPI kubeconfig secrets.
	VksClusterKubeconfigSecretType = "cluster.x-k8s.io/secret" //nolint:gosec

	// KubeconfigSecretDataKey is the key in the secret's Data map that holds the kubeconfig YAML.
	VksClusterKubeconfigSecretDataKey = "value"
)
