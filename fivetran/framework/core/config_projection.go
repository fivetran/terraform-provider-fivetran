package core

import (
	"context"
	"math/big"
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
		diags.AddError("Dynamic value is not an object", "Expected an object value for this dynamic attribute.")
		return nil, diags
	}
	return m, diags
}

type DynamicUnknownValue struct{}

func IsDynamicUnknownValue(value interface{}) bool {
	_, ok := value.(DynamicUnknownValue)
	return ok
}

// DynamicToMapPreserveUnknown converts dynamic objects like DynamicToMap, but preserves
// nested unknown values so plan-time validation can distinguish them from explicit nulls.
func DynamicToMapPreserveUnknown(ctx context.Context, dyn types.Dynamic) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	if dyn.IsNull() || dyn.IsUnknown() {
		return nil, diags
	}
	underlying := dyn.UnderlyingValue()
	if underlying == nil || underlying.IsNull() || underlying.IsUnknown() {
		return nil, diags
	}
	result := attrToInterfacePreserveUnknown(ctx, underlying)
	if result == nil {
		return nil, diags
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		diags.AddError("Dynamic value is not an object", "Expected an object value for this dynamic attribute.")
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
	case types.Number:
		n := v.ValueBigFloat()
		if n == nil {
			return nil
		}
		if i, acc := n.Int64(); acc == big.Exact {
			return i
		}
		f, _ := n.Float64()
		return f
	case types.Dynamic:
		return attrToInterface(ctx, v.UnderlyingValue())
	case types.Object:
		return convertObjectAttrs(ctx, v.Attributes())
	case types.Map:
		return convertObjectAttrs(ctx, v.Elements())
	case types.List:
		return convertSliceElems(ctx, v.Elements())
	case types.Set:
		return convertSliceElems(ctx, v.Elements())
	case types.Tuple:
		return convertSliceElems(ctx, v.Elements())
	}
	return nil
}

func attrToInterfacePreserveUnknown(ctx context.Context, val attr.Value) interface{} {
	if val == nil || val.IsNull() {
		return nil
	}
	if val.IsUnknown() {
		return DynamicUnknownValue{}
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
	case types.Number:
		n := v.ValueBigFloat()
		if n == nil {
			return nil
		}
		if i, acc := n.Int64(); acc == big.Exact {
			return i
		}
		f, _ := n.Float64()
		return f
	case types.Dynamic:
		return attrToInterfacePreserveUnknown(ctx, v.UnderlyingValue())
	case types.Object:
		return convertObjectAttrsPreserveUnknown(ctx, v.Attributes())
	case types.Map:
		return convertObjectAttrsPreserveUnknown(ctx, v.Elements())
	case types.List:
		return convertSliceElemsPreserveUnknown(ctx, v.Elements())
	case types.Set:
		return convertSliceElemsPreserveUnknown(ctx, v.Elements())
	case types.Tuple:
		return convertSliceElemsPreserveUnknown(ctx, v.Elements())
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

func convertObjectAttrsPreserveUnknown(ctx context.Context, elems map[string]attr.Value) map[string]interface{} {
	result := make(map[string]interface{}, len(elems))
	for k, v := range elems {
		result[k] = attrToInterfacePreserveUnknown(ctx, v)
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

func convertSliceElemsPreserveUnknown(ctx context.Context, elems []attr.Value) []interface{} {
	result := make([]interface{}, len(elems))
	for i, v := range elems {
		result[i] = attrToInterfacePreserveUnknown(ctx, v)
	}
	return result
}

// MapToDynamic converts a map back into a Terraform dynamic object value.
func MapToDynamic(ctx context.Context, m map[string]interface{}) (types.Dynamic, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m == nil {
		return types.DynamicNull(), diags
	}

	value, _, valueDiags := interfaceToAttrValue(ctx, m)
	diags.Append(valueDiags...)
	if diags.HasError() {
		return types.DynamicNull(), diags
	}

	return types.DynamicValue(value), diags
}

func interfaceToAttrValue(ctx context.Context, value interface{}) (attr.Value, attr.Type, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch v := value.(type) {
	case nil:
		return types.DynamicNull(), types.DynamicType, diags
	case string:
		return types.StringValue(v), types.StringType, diags
	case bool:
		return types.BoolValue(v), types.BoolType, diags
	case int:
		return types.Int64Value(int64(v)), types.Int64Type, diags
	case int8:
		return types.Int64Value(int64(v)), types.Int64Type, diags
	case int16:
		return types.Int64Value(int64(v)), types.Int64Type, diags
	case int32:
		return types.Int64Value(int64(v)), types.Int64Type, diags
	case int64:
		return types.Int64Value(v), types.Int64Type, diags
	case uint:
		return types.Int64Value(int64(v)), types.Int64Type, diags
	case uint8:
		return types.Int64Value(int64(v)), types.Int64Type, diags
	case uint16:
		return types.Int64Value(int64(v)), types.Int64Type, diags
	case uint32:
		return types.Int64Value(int64(v)), types.Int64Type, diags
	case uint64:
		return types.NumberValue(new(big.Float).SetUint64(v)), types.NumberType, diags
	case float32:
		return numberValue(float64(v)), types.NumberType, diags
	case float64:
		return numberValue(v), types.NumberType, diags
	case map[string]interface{}:
		attrTypes := make(map[string]attr.Type, len(v))
		attrValues := make(map[string]attr.Value, len(v))
		for key, nested := range v {
			nestedValue, nestedType, nestedDiags := interfaceToAttrValue(ctx, nested)
			diags.Append(nestedDiags...)
			attrTypes[key] = nestedType
			attrValues[key] = nestedValue
		}
		result, resultDiags := types.ObjectValue(attrTypes, attrValues)
		diags.Append(resultDiags...)
		return result, types.ObjectType{AttrTypes: attrTypes}, diags
	case []interface{}:
		elementTypes := make([]attr.Type, 0, len(v))
		elementValues := make([]attr.Value, 0, len(v))
		for _, nested := range v {
			nestedValue, nestedType, nestedDiags := interfaceToAttrValue(ctx, nested)
			diags.Append(nestedDiags...)
			elementTypes = append(elementTypes, nestedType)
			elementValues = append(elementValues, nestedValue)
		}
		result, resultDiags := types.TupleValue(elementTypes, elementValues)
		diags.Append(resultDiags...)
		return result, types.TupleType{ElemTypes: elementTypes}, diags
	}

	diags.AddError("Unsupported dynamic value", "The API returned a dynamic field value the provider does not know how to store in Terraform state.")
	return types.DynamicNull(), types.DynamicType, diags
}

func numberValue(value float64) types.Number {
	f := new(big.Float).SetFloat64(value)
	if i, acc := f.Int64(); acc == big.Exact {
		return types.NumberValue(new(big.Float).SetInt64(i))
	}
	return types.NumberValue(f)
}

// ProjectDynamic filters an API read-back (remote) down to managed keys.
func ProjectDynamic(remote, mask map[string]interface{}, slot *metadata.Property) map[string]interface{} {
	return project(remote, mask, slot)
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
		prop := SlotProp(slot, key)

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

	for key, remoteVal := range remote {
		if _, inMask := mask[key]; inMask {
			continue
		}
		prop := SlotProp(slot, key)
		if prop != nil && prop.Readonly {
			result[key] = remoteVal
		}
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
		prop := SlotProp(slot, k)
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
		prop := SlotProp(slot, k)
		if prop != nil && (prop.Readonly || prop.Immutable) {
			continue
		}
		if prop != nil && prop.Nullable {
			patch[k] = nil
		}
	}

	return patch
}

// SlotProp returns the child Property for key within a slot's Properties map,
// or nil if the slot, its Properties map, or the key is absent.
func SlotProp(slot *metadata.Property, key string) *metadata.Property {
	if slot == nil || slot.Properties == nil {
		return nil
	}
	return slot.Properties[key]
}
