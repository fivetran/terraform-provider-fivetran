package mock

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/actions"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestSchemaManagementFullFlow simulates the full lifecycle of a Fivetran connection
// with schema management across 3 schemas × 3 tables, with column-level config on
// selected tables.
//
// Note: terraform-plugin-testing does not support for_each string-keyed resources
// in state checks, so we use individually named resources that mirror what a
// for_each config would produce. The HCL uses locals to centralize configuration,
// demonstrating the variable-driven pattern used in production.
//
// Flow:
//  1. Create connection + reload schema via lifecycle action
//  2. Schema-level: ALLOW_ALL, disable "reporting" schema
//  3. Table-level: one resource per schema, each disables one table
//  4. Column-level: one resource per table needing column config
//  5. Unpause via connector_schedule
func TestSchemaManagementFullFlow(t *testing.T) {
	origBackoff := core.SchemaConflictBackoff()
	core.SetSchemaConflictBackoff(10 * time.Millisecond)
	defer core.SetSchemaConflictBackoff(origBackoff)

	actions.PollIntervalOverride = 100 * time.Millisecond
	defer func() { actions.PollIntervalOverride = 0 }()

	// --- Stateful mock data ---
	var mu sync.Mutex
	connectionCreated := false
	schemaReloaded := false
	paused := true
	syncFrequency := 360

	schemaEnabled := map[string]bool{
		"public": true, "analytics": true, "reporting": true,
	}
	tableEnabled := map[string]map[string]bool{
		"public":    {"users": true, "orders": true, "products": true},
		"analytics": {"events": true, "sessions": true, "pageviews": true},
		"reporting": {"daily": true, "weekly": true, "monthly": true},
	}
	columnState := map[string]map[string]map[string]map[string]any{
		"public": {
			"users": {
				"id":    {"enabled": true, "hashed": false, "is_primary_key": false},
				"name":  {"enabled": true, "hashed": false, "is_primary_key": false},
				"email": {"enabled": true, "hashed": false, "is_primary_key": false},
			},
			"orders": {
				"order_id": {"enabled": true, "hashed": false, "is_primary_key": false},
				"total":    {"enabled": true, "hashed": false, "is_primary_key": false},
				"status":   {"enabled": true, "hashed": false, "is_primary_key": false},
			},
		},
		"analytics": {
			"events": {
				"event_id": {"enabled": true, "hashed": false, "is_primary_key": false},
				"payload":  {"enabled": true, "hashed": false, "is_primary_key": false},
				"ts":       {"enabled": true, "hashed": false, "is_primary_key": false},
			},
		},
	}

	buildSchemaData := func() map[string]any {
		mu.Lock()
		defer mu.Unlock()
		schemas := map[string]any{}
		for sName, sEnabled := range schemaEnabled {
			tables := map[string]any{}
			for tName, tEnabled := range tableEnabled[sName] {
				tables[tName] = map[string]any{"enabled": tEnabled, "sync_mode": "LIVE"}
			}
			schemas[sName] = map[string]any{"enabled": sEnabled, "tables": tables}
		}
		return map[string]any{"schema_change_handling": "ALLOW_ALL", "schemas": schemas}
	}

	buildColumnData := func(schema, table string) map[string]any {
		mu.Lock()
		defer mu.Unlock()
		result := map[string]any{}
		if cols, ok := columnState[schema][table]; ok {
			for cName, c := range cols {
				result[cName] = map[string]any{
					"enabled": c["enabled"], "hashed": c["hashed"], "is_primary_key": c["is_primary_key"],
				}
			}
		}
		return result
	}

	connResp := func() map[string]any {
		return map[string]any{
			"id": "conn_id", "group_id": "group_id", "service": "postgres",
			"service_version": float64(1), "schema": "postgres_rds",
			"paused": paused, "pause_after_trial": false,
			"connected_by": "user_id", "created_at": "2024-01-01T00:00:00.000000Z",
			"succeeded_at": nil, "failed_at": nil,
			"sync_frequency": float64(syncFrequency), "schedule_type": "auto",
			"networking_method": "Directly",
			"status": map[string]any{
				"setup_state": "connected", "sync_state": "scheduled",
				"update_state": "on_schedule", "is_historical_sync": false,
				"tasks": []any{}, "warnings": []any{},
			},
			"setup_tests": []any{}, "config": map[string]any{},
		}
	}

	setupMock := func(t *testing.T) {
		mockClient.Reset()

		// Connection CRUD
		mockClient.When(http.MethodPost, "/v1/connections").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				connectionCreated = true
				paused = true
				return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", connResp()), nil
			},
		)
		mockClient.When(http.MethodGet, "/v1/connections/conn_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connResp()), nil
			},
		)
		mockClient.When(http.MethodPatch, "/v1/connections/conn_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				if p, ok := body["paused"]; ok {
					paused = p.(bool)
				}
				if sf, ok := body["sync_frequency"]; ok {
					syncFrequency = int(sf.(float64))
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connResp()), nil
			},
		)
		mockClient.When(http.MethodDelete, "/v1/connections/conn_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				connectionCreated = false
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)

		// Schema reload
		mockClient.When(http.MethodPost, "/v1/connections/conn_id/schemas/reload").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				schemaReloaded = true
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", buildSchemaData()), nil
			},
		)

		// Schema details
		mockClient.When(http.MethodGet, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if !schemaReloaded {
					return fivetranResponse(t, req, "NotFound_SchemaConfig", http.StatusNotFound,
						"Schema config not found", nil), nil
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", buildSchemaData()), nil
			},
		)

		// Schema-level PATCH
		mockClient.When(http.MethodPatch, "/v1/connections/conn_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				if schemas, ok := body["schemas"].(map[string]interface{}); ok {
					mu.Lock()
					for sName, sVal := range schemas {
						if sMap, ok := sVal.(map[string]interface{}); ok {
							if en, ok := sMap["enabled"]; ok {
								schemaEnabled[sName] = en.(bool)
							}
						}
					}
					mu.Unlock()
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", buildSchemaData()), nil
			},
		)

		// Table-level PATCH per schema
		for _, sName := range []string{"public", "analytics", "reporting"} {
			schemaName := sName
			mockClient.When(http.MethodPatch, fmt.Sprintf("/v1/connections/conn_id/schemas/%s", schemaName)).ThenCall(
				func(req *http.Request) (*http.Response, error) {
					body := requestBodyToJson(t, req)
					if tables, ok := body["tables"].(map[string]interface{}); ok {
						mu.Lock()
						for tName, tVal := range tables {
							if tMap, ok := tVal.(map[string]interface{}); ok {
								if en, ok := tMap["enabled"]; ok {
									tableEnabled[schemaName][tName] = en.(bool)
								}
							}
						}
						mu.Unlock()
					}
					return fivetranSuccessResponse(t, req, http.StatusOK, "Success", buildSchemaData()), nil
				},
			)
		}

		// Column list + PATCH for tables with column config
		for _, pair := range []struct{ schema, table string }{
			{"public", "users"}, {"public", "orders"}, {"analytics", "events"},
		} {
			s, tbl := pair.schema, pair.table
			mockClient.When(http.MethodGet,
				fmt.Sprintf("/v1/connections/conn_id/schemas/%s/tables/%s/columns", s, tbl)).ThenCall(
				func(req *http.Request) (*http.Response, error) {
					return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
						map[string]any{"columns": buildColumnData(s, tbl)}), nil
				},
			)
			mockClient.When(http.MethodPatch,
				fmt.Sprintf("/v1/connections/conn_id/schemas/%s/tables/%s", s, tbl)).ThenCall(
				func(req *http.Request) (*http.Response, error) {
					body := requestBodyToJson(t, req)
					if cols, ok := body["columns"].(map[string]interface{}); ok {
						mu.Lock()
						for cName, cVal := range cols {
							if cMap, ok := cVal.(map[string]interface{}); ok {
								if colMap, exists := columnState[s][tbl][cName]; exists {
									if en, ok := cMap["enabled"]; ok {
										colMap["enabled"] = en
									}
									if h, ok := cMap["hashed"]; ok {
										colMap["hashed"] = h
									}
									if pk, ok := cMap["is_primary_key"]; ok {
										colMap["is_primary_key"] = pk
									}
								}
							}
						}
						mu.Unlock()
					}
					return fivetranSuccessResponse(t, req, http.StatusOK, "Success", buildSchemaData()), nil
				},
			)
		}
	}

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { setupMock(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			if connectionCreated {
				t.Error("connection was not deleted")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: `
# ─── Configuration driven by locals ──────────────────────────────────
# In production, these would come from variables / tfvars.
# terraform-plugin-testing doesn't support for_each in state checks,
# so we expand the loop manually here, but the pattern is identical.

locals {
  connection_id = fivetran_connector.pg.id

  disabled_schemas = ["reporting"]

  schema_table_config = {
    public    = { disabled_tables = ["products"] }
    analytics = { disabled_tables = ["pageviews"] }
    reporting = { disabled_tables = ["monthly"] }
  }

  column_config = {
    public_users = {
      schema = "public", table = "users"
      disabled_columns = ["name"], hashed_columns = ["email"], pk_columns = ["id"]
    }
    public_orders = {
      schema = "public", table = "orders"
      disabled_columns = ["status"]
    }
    analytics_events = {
      schema = "analytics", table = "events"
      disabled_columns = ["payload"]
    }
  }
}

# ─── 1. Connection + schema reload ───────────────────────────────────

resource "fivetran_connector" "pg" {
  provider = fivetran-provider
  group_id = "group_id"
  service  = "postgres"
  destination_schema { prefix = "postgres_rds" }
  run_setup_tests    = false
  trust_certificates = false
  trust_fingerprints = false
  config {}
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.fivetran_connection_schema_reload.reload]
    }
  }
}

action "fivetran_connection_schema_reload" "reload" {
  provider = fivetran-provider
  config { connection_id = local.connection_id }
}

# ─── 2. Schema-level settings ────────────────────────────────────────

resource "fivetran_connection_schema_settings" "settings" {
  provider               = fivetran-provider
  connection_id          = local.connection_id
  schema_change_handling = "ALLOW_ALL"
  disabled_schemas       = local.disabled_schemas
  depends_on             = [fivetran_connector.pg]
}

# ─── 3. Table-level settings (one per schema) ────────────────────────

resource "fivetran_connection_schema_config" "public" {
  provider        = fivetran-provider
  connection_id   = local.connection_id
  schema_name     = "public"
  disabled_tables = local.schema_table_config["public"].disabled_tables
  depends_on      = [fivetran_connection_schema_settings.settings]
}

resource "fivetran_connection_schema_config" "analytics" {
  provider        = fivetran-provider
  connection_id   = local.connection_id
  schema_name     = "analytics"
  disabled_tables = local.schema_table_config["analytics"].disabled_tables
  depends_on      = [fivetran_connection_schema_settings.settings]
}

resource "fivetran_connection_schema_config" "reporting" {
  provider        = fivetran-provider
  connection_id   = local.connection_id
  schema_name     = "reporting"
  disabled_tables = local.schema_table_config["reporting"].disabled_tables
  depends_on      = [fivetran_connection_schema_settings.settings]
}

# ─── 4. Column-level settings (one per table needing column config) ──

resource "fivetran_connection_schema_table_config" "public_users" {
  provider            = fivetran-provider
  connection_id       = local.connection_id
  schema_name         = local.column_config["public_users"].schema
  table_name          = local.column_config["public_users"].table
  disabled_columns    = local.column_config["public_users"].disabled_columns
  hashed_columns      = local.column_config["public_users"].hashed_columns
  primary_key_columns = local.column_config["public_users"].pk_columns
  depends_on          = [fivetran_connection_schema_config.public]
}

resource "fivetran_connection_schema_table_config" "public_orders" {
  provider         = fivetran-provider
  connection_id    = local.connection_id
  schema_name      = local.column_config["public_orders"].schema
  table_name       = local.column_config["public_orders"].table
  disabled_columns = local.column_config["public_orders"].disabled_columns
  depends_on       = [fivetran_connection_schema_config.public]
}

resource "fivetran_connection_schema_table_config" "analytics_events" {
  provider         = fivetran-provider
  connection_id    = local.connection_id
  schema_name      = local.column_config["analytics_events"].schema
  table_name       = local.column_config["analytics_events"].table
  disabled_columns = local.column_config["analytics_events"].disabled_columns
  depends_on       = [fivetran_connection_schema_config.analytics]
}

# ─── 5. Unpause ──────────────────────────────────────────────────────

resource "fivetran_connector_schedule" "schedule" {
  provider       = fivetran-provider
  connector_id   = local.connection_id
  sync_frequency = "360"
  paused         = "false"
  schedule_type  = "auto"
  depends_on     = [
    fivetran_connection_schema_table_config.public_users,
    fivetran_connection_schema_table_config.public_orders,
    fivetran_connection_schema_table_config.analytics_events,
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_connector.pg", "service", "postgres"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_settings.settings", "disabled_schemas.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.public", "disabled_tables.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.analytics", "disabled_tables.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_config.reporting", "disabled_tables.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.public_users", "disabled_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.public_users", "hashed_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.public_users", "primary_key_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.public_orders", "disabled_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connection_schema_table_config.analytics_events", "disabled_columns.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.schedule", "paused", "false"),

					func(s *terraform.State) error {
						mu.Lock()
						defer mu.Unlock()

						if !schemaReloaded {
							t.Error("schema was not reloaded")
						}
						if schemaEnabled["reporting"] {
							t.Error("expected reporting schema to be disabled")
						}
						if tableEnabled["public"]["products"] {
							t.Error("expected public.products to be disabled")
						}
						if tableEnabled["analytics"]["pageviews"] {
							t.Error("expected analytics.pageviews to be disabled")
						}
						if tableEnabled["reporting"]["monthly"] {
							t.Error("expected reporting.monthly to be disabled")
						}
						if columnState["public"]["users"]["name"]["enabled"].(bool) {
							t.Error("expected public.users.name to be disabled")
						}
						if !columnState["public"]["users"]["email"]["hashed"].(bool) {
							t.Error("expected public.users.email to be hashed")
						}
						if !columnState["public"]["users"]["id"]["is_primary_key"].(bool) {
							t.Error("expected public.users.id to be primary key")
						}
						if columnState["public"]["orders"]["status"]["enabled"].(bool) {
							t.Error("expected public.orders.status to be disabled")
						}
						if columnState["analytics"]["events"]["payload"]["enabled"].(bool) {
							t.Error("expected analytics.events.payload to be disabled")
						}
						if paused {
							t.Error("expected connection to be unpaused")
						}
						return nil
					},
				),
			},
		},
	})
}
