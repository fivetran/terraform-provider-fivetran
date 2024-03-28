package mock

import (
	"testing"
)

// Test checks the following behavior: if no column settings defined for enabled table in BLOCK_ALL mode - do not manage columns for the table
// existing tables will save settings, new tables will be ignored
func TestSchemaDoesntTouchColumnsInBlockAllIfNoColumnSettingsMock(t *testing.T) {
	upstreamConfig := schemaConfigTestData{
		schemaChangeHandling: "ALLOW_ALL",
	}
	schema_1 := upstreamConfig.newSchema("schema_1", true)

	schema_1.newTable("table_1", true, nil).
		newColumnLocked("column_1", true, boolPtr(false)).
		newColumn("column_2", true, boolPtr(false)).
		newColumn("column_3", true, boolPtr(true))

	schema_1.newTable("table_2", true, nil).
		newColumnLocked("column_1", true, boolPtr(false)).
		newColumn("column_2", true, boolPtr(false)).
		newColumn("column_3", true, boolPtr(true))

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
		newColumnLocked("column_1", true, boolPtr(false)).
		newColumn("column_2", true, boolPtr(false)).
		newColumn("column_3", true, boolPtr(true))

	// table_2 enabled, existing columns saved settings
	schema_1response.newTable("table_2", false, nil).
		newColumnLocked("column_1", true, boolPtr(false)).
		newColumn("column_2", true, boolPtr(false)).
		newColumn("column_3", true, boolPtr(true))

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
		newColumnLocked("column_1", true, boolPtr(false)).
		newColumn("column_2", true, boolPtr(false)).
		newColumn("column_3", true, boolPtr(true))

	schema_1.newTableLocked("table_locked", true, nil)

	tfConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	tfConfig.newSchema("schema_1", true).
		newTable("table_1", true, nil).
		newColumn("column_2", true, nil)

	responseConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	schema_1response := responseConfig.newSchema("schema_1", true)
	schema_1response.newTable("table_1", true, nil).
		newColumnLocked("column_1", true, boolPtr(false)).
		newColumn("column_2", true, boolPtr(false)).
		newColumn("column_3", false, boolPtr(true))
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
	assertEqual(t, len(column3), 1)
	assertKeyExistsAndHasValue(t, column3, "enabled", false)
}

func TestIgnoreNoPatchAllowedColumnsMock(t *testing.T) {
	upstreamConfig := schemaConfigTestData{
		schemaChangeHandling: "ALLOW_ALL",
	}
	upstreamConfig.newSchema("schema_1", true).
		newTable("table_1", true, nil).
		newColumn("column_1", true, boolPtr(false)).
		newColumn("column_2", true, boolPtr(false)).
		newColumn("column_3", true, boolPtr(true))

	tfConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}

	tfConfig.newSchema("schema_1", true).
		newTable("table_1", true, nil).
		newColumn("column_2", true, nil)

	responseConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}
	responseConfig.newSchema("schema_1", true).
		newTable("table_1", true, nil).
		newColumn("column_1", false, boolPtr(false)).
		newColumn("column_2", true, boolPtr(false)).
		newColumn("column_3", false, boolPtr(true))

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
	assertEqual(t, len(column_1), 1)
	assertKeyExistsAndHasValue(t, column_1, "enabled", false)
	column_3 := assertKeyExists(t, columns, "column_3").(map[string]interface{})
	assertEqual(t, len(column_3), 1)
	assertKeyExistsAndHasValue(t, column_3, "enabled", false)
}

func TestConsistentWithUpstreamSchemaNoPatchMock(t *testing.T) {
	upstreamConfig := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}
	schema_1 := upstreamConfig.newSchema("schema_1", true)
	schema_1.newTable("table_1", true, stringPtr("SOFT_DELETE"))
	schema_1.newTable("table_2", true, stringPtr("LIVE"))
	schema_1.newTable("table_3", false, stringPtr("LIVE"))

	step1Config := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
	}
	step1Config_schema_1 := step1Config.newSchema("schema_1", true)
	step1Config_schema_1.newTable("table_1", true, stringPtr("SOFT_DELETE"))
	step1Config_schema_1.newTable("table_2", true, nil)

	step2Config := schemaConfigTestData{
		schemaChangeHandling: "BLOCK_ALL",
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
