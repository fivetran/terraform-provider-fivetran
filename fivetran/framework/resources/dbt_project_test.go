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
	dbtProjectResourceMockGetHandler    *mock.Handler
	dbtProjectResourceMockPostHandler   *mock.Handler
	dbtProjectResourceMockPatchHandler  *mock.Handler
	dbtProjectResourceMockDeleteHandler *mock.Handler

	dbtProjectResourceMockData map[string]interface{}

	dbtProjectResourceCreateMockGetHandler       *mock.Handler
	dbtProjectResourceCreateMockGetModelsHandler *mock.Handler
	dbtProjectResourceCreateMockPostHandler      *mock.Handler
	dbtProjectResourceCreateMockDeleteHandler    *mock.Handler
	dbtModelsDataSourceMockGetHandler            *mock.Handler

	dbtModelsDataSourceMockData      map[string]interface{}
	dbtProjectResourceCreateMockData map[string]interface{}
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
		},
		"status": "NOT_READY"
	}`
	tfmock.MockClient().Reset()

	dbtProjectResourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/dbt/projects/project_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", dbtProjectResourceMockData), nil
		},
	)

	dbtProjectResourceMockPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/dbt/projects/project_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)

			tfmock.AssertKeyExistsAndHasValue(t, body, "dbt_version", "dbt_version_1")
			tfmock.AssertKeyExistsAndHasValue(t, body, "target_name", "target_name_1")

			varsFromRequest := body["environment_vars"].([]interface{})
			tfmock.AssertEqual(t, len(varsFromRequest), 1)
			tfmock.AssertEqual(t, varsFromRequest[0], "environment_var_1")

			tfmock.AssertKeyExistsAndHasValue(t, body, "threads", float64(2))

			tfmock.AssertKeyExists(t, body, "project_config")

			config := body["project_config"].(map[string]interface{})

			tfmock.AssertKeyExistsAndHasValue(t, config, "git_branch", "git_branch_1")
			tfmock.AssertKeyExistsAndHasValue(t, config, "folder_path", "folder_path_1")

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

			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", dbtProjectResourceMockData), nil
		},
	)

	dbtProjectResourceMockPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/dbt/projects").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)

			tfmock.AssertKeyExistsAndHasValue(t, body, "group_id", "group_id")
			tfmock.AssertKeyExistsAndHasValue(t, body, "dbt_version", "dbt_version")
			tfmock.AssertKeyExistsAndHasValue(t, body, "default_schema", "default_schema")
			tfmock.AssertKeyExistsAndHasValue(t, body, "target_name", "target_name")
			varsFromRequest := body["environment_vars"].([]interface{})
			tfmock.AssertEqual(t, len(varsFromRequest), 1)
			tfmock.AssertEqual(t, varsFromRequest[0], "environment_var")

			tfmock.AssertKeyExistsAndHasValue(t, body, "threads", float64(1))
			tfmock.AssertKeyExistsAndHasValue(t, body, "type", "GIT")

			tfmock.AssertKeyExists(t, body, "project_config")

			config := body["project_config"].(map[string]interface{})

			tfmock.AssertKeyExistsAndHasValue(t, config, "git_remote_url", "git_remote_url")
			tfmock.AssertKeyExistsAndHasValue(t, config, "git_branch", "git_branch")
			tfmock.AssertKeyExistsAndHasValue(t, config, "folder_path", "folder_path")

			dbtProjectResourceMockData = tfmock.CreateMapFromJsonString(t, dbtProjectResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", dbtProjectResourceMockData), nil
		},
	)

	dbtProjectResourceMockDeleteHandler = tfmock.MockClient().When(http.MethodDelete,
		"/v1/dbt/projects/project_id",
	).ThenCall(
		func(req *http.Request) (*http.Response, error) {
			dbtProjectResourceMockData = nil
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
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
			ensure_readiness = false
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, dbtProjectResourceMockPostHandler.Interactions, 1)
				tfmock.AssertEqual(t, dbtProjectResourceMockGetHandler.Interactions, 0)
				tfmock.AssertNotEmpty(t, dbtProjectResourceMockData)
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
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.git_remote_url", "git_remote_url"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.git_branch", "git_branch"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.folder_path", "folder_path"),
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
				tfmock.AssertEqual(t, dbtProjectResourceMockPatchHandler.Interactions, 1)
				tfmock.AssertEqual(t, dbtProjectResourceMockGetHandler.Interactions, 2)
				tfmock.AssertNotEmpty(t, dbtProjectResourceMockData)
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
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.git_remote_url", "git_remote_url"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.git_branch", "git_branch_1"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.folder_path", "folder_path_1"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientDbtProjectResourceMappingTest(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
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

func setupMockClientDbtProjectResourceCreateTest(t *testing.T) {

	const (
		dbtModelsMappingResponseWithCursor = `
		{
			"items":[
				{
					"id": "model_id",
					"model_name": "model_name",
					"scheduled": true
				}
			],
			"next_cursor": "next_cursor"	
		}
		`

		dbtModelsMappingResponse = `
		{
			"items":[
				{
					"id": "model_id_2",
					"model_name": "model_name_2",
					"scheduled": false
				}
			],
			"next_cursor": null	
		}
		`
	)
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
	tfmock.MockClient().Reset()

	getIteration := 0

	dbtProjectResourceCreateMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/dbt/projects/project_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			getIteration = getIteration + 1
			if getIteration == 1 {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, dbtProjectResponse)), nil
			} else {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, dbtProjectResponseReady)), nil
			}

		},
	)

	dbtProjectResourceCreateMockGetModelsHandler = tfmock.MockClient().When(http.MethodGet, "/v1/dbt/models").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			project_id := req.URL.Query().Get("project_id")
			tfmock.AssertEqual(t, project_id, "project_id")
			cursor := req.URL.Query().Get("cursor")
			if cursor == "" {
				dbtModelsDataSourceMockData = tfmock.CreateMapFromJsonString(t, dbtModelsMappingResponseWithCursor)
			} else {
				tfmock.AssertEqual(t, cursor, "next_cursor")
				dbtModelsDataSourceMockData = tfmock.CreateMapFromJsonString(t, dbtModelsMappingResponse)
			}
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", dbtModelsDataSourceMockData), nil
		},
	)

	dbtProjectResourceCreateMockPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/dbt/projects").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)

			tfmock.AssertKeyExistsAndHasValue(t, body, "group_id", "group_id")
			tfmock.AssertKeyExistsAndHasValue(t, body, "dbt_version", "dbt_version")
			tfmock.AssertKeyExistsAndHasValue(t, body, "default_schema", "default_schema")
			tfmock.AssertKeyExistsAndHasValue(t, body, "target_name", "target_name")
			varsFromRequest := body["environment_vars"].([]interface{})
			tfmock.AssertEqual(t, len(varsFromRequest), 1)
			tfmock.AssertEqual(t, varsFromRequest[0], "environment_var")

			tfmock.AssertKeyExistsAndHasValue(t, body, "threads", float64(1))
			tfmock.AssertKeyExistsAndHasValue(t, body, "type", "GIT")

			tfmock.AssertKeyExists(t, body, "project_config")

			config := body["project_config"].(map[string]interface{})

			tfmock.AssertKeyExistsAndHasValue(t, config, "git_remote_url", "git_remote_url")
			tfmock.AssertKeyExistsAndHasValue(t, config, "git_branch", "git_branch")
			tfmock.AssertKeyExistsAndHasValue(t, config, "folder_path", "folder_path")

			dbtProjectResourceCreateMockData = tfmock.CreateMapFromJsonString(t, dbtProjectResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", dbtProjectResourceCreateMockData), nil
		},
	)

	dbtProjectResourceCreateMockDeleteHandler = tfmock.MockClient().When(http.MethodDelete,
		"/v1/dbt/projects/project_id",
	).ThenCall(
		func(req *http.Request) (*http.Response, error) {
			dbtProjectResourceCreateMockData = nil
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
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
				tfmock.AssertEqual(t, dbtProjectResourceCreateMockPostHandler.Interactions, 1)
				tfmock.AssertEqual(t, dbtProjectResourceCreateMockGetModelsHandler.Interactions, 2)
				tfmock.AssertEqual(t, dbtProjectResourceCreateMockGetHandler.Interactions, 2)
				tfmock.AssertNotEmpty(t, dbtProjectResourceCreateMockData)
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
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.git_remote_url", "git_remote_url"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.git_branch", "git_branch"),
			resource.TestCheckResourceAttr("fivetran_dbt_project.project", "project_config.folder_path", "folder_path"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientDbtProjectResourceCreateTest(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, dbtProjectResourceCreateMockDeleteHandler.Interactions, 1)
				tfmock.AssertEmpty(t, dbtProjectResourceCreateMockData)
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
