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
    proxyMappingResponse = `
    {
        "id": "id",
        "account_id": "account_id",
        "registred_at": "registred_at",
        "region": "region",
        "token": "token",
        "salt": "salt",
        "created_by": "created_by",
        "display_name": "display_name"
    }
    `
)

var (
    proxyDataSourceMockGetHandler *mock.Handler

    proxyDataSourceMockData map[string]interface{}
)

func setupMockClientProxyDataSourceConfigMapping(t *testing.T) {
    tfmock.MockClient().Reset()

    proxyDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/proxy/id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            proxyDataSourceMockData = tfmock.CreateMapFromJsonString(t, proxyMappingResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", proxyDataSourceMockData), nil
        },
    )
}

func TestDataSourceProxyConfigMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        data "fivetran_proxy" "test_proxy" {
            provider = fivetran-provider
            id = "id"
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, proxyDataSourceMockGetHandler.Interactions, 1)
                tfmock.AssertNotEmpty(t, proxyDataSourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("data.fivetran_proxy.test_proxy", "id", "id"),
            resource.TestCheckResourceAttr("data.fivetran_proxy.test_proxy", "account_id", "account_id"),
            resource.TestCheckResourceAttr("data.fivetran_proxy.test_proxy", "registred_at", "registred_at"),
            resource.TestCheckResourceAttr("data.fivetran_proxy.test_proxy", "group_region", "region"),
            resource.TestCheckResourceAttr("data.fivetran_proxy.test_proxy", "token", "token"),
            resource.TestCheckResourceAttr("data.fivetran_proxy.test_proxy", "salt", "salt"),
            resource.TestCheckResourceAttr("data.fivetran_proxy.test_proxy", "created_by", "created_by"),
            resource.TestCheckResourceAttr("data.fivetran_proxy.test_proxy", "display_name", "display_name"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientProxyDataSourceConfigMapping(t)
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
