package resources_test

import (
	"net/http"
	"testing"
	
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	teamPostHandler   *mock.Handler
	teamPatchHandler  *mock.Handler
	teamDeleteHandler *mock.Handler
	teamData map[string]interface{}
)

func setupMockClientTeamResource(t *testing.T) {
	tfmock.MockClient().Reset()
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

	teamPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/teams").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamData = tfmock.CreateMapFromJsonString(t, teamResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Team has been created", teamData), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/teams/team_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", teamData), nil
		},
	)

	teamPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/teams/team_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamData = tfmock.CreateMapFromJsonString(t, teamUpdatedResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Team has been updated", teamData), nil
		},
	)

	teamDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/teams/team_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "Team has been deleted", nil), nil
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
				tfmock.AssertEqual(t, teamPostHandler.Interactions, 1)
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
				tfmock.AssertEqual(t, teamPatchHandler.Interactions, 1)
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
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}
