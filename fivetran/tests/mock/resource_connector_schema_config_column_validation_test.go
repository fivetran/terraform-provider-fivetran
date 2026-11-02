package mock

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Order has a single column, AccountId, both in the schema config and in the columns endpoint.
const columnValidationSchema = `
	{
		"enable_new_by_default": false,
		"schema_change_handling": "BLOCK_ALL",
		"schemas": {
			"salesforce": {
				"name_in_destination": "salesforce",
				"enabled": true,
				"tables": {
					"Order": {
						"name_in_destination": "order",
						"enabled": true,
						"supports_columns_config": true,
						"enabled_patch_settings": { "allowed": true },
						"columns": {
							"AccountId": {
								"name_in_destination": "account_id",
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

const columnValidationTableColumns = `
	{
		"columns": {
			"AccountId": {
				"name_in_destination": "account_id",
				"enabled": true,
				"hashed": false,
				"enabled_patch_settings": { "allowed": true }
			}
		}
	}`

func columnValidationConfig(columns string) string {
	return `
	resource "fivetran_connector_schema_config" "test_schema" {
		provider = fivetran-provider
		connector_id = "connector_id"
		schema_change_handling = "BLOCK_ALL"
		validation_level = "COLUMNS"
		schemas = {
			"salesforce" = {
				enabled = true
				tables = {
					"Order" = {
						enabled = true
						columns = {` + columns + `
						}
					}
				}
			}
		}
	}`
}

func setupColumnValidationHandlers(t *testing.T) (patchHandler *mock.Handler) {
	mockClient.Reset()
	mockClient.When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]interface{}{
				"id":      "connector_id",
				"service": "salesforce",
			}), nil
		},
	)
	mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
				createMapFromJsonString(t, columnValidationSchema)), nil
		},
	)
	mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/reload").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
				createMapFromJsonString(t, columnValidationSchema)), nil
		},
	)
	// Batch columns endpoint unavailable: validation falls back to the per-table columns endpoint.
	mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/salesforce/fetch-source-columns").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranResponse(t, req, "NotImplemented", http.StatusNotImplemented, "Not implemented", nil), nil
		},
	)
	mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas/salesforce/tables/Order/columns").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
				createMapFromJsonString(t, columnValidationTableColumns)), nil
		},
	)
	return mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
				createMapFromJsonString(t, columnValidationSchema)), nil
		},
	)
}

// RD-1245106: with validation_level = "COLUMNS", a column that doesn't exist in the source
// fails the plan, and nothing is sent to the API.
func TestResourceSchemaConfigColumnsValidationRejectsNonExistentColumnMock(t *testing.T) {
	var patchHandler *mock.Handler

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			patchHandler = setupColumnValidationHandlers(t)
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: columnValidationConfig(`
							"AccountId" = { enabled = true }
							"non_existent" = { enabled = true }`),
				ExpectError: regexp.MustCompile("Column `non_existent` of table with name `Order` not found in\\s+source schema `salesforce`"),
			},
		},
	})

	assertEqual(t, patchHandler.Interactions, 0)
}

// Same setup with only existing columns: validation passes and the apply succeeds.
func TestResourceSchemaConfigColumnsValidationAcceptsExistingColumnsMock(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			setupColumnValidationHandlers(t)
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: columnValidationConfig(`
							"AccountId" = { enabled = true }`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "validation_level", "COLUMNS"),
					resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schemas.salesforce.tables.Order.columns.AccountId.enabled", "true"),
				),
			},
		},
	})
}
