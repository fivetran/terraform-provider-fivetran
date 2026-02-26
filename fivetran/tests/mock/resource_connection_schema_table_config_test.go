package mock

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// helpers to build column mock data
func col(enabled bool, hashed bool, isPK bool) map[string]any {
	return map[string]any{"enabled": enabled, "hashed": hashed, "is_primary_key": isPK}
}

func colWithPatchSettings(enabled bool, hashed bool, isPK bool, allowed bool, reason string) map[string]any {
	return map[string]any{
		"enabled": enabled, "hashed": hashed, "is_primary_key": isPK,
		"enabled_patch_settings": map[string]any{
			"allowed":     allowed,
			"reason_code": "SYSTEM_COLUMN",
			"reason":      reason,
		},
	}
}

func schemasWithColumns(policy string, columns map[string]any) map[string]any {
	return map[string]any{
		"schema_change_handling": policy,
		"schemas": map[string]any{
			"schema_1": map[string]any{
				"enabled": true,
				"tables": map[string]any{
					"table_1": map[string]any{
						"enabled": true,
						"columns": columns,
					},
				},
			},
		},
	}
}

// mockColumnList registers the column list API endpoint returning columns from columnsFn.
func mockColumnList(t *testing.T, columnsFn func() map[string]any) {
	mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas/schema_1/tables/table_1/columns").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
				map[string]any{"columns": columnsFn()}), nil
		},
	)
}

// TestConnectionSchemaTableConfigAllowAllDisabledColumns verifies creating a resource
// with ALLOW_ALL policy and disabled_columns.
func TestConnectionSchemaTableConfigAllowAllDisabledColumns(t *testing.T) {
	colAEnabled := true

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		columns := func() map[string]any {
			return map[string]any{
				"col_a": col(colAEnabled, false, false),
				"col_b": col(true, false, false),
			}
		}

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1/tables/table_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				colAEnabled = false
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
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
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	disabled_columns = ["col_a"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "disabled_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "id", "conn_id:schema_1:table_1"),
				),
			},
		},
	})
}

// TestConnectionSchemaTableConfigBlockAllEnabledColumns verifies creating a resource
// with BLOCK_ALL policy and enabled_columns.
func TestConnectionSchemaTableConfigBlockAllEnabledColumns(t *testing.T) {
	columns := func() map[string]any {
		return map[string]any{
			"col_a": col(false, false, false),
			"col_b": col(true, false, false),
			"col_c": col(false, false, false),
		}
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("BLOCK_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1/tables/table_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("BLOCK_ALL", columns())), nil
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
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	enabled_columns = ["col_b"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "enabled_columns.#", "1"),
				),
			},
		},
	})
}

// TestConnectionSchemaTableConfigHashedColumns verifies that hashed_columns
// are correctly applied and reflected in state.
func TestConnectionSchemaTableConfigHashedColumns(t *testing.T) {
	colBHashed := false

	columns := func() map[string]any {
		return map[string]any{
			"col_a": col(true, false, false),
			"col_b": col(true, colBHashed, false),
		}
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1/tables/table_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				colBHashed = true
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
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
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	hashed_columns = ["col_b"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "hashed_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "hashed_columns.0", "col_b"),
				),
			},
		},
	})
}

// TestConnectionSchemaTableConfigPrimaryKeyColumns verifies that primary_key_columns
// are correctly applied and reflected in state.
func TestConnectionSchemaTableConfigPrimaryKeyColumns(t *testing.T) {
	colAPK := false

	columns := func() map[string]any {
		return map[string]any{
			"col_a": col(true, false, colAPK),
			"col_b": col(true, false, false),
		}
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1/tables/table_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				colAPK = true
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
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
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	primary_key_columns = ["col_a"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "primary_key_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "primary_key_columns.0", "col_a"),
				),
			},
		},
	})
}

// TestConnectionSchemaTableConfigUpdate verifies updating columns across steps.
func TestConnectionSchemaTableConfigUpdate(t *testing.T) {
	colState := map[string]map[string]any{
		"col_a": {"enabled": true, "hashed": false, "is_primary_key": false},
		"col_b": {"enabled": true, "hashed": false, "is_primary_key": false},
		"col_c": {"enabled": true, "hashed": false, "is_primary_key": false},
	}

	columns := func() map[string]any {
		result := map[string]any{}
		for k, v := range colState {
			result[k] = col(v["enabled"].(bool), v["hashed"].(bool), v["is_primary_key"].(bool))
		}
		return result
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1/tables/table_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				if cols, ok := body["columns"].(map[string]interface{}); ok {
					for name, val := range cols {
						if cMap, ok := val.(map[string]interface{}); ok {
							if en, ok := cMap["enabled"]; ok {
								colState[name]["enabled"] = en
							}
							if h, ok := cMap["hashed"]; ok {
								colState[name]["hashed"] = h
							}
						}
					}
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
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
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	disabled_columns = ["col_b"]
	hashed_columns   = ["col_a"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "disabled_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "hashed_columns.#", "1"),
				),
			},
			{
				Config: `
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	disabled_columns = ["col_c"]
	hashed_columns   = ["col_b"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "disabled_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "disabled_columns.0", "col_c"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "hashed_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "hashed_columns.0", "col_b"),
				),
			},
		},
	})
}

// TestConnectionSchemaTableConfigImport verifies importing with "conn_id:schema:table".
func TestConnectionSchemaTableConfigImport(t *testing.T) {
	columns := func() map[string]any {
		return map[string]any{
			"col_a": col(true, false, false),
			"col_b": col(false, false, false),
		}
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1/tables/table_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
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
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	disabled_columns = ["col_b"]
}`,
			},
			{
				ResourceName:            "fivetran_connection_schema_table_config.test",
				ImportState:             true,
				ImportStateId:           "conn_id:schema_1:table_1",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"hashed_columns", "primary_key_columns"},
			},
		},
	})
}

// TestConnectionSchemaTableConfigSchemaNotLoaded verifies error when schema not loaded.
func TestConnectionSchemaTableConfigSchemaNotLoaded(t *testing.T) {
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
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	disabled_columns = ["col_a"]
}`,
				ExpectError: regexp.MustCompile("schema details not available"),
			},
		},
	})
}

// TestConnectionSchemaTableConfigTableNotFound verifies error when table is missing
// from the column list API response.
func TestConnectionSchemaTableConfigTableNotFound(t *testing.T) {
	setupMock := func(t *testing.T) {
		mockClient.Reset()
		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
					"schema_change_handling": "ALLOW_ALL",
					"schemas": map[string]any{
						"schema_1": map[string]any{
							"enabled": true,
							"tables":  map[string]any{},
						},
					},
				}), nil
			},
		)

		// Column list API returns 404 for nonexistent table
		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas/schema_1/tables/table_1/columns").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranResponse(t, req, "NotFound", http.StatusNotFound,
					"Table not found", nil), nil
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
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	disabled_columns = ["col_a"]
}`,
				ExpectError: regexp.MustCompile("unable to list columns"),
			},
		},
	})
}

// TestConnectionSchemaTableConfigConnectionDeletedUpstream verifies RemoveResource
// when connection is deleted upstream.
func TestConnectionSchemaTableConfigConnectionDeletedUpstream(t *testing.T) {
	getCallCount := 0

	columns := func() map[string]any {
		return map[string]any{"col_a": col(true, false, false)}
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()
		getCallCount = 0

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				getCallCount++
				if getCallCount <= 3 {
					return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
						schemasWithColumns("ALLOW_ALL", columns())), nil
				}
				return fivetranResponse(t, req, "NotFound_Connection", http.StatusNotFound,
					"Connection not found", nil), nil
			},
		)
		mockColumnList(t, columns)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1/tables/table_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
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
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
}`,
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestConnectionSchemaTableConfigEnabledPatchSettingsBlocked verifies that system
// columns that cannot be disabled produce an informative error.
func TestConnectionSchemaTableConfigEnabledPatchSettingsBlocked(t *testing.T) {
	columns := func() map[string]any {
		return map[string]any{
			"col_a":      col(true, false, false),
			"sys_col_id": colWithPatchSettings(true, false, true, false, "Primary key column cannot be disabled"),
		}
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	disabled_columns = ["sys_col_id"]
}`,
				ExpectError: regexp.MustCompile("(?s)cannot be.*disabled.*sys_col_id"),
			},
		},
	})
}

// TestConnectionSchemaTableConfigColumnDroppedFromSource verifies drift detection
// when a configured column disappears from the API response.
func TestConnectionSchemaTableConfigColumnDroppedFromSource(t *testing.T) {
	colDropped := false

	columns := func() map[string]any {
		result := map[string]any{
			"col_a": col(true, true, false),
		}
		if !colDropped {
			result["col_b"] = col(false, false, false)
		}
		return result
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1/tables/table_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
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
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	disabled_columns = ["col_b"]
	hashed_columns   = ["col_a"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "disabled_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "hashed_columns.#", "1"),
				),
			},
			{
				PreConfig:          func() { colDropped = true },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestConnectionSchemaTableConfigHashedColumnDriftDetection verifies that when
// hashed_columns IS configured and an external process hashes another column,
// drift is detected. When hashed_columns is NOT configured, external changes
// to column hashing are ignored.
func TestConnectionSchemaTableConfigHashedColumnDriftDetection(t *testing.T) {
	colCHashed := false

	columns := func() map[string]any {
		return map[string]any{
			"col_a": col(true, true, false),
			"col_b": col(true, false, false),
			"col_c": col(true, colCHashed, false),
		}
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1/tables/table_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				// hashed_columns IS configured — col_a is tracked
				Config: `
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	hashed_columns = ["col_a"]
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "hashed_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.test", "hashed_columns.0", "col_a"),
				),
			},
			{
				// col_c gets hashed externally — drift MUST be detected because
				// hashed_columns is configured (resource owns the setting)
				PreConfig:          func() { colCHashed = true },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestConnectionSchemaTableConfigUnconfiguredHashedIgnored verifies that when
// hashed_columns is NOT in the config, external hashing changes are ignored
// and no drift is reported.
func TestConnectionSchemaTableConfigUnconfiguredHashedIgnored(t *testing.T) {
	colBHashed := false

	columns := func() map[string]any {
		return map[string]any{
			"col_a": col(true, false, false),
			"col_b": col(true, colBHashed, false),
		}
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)

		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas/schema_1/tables/table_1").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				// No hashed_columns in config — only managing disabled_columns
				Config: `
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
}`,
			},
			{
				// col_b gets hashed externally — should NOT cause drift
				// because hashed_columns is not configured
				PreConfig: func() { colBHashed = true },
				Config: `
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
}`,
				PlanOnly: true,
			},
		},
	})
}

// TestConnectionSchemaTableConfigHashedColumnMustBeEnabled verifies that configuring
// a column in hashed_columns or primary_key_columns that would be disabled produces
// an error.
func TestConnectionSchemaTableConfigHashedColumnMustBeEnabled(t *testing.T) {
	columns := func() map[string]any {
		return map[string]any{
			"col_a": col(true, false, false),
			"col_b": col(true, false, false),
		}
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				// col_b is in disabled_columns AND hashed_columns — should fail
				Config: `
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	disabled_columns = ["col_b"]
	hashed_columns   = ["col_b"]
}`,
				ExpectError: regexp.MustCompile("(?s)hashed_columns or primary_key_columns.*disabled.*col_b"),
			},
		},
	})
}

// TestConnectionSchemaTableConfigPKColumnMustBeEnabled verifies that configuring
// a column in primary_key_columns that is also in disabled_columns produces an error.
func TestConnectionSchemaTableConfigPKColumnMustBeEnabled(t *testing.T) {
	columns := func() map[string]any {
		return map[string]any{
			"col_a": col(true, false, false),
			"col_b": col(true, false, false),
		}
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
					schemasWithColumns("ALLOW_ALL", columns())), nil
			},
		)
		mockColumnList(t, columns)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				// col_b is in both disabled_columns and primary_key_columns — should fail
				Config: `
resource "fivetran_connection_schema_table_config" "test" {
	provider      = fivetran-provider
	connection_id = "conn_id"
	schema_name   = "schema_1"
	table_name    = "table_1"
	disabled_columns    = ["col_b"]
	primary_key_columns = ["col_b"]
}`,
				ExpectError: regexp.MustCompile("(?s)hashed_columns or primary_key_columns.*disabled.*col_b"),
			},
		},
	})
}
