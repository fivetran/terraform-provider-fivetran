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
	destinationsDataSourceMockGetHandler *mock.Handler
	destinationsDataSourceMockData       map[string]interface{}
)

const (
	destinationsMappingResponse = `
{
    "items": [
      {
        "id": "destination_id",
        "service": "string",
        "region": "GCP_US_EAST4",
        "networking_method": "Directly",
        "setup_status": "CONNECTED",
        "daylight_saving_time_enabled": true,
        "private_link_id": "private_link_id",
        "group_id": "group_id",
        "time_zone_offset": "+3",
        "hybrid_deployment_agent_id": "hybrid_deployment_agent_id"
      }
    ],
    "next_cursor": null
  }
`
)

func setupMockClientDestinationsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	destinationsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/destinations").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			destinationsDataSourceMockData = tfmock.CreateMapFromJsonString(t, destinationsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", destinationsDataSourceMockData), nil
		},
	)
}

func TestDataSourceDestinationsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_destinations" "test_destinations" {
			provider = fivetran-provider
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationsDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, destinationsDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_destinations.test_destinations", "destinations.0.id", "destination_id"),
			resource.TestCheckResourceAttr("data.fivetran_destinations.test_destinations", "destinations.0.service", "string"),
			resource.TestCheckResourceAttr("data.fivetran_destinations.test_destinations", "destinations.0.region", "GCP_US_EAST4"),
			resource.TestCheckResourceAttr("data.fivetran_destinations.test_destinations", "destinations.0.group_id", "group_id"),
			resource.TestCheckResourceAttr("data.fivetran_destinations.test_destinations", "destinations.0.setup_status", "CONNECTED"),
			resource.TestCheckResourceAttr("data.fivetran_destinations.test_destinations", "destinations.0.networking_method", "Directly"),
			resource.TestCheckResourceAttr("data.fivetran_destinations.test_destinations", "destinations.0.hybrid_deployment_agent_id", "hybrid_deployment_agent_id"),
			resource.TestCheckResourceAttr("data.fivetran_destinations.test_destinations", "destinations.0.private_link_id", "private_link_id"),
			resource.TestCheckResourceAttr("data.fivetran_destinations.test_destinations", "destinations.0.time_zone_offset", "+3"),
			resource.TestCheckResourceAttr("data.fivetran_destinations.test_destinations", "destinations.0.daylight_saving_time_enabled", "true"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientDestinationsDataSourceConfigMapping(t)
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
