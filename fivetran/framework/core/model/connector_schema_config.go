package model

import (
	"context"
	"encoding/json"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/fivetrantypes"
	configSchema "github.com/fivetran/terraform-provider-fivetran/modules/connector/schema"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ConnectorSchemaResourceModel struct {
	Id                   types.String                  `tfsdk:"id"`
	ConnectorId          types.String                  `tfsdk:"connector_id"`
	GroupId              types.String                  `tfsdk:"group_id"`
	ConnectorName        types.String                  `tfsdk:"connector_name"`
	SchemaChangeHandling types.String                  `tfsdk:"schema_change_handling"`
	Schemas              types.Map                     `tfsdk:"schemas"`
	Schema               types.Set                     `tfsdk:"schema"`
	Timeouts             timeouts.Value                `tfsdk:"timeouts"`
	SchemasRaw           fivetrantypes.JsonSchemaValue `tfsdk:"schemas_json"`
	ValidationLevel      types.String                  `tfsdk:"validation_level"`
}

func (d *ConnectorSchemaResourceModel) IsValid() bool {
	noSchemaDefined := !(d.IsRawSchemaDefined() || d.IsMappedSchemaDefined() || d.IsLegacySchemaDefined())
	return noSchemaDefined || ((d.IsRawSchemaDefined() != d.IsMappedSchemaDefined()) != d.IsLegacySchemaDefined())
}

func (d *ConnectorSchemaResourceModel) getValidationLevels() (bool, bool) {
	validationLevel := d.ValidationLevel.ValueString()

	validateTables := false
	validateColumns := false

	if validationLevel != "NONE" {
		validateTables = true
	}
	if validationLevel == "COLUMNS" {
		validateColumns = true
	}
	return validateTables, validateColumns
}

func (d *ConnectorSchemaResourceModel) ValidateSchemaElements(response connections.ConnectionSchemaDetailsResponse, client fivetran.Client, ctx context.Context) (error, bool) {
	validateTables, validateColumns := d.getValidationLevels()
	if validateTables {
		return d.GetSchemaConfig().ValidateSchemas(
			d.ConnectorId.ValueString(),
			response.Data.Schemas,
			client,
			ctx,
			validateColumns)
	}
	return nil, false
}

func (d *ConnectorSchemaResourceModel) IsRawSchemaDefined() bool {
	return !d.SchemasRaw.IsUnknown() && !d.SchemasRaw.IsNull() && len(d.SchemasRaw.ValueString()) > 0
}

func (d *ConnectorSchemaResourceModel) IsLegacySchemaDefined() bool {
	return !d.Schema.IsUnknown() && !d.Schema.IsNull() && len(d.Schema.Elements()) > 0
}

func (d *ConnectorSchemaResourceModel) IsMappedSchemaDefined() bool {
	return !d.Schemas.IsUnknown() && !d.Schemas.IsNull() && len(d.Schemas.Elements()) > 0
}

func (d *ConnectorSchemaResourceModel) ReadFromResponse(response connections.ConnectionSchemaDetailsResponse, diag *diag.Diagnostics) {
	schemaObject := configSchema.SchemaConfig{}
	schemaObject.ReadFromResponse(response)
	schemas := schemaObject.GetSchemas(response.Data.SchemaChangeHandling, d.GetSchemaConfig(), diag)

	if d.IsLegacySchemaDefined() {
		d.Schema = d.getLegacySchemaItems(schemas)
		d.Schemas = d.getNullSchemas()
		d.SchemasRaw = fivetrantypes.NewJsonSchemaNull()
	}
	if d.IsMappedSchemaDefined() {
		d.Schema = d.getNullSchema()
		d.Schemas = d.getSchemasMap(schemas)
		d.SchemasRaw = fivetrantypes.NewJsonSchemaNull()
	}
	if d.IsRawSchemaDefined() {
		d.Schemas = d.getNullSchemas()
		d.Schema = d.getNullSchema()
		schemasJson := d.getSchemasRawValue(schemas)
		d.SchemasRaw = fivetrantypes.NewJsonSchemaValue(schemasJson)
	}

	// SAP connections will not accept schemaChangeHandling in request and will return ALLOW_COLUMNS as default
	if !d.SchemaChangeHandling.IsNull() && !d.SchemaChangeHandling.IsUnknown() && d.SchemaChangeHandling.ValueString() != "" {
		d.SchemaChangeHandling = types.StringValue(response.Data.SchemaChangeHandling)
	} else {
		d.SchemaChangeHandling = types.StringNull()
	}
}

func (d *ConnectorSchemaResourceModel) getNullSchema() basetypes.SetValue {
	columnAttrTypes := map[string]attr.Type{
		"name":           types.StringType,
		"enabled":        types.BoolType,
		"hashed":         types.BoolType,
		"is_primary_key": types.BoolType,
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
		"enabled":        types.BoolType,
		"hashed":         types.BoolType,
		"is_primary_key": types.BoolType,
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

	if d.IsLegacySchemaDefined() {
		result.ReadFromRawSourceData(d.getLegacySchemas(), d.SchemaChangeHandling.ValueString())
	}
	if d.IsMappedSchemaDefined() {
		result.ReadFromRawSourceData(d.getSchemas(), d.SchemaChangeHandling.ValueString())
	}
	if d.IsRawSchemaDefined() {
		result.ReadFromRawSourceData(d.getSchemasRaw(), d.SchemaChangeHandling.ValueString())
	}

	return result
}

func (d *ConnectorSchemaResourceModel) getSchemasRawValue(schemas []interface{}) string {
	result := mapRawSchemas(schemas)
	resultRawString, _ := json.Marshal(result)
	return string(resultRawString)
}

func (d *ConnectorSchemaResourceModel) getSchemasMap(schemas []interface{}) basetypes.MapValue {
	columnsAttrTypes := map[string]attr.Type{
		"enabled":        types.BoolType,
		"hashed":         types.BoolType,
		"is_primary_key": types.BoolType,
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

		if tableList, ok := schemaMap["table"]; ok {
			for _, t := range tableList.([]interface{}) {
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
				if columnList, ok := tableMap["column"]; ok {
					for _, c := range columnList.([]interface{}) {
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

						if columnMap["is_primary_key"] != nil {
							columnElements["is_primary_key"] = types.BoolValue(helpers.StrToBool(columnMap["is_primary_key"].(string)))
						} else {
							columnElements["is_primary_key"] = types.BoolNull()
						}
						columnValue, _ := types.ObjectValue(columnsAttrTypes, columnElements)
						columns[columnName] = columnValue
					}
				}
				if len(columns) > 0 {
					tableElements["columns"], _ = types.MapValue(types.ObjectType{AttrTypes: columnsAttrTypes}, columns)
				} else {
					tableElements["columns"] = types.MapNull(types.ObjectType{AttrTypes: columnsAttrTypes})
				}
				tableValue, _ := types.ObjectValue(tablesAttrTypes, tableElements)
				tables[tableName] = tableValue
			}
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
		"name":           types.StringType,
		"enabled":        types.BoolType,
		"hashed":         types.BoolType,
		"is_primary_key": types.BoolType,
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
		if tableList, ok := schemaMap["table"]; ok {
			for _, t := range tableList.([]interface{}) {
				tableMap := t.(map[string]interface{})
				tableName := tableMap["name"].(string)

				localTable := d.tryGetLocalTable(localSchema, tableName)

				columns := []attr.Value{}
				if columnList, ok := tableMap["column"]; ok {
					for _, c := range columnList.([]interface{}) {
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
						if columnMap["is_primary_key"] != nil {
							columnElements["is_primary_key"] = types.BoolValue(helpers.StrToBool(columnMap["is_primary_key"].(string)))
						} else {
							columnElements["is_primary_key"] = types.BoolNull()
						}
						columnValue, _ := types.ObjectValue(columnAttrTypes, columnElements)
						columns = append(columns, columnValue)
					}
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

							isPrimaryKey := columnElement.Attributes()["is_primary_key"].(basetypes.BoolValue)
							if !isPrimaryKey.IsUnknown() && !isPrimaryKey.IsNull() {
								column["is_primary_key"] = columnElement.Attributes()["is_primary_key"].(basetypes.BoolValue).ValueBool()
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

func (d *ConnectorSchemaResourceModel) getSchemasRaw() []interface{} {
	schemas := []interface{}{}
	rawSchemas := map[string]interface{}{}
	if e := json.Unmarshal([]byte(d.SchemasRaw.ValueString()), &rawSchemas); e == nil {
		for sName, si := range rawSchemas {
			schema := map[string]interface{}{
				"name": sName,
			}
			if sMap, ok := si.(map[string]interface{}); ok {
				if e, ok := sMap["enabled"].(bool); ok {
					schema["enabled"] = e
				}
				if t, ok := sMap["tables"]; ok {
					tables := []interface{}{}
					if rawTables, ok := t.(map[string]interface{}); ok {
						for tName, ti := range rawTables {
							table := map[string]interface{}{
								"name": tName,
							}
							if tMap, ok := ti.(map[string]interface{}); ok {
								if e, ok := tMap["enabled"].(bool); ok {
									table["enabled"] = e
								}
								if sm, ok := tMap["sync_mode"]; ok {
									table["sync_mode"] = sm
								}
								if c, ok := tMap["columns"]; ok {
									columns := []interface{}{}
									if rawColumns, ok := c.(map[string]interface{}); ok {
										for cName, ci := range rawColumns {
											column := map[string]interface{}{
												"name": cName,
											}
											if cMap, ok := ci.(map[string]interface{}); ok {
												if e, ok := cMap["enabled"].(bool); ok {
													column["enabled"] = e
												}
												if h, ok := cMap["hashed"].(bool); ok {
													column["hashed"] = h
												}
												if p, ok := cMap["is_primary_key"].(bool); ok {
													column["is_primary_key"] = p
												}
												columns = append(columns, column)
											}
										}
									}
									if len(columns) > 0 {
										table["column"] = columns
									}
								}
								tables = append(tables, table)
							}
						}
					}
					if len(tables) > 0 {
						schema["table"] = tables
					}
				}
				schemas = append(schemas, schema)
			}
		}
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
										//"hashed":  columnElement.Attributes()["hashed"].(basetypes.BoolValue).ValueBool(),
									}
									if !columnElement.Attributes()["hashed"].(basetypes.BoolValue).IsUnknown() {
										column["hashed"] = columnElement.Attributes()["hashed"].(basetypes.BoolValue).ValueBool()
									}
									if !columnElement.Attributes()["is_primary_key"].(basetypes.BoolValue).IsUnknown() {
										column["is_primary_key"] = columnElement.Attributes()["is_primary_key"].(basetypes.BoolValue).ValueBool()
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

func mapRawSchemas(schemas []interface{}) map[string]interface{} {
	columnKey := "columns"
	tableKey := "tables"
	mapColumns := func(columns []interface{}) map[string]interface{} {
		mappedColumns := map[string]interface{}{}
		for _, lc := range columns {
			lcMap := lc.(map[string]interface{})
			mappedColumn := map[string]interface{}{}
			for k, v := range lcMap {
				if k != "name" {
					mappedColumn[k] = v
				}
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
				if k != "name" && k != "column" && k != "columns" {
					mappedTable[k] = v
				}
			}
			if columns, ok := ltMap["column"].([]interface{}); ok {
				mappedTable[columnKey] = mapColumns(columns)
			} else {
				if columns, ok = ltMap["columns"].([]interface{}); ok {
					mappedTable[columnKey] = mapColumns(columns)
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
			if k != "name" && k != "table" && k != "tables" {
				mappedSchema[k] = v
			}
		}
		if tables, ok := lsMap["table"].([]interface{}); ok {
			mappedSchema[tableKey] = mapTables(tables)
		} else {
			if tables, ok = lsMap["tables"].([]interface{}); ok {
				mappedSchema[tableKey] = mapTables(tables)
			}
		}
		delete(mappedSchema, "table")
		mappedSchemas[lsMap["name"].(string)] = mappedSchema
	}
	return mappedSchemas
}

func mapLocalSchemas(schemas []interface{}) map[string]interface{} {
	columnKey := "column"
	tableKey := "table"
	mapColumns := func(columns []interface{}) map[string]interface{} {
		mappedColumns := map[string]interface{}{}
		for _, lc := range columns {
			lcMap := lc.(map[string]interface{})
			mappedColumn := map[string]interface{}{}
			for k, v := range lcMap {
				if k != "name" {
					mappedColumn[k] = v
				}
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
				if k != "name" {
					mappedTable[k] = v
				}
			}
			if columns, ok := ltMap["column"].([]interface{}); ok {
				mappedTable[columnKey] = mapColumns(columns)
			} else {
				if columns, ok = ltMap["columns"].([]interface{}); ok {
					mappedTable[columnKey] = mapColumns(columns)
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
			if k != "name" {
				mappedSchema[k] = v
			}
		}
		if tables, ok := lsMap["table"].([]interface{}); ok {
			mappedSchema[tableKey] = mapTables(tables)
		} else {
			if tables, ok = lsMap["tables"].([]interface{}); ok {
				mappedSchema[tableKey] = mapTables(tables)
			}
		}
		mappedSchemas[lsMap["name"].(string)] = mappedSchema
	}
	return mappedSchemas
}

func (d *ConnectorSchemaResourceModel) mapLocalSchemas() map[string]interface{} {
	schemas := d.getLegacySchemas()
	if len(schemas) == 0 {
		schemas = d.getSchemas()
	}
	return mapLocalSchemas(schemas)
}
