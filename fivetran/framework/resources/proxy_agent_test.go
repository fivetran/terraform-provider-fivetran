package resources_test

import (
	"net/http"
	"testing"
	
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	proxyAgentPostHandler   *mock.Handler
	proxyAgentGetHandler   *mock.Handler
	proxyAgentDeleteHandler *mock.Handler
	proxyAgentData map[string]interface{}
	proxyAgentDetailsData map[string]interface{}
)

func setupMockClientProxyResource(t *testing.T) {
	tfmock.MockClient().Reset()
	proxyAgentResponse :=
	`{
    		"client_cert": "client_cert",
    		"agent_id": "agent_id",
    		"auth_token": "auth_token",
    		"client_private_key": "client_private_key"
  	}`

	proxyAgentDetailsResponse :=
	`{
    		"id": "agent_id",
    		"registered_at": "registered_at",
    		"region": "group_region",
    		"created_by": "created_by",
    		"display_name": "display_name"
  	}`

	proxyAgentPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/proxy").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			proxyAgentData = tfmock.CreateMapFromJsonString(t, proxyAgentResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Proxy has been created", proxyAgentData), nil
		},
	)

	proxyAgentGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/proxy/agent_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			proxyAgentDetailsData = tfmock.CreateMapFromJsonString(t, proxyAgentDetailsResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", proxyAgentDetailsData), nil
		},
	)

	proxyAgentDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/proxy/agent_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "Proxy has been deleted", nil), nil
		},
	)
}

func TestResourceProxyMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_proxy_agent" "test_proxy_agent" {
                 provider = fivetran-provider

                 display_name = "display_name"
                 group_region = "group_region"
            }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, proxyAgentPostHandler.Interactions, 1)
				tfmock.AssertEqual(t, proxyAgentGetHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_proxy_agent.test_proxy_agent", "display_name", "display_name"),
			resource.TestCheckResourceAttr("fivetran_proxy_agent.test_proxy_agent", "group_region", "group_region"),
			resource.TestCheckResourceAttr("fivetran_proxy_agent.test_proxy_agent", "token", "auth_token"),
			resource.TestCheckResourceAttr("fivetran_proxy_agent.test_proxy_agent", "client_private_key", "client_private_key"),
			resource.TestCheckResourceAttr("fivetran_proxy_agent.test_proxy_agent", "client_cert", "client_cert"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientProxyResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, proxyAgentDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
