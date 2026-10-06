package mock

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const reloadUnsupportedCurrentSchema = `
	{
		"enable_new_by_default": false,
		"schema_change_handling": "ALLOW_ALL",
		"schemas": {
			"public": {
				"name_in_destination": "public",
				"enabled": true,
				"tables": {
					"table_1": {
						"name_in_destination": "table_1",
						"enabled": true,
						"supports_columns_config": true,
						"enabled_patch_settings": { "allowed": true },
						"columns": {
							"table_1_col_1": {
								"name_in_destination": "table_1_col_1",
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

// Upstream state after PATCH; table_2 shows up once the source picks it up, keeping the
// post-apply read consistent with the plan.
const reloadUnsupportedPatchedSchema = `
	{
		"enable_new_by_default": false,
		"schema_change_handling": "ALLOW_ALL",
		"schemas": {
			"public": {
				"name_in_destination": "public",
				"enabled": true,
				"tables": {
					"table_1": {
						"name_in_destination": "table_1",
						"enabled": true,
						"supports_columns_config": true,
						"enabled_patch_settings": { "allowed": true },
						"columns": {
							"table_1_col_1": {
								"name_in_destination": "table_1_col_1",
								"enabled": true,
								"hashed": true,
								"enabled_patch_settings": { "allowed": true }
							}
						}
					},
					"table_2": {
						"name_in_destination": "table_2",
						"enabled": true,
						"supports_columns_config": true,
						"enabled_patch_settings": { "allowed": true }
					}
				}
			}
		}
	}`

// table_2 isn't in the current schema, so validation asks for a reload.
const reloadUnsupportedConfig = `
	resource "fivetran_connector_schema_config" "test_schema" {
		provider = fivetran-provider
		connector_id = "connector_id"
		schema_change_handling = "ALLOW_ALL"
		schemas = {
			"public" = {
				enabled = true
				tables = {
					"table_1" = {
						enabled = true
						columns = {
							"table_1_col_1" = {
								enabled = true
								hashed = true
							}
						}
					}
					"table_2" = {
						enabled = true
					}
				}
			}
		}
	}`

func setupReloadUnsupportedConnectionDetails(t *testing.T, service string) {
	mockClient.When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]interface{}{
				"id":      "connector_id",
				"service": service,
			}), nil
		},
	)
}

func setupReloadUnsupportedMetadata(t *testing.T, service string, supportsSchemaReload bool) *mock.Handler {
	return mockClient.When(http.MethodGet, "/v1/metadata/connector-types/"+service).ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]interface{}{
				"id":                     service,
				"name":                   service,
				"type":                   "Database",
				"supports_schema_reload": supportsSchemaReload,
			}), nil
		},
	)
}

// Metadata says reload isn't supported: reload must not be called and planned edits must be
// applied against the current schema instead of a zero-value response.
func TestResourceSchemaConfigReloadUnsupportedByMetadataMock(t *testing.T) {
	var (
		reloadHandler     *mock.Handler
		patchHandler      *mock.Handler
		schemaGetResponse map[string]interface{}
		patchRequestBody  map[string]interface{}
	)

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				patchRequestBody = nil
				schemaGetResponse = createMapFromJsonString(t, reloadUnsupportedCurrentSchema)

				setupReloadUnsupportedConnectionDetails(t, "reload_unsupported_by_metadata")
				setupReloadUnsupportedMetadata(t, "reload_unsupported_by_metadata", false)

				mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaGetResponse), nil
					},
				)
				reloadHandler = mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/reload").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						t.Errorf("unexpected POST /schemas/reload call — reload is not supported for this connection type")
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaGetResponse), nil
					},
				)
				patchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						patchRequestBody = requestBodyToJson(t, req)
						schemaGetResponse = createMapFromJsonString(t, reloadUnsupportedPatchedSchema)
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaGetResponse), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				{
					Config: reloadUnsupportedConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertEqual(t, reloadHandler.Interactions, 0)
							// A single PATCH carrying the planned column edit — no spurious
							// schema_change_handling-only PATCH caused by a zeroed-out response.
							assertEqual(t, patchHandler.Interactions, 1)
							assertNotEmpty(t, patchRequestBody)
							_, hasSch := patchRequestBody["schema_change_handling"]
							assertEqual(t, hasSch, false)
							schemasBody := assertKeyExists(t, patchRequestBody, "schemas").(map[string]interface{})
							publicBody := assertKeyExists(t, schemasBody, "public").(map[string]interface{})
							tablesBody := assertKeyExists(t, publicBody, "tables").(map[string]interface{})
							table1Body := assertKeyExists(t, tablesBody, "table_1").(map[string]interface{})
							columnsBody := assertKeyExists(t, table1Body, "columns").(map[string]interface{})
							col1Body := assertKeyExists(t, columnsBody, "table_1_col_1").(map[string]interface{})
							assertEqual(t, col1Body["hashed"], true)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "id", "connector_id"),
					),
				},
			},
		},
	)
}

// Metadata says reload isn't supported and the connection has no schema settings yet: there's
// no state to proceed with, so it must fail without calling reload.
func TestResourceSchemaConfigReloadUnsupportedNoSchemaConfigMock(t *testing.T) {
	var (
		reloadHandler *mock.Handler
		patchHandler  *mock.Handler
	)

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()

				setupReloadUnsupportedConnectionDetails(t, "reload_unsupported_no_schema")
				setupReloadUnsupportedMetadata(t, "reload_unsupported_no_schema", false)

				mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranResponse(t, req,
							"NotFound_SchemaConfig", http.StatusNotFound,
							"Connector with id 'connector_id' doesn't have schema config", nil), nil
					},
				)
				reloadHandler = mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/reload").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						t.Errorf("unexpected POST /schemas/reload call — reload is not supported for this connection type")
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
					},
				)
				patchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						t.Errorf("unexpected PATCH /schemas call")
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, reloadHandler.Interactions, 0)
				assertEqual(t, patchHandler.Interactions, 0)
				return nil
			},
			Steps: []resource.TestStep{
				{
					Config:      reloadUnsupportedConfig,
					ExpectError: regexp.MustCompile(`does not support schema reload and the connection\s+doesn't\s+have\s+schema\s+settings\s+yet`),
				},
			},
		},
	)
}

// Metadata doesn't say anything about reload support: fall back to the reload API's
// NotSupported_SchemaReload response and proceed with the current schema.
func TestResourceSchemaConfigReloadUnsupportedByApiMock(t *testing.T) {
	var (
		reloadHandler     *mock.Handler
		patchHandler      *mock.Handler
		schemaGetResponse map[string]interface{}
		patchRequestBody  map[string]interface{}
	)

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				patchRequestBody = nil
				schemaGetResponse = createMapFromJsonString(t, reloadUnsupportedCurrentSchema)

				// No metadata handler: the metadata lookup fails and reloadSchema falls back to the API response.
				setupReloadUnsupportedConnectionDetails(t, "reload_unsupported_by_api")

				mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaGetResponse), nil
					},
				)
				reloadHandler = mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/reload").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranResponse(t, req,
							"NotSupported_SchemaReload", http.StatusBadRequest,
							"Schema reload is not supported for this connection type", nil), nil
					},
				)
				patchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						patchRequestBody = requestBodyToJson(t, req)
						schemaGetResponse = createMapFromJsonString(t, reloadUnsupportedPatchedSchema)
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaGetResponse), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				{
					Config: reloadUnsupportedConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							if reloadHandler.Interactions == 0 {
								t.Errorf("expected POST /schemas/reload to be attempted")
							}
							assertEqual(t, patchHandler.Interactions, 1)
							_, hasSch := patchRequestBody["schema_change_handling"]
							assertEqual(t, hasSch, false)
							schemasBody := assertKeyExists(t, patchRequestBody, "schemas").(map[string]interface{})
							publicBody := assertKeyExists(t, schemasBody, "public").(map[string]interface{})
							tablesBody := assertKeyExists(t, publicBody, "tables").(map[string]interface{})
							table1Body := assertKeyExists(t, tablesBody, "table_1").(map[string]interface{})
							columnsBody := assertKeyExists(t, table1Body, "columns").(map[string]interface{})
							col1Body := assertKeyExists(t, columnsBody, "table_1_col_1").(map[string]interface{})
							assertEqual(t, col1Body["hashed"], true)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "id", "connector_id"),
					),
				},
			},
		},
	)
}
