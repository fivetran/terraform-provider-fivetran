package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	teamDataSourceMockGetHandler *mock.Handler
	teamDataSourceMockData       map[string]interface{}
)

const (
	teamMappingResponse = `
	{
      "id": "team_id",
      "name": "test_team",
      "description": "test_description",
      "role": "Account Reviewer"
    }`
)

func setupMockClientTeamDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	teamDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/teams/team_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamDataSourceMockData = createMapFromJsonString(t, teamMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", teamDataSourceMockData), nil
		},
	)
}

func TestDataSourceTeamMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_team" "test_team" {
			provider = fivetran-provider
			id = "team_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, teamDataSourceMockGetHandler.Interactions, 4)
				assertNotEmpty(t, teamDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_team.test_team", "name", "test_team"),
			resource.TestCheckResourceAttr("data.fivetran_team.test_team", "description", "test_description"),
			resource.TestCheckResourceAttr("data.fivetran_team.test_team", "role", "Account Reviewer"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamDataSourceConfigMapping(t)
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
