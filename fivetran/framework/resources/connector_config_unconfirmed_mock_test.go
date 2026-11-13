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

// postgresConnectorTfConfig returns a postgres connector. extra lines are added to the config block.
func postgresConnectorTfConfig(extra string) string {
	return fmt.Sprintf(`
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
				host = "host"
				port = 5432
				user = "user"
				password = "password"
				%s
			}
		}`, extra)
}

// Returned by the API for postgres without database and tunnel_host. The sensitive password is masked, as the API does.
const postgresConnectorResponseWithoutDatabase = `
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
		"tasks": [],
		"warnings": []
	},
	"config": {
		"host": "host",
		"port": 5432,
		"user": "user",
		"password": "******"
	}
}
`

// Create: a configured config subfield (database) is missing from the successful create response.
// Known issue: the connector reads config back through getValue, which resolves the omitted field to
// null, so apply fails. fivetran_destination keeps the planned value and warns instead. When the same fix
// is applied to fivetran_connector, this test must expect a successful apply with database kept in state.
func TestConnectorConfigUnconfirmedOnCreateMock(t *testing.T) {
	var testConnectorData map[string]interface{}
	var connectorPostHandler *mock.Handler
	var connectorDeleteHandler *mock.Handler

	step1 := resource.TestStep{
		Config:      postgresConnectorTfConfig(`database = "database"`),
		ExpectError: regexp.MustCompile(`(?s)Provider produced inconsistent result after apply.*\.config\.database: was\s+cty\.StringVal\("database"\), but\s+now null`),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testConnectorData), nil
					},
				)

				connectorPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "database", "database")

						// database is accepted but not returned
						testConnectorData = tfmock.CreateMapFromJsonString(t, postgresConnectorResponseWithoutDatabase)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", testConnectorData), nil
					},
				)

				connectorDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testConnectorData = nil
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorPostHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorDeleteHandler.Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{step1},
		},
	)
}

// Update: a configured config subfield (tunnel_host) added in an update is missing from the PATCH response.
// Known issue: as on create, the omitted field is resolved to null and apply fails.
func TestConnectorConfigUnconfirmedOnUpdateMock(t *testing.T) {
	var testConnectorData map[string]interface{}
	var connectorPatchHandler *mock.Handler
	var connectorDeleteHandler *mock.Handler

	step1 := resource.TestStep{
		Config: postgresConnectorTfConfig(""),
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "config.tunnel_host"),
		),
	}

	step2 := resource.TestStep{
		Config:      postgresConnectorTfConfig(`tunnel_host = "tunnel_host"`),
		ExpectError: regexp.MustCompile(`(?s)Provider produced inconsistent result after apply.*\.config\.tunnel_host: was\s+cty\.StringVal\("tunnel_host"\),\s+but now null`),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testConnectorData), nil
					},
				)

				tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testConnectorData = tfmock.CreateMapFromJsonString(t, postgresConnectorResponseWithoutDatabase)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", testConnectorData), nil
					},
				)

				connectorPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/connections/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "tunnel_host", "tunnel_host")

						// tunnel_host is accepted but not returned
						testConnectorData = tfmock.CreateMapFromJsonString(t, postgresConnectorResponseWithoutDatabase)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testConnectorData), nil
					},
				)

				connectorDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testConnectorData = nil
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorPatchHandler.Interactions, 1)
				tfmock.AssertEqual(t, connectorDeleteHandler.Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{step1, step2},
		},
	)
}
