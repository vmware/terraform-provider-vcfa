---
page_title: "VMware Cloud Foundation Automation: vcfa_vks_cluster_class"
subcategory: ""
description: |-
  Provides a data source to read a VKS Cluster Class from VMware Cloud Foundation Automation.
---

# vcfa_vks_cluster_class

Provides a data source to read a VKS `ClusterClass` resource from VMware Cloud Foundation Automation.

A `ClusterClass` is a reusable template that defines the infrastructure, control plane, and worker node topology for creating VKS clusters via ClusterAPI's Managed Topology feature.

_Used by: **Tenant**_

## Example Usage

```hcl
# Read a system-wide VKS ClusterClass from the system namespace
data "vcfa_vks_cluster_class" "builtin" {
  context = {
    project   = "my-project"
    namespace = "my-namespace"
  }

  name   = "builtin-generic-v3.7.0"
  system = true
}

# Read a private VKS ClusterClass from a specific namespace
data "vcfa_vks_cluster_class" "custom" {
  context = {
    project   = "my-project"
    namespace = "my-namespace"
  }

  name = "my-custom-class"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the VKS ClusterClass to read.
- `context` - (Required) VCF Automation context required to look up this resource. See [Context](#context).
- `system` - (Optional, Computed) When `true`, the VKS ClusterClass is read from the system-wide public namespace (`vmware-system-vks-public`).

## Context

The `context` block contains the following required attributes:

- `project` - (Required) Name of the Project where the resource is located.
- `namespace` - (Required) Name of the Namespace where the resource is located.

## Attribute Reference

In addition to the arguments above, the following computed attributes are exported:

- `id` - Internal identifier.
- `metadata` - Standard Kubernetes object metadata. See [Metadata](#metadata).
- `availability_gates` - Set of additional conditions evaluated when determining the Cluster `Available` condition. See [Availability Gates](#availability-gates).
- `infrastructure` - Provider-specific infrastructure cluster template definition. See [Infrastructure](#infrastructure).
- `control_plane` - Control plane class definition. See [Control Plane](#control-plane).
- `workers` - Workers class configuration, containing MachineDeployment class definitions. See [Workers](#workers).
- `variables` - Set of variables that can be configured in a Cluster topology and used in patches. See [Variables](#variables).
- `patches` - Set of patches applied to customise referenced templates. See [Patches](#patches).
- `upgrade` - Upgrade configuration for clusters using this ClusterClass. See [Upgrade](#upgrade).
- `kubernetes_versions` - Ordered list of Kubernetes versions supported by this ClusterClass (oldest to newest).
- `status` - Observed state of the ClusterClass as reported by the controller. See [Status](#status).

## Metadata

The `metadata` attribute exposes the standard Kubernetes object metadata read from the ClusterClass:

- `name` - Name of the object.
- `generate_name` - Optional server-side prefix used to generate a unique name when `name` is not provided.
- `namespace` - Namespace of the object.
- `uid` - Universally unique identifier assigned by the server at creation time.
- `resource_version` - Opaque string used to detect object changes; treat as opaque.
- `generation` - Monotonically increasing sequence number for the desired state; incremented on every spec change.
- `creation_timestamp` - RFC3339 timestamp when the object was created.
- `deletion_timestamp` - RFC3339 timestamp when graceful deletion was requested; `null` when the object is not being deleted.
- `deletion_grace_period_seconds` - Seconds allowed for graceful termination before the object is removed from the system; only set when `deletion_timestamp` is also set.
- `labels` - Map of string key-value labels attached to the object.
- `annotations` - Map of string key-value annotations attached to the object.
- `finalizers` - Set of finalizer strings that must be empty before the object is deleted.
- `owner_references` - Set of objects that own this ClusterClass.
  - `api_version` - API version of the owner object.
  - `kind` - Kind of the owner object.
  - `name` - Name of the owner object.
  - `uid` - UID of the owner object.
  - `controller` - Whether this owner is the managing controller.
  - `block_owner_deletion` - Whether deletion of the owner is blocked until this object is also deleted.

## Availability Gates

The `availability_gates` attribute is a set of entries with the following structure:

- `condition_type` - Condition type in the Cluster's condition list to use as an availability gate.
- `polarity` - Polarity of the condition: `Positive` (true = healthy) or `Negative` (false = healthy).

## Infrastructure

The `infrastructure` attribute is an object with the following structure:

- `template_ref` - Reference to the infrastructure cluster template.
  - `api_version` - API version of the referenced resource.
  - `kind` - Kind of the referenced resource.
  - `name` - Name of the referenced resource.
- `naming` - Naming strategy for the generated infrastructure object.
  - `template` - Go template string used to derive the infrastructure object name.

## Control Plane

The `control_plane` attribute is an object with the following structure:

- `metadata` - Labels and annotations applied to the control plane object.
  - `labels` - Map of string key-value labels.
  - `annotations` - Map of string key-value annotations.
- `template_ref` - Reference to the control plane template.
  - `api_version` - API version of the referenced resource.
  - `kind` - Kind of the referenced resource.
  - `name` - Name of the referenced resource.
- `machine_infrastructure` - Infrastructure template for individual control plane machines (machine-based providers only).
  - `template_ref` - Reference to the machine infrastructure template. Same fields as `template_ref` above.
- `health_check` - MachineHealthCheck configuration for control plane machines.
  - `checks` - Conditions that mark a control plane machine as unhealthy.
    - `node_startup_timeout_seconds` - Maximum seconds before a node must have a `ProviderID`; `0` disables the check.
    - `unhealthy_node_conditions` - Set of node conditions that trigger unhealthy classification (logical OR).
      - `type` - Node condition type (e.g. `Ready`).
      - `status` - Required condition status (`True`, `False`, `Unknown`).
      - `timeout_seconds` - Duration the condition must persist before the machine is considered unhealthy.
    - `unhealthy_machine_conditions` - Set of machine conditions that trigger unhealthy classification (logical OR). Same fields as `unhealthy_node_conditions`.
  - `remediation` - How unhealthy control plane machines are remediated.
    - `trigger_if` - Thresholds that gate when remediation fires.
      - `unhealthy_less_than_or_equal_to` - Remediate only when unhealthy count ≤ this value (integer or percentage string, e.g. `"3"` or `"20%"`).
      - `unhealthy_in_range` - Remediate only when unhealthy count falls within this range (e.g. `"[3-5]"`).
    - `template_ref` - External remediation template; when set, delegates remediation to a custom controller.
      - `api_version` - API version of the referenced template.
      - `kind` - Kind of the referenced template.
      - `name` - Name of the referenced template.
- `naming` - Naming strategy for the control plane object.
  - `template` - Go template string.
- `deletion` - Machine deletion configuration for control plane nodes.
  - `node_drain_timeout_seconds` - Maximum seconds spent draining a node before deletion (`0` = unlimited).
  - `node_volume_detach_timeout_seconds` - Maximum seconds waiting for volume detachment (`0` = unlimited).
  - `node_deletion_timeout_seconds` - How long to retry deleting the Kubernetes Node object (`0` = indefinite).
- `taints` - Set of node taints managed by Cluster API on control plane nodes.
  - `key` - Taint key.
  - `value` - Taint value.
  - `effect` - Taint effect: `NoSchedule`, `PreferNoSchedule`, or `NoExecute`.
  - `propagation` - When the taint is propagated to Nodes: `Always` or `OnInitialization`.
- `readiness_gates` - Set of additional Machine conditions used to evaluate control plane readiness.
  - `condition_type` - Machine condition type used as a readiness gate.
  - `polarity` - `Positive` or `Negative`.

## Workers

The `workers` attribute is an object with the following structure:

- `machine_deployments` - Set of MachineDeployment class definitions. See [Machine Deployments](#machine-deployments).

### Machine Deployments

The `workers.machine_deployments` attribute is a set of entries with the following structure:

- `metadata` - Labels and annotations applied to the MachineDeployment object.
  - `labels` - Map of string key-value labels.
  - `annotations` - Map of string key-value annotations.
- `class` - Unique class name, referenceable from a Cluster topology.
- `bootstrap` - Bootstrap template.
  - `template_ref` - Reference to the bootstrap template.
    - `api_version`, `kind`, `name` - Template reference fields.
- `infrastructure` - Infrastructure template.
  - `template_ref` - Reference to the infrastructure template.
    - `api_version`, `kind`, `name` - Template reference fields.
- `health_check` - MachineHealthCheck configuration for worker machines.
  - `checks` - Conditions that mark a worker machine as unhealthy. Same structure as `control_plane.health_check.checks`.
  - `remediation` - How unhealthy worker machines are remediated.
    - `max_in_flight` - Maximum number or percentage of machines that can undergo remediation simultaneously (e.g. `"3"` or `"20%"`).
    - `trigger_if` - Thresholds that gate remediation. Same structure as `control_plane.health_check.remediation.trigger_if`.
    - `template_ref` - External remediation template. Same structure as `control_plane.health_check.remediation.template_ref`.
- `failure_domain` - Default failure domain for the machines in this class.
- `naming` - Naming strategy for the MachineDeployment object.
  - `template` - Go template string.
- `deletion` - Machine deletion configuration.
  - `order` - Order in which Machines are deleted when downscaling: `Random`, `Newest`, or `Oldest`.
  - `node_drain_timeout_seconds` - Maximum seconds spent draining a node before deletion (`0` = unlimited).
  - `node_volume_detach_timeout_seconds` - Maximum seconds waiting for volume detachment (`0` = unlimited).
  - `node_deletion_timeout_seconds` - How long to retry deleting the Kubernetes Node object (`0` = indefinite).
- `taints` - Set of node taints managed by Cluster API. Same structure as `control_plane.taints`.
- `min_ready_seconds` - Minimum seconds a new machine must be ready before it is considered available (default `0`).
- `readiness_gates` - Set of additional Machine conditions used to evaluate readiness. Same structure as `control_plane.readiness_gates`.
- `rollout` - Rollout configuration.
  - `strategy` - Rollout strategy.
    - `type` - Strategy type: `RollingUpdate` or `OnDelete`.
    - `rolling_update` - Configuration for the `RollingUpdate` strategy.
      - `max_unavailable` - Maximum number or percentage of unavailable machines during a rolling update.
      - `max_surge` - Maximum number or percentage of machines that can be scheduled above the desired count during a rolling update.

## Variables

The `variables` attribute is a set of entries with the following structure:

- `name` - Name of the variable.
- `required` - Whether the variable must be set in a topology.
- `schema` - OpenAPI v3 schema definition for the variable.
  - `open_api_v3_schema` - The schema serialised as a JSON string.

## Patches

The `patches` attribute is a set of entries with the following structure:

- `name` - Name of the patch.
- `description` - Human-readable description.
- `enabled_if` - Go template expression; the patch is enabled only when this evaluates to `"true"`. When unset the patch is always enabled.
- `definitions` - Set of inline patch definitions (mutually exclusive with `external`).
  - `selector` - Selects which templates this definition applies to.
    - `api_version` - Filters templates by API version.
    - `kind` - Filters templates by kind.
    - `match_resources` - Selects templates based on where they are referenced (results are ORed).
      - `control_plane` - Selects templates referenced in `spec.controlPlane`.
      - `infrastructure_cluster` - Selects templates referenced in `spec.infrastructure`.
      - `machine_deployment_class` - Selects templates in specific MachineDeploymentClasses.
        - `names` - Set of MachineDeploymentClass names to match.
  - `json_patches` - Set of JSON patches applied to matching templates.
    - `op` - Patch operation: `add`, `replace`, or `remove`.
    - `path` - JSON patch path (must start with `/spec/`).
    - `value` - Literal patch value serialised as a JSON string (mutually exclusive with `value_from`).
    - `value_from` - Dynamic patch value (mutually exclusive with `value`).
      - `variable` - Variable whose value is used (from `spec.variables` or ClusterAPI builtins).
      - `template` - Go template evaluated to produce the patch value.
- `external` - External patch definition delegated to a runtime extension (mutually exclusive with `definitions`).
  - `generate_patches_extension` - Name of the runtime extension called to generate patches.
  - `validate_topology_extension` - Name of the runtime extension called to validate the topology.
  - `discover_variables_extension` - Name of the runtime extension called to discover variables.
  - `settings` - Map of key-value settings passed to the extension.

## Upgrade

The `upgrade` attribute is an object with the following structure:

- `external` - External runtime extensions for upgrade operations.
  - `generate_upgrade_plan_extension` - Name of the runtime extension called to generate the upgrade plan.

## Status

The `status` attribute is an object with the following structure:

- `observed_generation` - Most recent generation of the ClusterClass observed by the controller.
- `conditions` - Set of current conditions reported by the controller (e.g. `VariablesReady`, `RefVersionsUpToDate`, `Paused`).
  - `type` - Condition type.
  - `status` - Condition status: `True`, `False`, or `Unknown`.
  - `observed_generation` - Generation that was current when this condition was last updated.
  - `last_transition_time` - RFC3339 timestamp of the last status transition.
  - `reason` - Machine-readable reason for the condition.
  - `message` - Human-readable message describing the condition.
- `variables` - Variables as resolved and observed by the controller (includes variables discovered from runtime extensions).
  - `name` - Name of the variable.
  - `definitions_conflict` - Whether multiple conflicting definitions exist for this variable name.
  - `definitions` - Set of all definitions for this variable.
    - `from` - Origin of the definition: `"inline"` for variables defined directly in the ClusterClass, or the patch name for variables discovered via runtime extensions.
    - `required` - Whether this definition marks the variable as required.
    - `schema` - OpenAPI v3 schema for this definition.
      - `open_api_v3_schema` - The schema serialised as a JSON string.
