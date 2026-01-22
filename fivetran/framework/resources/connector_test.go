package resources_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

	connectorMockUpdateHdPostHandler  *mock.Handler
	connectorMockUpdateHdGetHandler   *mock.Handler
	connectorMockUpdateHdPatchHandler *mock.Handler

	connectorMockData map[string]interface{}
	postRequestJson map[string]interface{}
	patchRequestJson map[string]interface{}
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
		"networking_method": "Directly",
		"hybrid_deployment_agent_id": "hybrid_deployment_agent_id",
		"data_delay_sensitivity": "NORMAL",
		"data_delay_threshold": 0,
        "status": {
            "setup_state": "incomplete",
            "sync_state": "paused",
            "update_state": "on_schedule",
            "is_historical_sync": true,
            "tasks": [],
            "warnings": []
        },
        "config": {
			"port": 123
		}
	}
	`

	coilConnectorCreateResponse1 = `
	{
		"id": "connector_id",
        "group_id": "group_id",
		"service": "microsoft_entra_id",
		"service_version": 0,
		"schema": "microsoft_entra_id",
		"connected_by": "user_id",
		"created_at": "2025-11-19T11:14:34.391353Z",
		"succeeded_at": null,
		"failed_at": null,
		"paused": true,
		"pause_after_trial": false,
		"sync_frequency": 360,
		"data_delay_threshold": 0,
		"data_delay_sensitivity": "NORMAL",
		"private_link_id": null,
		"networking_method": "Directly",
		"proxy_agent_id": null,
		"schedule_type": "auto",
		"status": {
			"setup_state": "connected",
			"schema_status": "ready",
			"sync_state": "paused",
			"update_state": "on_schedule",
			"is_historical_sync": true,
			"tasks": [],
			"warnings": []
		},
		"setup_tests": [
			{ "title": "Connecting to API", "status": "PASSED", "message": "" }
		],
		"config": {
			"tenant": "tenant1",
			"client_id": "client_id1",
			"client_secret": "******"
		}
	}
		`
	
	coilConnectorPatchResponse1 = `
	{
		"id": "connector_id",
        "group_id": "group_id",
		"service": "microsoft_entra_id",
		"service_version": 0,
		"schema": "microsoft_entra_id",
		"connected_by": "user_id",
		"created_at": "2025-11-19T11:14:34.391353Z",
		"failed_at": "2025-11-19T11:14:34.391353Z",
		"succeeded_at": "2025-11-19T11:14:34.391353Z",
		"paused": false,
		"pause_after_trial": false,
		"sync_frequency": 30,
		"data_delay_threshold": 0,
		"data_delay_sensitivity": "NORMAL",
		"private_link_id": null,
		"networking_method": "Directly",
		"proxy_agent_id": null,
		"schedule_type": "auto",
		"status": {
			"setup_state": "connected",
			"schema_status": "ready",
			"sync_state": "scheduled",
			"update_state": "on_schedule",
			"is_historical_sync": true,
			"tasks": [],
			"warnings": []
		},
		"config": {
			"tenant": "tenant1",
			"client_id": "client_id1",
			"client_secret": "******"
		}
	}
		`

	coilConnectorPatchResponse2 = `
	{
		"id": "connector_id",
        "group_id": "group_id",
		"service": "microsoft_entra_id",
		"service_version": 0,
		"schema": "microsoft_entra_id",
		"connected_by": "user_id",
		"created_at": "2025-11-19T11:14:34.391353Z",
		"failed_at": "2025-11-19T11:14:34.391353Z",
		"succeeded_at": "2025-11-19T11:14:34.391353Z",
		"paused": false,
		"pause_after_trial": false,
		"sync_frequency": 30,
		"data_delay_threshold": 0,
		"data_delay_sensitivity": "NORMAL",
		"private_link_id": null,
		"networking_method": "Directly",
		"proxy_agent_id": null,
		"schedule_type": "auto",
		"status": {
			"setup_state": "connected",
			"schema_status": "ready",
			"sync_state": "scheduled",
			"update_state": "on_schedule",
			"is_historical_sync": true,
			"tasks": [],
			"warnings": []
		},
		"config": {
			"tenant": "tenant1",
			"client_id": "client_id2"
		}
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
		"networking_method": "Directly",
		"data_delay_sensitivity": "NORMAL",
		"data_delay_threshold": 0,
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
		"daily_sync_time": "03:00",

		"connected_by": "user_id",
        "created_at": "2022-01-01T11:22:33.012345Z",
        "succeeded_at": null,
        "failed_at": null,
		"schedule_type": "auto",
		"networking_method": "Directly",
		"data_delay_sensitivity": "NORMAL",
		"data_delay_threshold": 0,
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

	connectorUpdateResponseHd1 = `
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
		"networking_method": "Directly",
		"data_delay_sensitivity": "NORMAL",
		"data_delay_threshold": 0,
		"hybrid_deployment_agent_id": "existing_agent_id",
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

	connectorUpdateResponseHd2 = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "postgres",
        "service_version": 1,
        "schema": "postgres",
        "paused": false,
        "pause_after_trial": false,
        "sync_frequency": 1440,
		"daily_sync_time": "03:00",

		"connected_by": "user_id",
        "created_at": "2022-01-01T11:22:33.012345Z",
        "succeeded_at": null,
        "failed_at": null,
		"schedule_type": "auto",
		"networking_method": "Directly",
		"data_delay_sensitivity": "NORMAL",
		"data_delay_threshold": 0,
		"hybrid_deployment_agent_id": "new_agent_id",
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
		"networking_method": "Directly",
		"data_delay_sensitivity": "NORMAL",
		"data_delay_threshold": 0,
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
			"properties":["property_2", "property_1"],
			"primary_keys":["primary_key_2", "primary_key_1"],
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
			"account_ids": ["value_2", "value_1"],
			"segments": ["value_2", "value_1"],
			"schema_registry_urls": ["value_2", "value_1"],
			"per_interaction_dimensions": ["value_2", "value_1"],
			"partners": ["value_2", "value_1"],
			"custom_floodlight_variables": ["value_2", "value_1"],
			"conversion_dimensions": ["value_2", "value_1"],
			"app_ids": ["value_2", "value_1"]
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

		data_delay_sensitivity = "NORMAL"
		data_delay_threshold = 0

		trust_certificates = false
		trust_fingerprints = false
		run_setup_tests = false

		config {
			packed_mode_tables = ["packed_mode_table_1", "packed_mode_table_2", "packed_mode_table_3"]
			properties = ["property_1", "property_2"]
			primary_keys = ["primary_key_1", "primary_key_2"]

			app_ids = ["value_1", "value_2"]
			conversion_dimensions = ["value_1", "value_2"]
			custom_floodlight_variables = ["value_1", "value_2"]
			partners = ["value_1", "value_2"]
			per_interaction_dimensions = ["value_1", "value_2"]
			schema_registry_urls = ["value_1", "value_2"]
		    segments = ["value_1", "value_2"]

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

	workdayConnectorPostResponse = `
	{
		"id": "connection_id",
		"group_id": "group_id",
		"service": "workday_adaptive",
		"service_version": 0,
		"schema": "workday_adaptive",
		"connected_by": "user_id",
		"created_at": "2026-01-01T05:27:48Z",
		"succeeded_at": null,
		"failed_at": null,
		"paused": true,
		"pause_after_trial": false,
		"sync_frequency": 360,
		"data_delay_threshold": 0,
		"data_delay_sensitivity": "NORMAL",
		"private_link_id": null,
		"networking_method": "Directly",
		"proxy_agent_id": null,
		"schedule_type": "auto",
		"status": {
			"setup_state": "connected",
			"schema_status": "ready",
			"sync_state": "paused",
			"update_state": "on_schedule",
			"is_historical_sync": true,
			"tasks": [],
			"warnings": []
		},
		"setup_tests": [
			{
			"title": "Validating report parameters",
			"status": "PASSED",
			"message": ""
			},
			{
			"title": "Validating adaptive credentials",
			"status": "PASSED",
			"message": ""
			}
		],
		"config": {
			"login": "mail@mail.com",
			"password": "******",
			"instance_code": "INSTANCE_CODE1",
			"reports": [
				{
					"table": "table_1",
					"version_sync_strategy": "SYNC_SELECT_VERSIONS",
					"versions": ["2945"],
					"accounts": [
						{ "id": "3", "flag": true, "include_descendants": true },
						{ "id": "4", "flag": true }
					],
					"dimensions": ["3", "4"],
					"include_zero_rows": true
				}
			]
		}
	}
	`

	workdayConnectorPatchResponse = `
	{
		"id": "connection_id",
		"group_id": "group_id",
		"service": "workday_adaptive",
		"service_version": 0,
		"schema": "workday_adaptive",
		"connected_by": "user_id",
		"created_at": "2026-01-01T05:27:48Z",
		"succeeded_at": null,
		"failed_at": null,
		"paused": true,
		"pause_after_trial": false,
		"sync_frequency": 360,
		"data_delay_threshold": 0,
		"data_delay_sensitivity": "NORMAL",
		"private_link_id": null,
		"networking_method": "Directly",
		"proxy_agent_id": null,
		"schedule_type": "auto",
		"status": {
			"setup_state": "connected",
			"schema_status": "ready",
			"sync_state": "paused",
			"update_state": "on_schedule",
			"is_historical_sync": true,
			"tasks": [],
			"warnings": []
		},
		"setup_tests": [
			{
			"title": "Validating report parameters",
			"status": "PASSED",
			"message": ""
			},
			{
			"title": "Validating adaptive credentials",
			"status": "PASSED",
			"message": ""
			}
		],
		"config": {
			"login": "mail@mail.com",
			"password": "******",
			"instance_code": "INSTANCE_CODE1",
			"reports": [
				{
					"table": "table_1",
					"version_sync_strategy": "SYNC_SELECT_VERSIONS",
					"versions": ["2945"],
					"accounts": [
						{ "id": "3", "flag": true, "include_descendants": false },
						{ "id": "4", "flag": true },
						{ "id": "5", "flag": true }
					],
					"dimensions": ["3", "4"],
					"include_zero_rows": true
				},
				{
					"table": "table_2",
					"version_sync_strategy": "SYNC_SELECT_VERSIONS",
					"versions": ["2945"],
					"accounts": [
						{ "id": "11", "flag": true },
						{ "id": "12", "flag": true }
					],
					"dimensions": ["11"],
					"include_zero_rows": true
				}
			]
		}
	}
	`
)

func setupMockClientConnectorResourceListMappingConfig(t *testing.T) {
	tfmock.MockClient().Reset()

	connectorListsMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorListsMockPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = tfmock.CreateMapFromJsonString(t, connectorConfigListsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorListsMockDelete = tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func setupMockClientConnectorResourceUpdate(t *testing.T) {
	tfmock.MockClient().Reset()
	updateIteration := 0

	checkPatternNotRepresentedIfNotSet := func(t *testing.T, body map[string]interface{}) {
		if c, ok := body["config"]; ok {
			config := c.(map[string]interface{})
			_, ok := config["pattern"]
			tfmock.AssertEqual(t, ok, false)
		}
	}

	connectorMockUpdateGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdatePostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			checkPatternNotRepresentedIfNotSet(t, tfmock.RequestBodyToJson(t, req))
			connectorMockData = tfmock.CreateMapFromJsonString(t, connectorUpdateResponse1)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdatePatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			updateIteration++
			checkPatternNotRepresentedIfNotSet(t, tfmock.RequestBodyToJson(t, req))
			if updateIteration == 1 {
				connectorMockData = tfmock.CreateMapFromJsonString(t, connectorUpdateResponse1)
			} else {
				connectorMockData = tfmock.CreateMapFromJsonString(t, connectorUpdateResponse2)
			}
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdateDelete = tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	tfmock.MockClient().When(http.MethodPost, "/v1/connections/connector_id/test").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func setupMockClientConnectorResourceUpdateHd(t *testing.T) {
	tfmock.MockClient().Reset()

	checkPatternNotRepresentedIfNotSet := func(t *testing.T, body map[string]interface{}) {
		if c, ok := body["config"]; ok {
			config := c.(map[string]interface{})
			_, ok := config["pattern"]
			tfmock.AssertEqual(t, ok, false)
		}
	}

	connectorMockUpdateHdGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdateHdPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			checkPatternNotRepresentedIfNotSet(t, tfmock.RequestBodyToJson(t, req))
			connectorMockData = tfmock.CreateMapFromJsonString(t, connectorUpdateResponseHd1)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdateHdPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			checkPatternNotRepresentedIfNotSet(t, tfmock.RequestBodyToJson(t, req))
			connectorMockData = tfmock.CreateMapFromJsonString(t, connectorUpdateResponseHd2)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdateDelete = tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	tfmock.MockClient().When(http.MethodPost, "/v1/connections/connector_id/test").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func setupMockClientConnectorResourceWorkday(t *testing.T) {
	tfmock.MockClient().Reset()
	connectorMockData = nil

	connectorMockUpdateGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connections/connection_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdatePostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			// body = tfmock.RequestBodyToJson(t, req))
			connectorMockData = tfmock.CreateMapFromJsonString(t, workdayConnectorPostResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdatePatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/connections/connection_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			// body = tfmock.RequestBodyToJson(t, req))
			connectorMockData = tfmock.CreateMapFromJsonString(t, workdayConnectorPatchResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdateDelete = tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connection_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func setupMockClientConnectorResourceEmptyConfig(t *testing.T) {
	tfmock.MockClient().Reset()

	connectorEmptyMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorEmptyMockPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = tfmock.CreateMapFromJsonString(t, connectorWithoutConfig)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorEmptyMockDelete = tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
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

			data_delay_sensitivity = "NORMAL"
			data_delay_threshold = 0

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
			schedule_type = "auto"
		}
		`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, connectorMockData)
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

			data_delay_sensitivity = "NORMAL"
			data_delay_threshold = 0

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
			daily_sync_time = "03:00"
		}
		`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 4)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 3)
				tfmock.AssertNotEmpty(t, connectorMockData)
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
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdateDelete.Interactions, 1)
				tfmock.AssertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}

func setupMockClientConnectorResourceUpdateConfig(t *testing.T) {
	tfmock.MockClient().Reset()
	updateIteration := 0

	connectorMockUpdateGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdatePostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			postRequestJson = tfmock.RequestBodyToJson(t, req)
			connectorMockData = tfmock.CreateMapFromJsonString(t, coilConnectorCreateResponse1)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdatePatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			updateIteration++

			patchRequestJson = tfmock.RequestBodyToJson(t, req)
			 if updateIteration <=2 {
				connectorMockData = tfmock.CreateMapFromJsonString(t, coilConnectorPatchResponse1)
			 } else {
				connectorMockData = tfmock.CreateMapFromJsonString(t, coilConnectorPatchResponse2)
			 }
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	connectorMockUpdateDelete = tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMockData = nil
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)

	tfmock.MockClient().When(http.MethodPost, "/v1/connections/connector_id/test").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMockData), nil
		},
	)
}

func TestResourceUpdateSensitiveFieldInConfigMock(t *testing.T) {
	//fivetran.Debug(true)
	step1 := resource.TestStep{
		Config: `
		locals {
		  config = jsondecode("{\"tenant_id\":\"tenant1\",\"client_id\":\"client_id1\",\"client_secret\":\"client_secret1\"}")
		}
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service           = "microsoft_entra_id"
			networking_method = "Directly"
			run_setup_tests   = true
			destination_schema {
				name = "microsoft_entra_id"
			}
			config {
				tenant        = sensitive(local.config.tenant_id)
				client_id     = sensitive(local.config.client_id)
				client_secret = sensitive(local.config.client_secret)
			}
		}

		resource "fivetran_connector_schedule" "test_connector_schedule" {
			provider = fivetran-provider

			connector_id   = fivetran_connector.test_connector.id
			sync_frequency = 30
			lifecycle {
				ignore_changes = [sync_frequency]
			}
		}
		`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, connectorMockData)
				postRequestConfig := postRequestJson["config"].(map[string]interface{})
				tfmock.AssertEqual(t, postRequestConfig["client_secret"], "client_secret1")
				tfmock.AssertEqual(t, postRequestConfig["client_id"], "client_id1")
				tfmock.AssertEqual(t, postRequestConfig["tenant"], "tenant1")

				tfmock.AssertEmpty(t, patchRequestJson["config"]) // patch invocation from fivetran_connector_schedule resource creation should not contain config
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "microsoft_entra_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.tenant", "tenant1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_secret", "client_secret1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_id", "client_id1"),
		),
	}

	step2 := resource.TestStep{
		Config: `
		locals {
		  config = jsondecode("{\"tenant_id\":\"tenant1\",\"client_id\":\"client_id1\",\"client_secret\":\"client_secret2\"}")
		}
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service           = "microsoft_entra_id"
			networking_method = "Directly"
			run_setup_tests   = true
			destination_schema {
				name = "microsoft_entra_id"
			}
			config {
				tenant        = sensitive(local.config.tenant_id)
				client_id     = sensitive(local.config.client_id)
				client_secret = sensitive(local.config.client_secret)
			}
		}

		resource "fivetran_connector_schedule" "test_connector_schedule" {
			provider = fivetran-provider

			connector_id   = fivetran_connector.test_connector.id
			sync_frequency = 30
			lifecycle {
				ignore_changes = [sync_frequency]
			}
		}
		`,
		PreConfig: func() {
			connectorMockUpdatePostHandler.Interactions = 0
			connectorMockUpdateGetHandler.Interactions = 0
			connectorMockUpdatePatchHandler.Interactions = 0
			postRequestJson = nil
			patchRequestJson = nil
		},
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 2)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 1)
				patchRequestConfig := patchRequestJson["config"].(map[string]interface{})
				tfmock.AssertEqual(t, patchRequestConfig["client_secret"], "client_secret2")
				tfmock.AssertEmpty(t, patchRequestConfig["client_id"])
				tfmock.AssertEmpty(t, patchRequestConfig["tenant"])
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "microsoft_entra_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.tenant", "tenant1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_secret", "client_secret2"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_id", "client_id1"),
		),
	}

	step3 := resource.TestStep{
		Config: `
		locals {
		  config = jsondecode("{\"tenant_id\":\"tenant1\",\"client_id\":\"client_id2\",\"client_secret\":\"client_secret2\"}")
		}
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service           = "microsoft_entra_id"
			networking_method = "Directly"
			run_setup_tests   = true
			destination_schema {
				name = "microsoft_entra_id"
			}
			config {
				tenant        = sensitive(local.config.tenant_id)
				client_id     = sensitive(local.config.client_id)
				client_secret = sensitive(local.config.client_secret)
			}
		}

		resource "fivetran_connector_schedule" "test_connector_schedule" {
			provider = fivetran-provider

			connector_id   = fivetran_connector.test_connector.id
			sync_frequency = 30
			lifecycle {
				ignore_changes = [sync_frequency]
			}
		}
		`,
		PreConfig: func() {
			connectorMockUpdatePostHandler.Interactions = 0
			connectorMockUpdateGetHandler.Interactions = 0
			connectorMockUpdatePatchHandler.Interactions = 0
			postRequestJson = nil
			patchRequestJson = nil
		},
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 2)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, connectorMockData)
				patchRequestConfig := patchRequestJson["config"].(map[string]interface{})
				
				tfmock.AssertEmpty(t, patchRequestConfig["client_secret"])
				tfmock.AssertEqual(t, patchRequestConfig["client_id"], "client_id2")
				tfmock.AssertEmpty(t, patchRequestConfig["tenant"])
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "microsoft_entra_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.tenant", "tenant1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_secret", "client_secret2"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_id", "client_id2"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceUpdateConfig(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdateDelete.Interactions, 1)
				tfmock.AssertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
			},
		},
	)
}

func TestResourceImportThenUpdateSensitiveFieldInConfigMock(t *testing.T) {

	step1 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider
		}
		`,
		ImportState: true,
		ResourceName: "fivetran_connector.test_connector",
		ImportStateId: "connector_id",
		ImportStatePersist: true,
		PreConfig: func() {
			connectorMockUpdatePostHandler.Interactions = 0
			connectorMockUpdateGetHandler.Interactions = 0
			connectorMockUpdatePatchHandler.Interactions = 0
			postRequestJson = nil
			patchRequestJson = nil
			connectorMockData = tfmock.CreateMapFromJsonString(t, coilConnectorPatchResponse1)
		},
		ImportStateCheck: tfmock.ComposeImportStateCheck(
			func(s []*terraform.InstanceState) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 0)
				return nil
			},
			tfmock.CheckImportResourceAttr("fivetran_connector", "connector_id", "service", "microsoft_entra_id"),
			tfmock.CheckImportResourceAttr("fivetran_connector", "connector_id", "config.tenant", "tenant1"),
			tfmock.CheckNoImportResourceAttr("fivetran_connector", "connector_id", "config.client_secret"), // sensitive field are not imported, as their values are unknown
			tfmock.CheckImportResourceAttr("fivetran_connector", "connector_id", "config.client_id", "client_id1"),
		),
	}

	step2 := resource.TestStep{
		Config: `
		locals {
		  config = jsondecode("{\"tenant_id\":\"tenant1\",\"client_id\":\"client_id1\",\"client_secret\":\"client_secret3\"}")
		}
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service           = "microsoft_entra_id"
			networking_method = "Directly"
			run_setup_tests   = true
			destination_schema {
				name = "microsoft_entra_id"
			}
			config {
				tenant        = sensitive(local.config.tenant_id)
				client_id     = sensitive(local.config.client_id)
				client_secret = sensitive(local.config.client_secret)
			}
		}
		`,
		PreConfig: func() {
			connectorMockUpdatePostHandler.Interactions = 0
			connectorMockUpdateGetHandler.Interactions = 0
			connectorMockUpdatePatchHandler.Interactions = 0
			postRequestJson = nil
			patchRequestJson = nil
		},
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 1)
				patchRequestConfig := patchRequestJson["config"].(map[string]interface{})
				tfmock.AssertEqual(t, patchRequestConfig["client_secret"], "client_secret3")
				tfmock.AssertEmpty(t, patchRequestConfig["client_id"])
				tfmock.AssertEmpty(t, patchRequestConfig["tenant"])
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "microsoft_entra_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.tenant", "tenant1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_secret", "client_secret3"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_id", "client_id1"),
		),
	}

	step3 := resource.TestStep{
		Config: `
		locals {
		  config = jsondecode("{\"tenant_id\":\"tenant1\",\"client_id\":\"client_id1\",\"client_secret\":\"client_secret4\"}")
		}
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service           = "microsoft_entra_id"
			networking_method = "Directly"
			run_setup_tests   = true
			destination_schema {
				name = "microsoft_entra_id"
			}
			config {
				tenant        = sensitive(local.config.tenant_id)
				client_id     = sensitive(local.config.client_id)
				client_secret = sensitive(local.config.client_secret)
			}
		}
		`,
		PreConfig: func() {
			connectorMockUpdatePostHandler.Interactions = 0
			connectorMockUpdateGetHandler.Interactions = 0
			connectorMockUpdatePatchHandler.Interactions = 0
			postRequestJson = nil
			patchRequestJson = nil
		},
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 1)
				patchRequestConfig := patchRequestJson["config"].(map[string]interface{})
				tfmock.AssertEqual(t, patchRequestConfig["client_secret"], "client_secret4")
				tfmock.AssertEmpty(t, patchRequestConfig["client_id"])
				tfmock.AssertEmpty(t, patchRequestConfig["tenant"])
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "microsoft_entra_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.tenant", "tenant1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_secret", "client_secret4"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_id", "client_id1"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceUpdateConfig(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdateDelete.Interactions, 1)
				tfmock.AssertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
			},
		},
	)
}

func TestResourceConsequentUpdatesOfSensitiveFieldInConfigMock(t *testing.T) {

	step1 := resource.TestStep{
		Config: `
		locals {
		  config = jsondecode("{\"tenant_id\":\"tenant1\",\"client_id\":\"client_id1\",\"client_secret\":\"client_secret5\"}")
		}
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service           = "microsoft_entra_id"
			networking_method = "Directly"
			run_setup_tests   = true
			destination_schema {
				name = "microsoft_entra_id"
			}
			config {
				tenant        = sensitive(local.config.tenant_id)
				client_id     = sensitive(local.config.client_id)
				client_secret = sensitive(local.config.client_secret)
			}
		}
		`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 0)
				postRequestConfig := postRequestJson["config"].(map[string]interface{})
				tfmock.AssertEqual(t, postRequestConfig["client_secret"], "client_secret5")
				tfmock.AssertEqual(t, postRequestConfig["client_id"], "client_id1")
				tfmock.AssertEqual(t, postRequestConfig["tenant"], "tenant1")
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "microsoft_entra_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.tenant", "tenant1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_secret", "client_secret5"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_id", "client_id1"),
		),
	}

	step2 := resource.TestStep{
		Config: `
		locals {
		  config = jsondecode("{\"tenant_id\":\"tenant1\",\"client_id\":\"client_id1\",\"client_secret\":\"client_secret6\"}")
		}
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service           = "microsoft_entra_id"
			networking_method = "Directly"
			run_setup_tests   = true
			destination_schema {
				name = "microsoft_entra_id"
			}
			config {
				tenant        = sensitive(local.config.tenant_id)
				client_id     = sensitive(local.config.client_id)
				client_secret = sensitive(local.config.client_secret)
			}
		}
		`,
		PreConfig: func() {
			connectorMockUpdatePostHandler.Interactions = 0
			connectorMockUpdateGetHandler.Interactions = 0
			connectorMockUpdatePatchHandler.Interactions = 0
			postRequestJson = nil
			patchRequestJson = nil
		},
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 1)
				patchRequestConfig := patchRequestJson["config"].(map[string]interface{})
				tfmock.AssertEqual(t, patchRequestConfig["client_secret"], "client_secret6")
				tfmock.AssertEmpty(t, patchRequestConfig["client_id"])
				tfmock.AssertEmpty(t, patchRequestConfig["tenant"])
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "microsoft_entra_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.tenant", "tenant1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_secret", "client_secret6"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_id", "client_id1"),
		),
	}

	step3 := resource.TestStep{
		Config: `
		locals {
		  config = jsondecode("{\"tenant_id\":\"tenant1\",\"client_id\":\"client_id1\",\"client_secret\":\"client_secret7\"}")
		}
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service           = "microsoft_entra_id"
			networking_method = "Directly"
			run_setup_tests   = true
			destination_schema {
				name = "microsoft_entra_id"
			}
			config {
				tenant        = sensitive(local.config.tenant_id)
				client_id     = sensitive(local.config.client_id)
				client_secret = sensitive(local.config.client_secret)
			}
		}
		`,
		PreConfig: func() {
			connectorMockUpdatePostHandler.Interactions = 0
			connectorMockUpdateGetHandler.Interactions = 0
			connectorMockUpdatePatchHandler.Interactions = 0
			postRequestJson = nil
			patchRequestJson = nil
		},
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 1)
				patchRequestConfig := patchRequestJson["config"].(map[string]interface{})
				tfmock.AssertEqual(t, patchRequestConfig["client_secret"], "client_secret7")
				tfmock.AssertEmpty(t, patchRequestConfig["client_id"])
				tfmock.AssertEmpty(t, patchRequestConfig["tenant"])
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "microsoft_entra_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.tenant", "tenant1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_secret", "client_secret7"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.client_id", "client_id1"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceUpdateConfig(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdateDelete.Interactions, 1)
				tfmock.AssertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
			},
		},
	)
}

func TestResourceConnectorHybridDeploymentMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "postgres"

			hybrid_deployment_agent_id = "existing_agent_id"
			data_delay_threshold = 0

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
		`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdateHdPostHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdateHdGetHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdateHdPatchHandler.Interactions, 0)
				tfmock.AssertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "postgres"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "hybrid_deployment_agent_id", "existing_agent_id"),
		),
	}

	step2 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "postgres"

			hybrid_deployment_agent_id = "new_agent_id"
			data_delay_threshold = 0

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
		`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdateHdPostHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdateHdGetHandler.Interactions, 2)
				tfmock.AssertEqual(t, connectorMockUpdateHdPatchHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "postgres"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "hybrid_deployment_agent_id", "new_agent_id"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceUpdateHd(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdateDelete.Interactions, 1)
				tfmock.AssertEmpty(t, connectorMockData)
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

			data_delay_sensitivity = "NORMAL"
			data_delay_threshold = 0

			destination_schema {
				prefix = "postgres"
			}

			trust_certificates = false
			trust_fingerprints = false
			run_setup_tests = false

			timeouts {
				create = "0"
			}

			config {

			}
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorEmptyMockPostHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorEmptyMockGetHandler.Interactions, 0)
				tfmock.AssertNotEmpty(t, connectorMockData)
				return nil
			},
			//resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "config"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceEmptyConfig(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorEmptyMockDelete.Interactions, 1)
				tfmock.AssertEmpty(t, connectorMockData)
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
				tfmock.AssertEqual(t, connectorListsMockPostHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorListsMockGetHandler.Interactions, 0)
				tfmock.AssertNotEmpty(t, connectorMockData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceListMappingConfig(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorListsMockDelete.Interactions, 1)
				tfmock.AssertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func createConnectorTestResponseJsonMock(id, groupId, service, schema, table_group_name, config string) string {
	template := `
	{
		"id": "%v",
		"group_id": "%v",
		"service": "%v",
		"service_version": 0,
		"schema": "%v",
		"table_group_name": "%v",
		"paused": true,
		"pause_after_trial": true,
		"connected_by": "monitoring_assuring",
		"created_at": "2020-03-11T15:03:55.743708Z",
		"succeeded_at": "2020-03-17T12:31:40.870504Z",
		"failed_at": "2021-01-15T10:55:00.056497Z",
		"sync_frequency": 360,
		"data_delay_sensitivity": "NORMAL",
		"data_delay_threshold": 0,
		"schedule_type": "auto",
		"networking_method": "Directly",
		"status": {
			"setup_state": "incomplete",
			"schema_status": "ready",
			"sync_state": "scheduled",
			"update_state": "delayed",
			"is_historical_sync": false,
			"tasks": [
				{
					"code": "reconnect",
					"message": "Reconnect"
				}
			],
			"warnings": []
		},
		"config": 
		%v
	}
	`
	return fmt.Sprintf(template, id, groupId, service, schema, table_group_name, config)
}

func TestResourceConnectorNoDestinationSchemaMock(t *testing.T) {
	step1 :=
		resource.TestStep{Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id           = "group_id"
			service            = "amplitude"
		}`,
			ExpectError: regexp.MustCompile("Unable to Create Connector Resource."),
		}

	resource.Test(
		t,
		resource.TestCase{

			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestResourceConnectorUnknownServiceMock(t *testing.T) {
	step1 :=
		resource.TestStep{Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id           = "group_id"
			service            = "unknown-service-name"

			data_delay_sensitivity = "NORMAL"
			data_delay_threshold = 0

			destination_schema {
				name = "schema"
			}
			config {}
		}`,
			ExpectError: regexp.MustCompile("Unable to Create Connector Resource."),
		}

	resource.Test(
		t,
		resource.TestCase{

			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestResourceConnectorMock(t *testing.T) {
	var postHandler *mock.Handler
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			service = "google_ads"
			group_id = "group_id"

			data_delay_sensitivity = "NORMAL"
			data_delay_threshold = 0

			destination_schema {
				name = "adwords_schema"
			}

			config {
				user = "user_name"
				password = "password"
				port = 5432
				account_ids = ["id1", "id2", "id3"]

				reports {
					table = "table1"
					report_type = "report_1"
					metrics = ["metric1", "metric2"]
				}
				reports {
					table = "table2"
					report_type = "report_2"
					metrics = ["metric2", "metric3"]
				}
			}
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, postHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "id", "connector_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "adwords_schema"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "group_id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "google_ads"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "destination_schema.name", "adwords_schema"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_sensitivity", "NORMAL"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_threshold", "0"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.user", "user_name"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.password", "password"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.port", "5432"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.account_ids.0", "id1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.account_ids.1", "id2"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.account_ids.2", "id3"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.0.report_type", "report_1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.0.metrics.0", "metric1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.0.metrics.1", "metric2"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.1.report_type", "report_2"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.1.metrics.0", "metric2"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.1.metrics.1", "metric3"),
		),
	}

	step2 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			service = "google_ads"
			group_id = "group_id"

			data_delay_sensitivity = "NORMAL"
			data_delay_threshold = 0
			
			destination_schema {
				name = "adwords_schema"
			}

			run_setup_tests = true
			trust_certificates = true
			trust_fingerprints = true

			config {
				user = "user_name_1"
				password = "password_1"
				port = 2345
				always_encrypted = false

				reports {
					table = "table1"
					report_type = "report_1"
					metrics = ["metric1", "metric2"]
				}
			}
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, postHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "id", "connector_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "adwords_schema"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "group_id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "google_ads"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "destination_schema.name", "adwords_schema"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_sensitivity", "NORMAL"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_threshold", "0"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.user", "user_name_1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.password", "password_1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.port", "2345"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.always_encrypted", "false"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.0.report_type", "report_1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.0.metrics.0", "metric1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.0.metrics.1", "metric2"),
		),
	}

	var responseData map[string]interface{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				//getHandler =
				tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if responseData == nil {
							return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "NotFound", nil), nil
						}
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
					},
				)

				tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
					},
				)

				postHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
					func(req *http.Request) (*http.Response, error) {

						body := tfmock.RequestBodyToJson(t, req)

						// Check the request
						tfmock.AssertKeyExistsAndHasValue(t, body, "service", "google_ads")
						tfmock.AssertKeyExistsAndHasValue(t, body, "group_id", "group_id")
						tfmock.AssertKeyExistsAndHasValue(t, body, "run_setup_tests", false)
						tfmock.AssertKeyExistsAndHasValue(t, body, "trust_certificates", false)
						tfmock.AssertKeyExistsAndHasValue(t, body, "trust_fingerprints", false)

						if config, ok := tfmock.AssertKeyExists(t, body, "config").(map[string]interface{}); ok {
							tfmock.AssertKeyExistsAndHasValue(t, config, "schema", "adwords_schema")
							tfmock.AssertKeyExistsAndHasValue(t, config, "user", "user_name")
							tfmock.AssertKeyExistsAndHasValue(t, config, "password", "password")
							tfmock.AssertKeyExistsAndHasValue(t, config, "port", float64(5432))
							if reports, ok := tfmock.AssertKeyExists(t, config, "reports").([]interface{}); ok {
								tfmock.AssertEqual(t, len(reports), 2)
							}
							if accountIds, ok := tfmock.AssertKeyExists(t, config, "account_ids").([]interface{}); ok {
								tfmock.AssertEqual(t, len(accountIds), 3)
							}
						}

						responseJson := createConnectorTestResponseJsonMock(
							"connector_id",
							"group_id",
							"google_ads",
							"adwords_schema",
							"adwords_schema",
							`{
								"user": "user_name",
								"password": "******",
								"port": 5432,
								"always_encrypted": true,
								"account_ids": ["id1", "id2", "id3"],
								"reports": [
									{
										"table": "table1",
										"report_type": "report_1",
										"metrics": ["metric1", "metric2"]
									},
									{
										"table": "table2",
										"report_type": "report_2",
										"metrics": ["metric2", "metric3"]
									}
								]
							}`,
						)

						responseData = tfmock.CreateMapFromJsonString(t, responseJson)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", responseData), nil
					},
				)

				tfmock.MockClient().When(http.MethodPatch, "/v1/connections/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {

						body := tfmock.RequestBodyToJson(t, req)

						// Check the request
						tfmock.AssertEqual(t, len(body), 2)

						if config, ok := tfmock.AssertKeyExists(t, body, "config").(map[string]interface{}); ok {
							tfmock.AssertKeyExistsAndHasValue(t, config, "account_ids", nil)
							tfmock.AssertKeyExistsAndHasValue(t, config, "user", "user_name_1")
							tfmock.AssertKeyExistsAndHasValue(t, config, "password", "password_1")
							tfmock.AssertKeyExistsAndHasValue(t, config, "port", float64(2345))
							if reports, ok := tfmock.AssertKeyExists(t, config, "reports").([]interface{}); ok {
								tfmock.AssertEqual(t, len(reports), 1)
							}
						}

						responseJson := createConnectorTestResponseJsonMock(
							"connector_id",
							"group_id",
							"google_ads",
							"adwords_schema",
							"adwords_schema",
							`{
								"user": "user_name_1",
								"password": "******",
								"port": 2345,
								"always_encrypted": false,
								"reports": [
									{
										"table": "table1",
										"report_type": "report_1",
										"metrics": ["metric1", "metric2"]
									}
								]
							}`,
						)

						responseData = tfmock.CreateMapFromJsonString(t, responseJson)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
					},
				)

			tfmock.MockClient().When(http.MethodPost, "/v1/connections/connector_id/test").ThenCall(
				func(req *http.Request) (*http.Response, error) {
					return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
				},
			)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}

func testConnectorCreateUpdate(t *testing.T,
	service, destinationSchema,
	configStep1, configStep2,
	schemaNameJson,
	configJsonStep1,
	configJsonStep2 string,
	bodyCheck1, bodyCheck2 func(*testing.T, map[string]interface{}),
	step1Check, step2Check resource.TestCheckFunc) {
	resourceConfigTemplate := `
	resource "fivetran_connector" "test_connector" {
		provider = fivetran-provider

		group_id           = "group_id"
		service            = "%v"
		run_setup_tests    = true
		trust_fingerprints = true
		trust_certificates = true
		
		destination_schema {
			%v
		}
		
		config {
			%v
		}
	}`

	step1 :=
		resource.TestStep{
			Config: fmt.Sprintf(resourceConfigTemplate, service, destinationSchema, configStep1),
		}
	if step1Check != nil {
		step1.Check = step1Check
	}
	step2 :=
		resource.TestStep{
			Config: fmt.Sprintf(resourceConfigTemplate, service, destinationSchema, configStep2),
		}
	if step2Check != nil {
		step2.Check = step2Check
	}
	var responseData map[string]interface{}
	preCheck := func() {
		tfmock.MockClient().Reset()

		//getHandler =
		tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if responseData == nil {
					return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "NotFound", nil), nil
				}
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
			},
		)

		tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)

		tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
			func(req *http.Request) (*http.Response, error) {

				if bodyCheck1 != nil {
					body := tfmock.RequestBodyToJson(t, req)
					bodyCheck1(t, body)
				}

				responseJson := createConnectorTestResponseJsonMock(
					"connector_id",
					"group_id",
					service,
					schemaNameJson,
					schemaNameJson,
					configJsonStep1,
				)

				responseData = tfmock.CreateMapFromJsonString(t, responseJson)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", responseData), nil
			},
		)

		tfmock.MockClient().When(http.MethodPatch, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {

				if bodyCheck2 != nil {
					body := tfmock.RequestBodyToJson(t, req)
					bodyCheck2(t, body)
				}

				responseJson := createConnectorTestResponseJsonMock(
					"connector_id",
					"group_id",
					service,
					schemaNameJson,
					schemaNameJson,
					configJsonStep2,
				)

				responseData = tfmock.CreateMapFromJsonString(t, responseJson)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
			},
		)
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck:                 preCheck,
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}

func TestConnectorSubFieldsSensitiveMock(t *testing.T) {
	step1 :=
		resource.TestStep{Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id           = "group_id"
			service            = "amplitude"
			run_setup_tests    = true
			trust_fingerprints = true
			trust_certificates = true
		  
			destination_schema {
			  name = "schema_name"
			}
		  
			config {
			  project_credentials {
				project    = "project_name"
				api_key    = "api_key"
				secret_key = "secret_key"
			  }
			}
		  }`,

			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("fivetran_connector.test_connector", "id", "connector_id"),
			),
		}

	var responseData map[string]interface{}

	preCheck := func() {
		tfmock.MockClient().Reset()

		//getHandler =
		tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if responseData == nil {
					return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "NotFound", nil), nil
				}
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
			},
		)

		tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)

		tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				responseJson := createConnectorTestResponseJsonMock(
					"connector_id",
					"group_id",
					"amplitude",
					"schema_name",
					"schema_name",
					`{
						"project_credentials": [
							{
								"project": "project_name",
								"api_key": "******",
								"secret_key": "******"
							}
						]
					}`,
				)

				responseData = tfmock.CreateMapFromJsonString(t, responseJson)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", responseData), nil
			},
		)
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck:                 preCheck,
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestConnectorCollectionSensitiveMock(t *testing.T) {
	step1 :=
		resource.TestStep{Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id           = "group_id"
			service            = "github"
			run_setup_tests    = true
			trust_fingerprints = true
			trust_certificates = true
		  
			destination_schema {
			  name = "schema_name"
			}
		  
			config {
			  pats = ["a", "b"]
			}
		  }`,

			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("fivetran_connector.test_connector", "id", "connector_id"),
			),
		}

	var responseData map[string]interface{}

	preCheck := func() {
		tfmock.MockClient().Reset()

		//getHandler =
		tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if responseData == nil {
					return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "NotFound", nil), nil
				}
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
			},
		)

		tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)

		tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				responseJson := createConnectorTestResponseJsonMock(
					"connector_id",
					"group_id",
					"github",
					"schema_name",
					"schema_name",
					`{
						"pats": ["******", "******"]
					}`,
				)

				responseData = tfmock.CreateMapFromJsonString(t, responseJson)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", responseData), nil
			},
		)
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck:                 preCheck,
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestConnectorNonNullableFieldNotConfiguredMock(t *testing.T) {
	step1 :=
		resource.TestStep{Config: `
	resource "fivetran_connector" "test_connector" {
		provider = fivetran-provider

		group_id           = "group_id"
		service            = "azure_blob_storage"
		run_setup_tests    = true
		trust_fingerprints = true
		trust_certificates = true
	  
		destination_schema {
		  name = "schema_name"
		  table_group_name = "name_of_table_in_snowflake_schema"
		}
	  
		config {
			container_name = "name_of_container"
			pattern = "some_file_pattern"
		}
	  }`,

			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("fivetran_connector.test_connector", "id", "connector_id"),
			),
		}

	var responseData map[string]interface{}

	preCheck := func() {
		tfmock.MockClient().Reset()

		//getHandler =
		tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if responseData == nil {
					return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "NotFound", nil), nil
				}
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
			},
		)

		tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)

		tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				responseJson := createConnectorTestResponseJsonMock(
					"connector_id",
					"group_id",
					"azure_blob_storage",
					"schema_name.name_of_table_in_snowflake_schema",
					"name_of_table_in_snowflake_schema",
					`{
						"container_name": "name_of_container",
						"pattern": "some_file_pattern",
						"json_delivery_mode": "Packed"
					}`,
				)

				responseData = tfmock.CreateMapFromJsonString(t, responseJson)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", responseData), nil
			},
		)
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck:                 preCheck,
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestConnectorConfigCollectionSubFieldsUpdateMock(t *testing.T) {
	testConnectorCreateUpdate(t,
		"google_ads",
		`name = "schema_name"`,
		`
		reports {
			table = "table1"
			metrics = ["metric1", "metric2"]
		}
		reports {
			table = "table2"
			metrics = ["metric2", "metric3"]
		}
		`,
		`
		reports {
			table = "table1"
			report_type = "Custom"
			metrics = ["metric1", "metric2"]
		}
		reports {
			table = "table2"
			metrics = ["metric2", "metric3"]
		}
		`,
		"schema_name",
		`{
			"reports": [
						{
							"table": "table1",
							"report_type": "Custom",
							"metrics": ["metric1", "metric2"]
						},
						{
							"table": "table2",
							"report_type": "Custom",
							"metrics": ["metric2", "metric3"]
						}
					]
		}`,
		`{
			"reports": [
						{
							"table": "table1",
							"report_type": "Custom",
							"metrics": ["metric1", "metric2"]
						},
						{
							"table": "table2",
							"report_type": "Custom",
							"metrics": ["metric2", "metric3"]
						}
					]
		}`,
		func(t *testing.T, body map[string]interface{}) {
			if config, ok := tfmock.AssertKeyExists(t, body, "config").(map[string]interface{}); ok {
				if reports, ok := tfmock.AssertKeyExists(t, config, "reports").([]interface{}); ok {
					tfmock.AssertEqual(t, len(reports), int(2))
				}
			}
		},
		func(t *testing.T, body map[string]interface{}) {
			if config, ok := tfmock.AssertKeyExists(t, body, "config").(map[string]interface{}); ok {
				if reports, ok := tfmock.AssertKeyExists(t, config, "reports").([]interface{}); ok {
					tfmock.AssertEqual(t, len(reports), int(2))
				}
			}
		},
		resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "id", "connector_id"),
		),
		nil,
	)
}

func TestConnectorConfigCollectionSingleFieldObjectsMock(t *testing.T) {
	testConnectorCreateUpdate(t,
		"reddit_ads",
		`name = "schema_name"`,
		`
		accounts_reddit_ads {
			name = "acc1"
		}
		accounts_reddit_ads {
			name = "acc2"
		}
		`,
		`
		accounts_reddit_ads {
			name = "acc2"
		}
		accounts_reddit_ads {
			name = "acc3"
		}
		`,
		"schema_name",
		`{
			"accounts": [
				{
					"name": "acc1"
				},
				{
					"name": "acc2"
				}
			]
		}`,
		`{
			"accounts": [
				{
					"name": "acc2"
				},
				{
					"name": "acc3"
				}
			]
		}`,
		nil, nil,
		resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "id", "connector_id"),
		),
		nil,
	)
}

func TestConnectorConfigemptyStringToNullConversionMock(t *testing.T) {
	testConnectorCreateUpdate(t,
		"google_play",
		`name = "schema_name"`,
		`
		bucket = ""
		`,
		`
		bucket = "bucket_name"
		`,
		"schema_name",
		`{
		}`,
		`{
			"bucket": "bucket_name"
		}`,
		nil, nil,
		resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "id", "connector_id"),
		),
		nil,
	)
}

func TestConnectorConfigEmptyListConversionMock(t *testing.T) {
	testConnectorCreateUpdate(t,
		"dynamodb",
		`name = "schema_name"`,
		`
		packed_mode_tables = []
		`,
		`
		packed_mode_tables = ["table"]
		`,
		"schema_name",
		`{
		}`,
		`{
			"packed_mode_tables": ["table"]
		}`,
		nil, nil,
		resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "id", "connector_id"),
		),
		nil,
	)
}

func TestResourceConnectorWorkdayServiceMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "workday_adaptive"
			data_delay_sensitivity = "NORMAL"
			networking_method = "Directly"
			destination_schema {
			  	name = "workday_adaptive"
			}
			config {
				login = "mail@mail.com"
				password = "password"
				instance_code = "INSTANCE_CODE1"
				reports {
				    table = "table_1"
					version_sync_strategy = "SYNC_SELECT_VERSIONS"
					include_zero_rows = true
					versions = ["2945"]
					accounts  { 
					  include_descendants = true
					  id = "3"
					  flag = true 
					}
					accounts { 
					  id = "4"
					  flag = true 
					}
					dimensions = ["3","4"]
				}
			}
		}
		`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 0)
				tfmock.AssertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "workday_adaptive"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.#", "1"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.0.table", "table_1"),
		),
	}

	step2 := resource.TestStep{
		PreConfig: func() {
			connectorMockUpdatePostHandler.Interactions = 0
			connectorMockUpdateGetHandler.Interactions = 0
			connectorMockUpdatePatchHandler.Interactions = 0
		},
		Config: `
		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "workday_adaptive"
			data_delay_sensitivity = "NORMAL"
			networking_method = "Directly"
			destination_schema {
			  	name = "workday_adaptive"
			}
			config {
				login = "mail@mail.com"
				password = "password"
				instance_code = "INSTANCE_CODE1"
				reports {
				    table = "table_1"
					version_sync_strategy = "SYNC_SELECT_VERSIONS"
					include_zero_rows = true
					versions = ["2945"]
					accounts  { 
					  include_descendants = false
					  id = "3"
					  flag = true 
					}
					accounts { 
					  id = "4"
					  flag = true 
					}
					accounts  { 
					  id = "5"
					  flag = true 
					}
					dimensions = ["3","4"]
				}
				reports {
				    table = "table_2"
					version_sync_strategy = "SYNC_SELECT_VERSIONS"
					include_zero_rows = true
					versions = ["2945"]
					accounts  { 
					  id = "11"
					  flag = true 
					}
					accounts { 
					  id = "12"
					  flag = true 
					}
					dimensions = ["11"]
				}
			}
		}
		`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdatePostHandler.Interactions, 0)
				tfmock.AssertEqual(t, connectorMockUpdateGetHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorMockUpdatePatchHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, connectorMockData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "workday_adaptive"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.reports.#", "2"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceWorkday(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorMockUpdateDelete.Interactions, 1)
				tfmock.AssertEmpty(t, connectorMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}
