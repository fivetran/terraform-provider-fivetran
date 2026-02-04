package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	externalLoggingMappingResponse = `
    {
        "id":                 "log_id",
        "group_id":           "group_id",
        "service":            "azure_monitor_log",
        "enabled":            false,
        "config":{
            "workspace_id":   	   "workspace_id",
            "primary_key":    	   "primary_key",
            "role_arn":            "role_arn",
            "region":              "region",
            "log_group_name":      "log_group_name",
            "sub_domain":          "sub_domain",
            "enable_ssl":          true,
            "channel":             "channel",
            "token":               "token",
            "external_id":         "external_id",
            "api_key":             "api_key",
            "host":                "host",
            "hostname":            "hostname",
            "port":                443,
            "project_id":          "project_id",
			"access_key_secret":   "access_key_secret",
			"access_key_id":       "access_key_id",
			"service_account_key": "service_account_key"
        }
    }
    `
)

var (
	externalLoggingDataSourceMockGetHandler *mock.Handler

	externalLoggingDataSourceMockData map[string]interface{}
)

func setupMockClientExternalLoggingDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	externalLoggingDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/external-logging/log_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			externalLoggingDataSourceMockData = tfmock.CreateMapFromJsonString(t, externalLoggingMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", externalLoggingDataSourceMockData), nil
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
				tfmock.AssertEqual(t, externalLoggingDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, externalLoggingDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "service", "azure_monitor_log"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "enabled", "false"),

			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.workspace_id", "workspace_id"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.primary_key", "primary_key"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.log_group_name", "log_group_name"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.sub_domain", "sub_domain"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.enable_ssl", "true"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.port", "443"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.role_arn", "role_arn"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.channel", "channel"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.token", "token"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.region", "region"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.external_id", "external_id"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.api_key", "api_key"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.host", "host"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.hostname", "hostname"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.project_id", "project_id"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.access_key_secret", "access_key_secret"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.access_key_id", "access_key_id"),
			resource.TestCheckResourceAttr("data.fivetran_external_logging.test_extlog", "config.service_account_key", "service_account_key"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientExternalLoggingDataSourceConfigMapping(t)
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
