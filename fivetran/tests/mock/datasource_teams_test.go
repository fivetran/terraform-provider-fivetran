package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	mockClient.Reset()

	teamsDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/teams").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamsDataSourceMockData = createMapFromJsonString(t, teamsMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", teamsDataSourceMockData), nil
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
				assertEqual(t, teamsDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, teamsDataSourceMockData)
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
