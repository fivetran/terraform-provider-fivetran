package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestDataSourceConnectorMock(t *testing.T) {
	var getHandler *mock.Handler
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_connector" "test_connector" {
			provider = fivetran-provider
			
			id = "connector_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, getHandler.Interactions, 2)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "id", "connector_id"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.user", "user_name"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.password", "******"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.port", "5432"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.always_encrypted", "true"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.account_ids.0", "id1"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.account_ids.1", "id2"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.account_ids.2", "id3"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.reports.0.report_type", "report_1"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.reports.0.metrics.0", "metric1"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.reports.0.metrics.1", "metric2"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.reports.1.report_type", "report_2"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.reports.1.metrics.0", "metric2"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "config.reports.1.metrics.1", "metric3"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				getHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						var responseData = tfmock.CreateMapFromJsonString(t, `
						{
							"id": "connector_id",
							"group_id": "group_id",
							"service": "google_ads",
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
							},
							"config": {
								"user": "user_name",
								"password": "******",
								"port": 5432,
								"always_encrypted": true,
								"account_ids": ["id1", "id2", "id3"],
								"reports": [
									{
										"report_type": "report_1",
										"metrics": ["metric1", "metric2"]
									},
									{
										"report_type": "report_2",
										"metrics": ["metric2", "metric3"]
									}
								]
							}
						}
						`)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
					},
				)
			},
			//Providers: testProviders,
			//ProtoV5ProviderFactories: tfmock.ProtoV5ProviderFactory,
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
