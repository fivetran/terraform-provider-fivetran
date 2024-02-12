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
	teamConnectorMembershipsDataSourceMockGetHandler *mock.Handler
	teamConnectorMembershipsDataSourceMockData       map[string]interface{}
)

const (
	teamConnectorMembershipsMappingResponse = `
    {
      "items": [
        {
          "id": "connector_id_1",
          "role": "Connector Administrator",
          "created_at": "2020-05-25T15:26:47.306509Z"
        },
        {
          "id": "connector_id_2",
          "role": "Connector Reviewer",
          "created_at": "2020-05-25T15:26:47.306509Z"
        }
      ],
      "next_cursor": null
    }`
)

func setupMockClientTeamConnectorMembershipsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	teamsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/teams").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamsDataSourceMockData = tfmock.CreateMapFromJsonString(t, teamsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", teamsDataSourceMockData), nil
		},
	)

	teamConnectorMembershipsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/teams/team_id/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamConnectorMembershipsDataSourceMockData = tfmock.CreateMapFromJsonString(t, teamConnectorMembershipsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", teamConnectorMembershipsDataSourceMockData), nil
		},
	)
}

func TestDataSourceTeamConnectorMembershipsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_team_connector_memberships" "test_team_connector_memberships" {
            provider     = fivetran-provider
            team_id      = "team_id"
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamConnectorMembershipsDataSourceMockGetHandler.Interactions, 2)
				tfmock.AssertNotEmpty(t, teamConnectorMembershipsDataSourceMockData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamConnectorMembershipsDataSourceConfigMapping(t)
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
