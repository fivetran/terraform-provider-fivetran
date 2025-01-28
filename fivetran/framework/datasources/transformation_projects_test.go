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
	transformationProjectsDataSourceMockGetHandler *mock.Handler
	transformationProjectsDataSourceMockData       map[string]interface{}
)

const (
	transformationProjectsMappingResponse = `
	{
		"items":[
     {
        "id": "string",
        "type": "DBT_GIT",
        "created_at": "created_at",
        "created_by_id": "string",
        "group_id": "string"
      },
     {
        "id": "string2",
        "type": "DBT_GIT",
        "created_at": "created_at_2",
        "created_by_id": "string2",
        "group_id": "string2"
      }
		],
		"next_cursor": null	
    }
	`
)

func setupMockClientTransformationProjectsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()
	transformationProjectsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/transformation-projects").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			transformationProjectsDataSourceMockData = tfmock.CreateMapFromJsonString(t, transformationProjectsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", transformationProjectsDataSourceMockData), nil
		},
	)
}

func TestDataSourceTransformationProjectsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_transformation_projects" "test_projects" {
			provider = fivetran-provider
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, transformationProjectsDataSourceMockGetHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_transformation_projects.test_projects", "projects.0.id", "string"),
			resource.TestCheckResourceAttr("data.fivetran_transformation_projects.test_projects", "projects.0.group_id", "string"),
			resource.TestCheckResourceAttr("data.fivetran_transformation_projects.test_projects", "projects.0.created_at", "created_at"),
			resource.TestCheckResourceAttr("data.fivetran_transformation_projects.test_projects", "projects.0.created_by_id", "string"),
			resource.TestCheckResourceAttr("data.fivetran_transformation_projects.test_projects", "projects.0.type", "DBT_GIT"),

			resource.TestCheckResourceAttr("data.fivetran_transformation_projects.test_projects", "projects.1.id", "string2"),
			resource.TestCheckResourceAttr("data.fivetran_transformation_projects.test_projects", "projects.1.group_id", "string2"),
			resource.TestCheckResourceAttr("data.fivetran_transformation_projects.test_projects", "projects.1.created_at", "created_at_2"),
			resource.TestCheckResourceAttr("data.fivetran_transformation_projects.test_projects", "projects.1.created_by_id", "string2"),
			resource.TestCheckResourceAttr("data.fivetran_transformation_projects.test_projects", "projects.1.type", "DBT_GIT"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTransformationProjectsDataSourceConfigMapping(t)
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
