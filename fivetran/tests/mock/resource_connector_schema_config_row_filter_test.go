package mock

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const rowFilterJSON = `{"name":"record_type","description":"Account selection","column_clauses":[{"column":"RecordTypeId","column_type":"STRING","operator":"IN","values":["test-record-type"]}]}`

func rowFilterConfig(filter string, enabled bool) string {
	field := ""
	if filter != "" {
		field = `,"row_filter":` + filter
	}
	return fmt.Sprintf(`resource "fivetran_connector_schema_config" "test" {
 provider = fivetran-provider
 connector_id = "connector_id"
 schema_change_handling = "ALLOW_ALL"
 validation_level = "NONE"
 schemas_json = <<JSON
 {"salesforce":{"enabled":true,"tables":{"Account":{"enabled":%t%s}}}}
 JSON
 }`, enabled, field)
}

func TestConnectorSchemaRowFilter(t *testing.T) {
	original := createMapFromJsonString(t, rowFilterJSON)
	updatedJSON := strings.ReplaceAll(rowFilterJSON, "test-record-type", "another-record-type")
	updated := createMapFromJsonString(t, updatedJSON)
	table := map[string]interface{}{"enabled": true}
	data := map[string]interface{}{"schema_change_handling": "ALLOW_ALL", "schemas": map[string]interface{}{"salesforce": map[string]interface{}{"enabled": true, "tables": map[string]interface{}{"Account": table}}}}
	var patches []map[string]interface{}
	check := func(want interface{}, managed bool) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			raw := s.RootModule().Resources["fivetran_connector_schema_config.test"].Primary.Attributes["schemas_json"]
			schemas := createMapFromJsonString(t, raw)
			stateTable := schemas["salesforce"].(map[string]interface{})["tables"].(map[string]interface{})["Account"].(map[string]interface{})
			got, present := stateTable["row_filter"]
			if present != managed || !reflect.DeepEqual(got, want) {
				return fmt.Errorf("state filter = %#v (present %t), want %#v (present %t)", got, present, want, managed)
			}
			return nil
		}
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			mockClient.Reset()
			mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
			})
			mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				got := body["schemas"].(map[string]interface{})["salesforce"].(map[string]interface{})["tables"].(map[string]interface{})["Account"].(map[string]interface{})
				patches = append(patches, got)
				for key, value := range got {
					table[key] = value
				}
				if table["row_filter"] == nil {
					delete(table, "row_filter")
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
			})
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: `resource "fivetran_connector_schema_config" "test" {
              provider = fivetran-provider
              connector_id = "connector_id"
              schema_change_handling = "ALLOW_ALL"
              validation_level = "NONE"
              schemas = { salesforce = { enabled = true, tables = { Account = { enabled = true } } } }
            }`},
			{Config: rowFilterConfig(rowFilterJSON, true), Check: check(original, true)},
			{Config: rowFilterConfig(rowFilterJSON, true), PlanOnly: true},
			{Config: rowFilterConfig(updatedJSON, true), Check: check(updated, true)},
			// Remote removal must produce a plan, then be repaired.
			{PreConfig: func() { delete(table, "row_filter") }, Config: rowFilterConfig(updatedJSON, true), PlanOnly: true, ExpectNonEmptyPlan: true},
			{Config: rowFilterConfig(updatedJSON, true), Check: check(updated, true)},
			// Omission relinquishes management, even while another table field changes.
			{Config: rowFilterConfig("", false), Check: check(nil, false)},
			{Config: rowFilterConfig("", false), PlanOnly: true},
			{Config: rowFilterConfig("null", false), Check: check(nil, true)},
			{Config: rowFilterConfig("null", false), PlanOnly: true},
		},
	})
	if len(patches) != 5 {
		t.Fatalf("got %d PATCH requests, want 5: %#v", len(patches), patches)
	}
	for i, want := range []interface{}{original, updated, updated} {
		if !reflect.DeepEqual(patches[i]["row_filter"], want) {
			t.Errorf("PATCH %d row_filter = %#v, want %#v", i, patches[i]["row_filter"], want)
		}
	}
	if _, ok := patches[3]["row_filter"]; ok {
		t.Error("omission must not write row_filter")
	}
	if value, ok := patches[4]["row_filter"]; !ok || value != nil {
		t.Error("deletion must send explicit null")
	}
}

func TestConnectorSchemaRowFilterAPIError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			mockClient.Reset()
			data := createMapFromJsonString(t, `{"schema_change_handling":"ALLOW_ALL","schemas":{"salesforce":{"enabled":true,"tables":{"Account":{"enabled":true}}}}}`)
			mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
			})
			mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(func(req *http.Request) (*http.Response, error) {
				return fivetranResponse(t, req, "InvalidInput", http.StatusBadRequest, "Row filtering is not supported", nil), nil
			})
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps:                    []resource.TestStep{{Config: rowFilterConfig(rowFilterJSON, true), ExpectError: regexp.MustCompile("Row filtering is not supported")}},
	})
}

func TestConnectorSchemaRowFilterMissingReadback(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			mockClient.Reset()
			data := createMapFromJsonString(t, `{"schema_change_handling":"ALLOW_ALL","schemas":{"salesforce":{"enabled":true,"tables":{"Account":{"enabled":true}}}}}`)
			for _, method := range []string{http.MethodGet, http.MethodPatch} {
				mockClient.When(method, "/v1/connections/connector_id/schemas").ThenCall(func(req *http.Request) (*http.Response, error) {
					return fivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
				})
			}
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps:                    []resource.TestStep{{Config: rowFilterConfig(rowFilterJSON, true), ExpectError: regexp.MustCompile("Provider produced inconsistent result")}},
	})
}

func TestConnectorSchemaRowFilterInvalidOperator(t *testing.T) {
	invalid := strings.Replace(rowFilterJSON, `"name":`, `"operator":"AND","name":`, 1)
	patches := 0
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			mockClient.Reset()
			data := createMapFromJsonString(t, fmt.Sprintf(`{"schema_change_handling":"ALLOW_ALL","schemas":{"salesforce":{"enabled":true,"tables":{"Account":{"enabled":true,"row_filter":%s}}}}}`, rowFilterJSON))
			mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
			})
			mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(func(req *http.Request) (*http.Response, error) {
				patches++
				return fivetranResponse(t, req, "InvalidInput", http.StatusBadRequest, "operator requires at least two clauses", nil), nil
			})
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: rowFilterConfig(rowFilterJSON, true)},
			{Config: rowFilterConfig(invalid, true), ExpectError: regexp.MustCompile("operator requires at least two clauses")},
		},
	})
	if patches != 1 {
		t.Fatalf("got %d PATCH requests, want 1", patches)
	}
}
