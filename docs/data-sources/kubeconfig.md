---
page_title: "VMware Cloud Foundation Automation: vcfa_kubeconfig"
subcategory: ""
description: |-
  Provides a data source to fetch the kubeconfig data from VMware Cloud Foundation Automation.
---

# Data Source: vcfa_kubeconfig

Provides a data source to fetch the [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) data from VMware Cloud Foundation Automation.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
data "vcfa_kubeconfig" "kube_config" {}

data "vcfa_kubeconfig" "kube_config_supervisor_namespace" {
  project_name              = "default-project"
  supervisor_namespace_name = "demo-supervisor-namespace"
}

# The kubeconfig can be used to configure the Kubernetes provider
provider "kubernetes" {
  host     = data.vcfa_kubeconfig.kube_config.host
  insecure = data.vcfa_kubeconfig.kube_config.insecure_skip_tls_verify
  token    = data.vcfa_kubeconfig.kube_config.token
}
```

## Argument Reference

The following arguments are supported:

- `project_name` - (Optional) The name of the Project where the Supervisor Namespace belongs to
- `supervisor_namespace_name` - (Optional) The name of the [Supervisor Namespace][vcfa_supervisor_namespace-ds] to retrieve the kubeconfig for

## Attribute Reference

- `host` - Hostname of the Kubernetes cluster
- `insecure_skip_tls_verify` - Whether to skip TLS verification when connecting to the Kubernetes cluster
- `token` - Bearer token for authentication to the Kubernetes cluster
- `user` - Bearer token username
- `context_name` - Name of the generated context
- `kube_config_raw` - Raw kubeconfig

[vcfa_supervisor_namespace-ds]: /providers/vmware/vcfa/latest/docs/data-sources/supervisor_namespace