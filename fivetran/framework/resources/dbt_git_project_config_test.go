package resources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	dbtGitProjectConfigResourceMockGetHandler    *mock.Handler
	dbtGitProjectConfigResourceMockPatchHandler  *mock.Handler
)

func setupMockClientDbtGitProjectConfigResourceCreateTest(t *testing.T) {
	dbtGitProjectConfigResponse := `
	{
		"id": "project_id",
		"group_id": "group_id",
		"dbt_version": "dbt_version",
		"created_at": "created_at",
		"created_by_id": "created_by_id",
		"public_key": "public_key",
		"default_schema": "default_schema",
		"target_name": "target_name",
		"environment_vars": ["environment_var"],
		"threads": 1,
		"type": "GIT",
		"project_config": {
			"git_remote_url": "git_remote_url",
			"git_branch": "git_branch",
			"folder_path": "folder_path"
		},
		"status": "READY"
	}`

	tfmock.MockClient().Reset()

	dbtGitProjectConfigResourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/dbt/projects/project_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, dbtGitProjectConfigResponse)), nil
		},
	)

	dbtGitProjectConfigResourceMockPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/dbt/projects/project_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, dbtGitProjectConfigResponse)), nil
		},
	)
}

func TestResourceDbtGitProjectConfigCreateMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_dbt_git_project_config" "project" {
			provider = fivetran-provider
			project_id = "project_id"

			git_remote_url = "git_remote_url"
			git_branch = "git_branch"
			folder_path = "folder_path"

			ensure_readiness = true
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, dbtGitProjectConfigResourceMockPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_dbt_git_project_config.project", "id", "project_id"),
			resource.TestCheckResourceAttr("fivetran_dbt_git_project_config.project", "git_remote_url", "git_remote_url"),
			resource.TestCheckResourceAttr("fivetran_dbt_git_project_config.project", "git_branch", "git_branch"),
			resource.TestCheckResourceAttr("fivetran_dbt_git_project_config.project", "folder_path", "folder_path"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientDbtGitProjectConfigResourceCreateTest(t)
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
