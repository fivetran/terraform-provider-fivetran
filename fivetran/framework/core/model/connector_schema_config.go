package model

import (
	"github.com/fivetran/go-fivetran/connectors"
	configSchema "github.com/fivetran/terraform-provider-fivetran/modules/connector/schema"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ConnectorSchemaResourceModel struct {
	Id                   types.String   `tfsdk:"id"`
	ConnectorId          types.String   `tfsdk:"connector_id"`
	SchemaChangeHandling types.String   `tfsdk:"schema_change_handling"`
	Schemas              types.Map      `tfsdk:"schemas"`
	Schema               types.Set      `tfsdk:"schema"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
}

func (d *ConnectorSchemaResourceModel) IsLegacySchemaDefined() bool {
	return !d.Schema.IsUnknown() && !d.Schema.IsNull() && len(d.Schema.Elements()) > 0
}

func (d *ConnectorSchemaResourceModel) IsMappedSchemaDefined() bool {
	return !d.Schemas.IsUnknown() && !d.Schemas.IsNull() && len(d.Schemas.Elements()) > 0
}

func (d *ConnectorSchemaResourceModel) ReadFromResponse(response connectors.ConnectorSchemaDetailsResponse) {
	schemaObject := configSchema.SchemaConfig{}
	schemaObject.ReadFromResponse(response)
	schemas := schemaObject.GetSchemas(response.Data.SchemaChangeHandling, d.GetSchemaConfig())

	if d.IsLegacySchemaDefined() {
		d.Schema = d.getLegacySchemaItems(schemas)
		d.Schemas = d.getNullSchemas()
	}
	if d.IsMappedSchemaDefined() {
		d.Schema = d.getNullSchema()
		d.Schemas = d.getSchemasMap(schemas)
	}

	d.SchemaChangeHandling = types.StringValue(response.Data.SchemaChangeHandling)
}

func (d *ConnectorSchemaResourceModel) getNullSchema() basetypes.SetValue {
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
	return types.SetNull(types.ObjectType{AttrTypes: schemaElemAttrTypes})
}

func (d *ConnectorSchemaResourceModel) getNullSchemas() basetypes.MapValue {
	columnsAttrTypes := map[string]attr.Type{
		"enabled": types.BoolType,
		"hashed":  types.BoolType,
	}

	tablesAttrTypes := map[string]attr.Type{
		"enabled":   types.BoolType,
		"sync_mode": types.StringType,
		"columns": types.MapType{
			ElemType: types.ObjectType{
				AttrTypes: columnsAttrTypes,
			},
		},
	}
	schemasAttrTypes := map[string]attr.Type{
		"enabled": types.BoolType,
		"tables": types.MapType{
			ElemType: types.ObjectType{
				AttrTypes: tablesAttrTypes,
			},
		},
	}
	return types.MapNull(types.ObjectType{AttrTypes: schemasAttrTypes})
}

// Get raw flat schema config from model
func (d *ConnectorSchemaResourceModel) GetSchemaConfig() configSchema.SchemaConfig {
	result := configSchema.SchemaConfig{}
	schemas := []interface{}{}
	if d.IsLegacySchemaDefined() {
		schemas = d.getLegacySchemas()
	}
	if d.IsMappedSchemaDefined() {
		schemas = d.getSchemas()
	}
	result.ReadFromRawSourceData(schemas, d.SchemaChangeHandling.ValueString())
	return result
}

func (d *ConnectorSchemaResourceModel) getSchemasMap(schemas []interface{}) basetypes.MapValue {
	columnsAttrTypes := map[string]attr.Type{
		"enabled": types.BoolType,
		"hashed":  types.BoolType,
	}

	tablesAttrTypes := map[string]attr.Type{
		"enabled":   types.BoolType,
		"sync_mode": types.StringType,
		"columns": types.MapType{
			ElemType: types.ObjectType{
				AttrTypes: columnsAttrTypes,
			},
		},
	}
	schemasAttrTypes := map[string]attr.Type{
		"enabled": types.BoolType,
		"tables": types.MapType{
			ElemType: types.ObjectType{
				AttrTypes: tablesAttrTypes,
			},
		},
	}
	schemasMap := map[string]attr.Value{}
	localSchemas := d.mapLocalSchemas()

	for _, v := range schemas {
		schemaMap := v.(map[string]interface{})
		schemaName := schemaMap["name"].(string)
		localSchema := d.tryGetLocalSchema(localSchemas, schemaName)
		schemaElements := map[string]attr.Value{}
		if _, ok := localSchema["enabled"]; ok {
			schemaElements["enabled"] = types.BoolValue(helpers.StrToBool(schemaMap["enabled"].(string)))
		} else {
			schemaElements["enabled"] = types.BoolNull()
		}

		tables := map[string]attr.Value{}

		for _, t := range schemaMap["table"].([]interface{}) {
			tableMap := t.(map[string]interface{})
			tableName := tableMap["name"].(string)
			localTable := d.tryGetLocalTable(localSchema, tableName)

			tableElements := map[string]attr.Value{}
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
			columns := map[string]attr.Value{}
			for _, c := range tableMap["column"].([]interface{}) {
				columnMap := c.(map[string]interface{})
				columnName := columnMap["name"].(string)

				localColumn := d.tryGetLocalColumn(localTable, columnName)

				columnElements := map[string]attr.Value{}

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
				columnValue, _ := types.ObjectValue(columnsAttrTypes, columnElements)
				columns[columnName] = columnValue
			}
			if len(columns) > 0 {
				tableElements["columns"], _ = types.MapValue(types.ObjectType{AttrTypes: columnsAttrTypes}, columns)
			} else {
				tableElements["columns"] = types.MapNull(types.ObjectType{AttrTypes: columnsAttrTypes})
			}
			tableValue, _ := types.ObjectValue(tablesAttrTypes, tableElements)
			tables[tableName] = tableValue
		}
		if len(tables) > 0 {
			schemaElements["tables"], _ = types.MapValue(types.ObjectType{AttrTypes: tablesAttrTypes}, tables)
		} else {
			schemaElements["tables"] = types.MapNull(types.ObjectType{AttrTypes: tablesAttrTypes})
		}
		schemaValue, _ := types.ObjectValue(schemasAttrTypes, schemaElements)
		schemasMap[schemaName] = schemaValue
	}

	result, _ := types.MapValue(types.ObjectType{AttrTypes: schemasAttrTypes}, schemasMap)
	return result
}

func (d *ConnectorSchemaResourceModel) tryGetLocalSchema(mappedSchemas map[string]interface{}, schema string) map[string]interface{} {
	if v, ok := mappedSchemas[schema]; ok {
		return v.(map[string]interface{})
	}
	return map[string]interface{}{}
}
func (d *ConnectorSchemaResourceModel) tryGetLocalTable(mappedSchema map[string]interface{}, table string) map[string]interface{} {
	if tables, ok := mappedSchema["table"].(map[string]interface{}); ok {
		if t, ok := tables[table]; ok {
			return t.(map[string]interface{})
		}
	}
	return map[string]interface{}{}
}
func (d *ConnectorSchemaResourceModel) tryGetLocalColumn(mappedTable map[string]interface{}, column string) map[string]interface{} {
	if columns, ok := mappedTable["column"].(map[string]interface{}); ok {
		if c, ok := columns[column]; ok {
			return c.(map[string]interface{})
		}
	}
	return map[string]interface{}{}
}

func (d *ConnectorSchemaResourceModel) getLegacySchemaItems(schemas []interface{}) basetypes.SetValue {
	schemaItems := []attr.Value{}
	localSchemas := d.mapLocalSchemas()
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
	for _, v := range schemas {
		schemaMap := v.(map[string]interface{})
		schemaName := schemaMap["name"].(string)

		localSchema := d.tryGetLocalSchema(localSchemas, schemaName)

		tables := []attr.Value{}
		for _, t := range schemaMap["table"].([]interface{}) {
			tableMap := t.(map[string]interface{})
			tableName := tableMap["name"].(string)

			localTable := d.tryGetLocalTable(localSchema, tableName)

			columns := []attr.Value{}
			for _, c := range tableMap["column"].([]interface{}) {
				columnMap := c.(map[string]interface{})
				columnName := columnMap["name"].(string)

				localColumn := d.tryGetLocalColumn(localTable, columnName)

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
		schemaItems = append(schemaItems, objectValue)
	}
	result, _ := types.SetValue(types.ObjectType{AttrTypes: schemaElemAttrTypes}, schemaItems)
	return result
}

func (d *ConnectorSchemaResourceModel) getLegacySchemas() []interface{} {
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
							if !enabledValue.IsUnknown() && !enabledValue.IsNull() {
								column["enabled"] = enabledValue.ValueBool()
							}

							hashedValue := columnElement.Attributes()["hashed"].(basetypes.BoolValue)
							if !hashedValue.IsUnknown() && !hashedValue.IsNull() {
								column["hashed"] = columnElement.Attributes()["hashed"].(basetypes.BoolValue).ValueBool()
							}
						}
						columns = append(columns, column)
					}
					table["name"] = tableElement.Attributes()["name"].(basetypes.StringValue).ValueString()

					syncModeValue := tableElement.Attributes()["sync_mode"].(basetypes.StringValue)
					if !syncModeValue.IsUnknown() && !syncModeValue.IsNull() {
						table["sync_mode"] = syncModeValue.ValueString()
					}

					enabledValue := tableElement.Attributes()["enabled"].(basetypes.BoolValue)
					if !enabledValue.IsUnknown() && !enabledValue.IsNull() {
						table["enabled"] = enabledValue.ValueBool()
					}

					table["column"] = columns
				}
				tables = append(tables, table)
			}

			schema["name"] = schemaElement.Attributes()["name"].(basetypes.StringValue).ValueString()
			enabledValue := schemaElement.Attributes()["enabled"].(basetypes.BoolValue)

			if !enabledValue.IsUnknown() && !enabledValue.IsNull() {
				schema["enabled"] = enabledValue.ValueBool()
			}
			schema["table"] = tables
		}
		schemas = append(schemas, schema)
	}
	return schemas
}

func (d *ConnectorSchemaResourceModel) getSchemas() []interface{} {
	schemas := []interface{}{}
	for sName, se := range d.Schemas.Elements() {
		if schemaElement, ok := se.(basetypes.ObjectValue); ok {
			schema := map[string]interface{}{
				"name":    sName,
				"enabled": schemaElement.Attributes()["enabled"].(basetypes.BoolValue).ValueBool(),
			}
			schemas = append(schemas, schema)

			if tablesMap, ok := schemaElement.Attributes()["tables"].(basetypes.MapValue); ok {
				tables := []interface{}{}
				for tName, te := range tablesMap.Elements() {
					if tableElement, ok := te.(basetypes.ObjectValue); ok {
						table := map[string]interface{}{
							"name":      tName,
							"enabled":   tableElement.Attributes()["enabled"].(basetypes.BoolValue).ValueBool(),
							"sync_mode": tableElement.Attributes()["sync_mode"].(basetypes.StringValue).ValueString(),
						}
						tables = append(tables, table)

						if columnsMap, ok := tableElement.Attributes()["columns"].(basetypes.MapValue); ok {
							columns := []interface{}{}
							for cName, ce := range columnsMap.Elements() {
								if columnElement, ok := ce.(basetypes.ObjectValue); ok {
									column := map[string]interface{}{
										"name":    cName,
										"enabled": columnElement.Attributes()["enabled"].(basetypes.BoolValue).ValueBool(),
										"hashed":  columnElement.Attributes()["hashed"].(basetypes.BoolValue).ValueBool(),
									}
									columns = append(columns, column)
								}
							}
							table["column"] = columns
						}

					}
				}
				schema["table"] = tables
			}
		}
	}
	return schemas
}

func (d *ConnectorSchemaResourceModel) mapLocalSchemas() map[string]interface{} {
	schemas := d.getLegacySchemas()
	if len(schemas) == 0 {
		schemas = d.getSchemas()
	}
	mapColumns := func(columns []interface{}) map[string]interface{} {
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
	mapTables := func(tables []interface{}) map[string]interface{} {
		mappedTables := map[string]interface{}{}
		for _, lt := range tables {
			ltMap := lt.(map[string]interface{})
			mappedTable := map[string]interface{}{}
			for k, v := range ltMap {
				mappedTable[k] = v
			}
			if columns, ok := ltMap["column"].([]interface{}); ok {
				mappedTable["column"] = mapColumns(columns)
			} else {
				if columns, ok = ltMap["columns"].([]interface{}); ok {
					mappedTable["column"] = mapColumns(columns)
				}
			}

			mappedTables[ltMap["name"].(string)] = mappedTable
		}
		return mappedTables
	}
	mappedSchemas := map[string]interface{}{}
	for _, ls := range schemas {
		lsMap := ls.(map[string]interface{})
		mappedSchema := map[string]interface{}{}
		for k, v := range lsMap {
			mappedSchema[k] = v
		}
		if tables, ok := lsMap["table"].([]interface{}); ok {
			mappedSchema["table"] = mapTables(tables)
		} else {
			if tables, ok = lsMap["tables"].([]interface{}); ok {
				mappedSchema["table"] = mapTables(tables)
			}
		}

		mappedSchemas[lsMap["name"].(string)] = mappedSchema
	}
	return mappedSchemas
}
