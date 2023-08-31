package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	schemaEmptyDefaultReloadHandler *mock.Handler
	schemaEmptyDefaultGetHandler    *mock.Handler
	schemaEmptyDefaultPatchHandler  *mock.Handler
	schemaEmptyDefaultData          map[string]interface{}

	schemaLockedGetHandler   *mock.Handler
	schemaLockedPatchHandler *mock.Handler
	schemaLockedData         map[string]interface{}

	schemaHashedAlignmentGetHandler   *mock.Handler
	schemaHashedAlignmentPatchHandler *mock.Handler
	schemaHashedAlignmentData         map[string]interface{}

	schemaConsistentWithUpstreamGetHandler    *mock.Handler
	schemaConsistentWithUpstreamPathchHandler *mock.Handler
	schemaConsistentWithUpstreamData          map[string]interface{}
)

const (
	schemaHashedColumnAlignmentJsonSchema = `
	{
		"enable_new_by_default": true,
		"schema_change_handling": "ALLOW_ALL",
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
								"enabled": true,
								"hashed": true,
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

	schemaWithLockedTableAndColumn = `
	{
		"schema_change_handling": "ALLOW_ALL",
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
								"enabled": true,
								"hashed": false,
								"enabled_patch_settings": {
									"allowed": false,
									"reason_code": "SYSTEM_COLUMN",
									"reason": "The column does not support exclusion as it is a Primary Key"
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
					},
					"table_2": {
						"name_in_destination": "table_2",
						"enabled": true,
						"sync_mode": "SOFT_DELETE",
						"enabled_patch_settings": {
							"allowed": false,
							"reason_code": "SYSTEM_TABLE"
						},
						"columns": {
							"column_3": {
								"name_in_destination": "column_3",
								"enabled": true,
								"hashed": false,
								"enabled_patch_settings": {
									"allowed": false,
									"reason_code": "SYSTEM_COLUMN",
									"reason": "The column does not support exclusion as it is a Primary Key"
								}
							},
							"column_4": {
								"name_in_destination": "column_4",
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
	schemaWithLockedTableAndColumn_blocked0 = `
	{
		"schema_change_handling": "BLOCK_ALL",
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
								"enabled": true,
								"hashed": false,
								"enabled_patch_settings": {
									"allowed": false,
									"reason_code": "SYSTEM_COLUMN",
									"reason": "The column does not support exclusion as it is a Primary Key"
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
					},
					"table_2": {
						"name_in_destination": "table_2",
						"enabled": true,
						"sync_mode": "SOFT_DELETE",
						"enabled_patch_settings": {
							"allowed": false,
							"reason_code": "SYSTEM_TABLE"
						},
						"columns": {
							"column_3": {
								"name_in_destination": "column_3",
								"enabled": true,
								"hashed": false,
								"enabled_patch_settings": {
									"allowed": false,
									"reason_code": "SYSTEM_COLUMN",
									"reason": "The column does not support exclusion as it is a Primary Key"
								}
							},
							"column_4": {
								"name_in_destination": "column_4",
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
	schemaWithLockedTableAndColumn_blocked1 = `
	{
		"schema_change_handling": "BLOCK_ALL",
		"schemas": {
			"schema_1": {
				"name_in_destination": "schema_1",
				"enabled": false,
				"tables": {
					"table_1": {
						"name_in_destination": "table_1",
						"enabled": false,
						"sync_mode": "SOFT_DELETE",
						"enabled_patch_settings": {
							"allowed": true
						},
						"columns": {
							"column_1": {
								"name_in_destination": "column_1",
								"enabled": true,
								"hashed": false,
								"enabled_patch_settings": {
									"allowed": false,
									"reason_code": "SYSTEM_COLUMN",
									"reason": "The column does not support exclusion as it is a Primary Key"
								}
							},
							"column_2": {
								"name_in_destination": "column_2",
								"enabled": false,
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
							"allowed": false,
							"reason_code": "SYSTEM_TABLE"
						},
						"columns": {
							"column_3": {
								"name_in_destination": "column_3",
								"enabled": true,
								"hashed": false,
								"enabled_patch_settings": {
									"allowed": false,
									"reason_code": "SYSTEM_COLUMN",
									"reason": "The column does not support exclusion as it is a Primary Key"
								}
							},
							"column_4": {
								"name_in_destination": "column_4",
								"enabled": false,
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

	schemaEmptyDefaultJsonSchema = `
	{
		"enable_new_by_default": true,
		"schema_change_handling": "ALLOW_ALL",
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
	schemaEmptyDefaultJsonBlockedSchema1 = `
	{
		"schema_change_handling": "BLOCK_ALL",
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
	schemaEmptyDefaultJsonBlockedSchema2 = `
	{
		"enable_new_by_default": false,
		"schema_change_handling": "BLOCK_ALL",
		"schemas": {
			"schema_1": {
				"name_in_destination": "schema_1",
				"enabled": false,
				"tables": {
					"table_1": {
						"name_in_destination": "table_1",
						"enabled": false,
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
							}
						}
					}
				}
			}
		}
	}
	`
	schemaEmptyDefaultJsonBlockedSchema3 = `
	{
		"enable_new_by_default": false,
		"schema_change_handling": "BLOCK_ALL",
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
)

func setupMockClientEmptyDefaultSchemaResource(t *testing.T) {
	mockClient.Reset()
	schemaEmptyDefaultData = nil

	updateIteration := 0

	schemaEmptyDefaultReloadHandler = mockClient.When(http.MethodPost, "/v1/connectors/connector_id/schemas/reload").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)

			// Check the request
			assertEqual(t, len(body), 1)
			assertEqual(t, body["exclude_mode"], "PRESERVE") // reload schema in PRESERVE mode

			// create schema structure
			schemaEmptyDefaultData = createMapFromJsonString(t, schemaEmptyDefaultJsonSchema)

			response := fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaEmptyDefaultData)

			return response, nil
		},
	)

	schemaEmptyDefaultGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if nil == schemaEmptyDefaultData {
				return fivetranResponse(t, req,
					"NotFound_SchemaConfig", http.StatusNotFound,
					"Connector with id 'connector_id' doesn't have schema config", nil), nil
			}

			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaEmptyDefaultData), nil
		},
	)

	schemaEmptyDefaultPatchHandler = mockClient.When(http.MethodPatch, "/v1/connectors/connector_id/schemas/").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)

			if updateIteration == 0 {
				// Check the request
				assertEqual(t, len(body), 1)
				assertEqual(t, body["schema_change_handling"], "BLOCK_ALL")

				// create schema structure
				schemaEmptyDefaultData = createMapFromJsonString(t, schemaEmptyDefaultJsonBlockedSchema1)
			}
			if updateIteration == 1 {
				// Check the request
				assertEqual(t, len(body), 1)
				assertNotEmpty(t, body["schemas"])

				schemasMap := body["schemas"].(map[string]interface{})

				schema1 := schemasMap["schema_1"].(map[string]interface{})

				assertEqual(t, len(schema1), 2)

				assertEqual(t, schema1["enabled"], false)
				assertNotEmpty(t, schema1["tables"])

				tablesMap := schema1["tables"].(map[string]interface{})

				table1 := tablesMap["table_1"].(map[string]interface{})

				assertEqual(t, len(table1), 2)

				assertEqual(t, table1["enabled"], false)
				assertNotEmpty(t, table1["columns"])

				columnsMap := table1["columns"].(map[string]interface{})

				column1 := columnsMap["column_1"].(map[string]interface{})

				assertEqual(t, column1["enabled"], false)

				// create schema structure
				schemaEmptyDefaultData = createMapFromJsonString(t, schemaEmptyDefaultJsonBlockedSchema2)
			}
			if updateIteration == 2 {
				// Check the request
				assertEqual(t, len(body), 1)
				assertNotEmpty(t, body["schemas"])

				schemasMap := body["schemas"].(map[string]interface{})

				schema1 := schemasMap["schema_1"].(map[string]interface{})

				assertEqual(t, len(schema1), 2)

				assertEqual(t, schema1["enabled"], true)
				assertNotEmpty(t, schema1["tables"])

				tablesMap := schema1["tables"].(map[string]interface{})

				table1 := tablesMap["table_1"].(map[string]interface{})

				assertEqual(t, len(table1), 2)

				assertEqual(t, table1["enabled"], true)
				assertNotEmpty(t, table1["columns"])

				columnsMap := table1["columns"].(map[string]interface{})

				column1 := columnsMap["column_1"].(map[string]interface{})

				assertEqual(t, column1["enabled"], true)

				// create schema structure
				schemaEmptyDefaultData = createMapFromJsonString(t, schemaEmptyDefaultJsonBlockedSchema3)
			}
			updateIteration++
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaEmptyDefaultData), nil
		},
	)
}

// This test checks the case when defined schema_change_handling mathches upstream state
// Schema isn't defined in config
// Resource should just reload schema, verify that upstream schema is aligned with config and then do nothing
// In step 2 schema_change_handling updated and schema config should be transformed according to state
// In step 3 aligned config added (this should not be detected in plan - no-drifts check)
// Step 4 applies effective config
func TestResourceEmptyDefaultSchemaMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "ALLOW_ALL"
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaEmptyDefaultReloadHandler.Interactions, 1) // 1 reload on create
				assertEqual(t, schemaEmptyDefaultGetHandler.Interactions, 2)    // 1 read attempt before reload, 1 read after create
				assertNotEmpty(t, schemaEmptyDefaultData)                       // schema initialised
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_ALL"),
		),
	}

	step2 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaEmptyDefaultPatchHandler.Interactions, 2)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
		),
	}

	step3 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
				schema {
					name = "schema_1"
					enabled = "false"
					table {
						name = "table_1"
						enabled = "false"
						column {
							name = "column_1"
							enabled = "false"
						}
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaEmptyDefaultPatchHandler.Interactions, 2)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
		),
	}

	step4 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
				schema {
					name = "schema_1"
					table {
						name = "table_1"
						column {
							name = "column_1"
							enabled = "true"
						}
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaEmptyDefaultPatchHandler.Interactions, 3)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientEmptyDefaultSchemaResource(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				// there is no possibility to destroy schema config - it alsways exists within the connector
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
				step4,
			},
		},
	)
}

func setupMockClientLockedPartsSchemaResource(t *testing.T) {
	mockClient.Reset()
	schemaLockedData = nil
	updateIteration := 0

	schemaLockedGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if nil == schemaLockedData {
				schemaLockedData = createMapFromJsonString(t, schemaWithLockedTableAndColumn)
			}
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaLockedData), nil
		},
	)

	schemaLockedPatchHandler = mockClient.When(http.MethodPatch, "/v1/connectors/connector_id/schemas/").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)
			if updateIteration == 0 {
				// Check the request
				assertEqual(t, len(body), 1)
				assertEqual(t, body["schema_change_handling"], "BLOCK_ALL")

				// create schema structure
				schemaLockedData = createMapFromJsonString(t, schemaWithLockedTableAndColumn_blocked0)
			}
			if updateIteration == 1 {
				// Check the request
				assertEqual(t, len(body), 1)
				assertNotEmpty(t, body["schemas"])

				schemasMap := body["schemas"].(map[string]interface{})

				schema1 := schemasMap["schema_1"].(map[string]interface{})

				assertEqual(t, len(schema1), 2)

				assertEqual(t, schema1["enabled"], false)
				assertNotEmpty(t, schema1["tables"])

				tablesMap := schema1["tables"].(map[string]interface{})

				assertEqual(t, len(tablesMap), 2)

				table1 := tablesMap["table_1"].(map[string]interface{})

				assertEqual(t, len(table1), 2)

				assertEqual(t, table1["enabled"], false)
				assertNotEmpty(t, table1["columns"])

				columnsMap := table1["columns"].(map[string]interface{})

				column2 := columnsMap["column_2"].(map[string]interface{})

				assertEqual(t, column2["enabled"], false)

				table2 := tablesMap["table_2"].(map[string]interface{})

				assertEqual(t, len(table2), 1)

				assertNotEmpty(t, table2["columns"])

				columnsMap2 := table2["columns"].(map[string]interface{})

				column4 := columnsMap2["column_4"].(map[string]interface{})

				assertEqual(t, column4["enabled"], false)

				// create schema structure
				schemaLockedData = createMapFromJsonString(t, schemaEmptyDefaultJsonBlockedSchema2)
			}
			updateIteration++
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaLockedData), nil
		},
	)
}

func setupMockClientHashedAlignmentSchemaResource(t *testing.T) {
	mockClient.Reset()
	schemaHashedAlignmentData = nil
	updateIteration := 0

	schemaHashedAlignmentGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if nil == schemaHashedAlignmentData {
				schemaHashedAlignmentData = createMapFromJsonString(t, schemaHashedColumnAlignmentJsonSchema)
			}
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaHashedAlignmentData), nil
		},
	)

	schemaHashedAlignmentPatchHandler = mockClient.When(http.MethodPatch, "/v1/connectors/connector_id/schemas/").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)
			if updateIteration == 0 {
				assertNotEmpty(t, body["schemas"])

				schemasMap := body["schemas"].(map[string]interface{})
				schema1 := schemasMap["schema_1"].(map[string]interface{})

				assertEqual(t, len(schema1), 2)

				assertNotEmpty(t, schema1["tables"])

				tablesMap := schema1["tables"].(map[string]interface{})
				table1 := tablesMap["table_1"].(map[string]interface{})

				assertEqual(t, len(table1), 2)

				assertNotEmpty(t, table1["columns"])

				columnsMap := table1["columns"].(map[string]interface{})
				column1 := columnsMap["column_1"].(map[string]interface{})

				assertEqual(t, column1["hashed"], false)

				schemaHashedAlignmentData = createMapFromJsonString(t, schemaEmptyDefaultJsonSchema)
			}

			updateIteration++
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaHashedAlignmentData), nil
		},
	)
}

func setupMockClientConsistentWithUpstreamResource(t *testing.T) {
	mockClient.Reset()
	schemaConsistentWithUpstreamData = nil

	schemaConsistentWithUpstreamGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if nil == schemaHashedAlignmentData {
				schemaConsistentWithUpstreamData = createMapFromJsonString(t, `
				{
					"enable_new_by_default": true,
					"schema_change_handling": "BLOCK_ALL",
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
									}
									
								}
							}
						}
					}
				}
				`)
			}
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaConsistentWithUpstreamData), nil
		},
	)

	schemaConsistentWithUpstreamPathchHandler = mockClient.When(http.MethodPatch, "/v1/connectors/connector_id/schemas/").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaConsistentWithUpstreamData), nil
		},
	)
}

func TestConsistentWithUpstreamSchemaMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
				schema {
					name = "schema_1"
					enabled = true
					table {
						name = "table_1"
						enabled = true
						sync_mode = "SOFT_DELETE"
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaConsistentWithUpstreamGetHandler.Interactions, 2) // 1 read attempt before reload, 1 read after create
				assertNotEmpty(t, schemaConsistentWithUpstreamData)                    // schema initialised
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.0.sync_mode", "SOFT_DELETE"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConsistentWithUpstreamResource(t)
			},
			Providers: testProviders,
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

func TestResourceHashedAlignmentSchemaMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "ALLOW_ALL"
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaHashedAlignmentGetHandler.Interactions, 2)   // 1 read attempt before reload, 1 read after create
				assertEqual(t, schemaHashedAlignmentPatchHandler.Interactions, 1) // Update hashed for column
				assertNotEmpty(t, schemaHashedAlignmentData)                      // schema initialised
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_ALL"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientHashedAlignmentSchemaResource(t)
			},
			Providers: testProviders,
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

func TestResourceLockedSchemaMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaLockedGetHandler.Interactions, 2)   // 1 read attempt before reload, 1 read after create
				assertEqual(t, schemaLockedPatchHandler.Interactions, 2) // Update SCM and align schema
				assertNotEmpty(t, schemaLockedData)                      // schema initialised
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientLockedPartsSchemaResource(t)
			},
			Providers: testProviders,
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
