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
	externalLogsDataSourceMockGetHandler *mock.Handler
	externalLogsDataSourceMockData       map[string]interface{}
)

const (
	externalLogsMappingResponse = `
{
    "items": [
      {
        "id": "log_id",
        "service": "string",
        "enabled": true
      }
    ],
    "next_cursor": null
  }
`
)

func setupMockClientExternalLogsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	externalLogsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/external-logging").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			externalLogsDataSourceMockData = tfmock.CreateMapFromJsonString(t, externalLogsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", externalLogsDataSourceMockData), nil
		},
	)
}

func TestDataSourceExternalLogsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_external_logs" "test_externalLogs" {
			provider = fivetran-provider
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, externalLogsDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, externalLogsDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_external_logs.test_externalLogs", "logs.0.id", "log_id"),
			resource.TestCheckResourceAttr("data.fivetran_external_logs.test_externalLogs", "logs.0.service", "string"),
			resource.TestCheckResourceAttr("data.fivetran_external_logs.test_externalLogs", "logs.0.enabled", "true"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientExternalLogsDataSourceConfigMapping(t)
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
