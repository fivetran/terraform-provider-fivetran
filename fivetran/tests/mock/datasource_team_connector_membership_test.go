package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	teamConnectorMembershipDataSourceMockGetHandler *mock.Handler
	teamConnectorMembershipDataSourceMockData       map[string]interface{}
)

const (
	teamConnectorMembershipMappingResponse = `
	{
    	"id": "connector_id",
        "role": "Connector Administrator",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`
)

func setupMockClientTeamConnectorMembershipDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	teamConnectorMembershipDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/teams/team_id/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamConnectorMembershipDataSourceMockData = createMapFromJsonString(t, teamConnectorMembershipMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", teamConnectorMembershipDataSourceMockData), nil
		},
	)
}

func TestDataSourceTeamConnectorMembershipMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_team_connector_membership" "test_team_connector_membership" {
			provider 	 = fivetran-provider
			team_id 	 = "team_id"
			connector_id = "connector_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, teamConnectorMembershipDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, teamConnectorMembershipDataSourceMockData)
				return nil
			},
            resource.TestCheckResourceAttr("data.fivetran_team_connector_membership.test_team_connector_membership", "team_id", "team_id"),
            resource.TestCheckResourceAttr("data.fivetran_team_connector_membership.test_team_connector_membership", "connector_id", "connector_id"),
            resource.TestCheckResourceAttr("data.fivetran_team_connector_membership.test_team_connector_membership", "role", "Connector Administrator"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamConnectorMembershipDataSourceConfigMapping(t)
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
