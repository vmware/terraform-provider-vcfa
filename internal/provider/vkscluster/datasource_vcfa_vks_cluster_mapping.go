// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkscluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

// mapVksClusterToDataSourceModel populates a vcfaVksClusterDataSourceModel by delegating
// all mapping logic to mapVksClusterToResourceModel and copying the shared fields.
func mapVksClusterToDataSourceModel(ctx context.Context, cluster *vcfatypes.VksCluster, model *vcfaVksClusterDataSourceModel, diags *diag.Diagnostics) {
	rsModel := &vcfaVksClusterResourceModel{}
	mapVksClusterToResourceModel(ctx, cluster, rsModel, diags)

	// Metadata attributes
	model.Metadata = rsModel.Metadata

	// Spec attributes
	model.AvailabilityGates = rsModel.AvailabilityGates
	model.ClusterClass = rsModel.ClusterClass
	model.ClusterNetwork = rsModel.ClusterNetwork
	model.ControlPlane = rsModel.ControlPlane
	model.ControlPlaneEndpoint = rsModel.ControlPlaneEndpoint
	model.MachineDeployments = rsModel.MachineDeployments
	model.Variables = rsModel.Variables
	model.Version = rsModel.Version

	// Status attributes
	model.Status = rsModel.Status
}
