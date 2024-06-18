package mock

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSchemaDisableColumnMissingInSchemaResponseMock(t *testing.T) {

	var schemaData map[string]interface{}

	var getHandler *mock.Handler
	var patchHandler *mock.Handler

	var schemasWoColumnsJsonResponse = `
{
	"enable_new_by_default": true,
	"schemas": {
		"schema_1": {
			"name_in_destination": "schema_1",
			"enabled": true,
			"tables": {
				"table_1": {
					"name_in_destination": "table_1",
					"sync_mode": "LIVE",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					},
					"columns": {
						"column_2": {
							"name_in_destination": "column_2",
							"enabled": true,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						}
					}
				}
			}
		}
	},
	"schema_change_handling": "%v"
}
	`

	var schemaWithColumnJsonResponse = `
{
	"schema_change_handling": "ALLOW_COLUMNS",
	"schemas": {
		"schema_1": {
			"name_in_destination": "schema_1",
			"enabled": true,
			"tables": {
				"table_1": {
					"name_in_destination": "table_1",
					"enabled": true,
					"sync_mode": "SOFT_DELETE",
					"enabled_patch_settings": {
						"allowed": true
					},
					"columns": {
						"column_1": {
							"name_in_destination": "column_1",
							"enabled": false,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						},
						"column_2": {
							"name_in_destination": "column_2",
							"enabled": true,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						}
					}
				}
			}
		}
	}
}
	`

	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				
				connector_id = "connector_id"
				schema_change_handling = "ALLOW_COLUMNS"
				schema {
					name = "schema_1"
					table {
						name = "table_1"
						enabled = true
						column {
							name = "column_1"
							enabled = false
						}
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, getHandler.Interactions, 3)   // 1 read attempt before reload, 1 read after create
				assertEqual(t, patchHandler.Interactions, 1) // Update SCM and align schema
				assertNotEmpty(t, schemaData)                // schema initialised
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_COLUMNS"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.0.column.0.enabled", "false"),
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
							schemaData = createMapFromJsonString(t, fmt.Sprintf(schemasWoColumnsJsonResponse, "ALLOW_ALL"))
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)

				patchHandler = mockClient.When(http.MethodPatch, "/v1/connectors/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := requestBodyToJson(t, req)

						assertEqual(t, len(body), 2)
						assertEqual(t, body["schema_change_handling"], "ALLOW_COLUMNS")

						assertKeyExists(t, body, "schemas")
						schemas := body["schemas"].(map[string]interface{})

						assertKeyExists(t, schemas, "schema_1")
						schema := schemas["schema_1"].(map[string]interface{})

						AssertKeyDoesNotExist(t, schema, "enabled")
						assertKeyExists(t, schema, "tables")
						tables := schema["tables"].(map[string]interface{})

						assertKeyExists(t, tables, "table_1")
						table := tables["table_1"].(map[string]interface{})

						assertKeyExists(t, table, "columns")
						AssertKeyDoesNotExist(t, table, "enabled")
						columns := table["columns"].(map[string]interface{})

						assertKeyExists(t, columns, "column_1")
						column := columns["column_1"].(map[string]interface{})

						assertKeyExistsAndHasValue(t, column, "enabled", false)

						// create schema structure
						schemaData = createMapFromJsonString(t, schemaWithColumnJsonResponse)

						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)
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
