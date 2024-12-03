package mock

import (
	"testing"
)

func TestUpstreamSchemaWithoutColumns(t *testing.T) {
	// initially schema doesn't contain any columns
	upstreamConfig := schemaConfigTestData{
		schemaChangeHandling: "ALLOW_ALL",
	}
	schema_1 := upstreamConfig.newSchema("schema_1", true)
	schema_1.newTable("table_1", true, nil)
	schema_1.newTable("table_2", true, nil)
	schema_1.newTable("table_3", true, nil)

	schema_2 := upstreamConfig.newSchema("schema_2", true)
	schema_2.newTable("table_1", true, nil)
	schema_2.newTable("table_2", true, nil)
	schema_2.newTable("table_3", true, nil)

	tfConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	// Enable only two tables
	tfConfig.newSchema("schema_1", true).
		newTable("table_2", true, nil)

	tfConfig.newSchema("schema_2", true).
		newTable("table_2", true, nil)

	responseConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	schema_1response := responseConfig.newSchema("schema_1", true)
	schema_1response.newTable("table_1", false, nil)
	// as table_2 stays enabled on switch to BLOCK_ALL - column settings are fetched from source and saved in config
	schema_1response.newTable("table_2", true, nil).
		newColumn("column_1", true, boolPtr(false), true).
		newColumn("column_2", true, boolPtr(false), false)
	schema_1response.newTable("table_3", false, nil)

	schema_2response := responseConfig.newSchema("schema_2", true)
	schema_2response.newTable("table_1", false, nil)
	// as table_2 stays enabled on switch to BLOCK_ALL - column settings are fetched from source and saved in config
	schema_2response.newTable("table_2", true, nil).
		newColumn("column_1", true, boolPtr(false), true).
		newColumn("column_2", true, boolPtr(false), false)
	schema_2response.newTable("table_3", false, nil)

	body := setupOneStepTest(t, upstreamConfig, tfConfig, responseConfig)

	assertKeyExistsAndHasValue(t, body, "schema_change_handling", "BLOCK_ALL")
	schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})

	schema1 := assertKeyExists(t, schemas, "schema_1").(map[string]interface{})
	assertEqual(t, len(schema1), 1)
	tables := assertKeyExists(t, schema1, "tables").(map[string]interface{})
	assertEqual(t, len(tables), 2)
	table11 := AssertKeyExists(t, tables, "table_1").(map[string]interface{})
	assertEqual(t, len(table11), 1)
	assertKeyExistsAndHasValue(t, table11, "enabled", false)
	table13 := AssertKeyExists(t, tables, "table_3").(map[string]interface{})
	assertEqual(t, len(table13), 1)
	assertKeyExistsAndHasValue(t, table13, "enabled", false)

	schema2 := assertKeyExists(t, schemas, "schema_2").(map[string]interface{})
	assertEqual(t, len(schema2), 1)
	tables = assertKeyExists(t, schema2, "tables").(map[string]interface{})
	assertEqual(t, len(tables), 2)
	table21 := AssertKeyExists(t, tables, "table_1").(map[string]interface{})
	assertEqual(t, len(table21), 1)
	assertKeyExistsAndHasValue(t, table21, "enabled", false)
	table23 := AssertKeyExists(t, tables, "table_3").(map[string]interface{})
	assertEqual(t, len(table23), 1)
	assertKeyExistsAndHasValue(t, table23, "enabled", false)
}

func TestUpstreamSchemaWithoutColumnsColumnConfigured(t *testing.T) {
	// initially schema doesn't contain any columns
	upstreamConfig := schemaConfigTestData{
		schemaChangeHandling: "ALLOW_ALL",
	}
	upstreamConfig.newSchema("schema_1", true).newTable("table_1", true, nil)

	tfConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	// Enable only two tables
	tfConfig.newSchema("schema_1", true).
		newTable("table_1", true, nil).
		newColumn("column_1", true, nil, true)

	responseConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	responseConfig.newSchema("schema_1", true).
		newTable("table_1", true, nil).
		newColumn("column_1", true, boolPtr(false), false). // column user configured in tf
		newColumn("column_2", true, boolPtr(false), false)  // column present in source, but not saved to standard config before switch to BA mode

	response2Config := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	response2Config.newSchema("schema_1", true).
		newTable("table_1", true, nil).
		newColumn("column_1", true, boolPtr(false), false). // column user configured in tf
		newColumn("column_2", false, boolPtr(false), false) // column set disbled after second patch

	bodies := setupComplexTestWithColumnsReload(
		t, upstreamConfig,
		[]schemaConfigTestData{tfConfig},
		[]schemaConfigTestData{responseConfig, response2Config},
		map[string]map[string][]columnsConfigTestData{
			"schema_1": map[string][]columnsConfigTestData{
				"table_1": []columnsConfigTestData{
					newColumnConfigTestData().newColumn("column_1", false, boolPtr(false), true).newColumn("column_2", false, boolPtr(false), false),
				},
			},
		})

	assertEqual(t, len(bodies), 2)

	body1 := bodies[0]

	assertKeyExistsAndHasValue(t, body1, "schema_change_handling", "BLOCK_ALL")
	schemas := assertKeyExists(t, body1, "schemas").(map[string]interface{})

	schema1 := assertKeyExists(t, schemas, "schema_1").(map[string]interface{})
	assertEqual(t, len(schema1), 1)
	tables := assertKeyExists(t, schema1, "tables").(map[string]interface{})
	assertEqual(t, len(tables), 1)
	table11 := AssertKeyExists(t, tables, "table_1").(map[string]interface{})
	assertEqual(t, len(table11), 1)

	columns := AssertKeyExists(t, table11, "columns").(map[string]interface{})

	assertEqual(t, len(columns), 1)
	column11 := AssertKeyExists(t, columns, "column_1").(map[string]interface{})
	assertEqual(t, len(column11), 2)
	assertKeyExistsAndHasValue(t, column11, "enabled", true)

	body2 := bodies[1]

	assertEqual(t, len(body2), 1)
	schemas = assertKeyExists(t, body2, "schemas").(map[string]interface{})

	schema1 = assertKeyExists(t, schemas, "schema_1").(map[string]interface{})
	assertEqual(t, len(schema1), 1)
	tables = assertKeyExists(t, schema1, "tables").(map[string]interface{})
	assertEqual(t, len(tables), 1)
	table11 = AssertKeyExists(t, tables, "table_1").(map[string]interface{})
	assertEqual(t, len(table11), 1)

	columns = AssertKeyExists(t, table11, "columns").(map[string]interface{})
	assertEqual(t, len(columns), 1)
	column12 := AssertKeyExists(t, columns, "column_2").(map[string]interface{})
	assertEqual(t, len(column12), 2)
	assertKeyExistsAndHasValue(t, column12, "enabled", false)
}

// Test checks the following behavior: if no column settings defined for enabled table in BLOCK_ALL mode - do not manage columns for the table
// existing tables will save settings, new tables will be ignored
func TestSchemaDoesntTouchColumnsInBlockAllIfNoColumnSettingsMock(t *testing.T) {
	upstreamConfig := schemaConfigTestData{
		schemaChangeHandling: "ALLOW_ALL",
	}
	schema_1 := upstreamConfig.newSchema("schema_1", true)

	schema_1.newTable("table_1", true, nil).
		newColumnLocked("column_1", true, boolPtr(false), true).
		newColumn("column_2", true, boolPtr(false), false).
		newColumn("column_3", true, boolPtr(true), false)

	schema_1.newTable("table_2", true, nil).
		newColumnLocked("column_1", true, boolPtr(false), true).
		newColumn("column_2", true, boolPtr(false), false).
		newColumn("column_3", true, boolPtr(true), false)

	// only schema_1.table_1 will stay enabled
	// table_2 will be disabled
	// no column settings passed in request
	tfConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	tfConfig.newSchema("schema_1", true).
		newTable("table_1", true, nil)

	responseConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	schema_1response := responseConfig.newSchema("schema_1", true)
	// table_1 enabled, existing columns saved settings
	schema_1response.newTable("table_1", true, nil).
		newColumnLocked("column_1", true, boolPtr(false), true).
		newColumn("column_2", true, boolPtr(false), false).
		newColumn("column_3", true, boolPtr(true), false)

	// table_2 enabled, existing columns saved settings
	schema_1response.newTable("table_2", false, nil).
		newColumnLocked("column_1", true, boolPtr(false), true).
		newColumn("column_2", true, boolPtr(false), false).
		newColumn("column_3", true, boolPtr(true), false)

	// act
	body := setupOneStepTest(t, upstreamConfig, tfConfig, responseConfig)

	// assert
	assertKeyExistsAndHasValue(t, body, "schema_change_handling", "BLOCK_ALL")
	schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})
	assertEqual(t, len(schemas), 1)
	schema1 := assertKeyExists(t, schemas, "schema_1").(map[string]interface{})
	tables := assertKeyExists(t, schema1, "tables").(map[string]interface{})
	assertEqual(t, len(tables), 1)
	table2 := assertKeyExists(t, tables, "table_2").(map[string]interface{})
	assertEqual(t, len(table2), 1)
	assertKeyExistsAndHasValue(t, table2, "enabled", false)
}

func TestSetupSchemaBlockAllMock(t *testing.T) {
	upstreamConfig := schemaConfigTestData{
		schemaChangeHandling: "ALLOW_ALL",
	}
	schema_1 := upstreamConfig.newSchema("schema_1", true)

	schema_1.newTable("table_1", true, nil).
		newColumnLocked("column_1", true, boolPtr(false), true).
		newColumn("column_2", true, boolPtr(false), false).
		newColumn("column_3", true, boolPtr(true), false)

	schema_1.newTableLocked("table_locked", true, nil)

	tfConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	tfConfig.newSchema("schema_1", true).
		newTable("table_1", true, nil).
		newColumn("column_2", true, nil, true)

	responseConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	schema_1response := responseConfig.newSchema("schema_1", true)
	schema_1response.newTable("table_1", true, nil).
		newColumnLocked("column_1", true, boolPtr(false), true).
		newColumn("column_2", true, boolPtr(false), false).
		newColumn("column_3", false, boolPtr(true), false)
	schema_1response.newTableLocked("table_locked", true, nil)

	// act
	body := setupOneStepTest(t, upstreamConfig, tfConfig, responseConfig)

	// assert
	assertKeyExistsAndHasValue(t, body, "schema_change_handling", "BLOCK_ALL")
	schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})
	assertEqual(t, len(schemas), 1)
	schema1 := assertKeyExists(t, schemas, "schema_1").(map[string]interface{})
	tables := assertKeyExists(t, schema1, "tables").(map[string]interface{})

	assertEqual(t, len(tables), 1)
	table1 := assertKeyExists(t, tables, "table_1").(map[string]interface{})
	columns := assertKeyExists(t, table1, "columns").(map[string]interface{})

	assertEqual(t, len(columns), 1)
	column3 := assertKeyExists(t, columns, "column_3").(map[string]interface{})
	assertEqual(t, len(column3), 2)
	assertKeyExistsAndHasValue(t, column3, "enabled", false)
}

func TestIgnoreNoPatchAllowedColumnsMock(t *testing.T) {
	upstreamConfig := schemaConfigTestData{
		schemaChangeHandling: "ALLOW_ALL",
	}
	upstreamConfig.newSchema("schema_1", true).
		newTable("table_1", true, nil).
		newColumn("column_1", true, boolPtr(false), true).
		newColumn("column_2", true, boolPtr(false), false).
		newColumn("column_3", true, boolPtr(true), false)

	tfConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	tfConfig.newSchema("schema_1", true).
		newTable("table_1", true, nil).
		newColumn("column_2", true, nil, true)

	responseConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}
	responseConfig.newSchema("schema_1", true).
		newTable("table_1", true, nil).
		newColumn("column_1", false, boolPtr(false), true).
		newColumn("column_2", true, boolPtr(false), false).
		newColumn("column_3", false, boolPtr(true), false)

	// act
	body := setupOneStepTest(t, upstreamConfig, tfConfig, responseConfig)

	// assert
	assertKeyExistsAndHasValue(t, body, "schema_change_handling", "BLOCK_ALL")
	schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})
	assertEqual(t, len(schemas), 1)
	schema_1 := assertKeyExists(t, schemas, "schema_1").(map[string]interface{})
	tables := assertKeyExists(t, schema_1, "tables").(map[string]interface{})
	assertEqual(t, len(tables), 1)
	table_1 := assertKeyExists(t, tables, "table_1").(map[string]interface{})
	columns := assertKeyExists(t, table_1, "columns").(map[string]interface{})
	assertEqual(t, len(columns), 2)
	column_1 := assertKeyExists(t, columns, "column_1").(map[string]interface{})
	assertEqual(t, len(column_1), 2)
	assertKeyExistsAndHasValue(t, column_1, "enabled", false)
	column_3 := assertKeyExists(t, columns, "column_3").(map[string]interface{})
	assertEqual(t, len(column_3), 2)
	assertKeyExistsAndHasValue(t, column_3, "enabled", false)
}

func TestConsistentWithUpstreamSchemaNoPatchMock(t *testing.T) {
	upstreamConfig := schemaConfigTestData{
		schemaChangeHandling: "ALLOW_COLUMNS",
	}
	schema_1 := upstreamConfig.newSchema("schema_1", true)
	schema_1.newTable("table_1", true, stringPtr("SOFT_DELETE"))
	schema_1.newTable("table_2", true, stringPtr("LIVE"))
	schema_1.newTable("table_3", false, stringPtr("LIVE"))
	disabled_schema := upstreamConfig.newSchema("disabled_schema", false)
	disabled_schema.newTable("table_1", true, stringPtr("SOFT_DELETE"))

	step1Config := schemaConfigTestData{
		schemaChangeHandling: "ALLOW_COLUMNS",
	}
	step1Config_schema_1 := step1Config.newSchema("schema_1", true)
	step1Config_schema_1.newTable("table_1", true, stringPtr("SOFT_DELETE"))
	step1Config_schema_1.newTable("table_2", true, nil)

	step2Config := schemaConfigTestData{
		schemaChangeHandling: "ALLOW_COLUMNS",
	}
	step2Config_schema_1 := step2Config.newSchema("schema_1", true)
	step2Config_schema_1.newTable("table_1", true, nil)
	step2Config_schema_1.newTable("table_2", true, stringPtr("LIVE"))

	bodies := setupComplexTest(t, upstreamConfig,
		[]schemaConfigTestData{step1Config, step2Config},
		[]schemaConfigTestData{upstreamConfig, upstreamConfig})

	// no PATCH requests were done because configs are consistent with upstream
	assertEmpty(t, bodies)
}
