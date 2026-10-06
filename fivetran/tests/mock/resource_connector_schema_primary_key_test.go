package mock

import (
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const primaryKeyTfConfig = `
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	connector_id = "connector_id"
	schema_change_handling = "BLOCK_ALL"
	schemas = {
		"public" = {
			enabled = true
			tables = {
				"users" = {
					enabled = true
					columns = {
						"id" = {
							enabled = true
							is_primary_key = true
						}
						"name" = {
							enabled = true
						}
					}
				}
			}
		}
	}
}
`

func primaryKeySchemaResponse(idIsPrimaryKey string) string {
	return `
{
	"enable_new_by_default": false,
	"schemas": {
		"public": {
			"name_in_destination": "public",
			"enabled": true,
			"tables": {
				"users": {
					"name_in_destination": "users",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					},
					"columns": {
						"id": {
							"name_in_destination": "id",
							"enabled": true,
							"is_primary_key": ` + idIsPrimaryKey + `,
							"enabled_patch_settings": {
								"allowed": true
							}
						},
						"name": {
							"name_in_destination": "name",
							"enabled": true,
							"is_primary_key": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						}
					}
				}
			}
		}
	},
	"schema_change_handling": "BLOCK_ALL"
}
`
}

func setupPrimaryKeyMocks(t *testing.T, succeededAt interface{}, initialIsPrimaryKey string, patchBodies *[]map[string]interface{}) {
	var schemaData map[string]interface{}

	mockClient.Reset()

	mockClient.When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]interface{}{
				"id":           "connector_id",
				"service":      "s3",
				"schema":       "public",
				"succeeded_at": succeededAt,
			}), nil
		},
	)

	mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if schemaData == nil {
				schemaData = createMapFromJsonString(t, primaryKeySchemaResponse(initialIsPrimaryKey))
			}
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
		},
	)

	mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			*patchBodies = append(*patchBodies, requestBodyToJson(t, req))
			schemaData = createMapFromJsonString(t, primaryKeySchemaResponse("true"))
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
		},
	)
}

func assertPatchSetsPrimaryKey(t *testing.T, patchBodies []map[string]interface{}) {
	t.Helper()
	assertNotEmpty(t, patchBodies)
	if len(patchBodies) == 0 {
		return
	}
	body := patchBodies[0]
	schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})
	public := assertKeyExists(t, schemas, "public").(map[string]interface{})
	tables := assertKeyExists(t, public, "tables").(map[string]interface{})
	users := assertKeyExists(t, tables, "users").(map[string]interface{})
	columns := assertKeyExists(t, users, "columns").(map[string]interface{})
	id := assertKeyExists(t, columns, "id").(map[string]interface{})
	assertKeyExistsAndHasValue(t, id, "is_primary_key", true)
}

// is_primary_key is sent to the API and read back into state for a not yet synced file connector
func TestResourceSchemaPrimaryKeyPreSync(t *testing.T) {
	var patchBodies []map[string]interface{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				patchBodies = nil
				setupPrimaryKeyMocks(t, nil, "false", &patchBodies)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: primaryKeyTfConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertPatchSetsPrimaryKey(t, patchBodies)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema",
							"schemas.public.tables.users.columns.id.is_primary_key", "true"),
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema",
							"schemas.public.tables.users.columns.name.is_primary_key", "false"),
					),
				},
			},
		},
	)
}

// changing is_primary_key on an already synced file connector fails before any PATCH is sent
func TestResourceSchemaPrimaryKeyPostSyncChangeFails(t *testing.T) {
	var patchBodies []map[string]interface{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				patchBodies = nil
				setupPrimaryKeyMocks(t, time.Now().Add(-24*time.Hour).Format(time.RFC3339), "false", &patchBodies)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      primaryKeyTfConfig,
					ExpectError: regexp.MustCompile("`is_primary_key` can no longer be changed"),
				},
				{
					// no PATCH was sent by the failed step
					Config:   primaryKeyTfConfig,
					PlanOnly: true,
					PreConfig: func() {
						assertEqual(t, len(patchBodies), 0)
					},
					ExpectNonEmptyPlan: true,
				},
			},
		},
	)
}

// an unchanged is_primary_key on an already synced file connector applies normally
func TestResourceSchemaPrimaryKeyPostSyncUnchanged(t *testing.T) {
	var patchBodies []map[string]interface{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				patchBodies = nil
				setupPrimaryKeyMocks(t, time.Now().Add(-24*time.Hour).Format(time.RFC3339), "true", &patchBodies)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: primaryKeyTfConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema",
							"schemas.public.tables.users.columns.id.is_primary_key", "true"),
					),
				},
			},
		},
	)
}

// the schemas response may not contain columns, they are fetched in batches to detect a changed is_primary_key
func TestResourceSchemaPrimaryKeyPostSyncColumnsFetched(t *testing.T) {
	var patchBodies []map[string]interface{}
	var batchHandler, perTableHandler *mock.Handler
	schemaWithoutColumns := `
{
	"enable_new_by_default": false,
	"schemas": {
		"public": {
			"name_in_destination": "public",
			"enabled": true,
			"tables": {
				"users": {
					"name_in_destination": "users",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				}
			}
		}
	},
	"schema_change_handling": "BLOCK_ALL"
}
`
	columnsResponse := `
{
	"columns": {
		"id": {
			"name_in_destination": "id",
			"enabled": true,
			"is_primary_key": false,
			"enabled_patch_settings": {
				"allowed": true
			}
		},
		"name": {
			"name_in_destination": "name",
			"enabled": true,
			"is_primary_key": false,
			"enabled_patch_settings": {
				"allowed": true
			}
		}
	}
}
`

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				patchBodies = nil
				mockClient.Reset()
				mockClient.When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]interface{}{
							"id":           "connector_id",
							"service":      "s3",
							"schema":       "public",
							"succeeded_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
						}), nil
					},
				)
				mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, schemaWithoutColumns)), nil
					},
				)
				batchHandler = mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/public/fetch-source-columns").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]interface{}{
							"tables": map[string]interface{}{"users": createMapFromJsonString(t, columnsResponse)},
						}), nil
					},
				)
				perTableHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas/public/tables/users/columns").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, columnsResponse)), nil
					},
				)
				mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						patchBodies = append(patchBodies, requestBodyToJson(t, req))
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, schemaWithoutColumns)), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      primaryKeyTfConfig,
					ExpectError: regexp.MustCompile("`is_primary_key` can no longer be changed"),
				},
				{
					Config:   primaryKeyTfConfig,
					PlanOnly: true,
					PreConfig: func() {
						assertEqual(t, len(patchBodies), 0)
						// columns come from the batch endpoint, as in column validation
						if batchHandler.Interactions == 0 {
							t.Errorf("expected columns to be fetched with the batch endpoint")
						}
						assertEqual(t, perTableHandler.Interactions, 0)
					},
					ExpectNonEmptyPlan: true,
				},
			},
		},
	)
}

// when connector_id changes (e.g. the connector was recreated), the new connector is updated, not the one in state
func TestResourceSchemaConnectorIdChangeUpdatesNewConnector(t *testing.T) {
	tfConfig := func(connectorId string) string {
		return `
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	connector_id = "` + connectorId + `"
	schema_change_handling = "BLOCK_ALL"
	schemas = {
		"public" = {
			enabled = true
			tables = {
				"users" = {
					enabled = true
				}
			}
		}
	}
}
`
	}
	patched := map[string]int{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				for _, id := range []string{"connector_id", "connector_id_2"} {
					id := id
					// the new connector starts with the table disabled, so updating it requires a PATCH
					upstream := createMapFromJsonString(t, primaryKeySchemaResponse("false"))
					if id == "connector_id_2" {
						upstream["schemas"].(map[string]interface{})["public"].(map[string]interface{})["tables"].(map[string]interface{})["users"].(map[string]interface{})["enabled"] = false
					}
					mockClient.When(http.MethodGet, "/v1/connections/"+id+"/schemas").ThenCall(
						func(req *http.Request) (*http.Response, error) {
							return fivetranSuccessResponse(t, req, http.StatusOK, "Success", upstream), nil
						},
					)
					mockClient.When(http.MethodPatch, "/v1/connections/"+id+"/schemas").ThenCall(
						func(req *http.Request) (*http.Response, error) {
							patched[id]++
							upstream = createMapFromJsonString(t, primaryKeySchemaResponse("false"))
							return fivetranSuccessResponse(t, req, http.StatusOK, "Success", upstream), nil
						},
					)
				}
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: tfConfig("connector_id"),
				},
				{
					PreConfig: func() {
						patched = map[string]int{}
					},
					Config: tfConfig("connector_id_2"),
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertEqual(t, patched["connector_id"], 0)
							assertEqual(t, patched["connector_id_2"], 1)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id_2"),
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "id", "connector_id_2"),
					),
				},
			},
		},
	)
}

// an unconfigured (computed) is_primary_key that changed upstream after sync doesn't block unrelated updates
func TestResourceSchemaPrimaryKeyComputedValueDoesntBlockUpdate(t *testing.T) {
	tfConfig := func(nameEnabled bool) string {
		return `
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	connector_id = "connector_id"
	schema_change_handling = "BLOCK_ALL"
	schemas = {
		"public" = {
			enabled = true
			tables = {
				"users" = {
					enabled = true
					columns = {
						"id" = {
							enabled = true
						}
						"name" = {
							enabled = ` + map[bool]string{true: "true", false: "false"}[nameEnabled] + `
						}
					}
				}
			}
		}
	}
}
`
	}
	var succeededAt interface{}
	var schemaData map[string]interface{}
	patches := 0

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				succeededAt = nil
				schemaData = createMapFromJsonString(t, primaryKeySchemaResponse("true"))
				mockClient.When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]interface{}{
							"id":           "connector_id",
							"service":      "s3",
							"schema":       "public",
							"succeeded_at": succeededAt,
						}), nil
					},
				)
				mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)
				mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						patches++
						body := requestBodyToJson(t, req)
						name := body["schemas"].(map[string]interface{})["public"].(map[string]interface{})["tables"].(map[string]interface{})["users"].(map[string]interface{})["columns"].(map[string]interface{})["name"].(map[string]interface{})
						columns := schemaData["schemas"].(map[string]interface{})["public"].(map[string]interface{})["tables"].(map[string]interface{})["users"].(map[string]interface{})["columns"].(map[string]interface{})
						columns["name"].(map[string]interface{})["enabled"] = name["enabled"]
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					// is_primary_key isn't configured, state gets the upstream value
					Config: tfConfig(true),
					Check: resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema",
						"schemas.public.tables.users.columns.id.is_primary_key", "true"),
				},
				{
					// connector synced and upstream primary key changed outside of terraform
					PreConfig: func() {
						succeededAt = time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
						columns := schemaData["schemas"].(map[string]interface{})["public"].(map[string]interface{})["tables"].(map[string]interface{})["users"].(map[string]interface{})["columns"].(map[string]interface{})
						columns["id"].(map[string]interface{})["is_primary_key"] = false
						patches = 0
					},
					Config: tfConfig(false),
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertEqual(t, patches, 1)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema",
							"schemas.public.tables.users.columns.name.enabled", "false"),
					),
				},
			},
		},
	)
}

// with group_id + connector_name and an unchanged configuration, a recreated connector is resolved at plan time and updated
func TestResourceSchemaGroupAndNameRecreatedConnectorUpdatesNewConnector(t *testing.T) {
	tfConfig := `
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	group_id = "group_id"
	connector_name = "public"
	schema_change_handling = "BLOCK_ALL"
	schemas = {
		"public" = {
			enabled = true
			tables = {
				"users" = {
					enabled = true
				}
			}
		}
	}
}
`
	currentConnectorId := "connector_id"
	patched := map[string]int{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				mockClient.When(http.MethodGet, "/v1/connections").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]interface{}{
							"items":       []interface{}{map[string]interface{}{"id": currentConnectorId, "group_id": "group_id", "schema": "public"}},
							"next_cursor": nil,
						}), nil
					},
				)
				for _, id := range []string{"connector_id", "connector_id_2"} {
					id := id
					// the new connector starts with the table disabled, so updating it requires a PATCH
					upstream := createMapFromJsonString(t, primaryKeySchemaResponse("false"))
					if id == "connector_id_2" {
						upstream["schemas"].(map[string]interface{})["public"].(map[string]interface{})["tables"].(map[string]interface{})["users"].(map[string]interface{})["enabled"] = false
					}
					mockClient.When(http.MethodGet, "/v1/connections/"+id+"/schemas").ThenCall(
						func(req *http.Request) (*http.Response, error) {
							return fivetranSuccessResponse(t, req, http.StatusOK, "Success", upstream), nil
						},
					)
					mockClient.When(http.MethodPatch, "/v1/connections/"+id+"/schemas").ThenCall(
						func(req *http.Request) (*http.Response, error) {
							patched[id]++
							upstream = createMapFromJsonString(t, primaryKeySchemaResponse("false"))
							return fivetranSuccessResponse(t, req, http.StatusOK, "Success", upstream), nil
						},
					)
				}
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: tfConfig,
					Check:  resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id"),
				},
				{
					PreConfig: func() {
						currentConnectorId = "connector_id_2"
						patched = map[string]int{}
					},
					Config: tfConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertEqual(t, patched["connector_id"], 0)
							assertEqual(t, patched["connector_id_2"], 1)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id_2"),
					),
				},
			},
		},
	)
}

// adding a column to an existing resource doesn't plan a null is_primary_key that conflicts with the upstream value
func TestResourceSchemaPrimaryKeyAddColumnMock(t *testing.T) {
	cfg := func(extra string) string {
		return `
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	connector_id = "connector_id"
	schema_change_handling = "ALLOW_ALL"
	schemas = {
		"public" = {
			enabled = true
			tables = {
				"users" = {
					enabled = true
					columns = {
						"id" = {
							enabled = true
						}` + extra + `
					}
				}
			}
		}
	}
}
`
	}
	upstream := `
{
	"schema_change_handling": "ALLOW_ALL",
	"schemas": {"public": {"name_in_destination": "public", "enabled": true, "tables": {"users": {
		"name_in_destination": "users", "enabled": true, "enabled_patch_settings": {"allowed": true},
		"columns": {
			"id":   {"name_in_destination": "id",   "enabled": true, "is_primary_key": true,  "enabled_patch_settings": {"allowed": true}},
			"name": {"name_in_destination": "name", "enabled": true, "is_primary_key": false, "enabled_patch_settings": {"allowed": true}}
		}}}}}
}`
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			mockClient.Reset()
			mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, upstream)), nil
			})
			mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, upstream)), nil
			})
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: cfg("")},
			{Config: cfg(`
						"name" = {
							enabled = true
						}`)},
		},
	})
}

// a recreated connector without schema settings yet is reloaded on update, the same way as on create.
// validation_level = "NONE" skips the plan-time reload, as when the connector is recreated in the same apply
// and connector_id is unknown at plan time
func TestResourceSchemaConnectorIdChangeNewConnectorWithoutSchemaMock(t *testing.T) {
	tfConfig := func(connectorId string) string {
		return `
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	connector_id = "` + connectorId + `"
	schema_change_handling = "BLOCK_ALL"
	validation_level = "NONE"
	schemas = {
		"public" = {
			enabled = true
			tables = {
				"users" = {
					enabled = true
				}
			}
		}
	}
}
`
	}
	var newConnectorSchema map[string]interface{}
	reloads := 0
	patches := 0

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, primaryKeySchemaResponse("false"))), nil
					},
				)
				mockClient.When(http.MethodGet, "/v1/connections/connector_id_2/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if newConnectorSchema == nil {
							return fivetranResponse(t, req, "NotFound_SchemaConfig", http.StatusNotFound,
								"Connector with id 'connector_id_2' doesn't have schema config", nil), nil
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", newConnectorSchema), nil
					},
				)
				mockClient.When(http.MethodPost, "/v1/connections/connector_id_2/schemas/reload").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						reloads++
						// reloaded schema has the table disabled, so it needs a PATCH
						newConnectorSchema = createMapFromJsonString(t, primaryKeySchemaResponse("false"))
						newConnectorSchema["schemas"].(map[string]interface{})["public"].(map[string]interface{})["tables"].(map[string]interface{})["users"].(map[string]interface{})["enabled"] = false
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", newConnectorSchema), nil
					},
				)
				mockClient.When(http.MethodPatch, "/v1/connections/connector_id_2/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						patches++
						newConnectorSchema = createMapFromJsonString(t, primaryKeySchemaResponse("false"))
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", newConnectorSchema), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: tfConfig("connector_id"),
				},
				{
					Config: tfConfig("connector_id_2"),
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertEqual(t, reloads, 1)
							assertEqual(t, patches, 1)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id_2"),
					),
				},
			},
		},
	)
}
