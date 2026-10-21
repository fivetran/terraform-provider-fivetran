package resources_test

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// facebook_ads holds accounts and the custom_tables fields, breakdowns, action_breakdowns and
// attribution_windows in Java HashSets, so the API returns them in an arbitrary order that changes
// between reads of an unchanged connection. The mock below returns a different order on every
// response, and the plan after every step must still be empty.

const facebookAdsMetadataResponse = `{
	"id": "facebook_ads",
	"name": "Facebook Ads",
	"config": {
		"type": "object",
		"properties": {
			"schema": {"type": "string", "immutable": true},
			"accounts": {"type": "array", "items": {"type": "string"}},
			"sync_mode": {"type": "string"},
			"custom_tables": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"table_name": {"type": "string"},
						"aggregation": {"type": "string"},
						"fields": {"type": "array", "items": {"type": "string"}},
						"breakdowns": {"type": "array", "items": {"type": "string"}},
						"action_breakdowns": {"type": "array", "items": {"type": "string"}},
						"attribution_windows": {"type": "array", "items": {"type": "string"}}
					}
				}
			}
		}
	},
	"auth": {"type": "object", "properties": {}}
}`

func facebookAdsConnectionTfConfig(accounts, t1Windows, t1Fields string) string {
	return `
		resource "fivetran_connection_v2" "test" {
			provider = fivetran-provider

			group_id                 = "group_id"
			service                  = "facebook_ads"
			destination_schema_names = "FIVETRAN_NAMING"

			config = {
				schema    = "facebook_ads"
				sync_mode = "SpecificAccounts"
				accounts  = ` + accounts + `
				custom_tables = [
					{
						table_name          = "t1"
						aggregation         = "Day"
						fields              = ` + t1Fields + `
						breakdowns          = ["age", "gender", "country"]
						action_breakdowns   = ["action_type", "action_device"]
						attribution_windows = ` + t1Windows + `
					},
					{
						table_name          = "t2"
						aggregation         = "Week"
						fields              = ["reach", "frequency"]
						breakdowns          = []
						action_breakdowns   = ["action_type"]
						attribution_windows = ["CLICK_1D", "VIEW_1D"]
					}
				]
			}
		}`
}

var facebookAdsSetTypedPaths = map[string]bool{
	"accounts":            true,
	"fields":              true,
	"breakdowns":          true,
	"action_breakdowns":   true,
	"attribution_windows": true,
}

// shuffleSetFields returns a deep copy of config with every set-typed list rotated by shift and,
// for odd shifts, reversed, like the API returning a HashSet in another iteration order.
func shuffleSetFields(value interface{}, key string, shift int) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for k, item := range v {
			result[k] = shuffleSetFields(item, k, shift)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = shuffleSetFields(item, key, shift)
		}
		if !facebookAdsSetTypedPaths[key] || len(result) < 2 {
			return result
		}
		rotated := make([]interface{}, len(result))
		for i := range result {
			rotated[i] = result[(i+shift)%len(result)]
		}
		if shift%2 == 1 {
			for i, j := 0, len(rotated)-1; i < j; i, j = i+1, j-1 {
				rotated[i], rotated[j] = rotated[j], rotated[i]
			}
		}
		return rotated
	}
	return value
}

func TestConnectionV2SetTypedFieldsOrderMock(t *testing.T) {
	var storedConfig map[string]interface{}
	responses := 0
	var patchBodies []map[string]interface{}
	var postHandler, patchHandler, deleteHandler *mock.Handler

	connectionResponse := func() map[string]interface{} {
		responses++
		return map[string]interface{}{
			"id":                       "connection_id",
			"group_id":                 "group_id",
			"service":                  "facebook_ads",
			"service_version":          1,
			"schema":                   "facebook_ads",
			"paused":                   true,
			"destination_schema_names": "FIVETRAN_NAMING",
			"status": map[string]interface{}{
				"setup_state":        "connected",
				"sync_state":         "paused",
				"update_state":       "on_schedule",
				"is_historical_sync": false,
				"tasks":              []interface{}{},
				"warnings":           []interface{}{},
			},
			"setup_tests": []interface{}{},
			"config":      shuffleSetFields(storedConfig, "", responses),
		}
	}

	accounts := `["act_1", "act_2", "act_3", "act_4"]`
	windows := `["CLICK_1D", "CLICK_7D", "CLICK_28D", "VIEW_1D"]`
	fields := `["spend", "impressions", "clicks"]`

	checkConfigOrder := func(accounts []string, windows []string) resource.TestCheckFunc {
		checks := []resource.TestCheckFunc{}
		for i, a := range accounts {
			checks = append(checks, resource.TestCheckResourceAttr("fivetran_connection_v2.test", "config.accounts."+strconv.Itoa(i), a))
		}
		for i, w := range windows {
			checks = append(checks, resource.TestCheckResourceAttr("fivetran_connection_v2.test", "config.custom_tables.0.attribution_windows."+strconv.Itoa(i), w))
		}
		return resource.ComposeAggregateTestCheckFunc(checks...)
	}

	// Create: the create response and every later read return the sets in a new order.
	// State keeps the configured order, so the plan after apply (and after refresh) is empty.
	create := resource.TestStep{
		Config: facebookAdsConnectionTfConfig(accounts, windows, fields),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, postHandler.Interactions, 1)
				return nil
			},
			checkConfigOrder([]string{"act_1", "act_2", "act_3", "act_4"}, []string{"CLICK_1D", "CLICK_7D", "CLICK_28D", "VIEW_1D"}),
			resource.TestCheckResourceAttr("fivetran_connection_v2.test", "config.custom_tables.0.fields.0", "spend"),
			resource.TestCheckResourceAttr("fivetran_connection_v2.test", "config.custom_tables.1.action_breakdowns.0", "action_type"),
		),
	}

	// A plan of the unchanged config, the refresh reading yet another order.
	planOnly := func(config string) resource.TestStep {
		return resource.TestStep{Config: config, PlanOnly: true}
	}
	addedConfig := facebookAdsConnectionTfConfig(accounts, `["CLICK_1D", "CLICK_7D", "CLICK_28D"]`, `["spend", "impressions", "clicks", "reach"]`)
	reorderedConfig := facebookAdsConnectionTfConfig(`["act_4", "act_2", "act_1", "act_3"]`, `["CLICK_28D", "CLICK_1D", "CLICK_7D"]`, `["reach", "clicks", "impressions", "spend"]`)

	// A change in two sets of t1: only custom_tables is sent, in the configured order. accounts is
	// returned reordered on every read but unchanged, so it is not sent.
	addWindow := resource.TestStep{
		Config: addedConfig,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, patchHandler.Interactions, 1)
				config := patchBodies[0]["config"].(map[string]interface{})
				tfmock.AssertKeyDoesNotExist(t, config, "accounts")
				tfmock.AssertKeyDoesNotExist(t, config, "schema")
				tables := config["custom_tables"].([]interface{})
				tfmock.AssertEqual(t, len(tables), 2)
				t1 := tables[0].(map[string]interface{})
				tfmock.AssertEqual(t, t1["attribution_windows"], []interface{}{"CLICK_1D", "CLICK_7D", "CLICK_28D"})
				tfmock.AssertEqual(t, t1["fields"], []interface{}{"spend", "impressions", "clicks", "reach"})
				return nil
			},
			checkConfigOrder([]string{"act_1", "act_2", "act_3", "act_4"}, []string{"CLICK_1D", "CLICK_7D", "CLICK_28D"}),
		),
	}

	// Only the configured order changes. This is also the first apply after upgrading the provider,
	// when state holds an order the API returned earlier. Nothing is sent (the API rejects a PATCH
	// without update parameters) and state takes the new configured order.
	reorderOnly := resource.TestStep{
		Config: reorderedConfig,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, patchHandler.Interactions, 1)
				return nil
			},
			checkConfigOrder([]string{"act_4", "act_2", "act_1", "act_3"}, []string{"CLICK_28D", "CLICK_1D", "CLICK_7D"}),
			resource.TestCheckResourceAttr("fivetran_connection_v2.test", "config.custom_tables.0.fields.0", "reach"),
		),
	}

	// An account added outside Terraform is real drift and must show up in the plan.
	upstreamDrift := resource.TestStep{
		PreConfig: func() {
			storedConfig["accounts"] = append(storedConfig["accounts"].([]interface{}), "act_5")
		},
		Config:             reorderedConfig,
		PlanOnly:           true,
		ExpectNonEmptyPlan: true,
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()
				storedConfig = nil
				responses = 0
				patchBodies = nil

				tfmock.MockClient().When(http.MethodGet, "/v1/metadata/connector-types/facebook_ads").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success",
							tfmock.CreateMapFromJsonString(t, facebookAdsMetadataResponse)), nil
					},
				)

				tfmock.MockClient().When(http.MethodGet, "/v1/connections/connection_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectionResponse()), nil
					},
				)

				postHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						storedConfig = body["config"].(map[string]interface{})
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectionResponse()), nil
					},
				)

				patchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/connections/connection_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						patchBodies = append(patchBodies, body)
						if config, ok := body["config"].(map[string]interface{}); ok {
							for k, v := range config {
								storedConfig[k] = v
							}
						}
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectionResponse()), nil
					},
				)

				deleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connection_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, deleteHandler.Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{
				create,
				planOnly(create.Config),
				planOnly(create.Config),
				addWindow,
				planOnly(addedConfig),
				reorderOnly,
				planOnly(reorderedConfig),
				upstreamDrift,
			},
		},
	)
}
