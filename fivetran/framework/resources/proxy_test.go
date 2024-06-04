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
	proxyPostHandler   *mock.Handler
	proxyGethHandler   *mock.Handler
	proxyDeleteHandler *mock.Handler
	proxyData map[string]interface{}
	proxyDetailsData map[string]interface{}
)

func setupMockClientProxyResource(t *testing.T) {
	tfmock.MockClient().Reset()
	proxyResponse :=
	`{
        "agent_id": "agent_id",
        "auth_token": "auth_token",
        "proxy_server_uri": "proxy_server_uri"
	}`

	proxyDetailsResponse :=
	`{
    		"id": "agent_id",
    		"account_id": "account_id",
    		"registred_at": "registred_at",
    		"region": "group_region",
    		"token": "auth_token",
    		"salt": "salt",
    		"created_by": "created_by",
    		"display_name": "display_name"
  	}`

	proxyPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/proxy").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			proxyData = tfmock.CreateMapFromJsonString(t, proxyResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Proxy has been created", proxyData), nil
		},
	)

	proxyGethHandler = tfmock.MockClient().When(http.MethodGet, "/v1/proxy/agent_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			proxyDetailsData = tfmock.CreateMapFromJsonString(t, proxyDetailsResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", proxyDetailsData), nil
		},
	)

	proxyDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/proxy/agent_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "Proxy has been deleted", nil), nil
		},
	)
}

func TestResourceProxyMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_proxy" "test_proxy" {
                 provider = fivetran-provider

                 display_name = "display_name"
                 group_region = "group_region"
            }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, proxyPostHandler.Interactions, 1)
				tfmock.AssertEqual(t, proxyGethHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_proxy.test_proxy", "display_name", "display_name"),
			resource.TestCheckResourceAttr("fivetran_proxy.test_proxy", "group_region", "group_region"),
			resource.TestCheckResourceAttr("fivetran_proxy.test_proxy", "token", "auth_token"),
			resource.TestCheckResourceAttr("fivetran_proxy.test_proxy", "proxy_server_uri", "proxy_server_uri"),
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
				tfmock.AssertEqual(t, proxyDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
