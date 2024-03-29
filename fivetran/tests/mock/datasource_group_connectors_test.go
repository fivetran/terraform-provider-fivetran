package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	groupConnectorsDataSourceMockGetHandler *mock.Handler
	groupConnectorsDataSourceMockData       map[string]interface{}
)

const (
	groupConnectorsMappingResponse = `
	{
        "items": [
            {
                "id": "iodize_impressive",
                "group_id": "group_id",
                "service": "salesforce",
                "service_version": 1,
                "schema": "salesforce",
                "connected_by": "concerning_batch",
                "created_at": "2018-07-21T22:55:21.724201Z",
                "succeeded_at": "2018-12-26T17:58:18.245Z",
                "failed_at": "2018-08-24T15:24:58.872491Z",
                "sync_frequency": 60,
                "status": {
                    "setup_state": "connected",
                    "sync_state": "paused",
                    "update_state": "delayed",
                    "is_historical_sync": false,
                    "tasks": [],
                    "warnings": []
                }
            }
        ],
        "next_cursor": null
    }
	`
)

func setupMockClientGroupConnectorsDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	groupConnectorsDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/groups/group_id/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			groupConnectorsDataSourceMockData = createMapFromJsonString(t, groupConnectorsMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", groupConnectorsDataSourceMockData), nil
		},
	)
}

func TestDataSourceGroupConnectorsMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_group_connectors" "test_group_connectors" {
			provider = fivetran-provider
			id = "group_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, groupConnectorsDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, groupConnectorsDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.id", "iodize_impressive"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.group_id", "group_id"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.service", "salesforce"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.service_version", "1"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.schema", "salesforce"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.connected_by", "concerning_batch"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.created_at", "2018-07-21 22:55:21.724201 +0000 UTC"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.succeeded_at", "2018-12-26 17:58:18.245 +0000 UTC"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.failed_at", "2018-08-24 15:24:58.872491 +0000 UTC"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.sync_frequency", "60"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.status.0.setup_state", "connected"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.status.0.sync_state", "paused"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.status.0.update_state", "delayed"),
			resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.0.status.0.is_historical_sync", "false"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientGroupConnectorsDataSourceConfigMapping(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
