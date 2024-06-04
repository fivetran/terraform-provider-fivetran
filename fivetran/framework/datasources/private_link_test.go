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
    privateLinkMappingResponse = `
    {
        "id": "id",
        "name": "name",
        "region": "region",
        "service": "service",
        "account_id": "account_id",
        "cloud_provider": "cloud_provider",
        "state": "state",
        "state_summary": "state_summary",
        "created_at": "created_at",
        "created_by": "created_by",
        "config": {
            "connection_service_name": "connection_service_name"
        }
    }
    `
)

var (
    privateLinkDataSourceMockGetHandler *mock.Handler

    privateLinkDataSourceMockData map[string]interface{}
)

func setupMockClientPrivateLinkDataSourceConfigMapping(t *testing.T) {
    tfmock.MockClient().Reset()

    privateLinkDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/private-links/id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            privateLinkDataSourceMockData = tfmock.CreateMapFromJsonString(t, privateLinkMappingResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", privateLinkDataSourceMockData), nil
        },
    )
}

func TestDataSourcePrivateLinkConfigMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        data "fivetran_private_link" "test_pl" {
            provider = fivetran-provider
            id = "id"
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, privateLinkDataSourceMockGetHandler.Interactions, 1)
                tfmock.AssertNotEmpty(t, privateLinkDataSourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("data.fivetran_private_link.test_pl", "name", "name"),
            resource.TestCheckResourceAttr("data.fivetran_private_link.test_pl", "region", "region"),
            resource.TestCheckResourceAttr("data.fivetran_private_link.test_pl", "service", "service"),
            resource.TestCheckResourceAttr("data.fivetran_private_link.test_pl", "cloud_provider", "cloud_provider"),
            resource.TestCheckResourceAttr("data.fivetran_private_link.test_pl", "state", "state"),
            resource.TestCheckResourceAttr("data.fivetran_private_link.test_pl", "state_summary", "state_summary"),
            resource.TestCheckResourceAttr("data.fivetran_private_link.test_pl", "created_at", "created_at"),
            resource.TestCheckResourceAttr("data.fivetran_private_link.test_pl", "created_by", "created_by"),

            resource.TestCheckResourceAttr("data.fivetran_private_link.test_pl", "config.connection_service_name", "connection_service_name"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientPrivateLinkDataSourceConfigMapping(t)
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
