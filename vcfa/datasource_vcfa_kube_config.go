package vcfa

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	cciKubernetesSubpath = "%s://%s/cci/kubernetes"
	labelKubeConfig      = "KubeConfig"
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
	cciClient := meta.(ClientContainer).cciClient
	tmClient := meta.(ClientContainer).tmClient

	kubeconfig, kubecfgValues, err := cciClient.GetKubeConfig(tmClient.Org, d.Get("project_name").(string), d.Get("supervisor_namespace_name").(string))
	if err != nil {
		return diag.Errorf("error creating %s: %s", labelKubeConfig, err)
	}

	kubeconfigBytes, err := json.MarshalIndent(kubeconfig, "", "  ")
	if err != nil {
		return diag.Errorf("error marshaling kubeconfig: %s", err)
	}

	d.SetId(kubecfgValues.ContextName)
	dSet(d, "host", kubecfgValues.ClusterServer)
	dSet(d, "insecure_skip_tls_verify", tmClient.InsecureFlag)
	dSet(d, "token", kubecfgValues.Token.Raw)
	dSet(d, "user", kubecfgValues.UserName)
	dSet(d, "context_name", kubeconfig.CurrentContext)
	dSet(d, "kube_config_raw", string(kubeconfigBytes))

	return nil
}

// func datasourceVcfaKubeConfigReadOld(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	tpClient := meta.(ClientContainer).cciClient
// 	tmClient := meta.(ClientContainer).tmClient

// 	clusterName := fmt.Sprintf("%s:%s", tmClient.Org, tmClient.Client.VCDHREF.Host)
// 	clusterServer := fmt.Sprintf(cciKubernetesSubpath, tmClient.Client.VCDHREF.Scheme, tmClient.Client.VCDHREF.Host)
// 	contextName := tmClient.Org

// 	projectName, okProjectName := d.GetOk("project_name")
// 	supervisorNamespaceName, okSupervisorNamespace := d.GetOk("supervisor_namespace_name")

// 	if okProjectName && okSupervisorNamespace {
// 		supervisorNamespace, err := tpClient.GetSupervisorNamespaceByName(projectName.(string), supervisorNamespaceName.(string))
// 		if err != nil {
// 			return diag.Errorf("error reading %s: %s", labelSupervisorNamespace, err)
// 		}
// 		readyStatus := false
// 		for _, condition := range supervisorNamespace.SupervisorNamespace.Status.Conditions {
// 			if strings.ToLower(condition.Type) == "ready" {
// 				if strings.ToLower(condition.Status) == "true" {
// 					readyStatus = true
// 				}
// 				break
// 			}
// 		}
// 		if !readyStatus {
// 			return diag.Errorf("%s %s is not in a ready status", labelSupervisorNamespace, supervisorNamespaceName)
// 		}
// 		if supervisorNamespace.SupervisorNamespace.Status.NamespaceEndpointURL == "" {
// 			return diag.Errorf("unable to retrieve the endpoint URL for %s %s", labelSupervisorNamespace, supervisorNamespaceName)
// 		}
// 		clusterName = fmt.Sprintf("%s:%s@%s", tmClient.Org, supervisorNamespaceName.(string), tmClient.Client.VCDHREF.Host)
// 		clusterServer = supervisorNamespace.SupervisorNamespace.Status.NamespaceEndpointURL
// 		contextName = fmt.Sprintf("%s:%s:%s", tmClient.Org, supervisorNamespaceName.(string), projectName.(string))
// 	}

// 	token, _, err := new(jwt.Parser).ParseUnverified(tmClient.Client.VCDToken, jwt.MapClaims{})
// 	if err != nil {
// 		return diag.Errorf("error parsing JWT token: %s", err)
// 	}
// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if !ok {
// 		return diag.FromErr(errors.New("could not parse claims from JWT token"))
// 	}
// 	preferredUsername, ok := claims["preferred_username"].(string)
// 	if !ok {
// 		return diag.FromErr(errors.New("could not parse preferred username from JWT token claims"))
// 	}
// 	username := fmt.Sprintf("%s:%s@%s", tmClient.Org, preferredUsername, tmClient.Client.VCDHREF.Host)

// 	kubeconfig := &clientcmdapi.Config{
// 		Kind:       "Config",
// 		APIVersion: clientcmdapi.SchemeGroupVersion.Version,
// 		Clusters: map[string]*clientcmdapi.Cluster{
// 			clusterName: {
// 				InsecureSkipTLSVerify: tmClient.InsecureFlag,
// 				Server:                clusterServer,
// 			},
// 		},
// 		Contexts: map[string]*clientcmdapi.Context{
// 			contextName: {
// 				Cluster:  clusterName,
// 				AuthInfo: username,
// 			},
// 		},
// 		AuthInfos: map[string]*clientcmdapi.AuthInfo{
// 			username: {
// 				Token: token.Raw,
// 			},
// 		},
// 		CurrentContext: contextName,
// 	}
// 	if okProjectName && okSupervisorNamespace {
// 		kubeconfig.Contexts[contextName].Namespace = supervisorNamespaceName.(string)
// 	}

// 	kubeconfigBytes, err := json.MarshalIndent(kubeconfig, "", "  ")
// 	if err != nil {
// 		return diag.Errorf("error marshaling kubeconfig: %s", err)
// 	}

// 	d.SetId(contextName)
// 	dSet(d, "host", clusterServer)
// 	dSet(d, "insecure_skip_tls_verify", tmClient.InsecureFlag)
// 	dSet(d, "token", token.Raw)
// 	dSet(d, "user", username)
// 	dSet(d, "context_name", contextName)
// 	dSet(d, "kube_config_raw", string(kubeconfigBytes))

// 	return nil
// }
