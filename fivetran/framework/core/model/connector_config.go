package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	configAttrTypes map[string]attr.Type
)

func getAttrTypes(configFieldsMap map[string]common.ConfigField) map[string]attr.Type {
	if len(configAttrTypes) == 0 {
		configAttrTypes = make(map[string]attr.Type)
		for fn, f := range configFieldsMap {
			configAttrTypes[fn] = attrTypeFromConfigField(f)
		}
	}
	return configAttrTypes
}

func GetTfTypes(configFieldsMap map[string]common.ConfigField, version int) map[string]tftypes.Type {
	newRes := map[string]tftypes.Type{}
	for fn, f := range configFieldsMap {
		if version < 2 && fn == "servers" {
			newRes[fn] = tftypes.String
		} else {
			newRes[fn] = tfTypeFromConfigField(f, version)
		}
	}
	return newRes
}

func tfTypeFromConfigField(cf common.ConfigField, version int) tftypes.Type {
	switch cf.FieldValueType {
	case common.Boolean:
		if version < 3 {
			return tftypes.String
		} else {
			return tftypes.Bool
		}
	case common.Integer:
		if version < 3 {
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
			subFields[fn] = tfTypeFromConfigField(f, version)
		}
		if version < 3 {
			return tftypes.Set{ElementType: tftypes.Object{AttributeTypes: subFields}}
		}
		return tftypes.Object{AttributeTypes: subFields}
	case common.ObjectList:
		subFields := map[string]tftypes.Type{}
		for fn, f := range cf.ItemFields {
			subFields[fn] = tfTypeFromConfigField(f, version)
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

func getValue(fieldType attr.Type, value, local interface{}, fieldsMap map[string]common.ConfigField, currentField *common.ConfigField, service string) attr.Value {
	if fieldType.Equal(types.StringType) {
		if value == nil {
			return types.StringNull()
		}
		if currentField != nil && currentField.Sensitive && local != nil {
			return types.StringValue(local.(string))
		}
		if t, ok := currentField.ItemType[service]; ok {
			if t == common.Integer {
				return types.StringValue(fmt.Sprintf("%v", value))
			}
		}
		return types.StringValue(value.(string))
	}
	if fieldType.Equal(types.BoolType) {
		if value == nil {
			return types.BoolNull()
		}
		return types.BoolValue(value.(bool))
	}
	if fieldType.Equal(types.Int64Type) {
		if value == nil {
			return types.Int64Null()
		}
		// value in json decoded response is always float64 for any kind of numbers
		return types.Int64Value(int64(value.(float64)))
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
				// field is not related to this particular service
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
				elements[fn] = getValue(et, nil, nil, cf.ItemFields, &cf, service)
			}
		}
		objectValue, _ := types.ObjectValue(complexType.AttributeTypes(), elements)
		return objectValue
	}

	if collectionType, ok := fieldType.(attr.TypeWithElementType); ok {
		if value == nil {
			if _, ok := collectionType.(basetypes.SetTypable); ok {
				return types.SetNull(collectionType.ElementType())
			} else {
				return types.ListNull(collectionType.ElementType())
			}
		}
		items := []attr.Value{}
		for _, v := range value.([]interface{}) {
			items = append(items, getValue(collectionType.ElementType(), v, nil, fieldsMap, currentField, service))
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
