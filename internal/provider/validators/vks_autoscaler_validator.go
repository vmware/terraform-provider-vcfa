// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

// ── VksMachineDeploymentHasScaling ───────────────────────────────────────────

// vksMachineDeploymentHasScalingValidator is a validator.Object that ensures
// every machine_deployments entry specifies at least one of: replicas,
// autoscaler.min_size, or autoscaler.max_size.
type vksMachineDeploymentHasScalingValidator struct{}

// VksMachineDeploymentHasScaling returns a validator.Object that emits a
// plan-time error when a machine_deployments entry specifies neither replicas
// nor any autoscaler bound.
func VksMachineDeploymentHasScaling() validator.Object {
	return vksMachineDeploymentHasScalingValidator{}
}

func (v vksMachineDeploymentHasScalingValidator) Description(_ context.Context) string {
	return "each machine_deployments entry must set \"replicas\" or at least one of \"autoscaler.min_size\" / \"autoscaler.max_size\""
}

func (v vksMachineDeploymentHasScalingValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v vksMachineDeploymentHasScalingValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()

	replicasAttr, replicasOk := attrs["replicas"]
	// Non-null means the attribute is configured — the value may be unknown at
	// validate time when set via a variable reference (e.g. var.x), but it will
	// be resolved before apply. Treat non-null as "user has configured this".
	replicasPresent := replicasOk && !replicasAttr.IsNull()

	// autoscalerHasBounds is true when at least one of min_size / max_size is
	// configured inside the autoscaler block. An empty autoscaler = {} with
	// neither bound set is treated the same as no autoscaler block at all.
	autoscalerHasBounds := autoscalerBoundsPresent(attrs)

	if replicasPresent && autoscalerHasBounds {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Conflicting scaling configuration",
			"\"replicas\" and \"autoscaler.min_size\" / \"autoscaler.max_size\" are mutually exclusive — use one or the other, not both.",
		)
		return
	}

	if replicasPresent || autoscalerHasBounds {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Missing scaling configuration",
		"Each \"machine_deployments\" entry must specify either \"replicas\" or at least one of \"autoscaler.min_size\" / \"autoscaler.max_size\".",
	)
}

// autoscalerBoundsPresent returns true when the autoscaler block exists and at
// least one of min_size / max_size is non-null. If the autoscaler block itself
// is unknown (e.g. from a module output) it is treated as having bounds so that
// the conflict/missing check is deferred rather than incorrectly erroring.
func autoscalerBoundsPresent(attrs map[string]attr.Value) bool {
	autoscalerAttr, ok := attrs["autoscaler"]
	if !ok || autoscalerAttr.IsNull() {
		return false
	}
	if autoscalerAttr.IsUnknown() {
		return true
	}
	autoscalerObj, ok := autoscalerAttr.(types.Object)
	if !ok {
		return false
	}
	autoscalerAttrs := autoscalerObj.Attributes()
	minSize, minOk := autoscalerAttrs["min_size"]
	maxSize, maxOk := autoscalerAttrs["max_size"]
	return (minOk && !minSize.IsNull()) || (maxOk && !maxSize.IsNull())
}

// ── VksAutoscalerAnnotationConflict ──────────────────────────────────────────

// vksAutoscalerAnnotationConflictValidator is a validator.Object that rejects
// any machine_deployments entry that specifies the autoscaler block AND also
// manually sets the corresponding autoscaler annotation keys in
// metadata.annotations.
type vksAutoscalerAnnotationConflictValidator struct{}

// VksAutoscalerAnnotationConflict returns a validator.Object that emits a
// plan-time error when both the autoscaler block and either of the
// autoscaler annotation keys are present on the same machine_deployments entry.
func VksAutoscalerAnnotationConflict() validator.Object {
	return vksAutoscalerAnnotationConflictValidator{}
}

func (v vksAutoscalerAnnotationConflictValidator) Description(_ context.Context) string {
	return fmt.Sprintf(
		"autoscaler and the annotations %q / %q in metadata.annotations are mutually exclusive",
		vcfatypes.AutoscalerMinSizeAnnotationKey, vcfatypes.AutoscalerMaxSizeAnnotationKey,
	)
}

func (v vksAutoscalerAnnotationConflictValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v vksAutoscalerAnnotationConflictValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()

	autoscalerAttr, ok := attrs["autoscaler"]
	if !ok || autoscalerAttr.IsNull() || autoscalerAttr.IsUnknown() {
		return
	}

	metadataAttr, ok := attrs["metadata"]
	if !ok || metadataAttr.IsNull() || metadataAttr.IsUnknown() {
		return
	}

	metaObj, ok := metadataAttr.(types.Object)
	if !ok {
		return
	}

	annotationsAttr, ok := metaObj.Attributes()["annotations"]
	if !ok || annotationsAttr.IsNull() || annotationsAttr.IsUnknown() {
		return
	}

	annotationsMap, ok := annotationsAttr.(types.Map)
	if !ok {
		return
	}

	elements := annotationsMap.Elements()
	var conflicting []string
	if _, found := elements[vcfatypes.AutoscalerMinSizeAnnotationKey]; found {
		conflicting = append(conflicting, fmt.Sprintf("%q", vcfatypes.AutoscalerMinSizeAnnotationKey))
	}
	if _, found := elements[vcfatypes.AutoscalerMaxSizeAnnotationKey]; found {
		conflicting = append(conflicting, fmt.Sprintf("%q", vcfatypes.AutoscalerMaxSizeAnnotationKey))
	}

	if len(conflicting) > 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Conflicting autoscaler and metadata annotations",
			fmt.Sprintf(
				"\"autoscaler\" and the annotation(s) %s in \"metadata.annotations\" are mutually exclusive. "+
					"Remove the annotation(s) from \"metadata.annotations\" and use \"autoscaler\" instead.",
				strings.Join(conflicting, ", "),
			),
		)
	}
}
