package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	dbtProjectResourceCreateMockGetHandler    *mock.Handler
	dbtProjectResourceCreateMockPostHandler   *mock.Handler
	dbtProjectResourceCreateMockDeleteHandler *mock.Handler

	dbtProjectResourceCreateMockData map[string]interface{}
)

func setupMockClientDbtProjectResourceCreateTest(t *testing.T) {
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
		},
		"status": "NOT_READY"
	}`

	dbtProjectResponseReady := `
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
	mockClient.Reset()

	getIteration := 0

	dbtProjectResourceCreateMockGetHandler = mockClient.When(http.MethodGet, "/v1/dbt/projects/project_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			getIteration = getIteration + 1
			if getIteration == 1 {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, dbtProjectResponse)), nil
			} else {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, dbtProjectResponseReady)), nil
			}

		},
	)

	dbtProjectResourceCreateMockPostHandler = mockClient.When(http.MethodPost, "/v1/dbt/projects").ThenCall(
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

			dbtProjectResourceCreateMockData = createMapFromJsonString(t, dbtProjectResponse)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", dbtProjectResourceCreateMockData), nil
		},
	)

	dbtProjectResourceCreateMockDeleteHandler = mockClient.When(http.MethodDelete,
		"/v1/dbt/projects/project_id",
	).ThenCall(
		func(req *http.Request) (*http.Response, error) {
			dbtProjectResourceCreateMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
		},
	)
}

func TestResourceDbtProjectCreateMock(t *testing.T) {
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
				assertEqual(t, dbtProjectResourceCreateMockPostHandler.Interactions, 1)
				assertEqual(t, dbtProjectResourceCreateMockGetHandler.Interactions, 3)
				assertNotEmpty(t, dbtProjectResourceCreateMockData)
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

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientDbtProjectResourceCreateTest(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, dbtProjectResourceCreateMockDeleteHandler.Interactions, 1)
				assertEmpty(t, dbtProjectResourceCreateMockData)
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
