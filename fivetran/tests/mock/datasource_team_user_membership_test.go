package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	teamUserMembershipDataSourceMockGetHandler *mock.Handler
	teamUserMembershipDataSourceMockData       map[string]interface{}
)

const (
	teamUserMembershipMappingResponse = `
	{
      "user_id": "user_id",
      "role": "Team Member"
    }`
)

func setupMockClientTeamUserMembershipDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	teamUserMembershipDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/teams/team_id/users/user_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamUserMembershipDataSourceMockData = createMapFromJsonString(t, teamUserMembershipMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", teamUserMembershipDataSourceMockData), nil
		},
	)
}

func TestDataSourceTeamUserMembershipMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_team_user_membership" "test_team_user_membership" {
			provider 	 = fivetran-provider
			team_id 	 = "team_id"
			user_id 	 = "user_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, teamUserMembershipDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, teamUserMembershipDataSourceMockData)
				return nil
			},
            resource.TestCheckResourceAttr("data.fivetran_team_user_membership.test_team_user_membership", "team_id", "team_id"),
            resource.TestCheckResourceAttr("data.fivetran_team_user_membership.test_team_user_membership", "user_id", "user_id"),
            resource.TestCheckResourceAttr("data.fivetran_team_user_membership.test_team_user_membership", "role", "Team Member"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamUserMembershipDataSourceConfigMapping(t)
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
