package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	mockClient.Reset()

	connectorsMetadataDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/metadata/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorsMetadataDataSourceMockData = createMapFromJsonString(t, connectorsMetadataMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorsMetadataDataSourceMockData), nil
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
				assertEqual(t, connectorsMetadataDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, connectorsMetadataDataSourceMockData)
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
