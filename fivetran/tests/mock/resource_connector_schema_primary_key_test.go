package mock

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestResourceSchemaPrimaryKeyPreSync tests that is_primary_key can be set before sync
func TestResourceSchemaPrimaryKeyPreSync(t *testing.T) {
	tfConfig := `
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

	jsonResponse := `
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
							"is_primary_key": true,
							"enabled_patch_settings": {
								"allowed": true
							}
						},
						"name": {
							"name_in_destination": "name",
							"enabled": true,
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

	var schemaData map[string]interface{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				schemaData = nil

				// Mock connector details - NOT synced yet (SucceededAt = zero time)
				mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						connectorResponse := map[string]interface{}{
							"data": map[string]interface{}{
								"id":                     "connector_id",
								"service":                "postgres",  // Not a file connector, but we're testing the schema works
								"connected_at":          time.Now().Format(time.RFC3339),
								"succeeded_at":          nil,          // Not synced yet
								"schema":                "public",
								"schema_change_handling": "BLOCK_ALL",
							},
							"code":    "Success",
							"message": "Success",
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorResponse), nil
					},
				)

				mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if nil == schemaData {
							schemaData = createMapFromJsonString(t, jsonResponse)
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)

				mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: tfConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "connector_id", "connector_id"),
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
						// Verify primary key was set
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema",
							"schemas.public.tables.users.columns.id.is_primary_key", "true"),
					),
				},
			},
		},
	)
}

// TestResourceSchemaPrimaryKeyPostSync tests that primary key changes post-sync are warned about
func TestResourceSchemaPrimaryKeyPostSync(t *testing.T) {
	tfConfig := `
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
					}
				}
			}
		}
	}
}
`

	jsonResponse := `
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
							"is_primary_key": true,
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

	var schemaData map[string]interface{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				schemaData = nil

				// Mock connector details - ALREADY synced (SucceededAt = past time)
				mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						connectorResponse := map[string]interface{}{
							"data": map[string]interface{}{
								"id":                "connector_id",
								"service":          "postgres",
								"connected_at":     time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
								"succeeded_at":     time.Now().Add(-24 * time.Hour).Format(time.RFC3339), // Synced 24 hours ago
								"schema":           "public",
								"schema_change_handling": "BLOCK_ALL",
							},
							"code":    "Success",
							"message": "Success",
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorResponse), nil
					},
				)

				mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if nil == schemaData {
							schemaData = createMapFromJsonString(t, jsonResponse)
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)

				mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: tfConfig,
					// Should succeed but show warnings about post-sync primary key configuration
					ExpectWarnings: []string{
						"Primary Key Configuration After Sync",
					},
				},
			},
		},
	)
}
