package mock

import (
    "net/http"
    "testing"

    "github.com/fivetran/go-fivetran/tests/mock"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
    metadataColumnsDataSourceMockGetHandler *mock.Handler
    metadataColumnsDataSourceMockData       map[string]interface{}
)

const (
    metadataColumnsMappingResponse = `
    {
        "items": [
        {
            "id": "NTY4ODgzNDI",
            "parent_id": "NjUwMTU",
            "name_in_source": "id",
            "name_in_destination": "id",
            "type_in_source": "Integer",
            "type_in_destination": "Integer",
            "is_primary_key": true,
            "is_foreign_key": false
        }],
        "next_cursor": null
    }`
)

func setupMockClientMetadataColumnsDataSourceConfigMapping(t *testing.T) {
    mockClient.Reset()

    metadataColumnsDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/metadata/connectors/connector_id/columns").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            metadataColumnsDataSourceMockData = createMapFromJsonString(t, metadataColumnsMappingResponse)
            return fivetranSuccessResponse(t, req, http.StatusOK, "Success", metadataColumnsDataSourceMockData), nil
        },
    )
}

func TestDataSourceMetadataColumnsMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        data "fivetran_metadata_columns" "test_metadata_columns" {
            id = "connector_id"
            provider = fivetran-provider
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, metadataColumnsDataSourceMockGetHandler.Interactions, 2)
                assertNotEmpty(t, metadataColumnsDataSourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("data.fivetran_metadata_columns.test_metadata_columns", "metadata_columns.0.id", "NTY4ODgzNDI"),
            resource.TestCheckResourceAttr("data.fivetran_metadata_columns.test_metadata_columns", "metadata_columns.0.parent_id", "NjUwMTU"),
            resource.TestCheckResourceAttr("data.fivetran_metadata_columns.test_metadata_columns", "metadata_columns.0.name_in_source", "id"),
            resource.TestCheckResourceAttr("data.fivetran_metadata_columns.test_metadata_columns", "metadata_columns.0.name_in_destination", "id"),
            resource.TestCheckResourceAttr("data.fivetran_metadata_columns.test_metadata_columns", "metadata_columns.0.type_in_source", "Integer"),
            resource.TestCheckResourceAttr("data.fivetran_metadata_columns.test_metadata_columns", "metadata_columns.0.type_in_destination", "Integer"),
            resource.TestCheckResourceAttr("data.fivetran_metadata_columns.test_metadata_columns", "metadata_columns.0.is_primary_key", "true"),
            resource.TestCheckResourceAttr("data.fivetran_metadata_columns.test_metadata_columns", "metadata_columns.0.is_foreign_key", "false"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientMetadataColumnsDataSourceConfigMapping(t)
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
