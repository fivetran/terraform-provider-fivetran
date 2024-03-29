package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	dbtProjectsDataSourceMockGetHandler *mock.Handler
	dbtProjectsDataSourceMockData       map[string]interface{}
)

const (
	dbtProjectsMappingResponseWithCursor = `
	{
		"items":[
			{
				"id": "project_id",
				"group_id": "group_id",
				"created_at": "created_at",
				"created_by_id": "user_id"
			}
		],
		"next_cursor": "next_cursor"	
    }
	`

	dbtProjectsMappingResponse = `
	{
		"items":[
			{
				"id": "project_id_2",
				"group_id": "group_id_2",
				"created_at": "created_at_2",
				"created_by_id": "user_id_2"
			}
		],
		"next_cursor": null	
    }
	`
)

func setupMockClientdbtProjectsDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()
	getIteration := 0
	dbtProjectsDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/dbt/projects").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			cursor := req.URL.Query().Get("cursor")
			if cursor == "" {
				dbtProjectsDataSourceMockData = createMapFromJsonString(t, dbtProjectsMappingResponseWithCursor)
			} else {
				assertEqual(t, cursor, "next_cursor")
				dbtProjectsDataSourceMockData = createMapFromJsonString(t, dbtProjectsMappingResponse)
			}
			getIteration = getIteration + 1
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", dbtProjectsDataSourceMockData), nil
		},
	)
}

func TestDataSourceDbtProjectsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_dbt_projects" "test_projects" {
			provider = fivetran-provider
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, dbtProjectsDataSourceMockGetHandler.Interactions, 4)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_dbt_projects.test_projects", "projects.0.id", "project_id"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_projects.test_projects", "projects.0.group_id", "group_id"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_projects.test_projects", "projects.0.created_at", "created_at"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_projects.test_projects", "projects.0.created_by_id", "user_id"),

			resource.TestCheckResourceAttr("data.fivetran_dbt_projects.test_projects", "projects.1.id", "project_id_2"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_projects.test_projects", "projects.1.group_id", "group_id_2"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_projects.test_projects", "projects.1.created_at", "created_at_2"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_projects.test_projects", "projects.1.created_by_id", "user_id_2"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientdbtProjectsDataSourceConfigMapping(t)
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
