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

// postgresConnectionV2TfConfig returns a postgres connection_v2. extra lines are added to the config object.
func postgresConnectionV2TfConfig(extra string) string {
	return fmt.Sprintf(`
		resource "fivetran_connection_v2" "test_connection" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "postgres"
			destination_schema_names = "FIVETRAN_NAMING"

			trust_certificates = false
			trust_fingerprints = false
			run_setup_tests = false

			config = {
				schema_prefix = "postgres"
				host = "host"
				port = 5432
				user = "user"
				password = "password"
				%s
			}
		}`, extra)
}

// Minimal connector metadata for postgres: password is a secret, the other fields are plain.
const postgresConnectionV2Metadata = `
{
	"id": "postgres",
	"name": "PostgreSQL",
	"config": {
		"type": "object",
		"properties": {
			"schema_prefix": {"type": "string", "immutable": true},
			"host": {"type": "string"},
			"port": {"type": "integer"},
			"user": {"type": "string"},
			"password": {"type": "string", "format": "password"},
			"database": {"type": "string"},
			"tunnel_host": {"type": "string"}
		}
	},
	"auth": {
		"type": "object",
		"properties": {}
	}
}
`

// postgresConnectionV2Response returns the API response for postgres. extra lines are added to the config object.
// database and tunnel_host are not returned unless they are in extra. The sensitive password is masked, as the API does.
func postgresConnectionV2Response(extra string) string {
	return fmt.Sprintf(`
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
	"destination_schema_names": "FIVETRAN_NAMING",
	"status": {
		"setup_state": "incomplete",
		"sync_state": "paused",
		"update_state": "on_schedule",
		"is_historical_sync": true,
		"tasks": [],
		"warnings": []
	},
	"config": {
		"schema_prefix": "postgres",
		"host": "host",
		"port": 5432,
		"user": "user",
		%s
		"password": "******"
	}
}
`, extra)
}

// setupConnectionV2UnconfirmedMock registers the metadata, GET and DELETE handlers shared by the tests below.
// data is the connection returned by GET; POST and PATCH handlers are registered by each test.
func setupConnectionV2UnconfirmedMock(t *testing.T, data *map[string]interface{}) *mock.Handler {
	tfmock.MockClient().Reset()

	tfmock.MockClient().When(http.MethodGet, "/v1/metadata/connector-types/postgres").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success",
				tfmock.CreateMapFromJsonString(t, postgresConnectionV2Metadata)), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/connections/connection_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", *data), nil
		},
	)

	return tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connection_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			*data = nil
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
		},
	)
}

// Create: a configured config field (database) is missing from the create and GET responses.
// The projection keeps the configured value for a field the response omits, so apply succeeds and
// refresh keeps the value too. The next plan is empty, so a value the API ignored is never shown as drift.
func TestConnectionV2ConfigUnconfirmedOnCreateMock(t *testing.T) {
	var testConnectionData map[string]interface{}
	var connectionPostHandler *mock.Handler
	var connectionDeleteHandler *mock.Handler

	step1 := resource.TestStep{
		Config: postgresConnectionV2TfConfig(`database = "database"`),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectionPostHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connection_v2.test_connection", "config.database", "database"),
			resource.TestCheckResourceAttr("fivetran_connection_v2.test_connection", "config.password", "password"),
			resource.TestCheckResourceAttr("fivetran_connection_v2.test_connection", "config.host", "host"),
		),
	}

	// Refresh only: the API still omits database, and the configured value is still kept
	step2 := resource.TestStep{
		RefreshState: true,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_connection_v2.test_connection", "config.database", "database"),
			resource.TestCheckResourceAttr("fivetran_connection_v2.test_connection", "config.password", "password"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				connectionDeleteHandler = setupConnectionV2UnconfirmedMock(t, &testConnectionData)

				connectionPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "database", "database")

						// database is accepted but not returned
						testConnectionData = tfmock.CreateMapFromJsonString(t, postgresConnectionV2Response(""))
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", testConnectionData), nil
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

// Create: a configured config field (database) is returned as null.
// Known issue: the projection stores the null as an untyped dynamic value, so apply fails. A fix must
// either keep the configured value or store a typed null; then this test must expect a successful apply.
func TestConnectionV2ConfigReturnedAsNullOnCreateMock(t *testing.T) {
	var testConnectionData map[string]interface{}
	var connectionDeleteHandler *mock.Handler

	step1 := resource.TestStep{
		Config:      postgresConnectionV2TfConfig(`database = "database"`),
		ExpectError: regexp.MustCompile(`(?s)Provider produced inconsistent result after apply.*wrong final value type: attribute\s+"database":\s+string required, but have dynamic`),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				connectionDeleteHandler = setupConnectionV2UnconfirmedMock(t, &testConnectionData)

				tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testConnectionData = tfmock.CreateMapFromJsonString(t, postgresConnectionV2Response(`"database": null,`))
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", testConnectionData), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectionDeleteHandler.Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{step1},
		},
	)
}

// Update: a configured config field (tunnel_host) added in an update is missing from the PATCH and GET responses.
// As on create, the configured value is kept after apply and after refresh.
func TestConnectionV2ConfigUnconfirmedOnUpdateMock(t *testing.T) {
	var testConnectionData map[string]interface{}
	var connectionPatchHandler *mock.Handler
	var connectionDeleteHandler *mock.Handler

	step1 := resource.TestStep{
		Config: postgresConnectionV2TfConfig(""),
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckNoResourceAttr("fivetran_connection_v2.test_connection", "config.tunnel_host"),
		),
	}

	step2 := resource.TestStep{
		Config: postgresConnectionV2TfConfig(`tunnel_host = "tunnel_host"`),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectionPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connection_v2.test_connection", "config.tunnel_host", "tunnel_host"),
		),
	}

	// Refresh only: the API still omits tunnel_host, and the configured value is still kept
	step3 := resource.TestStep{
		RefreshState: true,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_connection_v2.test_connection", "config.tunnel_host", "tunnel_host"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				connectionDeleteHandler = setupConnectionV2UnconfirmedMock(t, &testConnectionData)

				tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testConnectionData = tfmock.CreateMapFromJsonString(t, postgresConnectionV2Response(""))
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", testConnectionData), nil
					},
				)

				connectionPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/connections/connection_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "tunnel_host", "tunnel_host")

						// tunnel_host is accepted but not returned
						testConnectionData = tfmock.CreateMapFromJsonString(t, postgresConnectionV2Response(""))
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testConnectionData), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectionDeleteHandler.Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{step1, step2, step3},
		},
	)
}
