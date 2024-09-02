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
	proxyAgentMappingResponse = `
    {
        "id": "id",
        "account_id": "account_id",
        "registered_at": "registered_at",
        "region": "region",
        "token": "token",
        "salt": "salt",
        "created_by": "created_by",
        "display_name": "display_name"
    }
    `
)

var (
	proxyAgentDataSourceMockGetHandler *mock.Handler

	proxyAgentDataSourceMockData map[string]interface{}
)

func setupMockClientProxyAgentDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	proxyAgentDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/proxy/id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			proxyAgentDataSourceMockData = tfmock.CreateMapFromJsonString(t, proxyAgentMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", proxyAgentDataSourceMockData), nil
		},
	)
}

func TestDataSourceProxyAgentConfigMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_proxy_agent" "test_proxy_agent" {
            provider = fivetran-provider
            id = "id"
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, proxyAgentDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, proxyAgentDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_proxy_agent.test_proxy_agent", "id", "id"),
			resource.TestCheckResourceAttr("data.fivetran_proxy_agent.test_proxy_agent", "registred_at", "registered_at"),
			resource.TestCheckResourceAttr("data.fivetran_proxy_agent.test_proxy_agent", "group_region", "region"),
			resource.TestCheckResourceAttr("data.fivetran_proxy_agent.test_proxy_agent", "token", "token"),
			resource.TestCheckResourceAttr("data.fivetran_proxy_agent.test_proxy_agent", "salt", "salt"),
			resource.TestCheckResourceAttr("data.fivetran_proxy_agent.test_proxy_agent", "created_by", "created_by"),
			resource.TestCheckResourceAttr("data.fivetran_proxy_agent.test_proxy_agent", "display_name", "display_name"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientProxyAgentDataSourceConfigMapping(t)
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
