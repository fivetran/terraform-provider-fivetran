package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	connectionsDataSourceMockGetHandler *mock.Handler
	connectionsDataSourceMockData       map[string]interface{}
)

const (
	connectionsMappingResponse = `
	{
    "items": [
      {
        "id": "connection_id",
        "service": "string",
        "schema": "gsheets.table",
        "paused": false,
        "daily_sync_time": "14:00",
        "succeeded_at": "2024-12-01T15:43:29.013729Z",
        "sync_frequency": 360,
        "group_id": "group_id",
        "connected_by": "user_id",
        "service_version": 0,
        "created_at": "2024-12-01T15:43:29.013729Z",
        "failed_at": "2024-12-01T15:43:29.013729Z",
        "private_link_id": "string",
        "proxy_agent_id": "string",
        "networking_method": "Directly",
        "pause_after_trial": false,
        "data_delay_threshold": 0,
        "data_delay_sensitivity": "LOW",
        "schedule_type": "auto",
        "hybrid_deployment_agent_id": "string"
      }
    ],
    "next_cursor": null
  }
`

	connectionsMappingResponseWithCursor = `
	{
    "items": [
      {
        "id": "connection_id1",
        "service": "string",
        "schema": "gsheets.table",
        "paused": false,
        "daily_sync_time": "14:00",
        "succeeded_at": "2024-12-01T15:43:29.013729Z",
        "sync_frequency": 360,
        "group_id": "group_id",
        "connected_by": "user_id",
        "service_version": 0,
        "created_at": "2024-12-01T15:43:29.013729Z",
        "failed_at": "2024-12-01T15:43:29.013729Z",
        "private_link_id": "string",
        "proxy_agent_id": "string",
        "networking_method": "Directly",
        "pause_after_trial": false,
        "data_delay_threshold": 0,
        "data_delay_sensitivity": "LOW",
        "schedule_type": "auto",
        "hybrid_deployment_agent_id": "string"
      }
    ],
    "next_cursor": "next_cursor"
  }
`

)

func setupMockClientConnectionsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	connectionsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if req.URL.Query().Get("cursor") == "next_cursor" {
				connectionsDataSourceMockData = tfmock.CreateMapFromJsonString(t, connectionsMappingResponse)
			} else {
				connectionsDataSourceMockData = tfmock.CreateMapFromJsonString(t, connectionsMappingResponseWithCursor)
			}

			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectionsDataSourceMockData), nil
		},
	)
}

func setupMockClientConnectionsDataSourceFilteringByGroupIdAndSchema(t *testing.T) {
	tfmock.MockClient().Reset()

	connectionsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if req.URL.Query().Get("group_id") == "group_id" && req.URL.Query().Get("schema") == "gsheets.table" {
				connectionsDataSourceMockData = tfmock.CreateMapFromJsonString(t, connectionsMappingResponse)
			} else {
				connectionsDataSourceMockData = nil
			}

			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectionsDataSourceMockData), nil
		},
	)
}

func TestDataSourceConnectionsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_connections" "test" {
			provider = fivetran-provider
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectionsDataSourceMockGetHandler.Interactions, 2)
				tfmock.AssertNotEmpty(t, connectionsDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.id", "connection_id"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.name", "gsheets.table"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.group_id", "group_id"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.connected_by", "user_id"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.service", "string"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.service_version", "0"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.sync_frequency", "360"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.schedule_type", "auto"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.paused", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.pause_after_trial", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.daily_sync_time", "14:00"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.data_delay_sensitivity", "LOW"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.data_delay_threshold", "0"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.proxy_agent_id", "string"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.networking_method", "Directly"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.hybrid_deployment_agent_id", "string"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.private_link_id", "string"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.succeeded_at", "2024-12-01 15:43:29.013729 +0000 UTC"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.failed_at", "2024-12-01 15:43:29.013729 +0000 UTC"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test", "connections.0.created_at", "2024-12-01 15:43:29.013729 +0000 UTC"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectionsDataSourceConfigMapping(t)
			},
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

func TestDataSourceConnectionsFilteringByGroupIdAndSchema(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_connections" "test2" {
			provider = fivetran-provider
			group_id = "group_id"
			schema_name = "gsheets.table"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectionsDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, connectionsDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.id", "connection_id"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.name", "gsheets.table"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.group_id", "group_id"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.connected_by", "user_id"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.service", "string"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.service_version", "0"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.sync_frequency", "360"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.schedule_type", "auto"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.paused", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.pause_after_trial", "false"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.daily_sync_time", "14:00"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.data_delay_sensitivity", "LOW"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.data_delay_threshold", "0"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.proxy_agent_id", "string"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.networking_method", "Directly"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.hybrid_deployment_agent_id", "string"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.private_link_id", "string"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.succeeded_at", "2024-12-01 15:43:29.013729 +0000 UTC"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.failed_at", "2024-12-01 15:43:29.013729 +0000 UTC"),
			resource.TestCheckResourceAttr("data.fivetran_connections.test2", "connections.0.created_at", "2024-12-01 15:43:29.013729 +0000 UTC"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectionsDataSourceFilteringByGroupIdAndSchema(t)
			},
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
