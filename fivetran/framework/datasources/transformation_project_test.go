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
	transformationProjectDataSourceMockGetHandler *mock.Handler
	transformationProjectDataSourceMockData       map[string]interface{}
)

func setupMockClientTransformationProjectDataSourceMappingTest(t *testing.T) {
	transformationProjectResponse := `
{
    "id": "projectId",
    "type": "DBT_GIT",
    "status": "NOT_READY",
    "errors": [
      "string"
    ],
    "created_at": "created_at",
    "group_id": "group_id",
    "setup_tests": [
      {
        "title": "Test Title",
        "status": "FAILED",
        "message": "Error message",
        "details": "Error details"
      }
    ],
    "created_by_id": "created_by_id",
    "project_config": {
      "dbt_version": "dbt_version",
      "default_schema": "default_schema",
      "git_remote_url": "git_remote_url",
      "folder_path": "folder_path",
      "git_branch": "git_branch",
      "threads": 0,
      "target_name": "target_name",
      "environment_vars": [
        "environment_var"
      ],
      "public_key": "public_key"
    }
}`
	tfmock.MockClient().Reset()

	transformationProjectDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/transformation-projects/projectId").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			transformationProjectDataSourceMockData = tfmock.CreateMapFromJsonString(t, transformationProjectResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", transformationProjectDataSourceMockData), nil
		},
	)
}

func TestDataSourceTransformationProjectMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_transformation_project" "project" {
			provider = fivetran-provider
			id = "projectId"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, transformationProjectDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, transformationProjectDataSourceMockData)
				return nil
			},
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "id", "projectId"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "group_id", "group_id"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "created_at", "created_at"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "created_by_id", "created_by_id"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "type", "DBT_GIT"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "project_config.dbt_version", "dbt_version"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "project_config.public_key", "public_key"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "project_config.default_schema", "default_schema"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "project_config.target_name", "target_name"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "project_config.environment_vars.0", "environment_var"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "project_config.git_remote_url", "git_remote_url"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "project_config.git_branch", "git_branch"),
            resource.TestCheckResourceAttr("data.fivetran_transformation_project.project", "project_config.folder_path", "folder_path"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTransformationProjectDataSourceMappingTest(t)
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
