package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	connectorDataSourceMockGetHandler *mock.Handler

	connectorDataSourceMockData map[string]interface{}
)

func setupMockClientConnectorDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	connectorDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorDataSourceMockData = createMapFromJsonString(t, connectorMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorDataSourceMockData), nil
		},
	)
}

func TestDataSourceConnectorConfigMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_connector" "test_connector" {
			provider = fivetran-provider
			id = "connector_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, connectorDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "service", "google_sheets"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "service_version", "1"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "schedule_type", "auto"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.is_historical_sync", "true"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.update_state", "on_schedule"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.setup_state", "incomplete"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.sync_state", "paused"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.tasks.0.code", "task_code"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.tasks.0.message", "task_message"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.warnings.0.code", "warning_code"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "status.warnings.0.message", "warning_message"),

			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "name", "schema.table"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "sync_frequency", "5"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "paused", "true"),
			resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "pause_after_trial", "true"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorDataSourceConfigMapping(t)
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
