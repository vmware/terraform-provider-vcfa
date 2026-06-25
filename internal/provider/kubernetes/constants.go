// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import "regexp"

// Kubernetes validation regex patterns derived from upstream kubebuilder:validation:Pattern markers.
var (
	// ReDNSSubdomain matches a DNS subdomain: lowercase alphanumeric, hyphens, dots; start/end alphanumeric.
	ReDNSSubdomain = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

	// ReDNSLabel matches a DNS label: lowercase alphanumeric and hyphens only; start/end alphanumeric.
	ReDNSLabel = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

	// ReK8sKind matches a Kubernetes kind: alphabetic start, alphanumeric and hyphens.
	ReK8sKind = regexp.MustCompile(`^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$`)

	// ReQualifiedName matches a qualified condition/taint-key name: optional dns-subdomain prefix + label.
	ReQualifiedName = regexp.MustCompile(`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$`)

	// ReTaintKey matches a taint key: optional dns-subdomain/label prefix.
	ReTaintKey = regexp.MustCompile(`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/)?([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$`)

	// ReTaintValue matches a taint value: label value format.
	ReTaintValue = regexp.MustCompile(`^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$`)

	// ReUnhealthyInRange matches the unhealthy-in-range pattern, e.g. "[3-5]".
	ReUnhealthyInRange = regexp.MustCompile(`^\[[0-9]+-[0-9]+\]$`)

	// ReAPIVersion matches a fully-qualified API version: dns-subdomain/version.
	ReAPIVersion = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[a-z]([-a-z0-9]*[a-z0-9])?$`)
)
