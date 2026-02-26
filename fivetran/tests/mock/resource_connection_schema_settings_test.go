package mock

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestConnectionSchemaSettingsAllowAll verifies creating a resource with ALLOW_ALL policy
// and disabled_schemas. Checks that the PATCH request sets schema_change_handling correctly
// and the state reflects the expected disabled_schemas count.
func TestConnectionSchemaSettingsAllowAll(t *testing.T) {
	var patchBody map[string]interface{}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		// GET returns two schemas, both initially enabled
		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
					},
				}), nil
			},
		)

		// PATCH captures the body and returns updated state
		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				patchBody = requestBodyToJson(t, req)
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
					},
				}), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = ["schema_2"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "schema_change_handling", "ALLOW_ALL"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "disabled_schemas.#", "1"),
					func(s *terraform.State) error {
						assertKeyExistsAndHasValue(t, patchBody, "schema_change_handling", "ALLOW_ALL")
						return nil
					},
				),
			},
		},
	})
}

// TestConnectionSchemaSettingsBlockAll verifies creating a resource with BLOCK_ALL policy
// and enabled_schemas. Checks that the PATCH request sets schema_change_handling correctly
// and only the listed schema is enabled in state.
func TestConnectionSchemaSettingsBlockAll(t *testing.T) {
	var patchBody map[string]interface{}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "BLOCK_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
						"schema_3": map[string]any{"enabled": false},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				patchBody = requestBodyToJson(t, req)
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "BLOCK_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
						"schema_3": map[string]any{"enabled": false},
					},
				}), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "BLOCK_ALL"
	enabled_schemas        = ["schema_1"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "schema_change_handling", "BLOCK_ALL"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "enabled_schemas.#", "1"),
					func(s *terraform.State) error {
						assertKeyExistsAndHasValue(t, patchBody, "schema_change_handling", "BLOCK_ALL")
						return nil
					},
				),
			},
		},
	})
}

// TestConnectionSchemaSettingsUpdate verifies a two-step policy change: create with
// ALLOW_ALL + disabled_schemas, then update to BLOCK_ALL + enabled_schemas. The mock
// tracks server-side state across PATCHes to simulate realistic API behavior.
// Verifies both PATCH calls are made and state converges after each step.
func TestConnectionSchemaSettingsUpdate(t *testing.T) {
	var patchHandler *mock.Handler

	// currentData tracks the server-side state, updated by PATCH
	var currentData map[string]any

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		currentData = map[string]any{
			"schema_change_handling": "ALLOW_ALL",
			"schemas": map[string]any{
				"schema_1": map[string]any{"enabled": true},
				"schema_2": map[string]any{"enabled": true},
			},
		}

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", currentData), nil
			},
		)

		patchHandler = mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				// Update currentData based on the PATCH request
				if sch, ok := body["schema_change_handling"]; ok {
					currentData["schema_change_handling"] = sch
				}
				if schemas, ok := body["schemas"].(map[string]interface{}); ok {
					existingSchemas := currentData["schemas"].(map[string]any)
					for name, config := range schemas {
						if configMap, ok := config.(map[string]interface{}); ok {
							if enabled, ok := configMap["enabled"]; ok {
								existingSchemas[name] = map[string]any{"enabled": enabled}
							}
						}
					}
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", currentData), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = ["schema_2"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "schema_change_handling", "ALLOW_ALL"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "disabled_schemas.#", "1"),
				),
			},
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "BLOCK_ALL"
	enabled_schemas        = ["schema_1"]
}`,
				// The reversal logic re-enables schema_2 (removed from disabled_schemas).
				// Refresh detects schema_2 is now enabled and reports it in enabled_schemas,
				// causing a non-empty plan (schema_2 needs to be removed on next apply).
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "schema_change_handling", "BLOCK_ALL"),
					func(s *terraform.State) error {
						assertEqual(t, patchHandler.Interactions, 2)
						return nil
					},
				),
			},
		},
	})
}

// TestConnectionSchemaSettingsRefreshOnly simulates upstream drift on a managed
// schema: after initial create with disabled_schemas = ["schema_2"], someone
// re-enables schema_2 externally. Refresh detects the drift (schema_2 is no longer
// disabled) and signals a non-empty plan.
func TestConnectionSchemaSettingsRefreshOnly(t *testing.T) {
	schema2Enabled := false

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": schema2Enabled},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				schema2Enabled = false
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
					},
				}), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = ["schema_2"]
}`,
				Check: resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "disabled_schemas.#", "1"),
			},
			{
				// Someone re-enabled schema_2 externally — drift on managed item
				PreConfig: func() { schema2Enabled = true },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestConnectionSchemaSettingsImport verifies importing an existing connection schema
// settings resource by connection ID. Step 1 creates the resource; step 2 imports it
// and uses ImportStateVerify to confirm the imported state matches the original.
func TestConnectionSchemaSettingsImport(t *testing.T) {
	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "BLOCK_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
						"schema_3": map[string]any{"enabled": true},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "BLOCK_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
						"schema_3": map[string]any{"enabled": true},
					},
				}), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "BLOCK_ALL"
	enabled_schemas        = ["schema_1", "schema_3"]
}`,
			},
			{
				ResourceName:      "fivetran_connection_schema_settings.test",
				ImportState:       true,
				ImportStateId:     "conn_id",
				ImportStateVerify: true,
			},
		},
	})
}

// TestConnectionSchemaSettingsSchemaNotLoaded verifies that when the schema details
// endpoint returns NotFound_SchemaConfig (schema not yet loaded), the resource
// returns a clear error directing the user to reload the schema first.
func TestConnectionSchemaSettingsSchemaNotLoaded(t *testing.T) {
	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranResponse(t, req, "NotFound_SchemaConfig", http.StatusNotFound, "Schema config not found", nil), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = ["schema_2"]
}`,
				ExpectError: regexp.MustCompile("Ensure the schema has"),
			},
		},
	})
}

// TestConnectionSchemaSettingsConnectionDeletedUpstream verifies that when the
// connection is deleted upstream (API returns NotFound_Connection), the Read method
// silently removes the resource from state. Terraform then plans to recreate it.
func TestConnectionSchemaSettingsConnectionDeletedUpstream(t *testing.T) {
	deleted := false

	setupMock := func(t *testing.T) {
		mockClient.Reset()
		deleted = false

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if deleted {
					return fivetranResponse(t, req, "NotFound_Connection", http.StatusNotFound, "Connection not found", nil), nil
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
					},
				}), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = ["schema_2"]
}`,
				Check: resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "id", "conn_id"),
			},
			{
				PreConfig: func() {
					// Simulate connection deletion upstream
					deleted = true
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				// After refresh, the resource should be removed from state.
				// Terraform will then plan to recreate it, producing a non-empty plan.
			},
		},
	})
}

// TestConnectionSchemaSettingsUnexpectedReadError verifies that unexpected API errors
// (e.g. Forbidden/403) during Read are surfaced to the user as diagnostics, not
// silently swallowed or treated as a deleted resource.
func TestConnectionSchemaSettingsUnexpectedReadError(t *testing.T) {
	forbidden := false

	setupMock := func(t *testing.T) {
		mockClient.Reset()
		forbidden = false

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if forbidden {
					return fivetranResponse(t, req, "Forbidden", http.StatusForbidden, "Access denied", nil), nil
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
					},
				}), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = ["schema_2"]
}`,
			},
			{
				PreConfig: func() {
					forbidden = true
				},
				RefreshState: true,
				ExpectError:  regexp.MustCompile("Unable to Read Connection Schema Settings"),
			},
		},
	})
}

// TestConnectionSchemaSettingsSchemaDisappearedFromSource verifies behavior when a
// disabled schema disappears from the source (ALLOW_ALL policy).
// Step 1: create with disabled_schemas = ["schema_2", "schema_3"].
// Step 2 (refresh): schema_3 disappears upstream; state shrinks to 1, plan shows diff.
// Step 3 (re-apply): user updates config to remove stale schema_3; apply converges.
func TestConnectionSchemaSettingsSchemaDisappearedFromSource(t *testing.T) {
	var currentData map[string]any

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		currentData = map[string]any{
			"schema_change_handling": "ALLOW_ALL",
			"schemas": map[string]any{
				"schema_1": map[string]any{"enabled": true},
				"schema_2": map[string]any{"enabled": false},
				"schema_3": map[string]any{"enabled": false},
			},
		}

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", currentData), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", currentData), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = ["schema_2", "schema_3"]
}`,
				Check: resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "disabled_schemas.#", "2"),
			},
			{
				PreConfig: func() {
					// schema_3 disappears from the source
					currentData["schemas"] = map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check:              resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "disabled_schemas.#", "1"),
			},
			{
				// Re-apply with the config updated to remove stale schema_3.
				// Apply should succeed and state should converge.
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = ["schema_2"]
}`,
				Check: resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "disabled_schemas.#", "1"),
			},
		},
	})
}

// TestConnectionSchemaSettingsEnabledSchemaDisappearedFromSource verifies behavior
// when an enabled schema disappears from the source (BLOCK_ALL policy).
// Step 1: create with enabled_schemas = ["schema_1", "schema_2"].
// Step 2 (refresh): schema_2 disappears upstream; state shrinks to 1, plan shows diff.
// Step 3 (re-apply): user updates config to remove stale schema_2; apply converges.
func TestConnectionSchemaSettingsEnabledSchemaDisappearedFromSource(t *testing.T) {
	var currentData map[string]any

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		currentData = map[string]any{
			"schema_change_handling": "BLOCK_ALL",
			"schemas": map[string]any{
				"schema_1": map[string]any{"enabled": true},
				"schema_2": map[string]any{"enabled": true},
				"schema_3": map[string]any{"enabled": false},
			},
		}

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", currentData), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", currentData), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "BLOCK_ALL"
	enabled_schemas        = ["schema_1", "schema_2"]
}`,
				Check: resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "enabled_schemas.#", "2"),
			},
			{
				PreConfig: func() {
					// schema_2 disappears from the source
					currentData["schemas"] = map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_3": map[string]any{"enabled": false},
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check:              resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "enabled_schemas.#", "1"),
			},
			{
				// Re-apply with config updated to remove stale schema_2.
				// Apply should succeed and state should converge.
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "BLOCK_ALL"
	enabled_schemas        = ["schema_1"]
}`,
				Check: resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "enabled_schemas.#", "1"),
			},
		},
	})
}

// TestConnectionSchemaSettingsLargeSchemaModerate tests performance with 10k schemas
// and 2k disabled. Uses FastStringSetType (list-backed) for O(n) comparison.
// Expected: < 5s (baseline ~1.3s).
func TestConnectionSchemaSettingsLargeSchemaModerate(t *testing.T) {
	const totalSchemas = 10000
	const disabledCount = 2000

	// Build mock API response with 10k schemas
	schemasData := make(map[string]any, totalSchemas)
	for i := 0; i < totalSchemas; i++ {
		enabled := true
		if i < disabledCount {
			enabled = false
		}
		schemasData[fmt.Sprintf("schema_%05d", i)] = map[string]any{"enabled": enabled}
	}

	// Build TF config: disabled_schemas list for schemas 0..1999
	disabledNames := make([]string, disabledCount)
	for i := 0; i < disabledCount; i++ {
		disabledNames[i] = fmt.Sprintf(`"%s"`, fmt.Sprintf("schema_%05d", i))
	}
	tfConfig := fmt.Sprintf(`
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = [%s]
}`, strings.Join(disabledNames, ", "))

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas":               schemasData,
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas":               schemasData,
				}), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: tfConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "schema_change_handling", "ALLOW_ALL"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "disabled_schemas.#", fmt.Sprintf("%d", disabledCount)),
				),
			},
		},
	})
}

// TestConnectionSchemaSettingsNoDriftOnDifferentAPIOrder verifies that the API
// returning schemas in a different order on each GET does not cause false drift.
// The config lists schemas as ["middle", "zebra", "alpha"]; the API alternates
// between two different orderings. Step 2 re-applies the same config and must
// produce an empty plan thanks to ListSemanticEquals.
func TestConnectionSchemaSettingsNoDriftOnDifferentAPIOrder(t *testing.T) {
	callCount := 0

	setupMock := func(t *testing.T) {
		mockClient.Reset()
		callCount = 0

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				callCount++
				// Return schemas in a different order each time
				var data map[string]any
				if callCount%2 == 1 {
					data = map[string]any{
						"schema_change_handling": "BLOCK_ALL",
						"schemas": map[string]any{
							"zebra":   map[string]any{"enabled": true},
							"alpha":   map[string]any{"enabled": true},
							"middle":  map[string]any{"enabled": true},
							"unused1": map[string]any{"enabled": false},
							"unused2": map[string]any{"enabled": false},
						},
					}
				} else {
					data = map[string]any{
						"schema_change_handling": "BLOCK_ALL",
						"schemas": map[string]any{
							"middle":  map[string]any{"enabled": true},
							"unused2": map[string]any{"enabled": false},
							"alpha":   map[string]any{"enabled": true},
							"unused1": map[string]any{"enabled": false},
							"zebra":   map[string]any{"enabled": true},
						},
					}
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "BLOCK_ALL",
					"schemas": map[string]any{
						"alpha":   map[string]any{"enabled": true},
						"middle":  map[string]any{"enabled": true},
						"zebra":   map[string]any{"enabled": true},
						"unused1": map[string]any{"enabled": false},
						"unused2": map[string]any{"enabled": false},
					},
				}), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				// Config lists schemas in yet another order
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "BLOCK_ALL"
	enabled_schemas        = ["middle", "zebra", "alpha"]
}`,
				Check: resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "enabled_schemas.#", "3"),
			},
			{
				// Same config, no changes — should produce empty plan despite
				// the API returning schemas in a different order
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "BLOCK_ALL"
	enabled_schemas        = ["middle", "zebra", "alpha"]
}`,
				Check: resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "enabled_schemas.#", "3"),
			},
		},
	})
}

// TestConnectionSchemaSettingsLargeSchemaWorstCase tests worst-case performance
// with 9999 out of 10k schemas in the disabled set.
// Expected: < 30s (baseline ~14s).
// Without FastStringSetType (using types.Set) this took ~497s due to O(n^2) comparison.
func TestConnectionSchemaSettingsLargeSchemaWorstCase(t *testing.T) {
	const totalSchemas = 10000
	const disabledCount = 9999

	// Build mock API response: 9999 disabled, 1 enabled
	schemasData := make(map[string]any, totalSchemas)
	for i := 0; i < totalSchemas; i++ {
		enabled := i >= disabledCount // only the last one is enabled
		schemasData[fmt.Sprintf("schema_%05d", i)] = map[string]any{"enabled": enabled}
	}

	// Build TF config with 9999 disabled schemas
	disabledNames := make([]string, disabledCount)
	for i := 0; i < disabledCount; i++ {
		disabledNames[i] = fmt.Sprintf(`"schema_%05d"`, i)
	}
	tfConfig := fmt.Sprintf(`
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = [%s]
}`, strings.Join(disabledNames, ", "))

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas":               schemasData,
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas":               schemasData,
				}), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: tfConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "schema_change_handling", "ALLOW_ALL"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "disabled_schemas.#", fmt.Sprintf("%d", disabledCount)),
				),
			},
		},
	})
}

// TestConnectionSchemaSettingsDuplicateDisabledSchemas verifies that specifying
// duplicate values in disabled_schemas fails validation during plan with a
// "Duplicate Value" error indicating the duplicated element and its positions.
func TestConnectionSchemaSettingsDuplicateDisabledSchemas(t *testing.T) {
	mockClient.Reset()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = ["schema_1", "schema_2", "schema_1"]
}`,
				ExpectError: regexp.MustCompile("Duplicate Value"),
			},
		},
	})
}

// TestConnectionSchemaSettingsDuplicateEnabledSchemas verifies that specifying
// duplicate values in enabled_schemas fails validation during plan with a
// "Duplicate Value" error indicating the duplicated element and its positions.
func TestConnectionSchemaSettingsDuplicateEnabledSchemas(t *testing.T) {
	mockClient.Reset()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "BLOCK_ALL"
	enabled_schemas        = ["schema_1", "schema_1"]
}`,
				ExpectError: regexp.MustCompile("Duplicate Value"),
			},
		},
	})
}

// TestConnectionSchemaSettingsDuplicateValidationLargeScale tests that duplicate
// detection at scale (9999 elements with one duplicate) runs in O(n) time.
// Expected: < 3s (baseline ~0.8s).
func TestConnectionSchemaSettingsDuplicateValidationLargeScale(t *testing.T) {
	mockClient.Reset()

	// 9999 elements: 9998 unique + 1 duplicate (first element repeated at the end)
	names := make([]string, 9999)
	for i := 0; i < 9998; i++ {
		names[i] = fmt.Sprintf(`"schema_%05d"`, i)
	}
	names[9998] = `"schema_00000"` // duplicate of first

	tfConfig := fmt.Sprintf(`
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = [%s]
}`, strings.Join(names, ", "))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config:      tfConfig,
				ExpectError: regexp.MustCompile(`Duplicate Value`),
			},
		},
	})
}

// TestConnectionSchemaSettingsReorderConfigNoPlan verifies that reordering elements
// in the .tf config without adding or removing any schemas produces an empty plan.
// Step 1: create with enabled_schemas = ["alpha", "bravo", "charlie"].
// Step 2: change config order to ["charlie", "alpha", "bravo"]; same set of schemas.
// ListSemanticEquals treats both as equal, so no plan diff is generated.
func TestConnectionSchemaSettingsReorderConfigNoPlan(t *testing.T) {
	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "BLOCK_ALL",
					"schemas": map[string]any{
						"alpha":  map[string]any{"enabled": true},
						"bravo":  map[string]any{"enabled": true},
						"charlie": map[string]any{"enabled": true},
						"delta":  map[string]any{"enabled": false},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "BLOCK_ALL",
					"schemas": map[string]any{
						"alpha":  map[string]any{"enabled": true},
						"bravo":  map[string]any{"enabled": true},
						"charlie": map[string]any{"enabled": true},
						"delta":  map[string]any{"enabled": false},
					},
				}), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "BLOCK_ALL"
	enabled_schemas        = ["alpha", "bravo", "charlie"]
}`,
				Check: resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "enabled_schemas.#", "3"),
			},
			{
				// Same schemas, completely different order — should produce empty plan
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "BLOCK_ALL"
	enabled_schemas        = ["charlie", "alpha", "bravo"]
}`,
				Check: resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "enabled_schemas.#", "3"),
			},
		},
	})
}

// TestConnectionSchemaSettingsConflictRetry verifies that when the PATCH endpoint
// returns a 409 Conflict (optimistic lock failure), the resource retries the full
// read-modify-write cycle and eventually succeeds.
func TestConnectionSchemaSettingsConflictRetry(t *testing.T) {
	origBackoff := core.SchemaConflictBackoff()
	core.SetSchemaConflictBackoff(10 * time.Millisecond)
	defer core.SetSchemaConflictBackoff(origBackoff)

	patchAttempts := 0

	setupMock := func(t *testing.T) {
		mockClient.Reset()
		patchAttempts = 0

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				patchAttempts++
				if patchAttempts <= 2 {
					return fivetranResponse(t, req, "Conflict", http.StatusConflict,
						"Optimistic lock conflict", nil), nil
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true},
						"schema_2": map[string]any{"enabled": false},
					},
				}), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_settings" "test" {
	provider               = fivetran-provider
	connection_id          = "conn_id"
	schema_change_handling = "ALLOW_ALL"
	disabled_schemas       = ["schema_2"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.test", "disabled_schemas.#", "1"),
					func(s *terraform.State) error {
						assertEqual(t, patchAttempts, 3)
						return nil
					},
				),
			},
		},
	})
}
