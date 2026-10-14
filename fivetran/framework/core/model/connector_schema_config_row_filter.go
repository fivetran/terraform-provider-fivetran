package model

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var rowFilterClauseAttrTypes = map[string]attr.Type{
	"column":      types.StringType,
	"column_type": types.StringType,
	"operator":    types.StringType,
	"values":      types.ListType{ElemType: types.StringType},
}

var rowFilterAttrTypes = map[string]attr.Type{
	"name":        types.StringType,
	"description": types.StringType,
	"operator":    types.StringType,
	"column_clauses": types.ListType{
		ElemType: types.ObjectType{AttrTypes: rowFilterClauseAttrTypes},
	},
}

// rowFilterFromObject converts a configured row_filter into the API shape, null attributes are omitted
func rowFilterFromObject(value basetypes.ObjectValue) map[string]interface{} {
	result := map[string]interface{}{}
	attrs := value.Attributes()
	setStringIfKnown(result, attrs, "name")
	setStringIfKnown(result, attrs, "description")
	setStringIfKnown(result, attrs, "operator")
	if clausesList, ok := attrs["column_clauses"].(basetypes.ListValue); ok && !clausesList.IsNull() && !clausesList.IsUnknown() {
		clauses := []interface{}{}
		for _, ce := range clausesList.Elements() {
			clauseObject, ok := ce.(basetypes.ObjectValue)
			if !ok {
				continue
			}
			clauseAttrs := clauseObject.Attributes()
			clause := map[string]interface{}{}
			setStringIfKnown(clause, clauseAttrs, "column")
			setStringIfKnown(clause, clauseAttrs, "column_type")
			setStringIfKnown(clause, clauseAttrs, "operator")
			if valuesList, ok := clauseAttrs["values"].(basetypes.ListValue); ok && !valuesList.IsNull() && !valuesList.IsUnknown() {
				values := []interface{}{}
				for _, ve := range valuesList.Elements() {
					if v, ok := ve.(basetypes.StringValue); ok && !v.IsNull() && !v.IsUnknown() {
						values = append(values, v.ValueString())
					}
				}
				clause["values"] = values
			}
			clauses = append(clauses, clause)
		}
		result["column_clauses"] = clauses
	}
	return result
}

func setStringIfKnown(target map[string]interface{}, attrs map[string]attr.Value, key string) {
	if v, ok := attrs[key].(basetypes.StringValue); ok && !v.IsNull() && !v.IsUnknown() {
		target[key] = v.ValueString()
	}
}

// rowFilterToObject converts a row filter in the API shape into the row_filter attribute value, nil gives null
func rowFilterToObject(value interface{}) basetypes.ObjectValue {
	filter, ok := value.(map[string]interface{})
	if !ok {
		return types.ObjectNull(rowFilterAttrTypes)
	}
	clauseType := types.ObjectType{AttrTypes: rowFilterClauseAttrTypes}
	clausesValue := types.ListNull(clauseType)
	if rawClauses, ok := filter["column_clauses"].([]interface{}); ok {
		clauses := []attr.Value{}
		for _, rc := range rawClauses {
			rawClause, _ := rc.(map[string]interface{})
			valuesValue := types.ListNull(types.StringType)
			if rawValues, ok := rawClause["values"].([]interface{}); ok {
				values := []attr.Value{}
				for _, v := range rawValues {
					values = append(values, types.StringValue(fmt.Sprint(v)))
				}
				valuesValue, _ = types.ListValue(types.StringType, values)
			}
			clauseValue, _ := types.ObjectValue(rowFilterClauseAttrTypes, map[string]attr.Value{
				"column":      stringOrNull(rawClause["column"]),
				"column_type": stringOrNull(rawClause["column_type"]),
				"operator":    stringOrNull(rawClause["operator"]),
				"values":      valuesValue,
			})
			clauses = append(clauses, clauseValue)
		}
		clausesValue, _ = types.ListValue(clauseType, clauses)
	}
	result, _ := types.ObjectValue(rowFilterAttrTypes, map[string]attr.Value{
		"name":           stringOrNull(filter["name"]),
		"description":    stringOrNull(filter["description"]),
		"operator":       stringOrNull(filter["operator"]),
		"column_clauses": clausesValue,
	})
	return result
}

func stringOrNull(value interface{}) basetypes.StringValue {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(fmt.Sprint(value))
}
