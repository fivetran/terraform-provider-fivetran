package mock

import (
    "net/http"
    "testing"

    "github.com/fivetran/go-fivetran/tests/mock"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
    teamGroupMembershipPostHandler   *mock.Handler
    teamGroupMembershipPatchHandler  *mock.Handler
    teamGroupMembershipDeleteHandler *mock.Handler
    teamGroupMembershipTestHandler   *mock.Handler
    teamGroupMembershipData          map[string]interface{}
)

func setupMockClientTeamGroupMembershipResource(t *testing.T) {
    mockClient.Reset()
    teamGroupMembershipResponse := 
    `{
        "id": "test_group",
        "role": "Group Reviewer",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`

    teamGroupMembershipUpdatedResponse := 
    `{
        "id": "test_group",
        "role": "Group Administrator",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`

    teamGroupMembershipPostHandler = mockClient.When(http.MethodPost, "/v1/teams/test_team/groups").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            teamGroupMembershipData = createMapFromJsonString(t, teamGroupMembershipResponse)
            return fivetranSuccessResponse(t, req, http.StatusCreated, "Group membership has been created", teamGroupMembershipData), nil
        },
    )

    mockClient.When(http.MethodGet, "/v1/teams/test_team/groups/test_group").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return fivetranSuccessResponse(t, req, http.StatusOK, "", teamGroupMembershipData), nil
        },
    )

    teamGroupMembershipPatchHandler = mockClient.When(http.MethodPatch, "/v1/teams/test_team/groups/test_group").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            teamGroupMembershipData = createMapFromJsonString(t, teamGroupMembershipUpdatedResponse)
            return fivetranSuccessResponse(t, req, http.StatusOK, "Group membership has been updated", teamGroupMembershipData), nil
        },
    )

    teamGroupMembershipDeleteHandler = mockClient.When(http.MethodDelete, "/v1/teams/test_team/groups/test_group").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return fivetranSuccessResponse(t, req, 200, "Group membership has been deleted", nil), nil
        },
    )
}

func TestGroupMembershipResourceTeamMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
            resource "fivetran_team_group_membership" "test_team_group_membership" {
                 provider = fivetran-provider

                 team_id = "test_team"
                 group_id = "test_group"
                 role = "Group Reviewer"
            }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, teamGroupMembershipPostHandler.Interactions, 1)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_team_group_membership.test_team_group_membership", "team_id", "test_team"),
            resource.TestCheckResourceAttr("fivetran_team_group_membership.test_team_group_membership", "group_id", "test_group"),
            resource.TestCheckResourceAttr("fivetran_team_group_membership.test_team_group_membership", "role", "Group Reviewer"),
        ),
    }

    step2 := resource.TestStep{
        Config: `
            resource "fivetran_team_group_membership" "test_team_group_membership" {
                 provider = fivetran-provider

                 team_id = "test_team"
                 group_id = "test_group"
                 role = "Group Administrator"
            }`,
        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, teamGroupMembershipPatchHandler.Interactions, 2)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_team_group_membership.test_team_group_membership", "team_id", "test_team"),
            resource.TestCheckResourceAttr("fivetran_team_group_membership.test_team_group_membership", "group_id", "test_group"),
            resource.TestCheckResourceAttr("fivetran_team_group_membership.test_team_group_membership", "role", "Group Administrator"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientTeamGroupMembershipResource(t)
            },
            Providers: testProviders,
            CheckDestroy: func(s *terraform.State) error {
                assertEqual(t, teamGroupMembershipDeleteHandler.Interactions, 2)
                return nil
            },

            Steps: []resource.TestStep{
                step1,
                step2,
            },
        },
    )
}
