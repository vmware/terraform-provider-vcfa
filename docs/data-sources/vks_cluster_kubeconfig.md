---
page_title: "VMware Cloud Foundation Automation: vcfa_vks_cluster_kubeconfig"
subcategory: ""
description: |-
  Provides a data source to retrieve the kubeconfig of a VKS Cluster.
---

# vcfa_vks_cluster_kubeconfig

Provides a data source to retrieve the [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)
for a VKS (VMware Kubernetes Service) Cluster.

_Used by: **Tenant**_

## Example Usage

```hcl
data "vcfa_vks_cluster_kubeconfig" "my_cluster" {
  context = {
    project   = "my-project"
    namespace = "my-namespace"
  }
  name = "my-vks-cluster"
}

provider "kubernetes" {
  host                   = data.vcfa_vks_cluster_kubeconfig.my_cluster.host
  insecure               = data.vcfa_vks_cluster_kubeconfig.my_cluster.insecure_skip_tls_verify
  cluster_ca_certificate = base64decode(data.vcfa_vks_cluster_kubeconfig.my_cluster.certificate_authority_data)
  client_certificate     = base64decode(data.vcfa_vks_cluster_kubeconfig.my_cluster.client_certificate_data)
  client_key             = base64decode(data.vcfa_vks_cluster_kubeconfig.my_cluster.client_key_data)
}
```

## Argument Reference

The following arguments are supported:

- `context` - (Required) VCF Automation context required to locate the VKS Cluster:
  - `project` - (Required) Name of the Project where the VKS Cluster is located.
  - `namespace` - (Required) Name of the Supervisor Namespace where the VKS Cluster is located.
- `name` - (Required) Name of the VKS Cluster.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `host` - Kubernetes API server URL extracted from the kubeconfig.
- `insecure_skip_tls_verify` - Whether TLS verification is disabled for the Kubernetes API server.
- `kube_config_raw` - Full kubeconfig YAML content. This field is sensitive.
- `context_name` - Name of the current context in the kubeconfig.
- `user` - Name of the user entry in the kubeconfig.
- `token` - Bearer token for authenticating to the Kubernetes API server. Empty for clusters using certificate-based authentication. This field is sensitive.
- `certificate_authority_data` - Base64-encoded PEM certificate authority data for the cluster. Empty when not present in the kubeconfig. This field is sensitive.
- `client_certificate_data` - Base64-encoded PEM client certificate for authenticating to the cluster. Empty when not present in the kubeconfig. This field is sensitive.
- `client_key_data` - Base64-encoded PEM client key for authenticating to the cluster. Empty when not present in the kubeconfig. This field is sensitive.
