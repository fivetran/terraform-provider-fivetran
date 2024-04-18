package model

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func PrepareConfigAuthPatch(state, plan map[string]interface{}, service string, allFields map[string]common.ConfigField) map[string]interface{} {
	result := map[string]interface{}{}

	for k, v := range plan {
		// Include ALL fields from plan even if they have the same value as in state just for simplicity
		result[k] = v
	}

	// Filter out non nullable fields
	for k := range state {
		if _, ok := plan[k]; !ok {
			if f, ok := allFields[k]; ok {
				if f.Nullable || f.FieldValueType == common.ObjectList || f.FieldValueType == common.StringList {
					// If the field is not represented in plan (deleted from config)
					// And the field is nullable - it should be set to null explicitly
					result[k] = nil
				}
			}
		}
	}

	for k, pv := range plan {
		if sv, ok := state[k]; ok {
			if reflect.DeepEqual(pv, sv) {
				delete(result, k)
			}
		}
	}

	return result
}

func getAttrTypes(configFieldsMap map[string]common.ConfigField) map[string]attr.Type {
	result := make(map[string]attr.Type)
	for fn, f := range configFieldsMap {
		result[fn] = attrTypeFromConfigField(f)
	}
	return result
}

func GetTfTypes(configFieldsMap map[string]common.ConfigField, version int) map[string]tftypes.Type {
	newRes := map[string]tftypes.Type{}
	for fn, f := range configFieldsMap {
		if version < 2 && fn == "servers" {
			newRes[fn] = tftypes.String
		} else {
			newRes[fn] = tfTypeFromConfigField(f, version < 3)
		}
	}
	return newRes
}

func GetTfTypesDestination(configFieldsMap map[string]common.ConfigField, version int) map[string]tftypes.Type {
	newRes := map[string]tftypes.Type{}
	for fn, f := range configFieldsMap {
		newRes[fn] = tfTypeFromConfigField(f, version < 1)
	}
	return newRes
}

func tfTypeFromConfigField(cf common.ConfigField, primitivesAsString bool) tftypes.Type {
	switch cf.FieldValueType {
	case common.Boolean:
		if primitivesAsString {
			return tftypes.String
		} else {
			return tftypes.Bool
		}
	case common.Integer:
		if primitivesAsString {
			return tftypes.String
		} else {
			return tftypes.Number
		}
	case common.String:
		return tftypes.String
	case common.StringList:
		return tftypes.Set{ElementType: tftypes.String}
	case common.Object:
		subFields := map[string]tftypes.Type{}
		for fn, f := range cf.ItemFields {
			subFields[fn] = tfTypeFromConfigField(f, primitivesAsString)
		}
		if primitivesAsString {
			return tftypes.Set{ElementType: tftypes.Object{AttributeTypes: subFields}}
		}
		return tftypes.Object{AttributeTypes: subFields}
	case common.ObjectList:
		subFields := map[string]tftypes.Type{}
		for fn, f := range cf.ItemFields {
			subFields[fn] = tfTypeFromConfigField(f, primitivesAsString)
		}
		return tftypes.Set{ElementType: tftypes.Object{AttributeTypes: subFields}}
	}
	panic("Unknown FieldValueType " + cf.FieldValueType.String())
}

func attrTypeFromConfigField(cf common.ConfigField) attr.Type {
	switch cf.FieldValueType {
	case common.Boolean:
		return types.BoolType
	case common.Integer:
		return types.Int64Type
	case common.String:
		return types.StringType
	case common.StringList:
		return types.SetType{ElemType: types.StringType}
	case common.Object:
	case common.ObjectList:
		subFields := make(map[string]attr.Type)
		for fn, f := range cf.ItemFields {
			subFields[fn] = attrTypeFromConfigField(f)
		}
		if cf.FieldValueType == common.ObjectList {
			return types.SetType{ElemType: types.ObjectType{AttrTypes: subFields}}
		} else {
			return types.ObjectType{AttrTypes: subFields}
		}
	}
	return nil
}

func getStringValue(value, local interface{}, currentField *common.ConfigField, service string) types.String {
	if value == nil {
		if local != nil && local.(string) == "" {
			return types.StringValue("")
		}
		return types.StringNull()
	}
	if local == nil && !currentField.Readonly { // we should not set non-nullable value to the state if it's not configured by tf, we just ignore it
		return types.StringNull()
	}
	if currentField != nil && currentField.GetIsSensitive(service) && local != nil {
		return types.StringValue(local.(string))
	}
	if t, ok := currentField.ItemType[service]; ok {
		if t == common.Integer {
			return types.StringValue(fmt.Sprintf("%v", value))
		}
	}
	return types.StringValue(value.(string))
}

func getBoolValue(value, local interface{}, currentField *common.ConfigField) types.Bool {
	if value == nil || (local == nil && !currentField.Readonly) { // we should not set value to the state if it's not configured by tf
		return types.BoolNull()
	}
	if fValue, ok := value.(bool); ok {
		return types.BoolValue(fValue)
	}
	if sValue, ok := value.(string); ok {
		if sValue == "" {
			return types.BoolNull()
		}
		return types.BoolValue(helpers.StrToBool(sValue))
	}
	panic(fmt.Sprintf("Unable to read boolean value from %v", value))
}

func getIntValue(value, local interface{}, currentField *common.ConfigField) types.Int64 {
	if value == nil || (local == nil && !currentField.Readonly) { // we should not set value to the state if it's not configured by tf
		return types.Int64Null()
	}
	// value in json decoded response is always float64 for any kind of numbers
	if fValue, ok := value.(float64); ok {
		return types.Int64Value(int64(fValue))
	}
	if sValue, ok := value.(string); ok {
		if sValue == "" {
			return types.Int64Null()
		}
		i, err := strconv.Atoi(sValue)
		if err == nil {
			return types.Int64Value(int64(i))
		}
	}
	panic(fmt.Sprintf("Can't convert upstream value %v to types.Int64Type", value))
}

func getValue(
	fieldType attr.Type,
	value, local interface{},
	fieldsMap map[string]common.ConfigField,
	currentField *common.ConfigField,
	service string) attr.Value {
	if fieldType.Equal(types.StringType) {
		return getStringValue(value, local, currentField, service)
	}
	if fieldType.Equal(types.BoolType) {
		return getBoolValue(value, local, currentField)
	}
	if fieldType.Equal(types.Int64Type) {
		return getIntValue(value, local, currentField)
	}
	if complexType, ok := fieldType.(attr.TypeWithAttributeTypes); ok {
		if value == nil {
			return types.ObjectNull(complexType.AttributeTypes())
		}
		vMap := value.(map[string]interface{})
		var lMap map[string]interface{}
		if local != nil {
			lMap = local.(map[string]interface{})
		}
		elements := make(map[string]attr.Value)
		for fn, et := range complexType.AttributeTypes() {
			cf := fieldsMap[fn]

			if _, ok := fieldsMap[fn+"_"+service]; ok {
				// this field should be handled as service specific field, set this attr as nil
				elements[fn] = getValue(et, nil, nil, cf.ItemFields, &cf, service)
				continue
			}
			efn := fn
			if cf.ApiField != "" {
				efn = cf.ApiField
				// field is not related to this particular service, set this attr as nil
				if !strings.HasSuffix(fn, service) {
					elements[fn] = getValue(et, nil, nil, cf.ItemFields, &cf, service)
					continue
				}
			}
			// get upstream value
			if value, ok := vMap[efn]; ok {
				lValue := lMap[fn]
				elements[fn] = getValue(et, value, lValue, cf.ItemFields, &cf, service)
			} else {
				lValue := lMap[fn]
				elements[fn] = getValue(et, nil, lValue, cf.ItemFields, &cf, service)
			}
		}
		objectValue, _ := types.ObjectValue(complexType.AttributeTypes(), elements)
		return objectValue
	}

	if collectionType, ok := fieldType.(attr.TypeWithElementType); ok {
		if local != nil {
			localArray := local.([]interface{})
			if (currentField.GetIsSensitive(service)) || (value == nil && len(localArray) == 0) {
				items := []attr.Value{}
				for _, v := range localArray {
					items = append(items,
						getValue(collectionType.ElementType(), v, v, fieldsMap, currentField, service),
					)
				}
				if _, ok := collectionType.(basetypes.SetTypable); ok {
					setValue, _ := types.SetValue(collectionType.ElementType(), items)
					return setValue
				} else {
					listValue, _ := types.ListValue(collectionType.ElementType(), items)
					return listValue
				}
			}
		}
		if value == nil {
			if _, ok := collectionType.(basetypes.SetTypable); ok {
				return types.SetNull(collectionType.ElementType())
			} else {
				return types.ListNull(collectionType.ElementType())
			}
		}
		items := []attr.Value{}
		for _, v := range value.([]interface{}) {
			if currentField.ItemKeyField != "" && local != nil {
				keyField := currentField.ItemKeyField
				keyFields := strings.Split(keyField, "|")

				vMap := v.(map[string]interface{})

				if len(keyFields) > 1 {
					// choose key field
					fieldsMatch := 0
					for _, kf := range keyFields {
						if _, ok := vMap[kf]; ok { // if the field name represented in config - it could be key field
							keyField = kf
							fieldsMatch += 1
						}
					}
					// if more than one keyField found in config we don't know how to associate local and upstream values
					if fieldsMatch == 0 {
						panic("No key fields defined in configuration, can't associate upstream items")
					}
				}

				localCollection := local.([]interface{})

				for _, li := range localCollection {
					liMap := li.(map[string]interface{})
					if keyL, ok := liMap[keyField]; ok {
						if keyV, ok := vMap[keyField]; ok {
							if keyL == keyV {
								items = append(items,
									getValue(collectionType.ElementType(), v, li, fieldsMap, currentField, service),
								)
							}
						}
					}
				}
			} else {
				items = append(items,
					getValue(collectionType.ElementType(), v, v, fieldsMap, currentField, service),
				)
			}
		}
		if _, ok := collectionType.(basetypes.SetTypable); ok {
			setValue, _ := types.SetValue(collectionType.ElementType(), items)
			return setValue
		} else {
			listValue, _ := types.ListValue(collectionType.ElementType(), items)
			return listValue
		}
	}

	return nil
}

func getValueFromAttrValue(av attr.Value, fieldsMap map[string]common.ConfigField, currentField *common.ConfigField, service string) interface{} {
	if v, ok := av.(basetypes.StringValue); ok {
		if currentField != nil {
			if t, ok := currentField.ItemType[service]; ok {
				if t == common.Integer {
					res, err := strconv.Atoi(v.ValueString())
					if err != nil {
						panic(fmt.Sprintf("Can't convert value %v to int", v.ValueString()))
					}
					return res
				}
			}
		}
		return v.ValueString()
	}
	if v, ok := av.(basetypes.BoolValue); ok {
		return v.ValueBool()
	}
	if v, ok := av.(basetypes.Int64Value); ok {
		return v.ValueInt64()
	}
	if v, ok := av.(basetypes.ObjectValue); ok {
		result := make(map[string]interface{})
		for an, av := range v.Attributes() {
			if !av.IsUnknown() && !av.IsNull() {
				cf := fieldsMap[an]
				if scf, ok := fieldsMap[an+"_"+service]; ok {
					cf = scf
				}
				result[an] = getValueFromAttrValue(av, cf.ItemFields, &cf, service)
			}
		}
		return result
	}
	if v, ok := av.(basetypes.SetValue); ok {
		result := make([]interface{}, 0)
		for _, ev := range v.Elements() {
			result = append(result, getValueFromAttrValue(ev, fieldsMap, currentField, service))
		}
		return result
	}
	if v, ok := av.(basetypes.ListValue); ok {
		result := make([]interface{}, 0)
		for _, ev := range v.Elements() {
			result = append(result, getValueFromAttrValue(ev, fieldsMap, currentField, service))
		}
		return result
	}
	return nil
}

func patchServiceSpecificFields(
	result map[string]interface{},
	serviceName string,
	serviceFields map[string]common.ConfigField,
	allFields map[string]common.ConfigField) error {
	replacements := map[string]interface{}{}
	for rf := range result {
		if _, ok := serviceFields[rf]; !ok {
			// should lookup for service_field_name pattern
			serviceSpecificName := rf + "_" + serviceName
			if _, ok := serviceFields[serviceSpecificName]; ok {
				// field service_field_name found: prompt this field to user
				return fmt.Errorf("field `%v` isn't expected for service `%v`, try use `%v` instead", rf, serviceName, serviceSpecificName)
			}
		}
		if field, ok := allFields[rf]; ok {
			value := result[rf]
			if field.ApiField != "" && field.ApiField != rf {
				delete(result, rf)
				replacements[field.ApiField] = value
			}
		}
	}
	if len(replacements) > 0 {
		for k, v := range replacements {
			result[k] = v
		}
	}
	return nil
}
