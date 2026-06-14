// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vmware/terraform-provider-vcfa/internal/vcfatypes"
)

// vksOsImageAnnotationConflictValidator is a validator.Object that rejects any
// topology block that specifies os_image AND also manually sets the resolve-os-image
// annotation key in metadata.annotations.
type vksOsImageAnnotationConflictValidator struct{}

// VksOsImageAnnotationConflict returns a validator.Object that emits a plan-time
// error when os_image and the resolve-os-image annotation key are both present on the same
// topology block (control_plane or machine_deployments entry).
func VksOsImageAnnotationConflict() validator.Object {
	return vksOsImageAnnotationConflictValidator{}
}

func (v vksOsImageAnnotationConflictValidator) Description(_ context.Context) string {
	return fmt.Sprintf(
		"os_image and the annotation %q in metadata.annotations are mutually exclusive",
		vcfatypes.OsImageAnnotationKey,
	)
}

func (v vksOsImageAnnotationConflictValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v vksOsImageAnnotationConflictValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()

	osImageAttr, ok := attrs["os_image"]
	if !ok || osImageAttr.IsNull() || osImageAttr.IsUnknown() {
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

	if _, conflict := annotationsMap.Elements()[vcfatypes.OsImageAnnotationKey]; conflict {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Conflicting os_image and metadata annotation",
			fmt.Sprintf(
				"\"os_image\" and the annotation %q in \"metadata.annotations\" are mutually exclusive. "+
					"Remove the annotation from \"metadata.annotations\" and use \"os_image\" instead.",
				vcfatypes.OsImageAnnotationKey,
			),
		)
	}
}
