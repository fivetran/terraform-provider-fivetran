package mock

import (
	"fmt"
	"net/http"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestConnectionSchemaConfigAllowAllDisabledTables verifies creating a resource
// with ALLOW_ALL policy and disabled_tables. The PATCH is sent to the per-schema endpoint
// and the state shows the correct disabled_tables count.
func TestConnectionSchemaConfigAllowAllDisabledTables(t *testing.T) {
	var patchBody map[string]interface{}
	table2Enabled := true

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
								"table_2": map[string]any{"enabled": table2Enabled, "sync_mode": "LIVE"},
								"table_3": map[string]any{"enabled": true, "sync_mode": "LIVE"},
							},
						},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				patchBody = requestBodyToJson(t, req)
				table2Enabled = false
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
								"table_2": map[string]any{"enabled": false, "sync_mode": "LIVE"},
								"table_3": map[string]any{"enabled": true, "sync_mode": "LIVE"},
							},
						},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	disabled_tables = ["table_2"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "disabled_tables.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "id", "conn_id:schema_1"),
					func(s *terraform.State) error {
						assertNotEmpty(t, patchBody)
						return nil
					},
				),
			},
		},
	})
}

// TestConnectionSchemaConfigBlockAllEnabledTables verifies creating a resource
// with BLOCK_ALL policy and enabled_tables. Only the listed table is enabled.
func TestConnectionSchemaConfigBlockAllEnabledTables(t *testing.T) {
	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "BLOCK_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
								"table_2": map[string]any{"enabled": false, "sync_mode": "LIVE"},
								"table_3": map[string]any{"enabled": false, "sync_mode": "LIVE"},
							},
						},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "BLOCK_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
								"table_2": map[string]any{"enabled": false, "sync_mode": "LIVE"},
								"table_3": map[string]any{"enabled": false, "sync_mode": "LIVE"},
							},
						},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	enabled_tables = ["table_1"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "enabled_tables.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "id", "conn_id:schema_1"),
				),
			},
		},
	})
}

// TestConnectionSchemaConfigSyncMode verifies that sync_mode map is correctly
// sent in the PATCH request and reflected in state. The mock tracks state changes
// so the second GET (after PATCH) returns updated sync modes.
func TestConnectionSchemaConfigSyncMode(t *testing.T) {
	var patchBody map[string]interface{}
	t1SyncMode := "LIVE"
	t2SyncMode := "LIVE"

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": t1SyncMode},
								"table_2": map[string]any{"enabled": true, "sync_mode": t2SyncMode},
							},
						},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				patchBody = requestBodyToJson(t, req)
				if tables, ok := patchBody["tables"].(map[string]interface{}); ok {
					if t1, ok := tables["table_1"].(map[string]interface{}); ok {
						if sm, ok := t1["sync_mode"].(string); ok {
							t1SyncMode = sm
						}
					}
					if t2, ok := tables["table_2"].(map[string]interface{}); ok {
						if sm, ok := t2["sync_mode"].(string); ok {
							t2SyncMode = sm
						}
					}
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": t1SyncMode},
								"table_2": map[string]any{"enabled": true, "sync_mode": t2SyncMode},
							},
						},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	sync_mode = {
		"table_1" = "HISTORY"
		"table_2" = "SOFT_DELETE"
	}
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "sync_mode.table_1", "HISTORY"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "sync_mode.table_2", "SOFT_DELETE"),
					func(s *terraform.State) error {
						assertNotEmpty(t, patchBody)
						return nil
					},
				),
			},
		},
	})
}

// TestConnectionSchemaConfigUpdate verifies updating tables and sync modes
// across two steps. Step 1 disables table_2 and sets sync_mode for table_1.
// Step 2 changes to disable table_3 instead and updates sync_mode.
func TestConnectionSchemaConfigUpdate(t *testing.T) {
	tableState := map[string]map[string]any{
		"table_1": {"enabled": true, "sync_mode": "LIVE"},
		"table_2": {"enabled": true, "sync_mode": "LIVE"},
		"table_3": {"enabled": true, "sync_mode": "LIVE"},
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				tables := map[string]any{}
				for k, v := range tableState {
					tables[k] = map[string]any{"enabled": v["enabled"], "sync_mode": v["sync_mode"]}
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true, "tables": tables},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				if tables, ok := body["tables"].(map[string]interface{}); ok {
					for tName, tVal := range tables {
						if tMap, ok := tVal.(map[string]interface{}); ok {
							if en, ok := tMap["enabled"]; ok {
								tableState[tName]["enabled"] = en
							}
							if sm, ok := tMap["sync_mode"]; ok {
								tableState[tName]["sync_mode"] = sm
							}
						}
					}
				}
				tables := map[string]any{}
				for k, v := range tableState {
					tables[k] = map[string]any{"enabled": v["enabled"], "sync_mode": v["sync_mode"]}
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{"enabled": true, "tables": tables},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	disabled_tables = ["table_2"]
	sync_mode = {
		"table_1" = "HISTORY"
	}
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "disabled_tables.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "sync_mode.table_1", "HISTORY"),
				),
			},
			{
				Config: `
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	disabled_tables = ["table_3"]
	sync_mode = {
		"table_1" = "SOFT_DELETE"
	}
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "disabled_tables.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "sync_mode.table_1", "SOFT_DELETE"),
				),
			},
		},
	})
}

// TestConnectionSchemaConfigImport verifies importing a resource
// with "connection_id:schema_name" format. State should be populated from API.
func TestConnectionSchemaConfigImport(t *testing.T) {
	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
								"table_2": map[string]any{"enabled": false, "sync_mode": "HISTORY"},
							},
						},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
								"table_2": map[string]any{"enabled": false, "sync_mode": "HISTORY"},
							},
						},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	disabled_tables = ["table_2"]
}`,
			},
			{
				ResourceName:            "fivetran_connection_schema_config.test",
				ImportState:             true,
				ImportStateId:           "conn_id:schema_1",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sync_mode"},
			},
		},
	})
}

// TestConnectionSchemaConfigSchemaNotLoaded verifies that when the API returns
// NotFound_SchemaConfig, the resource reports an error asking the user to reload.
func TestConnectionSchemaConfigSchemaNotLoaded(t *testing.T) {
	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranResponse(t, req, "NotFound_SchemaConfig", http.StatusNotFound,
					"Schema config not found", nil), nil
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	disabled_tables = ["table_1"]
}`,
				ExpectError: regexp.MustCompile("Schema details not available|fivetran_connection_schema_reload"),
			},
		},
	})
}

// TestConnectionSchemaConfigSchemaNotFound verifies that when the API succeeds
// but the requested schema is not in the response, the resource reports an error.
func TestConnectionSchemaConfigSchemaNotFound(t *testing.T) {
	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"other_schema": map[string]any{
							"enabled": true,
							"tables":  map[string]any{},
						},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	disabled_tables = ["table_1"]
}`,
				ExpectError: regexp.MustCompile("not found"),
			},
		},
	})
}

// TestConnectionSchemaConfigConnectionDeletedUpstream verifies that when
// the connection is deleted upstream (NotFound_Connection), the resource is
// silently removed from state on refresh.
// TestConnectionSchemaConfigConflictRetry verifies that when the PATCH endpoint
// returns a 409 Conflict (optimistic lock failure), the resource retries the full
// read-modify-write cycle and eventually succeeds.
func TestConnectionSchemaConfigConflictRetry(t *testing.T) {
	origBackoff := core.SchemaConflictBackoff()
	core.SetSchemaConflictBackoff(10 * time.Millisecond)
	defer core.SetSchemaConflictBackoff(origBackoff)

	patchAttempts := 0
	table2Enabled := true

	setupMock := func(t *testing.T) {
		mockClient.Reset()
		patchAttempts = 0
		table2Enabled = true

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
								"table_2": map[string]any{"enabled": table2Enabled, "sync_mode": "LIVE"},
							},
						},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				patchAttempts++
				if patchAttempts <= 2 {
					return fivetranResponse(t, req, "Conflict", http.StatusConflict,
						"Optimistic lock conflict", nil), nil
				}
				table2Enabled = false
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
								"table_2": map[string]any{"enabled": false, "sync_mode": "LIVE"},
							},
						},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	disabled_tables = ["table_2"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "disabled_tables.#", "1"),
					func(s *terraform.State) error {
						assertEqual(t, patchAttempts, 3)
						return nil
					},
				),
			},
		},
	})
}

func TestConnectionSchemaConfigConnectionDeletedUpstream(t *testing.T) {
	getCallCount := 0

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				getCallCount++
				if getCallCount <= 3 {
					return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
						"schema_change_handling": "ALLOW_ALL",
						"schemas": map[string]any{
							"schema_1": map[string]any{
								"enabled": true,
								"tables": map[string]any{
									"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
								},
							},
						},
					}), nil
				}
				return fivetranResponse(t, req, "NotFound_Connection", http.StatusNotFound,
					"Connection not found", nil), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
							},
						},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
}`,
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestConnectionSchemaConfigSyncModeTracksOnlyConfiguredTables verifies that
// sync_mode in state only contains tables explicitly listed in the user's config.
// The schema has 3 tables with sync_modes, but the user only configures sync_mode
// for table_1. Step 1: create — state should have sync_mode only for table_1.
// Step 2: the API changes sync_mode on untracked table_3 from LIVE to HISTORY;
// a plan-only check should show no drift from untracked tables.
func TestConnectionSchemaConfigSyncModeTracksOnlyConfiguredTables(t *testing.T) {
	table3SyncMode := "LIVE"

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": "HISTORY"},
								"table_2": map[string]any{"enabled": true, "sync_mode": "SOFT_DELETE"},
								"table_3": map[string]any{"enabled": true, "sync_mode": table3SyncMode},
							},
						},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{"enabled": true, "sync_mode": "HISTORY"},
								"table_2": map[string]any{"enabled": true, "sync_mode": "SOFT_DELETE"},
								"table_3": map[string]any{"enabled": true, "sync_mode": table3SyncMode},
							},
						},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	sync_mode = {
		"table_1" = "HISTORY"
	}
}`,
				Check: resource.ComposeTestCheckFunc(
					// Only table_1 should be in sync_mode — table_2 and table_3 are untracked
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "sync_mode.%", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "sync_mode.table_1", "HISTORY"),
					resource.TestCheckNoResourceAttr("fivetran_connection_schema_config.test", "sync_mode.table_2"),
					resource.TestCheckNoResourceAttr("fivetran_connection_schema_config.test", "sync_mode.table_3"),
				),
			},
			{
				// Simulate upstream change: table_3 sync_mode changed from LIVE to HISTORY.
				// Since table_3 is not in the user's sync_mode config, this should NOT cause drift.
				PreConfig: func() { table3SyncMode = "HISTORY" },
				Config: `
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	sync_mode = {
		"table_1" = "HISTORY"
	}
}`,
				PlanOnly: true,
			},
		},
	})
}

// TestConnectionSchemaConfigTableDroppedFromSource verifies that when a table
// configured in disabled_tables and sync_mode is dropped from the source and
// disappears from the API response, the resource handles it gracefully:
// - Read removes the dropped table from state (no error)
// - Terraform detects drift and plans to reconcile
func TestConnectionSchemaConfigTableDroppedFromSource(t *testing.T) {
	tableDropped := false

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				tables := map[string]any{
					"table_1": map[string]any{"enabled": true, "sync_mode": "HISTORY"},
					"table_2": map[string]any{"enabled": false, "sync_mode": "LIVE"},
					"table_3": map[string]any{"enabled": true, "sync_mode": "LIVE"},
				}
				if tableDropped {
					// table_2 was dropped from source — no longer in API response
					delete(tables, "table_2")
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables":  tables,
						},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				tables := map[string]any{
					"table_1": map[string]any{"enabled": true, "sync_mode": "HISTORY"},
					"table_2": map[string]any{"enabled": false, "sync_mode": "LIVE"},
					"table_3": map[string]any{"enabled": true, "sync_mode": "LIVE"},
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables":  tables,
						},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	disabled_tables = ["table_2"]
	sync_mode = {
		"table_2" = "LIVE"
	}
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "disabled_tables.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "sync_mode.table_2", "LIVE"),
				),
			},
			{
				// table_2 dropped from source — Read should remove it from state,
				// causing drift (disabled_tables becomes empty, sync_mode loses table_2)
				PreConfig:          func() { tableDropped = true },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestConnectionSchemaConfigNewTableAppearsWithWrongState verifies drift detection
// when a new table appears in the API with an enabled state that contradicts the
// resource's policy-based expectation.
// Under ALLOW_ALL with disabled_tables = ["table_2"], all tables except table_2
// should be enabled. When table_3 appears with enabled=false (e.g. disabled by
// an external process), Read picks it up in disabled_tables — causing drift because
// the config only lists table_2. The subsequent apply should re-enable table_3.
func TestConnectionSchemaConfigNewTableAppearsWithWrongState(t *testing.T) {
	newTablePresent := false
	// Track table states for stateful mock
	table2Enabled := true
	table3Enabled := false // new table arrives disabled under ALLOW_ALL — wrong state

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				tables := map[string]any{
					"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
					"table_2": map[string]any{"enabled": table2Enabled, "sync_mode": "LIVE"},
				}
				if newTablePresent {
					tables["table_3"] = map[string]any{"enabled": table3Enabled, "sync_mode": "LIVE"}
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables":  tables,
						},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				if tables, ok := body["tables"].(map[string]interface{}); ok {
					for name, val := range tables {
						if tMap, ok := val.(map[string]interface{}); ok {
							if en, ok := tMap["enabled"]; ok {
								switch name {
								case "table_2":
									table2Enabled = en.(bool)
								case "table_3":
									table3Enabled = en.(bool)
								}
							}
						}
					}
				}
				tables := map[string]any{
					"table_1": map[string]any{"enabled": true, "sync_mode": "LIVE"},
					"table_2": map[string]any{"enabled": table2Enabled, "sync_mode": "LIVE"},
				}
				if newTablePresent {
					tables["table_3"] = map[string]any{"enabled": table3Enabled, "sync_mode": "LIVE"}
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables":  tables,
						},
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
				// Step 1: Create — table_1 enabled, table_2 disabled
				Config: `
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	disabled_tables = ["table_2"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "disabled_tables.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "disabled_tables.0", "table_2"),
				),
			},
			{
				// Step 2: New table_3 appears with enabled=false (wrong for ALLOW_ALL).
				// Refresh detects drift: disabled_tables now has [table_2, table_3] but
				// config only lists [table_2].
				PreConfig:          func() { newTablePresent = true },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Step 3: Apply with same config — should re-enable table_3 because
				// under ALLOW_ALL only table_2 should be disabled.
				Config: `
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	disabled_tables = ["table_2"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "disabled_tables.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "disabled_tables.0", "table_2"),
					func(s *terraform.State) error {
						// table_3 should have been re-enabled by the apply
						if !table3Enabled {
							t.Fatal("expected table_3 to be re-enabled after apply")
						}
						return nil
					},
				),
			},
		},
	})
}

// TestConnectionSchemaConfigSyncModeTableDropped verifies drift detection when a
// table tracked in sync_mode disappears from the API response. The user configures
// sync_mode for table_1 and table_2. After creation, table_2 is dropped from the
// source. On refresh, sync_mode should lose the table_2 entry, causing drift.
func TestConnectionSchemaConfigSyncModeTableDropped(t *testing.T) {
	tableDropped := false
	t1SyncMode := "LIVE"
	t2SyncMode := "LIVE"

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				tables := map[string]any{
					"table_1": map[string]any{"enabled": true, "sync_mode": t1SyncMode},
				}
				if !tableDropped {
					tables["table_2"] = map[string]any{"enabled": true, "sync_mode": t2SyncMode}
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables":  tables,
						},
					},
				}), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				if tables, ok := body["tables"].(map[string]interface{}); ok {
					if t1, ok := tables["table_1"].(map[string]interface{}); ok {
						if sm, ok := t1["sync_mode"].(string); ok {
							t1SyncMode = sm
						}
					}
					if t2, ok := tables["table_2"].(map[string]interface{}); ok {
						if sm, ok := t2["sync_mode"].(string); ok {
							t2SyncMode = sm
						}
					}
				}
				tables := map[string]any{
					"table_1": map[string]any{"enabled": true, "sync_mode": t1SyncMode},
				}
				if !tableDropped {
					tables["table_2"] = map[string]any{"enabled": true, "sync_mode": t2SyncMode}
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables":  tables,
						},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	sync_mode = {
		"table_1" = "HISTORY"
		"table_2" = "SOFT_DELETE"
	}
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "sync_mode.%", "2"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "sync_mode.table_1", "HISTORY"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.test", "sync_mode.table_2", "SOFT_DELETE"),
				),
			},
			{
				// table_2 dropped from source — sync_mode loses table_2, drift detected
				PreConfig:          func() { tableDropped = true },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestConnectionSchemaConfigEnabledPatchSettingsBlocked verifies that when the user
// tries to disable a table whose enabled_patch_settings.allowed is false (e.g. a
// system table), the resource fails with an informative error listing all blocked tables.
func TestConnectionSchemaConfigEnabledPatchSettingsBlocked(t *testing.T) {
	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables": map[string]any{
								"table_1": map[string]any{
									"enabled":   true,
									"sync_mode": "LIVE",
								},
								"system_table": map[string]any{
									"enabled":   true,
									"sync_mode": "LIVE",
									"enabled_patch_settings": map[string]any{
										"allowed":     false,
										"reason_code": "SYSTEM_TABLE",
										"reason":      "This is a system table and cannot be disabled",
									},
								},
								"another_system_table": map[string]any{
									"enabled":   true,
									"sync_mode": "LIVE",
									"enabled_patch_settings": map[string]any{
										"allowed":     false,
										"reason_code": "SYSTEM_TABLE",
										"reason":      "Required for replication",
									},
								},
							},
						},
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
resource "fivetran_connection_schema_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	disabled_tables = ["system_table", "another_system_table"]
}`,
				ExpectError: regexp.MustCompile("(?s)cannot be disabled.*another_system_table.*system_table"),
			},
		},
	})
}

// TestConnectionSchemaConfigConcurrentSchemasNoDeadlock verifies that multiple
// connection_schema_config resources for the same connection_id but different
// schemas are applied without deadlock and with serialized PATCH requests.
// Terraform applies independent resources in parallel; the per-connection mutex
// should serialize the PATCH calls so they never overlap.
func TestConnectionSchemaConfigConcurrentSchemasNoDeadlock(t *testing.T) {
	var concurrentPatches atomic.Int32
	var maxConcurrent atomic.Int32
	var totalPatches atomic.Int32

	schemaNames := []string{"schema_1", "schema_2", "schema_3"}

	// Track per-schema table_b enabled state; starts enabled, PATCH disables it
	tableBState := map[string]bool{}
	for _, name := range schemaNames {
		tableBState[name] = true
	}

	buildSchemas := func() map[string]any {
		schemas := map[string]any{}
		for _, name := range schemaNames {
			schemas[name] = map[string]any{
				"enabled": true,
				"tables": map[string]any{
					"table_a": map[string]any{"enabled": true, "sync_mode": "LIVE"},
					"table_b": map[string]any{"enabled": tableBState[name], "sync_mode": "LIVE"},
				},
			}
		}
		return schemas
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas":               buildSchemas(),
				}), nil
			},
		)

		for _, name := range schemaNames {
			schemaName := name // capture loop var
			mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/"+schemaName).ThenCall(
				func(req *http.Request) (*http.Response, error) {
					current := concurrentPatches.Add(1)
					// Track max concurrent PATCH calls
					for {
						old := maxConcurrent.Load()
						if current <= old || maxConcurrent.CompareAndSwap(old, current) {
							break
						}
					}
					totalPatches.Add(1)

					// Apply the change
					tableBState[schemaName] = false

					// Sleep to widen the concurrency window — if mutex is broken,
					// overlapping calls become much more likely
					time.Sleep(50 * time.Millisecond)

					concurrentPatches.Add(-1)

					return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
						"schema_change_handling": "ALLOW_ALL",
						"schemas":               buildSchemas(),
					}), nil
				},
			)
		}
	}

	config := ""
	for _, name := range schemaNames {
		config += fmt.Sprintf(`
resource "fivetran_connection_schema_config" "%s" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "%s"
	disabled_tables = ["table_b"]
}
`, name, name)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: func(s *terraform.State) error {
					if maxConcurrent.Load() > 1 {
						t.Errorf("detected %d concurrent PATCH calls, expected at most 1 (mutex should serialize)",
							maxConcurrent.Load())
					}
					if totalPatches.Load() != 3 {
						t.Errorf("expected 3 PATCH calls, got %d", totalPatches.Load())
					}
					return nil
				},
			},
		},
	})
}
