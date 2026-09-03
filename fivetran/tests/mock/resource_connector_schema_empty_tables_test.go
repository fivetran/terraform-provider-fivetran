package mock

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const schemaConfigEmptyTablesMapJson = `
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

// Reproduces "Provider produced inconsistent result after apply": with
// schema_change_handling = "ALLOW_ALL" every enabled upstream table matches the
// policy default and is filtered out of state, so an explicitly configured empty
// `tables = {}` map must be kept as an empty map, not degraded to null.
func TestResourceSchemaConfigEmptyTablesMapMock(t *testing.T) {
	var schemaData map[string]interface{}

	setupMockClient := func(t *testing.T) {
		mockClient.Reset()
		schemaData = nil

		mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if nil == schemaData {
					schemaData = createMapFromJsonString(t, schemaConfigEmptyTablesMapJson)
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
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
						enabled = true
						tables = {}
					}
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_ALL"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.schema_1.tables.%", "0"),
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
				// there is no possibility to destroy schema config - it always exists within the connector
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
