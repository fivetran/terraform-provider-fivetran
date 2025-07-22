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
    privateLinksMappingResponse = `
    {
        "items": [
        {
            "id": "id1",
            "name": "name1",
            "region": "region1",
            "service": "service1",
            "account_id": "account_id1",
            "cloud_provider": "cloud_provider1",
            "state": "state1",
            "state_summary": "state_summary1",
            "created_at": "created_at1",
            "created_by": "created_by1",
            "host": "host1"
        },
        {
            "id": "id2",
            "name": "name2",
            "region": "region2",
            "service": "service2",
            "account_id": "account_id2",
            "cloud_provider": "cloud_provider2",
            "state": "state2",
            "state_summary": "state_summary2",
            "created_at": "created_at2",
            "created_by": "created_by2",
            "host": "host2"
        }
        ],
        "next_cursor": null
    }`
)

var (
    privateLinksDataSourceMockGetHandler *mock.Handler

    privateLinksDataSourceMockData map[string]interface{}
)

func setupMockClientPrivateLinksDataSourceConfigMapping(t *testing.T) {
    tfmock.MockClient().Reset()

    privateLinksDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/private-links").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            privateLinksDataSourceMockData = tfmock.CreateMapFromJsonString(t, privateLinksMappingResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", privateLinksDataSourceMockData), nil
        },
    )
}

func TestDataSourcePrivateLinksConfigMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        data "fivetran_private_links" "test_pl" {
            provider = fivetran-provider
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, privateLinksDataSourceMockGetHandler.Interactions, 1)
                tfmock.AssertNotEmpty(t, privateLinksDataSourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.0.name", "name1"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.0.region", "region1"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.0.service", "service1"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.0.cloud_provider", "cloud_provider1"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.0.state", "state1"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.0.state_summary", "state_summary1"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.0.created_at", "created_at1"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.0.created_by", "created_by1"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.0.host", "host1"),

            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.1.name", "name2"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.1.region", "region2"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.1.service", "service2"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.1.cloud_provider", "cloud_provider2"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.1.state", "state2"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.1.state_summary", "state_summary2"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.1.created_at", "created_at2"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.1.created_by", "created_by2"),
            resource.TestCheckResourceAttr("data.fivetran_private_links.test_pl", "items.1.host", "host2"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientPrivateLinksDataSourceConfigMapping(t)
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
