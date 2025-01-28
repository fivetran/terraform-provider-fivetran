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
    transformationProjectResourceMockPatchHandler   *mock.Handler
    transformationProjectResourceMockDeleteHandler *mock.Handler

    transformationProjectResourceMockData map[string]interface{}
)

func setupMockClientTransformationProjectResourceMappingTest(t *testing.T) {
    transformationProjectResponse := `
{
    "id": "project_id",
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
      "git_branch": "git_branch",
      "threads": 0,
      "target_name": "target_name",
      "environment_vars": [
        "environment_var"
      ],
      "public_key": "public_key"
    }
  }`

    transformationProjectPatchedResponse := `
{
    "id": "project_id",
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
      "git_branch": "git_branch1",
      "threads": 1,
      "target_name": "target_name1",
      "environment_vars": [
        "environment_var1"
      ],
      "public_key": "public_key"
    }
  }`

    tfmock.MockClient().Reset()

    transformationProjectResourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/transformation-projects/project_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", transformationProjectResourceMockData), nil
        },
    )

    transformationProjectResourceMockPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/transformation-projects").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            body := tfmock.RequestBodyToJson(t, req)
            tfmock.AssertKeyExistsAndHasValue(t, body, "group_id", "group_id")
            tfmock.AssertKeyExistsAndHasValue(t, body, "type", "DBT_GIT")

            tfmock.AssertKeyExists(t, body, "project_config")
            config := body["project_config"].(map[string]interface{})
            tfmock.AssertKeyExistsAndHasValue(t, config, "git_remote_url", "git_remote_url")
            tfmock.AssertKeyExistsAndHasValue(t, config, "git_branch", "git_branch")
            tfmock.AssertKeyExistsAndHasValue(t, config, "dbt_version", "dbt_version")
            tfmock.AssertKeyExistsAndHasValue(t, config, "default_schema", "default_schema")
            tfmock.AssertKeyExistsAndHasValue(t, config, "target_name", "target_name")

            transformationProjectResourceMockData = tfmock.CreateMapFromJsonString(t, transformationProjectResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", transformationProjectResourceMockData), nil
        },
    )

    transformationProjectResourceMockPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/transformation-projects/project_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            body := tfmock.RequestBodyToJson(t, req)

            tfmock.AssertKeyDoesNotExist(t, body, "group_id")
            tfmock.AssertKeyDoesNotExist(t, body, "type")

            tfmock.AssertKeyExists(t, body, "project_config")
            config := body["project_config"].(map[string]interface{})
            tfmock.AssertKeyDoesNotExist(t, config, "git_remote_url")
            tfmock.AssertKeyExistsAndHasValue(t, config, "git_branch", "git_branch1")
            tfmock.AssertKeyExistsAndHasValue(t, config, "folder_path", "folder_path")
            tfmock.AssertKeyDoesNotExist(t, config, "dbt_version")
            tfmock.AssertKeyDoesNotExist(t, config, "default_schema")
            tfmock.AssertKeyExistsAndHasValue(t, config, "target_name", "target_name1")

            transformationProjectResourceMockData = tfmock.CreateMapFromJsonString(t, transformationProjectPatchedResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", transformationProjectResourceMockData), nil
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
                git_branch = "git_branch"
                dbt_version = "dbt_version"
                default_schema = "default_schema"
                threads = 0
                target_name = "target_name"
                environment_vars = ["environment_var"]
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
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "created_at", "created_at"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "created_by_id", "created_by_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "type", "DBT_GIT"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.dbt_version", "dbt_version"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.public_key", "public_key"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.default_schema", "default_schema"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.target_name", "target_name"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.environment_vars.0", "environment_var"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.git_remote_url", "git_remote_url"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.git_branch", "git_branch"),
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
                git_branch = "git_branch1"
                folder_path = "folder_path"
                dbt_version = "dbt_version"
                default_schema = "default_schema"
                threads = 1
                target_name = "target_name1"
                environment_vars = ["environment_var1"]
            }
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, transformationProjectResourceMockPostHandler.Interactions, 1)
                tfmock.AssertEqual(t, transformationProjectResourceMockPatchHandler.Interactions, 1)
                tfmock.AssertEqual(t, transformationProjectResourceMockGetHandler.Interactions, 2)
                tfmock.AssertNotEmpty(t, transformationProjectResourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "id", "project_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "group_id", "group_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "created_at", "created_at"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "created_by_id", "created_by_id"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "type", "DBT_GIT"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.dbt_version", "dbt_version"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.public_key", "public_key"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.default_schema", "default_schema"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.target_name", "target_name1"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.environment_vars.0", "environment_var1"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.git_remote_url", "git_remote_url"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.git_branch", "git_branch1"),
            resource.TestCheckResourceAttr("fivetran_transformation_project.project", "project_config.folder_path", "folder_path"),
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
