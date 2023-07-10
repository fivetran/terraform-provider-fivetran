package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	connectorEmptyMockGetHandler  *mock.Handler
	connectorEmptyMockPostHandler *mock.Handler
	connectorEmptyMockDelete      *mock.Handler

	connectorListsMockGetHandler  *mock.Handler
	connectorListsMockPostHandler *mock.Handler
	connectorListsMockDelete      *mock.Handler

	connectorMockUpdateGetHandler   *mock.Handler
	connectorMockUpdatePostHandler  *mock.Handler
	connectorMockUpdatePatchHandler *mock.Handler
	connectorMockUpdateDelete       *mock.Handler

	connectorMockData map[string]interface{}
)

const (
	connectorWithoutConfig = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "postgres",
        "service_version": 1,
        "schema": "postgres",
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
            "tasks": [],
            "warnings": []
        },
        "config": {}
	}
	`

	connectorUpdateResponse1 = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "postgres",
        "service_version": 1,
        "schema": "postgres",
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
            "user": "user",
			"password": "password"
		}
	}
	`

	connectorUpdateResponse2 = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "postgres",
        "service_version": 1,
        "schema": "postgres",
        "paused": false,
        "pause_after_trial": false,
        "sync_frequency": 1440,
		"daily_sync_time": "3:30",

		"connected_by": "user_id",
        "created_at": "2022-01-01T11:22:33.012345Z",
        "succeeded_at": null,
        "failed_at": null,
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
            "user": "user1",
			"password": "password1",
			"host": "host",
			"port": 123
		}
	}
	`

	connectorConfigListsMappingResponse = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "google_sheets",
        "service_version": 1,
        "schema": "google_sheets_schema.table",
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
			"packed_mode_tables":["packed_mode_table_3", "packed_mode_table_2", "packed_mode_table_1"],
			"properties":["property_2", "property_1"],` +

		// "app_ids": ["value_2", "value_1"],
		// "conversion_dimensions": ["value_2", "value_1"],
		// "custom_floodlight_variables": ["value_2", "value_1"],
		// "partners": ["value_2", "value_1"],
		// "per_interaction_dimensions": ["value_2", "value_1"],
		// "schema_registry_urls": ["value_2", "value_1"],
		// "topics": ["value_2", "value_1"],
		// "servers": ["value_2", "value_1"],
		// "segments": ["value_2", "value_1"],

		`"primary_keys":["primary_key_2", "primary_key_1"],
			"report_suites": ["value_2", "value_1"],
			"elements": ["value_2", "value_1"],
			"metrics": ["value_2", "value_1"],
			"advertisables": ["value_2", "value_1"],
			"dimensions": ["value_2", "value_1"],
			"selected_exports": ["value_2", "value_1"],
			"apps": ["value_2", "value_1"],
			"sales_accounts": ["value_2", "value_1"],
			"finance_accounts": ["value_2", "value_1"],
			"projects": ["value_2", "value_1"],
			"user_profiles": ["value_2", "value_1"],
			"report_configuration_ids": ["value_2", "value_1"],
			"accounts": ["value_2", "value_1"],
			"fields": ["value_2", "value_1"],
			"breakdowns": ["value_2", "value_1"],
			"action_breakdowns": ["value_2", "value_1"],
			"pages": ["value_2", "value_1"],
			"repositories": ["value_2", "value_1"],
			"dimension_attributes": ["value_2", "value_1"],
			"columns": ["value_2", "value_1"],
			"manager_accounts": ["value_2", "value_1"],
			"profiles": ["value_2", "value_1"],
			"site_urls": ["value_2", "value_1"],
			"api_keys": ["value_2", "value_1"],
			"advertisers_id": ["value_2", "value_1"],
			"hosts": ["value_2", "value_1"],
			"advertisers": ["value_2", "value_1"],
			"organizations": ["value_2", "value_1"],
			"account_ids": ["value_2", "value_1"]
		}
	}
	`

	connectorConfigListsMappingTfConfig = `
	resource "fivetran_connector" "test_connector" {
		provider = fivetran-provider

		group_id = "group_id"
		service = "google_sheets"

		destination_schema {
			name = "google_sheets_schema"
			table = "table"
		}

		trust_certificates = false
		trust_fingerprints = false
		run_setup_tests = false

		config {
			packed_mode_tables = ["packed_mode_table_1", "packed_mode_table_2", "packed_mode_table_3"]
			properties = ["property_1", "property_2"]
			primary_keys = ["primary_key_1", "primary_key_2"]

			# app_ids = ["value_1", "value_2"]
			# conversion_dimensions = ["value_1", "value_2"]
			# custom_floodlight_variables = ["value_1", "value_2"]
			# partners = ["value_1", "value_2"]
			# per_interaction_dimensions = ["value_1", "value_2"]
			# schema_registry_urls = ["value_1", "value_2"]
			# segments = ["value_1", "value_2"]

			metrics = ["value_1", "value_2"]
			advertisables = ["value_1", "value_2"]
			dimensions = ["value_1", "value_2"]
			selected_exports = ["value_1", "value_2"]
			apps = ["value_1", "value_2"]
			sales_accounts = ["value_1", "value_2"]
			finance_accounts = ["value_1", "value_2"]
			projects = ["value_1", "value_2"]
			user_profiles = ["value_1", "value_2"]
			report_configuration_ids = ["value_1", "value_2"]
			accounts = ["value_1", "value_2"]
			fields = ["value_1", "value_2"]
			breakdowns = ["value_1", "value_2"]
			action_breakdowns = ["value_1", "value_2"]
			pages = ["value_1", "value_2"]
			repositories = ["value_1", "value_2"]
			dimension_attributes = ["value_1", "value_2"]
			columns = ["value_1", "value_2"]
			manager_accounts = ["value_1", "value_2"]
			profiles = ["value_1", "value_2"]
			site_urls = ["value_1", "value_2"]
			api_keys = ["value_1", "value_2"]
			advertisers_id = ["value_1", "value_2"]
			hosts = ["value_1", "value_2"]
			advertisers = ["value_1", "value_2"]
			organizations = ["value_1", "value_2"]
			account_ids = ["value_1", "value_2"]
			elements = ["value_1", "value_2"]
			report_suites = ["value_1", "value_2"]
		}
	}
	`
)

func setupMockClientConnectorResourceListMappingConfig(t *testing.T) {
	mockClient.Reset()

	connectorListsMockGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorListsMockPostHandler = mockClient.When(http.MethodPost, "/v1/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = createMapFromJsonString(t, connectorConfigListsMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorListsMockDelete = mockClient.When(http.MethodDelete, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func setupMockClientConnectorResourceUpdate(t *testing.T) {
	mockClient.Reset()
	updateIteration := 0

	checkPatternNotRepresentedIfNotSet := func(t *testing.T, body map[string]interface{}) {
		if c, ok := body["config"]; ok {
			config := c.(map[string]interface{})
			_, ok := config["pattern"]
			assertEqual(t, ok, false)
		}
	}

	connectorMockUpdateGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdatePostHandler = mockClient.When(http.MethodPost, "/v1/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			checkPatternNotRepresentedIfNotSet(t, requestBodyToJson(t, req))
			connectorMockData = createMapFromJsonString(t, connectorUpdateResponse1)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdatePatchHandler = mockClient.When(http.MethodPatch, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			updateIteration++
			checkPatternNotRepresentedIfNotSet(t, requestBodyToJson(t, req))
			if updateIteration == 1 {
				connectorMockData = createMapFromJsonString(t, connectorUpdateResponse1)
			} else {
				connectorMockData = createMapFromJsonString(t, connectorUpdateResponse2)
			}
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdateDelete = mockClient.When(http.MethodDelete, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func setupMockClientConnectorResourceEmptyConfig(t *testing.T) {
	mockClient.Reset()

	connectorEmptyMockGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorEmptyMockPostHandler = mockClient.When(http.MethodPost, "/v1/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = createMapFromJsonString(t, connectorWithoutConfig)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorEmptyMockDelete = mockClient.When(http.MethodDelete, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func TestResourceConnectorUpdateMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "postgres"

			destination_schema {
				prefix = "postgres"
			}

			trust_certificates = false
			trust_fingerprints = false
			run_setup_tests = false

			config {
				user = "user"
				password = "password"
			}
		}

		resource "fivetran_connector_schedule" "test_connector_schedule" {
			provider = fivetran-provider

			connector_id = "connector_id"
			sync_frequency = 5
			paused = true
			pause_after_trial = true
		}
		`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorMockUpdatePostHandler.Interactions, 1)
				assertEqual(t, connectorMockUpdateGetHandler.Interactions, 3)
				assertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "postgres"),
		),
	}

	step2 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "postgres"

			destination_schema {
				prefix = "postgres"
			}

			trust_certificates = true
			trust_fingerprints = true
			run_setup_tests = true

			config {
				user = "user1"
				password = "password1"
				host = "host"
				port = "123"
			}
		}

		resource "fivetran_connector_schedule" "test_connector_schedule" {
			provider = fivetran-provider

			connector_id = "connector_id"
			sync_frequency = 1440
			paused = false
			pause_after_trial = false
			daily_sync_time = "3:30"
		}
		`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorMockUpdatePostHandler.Interactions, 1)
				assertEqual(t, connectorMockUpdateGetHandler.Interactions, 9)
				assertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "postgres"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceUpdate(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, connectorMockUpdateDelete.Interactions, 1)
				assertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}

func TestResourceConnectorEmptyConfigMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "postgres"

			destination_schema {
				prefix = "postgres"
			}

			trust_certificates = false
			trust_fingerprints = false
			run_setup_tests = false

			#config {}
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorEmptyMockPostHandler.Interactions, 1)
				assertEqual(t, connectorEmptyMockGetHandler.Interactions, 1)
				assertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "config"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceEmptyConfig(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, connectorEmptyMockDelete.Interactions, 1)
				assertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestResourceConnectorListsConfigMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: connectorConfigListsMappingTfConfig,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorListsMockPostHandler.Interactions, 1)
				assertEqual(t, connectorListsMockGetHandler.Interactions, 1)
				assertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "config"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceListMappingConfig(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, connectorListsMockDelete.Interactions, 1)
				assertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
