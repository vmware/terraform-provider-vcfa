// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ExtractStringMap converts an Optional types.Map into a map[string]string suitable
// for use in a Kubernetes ObjectMeta.  Returns nil when the attribute is null or unknown.
func ExtractStringMap(ctx context.Context, m types.Map, diags *diag.Diagnostics) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	result := make(map[string]string, len(m.Elements()))
	diags.Append(m.ElementsAs(ctx, &result, false)...)
	return result
}

// ObjFrom is a thin wrapper around types.ObjectValueFrom that appends any diagnostics
// to the provided diag.Diagnostics rather than returning them, keeping call sites concise.
func ObjFrom(ctx context.Context, attrTypes map[string]attr.Type, val any, diags *diag.Diagnostics) types.Object {
	obj, d := types.ObjectValueFrom(ctx, attrTypes, val)
	diags.Append(d...)
	return obj
}

// SetFrom is a thin wrapper around types.SetValueFrom that appends any diagnostics
// to the provided diag.Diagnostics rather than returning them, keeping call sites concise.
func SetFrom(ctx context.Context, elemType attr.Type, val any, diags *diag.Diagnostics) types.Set {
	set, d := types.SetValueFrom(ctx, elemType, val)
	diags.Append(d...)
	return set
}

// SanitizeUnknownForState walks every Terraform Plugin Framework value embedded in the model
// struct tree and replaces any "unknown" value with its null equivalent. Terraform rejects
// unknown state values after apply, so this must be called before resp.State.Set when the
// plan is stored directly as state (i.e. in Create).
//
// For scalar types (Bool, Int32, Int64, String, Map) the struct field is matched directly.
// For types.Object and types.Set the function also recurses *inside* the opaque framework
// value using tftypes so that unknown attributes at any nesting depth are nullified —
// including unknowns inside a known (user-configured) object.
func SanitizeUnknownForState(ctx context.Context, v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			SanitizeUnknownForState(ctx, v.Elem())
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			SanitizeUnknownForState(ctx, v.Index(i))
		}
	case reflect.Struct:
		if !v.CanAddr() {
			return
		}
		p := v.Addr().Interface()
		switch m := p.(type) {
		case *types.Bool:
			if m.IsUnknown() {
				*m = types.BoolNull()
			}
		case *types.Int32:
			if m.IsUnknown() {
				*m = types.Int32Null()
			}
		case *types.Int64:
			if m.IsUnknown() {
				*m = types.Int64Null()
			}
		case *types.String:
			if m.IsUnknown() {
				*m = types.StringNull()
			}
		case *types.Map:
			if m.IsUnknown() {
				*m = types.MapNull(m.ElementType(ctx))
			}
		case *types.List:
			if m.IsUnknown() {
				*m = types.ListNull(m.ElementType(ctx))
			} else if !m.IsNull() {
				tfVal, err := m.ToTerraformValue(ctx)
				if err == nil {
					nullified := NullifyUnknownsInTftypesValue(tfVal)
					if !nullified.Equal(tfVal) {
						newList, err2 := types.ListType{ElemType: m.ElementType(ctx)}.ValueFromTerraform(ctx, nullified)
						if err2 == nil {
							if l, ok := newList.(types.List); ok {
								*m = l
							}
						}
					}
				}
			}
		case *types.Set:
			if m.IsUnknown() {
				*m = types.SetNull(m.ElementType(ctx))
			} else if !m.IsNull() {
				// Known set: nullify any unknowns inside elements (e.g. nested objects).
				tfVal, err := m.ToTerraformValue(ctx)
				if err == nil {
					nullified := NullifyUnknownsInTftypesValue(tfVal)
					if !nullified.Equal(tfVal) {
						newSet, err2 := types.SetType{ElemType: m.ElementType(ctx)}.ValueFromTerraform(ctx, nullified)
						if err2 == nil {
							if s, ok := newSet.(types.Set); ok {
								*m = s
							}
						}
					}
				}
			}
		case *types.Object:
			if m.IsUnknown() {
				*m = types.ObjectNull(m.AttributeTypes(ctx))
			} else if !m.IsNull() {
				// Known object: nullify any unknown inner attributes at any depth.
				tfVal, err := m.ToTerraformValue(ctx)
				if err == nil {
					nullified := NullifyUnknownsInTftypesValue(tfVal)
					if !nullified.Equal(tfVal) {
						newObj, err2 := types.ObjectType{AttrTypes: m.AttributeTypes(ctx)}.ValueFromTerraform(ctx, nullified)
						if err2 == nil {
							if obj, ok := newObj.(types.Object); ok {
								*m = obj
							}
						}
					}
				}
			}
		default:
			for i := 0; i < v.NumField(); i++ {
				SanitizeUnknownForState(ctx, v.Field(i))
			}
		}
	}
}

// NullifyUnknownsInTftypesValue recursively replaces every unknown value in the tftypes
// value tree with a null of the same type. Known scalars, null values, and known non-null
// composites are returned unchanged (or with their children fixed).
func NullifyUnknownsInTftypesValue(v tftypes.Value) tftypes.Value {
	if !v.IsKnown() {
		return tftypes.NewValue(v.Type(), nil)
	}
	if v.IsNull() {
		return v
	}

	typ := v.Type()
	switch {
	case typ.Is(tftypes.Object{}):
		attrs := map[string]tftypes.Value{}
		if err := v.As(&attrs); err != nil {
			return v
		}
		nullified := make(map[string]tftypes.Value, len(attrs))
		changed := false
		for k, av := range attrs {
			n := NullifyUnknownsInTftypesValue(av)
			nullified[k] = n
			if !n.Equal(av) {
				changed = true
			}
		}
		if !changed {
			return v
		}
		return tftypes.NewValue(typ, nullified)

	case typ.Is(tftypes.Set{}):
		var elems []tftypes.Value
		if err := v.As(&elems); err != nil {
			return v
		}
		nullified := make([]tftypes.Value, len(elems))
		changed := false
		for i, e := range elems {
			n := NullifyUnknownsInTftypesValue(e)
			nullified[i] = n
			if !n.Equal(e) {
				changed = true
			}
		}
		if !changed {
			return v
		}
		return tftypes.NewValue(typ, nullified)

	case typ.Is(tftypes.List{}):
		var elems []tftypes.Value
		if err := v.As(&elems); err != nil {
			return v
		}
		nullified := make([]tftypes.Value, len(elems))
		changed := false
		for i, e := range elems {
			n := NullifyUnknownsInTftypesValue(e)
			nullified[i] = n
			if !n.Equal(e) {
				changed = true
			}
		}
		if !changed {
			return v
		}
		return tftypes.NewValue(typ, nullified)

	default:
		return v
	}
}
