---
page_title: "VMware Cloud Foundation Automation: vcfa_kubeconfig"
subcategory: ""
description: |-
  Provides a data source to fetch the kubeconfig data from VMware Cloud Foundation Automation.
---

# vcfa_kubeconfig

Provides a data source to fetch the [kubeconfig][kubeconfig] data from VMware Cloud Foundation Automation.

_Used by: **Tenant**_

## Example Usage for a VCFA context

To retrieve a [kubeconfig][kubeconfig] that allows managing VCFA Kubernetes resources:

```hcl
data "vcfa_kubeconfig" "kube_config" {}
```

## Example Usage for a VCFA Namespace context

To retrieve a [kubeconfig][kubeconfig] that allows managing Kubernetes resources inside a VCFA namespace:

``` hcl
data "vcfa_kubeconfig" "kube_config_supervisor_namespace" {
  project_name              = "default-project"
  supervisor_namespace_name = "demo-supervisor-namespace"
}
```

## Example Usage for the Kubernetes Provider

The datasource attributes can be used to configure the [Kubernetes provider][kubernetes_provider]:

```hcl
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
[kubeconfig]: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
[kubernetes_provider]: https://registry.terraform.io/providers/hashicorp/kubernetes
