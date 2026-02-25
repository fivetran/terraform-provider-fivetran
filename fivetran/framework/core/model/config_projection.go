package model

import (
	"context"
	"reflect"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DynamicToMapPublic is the exported wrapper for use outside the model package.
func DynamicToMapPublic(ctx context.Context, dyn types.Dynamic) map[string]any {
	return dynamicToMap(ctx, dyn)
}

// dynamicToMap converts a types.Dynamic holding an object value to map[string]any.
// Empty strings are converted to nil (treated as explicit clear operation).
// Null values are preserved as nil. Unknown values are skipped.
func dynamicToMap(ctx context.Context, dyn types.Dynamic) map[string]any {
	if dyn.IsNull() || dyn.IsUnknown() {
		return nil
	}
	underlying := dyn.UnderlyingValue()
	if underlying == nil || underlying.IsNull() || underlying.IsUnknown() {
		return nil
	}
	return attrValueToAny(ctx, underlying).(map[string]any)
}

func attrValueToAny(ctx context.Context, val attr.Value) any {
	if val == nil || val.IsNull() || val.IsUnknown() {
		return nil
	}

	switch v := val.(type) {
	case types.String:
		s := v.ValueString()
		if s == "" {
			return nil // empty string = explicit clear
		}
		return s
	case types.Bool:
		return v.ValueBool()
	case types.Int64:
		return v.ValueInt64()
	case types.Float64:
		return v.ValueFloat64()
	case types.Object:
		return convertMapElements(ctx, v.Attributes())
	case types.Map:
		return convertMapElements(ctx, v.Elements())
	case types.List:
		return convertSliceElements(ctx, v.Elements())
	case types.Set:
		return convertSliceElements(ctx, v.Elements())
	}
	return nil
}

func convertMapElements(ctx context.Context, elements map[string]attr.Value) map[string]any {
	result := map[string]any{}
	for k, av := range elements {
		result[k] = attrValueToAny(ctx, av)
	}
	return result
}

func convertSliceElements(ctx context.Context, elements []attr.Value) []any {
	result := []any{}
	for _, av := range elements {
		result = append(result, attrValueToAny(ctx, av))
	}
	return result
}

// project filters remoteConfig to only the keys present in mask (the local state).
// Sensitive fields: keep local (state) value — never overwrite from remote masked value.
// Readonly fields: skip — they are computed and not user-managed.
// Keys in mask but absent in remote: set to nil to surface drift.
func project(remote, mask map[string]any, meta *metadata.ConnectorMetadata) map[string]any {
	result := map[string]any{}
	for key, maskVal := range mask {
		prop := findProperty(meta, key)

		if prop != nil && prop.Sensitive {
			result[key] = maskVal // keep local value, never read from remote
			continue
		}

		if prop != nil && prop.Readonly {
			continue // skip readonly fields
		}

		remoteVal, exists := remote[key]
		if !exists {
			result[key] = nil // field absent in remote → nil triggers drift if user has value
			continue
		}

		// recurse for nested objects
		if nestedMask, ok := maskVal.(map[string]any); ok {
			if nestedRemote, ok := remoteVal.(map[string]any); ok {
				result[key] = project(nestedRemote, nestedMask, meta)
				continue
			}
		}

		result[key] = remoteVal
	}
	return result
}

// findProperty looks up a property by name in the connector metadata config schema.
func findProperty(meta *metadata.ConnectorMetadata, key string) *metadata.Property {
	if meta == nil {
		return nil
	}
	if prop, ok := meta.Config.Properties[key]; ok {
		return prop
	}
	return nil
}

// PrepareConfigPatchDynamic builds the PATCH payload for a dynamic config.
// Key fix over PrepareConfigAuthPatch: any field removed from config sends null to the API,
// regardless of whether it was marked nullable in fields.json.
func PrepareConfigPatchDynamic(state, plan map[string]any, meta *metadata.ConnectorMetadata) map[string]any {
	result := map[string]any{}

	// include plan fields, skipping unchanged ones
	for k, pv := range plan {
		if sv, ok := state[k]; !ok || !reflect.DeepEqual(pv, sv) {
			result[k] = pv
		}
	}

	// fields in state but absent in plan: send null to clear them
	for k := range state {
		if _, ok := plan[k]; !ok {
			prop := findProperty(meta, k)
			if prop == nil || !prop.Readonly {
				result[k] = nil
			}
		}
	}

	return result
}
