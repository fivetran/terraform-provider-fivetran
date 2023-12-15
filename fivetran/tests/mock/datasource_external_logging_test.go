package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	externalLoggingDataSourceMockGetHandler *mock.Handler

	externalLoggingDataSourceMockData map[string]interface{}
)

func setupMockClientExternalLoggingDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	externalLoggingDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/external-logging/log_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			externalLoggingDataSourceMockData = createMapFromJsonString(t, externalLoggingMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", externalLoggingDataSourceMockData), nil
		},
	)
}

func TestDataSourceExternalLoggingConfigMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_external_logging" "test_extlog" {
			provider = fivetran-provider
			id = "log_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, externalLoggingDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, externalLoggingDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "service", "azure_monitor_log"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "enabled", "false"),

			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.workspace_id", "workspace_id"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.primary_key", "primary_key"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.log_group_name", "log_group_name"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.sub_domain", "sub_domain"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.enable_ssl", "true"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.port", "443"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.role_arn", "role_arn"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.channel", "channel"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.token", "token"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.region", "region"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.external_id", "external_id"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.api_key", "api_key"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.host", "host"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.0.hostname", "hostname"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientExternalLoggingDataSourceConfigMapping(t)
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
