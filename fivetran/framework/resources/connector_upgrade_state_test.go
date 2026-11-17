package resources_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"testing"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Feeds prior-version state JSON straight to the state upgrader through the provider protocol.
// It checks the type conversions of every version without a registry download, including
// versions that are hard to produce with a real provider (0 and 1).
func TestConnectorUpgradeStateFromRawState(t *testing.T) {
	authJSON := map[string]interface{}{"refresh_token": "refresh_token"}
	setOf := func(m map[string]interface{}) []interface{} { return []interface{}{m} }
	destination := map[string]interface{}{"name": "schema", "table": "table", "prefix": "prefix"}

	// From version 3 on, config, auth and destination_schema are blocks and primitives are typed.
	blocksState := func(extra map[string]interface{}) map[string]interface{} {
		state := map[string]interface{}{
			"id":                 "connector_id",
			"group_id":           "group_id",
			"service":            "google_sheets",
			"run_setup_tests":    false,
			"trust_certificates": true,
			"trust_fingerprints": false,
			"config": map[string]interface{}{
				"sheet_id":         "sheet_id",
				"always_encrypted": true,
				"agent_port":       5432,
				"account_ids":      []interface{}{"a", "b"},
			},
			"auth":               authJSON,
			"destination_schema": destination,
		}
		for k, v := range extra {
			state[k] = v
		}
		return state
	}
	// Before, blocks were sets of one element and booleans and numbers were strings.
	setsState := func(extra map[string]interface{}) map[string]interface{} {
		state := map[string]interface{}{
			"id":                 "connector_id",
			"group_id":           "group_id",
			"service":            "google_sheets",
			"run_setup_tests":    "false",
			"trust_certificates": "true",
			"trust_fingerprints": "false",
			"config": setOf(map[string]interface{}{
				"sheet_id":         "sheet_id",
				"always_encrypted": "true",
				"agent_port":       "5432",
				"account_ids":      []interface{}{"a", "b"},
			}),
			"auth":               setOf(authJSON),
			"destination_schema": setOf(destination),
		}
		for k, v := range extra {
			state[k] = v
		}
		return state
	}

	tests := []struct {
		name    string
		version int64
		state   map[string]interface{}
		agentID string // expected hybrid_deployment_agent_id, "" for null
	}{
		{name: "version_0", version: 0, state: setsState(map[string]interface{}{"sync_frequency": "60", "paused": "true"})},
		{name: "version_1", version: 1, state: setsState(nil)},
		{name: "version_2", version: 2, state: setsState(nil)},
		{
			name: "version_3_local_processing_agent_id", version: 3, agentID: "local",
			state: blocksState(map[string]interface{}{"local_processing_agent_id": "local"}),
		},
		{
			name: "version_3_hybrid_deployment_agent_id", version: 3, agentID: "hybrid",
			state: blocksState(map[string]interface{}{"hybrid_deployment_agent_id": "hybrid"}),
		},
		{
			name: "version_3_both_agent_ids_prefers_hybrid", version: 3, agentID: "hybrid",
			state: blocksState(map[string]interface{}{"local_processing_agent_id": "local", "hybrid_deployment_agent_id": "hybrid"}),
		},
		{
			name: "version_3_removed_auth_field", version: 3,
			state: blocksState(map[string]interface{}{"auth": map[string]interface{}{"refresh_token": "refresh_token", "aws_access_key": "key"}}),
		},
		// Version 4 is the current schema version, so the framework never calls upgrader 4:
		// this case covers the same-version path, which decodes the state as is.
		{
			name: "version_4_same_version", version: 4, agentID: "hybrid",
			state: blocksState(map[string]interface{}{"hybrid_deployment_agent_id": "hybrid"}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			attrs := upgradeConnectorRawState(t, tc.version, tc.state)

			assertStringAttr(t, attrs, "id", "connector_id")
			assertStringAttr(t, attrs, "group_id", "group_id")
			assertStringAttr(t, attrs, "service", "google_sheets")

			// Strings from the old schema become booleans.
			assertBoolAttr(t, attrs, "run_setup_tests", false)
			assertBoolAttr(t, attrs, "trust_certificates", true)
			assertBoolAttr(t, attrs, "trust_fingerprints", false)

			// Single-element sets become blocks, and typed config fields are converted.
			config := objectAttr(t, attrs, "config")
			assertStringAttr(t, config, "sheet_id", "sheet_id")
			assertBoolAttr(t, config, "always_encrypted", true)
			assertNumberAttr(t, config, "agent_port", 5432)
			assertStringSetAttr(t, config, "account_ids", "a", "b")
			assertStringAttr(t, objectAttr(t, attrs, "auth"), "refresh_token", "refresh_token")
			destinationSchema := objectAttr(t, attrs, "destination_schema")
			assertStringAttr(t, destinationSchema, "name", "schema")
			assertStringAttr(t, destinationSchema, "table", "table")
			assertStringAttr(t, destinationSchema, "prefix", "prefix")
			// destination_schema.table_group_name did not exist before.
			assertNullAttr(t, destinationSchema, "table_group_name")

			// Attributes added after the prior version are null.
			for _, added := range []string{"networking_method", "proxy_agent_id", "private_link_id", "data_delay_sensitivity", "data_delay_threshold", "data_checks"} {
				assertNullAttr(t, attrs, added)
			}

			if tc.agentID == "" {
				assertNullAttr(t, attrs, "hybrid_deployment_agent_id")
			} else {
				assertStringAttr(t, attrs, "hybrid_deployment_agent_id", tc.agentID)
			}
		})
	}
}

// An empty set, which is how an absent optional block was stored, must become a null block.
func TestConnectorUpgradeStateEmptySetsBecomeNullBlocks(t *testing.T) {
	blocks := []string{"config", "auth", "destination_schema"}
	for _, version := range []int64{0, 1, 2} {
		t.Run(fmt.Sprintf("version_%d", version), func(t *testing.T) {
			state := map[string]interface{}{"id": "connector_id", "group_id": "group_id", "service": "google_sheets"}
			for _, block := range blocks {
				state[block] = []interface{}{}
			}
			attrs := upgradeConnectorRawState(t, version, state)

			assertStringAttr(t, attrs, "id", "connector_id")
			for _, block := range blocks {
				assertNullAttr(t, attrs, block)
			}
		})
	}
}

func upgradeConnectorRawState(t *testing.T, version int64, state map[string]interface{}) map[string]tftypes.Value {
	t.Helper()
	ctx := context.Background()
	server := providerserver.NewProtocol6(framework.FivetranProvider())()

	schemaResp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	schema, ok := schemaResp.ResourceSchemas["fivetran_connector"]
	if !ok {
		t.Fatal("fivetran_connector schema not found")
	}

	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.UpgradeResourceState(ctx, &tfprotov6.UpgradeResourceStateRequest{
		TypeName: "fivetran_connector",
		Version:  version,
		RawState: &tfprotov6.RawState{JSON: raw},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("state upgrade: %s: %s", d.Summary, d.Detail)
		}
	}
	if resp.UpgradedState == nil {
		t.Fatal("missing upgraded state")
	}
	value, err := resp.UpgradedState.Unmarshal(schema.ValueType())
	if err != nil {
		t.Fatalf("upgraded state does not match the current schema: %v", err)
	}
	var attrs map[string]tftypes.Value
	if err := value.As(&attrs); err != nil {
		t.Fatal(err)
	}
	return attrs
}

func objectAttr(t *testing.T, attrs map[string]tftypes.Value, name string) map[string]tftypes.Value {
	t.Helper()
	var obj map[string]tftypes.Value
	if err := attrs[name].As(&obj); err != nil || attrs[name].IsNull() {
		t.Fatalf("%s must be an object, got %v (%v)", name, attrs[name], err)
	}
	return obj
}

func assertStringAttr(t *testing.T, attrs map[string]tftypes.Value, name, expected string) {
	t.Helper()
	var actual string
	if err := attrs[name].As(&actual); err != nil {
		t.Errorf("%s: expected %q, got %v", name, expected, attrs[name])
		return
	}
	tfmock.AssertEqual(t, actual, expected)
}

func assertNumberAttr(t *testing.T, attrs map[string]tftypes.Value, name string, expected int64) {
	t.Helper()
	actual := new(big.Float)
	if attrs[name].IsNull() || attrs[name].As(&actual) != nil {
		t.Errorf("%s: expected %d, got %v", name, expected, attrs[name])
		return
	}
	if actual.Cmp(big.NewFloat(float64(expected))) != 0 {
		t.Errorf("%s: expected %d, got %v", name, expected, actual)
	}
}

func assertStringSetAttr(t *testing.T, attrs map[string]tftypes.Value, name string, expected ...string) {
	t.Helper()
	var elements []tftypes.Value
	if attrs[name].IsNull() || attrs[name].As(&elements) != nil {
		t.Errorf("%s: expected set %v, got %v", name, expected, attrs[name])
		return
	}
	actual := make([]string, 0, len(elements))
	for _, e := range elements {
		var v string
		if err := e.As(&v); err != nil {
			t.Errorf("%s: element %v is not a string", name, e)
			return
		}
		actual = append(actual, v)
	}
	sort.Strings(actual)
	sort.Strings(expected)
	tfmock.AssertEqual(t, actual, expected)
}

func assertBoolAttr(t *testing.T, attrs map[string]tftypes.Value, name string, expected bool) {
	t.Helper()
	// As on a null leaves false without an error, so null must be rejected explicitly.
	var actual bool
	if attrs[name].IsNull() || attrs[name].As(&actual) != nil {
		t.Errorf("%s: expected %v, got %v", name, expected, attrs[name])
		return
	}
	tfmock.AssertEqual(t, actual, expected)
}

func assertNullAttr(t *testing.T, attrs map[string]tftypes.Value, name string) {
	t.Helper()
	if v, ok := attrs[name]; !ok || !v.IsNull() {
		t.Errorf("%s must be null after upgrade, got %v", name, v)
	}
}
