package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	dbtProjectResourceMockGetHandler    *mock.Handler
	dbtProjectResourceMockPostHandler   *mock.Handler
	dbtProjectResourceMockPatchHandler  *mock.Handler
	dbtProjectResourceMockDeleteHandler *mock.Handler

	dbtProjectResourceMockData map[string]interface{}
)

func setupMockClientDbtProjectResourceMappingTest(t *testing.T) {
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
		"environment_vars": ["environment_var"],
		"threads": 1,
		"type": "GIT",
		"project_config": {
			"git_remote_url": "git_remote_url",
			"git_branch": "git_branch",
			"folder_path": "folder_path"
		}
	}`
	mockClient.Reset()

	dbtProjectResourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/dbt/projects/project_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", dbtProjectResourceMockData), nil
		},
	)

	dbtProjectResourceMockPatchHandler = mockClient.When(http.MethodPatch, "/v1/dbt/projects/project_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)

			assertKeyExistsAndHasValue(t, body, "dbt_version", "dbt_version_1")
			assertKeyExistsAndHasValue(t, body, "target_name", "target_name_1")

			varsFromRequest := body["environment_vars"].([]interface{})
			assertEqual(t, len(varsFromRequest), 1)
			assertEqual(t, varsFromRequest[0], "environment_var_1")

			assertKeyExistsAndHasValue(t, body, "threads", float64(2))

			assertKeyExists(t, body, "project_config")

			config := body["project_config"].(map[string]interface{})

			assertKeyExistsAndHasValue(t, config, "git_branch", "git_branch_1")
			assertKeyExistsAndHasValue(t, config, "folder_path", "folder_path_1")

			for k, v := range body {
				if k != "project_config" {
					dbtProjectResourceMockData[k] = v
				} else {
					projectConfig := dbtProjectResourceMockData[k].(map[string]interface{})
					for ck, cv := range config {
						projectConfig[ck] = cv
					}
				}
			}

			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", dbtProjectResourceMockData), nil
		},
	)

	dbtProjectResourceMockPostHandler = mockClient.When(http.MethodPost, "/v1/dbt/projects").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)

			assertKeyExistsAndHasValue(t, body, "group_id", "group_id")
			assertKeyExistsAndHasValue(t, body, "dbt_version", "dbt_version")
			assertKeyExistsAndHasValue(t, body, "default_schema", "default_schema")
			assertKeyExistsAndHasValue(t, body, "target_name", "target_name")
			varsFromRequest := body["environment_vars"].([]interface{})
			assertEqual(t, len(varsFromRequest), 1)
			assertEqual(t, varsFromRequest[0], "environment_var")

			assertKeyExistsAndHasValue(t, body, "threads", float64(1))
			assertKeyExistsAndHasValue(t, body, "type", "GIT")

			assertKeyExists(t, body, "project_config")

			config := body["project_config"].(map[string]interface{})

			assertKeyExistsAndHasValue(t, config, "git_remote_url", "git_remote_url")
			assertKeyExistsAndHasValue(t, config, "git_branch", "git_branch")
			assertKeyExistsAndHasValue(t, config, "folder_path", "folder_path")

			dbtProjectResourceMockData = createMapFromJsonString(t, dbtProjectResponse)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", dbtProjectResourceMockData), nil
		},
	)

	dbtProjectResourceMockDeleteHandler = mockClient.When(http.MethodDelete,
		"/v1/projects/project_id", //"/v1/dbt/projects/project_id",
	).ThenCall(
		func(req *http.Request) (*http.Response, error) {
			dbtProjectResourceMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
		},
	)
}

func TestResourceDbtProjectMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_dbt_project" "project" {
			provider = fivetran-provider
			group_id = "group_id"
			dbt_version = "dbt_version"
			default_schema = "default_schema"
			target_name = "target_name"
			environment_vars = ["environment_var"]
			threads = 1
			type = "GIT"
			project_config {
				git_remote_url = "git_remote_url"
				git_branch = "git_branch"
				folder_path = "folder_path"
			}
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, dbtProjectResourceMockPostHandler.Interactions, 1)
				assertEqual(t, dbtProjectResourceMockGetHandler.Interactions, 1)
				assertNotEmpty(t, dbtProjectResourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "id", "project_id"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "group_id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "dbt_version", "dbt_version"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "created_at", "created_at"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "created_by_id", "created_by_id"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "public_key", "public_key"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "default_schema", "default_schema"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "target_name", "target_name"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "environment_vars.0", "environment_var"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "threads", "1"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "type", "GIT"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.0.git_remote_url", "git_remote_url"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.0.git_branch", "git_branch"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.0.folder_path", "folder_path"),
		),
	}

	step2 := resource.TestStep{
		Config: `
		resource "fivetran_dbt_project" "project" {
			provider = fivetran-provider
			group_id = "group_id"
			dbt_version = "dbt_version_1"
			default_schema = "default_schema"
			target_name = "target_name_1"
			environment_vars = ["environment_var_1"]
			threads = 2
			type = "GIT"
			project_config {
				git_remote_url = "git_remote_url"
				git_branch = "git_branch_1"
				folder_path = "folder_path_1"
			}
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, dbtProjectResourceMockPatchHandler.Interactions, 1)
				assertEqual(t, dbtProjectResourceMockGetHandler.Interactions, 4)
				assertNotEmpty(t, dbtProjectResourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "id", "project_id"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "group_id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "dbt_version", "dbt_version_1"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "created_at", "created_at"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "created_by_id", "created_by_id"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "public_key", "public_key"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "default_schema", "default_schema"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "target_name", "target_name_1"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "environment_vars.0", "environment_var_1"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "threads", "2"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "type", "GIT"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.0.git_remote_url", "git_remote_url"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.0.git_branch", "git_branch_1"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.0.folder_path", "folder_path_1"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientDbtProjectResourceMappingTest(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}
