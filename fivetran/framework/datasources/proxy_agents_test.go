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
    proxyAgentsMappingResponse = `
    {
		"items": [
    	{
        	"id": "id1",
        	"registred_at": "registred_at1",
        	"region": "region1",
        	"token": "token1",
        	"salt": "salt1",
        	"created_by": "created_by1",
        	"display_name": "display_name1"
    	},
    	{
        	"id": "id2",
        	"registred_at": "registred_at2",
        	"region": "region2",
        	"token": "token2",
        	"salt": "salt2",
        	"created_by": "created_by2",
        	"display_name": "display_name2"
    	}
        ],
        "next_cursor": null
    }`
)

var (
    proxyAgentsDataSourceMockGetHandler *mock.Handler

    proxyAgentsDataSourceMockData map[string]interface{}
)

func setupMockClientProxyAgentsDataSourceConfigMapping(t *testing.T) {
    tfmock.MockClient().Reset()

    proxyAgentsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/proxy").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            proxyAgentsDataSourceMockData = tfmock.CreateMapFromJsonString(t, proxyAgentsMappingResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", proxyAgentsDataSourceMockData), nil
        },
    )
}

func TestDataSourceProxyAgentsConfigMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        data "fivetran_proxy_agents" "test_proxy_agents" {
            provider = fivetran-provider
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, proxyAgentsDataSourceMockGetHandler.Interactions, 1)
                tfmock.AssertNotEmpty(t, proxyAgentsDataSourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.0.id", "id1"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.0.registred_at", "registred_at1"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.0.group_region", "region1"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.0.token", "token1"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.0.salt", "salt1"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.0.created_by", "created_by1"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.0.display_name", "display_name1"),

            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.1.id", "id2"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.1.registred_at", "registred_at2"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.1.group_region", "region2"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.1.token", "token2"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.1.salt", "salt2"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.1.created_by", "created_by2"),
            resource.TestCheckResourceAttr("data.fivetran_proxy_agents.test_proxy_agents", "items.1.display_name", "display_name2"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientProxyAgentsDataSourceConfigMapping(t)
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
