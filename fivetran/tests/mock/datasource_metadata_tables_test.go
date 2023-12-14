package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	metadataTablesDataSourceMockGetHandler *mock.Handler
	metadataTablesDataSourceMockData       map[string]interface{}
)

const (
	metadataTablesMappingResponse = `
    {
        "items": [
        {
            "id": "NjUwMTU",
            "parent_id": "bWFpbl9wdWJsaWM",
            "name_in_source": "User Accounts",
            "name_in_destination": "user_accounts"
        }],
        "next_cursor": null
    }`
)

func setupMockClientMetadataTablesDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	metadataTablesDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/metadata/connectors/connector_id/tables").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			metadataTablesDataSourceMockData = createMapFromJsonString(t, metadataTablesMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", metadataTablesDataSourceMockData), nil
		},
	)
}

func TestDataSourceMetadataTablesMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_metadata_tables" "test_metadata_tables" {
            id = "connector_id"
            provider = fivetran-provider
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, metadataTablesDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, metadataTablesDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_metadata_tables.test_metadata_tables", "metadata_tables.0.id", "NjUwMTU"),
			resource.TestCheckResourceAttr("data.fivetran_metadata_tables.test_metadata_tables", "metadata_tables.0.parent_id", "bWFpbl9wdWJsaWM"),
			resource.TestCheckResourceAttr("data.fivetran_metadata_tables.test_metadata_tables", "metadata_tables.0.name_in_source", "User Accounts"),
			resource.TestCheckResourceAttr("data.fivetran_metadata_tables.test_metadata_tables", "metadata_tables.0.name_in_destination", "user_accounts"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientMetadataTablesDataSourceConfigMapping(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
