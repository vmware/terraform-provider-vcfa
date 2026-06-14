---
page_title: "VMware Cloud Foundation Automation: vcfa_vks_cluster"
subcategory: ""
description: |-
  Provides a data source to read a VKS Cluster from VMware Cloud Foundation Automation.
---

# vcfa_vks_cluster

Provides a data source to read a VKS Cluster resource from VMware Cloud Foundation Automation.

_Used by: **Tenant**_

## Example Usage

```hcl
data "vcfa_vks_cluster" "vks_cluster" {
  context = {
    project   = "my-project"
    namespace = "my-namespace"
  }

  name = "my-cluster"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the VKS Cluster to read.
- `context` - (Required) VCF Automation context required to look up this resource. See [Context](#context).

## Context

The `context` block contains the following required attributes:

- `project` - (Required) Name of the Project where the resource is located.
- `namespace` - (Required) Name of the Namespace where the resource is located.

## Attribute Reference

In addition to the arguments above, the following computed attributes are exported:

- `id` - Internal identifier.
- `metadata` - Standard Kubernetes object metadata. See [Metadata](#metadata).
- `cluster_class` - Reference to the ClusterClass used by this cluster. See [Cluster Class](#cluster-class).
- `version` - Desired Kubernetes Release version for the cluster.
- `availability_gates` - Additional conditions evaluated when determining the cluster's `Available` condition. See [Availability Gates](#availability-gates).
- `cluster_network` - Cluster-wide network configuration. See [Cluster Network](#cluster-network).
- `control_plane` - Topology configuration for the control plane. See [Control Plane](#control-plane).
- `control_plane_endpoint` - The externally reachable API server endpoint for the cluster. See [Control Plane Endpoint](#control-plane-endpoint).
- `machine_deployments` - Set of MachineDeployment topology entries. See [Machine Deployments](#machine-deployments).
- `variables` - Cluster-level variable values passed to ClusterClass patches. See [Variables](#variables).
- `status` - Observed state of the VKS Cluster. See [Status](#status).

## Cluster Class

The `cluster_class` attribute has the following structure:

- `name` - Name of the ClusterClass.
- `namespace` - Namespace of the ClusterClass.

## Cluster Network

The `cluster_network` attribute has the following structure:

- `service_domain` - DNS domain for Services inside the cluster (default: `cluster.local`).
- `pods` - Pod network CIDR configuration.
  - `cidr_blocks` - Set of CIDR blocks allocated for pod IP addresses.
- `services` - Service network CIDR configuration.
  - `cidr_blocks` - Set of CIDR blocks allocated for Service VIPs.

## Control Plane Endpoint

The `control_plane_endpoint` attribute has the following structure:

- `host` - Hostname or IP address of the Kubernetes API server.
- `port` - TCP port of the Kubernetes API server.

## Availability Gates

The `availability_gates` attribute is a set of entries with the following structure:

- `condition_type` - Condition type in the Cluster's condition list used as an availability gate.
- `polarity` - Polarity of the condition: `Positive` (true = healthy) or `Negative` (false = healthy).

## Control Plane

The `control_plane` attribute has the following structure:

- `metadata` - Labels and annotations merged with the ClusterClass control plane metadata at runtime.
  - `labels` - Map of string key-value labels.
  - `annotations` - Map of string key-value annotations.
- `replicas` - Desired number of control plane nodes.
- `os_image` - OS image selection for control plane machines, decoded from the `run.tanzu.vmware.com/resolve-os-image` annotation.
  - `name` - OS image name (e.g. `"ubuntu"`).
  - `version` - OS image version (e.g. `"22.04"`).
- `rollout` - Rolling update configuration.
  - `after` - RFC3339 timestamp after which a rollout is triggered even with no spec changes.
- `health_check` - MachineHealthCheck configuration for control plane machines. See [Health Check](#health-check).
- `deletion` - Machine deletion configuration for control plane nodes. See [Deletion](#deletion).
- `variable_overrides` - Set of variable overrides applied to the control plane topology element. See [Variables](#variables).
- `taints` - Node taints managed by Cluster API on control plane nodes. See [Taints](#taints).
- `readiness_gates` - Additional conditions included when evaluating Machine Ready on control plane nodes. See [Readiness Gates](#readiness-gates).

## Machine Deployments

The `machine_deployments` attribute is a set of entries with the following structure:

- `class` - Name of the MachineDeploymentClass defined in the ClusterClass.
- `name` - Unique identifier for this MachineDeployment within the cluster topology.
- `metadata` - Labels and annotations merged with the ClusterClass MachineDeployment metadata at runtime.
  - `labels` - Map of string key-value labels.
  - `annotations` - Map of string key-value annotations.
- `failure_domain` - Failure domain for the machines in this deployment.
- `replicas` - Desired number of worker nodes in this deployment.
- `os_image` - OS image selection for this MachineDeployment's machines, decoded from the `run.tanzu.vmware.com/resolve-os-image` annotation.
  - `name` - OS image name (e.g. `"ubuntu"`).
  - `version` - OS image version (e.g. `"22.04"`).
- `autoscaler` - Cluster Autoscaler bounds decoded from the `cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size` and `cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size` annotations.
  - `min_size` - Minimum number of nodes the autoscaler can scale down to.
  - `max_size` - Maximum number of nodes the autoscaler can scale up to.
- `min_ready_seconds` - Minimum seconds a Machine must be ready before it is considered available.
- `health_check` - MachineHealthCheck configuration. See [Health Check](#health-check).
- `deletion` - Machine deletion configuration. See [Deletion](#deletion).
- `rollout` - Rolling update configuration.
  - `after` - RFC3339 timestamp after which a rollout is triggered even with no spec changes.
  - `strategy` - Rollout strategy.
    - `type` - Strategy type: `RollingUpdate` or `OnDelete`.
    - `rolling_update` - Configuration for the `RollingUpdate` strategy.
      - `max_unavailable` - Maximum unavailable machines during a rolling update (absolute number or percentage).
      - `max_surge` - Maximum machines that can be scheduled above the desired count (absolute number or percentage).
- `variable_overrides` - Set of variable overrides applied to this MachineDeployment topology element. See [Variables](#variables).
- `taints` - Node taints managed by Cluster API on this MachineDeployment's nodes. See [Taints](#taints).
- `readiness_gates` - Additional conditions included when evaluating Machine Ready on this MachineDeployment. See [Readiness Gates](#readiness-gates).

## Variables

The `variables` attribute (and `*.variable_overrides`) is a set of entries with the following structure:

- `name` - Variable name.
- `value` - Variable value serialised as a JSON string.

## Health Check

The `health_check` attribute has the following structure:

- `enabled` - Whether a MachineHealthCheck object is created.
- `checks` - Conditions that classify a machine as unhealthy.
  - `node_startup_timeout_seconds` - Maximum seconds before a Machine is considered unhealthy if its Node does not appear (`0` = disabled).
  - `unhealthy_node_conditions` - Set of Node conditions that trigger unhealthy classification (logical OR).
    - `type` - Node condition type.
    - `status` - Required condition status: `True`, `False`, or `Unknown`.
    - `timeout_seconds` - Duration (seconds) the node must be in this state before being deemed unhealthy.
- `remediation` - How unhealthy machines are remediated.
  - `trigger_if` - Thresholds that gate when remediation fires.
    - `unhealthy_less_than_or_equal_to` - Remediate only when unhealthy count ≤ this value (absolute number or percentage, e.g. `"3"` or `"20%"`).
    - `unhealthy_in_range` - Remediate only when unhealthy count falls within this range (e.g. `"[3-5]"`).

The `machine_deployments[*].health_check.remediation` block additionally contains:

- `max_in_flight` - Maximum number or percentage of machines that can be simultaneously remediated.

## Deletion

The `deletion` attribute has the following structure:

- `node_drain_timeout_seconds` - Maximum seconds spent draining a node before deletion (`0` = unlimited).
- `node_volume_detach_timeout_seconds` - Maximum seconds waiting for volume detachment (`0` = unlimited).
- `node_deletion_timeout_seconds` - Seconds the controller tries to delete the Kubernetes Node object before giving up (`0` = retry indefinitely).

The `machine_deployments[*].deletion` block additionally contains:

- `order` - Order in which Machines are deleted when downscaling: `Random`, `Newest`, or `Oldest`.

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
