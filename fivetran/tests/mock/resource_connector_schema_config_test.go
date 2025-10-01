package mock

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

	listConnectionsHandler *mock.Handler
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
	schemaEmptyDefaultJsonBlockedSchema3 = `
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

	schemaEmptyDefaultJsonBlockedSchema4 = `
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

	schemaEmptyDefaultReloadHandler = mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/reload").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)

			// Check the request
			assertEqual(t, len(body), 1)
			//assertEqual(t, body["exclude_mode"], "PRESERVE") // reload schema in PRESERVE mode
			assertKeyExists(t, body, "exclude_mode")

			// create schema structure
			if updateIteration == 0 {
				schemaEmptyDefaultData = createMapFromJsonString(t, schemaEmptyDefaultJsonSchema)
			}
			if updateIteration == 1 {
				schemaEmptyDefaultData = createMapFromJsonString(t, schemaEmptyDefaultJsonBlockedSchema2)
			}
			if updateIteration == 2 {
				schemaEmptyDefaultData = createMapFromJsonString(t, schemaEmptyDefaultJsonBlockedSchema3)
			}
			if updateIteration > 2 {
				schemaEmptyDefaultData = createMapFromJsonString(t, schemaEmptyDefaultJsonBlockedSchema4)
			}

			response := fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaEmptyDefaultData)

			return response, nil
		},
	)

	schemaEmptyDefaultGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if nil == schemaEmptyDefaultData {
				return fivetranResponse(t, req,
					"NotFound_SchemaConfig", http.StatusNotFound,
					"Connector with id 'connector_id' doesn't have schema config", nil), nil
			}

			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaEmptyDefaultData), nil
		},
	)

	schemaEmptyDefaultPatchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)
			if updateIteration == 0 {
				// Check the request
				assertEqual(t, len(body), 2)

				assertEqual(t, body["schema_change_handling"], "BLOCK_ALL")
				assertNotEmpty(t, body["schemas"])

				schemasMap := body["schemas"].(map[string]interface{})

				schema1 := schemasMap["schema_1"].(map[string]interface{})

				assertEqual(t, len(schema1), 1)

				assertEqual(t, schema1["enabled"], false)

				// create schema structure
				schemaEmptyDefaultData = createMapFromJsonString(t, schemaEmptyDefaultJsonBlockedSchema2)
			}
			if updateIteration == 1 {
				// Check the request
				assertEqual(t, len(body), 1)

				schemasMap := AssertKeyExists(t, body, "schemas").(map[string]interface{})

				schema1 := AssertKeyExists(t, schemasMap, "schema_1").(map[string]interface{})

				tablesMap := AssertKeyExists(t, schema1, "tables").(map[string]interface{})

				table1 := AssertKeyExists(t, tablesMap, "table_1").(map[string]interface{})

				assertKeyExistsAndHasValue(t, table1, "enabled", false)

				columnsMap := AssertKeyExists(t, table1, "columns").(map[string]interface{})

				column1 := AssertKeyExists(t, columnsMap, "column_1").(map[string]interface{})

				assertKeyExistsAndHasValue(t, column1, "enabled", false)

				// create schema structure
				schemaEmptyDefaultData = createMapFromJsonString(t, schemaEmptyDefaultJsonBlockedSchema3)
			}

			if updateIteration == 2 {
				// Check the request
				assertEqual(t, len(body), 1)

				schemasMap := AssertKeyExists(t, body, "schemas").(map[string]interface{})

				schema1 := AssertKeyExists(t, schemasMap, "schema_1").(map[string]interface{})

				assertKeyExistsAndHasValue(t, schema1, "enabled", true)

				tablesMap := AssertKeyExists(t, schema1, "tables").(map[string]interface{})

				table1 := AssertKeyExists(t, tablesMap, "table_1").(map[string]interface{})

				assertKeyExistsAndHasValue(t, table1, "enabled", true)

				columnsMap := AssertKeyExists(t, table1, "columns").(map[string]interface{})

				column1 := AssertKeyExists(t, columnsMap, "column_1").(map[string]interface{})

				assertKeyExistsAndHasValue(t, column1, "enabled", true)

				// create schema structure
				schemaEmptyDefaultData = createMapFromJsonString(t, schemaEmptyDefaultJsonBlockedSchema4)
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
				assertEqual(t, schemaEmptyDefaultReloadHandler.Interactions, 1)
				assertEqual(t, schemaEmptyDefaultGetHandler.Interactions, 3)
				assertEqual(t, schemaEmptyDefaultPatchHandler.Interactions, 0)
				assertNotEmpty(t, schemaEmptyDefaultData) // schema initialised
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
				assertEqual(t, schemaEmptyDefaultPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
			resource.TestCheckNoResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0"),
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
				//assertEqual(t, schemaEmptyDefaultPatchHandler.Interactions, 1)
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
				//assertEqual(t, schemaEmptyDefaultPatchHandler.Interactions, 2)
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
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
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

	schemaLockedGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if nil == schemaLockedData {
				schemaLockedData = createMapFromJsonString(t, schemaWithLockedTableAndColumn)
			}
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaLockedData), nil
		},
	)

	schemaLockedPatchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)

			if updateIteration == 0 {
				// Check the request
				assertEqual(t, len(body), 2)
				assertKeyExistsAndHasValue(t, body, "schema_change_handling", "BLOCK_ALL")

				schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})

				schema_1 := assertKeyExists(t, schemas, "schema_1").(map[string]interface{})

				assertKeyExistsAndHasValue(t, schema_1, "enabled", false)

				// create schema structure
				schemaLockedData = createMapFromJsonString(t, schemaWithLockedTableAndColumn_blocked1)
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

	schemaHashedAlignmentGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if nil == schemaHashedAlignmentData {
				schemaHashedAlignmentData = createMapFromJsonString(t, schemaHashedColumnAlignmentJsonSchema)
			}
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaHashedAlignmentData), nil
		},
	)

	schemaHashedAlignmentPatchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)
			if updateIteration == 0 {
				assertNotEmpty(t, body["schemas"])

				schemasMap := body["schemas"].(map[string]interface{})
				schema1 := schemasMap["schema_1"].(map[string]interface{})

				assertEqual(t, len(schema1), 1)

				assertNotEmpty(t, schema1["tables"])

				tablesMap := schema1["tables"].(map[string]interface{})
				table1 := tablesMap["table_1"].(map[string]interface{})

				assertEqual(t, len(table1), 1)

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

func TestSyncModeMock(t *testing.T) {
	setupMockClientConsistentWithUpstreamResource := func(t *testing.T) {
		mockClient.Reset()
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
										"enabled_patch_settings": {
											"allowed": true
										}
									},
									"table_2": {
										"name_in_destination": "table_2",
										"enabled": true,
										"sync_mode": "LIVE",
										"enabled_patch_settings": {
											"allowed": true
										}
									},
									"table_3": {
										"name_in_destination": "table_3",
										"enabled": false,
										"sync_mode": "LIVE",
										"enabled_patch_settings": {
											"allowed": true
										}
									}
								}
							}
						}
					}
					`)

		schemaConsistentWithUpstreamGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaConsistentWithUpstreamData), nil
			},
		)

		schemaConsistentWithUpstreamPathchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				fmt.Print(body)
				syncMode := body["schemas"].(map[string]interface{})["schema_1"].(map[string]interface{})["tables"].(map[string]interface{})["table_1"].(map[string]interface{})["sync_mode"].(string)
				assertEqual(t, syncMode, "HISTORY")

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
										"sync_mode": "HISTORY",
										"enabled_patch_settings": {
											"allowed": true
										}
									},
									"table_2": {
										"name_in_destination": "table_2",
										"enabled": true,
										"sync_mode": "LIVE",
										"enabled_patch_settings": {
											"allowed": true
										}
									},
									"table_3": {
										"name_in_destination": "table_3",
										"enabled": false,
										"sync_mode": "LIVE",
										"enabled_patch_settings": {
											"allowed": true
										}
									}
								}
							}
						}
					}
					`)

				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaConsistentWithUpstreamData), nil
			},
		)
	}

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
						sync_mode = "HISTORY"
					}
					table {
						name = "table_2"
						enabled = true
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaConsistentWithUpstreamGetHandler.Interactions, 3) // 1 read attempt before reload, 1 read after create
				assertNotEmpty(t, schemaConsistentWithUpstreamData)                    // schema initialised
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.0.sync_mode", "HISTORY"),
			resource.TestCheckNoResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.1.sync_mode"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConsistentWithUpstreamResource(t)
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

func TestConsistentWithUpstreamSchemaMock(t *testing.T) {
	setupMockClientConsistentWithUpstreamResource := func(t *testing.T) {
		mockClient.Reset()
		schemaConsistentWithUpstreamData = nil

		schemaConsistentWithUpstreamGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
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
									},
									"table_2": {
										"name_in_destination": "table_2",
										"enabled": true,
										"sync_mode": "LIVE",
										"enabled_patch_settings": {
											"allowed": true
										}
									},
									"table_3": {
										"name_in_destination": "table_3",
										"enabled": false,
										"sync_mode": "LIVE",
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

		schemaConsistentWithUpstreamPathchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				fmt.Print(body)
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaConsistentWithUpstreamData), nil
			},
		)
	}

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
					table {
						name = "table_2"
						enabled = true
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaConsistentWithUpstreamGetHandler.Interactions, 3)
				assertNotEmpty(t, schemaConsistentWithUpstreamData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.0.sync_mode", "SOFT_DELETE"),
			resource.TestCheckNoResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.1.sync_mode"),
		),
	}

	step2 := resource.TestStep{
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
					}
					table {
						name = "table_2"
						enabled = true
						sync_mode = "LIVE"
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
			resource.TestCheckNoResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.0.sync_mode"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.1.sync_mode", "LIVE"),
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
					enabled = true
					table {
						name = "table_1"
						enabled = true
					}
					table {
						name = "table_2"
						enabled = true
						sync_mode = "LIVE"
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
			resource.TestCheckNoResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.0.sync_mode"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.1.sync_mode", "LIVE"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConsistentWithUpstreamResource(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				// there is no possibility to destroy schema config - it alsways exists within the connector
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
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
				assertEqual(t, schemaHashedAlignmentGetHandler.Interactions, 3)   // 1 read attempt before reload, 1 read after create
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
				assertEqual(t, schemaLockedGetHandler.Interactions, 3)   // 1 read attempt before reload, 1 read after create
				assertEqual(t, schemaLockedPatchHandler.Interactions, 1) // Update SCM and align schema
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

func TestConsistentWithUpstreamSchemaMappedMock(t *testing.T) {
	setupMockClientConsistentWithUpstreamResource := func(t *testing.T) {
		mockClient.Reset()
		schemaConsistentWithUpstreamData = nil

		schemaConsistentWithUpstreamGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if nil == schemaConsistentWithUpstreamData {
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
									},
									"table_2": {
										"name_in_destination": "table_2",
										"enabled": true,
										"sync_mode": "LIVE",
										"enabled_patch_settings": {
											"allowed": true
										}
									},
									"table_3": {
										"name_in_destination": "table_3",
										"enabled": false,
										"sync_mode": "LIVE",
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

		schemaConsistentWithUpstreamPathchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				fmt.Print(body)
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaConsistentWithUpstreamData), nil
			},
		)
	}

	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
				schemas = {
					"schema_1" = {
						enabled = true
						tables = {
							"table_1" = {
								enabled = true
								sync_mode = "SOFT_DELETE"
							}
							"table_2" = {
								enabled = true
							}
						}
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaConsistentWithUpstreamGetHandler.Interactions, 3)
				assertNotEmpty(t, schemaConsistentWithUpstreamData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_1.sync_mode", "SOFT_DELETE"),
			resource.TestCheckNoResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_2.sync_mode"),
		),
	}

	step2 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
				schemas = {
					"schema_1" = {
						enabled = true
						tables = {
							"table_1" = {
								enabled = true
							}
							"table_2" = {
								enabled = true
								sync_mode = "LIVE"
							}
						}
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
			resource.TestCheckNoResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_1.sync_mode"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_2.sync_mode", "LIVE"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConsistentWithUpstreamResource(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				// there is no possibility to destroy schema config - it alsways exists within the connector
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}

func setupMockClientListConnections(t *testing.T) {

	const connectionsMappingResponse = `
	{
    "items": [
      {
        "id": "connector_id",
        "service": "string",
        "schema": "schema_name",
        "paused": false,
        "daily_sync_time": "14:00",
        "succeeded_at": "2024-12-01T15:43:29.013729Z",
        "sync_frequency": 360,
        "group_id": "group_id",
        "connected_by": "user_id",
        "service_version": 0,
        "created_at": "2024-12-01T15:43:29.013729Z",
        "failed_at": "2024-12-01T15:43:29.013729Z",
        "private_link_id": "string",
        "proxy_agent_id": "string",
        "networking_method": "Directly",
        "pause_after_trial": false,
        "data_delay_threshold": 0,
        "data_delay_sensitivity": "LOW",
        "schedule_type": "auto",
        "hybrid_deployment_agent_id": "string"
      }
    ],
    "next_cursor": null
  }
`
	listConnectionsHandler = mockClient.When(http.MethodGet, "/v1/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			var connectionsListData map[string]interface{}
			if req.URL.Query().Get("group_id") == "group_id" && req.URL.Query().Get("schema") == "schema_name" {
				connectionsListData = createMapFromJsonString(t, connectionsMappingResponse)
			} else {
				connectionsListData = nil
			}

			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectionsListData), nil
		},
	)
}

// This test checks that schema config can be created by group id and connector name
func TestResourceByGroupIdAndConnectorNameMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider

				group_id = "group_id"
				connector_name = "schema_name"

				schema_change_handling = "ALLOW_ALL"
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, listConnectionsHandler.Interactions, 1)
				assertEqual(t, schemaEmptyDefaultReloadHandler.Interactions, 1)
				assertEqual(t, schemaEmptyDefaultGetHandler.Interactions, 3)
				assertEqual(t, schemaEmptyDefaultPatchHandler.Interactions, 0)
				assertNotEmpty(t, schemaEmptyDefaultData) // schema initialised
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "group_id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_name", "schema_name"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_ALL"),
		),
	}

	step2 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider

				group_id = "group_id"
				connector_name = "schema_name"

				schema_change_handling = "BLOCK_ALL"
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaEmptyDefaultPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "group_id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_name", "schema_name"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
			resource.TestCheckNoResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientEmptyDefaultSchemaResource(t)
				setupMockClientListConnections(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				// there is no possibility to destroy schema config - it alsways exists within the connector
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}

// Test to verify that is_primary_key field is computed-only and doesn't cause noisy plan diffs
func TestResourceSchemaPrimaryKeyComputedMock(t *testing.T) {
	var (
		schemaPrimaryKeyGetHandler   *mock.Handler
		schemaPrimaryKeyPatchHandler *mock.Handler
		schemaPrimaryKeyData         map[string]interface{}
	)

	setupMockClientPrimaryKeyResource := func(t *testing.T) {
		mockClient.Reset()
		schemaPrimaryKeyData = nil

		// Mock GET handler
		schemaPrimaryKeyGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if schemaPrimaryKeyData == nil {
					schemaPrimaryKeyData = createMapFromJsonString(t, `
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
										"enabled_patch_settings": {
											"allowed": true
										},
										"columns": {
											"id": {
												"name_in_destination": "id",
												"enabled": true,
												"hashed": false,
												"is_primary_key": true,
												"enabled_patch_settings": {
													"allowed": true
												}
											},
											"name": {
												"name_in_destination": "name",
												"enabled": true,
												"hashed": false,
												"is_primary_key": false,
												"enabled_patch_settings": {
													"allowed": true
												}
											}
										}
									},
									"table_2": {
										"name_in_destination": "table_2",
										"enabled": true,
										"enabled_patch_settings": {
											"allowed": true
										},
										"columns": {
											"id": {
												"name_in_destination": "id",
												"enabled": true,
												"hashed": false,
												"is_primary_key": true,
												"enabled_patch_settings": {
													"allowed": true
												}
											},
											"value": {
												"name_in_destination": "value",
												"enabled": true,
												"hashed": false,
												"is_primary_key": false,
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
					`)
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaPrimaryKeyData), nil
			},
		)

		// Mock PATCH handler - verifies no is_primary_key field is sent in requests
		schemaPrimaryKeyPatchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)

				// Helper function to validate columns don't contain is_primary_key
				validateColumns := func(columns map[string]interface{}) {
					for columnName, columnVal := range columns {
						columnMap, ok := columnVal.(map[string]interface{})
						if !ok {
							continue
						}

						// Assert that is_primary_key is NOT in any column
						if _, exists := columnMap["is_primary_key"]; exists {
							t.Errorf("is_primary_key should not be sent in PATCH requests, but was present in column %s", columnName)
						}

						// Verify hashed field is properly sent for the name column
						if columnName == "name" {
							if hashed, ok := columnMap["hashed"].(bool); !ok || !hashed {
								t.Errorf("Expected hashed=true for name column, but got: %v", columnMap["hashed"])
							}
						}
					}
				}

				// Walk through the schema structure and validate columns
				schemas, ok := body["schemas"].(map[string]interface{})
				if !ok {
					return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaPrimaryKeyData), nil
				}

				for _, schemaVal := range schemas {
					schemaMap, ok := schemaVal.(map[string]interface{})
					if !ok {
						continue
					}

					tables, ok := schemaMap["tables"].(map[string]interface{})
					if !ok {
						continue
					}

					for _, tableVal := range tables {
						tableMap, ok := tableVal.(map[string]interface{})
						if !ok {
							continue
						}

						columns, ok := tableMap["columns"].(map[string]interface{})
						if ok {
							validateColumns(columns)
						}
					}
				}

				// Update the mock data with the change
				schemaPrimaryKeyData = createMapFromJsonString(t, `
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
									"enabled_patch_settings": {
										"allowed": true
									},
									"columns": {
										"id": {
											"name_in_destination": "id",
											"enabled": true,
											"hashed": false,
											"is_primary_key": true,
											"enabled_patch_settings": {
												"allowed": true
											}
										},
										"name": {
											"name_in_destination": "name",
											"enabled": true,
											"hashed": true,
											"is_primary_key": false,
											"enabled_patch_settings": {
												"allowed": true
											}
										}
									}
								},
								"table_2": {
									"name_in_destination": "table_2",
									"enabled": true,
									"enabled_patch_settings": {
										"allowed": true
									},
									"columns": {
										"id": {
											"name_in_destination": "id",
											"enabled": true,
											"hashed": false,
											"is_primary_key": true,
											"enabled_patch_settings": {
												"allowed": true
											}
										},
										"value": {
											"name_in_destination": "value",
											"enabled": true,
											"hashed": false,
											"is_primary_key": false,
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
				`)

				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaPrimaryKeyData), nil
			},
		)
	}

	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "ALLOW_ALL"
				schemas = {
					"schema_1" = {
						tables = {
							"table_1" = {
								columns = {
									"id" = {
										enabled = true
									}
									"name" = {
										enabled = true
									}
								}
							}
							"table_2" = {
								columns = {
									"id" = {
										enabled = true
									}
									"value" = {
										enabled = true
									}
								}
							}
						}
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaPrimaryKeyGetHandler.Interactions, 3)
				assertNotEmpty(t, schemaPrimaryKeyData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_ALL"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_1.columns.id.enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_1.columns.name.enabled", "true"),
		),
	}

	// Step 2: Make a change to a column (hash it) without touching is_primary_key
	// This verifies that is_primary_key values don't create diffs for unchanged columns
	step2 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "ALLOW_ALL"
				schemas = {
					"schema_1" = {
						tables = {
							"table_1" = {
								columns = {
									"id" = {
										enabled = true
									}
									"name" = {
										enabled = true
										hashed = true
									}
								}
							}
							"table_2" = {
								columns = {
									"id" = {
										enabled = true
									}
									"value" = {
										enabled = true
									}
								}
							}
						}
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaPrimaryKeyPatchHandler.Interactions, 2)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_1.columns.name.hashed", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.table_1.columns.name.enabled", "true"),
		),
	}

	// Step 3: Plan-only test to verify no changes are detected when config is unchanged
	// This explicitly tests that is_primary_key values from API don't show up as plan diffs
	step3 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "ALLOW_ALL"
				schemas = {
					"schema_1" = {
						tables = {
							"table_1" = {
								columns = {
									"id" = {
										enabled = true
									}
									"name" = {
										enabled = true
										hashed = true
									}
								}
							}
							"table_2" = {
								columns = {
									"id" = {
										enabled = true
									}
									"value" = {
										enabled = true
									}
								}
							}
						}
					}
				}
			}`,
		PlanOnly:           true,
		ExpectNonEmptyPlan: false, // Expect no changes - is_primary_key should not cause diffs
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientPrimaryKeyResource(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
				step2,
				step3, // Plan-only step to verify no drift from is_primary_key
			},
		},
	)
}
