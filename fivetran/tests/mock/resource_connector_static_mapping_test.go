package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	connectorMockGetHandler  *mock.Handler
	connectorMockPostHandler *mock.Handler
	connectorMockDelete      *mock.Handler
	connectorMappingMockData map[string]interface{}
)

const (
	connectorConfigMappingTfConfig = `
	resource "fivetran_connector" "test_connector" {
		provider = fivetran-provider

		group_id = "group_id"
		service = "google_sheets"

		destination_schema {
			name = "schema"
			table = "table"
		}

		trust_certificates = false
		trust_fingerprints = false
		run_setup_tests = false

		auth {
			refresh_token = "refresh_token"
			access_token = "access_token"
			realm_id = "realm_id"
			client_access {
				client_id = "client_id"
				client_secret = "client_secret"
				user_agent = "user_agent"
				developer_token = "developer_token"
			}
		}
	}
	`

	connectorMappingResponse = `
	{
		"id": "connector_id",
        "group_id": "group_id",
        "service": "google_sheets",
        "service_version": 1,
        "schema": "schema.table",
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
        }]
	}
	`
)

func setupMockClientConnectorResourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	connectorMockGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMappingMockData), nil
		},
	)

	connectorMockPostHandler = mockClient.When(http.MethodPost, "/v1/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)

			assertKeyExists(t, body, "auth")

			auth := body["auth"].(map[string]interface{})

			assertKeyExistsAndHasValue(t, auth, "refresh_token", "refresh_token")
			assertKeyExistsAndHasValue(t, auth, "access_token", "access_token")
			assertKeyExistsAndHasValue(t, auth, "realm_id", "realm_id")

			assertKeyExists(t, auth, "client_access")

			clientAccess := auth["client_access"].(map[string]interface{})
			assertKeyExistsAndHasValue(t, clientAccess, "client_id", "client_id")
			assertKeyExistsAndHasValue(t, clientAccess, "client_secret", "client_secret")
			assertKeyExistsAndHasValue(t, clientAccess, "user_agent", "user_agent")
			assertKeyExistsAndHasValue(t, clientAccess, "developer_token", "developer_token")

			connectorMappingMockData = createMapFromJsonString(t, connectorMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorMappingMockData), nil
		},
	)

	connectorMockDelete = mockClient.When(http.MethodDelete, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorMappingMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMappingMockData), nil
		},
	)
}

func TestResourceConnectorConfigMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: connectorConfigMappingTfConfig,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorMockPostHandler.Interactions, 1)
				assertEqual(t, connectorMockGetHandler.Interactions, 1)
				assertNotEmpty(t, connectorMappingMockData)
				return nil
			},

			// check auth fields
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.0.refresh_token", "refresh_token"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.0.access_token", "access_token"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.0.realm_id", "realm_id"),

			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.0.client_access.0.client_id", "client_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.0.client_access.0.client_secret", "client_secret"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.0.client_access.0.user_agent", "user_agent"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.0.client_access.0.developer_token", "developer_token"),

			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "google_sheets"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "schema.table"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "false"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "false"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "false"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceConfigMapping(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, connectorMockDelete.Interactions, 1)
				assertEmpty(t, connectorMappingMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
