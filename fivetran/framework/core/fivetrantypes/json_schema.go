package fivetrantypes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*JsonSchemaValue)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*JsonSchemaValue)(nil)
)

type JsonSchemaValue struct {
	basetypes.StringValue
}

func (v JsonSchemaValue) Type(_ context.Context) attr.Type {
	return JsonSchemaType{}
}

func (v JsonSchemaValue) Equal(o attr.Value) bool {
	other, ok := o.(JsonSchemaValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v JsonSchemaValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(JsonSchemaValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	result, err := jsonEqual(newValue.ValueString(), v.ValueString())
	if err != nil {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected error occurred while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return false, diags
	}

	result, err = schemaEqual(newValue.ValueString(), v.ValueString())
	if err != nil {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected error occurred while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return false, diags
	}

	return result, diags
}

func schemaEqual(s1, s2 string) (bool, error) {
	schema1, err := unmarshalSchema(s1)
	if err != nil {
		return false, err
	}
	schema2, err := unmarshalSchema(s2)
	if err != nil {
		return false, err
	}
	schema1json, err := json.Marshal(schema1)
	if err != nil {
		return false, err
	}
	schema2json, err := json.Marshal(schema2)
	if err != nil {
		return false, err
	}
	schema1string := string(schema1json)
	schema2string := string(schema2json)
	return schema1string == schema2string, nil
}

func unmarshalSchema(schema string) (map[string]interface{}, error) {
	var schemasMap map[string]interface{}
	err := json.Unmarshal([]byte(schema), &schemasMap)
	schemas := map[string]interface{}{}
	if err != nil {
		return schemas, err
	}

	for sName, sValue := range schemasMap {
		if sMap, ok := sValue.(map[string]interface{}); ok {
			sResult := map[string]interface{}{}
			if sEnabled, ok := sMap["enabled"]; ok {
				if sEnabledBool, ok := helpers.GetBoolOk(sEnabled); ok {
					sResult["enabled"] = sEnabledBool
				}
			}
			if sTables, ok := sMap["tables"]; ok {
				if sTablesMap, ok := sTables.(map[string]interface{}); ok {
					tables := map[string]interface{}{}
					for tName, tValue := range sTablesMap {
						if tMap, ok := tValue.(map[string]interface{}); ok {
							tResult := map[string]interface{}{}
							if tEnabled, ok := tMap["enabled"]; ok {
								if tEnabledBool, ok := helpers.GetBoolOk(tEnabled); ok {
									tResult["enabled"] = tEnabledBool
								}
							}
							if tSyncMode, ok := tMap["sync_mode"]; ok {
								tResult["sync_mode"] = tSyncMode
							}
							if tColumns, ok := tMap["columns"]; ok {
								if tColumnsMap, ok := tColumns.(map[string]interface{}); ok {
									columns := map[string]interface{}{}
									for cName, cValue := range tColumnsMap {
										if cMap, ok := cValue.(map[string]interface{}); ok {
											cResult := map[string]interface{}{}
											if cEnabled, ok := cMap["enabled"]; ok {
												if cEnabledBool, ok := helpers.GetBoolOk(cEnabled); ok {
													cResult["enabled"] = cEnabledBool
												}
											}
											if cHashed, ok := cMap["hashed"]; ok {
												if cHashedBool, ok := helpers.GetBoolOk(cHashed); ok {
													cResult["hashed"] = cHashedBool
												}
											}
											if len(cResult) > 0 {
												columns[cName] = cResult
											}
										}
									}
									if len(columns) > 0 {
										tResult["columns"] = columns
									}
								}
							}
							if len(tResult) > 0 {
								tables[tName] = tResult
							}
						}
					}
					if len(tables) > 0 {
						sResult["tables"] = tables
					}
				}
			}
			if len(sResult) > 0 {
				schemas[sName] = sResult
			}
		}
	}

	return schemas, nil
}

func (v JsonSchemaValue) Unmarshal(target any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v.IsNull() {
		diags.Append(diag.NewErrorDiagnostic("JSON Schema Unmarshal Error", "json string value is null"))
		return diags
	}

	if v.IsUnknown() {
		diags.Append(diag.NewErrorDiagnostic("JSON Schema Unmarshal Error", "json string value is unknown"))
		return diags
	}

	err := json.Unmarshal([]byte(v.ValueString()), target)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("JSON Schema Unmarshal Error", err.Error()))
	}

	return diags
}

func NewJsonSchemaNull() JsonSchemaValue {
	return JsonSchemaValue{
		StringValue: basetypes.NewStringNull(),
	}
}

func NewJsonSchemaUnknown() JsonSchemaValue {
	return JsonSchemaValue{
		StringValue: basetypes.NewStringUnknown(),
	}
}

func NewJsonSchemaValue(value string) JsonSchemaValue {
	return JsonSchemaValue{
		StringValue: basetypes.NewStringValue(value),
	}
}

func NewJsonSchemaPointerValue(value *string) JsonSchemaValue {
	return JsonSchemaValue{
		StringValue: basetypes.NewStringPointerValue(value),
	}
}
