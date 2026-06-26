---
page_title: "VMware Cloud Foundation Automation: vcfa_vks_cluster"
subcategory: ""
description: |-
  Provides a resource to manage VKS Clusters in VMware Cloud Foundation Automation.
---

# vcfa_vks_cluster

Provides a resource to manage VKS Cluster resources in VMware Cloud Foundation Automation.

_Used by: **Tenant**_

## Example Usage

### Basic cluster with fixed replica counts

```hcl
resource "vcfa_vks_cluster" "example" {
  context = {
    project   = "my-project"
    namespace = "my-namespace"
  }

  name = "my-cluster"

  labels = {
    "env"        = "production"
    "managed-by" = "terraform"
  }

  annotations = {
    "owner" = "platform-team"
  }

  cluster_class = {
    name = "builtin-generic-v3.7.0"
  }
  version = "v1.34.1+vmware.1"

  cluster_network = {
    services = {
      cidr_blocks = ["10.96.0.0/12"]
    }
  }

  variables = [
    { name = "vmClass", value = "best-effort-small" },
    { name = "storageClass", value = "development" },
  ]

  control_plane = {
    replicas = 1
  }

  machine_deployments = [
    {
      name     = "default"
      class    = "node-pool"
      replicas = 2
    }
  ]
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required, Forces new resource) Name of the VKS Cluster. Must be RFC 1123 DNS subdomain compliant.
- `context` - (Required, Forces new resource) VCF Automation context for managing this cluster; changing either field forces replacement. See [Context](#context).
- `cluster_class` - (Required, Forces new resource) Reference to the ClusterClass used by this cluster. See [Cluster Class](#cluster-class).
- `version` - (Required) Desired Kubernetes Release version for the cluster (e.g. `v1.34.1+vmware.1`).
- `cluster_network` - (Required, Forces new resource) Cluster-wide network configuration. See [Cluster Network](#cluster-network).
- `control_plane` - (Required) Topology configuration for the control plane. See [Control Plane](#control-plane).
- `variables` - (Required) Cluster-level variable values passed to ClusterClass patches. Must include at least `vmClass` and `storageClass` entries. See [Variables](#variables).
- `machine_deployments` - (Optional) Set of MachineDeployment topology entries. See [Machine Deployments](#machine-deployments).
- `labels` - (Optional) User-managed labels to set on the cluster's `ObjectMeta`. Only the keys declared here are tracked; any labels injected by the backend are silently ignored and never appear in plan diffs. Must contain at least one entry when set. See [Labels and Annotations](#labels-and-annotations).
- `annotations` - (Optional) User-managed annotations to set on the cluster's `ObjectMeta`. Only the keys declared here are tracked; any annotations injected by the backend are silently ignored and never appear in plan diffs. Must contain at least one entry when set. See [Labels and Annotations](#labels-and-annotations).
- `dry_run_validation` - (Optional) When `true`, a dry-run Create or Update request is sent to the backend during `terraform plan` and `terraform apply` to validate the cluster configuration before any changes are committed. Backend validation errors are surfaced as plan errors. Defaults to `false`.
- `wait_for` - (Optional) Controls whether create/update/delete operations block until the cluster reaches a desired state. See [Wait For](#wait-for).
- `timeouts` - (Optional) Operation timeouts. See [Timeouts](#timeouts).

-> The `version` attribute accepts both the VKS Kubernetes Release `name` and `version`. If the Kubernetes Release `name` is provided (e.g. `v1.34.1---vmware.1-vkr.4`), the backend converts it to its canonical form (e.g. `v1.34.1+vmware.1`), which will show as a diff on subsequent plans. Use the VKS Kubernetes Release `version` to avoid this, or add `version` to the [lifecycle.ignore_changes](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes) resource argument.

## Attribute Reference

In addition to the arguments above, the following computed attributes are exported:

- `id` - Internal identifier.
- `metadata` - Standard Kubernetes object metadata. See [Metadata](#metadata).
- `availability_gates` - Additional conditions evaluated when determining the cluster's `Available` condition. See [Availability Gates](#availability-gates).
- `control_plane_endpoint` - The externally reachable API server endpoint for the cluster. See [Control Plane Endpoint](#control-plane-endpoint).
- `status` - Observed state of the VKS Cluster. See [Status](#status).

## Context

The `context` block contains the following required attributes:

- `project` - (Required) Name of the Project where the resource is located.
- `namespace` - (Required) Name of the Namespace where the resource is located.

## Cluster Class

The `cluster_class` argument has the following structure:

- `name` - (Required) Name of the ClusterClass (1–253 characters; DNS subdomain format).
- `namespace` - (Optional) Namespace of the ClusterClass (1–63 characters; DNS label format).

## Cluster Network

The `cluster_network` argument has the following structure:

- `service_domain` - (Optional) DNS domain for Services inside the cluster.
- `pods` - (Optional) Pod network CIDR configuration.
  - `cidr_blocks` - (Required) Set of CIDR blocks allocated for pod IP addresses.
- `services` - (Required) Service network CIDR configuration.
  - `cidr_blocks` - (Required) Set of CIDR blocks allocated for Service VIPs.

## Control Plane

The `control_plane` argument has the following structure:

- `replicas` - (Required) Desired number of control plane nodes. Must be `1`, `3`, or `5`.
- `metadata` - (Optional) Labels and annotations merged with the ClusterClass control plane metadata at runtime.
  - `labels` - (Optional) Map of string key-value labels.
  - `annotations` - (Optional) Map of string key-value annotations.
- `os_image` - (Optional) OS image selection for control plane machines. See [OS Image](#os-image).
- `rollout` - (Optional) Rolling update configuration.
  - `after` - (Required) RFC3339 timestamp after which a rollout is triggered even with no spec changes (e.g. `"2026-06-30T03:00:00Z"`). Required when the `rollout` block is specified.
- `health_check` - (Optional) MachineHealthCheck configuration for control plane machines. See [Health Check](#health-check).
- `deletion` - (Optional) Machine deletion configuration. See [Deletion](#deletion).
- `variable_overrides` - (Optional) Set of variable overrides applied to the control plane topology element (1–1000 entries). See [Variables](#variables).
- `taints` - (Computed) Node taints on control plane nodes. See [Taints](#taints).
- `readiness_gates` - (Computed) Additional conditions included when evaluating Machine Ready on control plane nodes. See [Readiness Gates](#readiness-gates).

## Machine Deployments

The `machine_deployments` argument is a set of entries with the following structure:

- `class` - (Required) Name of the MachineDeploymentClass defined in the ClusterClass (1–256 characters).
- `name` - (Required) Unique identifier for this MachineDeployment within the cluster topology (1–63 characters; DNS subdomain format).
- `metadata` - (Optional) Labels and annotations merged with the ClusterClass MachineDeployment metadata at runtime.
  - `labels` - (Optional) Map of string key-value labels.
  - `annotations` - (Optional) Map of string key-value annotations.
- `replicas` - (Optional) Desired number of worker nodes in this deployment. Mutually exclusive with `autoscaler`. At least one of `replicas`, `autoscaler.min_size`, or `autoscaler.max_size` must be specified.
- `autoscaler` - (Optional) Cluster Autoscaler bounds for this MachineDeployment. Mutually exclusive with `replicas`.  At least one of `min_size` or `max_size` must be provided. See [Autoscaler](#autoscaler).
- `os_image` - (Optional) OS image selection for this MachineDeployment's machines. See [OS Image](#os-image).
- `failure_domain` - (Optional) Failure domain for the machines in this deployment (1–256 characters).
- `min_ready_seconds` - (Optional) Minimum seconds a Machine must be ready before it is considered available (`0` = immediate).
- `health_check` - (Optional) MachineHealthCheck configuration. See [Health Check](#health-check).
- `deletion` - (Optional) Machine deletion configuration. See [Deletion](#deletion).
- `rollout` - (Optional) Rolling update configuration.
  - `after` - (Required) RFC3339 timestamp after which a rollout is triggered even with no spec changes (e.g. `"2026-06-30T03:00:00Z"`). Required when the `rollout` block is specified.
  - `strategy` - (Optional) Rollout strategy.
    - `type` - (Required) Strategy type: `RollingUpdate` or `OnDelete`.
    - `rolling_update` - (Optional) Rolling update config; applicable when `type` is `RollingUpdate`.
      - `max_unavailable` - (Optional) Maximum unavailable machines during update (absolute number or percentage, e.g. `"5"` or `"10%"`).
      - `max_surge` - (Optional) Maximum machines that can be scheduled above the desired count (absolute number or percentage).
- `variable_overrides` - (Optional) Set of variable overrides applied to this MachineDeployment topology element (1–1000 entries). See [Variables](#variables).
- `taints` - (Computed) Node taints on this MachineDeployment's nodes. See [Taints](#taints).
- `readiness_gates` - (Computed) Additional conditions included when evaluating Machine Ready on this MachineDeployment. See [Readiness Gates](#readiness-gates).

## Variables

The `variables` argument (and `*.variable_overrides`) is a set of entries with the following structure:

- `name` - (Required) Variable name (1–256 characters).
- `value` - (Required) Variable value serialised as a JSON string.

The available variables and their nested properties depend on the VKS version. See [Variable availability across VKS versions](https://developer.broadcom.com/xapis/vmware-vsphere-kubernetes-service/latest/variable-docs.html).

## Labels and Annotations

The top-level `labels` and `annotations` arguments set arbitrary key-value metadata directly on the cluster's Kubernetes `ObjectMeta`. They implement **partial management**: only the keys you declare in Terraform configuration are tracked in state. Any labels or annotations injected by the platform or by Cluster API controllers are stored on the cluster object but are silently filtered out when Terraform reads state back, so they never produce a plan diff.

**Difference from topology `metadata`:**

The `metadata` attribute (Computed, read-only) reflects the _full_ `ObjectMeta` as stored in the API, including every backend-injected label and annotation. The top-level `labels` / `annotations` arguments are the writable counterpart and expose only the subset you control.

Similarly, the `metadata` blocks inside `control_plane` and `machine_deployments` set labels and annotations on the _topology_ objects (i.e. the `ClusterTopology` metadata that Cluster API merges with the ClusterClass at runtime), not on the cluster `ObjectMeta` itself.

## OS Image

The `os_image` argument (available on both `control_plane` and `machine_deployments` entries) selects the OS image for cluster machines. When configured, it automatically injects the `run.tanzu.vmware.com/resolve-os-image` annotation into the topology metadata; you should not set this annotation manually in `metadata.annotations` at the same time.

- `name` - (Required) OS image name (e.g. `"ubuntu"`, `"photon"`).
- `version` - (Optional) OS image version (e.g. `"22.04"`).

## Autoscaler

The `autoscaler` argument (available on `machine_deployments` entries only) configures the VKS Cluster Autoscaler node group bounds. When configured, it automatically injects the corresponding `cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size` and `cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size` annotations; you should not set these annotations manually in `metadata.annotations` at the same time.

- `min_size` - (Optional) Minimum number of nodes the autoscaler can scale down to.
- `max_size` - (Optional) Maximum number of nodes the autoscaler can scale up to.

At least one of `min_size` or `max_size` must be provided. `autoscaler` is mutually exclusive with `replicas`.

## Health Check

The `health_check` argument has the following structure:

- `enabled` - (Optional) Whether a MachineHealthCheck object is created.
- `checks` - (Optional) Conditions that classify a machine as unhealthy.
  - `node_startup_timeout_seconds` - (Optional) Maximum seconds before a Machine is considered unhealthy if its Node does not appear (`0` = disabled, default 10 minutes). When non-zero must be at least `30`.
  - `unhealthy_node_conditions` - (Optional) Set of Node conditions that trigger unhealthy classification (logical OR).
    - `type` - (Required) Node condition type.
    - `status` - (Required) Required condition status: `True`, `False`, or `Unknown`.
    - `timeout_seconds` - (Required) Duration (seconds) the node must be in this state before being deemed unhealthy.
- `remediation` - (Optional) How unhealthy machines are remediated.
  - `trigger_if` - (Optional) Thresholds that gate when remediation fires.
    - `unhealthy_less_than_or_equal_to` - (Optional) Remediate only when unhealthy count ≤ this value (absolute number or percentage, e.g. `"3"` or `"20%"`).
    - `unhealthy_in_range` - (Optional) Remediate only when unhealthy count falls within this range (e.g. `"[3-5]"`).

The `machine_deployments[*].health_check.remediation` block additionally supports:

- `max_in_flight` - (Optional) Maximum number or percentage of machines that can be simultaneously remediated.

## Deletion

The `deletion` argument has the following structure:

- `node_drain_timeout_seconds` - (Optional) Maximum seconds spent draining a node before deletion (`0` = unlimited).
- `node_volume_detach_timeout_seconds` - (Optional) Maximum seconds waiting for volume detachment (`0` = unlimited).
- `node_deletion_timeout_seconds` - (Optional) Seconds the controller tries to delete the Kubernetes Node object before giving up (`0` = retry indefinitely, default 10).

The `machine_deployments[*].deletion` block additionally supports:

- `order` - (Optional) Order in which Machines are deleted when downscaling: `Random`, `Newest`, or `Oldest` (default: `Random`).

## Taints

The `*.taints` attribute is a set of entries with the following structure:

- `key` - Taint key.
- `value` - Taint value.
- `effect` - Taint effect: `NoSchedule`, `PreferNoSchedule`, or `NoExecute`.
- `propagation` - When the taint is propagated to Nodes: `Always` or `OnInitialization`.

## Readiness Gates

The `*.readiness_gates` attribute is a set of entries with the following structure:

- `condition_type` - Condition type.
- `polarity` - `Positive` or `Negative`.

## Wait For

The `wait_for` argument has the following structure:

- `available` - (Optional) When `true`, Create and Update operations block until the cluster's `Available` condition is `True`. Set to `false` (default) to return immediately after the API call.
- `deleted` - (Optional) When `true`, Delete operation blocks until the cluster is fully removed. Set to `false` (default) to return immediately after the delete API call.

## Timeouts

The `timeouts` block allows you to specify timeouts for certain actions:

- `create` - (Default `30m`) How long to wait for a Cluster to be available during a Create operation. Only applicable when the `wait_for.available` attribute is set to `true`.
- `update` - (Default `30m`) How long to wait for a Cluster to be available during an Update operation. Only applicable when the `wait_for.available` attribute is set to `true`.
- `delete` - (Default `10m`) How long to wait for a Cluster to be deleted. Only applicable when the `wait_for.deleted` attribute is set to `true`.

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
- `owner_references` - Set of objects that own this Cluster.
  - `api_version` - API version of the owner object.
  - `kind` - Kind of the owner object.
  - `name` - Name of the owner object.
  - `uid` - UID of the owner object.
  - `controller` - Whether this owner is the managing controller.
  - `block_owner_deletion` - Whether deletion of the owner is blocked until this object is also deleted.

## Availability Gates

The `availability_gates` attribute is a set of entries with the following structure:

- `condition_type` - Condition type in the Cluster's condition list used as an availability gate.
- `polarity` - Polarity of the condition: `Positive` (true = healthy) or `Negative` (false = healthy).

## Control Plane Endpoint

The `control_plane_endpoint` attribute has the following structure:

- `host` - Hostname or IP address of the Kubernetes API server.
- `port` - TCP port of the Kubernetes API server.

## Status

The `status` attribute has the following structure:

- `phase` - Current lifecycle phase of the cluster: `Pending`, `Provisioning`, `Provisioned`, `Deleting`, `Failed`, or `Unknown`.
- `observed_generation` - Most recent generation of the Cluster spec observed by the controller.
- `initialization` - One-time initialisation milestones.
  - `infrastructure_provisioned` - `true` once the infrastructure has been fully provisioned.
  - `control_plane_initialized` - `true` once the control plane is functional enough to accept requests.
- `control_plane` - Replica counts for the control plane.
  - `desired_replicas` - Total desired control plane machines.
  - `replicas` - Total control plane machines, including those being provisioned or deleted.
  - `up_to_date_replicas` - Control plane machines running the latest spec.
  - `ready_replicas` - Control plane machines in the `Ready` state.
  - `available_replicas` - Control plane machines that have been ready for at least `minReadySeconds`.
- `workers` - Aggregate replica counts across all worker MachineDeployments.
  - `desired_replicas` - Total desired worker machines.
  - `replicas` - Total worker machines, including those being provisioned or deleted.
  - `up_to_date_replicas` - Worker machines running the latest spec.
  - `ready_replicas` - Worker machines in the `Ready` state.
  - `available_replicas` - Worker machines that have been ready for at least `minReadySeconds`.
- `failure_domains` - Failure domains discovered from the infrastructure provider and available for scheduling.
  - `name` - Name of the failure domain.
  - `control_plane` - Whether this failure domain is suitable for control plane machines.
  - `attributes` - Map of free-form key-value attributes provided by the infrastructure provider.
- `conditions` - Set of conditions reported by the ClusterAPI controller.
  - `type` - Condition type (e.g. `Available`, `TopologyReconciled`).
  - `status` - Condition status: `True`, `False`, or `Unknown`.
  - `observed_generation` - Generation that was current when this condition was last updated.
  - `last_transition_time` - RFC3339 timestamp of the last status transition.
  - `reason` - Machine-readable reason for the condition.
  - `message` - Human-readable message describing the condition.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows also code generation. See [Importing resources][importing-resources] for more information.

An existing VKS Cluster can be [imported][docs-import] into this resource via its composite identifier.
For example, using this structure, representing an existing VKS Cluster that was **not** created using Terraform:

```hcl
resource "vcfa_vks_cluster" "existing" {
  context = {
    project   = "my-project"
    namespace = "my-namespace"
  }

  name = "my-cluster"

  cluster_class = {
    name = "builtin-generic-v3.7.0"
  }
  version = "v1.34.1+vmware.1"

  cluster_network = {
    services = {
      cidr_blocks = ["10.96.0.0/12"]
    }
  }

  variables = [
    { name = "vmClass", value = "best-effort-small" },
    { name = "storageClass", value = "development" },
  ]

  control_plane = {
    replicas = 1
  }

  machine_deployments = [
    {
      name     = "default"
      class    = "node-pool"
      replicas = 2
    }
  ]
}
```

You can import such VKS Cluster into terraform state using this command:

```shell
terraform import vcfa_vks_cluster.existing "my-project.my-namespace.my-cluster"
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

[docs-import]: https://developer.hashicorp.com/terraform/cli/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
