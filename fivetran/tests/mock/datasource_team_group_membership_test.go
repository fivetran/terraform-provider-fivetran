package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	teamGroupMembershipDataSourceMockGetHandler *mock.Handler
	teamGroupMembershipDataSourceMockData       map[string]interface{}
)

const (
	teamGroupMembershipMappingResponse = `
	{
          "id": "group_id",
          "role": "Destination Administrator",
          "created_at": "2020-05-25T15:26:47.306509Z"
    }`
)

func setupMockClientTeamGroupMembershipDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	teamGroupMembershipDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/teams/team_id/groups/group_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamGroupMembershipDataSourceMockData = createMapFromJsonString(t, teamGroupMembershipMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", teamGroupMembershipDataSourceMockData), nil
		},
	)
}

func TestDataSourceTeamGroupMembershipMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_team_group_membership" "test_team_group_membership" {
			provider 	 = fivetran-provider
			team_id 	 = "team_id"
			group_id 	 = "group_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, teamGroupMembershipDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, teamGroupMembershipDataSourceMockData)
				return nil
			},
            resource.TestCheckResourceAttr("data.fivetran_team_group_membership.test_team_group_membership", "team_id", "team_id"),
            resource.TestCheckResourceAttr("data.fivetran_team_group_membership.test_team_group_membership", "group_id", "group_id"),
            resource.TestCheckResourceAttr("data.fivetran_team_group_membership.test_team_group_membership", "role", "Destination Administrator"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamGroupMembershipDataSourceConfigMapping(t)
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
