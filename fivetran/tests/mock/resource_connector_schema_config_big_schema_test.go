package mock

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceSchemaBigSchemaTestMock(t *testing.T) {

	schemasCount := 5
	tablesCount := 200

	generateJsonResponse := func(schemaCount, tableCount int, sch string) string {
		tableTempate :=
			`
				"%v": {
					"name_in_destination": "%v",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				}
		`
		schemaTemplate :=
			`
		"%v": {
			"name_in_destination": "%v",
			"enabled": true,
			"tables": {
%v
			}
		}
		`
		schemas := ""
		for si := 0; si < schemaCount; si++ {
			if si > 0 {
				schemas = schemas + ","
			}
			tables := ""
			for ti := 0; ti < tableCount; ti++ {
				if ti > 0 {
					tables = tables + ","
				}
				tName := fmt.Sprintf("table_%v", ti)
				tables = tables + fmt.Sprintf(tableTempate, tName, tName)
			}
			sName := fmt.Sprintf("schema_%v", si)
			schemas = schemas + fmt.Sprintf(schemaTemplate, sName, sName, tables)
		}
		result := fmt.Sprintf(
			`
{
	"enable_new_by_default": false,
	"schemas": {
%v
	},
	"schema_change_handling": "%v"
}
			`,
			schemas,
			sch,
		)
		fmt.Println("Response generation done")
		return result
	}

	generateTfConfig := func(schemaCount, tableCount int, sch string) string {
		tableTempate :=
			`
			"%v" = {
				enabled = true
			}
		`
		schemaTemplate :=
			`
	"%v" = {
		enabled = true
		tables = {
%v
		}
	}
		`
		schemas := ""
		for si := 0; si < schemaCount; si++ {
			tables := ""
			for ti := 0; ti < tableCount; ti++ {
				tName := fmt.Sprintf("table_%v", ti)
				tables = tables + fmt.Sprintf(tableTempate, tName)
			}
			sName := fmt.Sprintf("schema_%v", si)
			schemas = schemas + fmt.Sprintf(schemaTemplate, sName, tables)
		}
		result := fmt.Sprintf(
			`
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	connector_id = "connector_id"
	schema_change_handling = "%v"
	schemas = {
%v
	}
}
			`,
			sch,
			schemas,
		)
		fmt.Println("Config generation done")
		return result
	}

	var schemaData map[string]interface{}

	var getHandler *mock.Handler
	//var patchHandler *mock.Handler

	step1 := resource.TestStep{
		Config: generateTfConfig(schemasCount, tablesCount, "BLOCK_ALL"),

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, getHandler.Interactions, 1)
				//assertEqual(t, patchHandler.Interactions, 1)
				assertNotEmpty(t, schemaData)
				return nil
			},
			//resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_COLUMNS"),
			//resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.0.column.0.enabled", "false"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				schemaData = nil

				getHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if nil == schemaData {
							schemaData = createMapFromJsonString(t, generateJsonResponse(schemasCount, tablesCount, "BLOCK_ALL"))
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)

				// patchHandler = mockClient.When(http.MethodPatch, "/v1/connectors/connector_id/schemas").ThenCall(
				// 	func(req *http.Request) (*http.Response, error) {
				// 		body := requestBodyToJson(t, req)

				// 		assertEqual(t, len(body), 2)
				// 		assertEqual(t, body["schema_change_handling"], "ALLOW_COLUMNS")

				// 		assertKeyExists(t, body, "schemas")
				// 		schemas := body["schemas"].(map[string]interface{})

				// 		assertKeyExists(t, schemas, "schema_1")
				// 		schema := schemas["schema_1"].(map[string]interface{})

				// 		AssertKeyDoesNotExist(t, schema, "enabled")
				// 		assertKeyExists(t, schema, "tables")
				// 		tables := schema["tables"].(map[string]interface{})

				// 		assertKeyExists(t, tables, "table_1")
				// 		table := tables["table_1"].(map[string]interface{})

				// 		assertKeyExists(t, table, "columns")
				// 		AssertKeyDoesNotExist(t, table, "enabled")
				// 		columns := table["columns"].(map[string]interface{})

				// 		assertKeyExists(t, columns, "column_1")
				// 		column := columns["column_1"].(map[string]interface{})

				// 		assertKeyExistsAndHasValue(t, column, "enabled", false)

				// 		// create schema structure
				// 		// schemaData = createMapFromJsonString(t, schemaWithColumnJsonResponse)

				// 		return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
				// 	},
				// )
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				// there is no possibility to destroy schema config - it alsways exists within the connector
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
