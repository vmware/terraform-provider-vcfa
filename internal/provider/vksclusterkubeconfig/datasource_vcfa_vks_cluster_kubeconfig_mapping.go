// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vksclusterkubeconfig

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

func mapVksClusterKubeconfigToModel(_ context.Context, clusterName string, secret *corev1.Secret, model *vcfaVksClusterKubeconfigModel, diags *diag.Diagnostics) {
	kubeConfigBytes, ok := secret.Data[vcfatypes.VksClusterKubeconfigSecretDataKey]
	if !ok || len(kubeConfigBytes) == 0 {
		diags.AddError(
			fmt.Sprintf("error reading %s for %s %s", vcfatypes.LabelVksClusterKubeconfig, vcfatypes.LabelVksCluster, clusterName),
			fmt.Sprintf("secret %s does not contain the %s key", secret.Name, vcfatypes.VksClusterKubeconfigSecretDataKey),
		)
		return
	}
	model.KubeConfigRaw = types.StringValue(string(kubeConfigBytes))

	cfg, err := clientcmd.Load(kubeConfigBytes)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("error reading %s for %s %s", vcfatypes.LabelVksClusterKubeconfig, vcfatypes.LabelVksCluster, clusterName),
			fmt.Sprintf("could not parse kubeconfig from secret %s: %s", secret.Name, err),
		)
		return
	}
	model.ContextName = types.StringValue(cfg.CurrentContext)

	ctxInfo, ok := cfg.Contexts[cfg.CurrentContext]
	if !ok {
		diags.AddError(
			fmt.Sprintf("error reading %s for %s %s", vcfatypes.LabelVksClusterKubeconfig, vcfatypes.LabelVksCluster, clusterName),
			fmt.Sprintf("kubeconfig context %s not found", cfg.CurrentContext),
		)
		return
	}

	model.User = types.StringValue(ctxInfo.AuthInfo)
	if clusterInfo, ok := cfg.Clusters[ctxInfo.Cluster]; ok {
		model.Host = types.StringValue(clusterInfo.Server)
		model.InsecureSkipTLSVerify = types.BoolValue(clusterInfo.InsecureSkipTLSVerify)
		if len(clusterInfo.CertificateAuthorityData) > 0 {
			model.CertificateAuthorityData = types.StringValue(base64.StdEncoding.EncodeToString(clusterInfo.CertificateAuthorityData))
		}
	}
	if authInfo, ok := cfg.AuthInfos[ctxInfo.AuthInfo]; ok {
		model.Token = types.StringValue(authInfo.Token)
		if len(authInfo.ClientCertificateData) > 0 {
			model.ClientCertificateData = types.StringValue(base64.StdEncoding.EncodeToString(authInfo.ClientCertificateData))
		}
		if len(authInfo.ClientKeyData) > 0 {
			model.ClientKeyData = types.StringValue(base64.StdEncoding.EncodeToString(authInfo.ClientKeyData))
		}
	}
}
