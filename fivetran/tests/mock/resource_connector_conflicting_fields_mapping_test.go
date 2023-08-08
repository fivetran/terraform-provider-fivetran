package mock

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	connectorConflictingMockGetHandler  *mock.Handler
	connectorConflictingMockPostHandler *mock.Handler
	connectorConflictingMockDelete      *mock.Handler
	connectorConflictingMappingMockData map[string]interface{}
)

const (
	connectorConfigConflictingMappingTfConfig = `
	resource "fivetran_connector" "test_connector" {
		provider = fivetran-provider

		group_id = "group_id"
		service = "%v"

		destination_schema {
			%v
		}

		trust_certificates = false
		trust_fingerprints = false
		run_setup_tests = false

		config {
			%v
		}
	}`

	connectorConflictingMappingResponse = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "%v",
        "service_version": 1,
        "schema": "%v",
        "paused": true,
        "pause_after_trial": true,
        "connected_by": "user_id",
        "created_at": "2022-01-01T11:22:33.012345Z",
        "succeeded_at": null,
        "failed_at": null,
        "sync_frequency": 5,
		"schedule_type": "auto",
        "status": {
            "setup_state": "incomplete",
            "sync_state": "paused",
            "update_state": "on_schedule",
            "is_historical_sync": true,
            "tasks": [{
				"code":"task_code",
				"message":"task_message"
			}],
            "warnings": [{
				"code":"warning_code",
				"message":"warning_message"
			}]
        },
        "setup_tests": [{
            "title": "Validate Login",
            "status": "FAILED",
            "message": "Invalid login credentials"
        }],
        "config": {
			%v
		}
	}
	`
)

func getTfConfigForConflictingFields(service, destinationSchema, configTf string) string {
	return fmt.Sprintf(connectorConfigConflictingMappingTfConfig, service, destinationSchema, configTf)
}

func getJsonConfigForConflictingFields(service, schema, configJson string) string {
	return fmt.Sprintf(connectorConflictingMappingResponse, service, schema, configJson)
}

func setupMockClientConnectorResourceConfigConflictingFieldsMapping(t *testing.T, service, schema, configJson string) {
	mockClient.Reset()

	connectorConflictingMockGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorConflictingMappingMockData), nil
		},
	)

	connectorConflictingMockPostHandler = mockClient.When(http.MethodPost, "/v1/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			//body := requestBodyToJson(t, req)

			//assertKeyExists(t, body, "config")

			//config := body["config"].(map[string]interface{})

			connectorConflictingMappingMockData = createMapFromJsonString(t, getJsonConfigForConflictingFields(service, schema, configJson))
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorConflictingMappingMockData), nil
		},
	)

	connectorConflictingMockDelete = mockClient.When(http.MethodDelete, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorConflictingMappingMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorConflictingMappingMockData), nil
		},
	)
}

//func testConflictingField(tfConfig, jsonResponse string, checkFunc resource.TestCheckFunc, )

func TestResourceConnectorConfigConflictingFieldsMappingMock(t *testing.T) {
	testResourceConnectorConfigConflictingFieldsMappingMock(t,
		"pendo",
		`
		name = "pendo"
		`,
		"pendo",
		`
	 	app_ids = ["app_id"]
	 	`, `
		 "app_ids":["app_id"]
		`,
	)

	testResourceConnectorConfigConflictingFieldsMappingMock(t,
		"appsflyer",
		`
		name = "appsflyer"
		`,
		"appsflyer",
		`
	 	app_ids_appsflyer {
	 		app_id = "app_id"
	 	}
	 	`, `
		 "app_ids":[{"app_id":"app_id"}]
		`,
	)

	testResourceConnectorConfigConflictingFieldsMappingMock(t,
		"linkedin_ads",
		`
		name = "linkedin"
		`,
		"linkedin",
		`
		reports_linkedin_ads = ["report"]
	 	`, `
		 "reports":["report"]
		`,
	)

	testResourceConnectorConfigConflictingFieldsMappingMock(t,
		"google_analytics",
		`
		name = "google_analytics"
		`,
		"google_analytics",
		`
		reports {
			aggregation = "aggregation"
			attributes = ["attribute"]
			config_type = "config_type"
			dimensions = ["dimension"]
			fields = ["field"]
			filter = "filter"
			filter_field_name = "filter_field_name"
			filter_value = "filter_value"
			metrics = ["metric"]
			prebuilt_report = "prebuilt_report"
			report_type = "report_type"
			search_types = ["search_type"]
			segment_ids = ["segment_id"]
			segments = ["segment"]
			table = "table1"
		}
		reports {
			aggregation = "aggregation"
			attributes = ["attribute"]
			config_type = "config_type"
			dimensions = ["dimension"]
			fields = ["field"]
			filter = "filter"
			filter_field_name = "filter_field_name"
			filter_value = "filter_value"
			metrics = ["metric"]
			prebuilt_report = "prebuilt_report"
			report_type = "report_type"
			search_types = ["search_type"]
			segment_ids = ["segment_id"]
			segments = ["segment"]
			table = "table2"
		}
	 	`, `
		 "reports": [{
			"aggregation" : "aggregation",
			"attributes" : ["attribute"],
			"config_type" : "config_type",
			"dimensions" : ["dimension"],
			"fields" : ["field"],
			"filter" : "filter",
			"filter_field_name" : "filter_field_name",
			"filter_value" : "filter_value",
			"metrics" : ["metric"],
			"prebuilt_report" : "prebuilt_report",
			"report_type" : "report_type",
			"search_types" : ["search_type"],
			"segment_ids" : ["segment_id"],
			"segments" : ["segment"],
			"table" : "table1"
		 },{
			"aggregation" : "aggregation",
			"attributes" : ["attribute"],
			"config_type" : "config_type",
			"dimensions" : ["dimension"],
			"fields" : ["field"],
			"filter" : "filter",
			"filter_field_name" : "filter_field_name",
			"filter_value" : "filter_value",
			"metrics" : ["metric"],
			"prebuilt_report" : "prebuilt_report",
			"report_type" : "report_type",
			"search_types" : ["search_type"],
			"segment_ids" : ["segment_id"],
			"segments" : ["segment"],
			"table" : "table2"
		 }]
		`,
	)
}

func testResourceConnectorConfigConflictingFieldsMappingMock(t *testing.T, service, destinationSchema, schema, tfConfig, jsonConfig string) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: getTfConfigForConflictingFields(service, destinationSchema, tfConfig),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorConflictingMockPostHandler.Interactions, 1)
				assertEqual(t, connectorConflictingMockGetHandler.Interactions, 1)
				assertNotEmpty(t, connectorConflictingMappingMockData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceConfigConflictingFieldsMapping(t, service, schema, jsonConfig)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, connectorConflictingMockDelete.Interactions, 1)
				assertEmpty(t, connectorConflictingMappingMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
