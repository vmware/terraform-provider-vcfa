// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkscluster

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── DataSource Top-level model ───────────────────────────────────────────────

type vcfaVksClusterDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Context types.Object `tfsdk:"context"`
	Name    types.String `tfsdk:"name"`

	// Metadata attributes
	Metadata types.Object `tfsdk:"metadata"`

	// Spec attributes
	AvailabilityGates    types.Set    `tfsdk:"availability_gates"`
	ClusterClass         types.Object `tfsdk:"cluster_class"`
	ClusterNetwork       types.Object `tfsdk:"cluster_network"`
	ControlPlane         types.Object `tfsdk:"control_plane"`
	ControlPlaneEndpoint types.Object `tfsdk:"control_plane_endpoint"`
	MachineDeployments   types.Set    `tfsdk:"machine_deployments"`
	Variables            types.Set    `tfsdk:"variables"`
	Version              types.String `tfsdk:"version"`

	// Status attributes
	Status types.Object `tfsdk:"status"`
}
