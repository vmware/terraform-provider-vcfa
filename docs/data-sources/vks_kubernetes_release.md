---
page_title: "VMware Cloud Foundation Automation: vcfa_vks_kubernetes_release"
subcategory: ""
description: |-
  Provides a data source to read a VKS Kubernetes Release from VMware Cloud Foundation Automation.
---

# vcfa_vks_kubernetes_release

Provides a data source to read a VKS `KubernetesRelease` resource from VMware Cloud Foundation Automation.

A `KubernetesRelease` is a read-only, immutable object created and managed by the Kubernetes Service. It describes a specific Kubernetes release available for provisioning VKS clusters, including the exact component image references for `etcd`, `coredns`, `pause`, and `kube-vip`.

_Used by: **Tenant**_

## Example Usage

```hcl
data "vcfa_vks_kubernetes_release" "kubernetes_release" {
  context = {
    project   = "my-project"
    namespace = "my-namespace"
  }

  name = "v1.34.1---vmware.1-vkr.4"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the KubernetesRelease to read (e.g. `v1.34.1---vmware.1-vkr.4`).
- `context` - (Required) VCF Automation context for looking up the KubernetesRelease. See [Context](#context).

## Context

The `context` attribute has the following structure:

- `project` - (Required) Name of the Project where the resource is located.
- `namespace` - (Required) Name of the Namespace where the resource is located.

## Attribute Reference

In addition to the arguments above, the following computed attributes are exported:

- `id` - Internal identifier.
- `metadata` - Standard Kubernetes object metadata. See [Metadata](#metadata).
- `version` - Fully qualified Semantic Versioning conformant version of the KubernetesRelease.
- `kubernetes` - Details about the Kubernetes distribution shipped by this release. See [Kubernetes](#kubernetes).
- `os_images` - Set of OSImage object names shipped with this release.
- `bootstrap_packages` - Set of bootstrap package object names shipped with this release.
- `status` - Observed state of the KubernetesRelease. See [Status](#status).

## Metadata

The `metadata` attribute exposes the standard Kubernetes object metadata:

- `name` - Name of the object.
- `generate_name` - Optional server-side prefix used to generate a unique name.
- `namespace` - Namespace of the object.
- `uid` - Universally unique identifier assigned by the server at creation time.
- `resource_version` - Opaque string used to detect object changes.
- `generation` - Monotonically increasing sequence number for the desired state.
- `creation_timestamp` - RFC3339 timestamp when the object was created.
- `deletion_timestamp` - RFC3339 timestamp when graceful deletion was requested; `null` when not being deleted.
- `deletion_grace_period_seconds` - Seconds allowed for graceful termination before removal from the system.
- `labels` - Map of string key-value labels attached to the object.
- `annotations` - Map of string key-value annotations attached to the object.
- `finalizers` - Set of finalizer strings that must be empty before the object is deleted.
- `owner_references` - Set of objects that own this KubernetesRelease.
  - `api_version` - API version of the owner object.
  - `kind` - Kind of the owner object.
  - `name` - Name of the owner object.
  - `uid` - UID of the owner object.
  - `controller` - Whether this owner is the managing controller.
  - `block_owner_deletion` - Whether deletion of the owner is blocked until this object is also deleted.

## Kubernetes

The `kubernetes` attribute has the following structure:

- `version` - Semantic versioning conformant version of the Kubernetes build shipped by this release.
- `image_repository` - Container image registry used to pull Kubernetes component images.
- `etcd` - Container image details for etcd. See [Container Image Info](#container-image-info).
- `pause` - Container image details for pause. See [Container Image Info](#container-image-info).
- `coredns` - Container image details for CoreDNS. See [Container Image Info](#container-image-info).
- `kube_vip` - Container image details for kube-vip. See [Container Image Info](#container-image-info).

## Container Image Info

The `etcd`, `pause`, `coredns`, and `kube_vip` attributes share the following structure:

- `image_repository` - Container image registry to pull images from. When empty, defaults to the `kubernetes.image_repository`.
- `image_tag` - Container image tag.

## Status

The `status` attribute has the following structure:

- `conditions` - Set of conditions reported by the controller. See [Conditions](#conditions).

## Conditions

The `status.conditions` attribute is a set of entries with the following structure:

- `type` - Condition type (e.g. `Ready`).
- `status` - Condition status: `True`, `False`, or `Unknown`.
- `last_transition_time` - RFC3339 timestamp of the last status transition.
- `reason` - Machine-readable reason for the condition.
- `message` - Human-readable message describing the condition.
