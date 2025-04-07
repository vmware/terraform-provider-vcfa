/*
 * // © Broadcom. All Rights Reserved.
 * // The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
 * // SPDX-License-Identifier: MPL-2.0
 */

package vcfa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api/v1"
)

func datasourceVcfaKubeConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaKubeConfigRead,
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  fmt.Sprintf("The name of the Project where the %s belongs to", labelSupervisorNamespace),
				RequiredWith: []string{"supervisor_namespace_name"},
			},
			"supervisor_namespace_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  fmt.Sprintf("The name of the %s to retrieve the kubeconfig for", labelSupervisorNamespace),
				RequiredWith: []string{"project_name"},
			},
			"host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Hostname of the Kubernetes cluster",
			},
			"insecure_skip_tls_verify": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether to skip TLS verification when connecting to the Kubernetes cluster",
			},
			"token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Bearer token for authentication to the Kubernetes cluster",
				Sensitive:   true,
			},
			"user": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Bearer token username",
			},
			"context_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the generated context",
			},
			"kube_config_raw": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Raw kubeconfig",
				Sensitive:   true,
			},
		},
	}
}

func datasourceVcfaKubeConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	clusterName := fmt.Sprintf("%s:%s", tmClient.Org, tmClient.Client.VCDHREF.Host)
	clusterServer := fmt.Sprintf(ccitypes.KubernetesSubpath, tmClient.Client.VCDHREF.Scheme, tmClient.Client.VCDHREF.Host)
	contextName := tmClient.Org

	projectName, okProjectName := d.GetOk("project_name")
	supervisorNamespaceName, okSupervisorNamespace := d.GetOk("supervisor_namespace_name")
	if okProjectName && okSupervisorNamespace {
		supervisorNamespace, err := readSupervisorNamespace(tmClient, projectName.(string), supervisorNamespaceName.(string))
		if err != nil {
			return diag.Errorf("error reading %s: %s", labelSupervisorNamespace, err)
		}
		readyStatus := false
		for _, condition := range supervisorNamespace.Status.Conditions {
			if strings.ToLower(condition.Type) == "ready" {
				if strings.ToLower(condition.Status) == "true" {
					readyStatus = true
				}
				break
			}
		}
		if !readyStatus {
			return diag.Errorf("%s %s is not in a ready status", labelSupervisorNamespace, supervisorNamespaceName)
		}
		if supervisorNamespace.Status.NamespaceEndpointURL == "" {
			return diag.Errorf("unable to retrieve the endpoint URL for %s %s", labelSupervisorNamespace, supervisorNamespaceName)
		}
		clusterName = fmt.Sprintf("%s:%s@%s", tmClient.Org, supervisorNamespaceName.(string), tmClient.Client.VCDHREF.Host)
		clusterServer = supervisorNamespace.Status.NamespaceEndpointURL
		contextName = fmt.Sprintf("%s:%s:%s", tmClient.Org, supervisorNamespaceName.(string), projectName.(string))
	}

	token, _, err := new(jwt.Parser).ParseUnverified(tmClient.Client.VCDToken, jwt.MapClaims{})
	if err != nil {
		return diag.Errorf("error parsing JWT token: %s", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return diag.FromErr(errors.New("could not parse claims from JWT token"))
	}
	preferredUsername, ok := claims["preferred_username"].(string)
	if !ok {
		return diag.FromErr(errors.New("could not parse preferred username from JWT token claims"))
	}
	username := fmt.Sprintf("%s:%s@%s", tmClient.Org, preferredUsername, tmClient.Client.VCDHREF.Host)

	kubeconfig := &clientcmdapi.Config{
		Kind:       "Config",
		APIVersion: clientcmdapi.SchemeGroupVersion.Version,
		Clusters: []clientcmdapi.NamedCluster{{
			Name: clusterName,
			Cluster: clientcmdapi.Cluster{
				InsecureSkipTLSVerify: tmClient.InsecureFlag,
				Server:                clusterServer,
			},
		}},
		Contexts: []clientcmdapi.NamedContext{
			{
				Name: contextName,
				Context: clientcmdapi.Context{
					Cluster:  clusterName,
					AuthInfo: username,
				},
			},
		},
		AuthInfos: []clientcmdapi.NamedAuthInfo{
			{
				Name: username,
				AuthInfo: clientcmdapi.AuthInfo{
					Token: token.Raw,
				},
			},
		},
		CurrentContext: contextName,
	}
	if okProjectName && okSupervisorNamespace {
		kubeconfig.Contexts[0].Context.Namespace = supervisorNamespaceName.(string)
	}

	kubeconfigBytes, err := json.MarshalIndent(kubeconfig, "", "  ")
	if err != nil {
		return diag.Errorf("error marshaling kubeconfig: %s", err)
	}

	d.SetId(contextName)
	dSet(d, "host", clusterServer)
	dSet(d, "insecure_skip_tls_verify", tmClient.InsecureFlag)
	dSet(d, "token", token.Raw)
	dSet(d, "user", username)
	dSet(d, "context_name", contextName)
	dSet(d, "kube_config_raw", string(kubeconfigBytes))

	return nil
}
