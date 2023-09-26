package mock

import (
    "net/http"
    "testing"

    "github.com/fivetran/go-fivetran/tests/mock"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
    teamPostHandler   *mock.Handler
    teamPatchHandler  *mock.Handler
    teamDeleteHandler *mock.Handler
    teamTestHandler   *mock.Handler
    teamData          map[string]interface{}
)

func setupMockClientTeamResource(t *testing.T) {
    mockClient.Reset()
    teamResponse := 
    `{
      "id": "team_id",
      "name": "test_team",
      "description": "test_description",
      "role": "Account Reviewer"
    }`

    teamUpdatedResponse := 
    `{
      "id": "team_id",
      "name": "test_team_2",
      "description": "test_description",
      "role": "Account Reviewer"
    }`

    teamPostHandler = mockClient.When(http.MethodPost, "/v1/teams").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            teamData = createMapFromJsonString(t, teamResponse)
            return fivetranSuccessResponse(t, req, http.StatusCreated, "Team has been created", teamData), nil
        },
    )

    mockClient.When(http.MethodGet, "/v1/teams/team_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return fivetranSuccessResponse(t, req, http.StatusOK, "", teamData), nil
        },
    )

    teamPatchHandler = mockClient.When(http.MethodPatch, "/v1/teams/team_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            teamData = createMapFromJsonString(t, teamUpdatedResponse)
            return fivetranSuccessResponse(t, req, http.StatusOK, "Team has been updated", teamData), nil
        },
    )

    teamDeleteHandler = mockClient.When(http.MethodDelete, "/v1/teams/team_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return fivetranSuccessResponse(t, req, 200, "Team has been deleted", nil), nil
        },
    )
}

func TestResourceTeamMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
            resource "fivetran_team" "test_team" {
                 provider = fivetran-provider

                 name = "test_team"
                 description = "test_description"
                 role = "Account Reviewer"
            }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, teamPostHandler.Interactions, 1)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_team.test_team", "name", "test_team"),
            resource.TestCheckResourceAttr("fivetran_team.test_team", "description", "test_description"),
            resource.TestCheckResourceAttr("fivetran_team.test_team", "role", "Account Reviewer"),
        ),
    }

    step2 := resource.TestStep{
        Config: `
            resource "fivetran_team" "test_team" {
                 provider = fivetran-provider

                 name = "test_team_2"
                 description = "test_description"
                 role = "Account Reviewer"
            }`,
        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, teamPatchHandler.Interactions, 1)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_team.test_team", "name", "test_team_2"),
            resource.TestCheckResourceAttr("fivetran_team.test_team", "description", "test_description"),
            resource.TestCheckResourceAttr("fivetran_team.test_team", "role", "Account Reviewer"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientTeamResource(t)
            },
            Providers: testProviders,
            CheckDestroy: func(s *terraform.State) error {
                assertEqual(t, teamDeleteHandler.Interactions, 1)
                return nil
            },

            Steps: []resource.TestStep{
                step1,
                step2,
            },
        },
    )
}
