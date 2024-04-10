package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	connectorsMetadataDataSourceMockGetHandler *mock.Handler
	connectorsMetadataDataSourceMockData       map[string]interface{}
)

const (
	connectorsMetadataMappingResponse = `
	{
		"items":[
			{
                "id": "facebook_ad_account",
                "name": "Facebook Ad Account",
                "type": "Marketing",
                "description": "Facebook Ad Account provides attribute data on Facebook Ad Accounts",
                "icon_url": "https://fivetran.com/integrations/facebook/resources/facebook.svg",
                "icons": [
                    "https://fivetran.com/integrations/facebook/resources/facebook-logo.svg",
                    "https://fivetran.com/integrations/facebook/resources/facebook-logo.png"
                ],
                "link_to_docs": "https://fivetran.com/docs/applications/facebook-ad-account",
                "link_to_erd": "https://fivetran.com/docs/applications/facebook-ad-account#schemainformation"
            }
		],
		"next_cursor": null	
    }
	`
)

func setupMockClientConnectorsMetadataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	connectorsMetadataDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/metadata/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorsMetadataDataSourceMockData = tfmock.CreateMapFromJsonString(t, connectorsMetadataMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorsMetadataDataSourceMockData), nil
		},
	)
}

func TestDataSourceConnectorsMetadataMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_connectors_metadata" "test_metadata" {
			provider = fivetran-provider
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, connectorsMetadataDataSourceMockGetHandler.Interactions, 2)
				tfmock.AssertNotEmpty(t, connectorsMetadataDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_connectors_metadata.test_metadata", "sources.0.id", "facebook_ad_account"),
			resource.TestCheckResourceAttr("data.fivetran_connectors_metadata.test_metadata", "sources.0.name", "Facebook Ad Account"),
			resource.TestCheckResourceAttr("data.fivetran_connectors_metadata.test_metadata", "sources.0.type", "Marketing"),
			resource.TestCheckResourceAttr("data.fivetran_connectors_metadata.test_metadata", "sources.0.description", "Facebook Ad Account provides attribute data on Facebook Ad Accounts"),
			resource.TestCheckResourceAttr("data.fivetran_connectors_metadata.test_metadata", "sources.0.icon_url", "https://fivetran.com/integrations/facebook/resources/facebook.svg"),
			resource.TestCheckResourceAttr("data.fivetran_connectors_metadata.test_metadata", "sources.0.link_to_docs", "https://fivetran.com/docs/applications/facebook-ad-account"),
			resource.TestCheckResourceAttr("data.fivetran_connectors_metadata.test_metadata", "sources.0.link_to_erd", "https://fivetran.com/docs/applications/facebook-ad-account#schemainformation"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorsMetadataSourceConfigMapping(t)
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
