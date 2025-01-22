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
    transformationProjectResourceMockGetHandler    *mock.Handler
    transformationProjectResourceMockPostHandler   *mock.Handler
    transformationProjectResourceMockPatchHandler  *mock.Handler
    transformationProjectResourceMockDeleteHandler *mock.Handler

    transformationProjectResourceMockData map[string]interface{}

    transformationProjectResourceCreateMockGetHandler       *mock.Handler
    transformationProjectResourceCreateMockGetModelsHandler *mock.Handler
    transformationProjectResourceCreateMockPostHandler      *mock.Handler
    transformationProjectResourceCreateMockDeleteHandler    *mock.Handler
    transformationModelsDataSourceMockGetHandler            *mock.Handler

    transformationModelsDataSourceMockData      map[string]interface{}
    transformationProjectResourceCreateMockData map[string]interface{}
)

func setupMockClientTransformationProjectResourceMappingTest(t *testing.T) {
    transformationProjectResponse := `
{
    "id": "string",
    "type": "DBT_GIT",
    "status": "NOT_READY",
    "errors": [
      "string"
    ],
    "created_at": "2019-08-24T14:15:22Z",
    "group_id": "string",
    "setup_tests": [
      {
        "title": "Test Title",
        "status": "FAILED",
        "message": "Error message",
        "details": "Error details"
      }
    ],
    "created_by_id": "string",
    "project_config": {
      "dbt_version": "string",
      "default_schema": "string",
      "git_remote_url": "string",
      "folder_path": "string",
      "git_branch": "string",
      "threads": 0,
      "target_name": "string",
      "environment_vars": [
        "string"
      ],
      "public_key": "string"
    }
  }`
    tfmock.MockClient().Reset()

    transformationProjectResourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/transformation-projects/project_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", transformationProjectResourceMockData), nil
        },
    )

    transformationProjectResourceMockPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/transformation-projects/project_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            body := tfmock.RequestBodyToJson(t, req)

            tfmock.AssertKeyExistsAndHasValue(t, body, "transformation_version", "transformation_version_1")
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
                    transformationProjectResourceMockData[k] = v
                } else {
                    projectConfig := transformationProjectResourceMockData[k].(map[string]interface{})
                    for ck, cv := range config {
                        projectConfig[ck] = cv
                    }
                }
            }

            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", transformationProjectResourceMockData), nil
        },
    )

    transformationProjectResourceMockPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/transformation-projects").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            body := tfmock.RequestBodyToJson(t, req)

            tfmock.AssertKeyExistsAndHasValue(t, body, "group_id", "group_id")
            tfmock.AssertKeyExistsAndHasValue(t, body, "transformation_version", "transformation_version")
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

            transformationProjectResourceMockData = tfmock.CreateMapFromJsonString(t, transformationProjectResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", transformationProjectResourceMockData), nil
        },
    )

    transformationProjectResourceMockDeleteHandler = tfmock.MockClient().When(http.MethodDelete,
        "/v1/transformation-projects/project_id",
    ).ThenCall(
        func(req *http.Request) (*http.Response, error) {
            transformationProjectResourceMockData = nil
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
        },
    )
}

func TestResourceTransformationProjectMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        resource "fivetran_transformation_project" "project" {
            provider = fivetran-provider
            group_id = "group_id"
            type = "DBT_GIT"
            run_tests = true

            project_config {
                git_remote_url = "git_remote_url"
                git_branch = "git_branch_1"
                folder_path = "folder_path_1"
                dbt_version = "string"
                default_schema = "string"
                threads = 0
                target_name = "string"
            }
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, transformationProjectResourceMockPostHandler.Interactions, 1)
                tfmock.AssertEqual(t, transformationProjectResourceMockGetHandler.Interactions, 0)
                tfmock.AssertNotEmpty(t, transformationProjectResourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "id", "project_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "group_id", "group_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "transformation_version", "transformation_version"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "created_at", "created_at"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "created_by_id", "created_by_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "public_key", "public_key"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "default_schema", "default_schema"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "target_name", "target_name"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "environment_vars.0", "environment_var"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "threads", "1"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "type", "GIT"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.git_remote_url", "git_remote_url"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.git_branch", "git_branch"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.folder_path", "folder_path"),
        ),
    }

    step2 := resource.TestStep{
        Config: `
        resource "fivetran_transformation_project" "project" {
            provider = fivetran-provider
            group_id = "group_id"
            type = "DBT_GIT"
            run_tests = true

            project_config {
                git_remote_url = "git_remote_url"
                git_branch = "git_branch_1"
                folder_path = "folder_path_1"
                dbt_version = "string"
                default_schema = "string"
                threads = 1
                target_name = "string"
            }
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, transformationProjectResourceMockPatchHandler.Interactions, 1)
                tfmock.AssertEqual(t, transformationProjectResourceMockGetHandler.Interactions, 2)
                tfmock.AssertNotEmpty(t, transformationProjectResourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "id", "project_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "group_id", "group_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "transformation_version", "transformation_version_1"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "created_at", "created_at"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "created_by_id", "created_by_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "public_key", "public_key"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "default_schema", "default_schema"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "target_name", "target_name_1"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "environment_vars.0", "environment_var_1"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "threads", "2"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "type", "GIT"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.git_remote_url", "git_remote_url"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.git_branch", "git_branch_1"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.folder_path", "folder_path_1"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientTransformationProjectResourceMappingTest(t)
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

func setupMockClientTransformationProjectResourceCreateTest(t *testing.T) {

    transformationProjectResponse := `
{
    "id": "string",
    "type": "DBT_GIT",
    "status": "NOT_READY",
    "errors": [
      "string"
    ],
    "created_at": "2019-08-24T14:15:22Z",
    "group_id": "string",
    "setup_tests": [
      {
        "title": "Test Title",
        "status": "FAILED",
        "message": "Error message",
        "details": "Error details"
      }
    ],
    "created_by_id": "string",
    "project_config": {
      "dbt_version": "string",
      "default_schema": "string",
      "git_remote_url": "string",
      "folder_path": "string",
      "git_branch": "string",
      "threads": 0,
      "target_name": "string",
      "environment_vars": [
        "string"
      ],
      "public_key": "string"
    }
  }`

    transformationProjectResponseReady := `
{
    "id": "string",
    "type": "DBT_GIT",
    "status": "NOT_READY",
    "errors": [
      "string"
    ],
    "created_at": "2019-08-24T14:15:22Z",
    "group_id": "string",
    "setup_tests": [
      {
        "title": "Test Title",
        "status": "FAILED",
        "message": "Error message",
        "details": "Error details"
      }
    ],
    "created_by_id": "string",
    "project_config": {
      "dbt_version": "string",
      "default_schema": "string",
      "git_remote_url": "string",
      "folder_path": "string",
      "git_branch": "string",
      "threads": 0,
      "target_name": "string",
      "environment_vars": [
        "string"
      ],
      "public_key": "string"
    }
  }`
    tfmock.MockClient().Reset()

    getIteration := 0

    transformationProjectResourceCreateMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/transformation-projects/project_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            getIteration = getIteration + 1
            if getIteration == 1 {
                return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, transformationProjectResponse)), nil
            } else {
                return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, transformationProjectResponseReady)), nil
            }

        },
    )

    transformationProjectResourceCreateMockPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/transformation-projects").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            body := tfmock.RequestBodyToJson(t, req)

            tfmock.AssertKeyExistsAndHasValue(t, body, "group_id", "group_id")
            tfmock.AssertKeyExistsAndHasValue(t, body, "transformation_version", "transformation_version")
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

            transformationProjectResourceCreateMockData = tfmock.CreateMapFromJsonString(t, transformationProjectResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", transformationProjectResourceCreateMockData), nil
        },
    )

    transformationProjectResourceCreateMockDeleteHandler = tfmock.MockClient().When(http.MethodDelete,
        "/v1/transformation-projects/project_id",
    ).ThenCall(
        func(req *http.Request) (*http.Response, error) {
            transformationProjectResourceCreateMockData = nil
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
        },
    )
}

func TestResourceTransformationProjectCreateMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        resource "fivetran_transformation_project" "project" {
            provider = fivetran-provider
            group_id = "group_id"
            type = "DBT_GIT"
            run_tests = true

            project_config {
                git_remote_url = "git_remote_url"
                git_branch = "git_branch_1"
                folder_path = "folder_path_1"
                dbt_version = "string"
                default_schema = "string"
                threads = 0
                target_name = "string"
            }
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, transformationProjectResourceCreateMockPostHandler.Interactions, 1)
                tfmock.AssertEqual(t, transformationProjectResourceCreateMockGetModelsHandler.Interactions, 0)
                tfmock.AssertNotEmpty(t, transformationProjectResourceCreateMockData)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "id", "project_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "group_id", "group_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "transformation_version", "transformation_version"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "created_at", "created_at"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "created_by_id", "created_by_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "public_key", "public_key"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "default_schema", "default_schema"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "target_name", "target_name"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "environment_vars.0", "environment_var"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "threads", "1"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "type", "GIT"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.git_remote_url", "git_remote_url"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.git_branch", "git_branch"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.folder_path", "folder_path"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientTransformationProjectResourceCreateTest(t)
            },
            ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
            CheckDestroy: func(s *terraform.State) error {
                tfmock.AssertEqual(t, transformationProjectResourceCreateMockDeleteHandler.Interactions, 1)
                tfmock.AssertEmpty(t, transformationProjectResourceCreateMockData)
                return nil
            },
            Steps: []resource.TestStep{
                step1,
            },
        },
    )
}
