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
	teamsDataSourceMockGetHandler *mock.Handler
	teamsDataSourceMockData       map[string]interface{}
)

const (
	teamsMappingResponse = `
    {
        "items":[
            {
              "id": "team_id",
              "name": "Head Team",
              "description": "Head Team description",
              "role": "Account Administrator"
            }],
        "next_cursor": null
    }`
)

func setupMockClientTeamsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	teamsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/teams").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamsDataSourceMockData = tfmock.CreateMapFromJsonString(t, teamsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", teamsDataSourceMockData), nil
		},
	)
}

func TestDataSourceTeamsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_teams" "test_teams" {
            provider = fivetran-provider
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamsDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, teamsDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_teams.test_teams", "teams.0.name", "Head Team"),
			resource.TestCheckResourceAttr("data.fivetran_teams.test_teams", "teams.0.description", "Head Team description"),
			resource.TestCheckResourceAttr("data.fivetran_teams.test_teams", "teams.0.role", "Account Administrator"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamsDataSourceConfigMapping(t)
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
