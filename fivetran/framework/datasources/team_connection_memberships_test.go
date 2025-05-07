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
	teamConnectionMembershipsDataSourceMockGetHandler *mock.Handler
	teamConnectionMembershipsDataSourceMockData       map[string]interface{}
)

const (
	teamConnectionMembershipsMappingResponse = `
    {
      "items": [
        {
          "id": "connection_id_1",
          "role": "Connection Administrator",
          "created_at": "2020-05-25T15:26:47.306509Z"
        },
        {
          "id": "connection_id_2",
          "role": "Connection Reviewer",
          "created_at": "2020-05-25T15:26:47.306509Z"
        }
      ],
      "next_cursor": null
    }`
)

func setupMockClientTeamConnectionMembershipsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	teamsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/teams").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamsDataSourceMockData = tfmock.CreateMapFromJsonString(t, teamsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", teamsDataSourceMockData), nil
		},
	)

	teamConnectionMembershipsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/teams/team_id/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamConnectionMembershipsDataSourceMockData = tfmock.CreateMapFromJsonString(t, teamConnectionMembershipsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", teamConnectionMembershipsDataSourceMockData), nil
		},
	)
}

func TestDataSourceTeamConnectionMembershipsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_team_connection_memberships" "test" {
            provider     = fivetran-provider
            id           = "team_id"
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamConnectionMembershipsDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, teamConnectionMembershipsDataSourceMockData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamConnectionMembershipsDataSourceConfigMapping(t)
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
