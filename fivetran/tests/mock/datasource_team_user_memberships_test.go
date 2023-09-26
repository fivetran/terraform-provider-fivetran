package mock

import (
    "net/http"
    "testing"

    "github.com/fivetran/go-fivetran/tests/mock"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
    teamUserMembershipsDataSourceMockGetHandler *mock.Handler
    teamUserMembershipsDataSourceMockData       map[string]interface{}
)

const (
    teamUserMembershipsMappingResponse = `
    {
      "items": [
        {
          "user_id": "user_id_1",
          "role": "Team Member"
        },
        {
          "user_id": "user_id_2",
          "role": "Team Manager"
        }
      ],
      "next_cursor": null
    }`
)

func setupMockClientTeamUserMembershipsDataSourceConfigMapping(t *testing.T) {
    mockClient.Reset()

    teamsDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/teams").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            teamsDataSourceMockData = createMapFromJsonString(t, teamsMappingResponse)
            return fivetranSuccessResponse(t, req, http.StatusOK, "Success", teamsDataSourceMockData), nil
        },
    )

    teamUserMembershipsDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/teams/team_id/users").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            teamUserMembershipsDataSourceMockData = createMapFromJsonString(t, teamUserMembershipsMappingResponse)
            return fivetranSuccessResponse(t, req, http.StatusOK, "Success", teamUserMembershipsDataSourceMockData), nil
        },
    )
}

func TestDataSourceTeamUserMembershipsMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        data "fivetran_team_user_memberships" "test_team_user_memberships" {
            provider     = fivetran-provider
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, teamUserMembershipsDataSourceMockGetHandler.Interactions, 2)
                assertNotEmpty(t, teamUserMembershipsDataSourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("data.fivetran_team_user_memberships.test_team_user_memberships", "memberships.0.team_id", "team_id"),
            resource.TestCheckResourceAttr("data.fivetran_team_user_memberships.test_team_user_memberships", "memberships.0.user_id", "user_id_2"),
            resource.TestCheckResourceAttr("data.fivetran_team_user_memberships.test_team_user_memberships", "memberships.0.role", "Team Manager"),

            resource.TestCheckResourceAttr("data.fivetran_team_user_memberships.test_team_user_memberships", "memberships.1.team_id", "team_id"),
            resource.TestCheckResourceAttr("data.fivetran_team_user_memberships.test_team_user_memberships", "memberships.1.user_id", "user_id_1"),
            resource.TestCheckResourceAttr("data.fivetran_team_user_memberships.test_team_user_memberships", "memberships.1.role", "Team Member"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientTeamUserMembershipsDataSourceConfigMapping(t)
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
