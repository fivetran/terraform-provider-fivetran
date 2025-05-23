package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestDataSourceGroupConnectionsMappingMock(t *testing.T) {
	var getHandler *mock.Handler
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
			data "fivetran_group_connections" "test" {
				provider = fivetran-provider
				id = "group"
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, getHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_group_connections.test", "id", "group"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				getHandler = tfmock.MockClient().When(http.MethodGet, "/v1/groups/group/connections").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						var responseData = tfmock.CreateMapFromJsonString(t, `
    				{
        				"items":[
            			{
							"id": "connection_id",
							"group_id": "group_id",
							"service": "reddit_ads",
							"service_version": 4,
							"schema": "adwords_schema",
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
							"status": {
								"setup_state": "broken",
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
								"warnings": [
									{
										"code": "warning",
										"message": "Warning"
									}
								]
							}
						},
						{
							"id": "connection_id_2",
							"group_id": "group_id",
							"service": "reddit_ads",
							"service_version": 4,
							"schema": "adwords_schema",
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
							"status": {
								"setup_state": "broken",
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
								"warnings": [
									{
										"code": "warning",
										"message": "Warning"
									}
								]
							}
						}
						],
        				"next_cursor": null
    				}`)
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
			},
		},
	)
}
