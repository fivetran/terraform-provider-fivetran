package resources_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// postgresConnectionTfConfig returns a postgres connection. config is the JSON-encoded config.
func postgresConnectionTfConfig(config string) string {
	return fmt.Sprintf(`
		resource "fivetran_connection" "test_connection" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "postgres"

			destination_schema {
				prefix = "postgres"
			}

			trust_certificates = false
			trust_fingerprints = false
			run_setup_tests = false

			config = jsonencode(%s)
		}`, config)
}

// Returned by the API for postgres without database. The sensitive password is masked, as the API does.
const postgresConnectionResponseWithoutDatabase = `
{
	"id": "connection_id",
	"group_id": "group_id",
	"service": "postgres",
	"service_version": 1,
	"schema": "postgres",
	"paused": true,
	"pause_after_trial": true,
	"connected_by": "user_id",
	"created_at": "2022-01-01T11:22:33.012345Z",
	"sync_frequency": 360,
	"schedule_type": "auto",
	"networking_method": "Directly",
	"data_delay_sensitivity": "NORMAL",
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

// The config as stored in state: jsonencode sorts the keys.
const postgresConnectionConfigJson = `{"database":"database","host":"host","password":"password","port":5432,"user":"user"}`

// Create: a configured config field (database) is missing from the create and GET responses.
// fivetran_connection does not read config back from the API, so the planned config is stored as is:
// apply succeeds, and refresh keeps the value without showing drift.
func TestConnectionConfigUnconfirmedOnCreateMock(t *testing.T) {
	var testConnectionData map[string]interface{}
	var connectionPostHandler *mock.Handler
	var connectionDeleteHandler *mock.Handler

	tfConfig := postgresConnectionTfConfig(`{
		host = "host"
		port = 5432
		user = "user"
		password = "password"
		database = "database"
	}`)

	step1 := resource.TestStep{
		Config: tfConfig,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectionPostHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connection.test_connection", "config", postgresConnectionConfigJson),
		),
	}

	// Refresh only: the API still omits database, and the planned config is still kept
	step2 := resource.TestStep{
		RefreshState: true,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_connection.test_connection", "config", postgresConnectionConfigJson),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				tfmock.MockClient().When(http.MethodGet, "/v1/connections/connection_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testConnectionData), nil
					},
				)

				connectionPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "database", "database")

						// database is accepted but not returned
						testConnectionData = tfmock.CreateMapFromJsonString(t, postgresConnectionResponseWithoutDatabase)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", testConnectionData), nil
					},
				)

				connectionDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connection_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testConnectionData = nil
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectionDeleteHandler.Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{step1, step2},
		},
	)
}
