package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	mockClient.Reset()

	teamsDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/teams").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamsDataSourceMockData = createMapFromJsonString(t, teamsMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", teamsDataSourceMockData), nil
		},
	)

	teamGroupMembershipsDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/teams/team_id/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamGroupMembershipsDataSourceMockData = createMapFromJsonString(t, teamGroupMembershipsMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", teamGroupMembershipsDataSourceMockData), nil
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
				assertEqual(t, teamGroupMembershipsDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, teamGroupMembershipsDataSourceMockData)
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
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
