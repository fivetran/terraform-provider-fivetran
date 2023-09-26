package mock

import (
    "net/http"
    "testing"

    "github.com/fivetran/go-fivetran/tests/mock"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
    teamUserMembershipPostHandler   *mock.Handler
    teamUserMembershipPatchHandler  *mock.Handler
    teamUserMembershipDeleteHandler *mock.Handler
    teamUserMembershipTestHandler   *mock.Handler
    teamUserMembershipData          map[string]interface{}
)

func setupMockClientTeamUserMembershipResource(t *testing.T) {
    mockClient.Reset()
    teamUserMembershipResponse := 
    `{
      "user_id": "test_user",
      "role": "Team Member"
    }`

    teamUserMembershipUpdatedResponse := 
    `{
      "user_id": "test_user",
      "role": "Team Manager"
    }`

    teamUserMembershipPostHandler = mockClient.When(http.MethodPost, "/v1/teams/test_team/users").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            teamUserMembershipData = createMapFromJsonString(t, teamUserMembershipResponse)
            return fivetranSuccessResponse(t, req, http.StatusCreated, "User membership has been created", teamUserMembershipData), nil
        },
    )

    mockClient.When(http.MethodGet, "/v1/teams/test_team/users/test_user").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return fivetranSuccessResponse(t, req, http.StatusOK, "", teamUserMembershipData), nil
        },
    )

    teamUserMembershipPatchHandler = mockClient.When(http.MethodPatch, "/v1/teams/test_team/users/test_user").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            teamUserMembershipData = createMapFromJsonString(t, teamUserMembershipUpdatedResponse)
            return fivetranSuccessResponse(t, req, http.StatusOK, "User membership has been updated", teamUserMembershipData), nil
        },
    )

    teamUserMembershipDeleteHandler = mockClient.When(http.MethodDelete, "/v1/teams/test_team/users/test_user").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return fivetranSuccessResponse(t, req, 200, "User membership has been deleted", nil), nil
        },
    )
}

func TestUserMembershipResourceTeamMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
            resource "fivetran_team_user_membership" "test_team_user_membership" {
                 provider = fivetran-provider

                 team_id = "test_team"
                 user_id = "test_user"
                 role = "Team Member"
            }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, teamUserMembershipPostHandler.Interactions, 1)
                return nil
            },
                    resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "team_id", "test_team"),
                    resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "user_id", "test_user"),
                    resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "role", "Team Member"),
        ),
    }

    step2 := resource.TestStep{
        Config: `
            resource "fivetran_team_user_membership" "test_team_user_membership" {
                 provider = fivetran-provider

                 team_id = "test_team"
                 user_id = "test_user"
                 role = "Team Manager"
            }`,
        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, teamUserMembershipPatchHandler.Interactions, 1)
                return nil
            },
                    resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "team_id", "test_team"),
                    resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "user_id", "test_user"),
                    resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "role", "Team Manager"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientTeamUserMembershipResource(t)
            },
            Providers: testProviders,
            CheckDestroy: func(s *terraform.State) error {
                assertEqual(t, teamUserMembershipDeleteHandler.Interactions, 1)
                return nil
            },

            Steps: []resource.TestStep{
                step1,
                step2,
            },
        },
    )
}
