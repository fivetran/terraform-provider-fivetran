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

	var columnsListResponse = `
{
	"columns": {
		"column_1": {
			"name_in_destination": "column_1",
			"enabled": true,
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
`

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
				assertEqual(t, getHandler.Interactions, 2)   // 1 read attempt before reload, 1 read after create
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

				getHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if nil == schemaData {
							schemaData = createMapFromJsonString(t, fmt.Sprintf(schemasWoColumnsJsonResponse, "ALLOW_ALL"))
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)

				mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas/schema_1/tables/table_1/columns").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, columnsListResponse)), nil
					},
				)

				patchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := requestBodyToJson(t, req)

						assertEqual(t, len(body), 2)
						assertEqual(t, body["schema_change_handling"], "ALLOW_COLUMNS")

						schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})

						schema := assertKeyExists(t, schemas, "schema_1").(map[string]interface{})

						AssertKeyDoesNotExist(t, schema, "enabled")
						tables := assertKeyExists(t, schema, "tables").(map[string]interface{})

						table := assertKeyExists(t, tables, "table_1").(map[string]interface{})

						columns := assertKeyExists(t, table, "columns").(map[string]interface{})
						AssertKeyDoesNotExist(t, table, "enabled")

						column := assertKeyExists(t, columns, "column_1").(map[string]interface{})

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

func TestResourceSchemaEmptyColumnsListMock(t *testing.T) {

	var schemaData map[string]interface{}

	var getHandler *mock.Handler
	var patchHandler *mock.Handler

	var columnsListResponse = `
{
	"columns": {
		"column_1": {
			"name_in_destination": "column_1",
			"enabled": true,
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
}`
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
						"column_1": {
							"name_in_destination": "column_1",
							"enabled": true,
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
						},
						"column_3": {
							"name_in_destination": "column_3",
							"enabled": true,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						}
					}
				},
				"table_2": {
					"name_in_destination": "table_2",
					"sync_mode": "LIVE",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					},
					"columns": {
						"column_2_1": {
							"name_in_destination": "column_2_1",
							"enabled": false,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						},
						"column_2_2": {
							"name_in_destination": "column_2_2",
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
}`
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
							"enabled": false,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						},
						"column_3": {
							"name_in_destination": "column_3",
							"enabled": true,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						}
					}
				},
				"table_2": {
					"name_in_destination": "table_2",
					"enabled": true,
					"sync_mode": "SOFT_DELETE",
					"enabled_patch_settings": {
						"allowed": true
					},
					"columns": {
						"column_2_1": {
							"name_in_destination": "column_2_1",
							"enabled": false,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						},
						"column_2_2": {
							"name_in_destination": "column_2_2",
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
}`

	step1 := resource.TestStep{
		Config: `
			locals {
				tables = {
					"table_1" = {
						name = "table_1"
						disabled_columns      = ["column_1", "column_2"]
					}
					"table_2" = {
						name = "table_2"
						disabled_columns      = [] # empty list
					}
				}
			}

			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider

				connector_id           = "connector_id"
				schema_change_handling = "ALLOW_COLUMNS"

				schemas = {
					"schema_1" = {
						enabled = true
						tables = {
							for table in local.tables : table.name => {
								enabled   = true
								sync_mode = "SOFT_DELETE"
								columns = {
									for column in table.disabled_columns : column => {
										enabled = false
									}
								}
							}
						}
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, getHandler.Interactions, 2)   // 1 read attempt before reload, 1 read after create
				assertEqual(t, patchHandler.Interactions, 1) // Update SCM and align schema
				assertNotEmpty(t, schemaData)                // schema initialised
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_COLUMNS"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_1.enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_1.sync_mode", "SOFT_DELETE"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_2.enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_2.sync_mode", "SOFT_DELETE"),

			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_1.columns.%", "2"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_1.columns.column_1.enabled", "false"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_1.columns.column_2.enabled", "false"),

			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_2.columns.%", "0"),
		),
	}

	step2 := resource.TestStep{
		RefreshState: true,
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				schemaData = nil

				getHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if nil == schemaData {
							schemaData = createMapFromJsonString(t, fmt.Sprintf(schemasWoColumnsJsonResponse, "ALLOW_ALL"))
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)

				mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas/schema_1/tables/table_1/columns").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, columnsListResponse)), nil
					},
				)

				patchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := 
						requestBodyToJson(t, req)

						assertEqual(t, body["schema_change_handling"], "ALLOW_COLUMNS")

						schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})

						schema := assertKeyExists(t, schemas, "schema_1").(map[string]interface{})

						AssertKeyDoesNotExist(t, schema, "enabled")
						tables := assertKeyExists(t, schema, "tables").(map[string]interface{})

						assertKeyExists(t, tables, "table_1")
						assertKeyExists(t, tables, "table_2")

						// create schema structure
						schemaData = createMapFromJsonString(t, schemaWithColumnJsonResponse)

						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				// there is no possibility to destroy schema config - it always exists within the connector
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}

func TestResourceSchemaEmptyColumnsListOldSchemaMock(t *testing.T) {

	var schemaData map[string]interface{}

	var getHandler *mock.Handler
	var patchHandler *mock.Handler

	var columnsListResponse = `
{
	"columns": {
		"column_1": {
			"name_in_destination": "column_1",
			"enabled": true,
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
}`
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
						"column_1": {
							"name_in_destination": "column_1",
							"enabled": true,
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
						},
						"column_3": {
							"name_in_destination": "column_3",
							"enabled": true,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						}
					}
				},
				"table_2": {
					"name_in_destination": "table_2",
					"sync_mode": "LIVE",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					},
					"columns": {
						"column_2_1": {
							"name_in_destination": "column_2_1",
							"enabled": false,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						},
						"column_2_2": {
							"name_in_destination": "column_2_2",
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
}`
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
							"enabled": false,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						},
						"column_3": {
							"name_in_destination": "column_3",
							"enabled": true,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						}
					}
				},
				"table_2": {
					"name_in_destination": "table_2",
					"enabled": true,
					"sync_mode": "SOFT_DELETE",
					"enabled_patch_settings": {
						"allowed": true
					},
					"columns": {
						"column_2_1": {
							"name_in_destination": "column_2_1",
							"enabled": false,
							"hashed": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						},
						"column_2_2": {
							"name_in_destination": "column_2_2",
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
}`

	step1 := resource.TestStep{
		Config: `
			locals {
				tables = {
					"table_1" = {
						name = "table_1"
						disabled_columns      = ["column_1", "column_2"]
					}
					"table_2" = {
						name = "table_2"
						disabled_columns      = [] # empty list
					}
				}
			}

			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider

				connector_id           = "connector_id"
				schema_change_handling = "ALLOW_COLUMNS"

				schema {
					name = "schema_1"
					dynamic "table" {
						for_each = local.tables
						iterator = tables
						content {
							name      = tables.key
							sync_mode = "SOFT_DELETE"
							enabled   = true
							dynamic "column" {
								for_each = tables.value.disabled_columns
								content {
									name    = column.value
									enabled = false
								}
							}
						}
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, getHandler.Interactions, 2)   // 1 read attempt before reload, 1 read after create
				assertEqual(t, patchHandler.Interactions, 1) // Update SCM and align schema
				assertNotEmpty(t, schemaData)                // schema initialised
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_COLUMNS"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.0.enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.0.sync_mode", "SOFT_DELETE"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.1.enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.1.sync_mode", "SOFT_DELETE"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				schemaData = nil

				getHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if nil == schemaData {
							schemaData = createMapFromJsonString(t, fmt.Sprintf(schemasWoColumnsJsonResponse, "ALLOW_ALL"))
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)

				mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas/schema_1/tables/table_1/columns").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, columnsListResponse)), nil
					},
				)

				patchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := 
						requestBodyToJson(t, req)

						assertEqual(t, body["schema_change_handling"], "ALLOW_COLUMNS")

						schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})

						schema := assertKeyExists(t, schemas, "schema_1").(map[string]interface{})

						AssertKeyDoesNotExist(t, schema, "enabled")
						tables := assertKeyExists(t, schema, "tables").(map[string]interface{})

						assertKeyExists(t, tables, "table_1")
						assertKeyExists(t, tables, "table_2")

						// create schema structure
						schemaData = createMapFromJsonString(t, schemaWithColumnJsonResponse)

						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				// there is no possibility to destroy schema config - it always exists within the connector
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

// TestResourceSchemaDisabledSchemasWithUnexpectedTablesInResponseMock reproduces GitHub Issue #589.
//
// When schemas are configured with enabled = false and no tables defined locally,
// but the Fivetran API response returns those schemas with tables after PATCH,
// the provider was incorrectly including those tables in state, causing:
//
//	"Provider produced inconsistent result after apply:
//	 .schemas["SCHEMA3"].tables: new element "SUPPLIER" has appeared."
func TestResourceSchemaDisabledSchemasWithUnexpectedTablesInResponseMock(t *testing.T) {

	var schemaData map[string]interface{}

	var getHandler *mock.Handler
	var patchHandler *mock.Handler
	var postSchemasReloadHandler *mock.Handler

	// Initial upstream: all schemas enabled
	var initialSchemaResponse = `
{
	"schema_change_handling": "ALLOW_ALL",
	"schemas": {
		"SOURCE": {
			"name_in_destination": "schema1_source",
			"enabled": true,
			"tables": {
				"ORDERS": {
					"name_in_destination": "orders",
					"sync_mode": "SOFT_DELETE",
					"enabled": true,
					"supports_columns_config": true,
					"supports_history_mode": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				},
				"DELIVERIES": {
					"name_in_destination": "deliveries",
					"sync_mode": "SOFT_DELETE",
					"enabled": true,
					"supports_columns_config": true,
					"supports_history_mode": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				}
			}
		},
		"SCHEMA2": {
			"name_in_destination": "schema1_schema2",
			"enabled": true,
			"tables": {
				"SALES": {
					"name_in_destination": "sales",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				}
			}
		},
		"SCHEMA3": {
			"name_in_destination": "schema1_schema3",
			"enabled": true,
			"tables": {
				"SUPPLIER": {
					"name_in_destination": "supplier",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				}
			}
		}
	}
}`

	// After PATCH: SCHEMA2 and SCHEMA3 are disabled, but the API still returns their tables.
	var afterPatchSchemaResponse = `
{
	"schema_change_handling": "ALLOW_COLUMNS",
	"schemas": {
		"SOURCE": {
			"name_in_destination": "schema1_source",
			"enabled": true,
			"tables": {
				"ORDERS": {
					"name_in_destination": "orders",
					"sync_mode": "SOFT_DELETE",
					"enabled": true,
					"supports_columns_config": true,
					"supports_history_mode": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				},
				"DELIVERIES": {
					"name_in_destination": "deliveries",
					"sync_mode": "SOFT_DELETE",
					"enabled": false,
					"supports_columns_config": true,
					"supports_history_mode": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				}
			}
		},
		"SCHEMA2": {
			"name_in_destination": "schema1_schema2",
			"enabled": false,
			"tables": {
				"SALES": {
					"name_in_destination": "sales",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				}
			}
		},
		"SCHEMA3": {
			"name_in_destination": "schema1_schema3",
			"enabled": false,
			"tables": {
				"SUPPLIER": {
					"name_in_destination": "supplier",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				}
			}
		}
	}
}`

	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider

				connector_id           = "connector_id"
				schema_change_handling = "ALLOW_COLUMNS"

				schemas = {
					"SOURCE" = {
						enabled = true
						tables = {
							"ORDERS" = {
								enabled = true
							}
						}
					}
					"SCHEMA2" = {
						enabled = false,
						tables = {
							"SALES" = {
								enabled = false
							}
						}
					}
					"SCHEMA3" = {
						enabled = false
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, getHandler.Interactions, 2)
				assertEqual(t, postSchemasReloadHandler.Interactions, 1)
				assertEqual(t, patchHandler.Interactions, 1)
				assertNotEmpty(t, schemaData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_COLUMNS"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.SOURCE.enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.SOURCE.tables.ORDERS.enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.SCHEMA2.enabled", "false"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.SCHEMA3.enabled", "false"),
			
			resource.TestCheckNoResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.SOURCE.tables.DELIVERIES"),
			resource.TestCheckNoResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.SCHEMA3.tables.SUPPLIER"),
			resource.TestCheckNoResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.SCHEMA2.tables.SALES"),
		),
	}

	// Refresh to ensure the state remains stable (no perpetual diff after initial apply).
	step2 := resource.TestStep{
		RefreshState: true,
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				schemaData = nil

				getHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if nil == schemaData {
							return fivetranResponse(t, req,
								"NotFound_SchemaConfig", http.StatusNotFound,
								"Connector with id 'connector_id' doesn't have schema config", nil), nil
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)

				postSchemasReloadHandler = mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/reload").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						schemaData = createMapFromJsonString(t, initialSchemaResponse)
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)

				patchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := requestBodyToJson(t, req)

						assertKeyExistsAndHasValue(t, body, "schema_change_handling", "ALLOW_COLUMNS")
						schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})

						schemaSource := assertKeyExists(t, schemas, "SOURCE").(map[string]interface{})
						tables := assertKeyExists(t, schemaSource, "tables").(map[string]interface{})
						assertKeyExistsAndHasValue(t, tables["DELIVERIES"].(map[string]interface{}), "enabled", false)

						schema2 := assertKeyExists(t, schemas, "SCHEMA2").(map[string]interface{})
						assertKeyExistsAndHasValue(t, schema2, "enabled", false)

						schema3 := assertKeyExists(t, schemas, "SCHEMA3").(map[string]interface{})
						assertKeyExistsAndHasValue(t, schema3, "enabled", false)

						schemaData = createMapFromJsonString(t, afterPatchSchemaResponse)
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				// Schema config always exists within the connector — nothing to destroy.
				return nil
			},
			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}