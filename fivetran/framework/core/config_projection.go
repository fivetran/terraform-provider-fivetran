package core

import (
	"context"
	"encoding/json"
	"math"
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

// MapToDynamicWithTemplate is like MapToDynamic, but returns template itself when it already holds
// the same values as m. MapToDynamic always builds an object, while a configuration may use a map,
// list or set (for example config = tomap({...})), and Terraform rejects a result whose type differs
// from the planned one.
func MapToDynamicWithTemplate(ctx context.Context, m map[string]interface{}, template types.Dynamic) (types.Dynamic, diag.Diagnostics) {
	if !template.IsNull() && !template.IsUnknown() {
		templateMap, diags := DynamicToMap(ctx, template)
		if !diags.HasError() && sameJSON(templateMap, m) {
			return template, nil
		}
	}
	return MapToDynamic(ctx, m)
}

// sameJSON compares two values by their JSON form
func sameJSON(a, b map[string]interface{}) bool {
	if a == nil || b == nil {
		return false
	}
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)
	return aErr == nil && bErr == nil && string(aJSON) == string(bJSON)
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
	return project(remote, mask, slot, "", "")
}

// ProjectDynamicWithService is like ProjectDynamic, but keeps the mask order of set-typed fields
// of the given service. The API returns those in an arbitrary order on every read, so taking the
// remote order as-is would show a diff on every plan.
func ProjectDynamicWithService(remote, mask map[string]interface{}, slot *metadata.Property, service string) map[string]interface{} {
	return project(remote, mask, slot, service, "")
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
//   - set-typed list (see setTypedFields): keep the mask order (see alignSetOrder)
//   - list of objects: recurse element-wise by index (see projectList)
//   - normal: take remote value
//
// basePath is the dot-separated path of remote within the dynamic field, without list indexes
// (e.g. "custom_tables" for the items of custom_tables). It is used to look up set-typed fields.
func project(remote, mask map[string]interface{}, slot *metadata.Property, service, basePath string) map[string]interface{} {
	result := make(map[string]interface{}, len(mask))

	for key, maskVal := range mask {
		prop := SlotProp(slot, key)
		fieldPath := joinFieldPath(basePath, key)

		if prop != nil && prop.Format == "password" {
			result[key] = maskVal
			continue
		}

		remoteVal, inRemote := remote[key]
		if !inRemote {
			result[key] = maskVal
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
				result[key] = project(nestedRemote, nestedMask, nestedSlot, service, fieldPath)
				continue
			}
		}

		if maskList, ok := maskVal.([]interface{}); ok {
			if remoteList, ok2 := remoteVal.([]interface{}); ok2 {
				if isSetTypedField(service, fieldPath) {
					result[key] = alignSetOrder(remoteList, maskList)
					continue
				}
				var itemSlot *metadata.Property
				if prop != nil {
					itemSlot = prop.Items
				}
				result[key] = projectList(remoteList, maskList, itemSlot, service, fieldPath)
				continue
			}
		}

		result[key] = remoteVal
	}

	return result
}

// projectList projects object elements of a remote list onto the mask element at the same index,
// so upstream-only keys inside list items (e.g. new fields added to each row of s3 `files`) are dropped.
// Items are paired by index only when both lists have the same length: after an upstream insert or removal
// the pairing is unreliable, so the remote list is kept as-is to surface drift. Non-object elements are kept as-is.
func projectList(remote, mask []interface{}, itemSlot *metadata.Property, service, listPath string) []interface{} {
	if len(remote) != len(mask) {
		return remote
	}
	result := make([]interface{}, len(remote))
	for i, remoteItem := range remote {
		result[i] = remoteItem
		maskItem, ok := mask[i].(map[string]interface{})
		if !ok {
			continue
		}
		if remoteObject, ok := remoteItem.(map[string]interface{}); ok {
			result[i] = project(remoteObject, maskItem, itemSlot, service, listPath)
		}
	}
	return result
}

// alignSetOrder returns the remote elements of a set-typed list in the order of the prior (mask) list.
// Elements still returned by the API keep their prior position, and new ones follow in the API order.
// So a reorder-only read gives back the prior list exactly, and real drift shows only the elements
// that were added or removed upstream, not a reshuffle of the whole list.
func alignSetOrder(remote, prior []interface{}) []interface{} {
	remaining := make(map[string]int, len(remote))
	for _, remoteItem := range remote {
		if key, ok := setElementKey(remoteItem); ok {
			remaining[key]++
		}
	}

	result := make([]interface{}, 0, len(remote))
	matched := make(map[string]int, len(prior))
	for _, priorItem := range prior {
		key, ok := setElementKey(priorItem)
		if !ok || remaining[key] == 0 {
			continue
		}
		remaining[key]--
		matched[key]++
		// Keep the prior value so its type matches the one in state (int64 vs float64).
		result = append(result, priorItem)
	}

	for _, remoteItem := range remote {
		key, ok := setElementKey(remoteItem)
		if ok && matched[key] > 0 {
			matched[key]--
			continue
		}
		result = append(result, remoteItem)
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
	return prepareConfigPatchDynamic(plan, state, slot, "")
}

// PrepareConfigPatchDynamicWithService is like PrepareConfigPatchDynamic but accounts for
// set-typed fields that should be compared as unordered collections.
// service is used to look up which fields are set-typed.
func PrepareConfigPatchDynamicWithService(plan, state map[string]interface{}, slot *metadata.Property, service string) map[string]interface{} {
	return prepareConfigPatchDynamic(plan, state, slot, service)
}

func prepareConfigPatchDynamic(plan, state map[string]interface{}, slot *metadata.Property, service string) map[string]interface{} {
	patch := make(map[string]interface{})

	for k, planVal := range plan {
		prop := SlotProp(slot, k)
		if prop != nil && (prop.Readonly || prop.Immutable) {
			continue
		}
		stateVal, inState := state[k]
		if !inState {
			patch[k] = planVal
			continue
		}

		if dynamicValuesEqualRecursive(planVal, stateVal, service, k) {
			continue
		}
		patch[k] = planVal
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

// DynamicValuesEqualWithService compares two dynamic values, using set semantics for
// set-typed fields when service is non-empty. This is exported for use by resources.
func DynamicValuesEqualWithService(left, right interface{}, service string) bool {
	return dynamicValuesEqualRecursive(left, right, service, "")
}

// DynamicValuesEqualAtPath is like DynamicValuesEqualWithService for values found at fieldPath
// (dot-separated, without list indexes) within the dynamic field, so a set-typed field is
// recognized when the compared values are the field itself.
func DynamicValuesEqualAtPath(left, right interface{}, service, fieldPath string) bool {
	return dynamicValuesEqualRecursive(left, right, service, fieldPath)
}

// IsSetTypedField reports whether the field at fieldPath (dot-separated, without list indexes)
// of the given service holds an unordered collection.
func IsSetTypedField(service, fieldPath string) bool {
	return isSetTypedField(service, fieldPath)
}

// dynamicValuesEqualRecursive recursively compares values, with set semantics for set-typed fields.
func dynamicValuesEqualRecursive(left, right interface{}, service, fieldPath string) bool {
	// Check types first
	lType := reflect.TypeOf(left)
	rType := reflect.TypeOf(right)
	if lType != rType {
		return false
	}

	switch l := left.(type) {
	case []interface{}:
		r := right.([]interface{})
		if isSetTypedField(service, fieldPath) {
			return multisetEqual(l, r)
		}
		if len(l) != len(r) {
			return false
		}
		for i, lv := range l {
			if !dynamicValuesEqualRecursive(lv, r[i], service, fieldPath) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		r := right.(map[string]interface{})
		if len(l) != len(r) {
			return false
		}
		for k, lv := range l {
			rv, ok := r[k]
			if !ok {
				return false
			}
			if !dynamicValuesEqualRecursive(lv, rv, service, joinFieldPath(fieldPath, k)) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(left, right)
	}
}

// SlotProp returns the child Property for key within a slot's Properties map,
// or nil if the slot, its Properties map, or the key is absent.
func SlotProp(slot *metadata.Property, key string) *metadata.Property {
	if slot == nil || slot.Properties == nil {
		return nil
	}
	return slot.Properties[key]
}

// setTypedFields maps service names to their set-typed field paths.
// A value in this map indicates that the field, when present at the given path,
// should be compared as an unordered set rather than an ordered list.
//
// These are fields the connector holds in a Java Set (usually a HashSet). The API serializes them
// in the set's iteration order, which is unrelated to the order the user sent. For enum elements
// it even differs between API instances, because enum hash codes are identity-based, so two reads
// of an unchanged connection can return different orders. Connector metadata does not mark these
// arrays as unordered (no uniqueItems), so they have to be listed here.
var setTypedFields = map[string]map[string]bool{
	"facebook_ads": {
		"accounts":                           true,
		"custom_prebuilt_report_table_names": true,
		"conflict_table_names":               true,
		"fields":                             true,
		"breakdowns":                         true,
		"action_breakdowns":                  true,
		"attribution_windows":                true,
		"custom_tables.fields":               true,
		"custom_tables.breakdowns":           true,
		"custom_tables.action_breakdowns":    true,
		"custom_tables.attribution_windows":  true,
	},
}

// isSetTypedField returns true if the given field path in the given service
// should be compared as an unordered set.
func isSetTypedField(service string, fieldPath string) bool {
	serviceFields, ok := setTypedFields[service]
	if !ok {
		return false
	}
	return serviceFields[fieldPath]
}

func joinFieldPath(base, key string) string {
	if base == "" {
		return key
	}
	return base + "." + key
}

// setElementKey returns a key that is the same for equal elements of a set-typed list, so they can
// be counted in a map. Numbers are compared by value, because the API read-back decodes them as
// float64 while values from state are int64 (see normalizeNumbers). Objects are compared regardless of key order, since
// json.Marshal sorts map keys. ok is false for elements that can't be compared, e.g. ones holding
// unknown plan values, which would otherwise encode like an empty object.
func setElementKey(value interface{}) (string, bool) {
	if containsDynamicUnknown(value) {
		return "", false
	}
	encoded, err := json.Marshal(normalizeNumbers(value))
	if err != nil {
		return "", false
	}
	return string(encoded), true
}

func containsDynamicUnknown(value interface{}) bool {
	switch v := value.(type) {
	case DynamicUnknownValue:
		return true
	case []interface{}:
		for _, item := range v {
			if containsDynamicUnknown(item) {
				return true
			}
		}
	case map[string]interface{}:
		for _, item := range v {
			if containsDynamicUnknown(item) {
				return true
			}
		}
	}
	return false
}

// normalizeNumbers gives equal numbers one representation: integers stay exact int64, and a float
// becomes int64 only when it is a whole number in the int64 range. Converting integers to float64
// instead would merge distinct values above 2^53 (e.g. 9007199254740992 and 9007199254740993) and
// hide a real change.
func normalizeNumbers(value interface{}) interface{} {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float32:
		return normalizeFloat(float64(v))
	case float64:
		return normalizeFloat(v)
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = normalizeNumbers(item)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for k, item := range v {
			result[k] = normalizeNumbers(item)
		}
		return result
	}
	return value
}

func normalizeFloat(value float64) interface{} {
	// -2^63 and 2^63 are exact as float64; the upper bound itself is out of the int64 range.
	if value == math.Trunc(value) && value >= math.MinInt64 && value < -math.MinInt64 {
		return int64(value)
	}
	return value
}

// multisetEqual returns true if two slices contain the same elements in any order, with the
// same number of duplicates. Elements are compared by setElementKey.
func multisetEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[string]int, len(a))
	for _, item := range a {
		key, ok := setElementKey(item)
		if !ok {
			return false
		}
		counts[key]++
	}
	for _, item := range b {
		key, ok := setElementKey(item)
		if !ok || counts[key] == 0 {
			return false
		}
		counts[key]--
	}
	return true
}
