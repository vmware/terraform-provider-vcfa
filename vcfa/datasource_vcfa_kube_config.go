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
