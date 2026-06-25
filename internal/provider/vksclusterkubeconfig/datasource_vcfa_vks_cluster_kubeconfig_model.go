// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vksclusterkubeconfig

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── Top-level model ──────────────────────────────────────────────────────────

type vcfaVksClusterKubeconfigModel struct {
	ID      types.String `tfsdk:"id"`
	Context types.Object `tfsdk:"context"`
	Name    types.String `tfsdk:"name"`

	Host                     types.String `tfsdk:"host"`
	InsecureSkipTLSVerify    types.Bool   `tfsdk:"insecure_skip_tls_verify"`
	KubeConfigRaw            types.String `tfsdk:"kube_config_raw"`
	ContextName              types.String `tfsdk:"context_name"`
	User                     types.String `tfsdk:"user"`
	Token                    types.String `tfsdk:"token"`
	CertificateAuthorityData types.String `tfsdk:"certificate_authority_data"`
	ClientCertificateData    types.String `tfsdk:"client_certificate_data"`
	ClientKeyData            types.String `tfsdk:"client_key_data"`
}
