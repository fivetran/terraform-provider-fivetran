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

func setupReloadUnsupportedSchemaHandlers(t *testing.T, notFound bool) (reloadHandler, patchHandler *mock.Handler) {
	mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if notFound {
				return fivetranResponse(t, req,
					"NotFound_SchemaConfig", http.StatusNotFound,
					"Connector with id 'connector_id' doesn't have schema config", nil), nil
			}
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
				createMapFromJsonString(t, reloadUnsupportedCurrentSchema)), nil
		},
	)
	patchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			t.Errorf("unexpected PATCH /schemas call — schema config can't be aligned without a reload")
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
		},
	)
	return reloadHandler, patchHandler
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

// runReloadUnsupportedTest asserts that an unsupported reload fails with expectError and never
// PATCHes a schema that is missing configured elements.
func runReloadUnsupportedTest(t *testing.T, setup func() (reloadHandler, patchHandler *mock.Handler), expectedReloadCalls bool, expectError *regexp.Regexp) {
	var reloadHandler, patchHandler *mock.Handler

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				reloadHandler, patchHandler = setup()
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				{
					Config:      reloadUnsupportedConfig,
					ExpectError: expectError,
				},
			},
		},
	)

	if expectedReloadCalls && reloadHandler.Interactions == 0 {
		t.Errorf("expected POST /schemas/reload to be attempted")
	}
	if !expectedReloadCalls {
		assertEqual(t, reloadHandler.Interactions, 0)
	}
	assertEqual(t, patchHandler.Interactions, 0)
}

func reloadForbiddenHandler(t *testing.T) *mock.Handler {
	return mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/reload").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			t.Errorf("unexpected POST /schemas/reload call — reload is not supported for this connection type")
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
		},
	)
}

// Metadata says reload isn't supported and table_2 is missing from the current schema: the
// original validation error must be reported instead of silently applying without table_2.
func TestResourceSchemaConfigReloadUnsupportedByMetadataMock(t *testing.T) {
	runReloadUnsupportedTest(t, func() (*mock.Handler, *mock.Handler) {
		setupReloadUnsupportedConnectionDetails(t, "reload_unsupported_by_metadata")
		setupReloadUnsupportedMetadata(t, "reload_unsupported_by_metadata", false)
		_, patchHandler := setupReloadUnsupportedSchemaHandlers(t, false)
		return reloadForbiddenHandler(t), patchHandler
	}, false, regexp.MustCompile(`does\s+not\s+support\s+schema\s+reload,\s+so\s+the\s+schema\s+configuration\s+can't\s+be(?s:.*)validation_level(?s:.*)table_2`))
}

// Metadata says reload isn't supported and the connection has no schema settings yet: there's
// no state to proceed with, so it must fail without calling reload.
func TestResourceSchemaConfigReloadUnsupportedNoSchemaConfigMock(t *testing.T) {
	runReloadUnsupportedTest(t, func() (*mock.Handler, *mock.Handler) {
		setupReloadUnsupportedConnectionDetails(t, "reload_unsupported_no_schema")
		setupReloadUnsupportedMetadata(t, "reload_unsupported_no_schema", false)
		_, patchHandler := setupReloadUnsupportedSchemaHandlers(t, true)
		return reloadForbiddenHandler(t), patchHandler
	}, false, regexp.MustCompile(`does not support schema reload and the connection\s+doesn't\s+have\s+schema\s+settings\s+yet`))
}

// Metadata doesn't say anything about reload support: fall back to the reload API's
// NotSupported_SchemaReload response and still report the original validation error.
func TestResourceSchemaConfigReloadUnsupportedByApiMock(t *testing.T) {
	runReloadUnsupportedTest(t, func() (*mock.Handler, *mock.Handler) {
		// No metadata handler: the metadata lookup fails and reloadSchema falls back to the API response.
		setupReloadUnsupportedConnectionDetails(t, "reload_unsupported_by_api")
		_, patchHandler := setupReloadUnsupportedSchemaHandlers(t, false)
		reloadHandler := mockClient.When(http.MethodPost, "/v1/connections/connector_id/schemas/reload").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranResponse(t, req,
					"NotSupported_SchemaReload", http.StatusBadRequest,
					"Schema reload is not supported for this connection type", nil), nil
			},
		)
		return reloadHandler, patchHandler
	}, true, regexp.MustCompile(`does\s+not\s+support\s+schema\s+reload,\s+so\s+the\s+schema\s+configuration\s+can't\s+be(?s:.*)validation_level(?s:.*)table_2(?s:.*)API message: Schema reload is not supported`))
}
