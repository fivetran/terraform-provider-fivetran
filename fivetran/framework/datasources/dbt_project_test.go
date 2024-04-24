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
	dbtProjectDataSourceMockGetHandler *mock.Handler
	dbtProjectDataSourceMockData       map[string]interface{}
)

func setupMockClientDbtProjectDataSourceMappingTest(t *testing.T) {
	dbtProjectResponse := `
	{
		"id": "project_id",
		"group_id": "group_id",
		"dbt_version": "dbt_version",
		"created_at": "created_at",
		"created_by_id": "created_by_id",
		"public_key": "public_key",
		"default_schema": "default_schema",
		"target_name": "target_name",
		"environment_vars": ["DBT_VARIABLE_1=VALUE"],
		"threads": 1,
		"type": "GIT",
		"project_config": {
			"git_remote_url": "git_remote_url",
			"git_branch": "git_branch",
			"folder_path": "folder_path"
		},
		"status":"NOT_READY"
	}`
	tfmock.MockClient().Reset()

	dbtProjectDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/dbt/projects/project_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			dbtProjectDataSourceMockData = tfmock.CreateMapFromJsonString(t, dbtProjectResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", dbtProjectDataSourceMockData), nil
		},
	)
}

func TestDataSourceDbtProjectMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_dbt_project" "project" {
			provider = fivetran-provider
			id = "project_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, dbtProjectDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, dbtProjectDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "id", "project_id"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "group_id", "group_id"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "dbt_version", "dbt_version"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "created_at", "created_at"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "created_by_id", "created_by_id"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "public_key", "public_key"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "default_schema", "default_schema"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "target_name", "target_name"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "environment_vars.0", "DBT_VARIABLE_1=VALUE"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "threads", "1"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "type", "GIT"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "project_config.git_remote_url", "git_remote_url"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "project_config.git_branch", "git_branch"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_project.project", "project_config.folder_path", "folder_path"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientDbtProjectDataSourceMappingTest(t)
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
