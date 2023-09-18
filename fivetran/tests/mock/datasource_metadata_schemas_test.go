package mock

import (
    "net/http"
    "testing"

    "github.com/fivetran/go-fivetran/tests/mock"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
    metadataSchemasDataSourceMockGetHandler *mock.Handler
    metadataSchemasDataSourceMockData       map[string]interface{}
)

const (
    metadataSchemasMappingResponse = `
    {
        "items": [
        {
            "id": "bWFpbl9wdWJsaWM",
            "name_in_source": "Main Public",
            "name_in_destination": "main_public"
        }],
        "next_cursor": null
    }`
)

func setupMockClientMetadataSchemasDataSourceConfigMapping(t *testing.T) {
    mockClient.Reset()

    metadataSchemasDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/metadata/connectors/connector_id/schemas").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            metadataSchemasDataSourceMockData = createMapFromJsonString(t, metadataSchemasMappingResponse)
            return fivetranSuccessResponse(t, req, http.StatusOK, "Success", metadataSchemasDataSourceMockData), nil
        },
    )
}

func TestDataSourceMetadataSchemasMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        data "fivetran_metadata_schemas" "test_metadata_schemas" {
            id = "connector_id"
            provider = fivetran-provider
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, metadataSchemasDataSourceMockGetHandler.Interactions, 2)
                assertNotEmpty(t, metadataSchemasDataSourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("data.fivetran_metadata_schemas.test_metadata_schemas", "metadata_schemas.0.id", "bWFpbl9wdWJsaWM"),
            resource.TestCheckResourceAttr("data.fivetran_metadata_schemas.test_metadata_schemas", "metadata_schemas.0.name_in_source", "Main Public"),
            resource.TestCheckResourceAttr("data.fivetran_metadata_schemas.test_metadata_schemas", "metadata_schemas.0.name_in_destination", "main_public"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientMetadataSchemasDataSourceConfigMapping(t)
            },
            Providers: testProviders,
            CheckDestroy: func(s *terraform.State) error {
                return nil
            },
            Steps: []resource.TestStep{
                step1,
            },
        },
    )
}
