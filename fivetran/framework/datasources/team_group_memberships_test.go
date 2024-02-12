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
	teamGroupMembershipsDataSourceMockGetHandler *mock.Handler
	teamGroupMembershipsDataSourceMockData       map[string]interface{}
)

const (
	teamGroupMembershipsMappingResponse = `
    {
      "items": [
        {
          "id": "group_id_1",
          "role": "Destination Administrator",
          "created_at": "2020-05-25T15:26:47.306509Z"
        },
        {
          "id": "group_id_2",
          "role": "Destination Reviewer",
          "created_at": "2020-05-25T15:26:47.306509Z"
        }
      ],
      "next_cursor": null
    }`
)

func setupMockClientTeamGroupMembershipsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	teamsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/teams").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamsDataSourceMockData = tfmock.CreateMapFromJsonString(t, teamsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", teamsDataSourceMockData), nil
		},
	)

	teamGroupMembershipsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/teams/team_id/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamGroupMembershipsDataSourceMockData = tfmock.CreateMapFromJsonString(t, teamGroupMembershipsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", teamGroupMembershipsDataSourceMockData), nil
		},
	)
}

func TestDataSourceTeamGroupMembershipsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_team_group_memberships" "test_team_group_memberships" {
            provider     = fivetran-provider
            team_id      = "team_id"
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamGroupMembershipsDataSourceMockGetHandler.Interactions, 2)
				tfmock.AssertNotEmpty(t, teamGroupMembershipsDataSourceMockData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamGroupMembershipsDataSourceConfigMapping(t)
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
