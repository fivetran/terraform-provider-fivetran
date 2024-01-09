package model

import (
	"github.com/fivetran/go-fivetran/connectors"
	configSchema "github.com/fivetran/terraform-provider-fivetran/modules/connector/schema"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ConnectorSchemaResourceModel struct {
	Id                   types.String `tfsdk:"id"`
	ConnectorId          types.String `tfsdk:"connector_id"`
	SchemaChangeHandling types.String `tfsdk:"schema_change_handling"`
	Schema               types.Set    `tfsdk:"schema"`
}

func mapSchemas(schemas []interface{}) map[string]interface{} {
	mappedSchemas := map[string]interface{}{}

	for _, ls := range schemas {
		lsMap := ls.(map[string]interface{})
		mappedSchema := map[string]interface{}{}
		for k, v := range lsMap {
			mappedSchema[k] = v
		}
		mappedSchema["table"] = mapTables(lsMap["table"].([]interface{}))
		mappedSchemas[lsMap["name"].(string)] = mappedSchema
	}

	return mappedSchemas
}

func mapTables(tables []interface{}) map[string]interface{} {
	mappedTables := map[string]interface{}{}

	for _, lt := range tables {
		ltMap := lt.(map[string]interface{})
		mappedTable := map[string]interface{}{}
		for k, v := range ltMap {
			mappedTable[k] = v
		}
		mappedTable["column"] = mapColumns(ltMap["column"].([]interface{}))
		mappedTables[ltMap["name"].(string)] = mappedTable
	}

	return mappedTables
}

func mapColumns(columns []interface{}) map[string]interface{} {
	mappedColumns := map[string]interface{}{}
	for _, lc := range columns {
		lcMap := lc.(map[string]interface{})
		mappedColumn := map[string]interface{}{}
		for k, v := range lcMap {
			mappedColumn[k] = v
		}
		mappedColumns[lcMap["name"].(string)] = mappedColumn
	}
	return mappedColumns
}

func tryGetLocalSchema(mappedSchemas map[string]interface{}, schema string) map[string]interface{} {
	if v, ok := mappedSchemas[schema]; ok {
		return v.(map[string]interface{})
	}
	return map[string]interface{}{}
}

func tryGetLocalTable(mappedSchema map[string]interface{}, table string) map[string]interface{} {
	if tables, ok := mappedSchema["table"].(map[string]interface{}); ok {
		if t, ok := tables[table]; ok {
			return t.(map[string]interface{})
		}
	}
	return map[string]interface{}{}
}

func tryGetLocalColumn(mappedTable map[string]interface{}, column string) map[string]interface{} {
	if columns, ok := mappedTable["column"].(map[string]interface{}); ok {
		if c, ok := columns[column]; ok {
			return c.(map[string]interface{})
		}
	}
	return map[string]interface{}{}
}

func (d *ConnectorSchemaResourceModel) ReadFromResponse(response connectors.ConnectorSchemaDetailsResponse) {
	schemaObject := configSchema.SchemaConfig{}
	schemaObject.ReadFromResponse(response)

	localSchemas := mapSchemas(d.getSchemas(true))

	schemas := schemaObject.GetSchemas(response.Data.SchemaChangeHandling, d.GetSchemaConfig())

	columnAttrTypes := map[string]attr.Type{
		"name":    types.StringType,
		"enabled": types.BoolType,
		"hashed":  types.BoolType,
	}

	tableAttrTypes := map[string]attr.Type{
		"name":      types.StringType,
		"enabled":   types.BoolType,
		"sync_mode": types.StringType,
		"column": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: columnAttrTypes,
			},
		},
	}

	schemaElemAttrTypes := map[string]attr.Type{
		"name":    types.StringType,
		"enabled": types.BoolType,
		"table": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: tableAttrTypes,
			},
		},
	}

	items := []attr.Value{}
	for _, v := range schemas {
		schemaMap := v.(map[string]interface{})
		schemaName := schemaMap["name"].(string)

		localSchema := tryGetLocalSchema(localSchemas, schemaName)

		tables := []attr.Value{}
		for _, t := range schemaMap["table"].([]interface{}) {
			tableMap := t.(map[string]interface{})
			tableName := tableMap["name"].(string)

			localTable := tryGetLocalTable(localSchema, tableName)

			columns := []attr.Value{}
			for _, c := range tableMap["column"].([]interface{}) {
				columnMap := c.(map[string]interface{})
				columnName := columnMap["name"].(string)

				localColumn := tryGetLocalColumn(localTable, columnName)

				columnElements := map[string]attr.Value{}
				columnElements["name"] = types.StringValue(columnName)

				if _, ok := localColumn["enabled"]; ok {
					columnElements["enabled"] = types.BoolValue(helpers.StrToBool(columnMap["enabled"].(string)))
				} else {
					columnElements["enabled"] = types.BoolNull()
				}

				if _, ok := localColumn["hashed"]; ok {
					columnElements["hashed"] = types.BoolValue(helpers.StrToBool(columnMap["hashed"].(string)))
				} else {
					columnElements["hashed"] = types.BoolNull()
				}
				columnValue, _ := types.ObjectValue(columnAttrTypes, columnElements)
				columns = append(columns, columnValue)
			}
			tableElements := map[string]attr.Value{}
			tableElements["name"] = types.StringValue(tableName)

			tableElements["sync_mode"] = types.StringNull()

			if _, ok := localTable["sync_mode"]; ok {
				if sm, ok := tableMap["sync_mode"].(string); ok {
					tableElements["sync_mode"] = types.StringValue(sm)
				}
			}

			if _, ok := localTable["enabled"]; ok {
				tableElements["enabled"] = types.BoolValue(helpers.StrToBool(tableMap["enabled"].(string)))
			} else {
				tableElements["enabled"] = types.BoolNull()
			}
			tableElements["column"], _ = types.SetValue(types.ObjectType{AttrTypes: columnAttrTypes}, columns)
			tableValue, _ := types.ObjectValue(tableAttrTypes, tableElements)
			tables = append(tables, tableValue)
		}
		schemaElements := map[string]attr.Value{}
		schemaElements["name"] = types.StringValue(schemaName)
		if _, ok := localSchema["enabled"]; ok {
			schemaElements["enabled"] = types.BoolValue(helpers.StrToBool(schemaMap["enabled"].(string)))
		} else {
			schemaElements["enabled"] = types.BoolNull()
		}
		schemaElements["table"], _ = types.SetValue(types.ObjectType{AttrTypes: tableAttrTypes}, tables)
		objectValue, _ := types.ObjectValue(schemaElemAttrTypes, schemaElements)
		items = append(items, objectValue)
	}

	d.SchemaChangeHandling = types.StringValue(response.Data.SchemaChangeHandling)
	d.Schema, _ = types.SetValue(types.ObjectType{AttrTypes: schemaElemAttrTypes}, items)
}

func (d *ConnectorSchemaResourceModel) getSchemas(checkUnknowns bool) []interface{} {
	schemas := []interface{}{}
	for _, se := range d.Schema.Elements() {
		schema := map[string]interface{}{}
		if schemaElement, ok := se.(basetypes.ObjectValue); ok {
			tables := []interface{}{}
			tableSet := schemaElement.Attributes()["table"].(basetypes.SetValue)
			for _, te := range tableSet.Elements() {
				table := map[string]interface{}{}
				if tableElement, ok := te.(basetypes.ObjectValue); ok {
					columns := []interface{}{}
					columnSet := tableElement.Attributes()["column"].(basetypes.SetValue)
					for _, ce := range columnSet.Elements() {
						column := map[string]interface{}{}
						if columnElement, ok := ce.(basetypes.ObjectValue); ok {
							column["name"] = columnElement.Attributes()["name"].(basetypes.StringValue).ValueString()

							enabledValue := columnElement.Attributes()["enabled"].(basetypes.BoolValue)
							if (!enabledValue.IsUnknown() && !enabledValue.IsNull()) || !checkUnknowns {
								column["enabled"] = enabledValue.ValueBool()
							}

							hashedValue := columnElement.Attributes()["hashed"].(basetypes.BoolValue)
							if (!hashedValue.IsUnknown() && !hashedValue.IsNull()) || !checkUnknowns {
								column["hashed"] = columnElement.Attributes()["hashed"].(basetypes.BoolValue).ValueBool()
							}
						}
						columns = append(columns, column)
					}
					table["name"] = tableElement.Attributes()["name"].(basetypes.StringValue).ValueString()

					syncModeValue := tableElement.Attributes()["sync_mode"].(basetypes.StringValue)
					if (!syncModeValue.IsUnknown() && !syncModeValue.IsNull()) || !checkUnknowns {
						table["sync_mode"] = syncModeValue.ValueString()
					}

					enabledValue := tableElement.Attributes()["enabled"].(basetypes.BoolValue)
					if (!enabledValue.IsUnknown() && !enabledValue.IsNull()) || !checkUnknowns {
						table["enabled"] = enabledValue.ValueBool()
					}

					table["column"] = columns
				}
				tables = append(tables, table)
			}

			schema["name"] = schemaElement.Attributes()["name"].(basetypes.StringValue).ValueString()
			enabledValue := schemaElement.Attributes()["enabled"].(basetypes.BoolValue)

			if (!enabledValue.IsUnknown() && !enabledValue.IsNull()) || !checkUnknowns {
				schema["enabled"] = enabledValue.ValueBool()
			}
			schema["table"] = tables
		}
		schemas = append(schemas, schema)
	}
	return schemas
}

// Get raw flat schema config from model
func (d *ConnectorSchemaResourceModel) GetSchemaConfig() configSchema.SchemaConfig {
	result := configSchema.SchemaConfig{}
	result.ReadFromRawSourceData(d.getSchemas(true), d.SchemaChangeHandling.ValueString())
	return result
}
