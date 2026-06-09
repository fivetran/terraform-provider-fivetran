package framework

import (
	"context"
	"reflect"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DynamicToMap converts a types.Dynamic holding an object value to map[string]interface{}.
// Returns nil for null or unknown values. Empty strings are passed through as-is.
// Diagnostics are returned rather than panicking on unexpected types.
func DynamicToMap(ctx context.Context, dyn types.Dynamic) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	if dyn.IsNull() || dyn.IsUnknown() {
		return nil, diags
	}
	underlying := dyn.UnderlyingValue()
	if underlying == nil || underlying.IsNull() || underlying.IsUnknown() {
		return nil, diags
	}
	result := attrToInterface(ctx, underlying)
	if result == nil {
		return nil, diags
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		diags.AddError("Dynamic config is not an object", "Expected an object value for dynamic config.")
		return nil, diags
	}
	return m, diags
}

func attrToInterface(ctx context.Context, val attr.Value) interface{} {
	if val == nil || val.IsNull() || val.IsUnknown() {
		return nil
	}
	switch v := val.(type) {
	case types.String:
		return v.ValueString()
	case types.Bool:
		return v.ValueBool()
	case types.Int64:
		return v.ValueInt64()
	case types.Float64:
		return v.ValueFloat64()
	case types.Object:
		return convertObjectAttrs(ctx, v.Attributes())
	case types.Map:
		return convertObjectAttrs(ctx, v.Elements())
	case types.List:
		return convertSliceElems(ctx, v.Elements())
	case types.Set:
		return convertSliceElems(ctx, v.Elements())
	}
	return nil
}

func convertObjectAttrs(ctx context.Context, elems map[string]attr.Value) map[string]interface{} {
	result := make(map[string]interface{}, len(elems))
	for k, v := range elems {
		result[k] = attrToInterface(ctx, v)
	}
	return result
}

func convertSliceElems(ctx context.Context, elems []attr.Value) []interface{} {
	result := make([]interface{}, len(elems))
	for i, v := range elems {
		result[i] = attrToInterface(ctx, v)
	}
	return result
}

// project filters an API read-back (remote) down to the keys the user manages (mask/plan),
// applying per-field metadata rules. slot is the metadata Property node for the dynamic field
// (e.g. &meta.Config or &meta.Auth); pass nil for v1 fallback (pass-through, no filtering).
//
// Per-field rules:
//   - format=="password": preserve local (mask) value — never overwrite with "****"
//   - readonly: store remote value so user can reference it; no drift tracking
//   - key in mask but absent from remote: set nil to surface drift
//   - nested object: recurse
//   - normal: take remote value
func project(remote, mask map[string]interface{}, slot *metadata.Property) map[string]interface{} {
	result := make(map[string]interface{}, len(mask))

	for key, maskVal := range mask {
		prop := slotProp(slot, key)

		if prop != nil && prop.Format == "password" {
			result[key] = maskVal
			continue
		}

		remoteVal, inRemote := remote[key]
		if !inRemote {
			result[key] = nil
			continue
		}

		if prop != nil && prop.Readonly {
			result[key] = remoteVal
			continue
		}

		if nestedMask, ok := maskVal.(map[string]interface{}); ok {
			if nestedRemote, ok2 := remoteVal.(map[string]interface{}); ok2 {
				var nestedSlot *metadata.Property
				if prop != nil {
					nestedSlot = prop
				}
				result[key] = project(nestedRemote, nestedMask, nestedSlot)
				continue
			}
		}

		result[key] = remoteVal
	}

	return result
}

// PrepareConfigPatchDynamic builds the minimal PATCH payload for an Update.
// slot is the metadata Property node for the dynamic field (e.g. &meta.Config or &meta.Auth);
// pass nil when metadata is unavailable (changed-field diff only, no nullable/immutable rules).
//
// Rules:
//   - readonly and immutable fields are never sent
//   - unchanged fields are omitted
//   - empty string "" is sent as-is, never coerced to nil
//   - field removed from plan + nullable: sent as nil (JSON null clears on server)
//   - field removed from plan + non-nullable or no metadata: omitted
func PrepareConfigPatchDynamic(plan, state map[string]interface{}, slot *metadata.Property) map[string]interface{} {
	patch := make(map[string]interface{})

	for k, planVal := range plan {
		prop := slotProp(slot, k)
		if prop != nil && (prop.Readonly || prop.Immutable) {
			continue
		}
		stateVal, inState := state[k]
		if !inState || !reflect.DeepEqual(planVal, stateVal) {
			patch[k] = planVal
		}
	}

	for k := range state {
		if _, inPlan := plan[k]; inPlan {
			continue
		}
		prop := slotProp(slot, k)
		if prop != nil && (prop.Readonly || prop.Immutable) {
			continue
		}
		if prop != nil && prop.Nullable {
			patch[k] = nil
		}
	}

	return patch
}

func slotProp(slot *metadata.Property, key string) *metadata.Property {
	if slot == nil || slot.Properties == nil {
		return nil
	}
	return slot.Properties[key]
}
