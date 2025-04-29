package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

		config {

		}

		auth {
			refresh_token = "refresh_token"
			access_token = "access_token"
			previous_refresh_token = "previous_refresh_token"
			realm_id = "realm_id"
			user_access_token = "user_access_token"
			consumer_secret = "consumer_secret"
			consumer_key = "consumer_key"
			oauth_token = "oauth_token"
			oauth_token_secret = "oauth_token_secret"
			role_arn = "role_arn"
			aws_access_key = "aws_access_key"
			aws_secret_key = "aws_secret_key"
			client_id = "client_id"
			key_id = "key_id"
			team_id = "team_id"
			client_secret = "client_secret"

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
		"networking_method": "Directly",
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

	connectorMockGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorMappingMockData), nil
		},
	)

	connectorMockPostHandler = mockClient.When(http.MethodPost, "/v1/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)

			assertKeyExists(t, body, "auth")

			auth := body["auth"].(map[string]interface{})

			assertKeyExistsAndHasValue(t, auth, "refresh_token", "refresh_token")
			assertKeyExistsAndHasValue(t, auth, "access_token", "access_token")
			assertKeyExistsAndHasValue(t, auth, "realm_id", "realm_id")
			assertKeyExistsAndHasValue(t, auth, "previous_refresh_token", "previous_refresh_token")
			assertKeyExistsAndHasValue(t, auth, "user_access_token", "user_access_token")
			assertKeyExistsAndHasValue(t, auth, "consumer_secret", "consumer_secret")
			assertKeyExistsAndHasValue(t, auth, "consumer_key", "consumer_key")
			assertKeyExistsAndHasValue(t, auth, "oauth_token", "oauth_token")
			assertKeyExistsAndHasValue(t, auth, "oauth_token_secret", "oauth_token_secret")
			assertKeyExistsAndHasValue(t, auth, "role_arn", "role_arn")
			assertKeyExistsAndHasValue(t, auth, "aws_access_key", "aws_access_key")
			assertKeyExistsAndHasValue(t, auth, "aws_secret_key", "aws_secret_key")
			assertKeyExistsAndHasValue(t, auth, "client_id", "client_id")
			assertKeyExistsAndHasValue(t, auth, "key_id", "key_id")
			assertKeyExistsAndHasValue(t, auth, "team_id", "team_id")
			assertKeyExistsAndHasValue(t, auth, "client_secret", "client_secret")

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

	connectorMockDelete = mockClient.When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
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
				//assertEqual(t, connectorMockGetHandler.Interactions, 1)
				assertNotEmpty(t, connectorMappingMockData)
				return nil
			},

			// check auth fields
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.refresh_token", "refresh_token"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.access_token", "access_token"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.realm_id", "realm_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.previous_refresh_token", "previous_refresh_token"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.user_access_token", "user_access_token"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.consumer_secret", "consumer_secret"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.consumer_key", "consumer_key"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.oauth_token", "oauth_token"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.oauth_token_secret", "oauth_token_secret"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.role_arn", "role_arn"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.aws_access_key", "aws_access_key"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.aws_secret_key", "aws_secret_key"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.client_id", "client_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.key_id", "key_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.team_id", "team_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.client_secret", "client_secret"),

			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.client_access.client_id", "client_id"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.client_access.client_secret", "client_secret"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.client_access.user_agent", "user_agent"),
			resource.TestCheckResourceAttr("fivetran_connector.test_connector", "auth.client_access.developer_token", "developer_token"),

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
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
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
