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

// Test to verify that is_primary_key defaults to false when API doesn't provide it
// This addresses the specific case where API responses don't include is_primary_key
func TestResourceSchemaPrimaryKeyNullDefaultMock(t *testing.T) {
	var (
		schemaPrimaryKeyNullGetHandler   *mock.Handler
		schemaPrimaryKeyNullPatchHandler *mock.Handler
		schemaPrimaryKeyNullData         map[string]interface{}
	)

	setupMockClientPrimaryKeyNullResource := func(t *testing.T) {
		mockClient.Reset()
		schemaPrimaryKeyNullData = nil

		// Mock GET handler - API response WITHOUT is_primary_key (like customer's case)
		schemaPrimaryKeyNullGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if schemaPrimaryKeyNullData == nil {
					schemaPrimaryKeyNullData = createMapFromJsonString(t, `
					{
						"enable_new_by_default": true,
						"schema_change_handling": "ALLOW_ALL",
						"schemas": {
							"lendable": {
								"name_in_destination": "lendable",
								"enabled": true,
								"tables": {
									"account": {
										"name_in_destination": "account",
										"enabled": true,
										"enabled_patch_settings": {
											"allowed": true
										},
										"columns": {
											"balance": {
												"name_in_destination": "balance",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": {
													"allowed": true
												}
											},
											"id": {
												"name_in_destination": "id",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": {
													"allowed": true
												}
											},
											"label": {
												"name_in_destination": "label",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": {
													"allowed": true
												}
											},
											"type": {
												"name_in_destination": "type",
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
					`)
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaPrimaryKeyNullData), nil
			},
		)

		// Mock PATCH handler - verifies no is_primary_key is sent
		schemaPrimaryKeyNullPatchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
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
					}
				}

				// Walk through the schema structure and validate columns
				schemas, ok := body["schemas"].(map[string]interface{})
				if !ok {
					return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaPrimaryKeyNullData), nil
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

				// Update mock data - preserve the hashed field that was set to true
				schemaPrimaryKeyNullData = createMapFromJsonString(t, `
				{
					"enable_new_by_default": true,
					"schema_change_handling": "ALLOW_ALL",
					"schemas": {
						"lendable": {
							"name_in_destination": "lendable",
							"enabled": true,
							"tables": {
								"account": {
									"name_in_destination": "account",
									"enabled": true,
									"enabled_patch_settings": {
										"allowed": true
									},
									"columns": {
										"balance": {
											"name_in_destination": "balance",
											"enabled": true,
											"hashed": false,
											"enabled_patch_settings": {
												"allowed": true
											}
										},
										"id": {
											"name_in_destination": "id",
											"enabled": true,
											"hashed": false,
											"enabled_patch_settings": {
												"allowed": true
											}
										},
										"label": {
											"name_in_destination": "label",
											"enabled": true,
											"hashed": false,
											"enabled_patch_settings": {
												"allowed": true
											}
										},
										"type": {
											"name_in_destination": "type",
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
				`)

				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaPrimaryKeyNullData), nil
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
					"lendable" = {
						tables = {
							"account" = {
								columns = {
									"balance" = {
										enabled = true
									}
									"id" = {
										enabled = true
									}
									"label" = {
										enabled = true
									}
									"type" = {
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
				assertEqual(t, schemaPrimaryKeyNullGetHandler.Interactions, 3)
				assertNotEmpty(t, schemaPrimaryKeyNullData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_ALL"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.lendable.tables.account.columns.balance.enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.lendable.tables.account.columns.id.enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.lendable.tables.account.columns.label.enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.lendable.tables.account.columns.type.enabled", "true"),
		),
	}

	// Step 2: Make a change to one column
	step2 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "ALLOW_ALL"
				schemas = {
					"lendable" = {
						tables = {
							"account" = {
								columns = {
									"balance" = {
										enabled = true
									}
									"id" = {
										enabled = true
									}
									"label" = {
										enabled = true
									}
									"type" = {
										enabled = true
										hashed = true
									}
								}
							}
						}
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaPrimaryKeyNullPatchHandler.Interactions, 2)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.lendable.tables.account.columns.type.hashed", "true"),
		),
	}

	// Step 3: Plan-only test - this is the critical test that verifies no is_primary_key diffs
	// This simulates the customer's exact scenario from the screenshot
	step3 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "ALLOW_ALL"
				schemas = {
					"lendable" = {
						tables = {
							"account" = {
								columns = {
									"balance" = {
										enabled = true
									}
									"id" = {
										enabled = true
									}
									"label" = {
										enabled = true
									}
									"type" = {
										enabled = true
										hashed = true
									}
								}
							}
						}
					}
				}
			}`,
		PlanOnly:           true,
		ExpectNonEmptyPlan: false,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				resource := s.RootModule().Resources["fivetran_connector_schema_config.test_schema"]
				if resource == nil {
					return fmt.Errorf("resource not found in state")
				}

				columns := []string{"balance", "id", "label", "type"}
				for _, col := range columns {
					key := fmt.Sprintf("schemas.lendable.tables.account.columns.%s.is_primary_key", col)
					if val, ok := resource.Primary.Attributes[key]; !ok {
						return fmt.Errorf("is_primary_key not found for column %s", col)
					} else if val != "" {
						return fmt.Errorf("is_primary_key for column %s should be null (empty string), got '%s'", col, val)
					}
				}

				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientPrimaryKeyNullResource(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
				step2,
				step3, // This step verifies no is_primary_key noise in plans
			},
		},
	)
}

func TestResourceSchemaPrimaryKeyLargeSchemaNoNoiseMock(t *testing.T) {
	var (
		schemaLargeGetHandler   *mock.Handler
		schemaLargePatchHandler *mock.Handler
		schemaLargeData         map[string]interface{}
	)

	setupMockClientLargeSchema := func(t *testing.T) {
		mockClient.Reset()
		schemaLargeData = nil

		// Mock GET handler - Simulates Aurora MySQL with multiple tables/columns WITHOUT is_primary_key
		schemaLargeGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if schemaLargeData == nil {
					// Simulating a schema with multiple tables and columns (like customer's Aurora MySQL)
					schemaLargeData = createMapFromJsonString(t, `
					{
						"enable_new_by_default": true,
						"schema_change_handling": "ALLOW_ALL",
						"schemas": {
							"public": {
								"name_in_destination": "public",
								"enabled": true,
								"tables": {
									"users": {
										"name_in_destination": "users",
										"enabled": true,
										"enabled_patch_settings": {"allowed": true},
										"columns": {
											"id": {"name_in_destination": "id", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}},
											"email": {"name_in_destination": "email", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}},
											"name": {"name_in_destination": "name", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}},
											"created_at": {"name_in_destination": "created_at", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}}
										}
									},
									"orders": {
										"name_in_destination": "orders",
										"enabled": true,
										"enabled_patch_settings": {"allowed": true},
										"columns": {
											"id": {"name_in_destination": "id", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}},
											"user_id": {"name_in_destination": "user_id", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}},
											"total": {"name_in_destination": "total", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}},
											"status": {"name_in_destination": "status", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}}
										}
									},
									"products": {
										"name_in_destination": "products",
										"enabled": true,
										"enabled_patch_settings": {"allowed": true},
										"columns": {
											"id": {"name_in_destination": "id", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}},
											"name": {"name_in_destination": "name", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}},
											"price": {"name_in_destination": "price", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}},
											"description": {"name_in_destination": "description", "enabled": true, "hashed": false, "enabled_patch_settings": {"allowed": true}}
										}
									}
								}
							}
						}
					}
					`)
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaLargeData), nil
			},
		)

		// Mock PATCH handler
		schemaLargePatchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)

				// Validate no is_primary_key in request
				validateColumns := func(columns map[string]interface{}) {
					for columnName, columnVal := range columns {
						columnMap, ok := columnVal.(map[string]interface{})
						if !ok {
							continue
						}
						if _, exists := columnMap["is_primary_key"]; exists {
							t.Errorf("is_primary_key should not be sent in PATCH requests, but was present in column %s", columnName)
						}
					}
				}

				schemas, ok := body["schemas"].(map[string]interface{})
				if ok {
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
				}

				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaLargeData), nil
			},
		)
	}

	// Step 1: Create resource with all columns
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "ALLOW_ALL"
				schemas = {
					"public" = {
						tables = {
							"users" = {
								columns = {
									"id" = { enabled = true }
									"email" = { enabled = true }
									"name" = { enabled = true }
									"created_at" = { enabled = true }
								}
							}
							"orders" = {
								columns = {
									"id" = { enabled = true }
									"user_id" = { enabled = true }
									"total" = { enabled = true }
									"status" = { enabled = true }
								}
							}
							"products" = {
								columns = {
									"id" = { enabled = true }
									"name" = { enabled = true }
									"price" = { enabled = true }
									"description" = { enabled = true }
								}
							}
						}
					}
				}
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaLargeGetHandler.Interactions, 3)
				// PATCH handler called during create (at least once)
				if schemaLargePatchHandler.Interactions < 1 {
					return fmt.Errorf("expected at least 1 PATCH interaction, got %d", schemaLargePatchHandler.Interactions)
				}
				return nil
			},
		),
	}

	// Step 2: Plan-only test - NO changes should be shown
	step2 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "ALLOW_ALL"
				schemas = {
					"public" = {
						tables = {
							"users" = {
								columns = {
									"id" = { enabled = true }
									"email" = { enabled = true }
									"name" = { enabled = true }
									"created_at" = { enabled = true }
								}
							}
							"orders" = {
								columns = {
									"id" = { enabled = true }
									"user_id" = { enabled = true }
									"total" = { enabled = true }
									"status" = { enabled = true }
								}
							}
							"products" = {
								columns = {
									"id" = { enabled = true }
									"name" = { enabled = true }
									"price" = { enabled = true }
									"description" = { enabled = true }
								}
							}
						}
					}
				}
			}`,
		PlanOnly:           true,
		ExpectNonEmptyPlan: false,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				// Verify all 12 columns have is_primary_key = false in state
				tables := map[string][]string{
					"users":    {"id", "email", "name", "created_at"},
					"orders":   {"id", "user_id", "total", "status"},
					"products": {"id", "name", "price", "description"},
				}

				resource := s.RootModule().Resources["fivetran_connector_schema_config.test_schema"]
				if resource == nil {
					return fmt.Errorf("resource not found in state")
				}

				for table, columns := range tables {
					for _, col := range columns {
						key := fmt.Sprintf("schemas.public.tables.%s.columns.%s.is_primary_key", table, col)
						if val, ok := resource.Primary.Attributes[key]; !ok {
							return fmt.Errorf("is_primary_key not found for %s.%s", table, col)
						} else if val != "" {
							return fmt.Errorf("is_primary_key for %s.%s should be null (empty string), got '%s'", table, col, val)
						}
					}
				}

				t.Logf("✓ SUCCESS: All 12 columns have is_primary_key=null in state")
				t.Logf("✓ SUCCESS: Plan shows NO changes (no is_primary_key noise)")
				t.Logf("✓ This fix prevents the customer's issue of hundreds of noisy plan lines!")

				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientLargeSchema(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}

func TestResourceSchemaConsequentGetReturnsLessColumnsMock(t *testing.T) {
	var (
		schemaGetHandler   *mock.Handler
		schemaPatchHandler *mock.Handler
		schemaReloadPostHandler   *mock.Handler
		schemaResponseData map[string]interface{}
	)
	
	schemaGetResponse1 :=  `
					{
						"enable_new_by_default": false,
						"schema_change_handling": "BLOCK_ALL",
						"schemas": {
							"public": {
								"name_in_destination": "public",
								"enabled": true,
								"tables": {
									"table_1": {
										"name_in_destination": "table_1",
										"enabled": false,
										"supports_columns_config": true,
										"sync_mode": "SOFT_DELETE",
										"enabled_patch_settings": { "allowed": true },
										"columns": {
											"table_1_col_1": {
												"name_in_destination": "table_1_col_1",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": {
													"allowed": false,
													"reason_code": "SYSTEM_COLUMN",
													"reason": "Column does not support exclusion as it is a Primary Key"
												}
											}
										}
									},
									"table_2": {
										"name_in_destination": "table_2",
										"enabled": true,
										"supports_columns_config": true,
										"sync_mode": "SOFT_DELETE",
										"enabled_patch_settings": { "allowed": true },
										"columns": {
											"table_2_col_1": {
												"name_in_destination": "table_2_col_1",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": { "allowed": true }
											},
											"table_2_col_2": {
												"name_in_destination": "table_2_col_2",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": { "allowed": true }
											}
										}
									},
									"table_3": {
										"name_in_destination": "table_3",
										"enabled": true,
										"supports_columns_config": true,
										"sync_mode": "SOFT_DELETE",
										"enabled_patch_settings": { "allowed": true },
										"columns": {
											"table_3_col_1": {
												"name_in_destination": "table_3_col_1",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": { "allowed": true }
											}
										}
									}
								}
							},
							"private": {
								"name_in_destination": "private",
								"enabled": true,
								"tables": {
									"table_1": {
										"name_in_destination": "table_1",
										"enabled": true,
										"supports_columns_config": true,
										"sync_mode": "SOFT_DELETE",
										"enabled_patch_settings": { "allowed": true },
										"columns": {
											"table_1_col_1": {
												"name_in_destination": "table_1_col_1",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": {
													"allowed": false,
													"reason_code": "SYSTEM_COLUMN",
													"reason": "Column does not support exclusion as it is a Primary Key"
												}
											}
										}
									}
								}
							}
						}
					}`

	schemaGetResponseAbsentSchema :=  `
					{
						"enable_new_by_default": false,
						"schema_change_handling": "BLOCK_ALL",
						"schemas": {
							"public": {
								"name_in_destination": "public",
								"enabled": true,
								"tables": {
									"table_1": {
										"name_in_destination": "table_1",
										"enabled": false,
										"supports_columns_config": true,
										"sync_mode": "SOFT_DELETE",
										"enabled_patch_settings": { "allowed": true },
										"columns": {
											"table_1_col_1": {
												"name_in_destination": "table_1_col_1",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": {
													"allowed": false,
													"reason_code": "SYSTEM_COLUMN",
													"reason": "Column does not support exclusion as it is a Primary Key"
												}
											}
										}
									},
									"table_2": {
										"name_in_destination": "table_2",
										"enabled": true,
										"supports_columns_config": true,
										"sync_mode": "SOFT_DELETE",
										"enabled_patch_settings": { "allowed": true },
										"columns": {
											"table_2_col_1": {
												"name_in_destination": "table_2_col_1",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": { "allowed": true }
											},
											"table_2_col_2": {
												"name_in_destination": "table_2_col_2",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": { "allowed": true }
											}
										}
									},
									"table_3": {
										"name_in_destination": "table_3",
										"enabled": true,
										"supports_columns_config": true,
										"sync_mode": "SOFT_DELETE",
										"enabled_patch_settings": { "allowed": true },
										"columns": {
											"table_3_col_1": {
												"name_in_destination": "table_3_col_1",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": { "allowed": true }
											}
										}
									}
								}
							}
						}
					}`

	schemaGetResponseAbsentTable :=  `
					{
						"enable_new_by_default": false,
						"schema_change_handling": "BLOCK_ALL",
						"schemas": {
							"public": {
								"name_in_destination": "public",
								"enabled": true,
								"tables": {
									"table_1": {
										"name_in_destination": "table_1",
										"enabled": false,
										"supports_columns_config": true,
										"sync_mode": "SOFT_DELETE",
										"enabled_patch_settings": { "allowed": true },
										"columns": {
											"table_1_col_1": {
												"name_in_destination": "table_1_col_1",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": {
													"allowed": false,
													"reason_code": "SYSTEM_COLUMN",
													"reason": "Column does not support exclusion as it is a Primary Key"
												}
											}
										}
									},
									"table_2": {
										"name_in_destination": "table_2",
										"enabled": true,
										"supports_columns_config": true,
										"sync_mode": "SOFT_DELETE",
										"enabled_patch_settings": { "allowed": true },
										"columns": {
											"table_2_col_1": {
												"name_in_destination": "table_2_col_1",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": { "allowed": true }
											},
											"table_2_col_2": {
												"name_in_destination": "table_2_col_2",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": { "allowed": true }
											}
										}
									}
								}
							}
						}
					}`

	schemaGetResponseWithAbsentColumn :=  `
					{
						"enable_new_by_default": false,
						"schema_change_handling": "BLOCK_ALL",
						"schemas": {
							"public": {
								"name_in_destination": "public",
								"enabled": true,
								"tables": {
									"table_1": {
										"name_in_destination": "table_1",
										"enabled": false,
										"supports_columns_config": true,
										"sync_mode": "SOFT_DELETE",
										"enabled_patch_settings": { "allowed": true },
										"columns": {
											"table_1_col_1": {
												"name_in_destination": "table_1_col_1",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": {
													"allowed": false,
													"reason_code": "SYSTEM_COLUMN",
													"reason": "Column does not support exclusion as it is a Primary Key"
												}
											}
										}
									},
									"table_2": {
										"name_in_destination": "table_2",
										"enabled": true,
										"supports_columns_config": true,
										"sync_mode": "SOFT_DELETE",
										"enabled_patch_settings": { "allowed": true },
										"columns": {
											"table_2_col_2": {
												"name_in_destination": "table_2_col_2",
												"enabled": true,
												"hashed": false,
												"enabled_patch_settings": { "allowed": true }
											}
										}
									}
								}
							}
						}
					}`

	setupMockClientLargeSchema := func(t *testing.T) {
		mockClient.Reset()

		// Mock GET handler - Simulates Aurora MySQL with multiple tables/columns WITHOUT is_primary_key
		schemaGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaResponseData), nil
			},
		)

		// Mock PATCH handler
		schemaPatchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaResponseData), nil
			},
		)

		schemaReloadPostHandler = mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/reload").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaResponseData), nil
			},
		)
	}

	resetInvocationCounts := func() {
		schemaGetHandler.Interactions = 0
		schemaPatchHandler.Interactions = 0
		schemaReloadPostHandler.Interactions = 0
	}

	step1 := resource.TestStep{
		PreConfig: func() {
			schemaResponseData = createMapFromJsonString(t, schemaGetResponse1)
		},
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
				schemas = {
					"public" = {
						enabled = true
						tables = {
							"table_2" = {
								sync_mode = "SOFT_DELETE"
								enabled = true
								columns = {
									"table_2_col_1" = { 
										enabled = true
										hashed = false
									}
									"table_2_col_2" = { 
										enabled = true
										hashed = false
									}
								}
							}
							"table_3" = {
								sync_mode = "SOFT_DELETE"
								enabled = true
								columns = {
									"table_3_col_1" = { 
										enabled = true
										hashed = false
									}
								}
							}
						}
					}
					"private" = {
						enabled = true
						tables = {
							"table_1" = {
								sync_mode = "SOFT_DELETE"
								enabled = true
								columns = {
									"table_1_col_1" = { 
										enabled = true
										hashed = false
									}
								}
							}
						}
					}
				}
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
					assertEqual(t, schemaGetHandler.Interactions, 3)
					assertEqual(t, schemaPatchHandler.Interactions, 0)
					assertEqual(t, schemaReloadPostHandler.Interactions, 0)
					return nil
				},
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.%", "2"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.%", "2"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.sync_mode", "SOFT_DELETE"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.%", "2"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_1.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_1.hashed", "false"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.hashed", "false"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.sync_mode", "SOFT_DELETE"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.%", "1"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.table_3_col_1.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.table_3_col_1.hashed", "false"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.private.tables.%", "1"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.private.tables.table_1.sync_mode", "SOFT_DELETE"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.private.tables.table_1.columns.%", "1"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.private.tables.table_1.columns.table_1_col_1.enabled", "true"),
			),
	}
    // Step 2: API returns less schemas
	step2 := resource.TestStep{
		PreConfig: func() {
			resetInvocationCounts()
			schemaResponseData = createMapFromJsonString(t, schemaGetResponseAbsentSchema)
		},
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
				schemas = {
					"public" = {
						enabled = true
						tables = {
							"table_2" = {
								sync_mode = "SOFT_DELETE"
								enabled = true
								columns = {
									"table_2_col_1" = { 
										enabled = true
										hashed = false
									}
									"table_2_col_2" = { 
										enabled = true
										hashed = false
									}
								}
							}
							"table_3" = {
								sync_mode = "SOFT_DELETE"
								enabled = true
								columns = {
									"table_3_col_1" = { 
										enabled = true
										hashed = false
									}
								}
							}
						}
					}
				}
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
					assertEqual(t, schemaGetHandler.Interactions, 4)
					assertEqual(t, schemaPatchHandler.Interactions, 0)
					assertEqual(t, schemaReloadPostHandler.Interactions, 0)
					return nil
				},
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.%", "1"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.%", "2"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.sync_mode", "SOFT_DELETE"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.%", "2"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_1.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_1.hashed", "false"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.hashed", "false"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.sync_mode", "SOFT_DELETE"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.%", "1"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.table_3_col_1.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.table_3_col_1.hashed", "false"),
			),
	}
	// Step 3: API returns less tables
	step3 := resource.TestStep{
		PreConfig: func() {
			resetInvocationCounts()
			schemaResponseData = createMapFromJsonString(t, schemaGetResponseAbsentTable)
		},
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
				schemas = {
					"public" = {
						enabled = true
						tables = {
							"table_2" = {
								sync_mode = "SOFT_DELETE"
								enabled = true
								columns = {
									"table_2_col_1" = { 
										enabled = true
										hashed = false
									}
									"table_2_col_2" = { 
										enabled = true
										hashed = false
									}
								}
							}
						}
					}
				}
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
					assertEqual(t, schemaGetHandler.Interactions, 4)
					assertEqual(t, schemaPatchHandler.Interactions, 0)
					return nil
				},
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.%", "1"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.sync_mode", "SOFT_DELETE"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.%", "2"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_1.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_1.hashed", "false"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.hashed", "false"),
			),
	}

	//Step 4: API returned less columns on consequent GET
	step4 := resource.TestStep{
		PreConfig: func() {
			resetInvocationCounts()
			schemaResponseData = createMapFromJsonString(t, schemaGetResponseWithAbsentColumn)
		},
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
				schemas = {
					"public" = {
						tables = {
							"table_2" = {
								sync_mode = "SOFT_DELETE"
								enabled = true
								columns = {
									"table_2_col_2" = { 
										enabled = true
										hashed = false
									}
								}
							}
						}
					}
				}
				#validation_level       = "TABLES"
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.%", "1"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.sync_mode", "SOFT_DELETE"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.%", "1"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.hashed", "false"),
			),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientLargeSchema(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
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

func TestResourceSchemaReloadUsesPreserveModeMock(t *testing.T) {
	var (
		schemaGetHandler   *mock.Handler
		schemaPatchHandler *mock.Handler
		schemaReloadPostHandler   *mock.Handler
		schemaResponseData map[string]interface{}
		schemasPatchRequestBody map[string]interface{}
		schemasReloadBody map[string]interface{}
	)
	
	schemaGetResponseAllTables :=  `
			{
				"enable_new_by_default": false,
				"schema_change_handling": "BLOCK_ALL",
				"schemas": {
					"public": {
						"name_in_destination": "public",
						"enabled": true,
						"tables": {
							"table_1": {
								"name_in_destination": "table_1",
								"enabled": false,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_1_col_1": {
										"name_in_destination": "table_1_col_1",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": {
											"allowed": false,
											"reason_code": "SYSTEM_COLUMN",
											"reason": "Column does not support exclusion as it is a Primary Key"
										}
									}
								}
							},
							"table_2": {
								"name_in_destination": "table_2",
								"enabled": true,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_2_col_1": {
										"name_in_destination": "table_2_col_1",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									},
									"table_2_col_2": {
										"name_in_destination": "table_2_col_2",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									}
								}
							},
							"table_3": {
								"name_in_destination": "table_3",
								"enabled": true,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_3_col_1": {
										"name_in_destination": "table_3_col_1",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									}
								}
							}
						}
					}
				}
			}`

	schemaGetResponseAbsentTable :=  `
			{
				"enable_new_by_default": false,
				"schema_change_handling": "BLOCK_ALL",
				"schemas": {
					"public": {
						"name_in_destination": "public",
						"enabled": true,
						"tables": {
							"table_1": {
								"name_in_destination": "table_1",
								"enabled": false,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_1_col_1": {
										"name_in_destination": "table_1_col_1",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": {
											"allowed": false,
											"reason_code": "SYSTEM_COLUMN",
											"reason": "Column does not support exclusion as it is a Primary Key"
										}
									}
								}
							},
							"table_2": {
								"name_in_destination": "table_2",
								"enabled": true,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_2_col_1": {
										"name_in_destination": "table_2_col_1",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									},
									"table_2_col_2": {
										"name_in_destination": "table_2_col_2",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									}
								}
							}
						}
					}
				}
			}`
	
	schemaGetResponseAllTablesAfterSchemaReload :=  `
			{
				"enable_new_by_default": false,
				"schema_change_handling": "BLOCK_ALL",
				"schemas": {
					"public": {
						"name_in_destination": "public",
						"enabled": true,
						"tables": {
							"table_1": {
								"name_in_destination": "table_1",
								"enabled": false,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_1_col_1": {
										"name_in_destination": "table_1_col_1",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": {
											"allowed": false,
											"reason_code": "SYSTEM_COLUMN",
											"reason": "Column does not support exclusion as it is a Primary Key"
										}
									}
								}
							},
							"table_2": {
								"name_in_destination": "table_2",
								"enabled": true,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_2_col_1": {
										"name_in_destination": "table_2_col_1",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									},
									"table_2_col_2": {
										"name_in_destination": "table_2_col_2",
										"enabled": false,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									},
									"table_2_col_3": {
										"name_in_destination": "table_2_col_3",
										"enabled": true,
										"hashed": true,
										"enabled_patch_settings": { "allowed": true }
									}
								}
							},
							"table_3": {
								"name_in_destination": "table_3",
								"enabled": false,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_3_col_1": {
										"name_in_destination": "table_3_col_1",
										"enabled": true,
										"hashed": true,
										"enabled_patch_settings": { "allowed": true }
									},
									"table_3_col_2": {
										"name_in_destination": "table_3_col_2",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									}
								}
							},
							"table_4": {
								"name_in_destination": "table_4",
								"enabled": true,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_4_col_1": {
										"name_in_destination": "table_4_col_1",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									}
								}
							}
						}
					}
				}
			}`
	
	schemaGetResponseAllTablesAfterPatch :=  `
			{
				"enable_new_by_default": false,
				"schema_change_handling": "BLOCK_ALL",
				"schemas": {
					"public": {
						"name_in_destination": "public",
						"enabled": true,
						"tables": {
							"table_1": {
								"name_in_destination": "table_1",
								"enabled": false,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_1_col_1": {
										"name_in_destination": "table_1_col_1",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": {
											"allowed": false,
											"reason_code": "SYSTEM_COLUMN",
											"reason": "Column does not support exclusion as it is a Primary Key"
										}
									}
								}
							},
							"table_2": {
								"name_in_destination": "table_2",
								"enabled": true,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_2_col_1": {
										"name_in_destination": "table_2_col_1",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									},
									"table_2_col_2": {
										"name_in_destination": "table_2_col_2",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									},
									"table_2_col_3": {
										"name_in_destination": "table_2_col_3",
										"enabled": false,
										"hashed": true,
										"enabled_patch_settings": { "allowed": true }
									}
								}
							},
							"table_3": {
								"name_in_destination": "table_3",
								"enabled": true,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_3_col_1": {
										"name_in_destination": "table_3_col_1",
										"enabled": true,
										"hashed": true,
										"enabled_patch_settings": { "allowed": true }
									},
									"table_3_col_2": {
										"name_in_destination": "table_3_col_2",
										"enabled": false,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									}
								}
							},
							"table_4": {
								"name_in_destination": "table_4",
								"enabled": false,
								"supports_columns_config": true,
								"sync_mode": "SOFT_DELETE",
								"enabled_patch_settings": { "allowed": true },
								"columns": {
									"table_4_col_1": {
										"name_in_destination": "table_4_col_1",
										"enabled": true,
										"hashed": false,
										"enabled_patch_settings": { "allowed": true }
									}
								}
							}
						}
					}
				}
			}`

	setupMockClient := func(t *testing.T) {
		mockClient.Reset()

		// Mock GET handler
		schemaGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaResponseData), nil
			},
		)

		// Mock PATCH handler
		schemaPatchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				schemasPatchRequestBody = requestBodyToJson(t, req)

				// return modified tables after PATCH /v1/schemas
				schemaResponseData = createMapFromJsonString(t, schemaGetResponseAllTablesAfterPatch)
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaResponseData), nil
			},
		)

		schemaReloadPostHandler = mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/reload").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				schemasReloadBody = requestBodyToJson(t, req)

				// return all the tables after /v1/schemas/reload
				schemaResponseData = createMapFromJsonString(t, schemaGetResponseAllTablesAfterSchemaReload)
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaResponseData), nil
			},
		)
	}

	resetInvocationCounts := func() {
		schemaGetHandler.Interactions = 0
		schemaPatchHandler.Interactions = 0
		schemaReloadPostHandler.Interactions = 0
		schemasPatchRequestBody = nil
		schemasReloadBody = nil
	}

	step1 := resource.TestStep{
		PreConfig: func() {
			schemaResponseData = createMapFromJsonString(t, schemaGetResponseAllTables)
		},
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
				schemas = {
					"public" = {
						enabled = true
						tables = {
							"table_2" = {
								sync_mode = "SOFT_DELETE"
								enabled = true
								columns = {
									"table_2_col_1" = { 
										enabled = true
										hashed = false
									}
									"table_2_col_2" = { 
										enabled = true
										hashed = false
									}
								}
							}
							"table_3" = {
								sync_mode = "SOFT_DELETE"
								enabled = true
								columns = {
									"table_3_col_1" = { 
										enabled = true
										hashed = false
									}
								}
							}
						}
					}
				}
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
					assertEqual(t, schemaGetHandler.Interactions, 3)
					assertEqual(t, schemaPatchHandler.Interactions, 0)
					assertEqual(t, schemaReloadPostHandler.Interactions, 0)
					return nil
				},
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.%", "1"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.%", "2"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.sync_mode", "SOFT_DELETE"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.%", "2"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_1.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_1.hashed", "false"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.hashed", "false"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.sync_mode", "SOFT_DELETE"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.%", "1"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.table_3_col_1.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.table_3_col_1.hashed", "false"),
			),
	}

	// Step 2: API returns less tables
	step2 := resource.TestStep{
		PreConfig: func() {
			resetInvocationCounts()
			schemaResponseData = createMapFromJsonString(t, schemaGetResponseAbsentTable)
		},
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"
				schemas = {
					"public" = {
						enabled = true
						tables = {
							"table_2" = {
								sync_mode = "SOFT_DELETE"
								enabled = true
								columns = {
									"table_2_col_1" = { 
										enabled = true
										hashed = false
									}
									"table_2_col_2" = { 
										enabled = true
										hashed = false
									}
								}
							}
							"table_3" = {
								sync_mode = "SOFT_DELETE"
								enabled = true
								columns = {
									"table_3_col_1" = { 
										enabled = true
										hashed = true
									}
								}
							}
						}
					}
				}
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
					assertEqual(t, schemasReloadBody["exclude_mode"], "PRESERVE")
					
					assertKeyExists(t, schemasPatchRequestBody, "schemas")
				 	assertKeyExists(t, schemasPatchRequestBody["schemas"].(map[string]interface {}), "public")
				 	patchSchema := schemasPatchRequestBody["schemas"].(map[string]interface {})["public"].(map[string]interface {})
					assertKeyExists(t, patchSchema, "tables")
					patchTables := patchSchema["tables"].(map[string]interface {})
					assertEqual(t, len(patchTables), 3)
					
					AssertKeyDoesNotExist(t, patchTables, "table_1")

					assertKeyExists(t, patchTables, "table_2")
					table2Patch := patchTables["table_2"].(map[string]interface {})
					AssertKeyDoesNotExist(t, table2Patch, "enabled")
					assertKeyExists(t, table2Patch, "columns")
					patchTable2Columns := table2Patch["columns"].(map[string]interface {})
					assertEqual(t, len(patchTable2Columns), 2)

					assertKeyExists(t, patchTable2Columns, "table_2_col_2")
					table2Col2Patch := patchTable2Columns["table_2_col_2"].(map[string]interface {})
					assertEqual(t, table2Col2Patch["enabled"], true)

					assertKeyExists(t, patchTable2Columns, "table_2_col_3")
					table2Col3Patch := patchTable2Columns["table_2_col_3"].(map[string]interface {})
					assertEqual(t, table2Col3Patch["enabled"], false)
					assertEqual(t, table2Col3Patch["is_primary_key"], nil)

					assertKeyExists(t, patchTables, "table_3")
					table3Patch := patchTables["table_3"].(map[string]interface {})
					assertEqual(t, table3Patch["enabled"], true)
					assertKeyExists(t, table3Patch, "columns")
					patchTable3Columns := table3Patch["columns"].(map[string]interface {})
					assertEqual(t, len(patchTable3Columns), 1)
					AssertKeyExists(t, patchTable3Columns, "table_3_col_2")
					table3Col2Patch := patchTable3Columns["table_3_col_2"].(map[string]interface {})
					assertEqual(t, table3Col2Patch["enabled"], false)
					assertEqual(t, table3Col2Patch["is_primary_key"], nil) 

					assertKeyExists(t, patchTables, "table_4")
					table4Patch := patchTables["table_4"].(map[string]interface {})
					assertEqual(t, len(table4Patch), 1)
					assertEqual(t, table4Patch["enabled"], false)

					assertEqual(t, schemaReloadPostHandler.Interactions, 1)
					assertEqual(t, schemaGetHandler.Interactions, 4)
					assertEqual(t, schemaPatchHandler.Interactions, 1)
					return nil
				},
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.%", "2"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.sync_mode", "SOFT_DELETE"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.%", "2"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_1.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_1.hashed", "false"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_2.columns.table_2_col_2.hashed", "false"),

				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.sync_mode", "SOFT_DELETE"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.%", "1"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.table_3_col_1.enabled", "true"),
				resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.public.tables.table_3.columns.table_3_col_1.hashed", "true"),
			),
		}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClient(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}