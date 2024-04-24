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
	teamUserMembershipsDataSourceMockGetHandler *mock.Handler
	teamUserMembershipsDataSourceMockData       map[string]interface{}
)

const (
	teamUserMembershipsMappingResponse = `
    {
      "items": [
        {
          "id": "user_id_1",
          "role": "Team Member"
        },
        {
          "id": "user_id_2",
          "role": "Team Manager"
        }
      ],
      "next_cursor": null
    }`
)

func setupMockClientTeamUserMembershipsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	teamsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/teams").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamsDataSourceMockData = tfmock.CreateMapFromJsonString(t, teamsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", teamsDataSourceMockData), nil
		},
	)

	teamUserMembershipsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/teams/team_id/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamUserMembershipsDataSourceMockData = tfmock.CreateMapFromJsonString(t, teamUserMembershipsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", teamUserMembershipsDataSourceMockData), nil
		},
	)
}

func TestDataSourceTeamUserMembershipsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_team_user_memberships" "test_team_user_memberships" {
            provider     = fivetran-provider
            team_id      = "team_id"
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamUserMembershipsDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, teamUserMembershipsDataSourceMockData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamUserMembershipsDataSourceConfigMapping(t)
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
