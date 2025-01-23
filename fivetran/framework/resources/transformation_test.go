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
    transformationGitPostHandler            *mock.Handler
    transformationQuickstartPostHandler     *mock.Handler
    transformationGitData                   map[string]interface{}
    transformationQuickstartData            map[string]interface{}

    transformationGitDeleteHandler          *mock.Handler
    transformationQuickstartDeleteHandler   *mock.Handler

    gitResponse = `{
    "id": "transformation_id",
    "status": "status",
    "schedule": {
      "cron": [
        "cron1","cron2"
      ],
      "interval": 601,
      "smart_syncing": true,
      "connection_ids": [
        "connection_id1",
        "connection_id2"
      ],
      "schedule_type": "schedule_type1",
      "days_of_week": [
        "days_of_week1",
        "days_of_week2"
      ],
      "time_of_day": "time_of_day1"
    },
    "type": "DBT_CORE",
    "paused": true,
    "created_at": "created_at",
    "output_model_names": [
      "output_model_name1",
      "output_model_name2"
    ],
    "created_by_id": "created_by_id",
    "transformation_config": {
      "project_id": "project_id",
      "name": "name",
      "steps": [
        {
          "name": "name1",
          "command": "command1"
        },
        {
          "name": "name2",
          "command": "command2"
        }
      ]
    }
  }`

 quickstartResponse = `{
    "id": "transformation_id",
    "status": "status",
    "schedule": {
      "cron": [
        "cron1","cron2"
      ],
      "interval": 601,
      "smart_syncing": true,
      "connection_ids": [
        "connection_id1",
        "connection_id2"
      ],
      "schedule_type": "schedule_type1",
      "days_of_week": [
        "days_of_week1",
        "days_of_week2"
      ],
      "time_of_day": "time_of_day1"
    },
    "type": "QUICKSTART",
    "paused": true,
    "created_at": "created_at",
    "output_model_names": [
      "output_model_name1",
      "output_model_name2"
    ],
    "created_by_id": "created_by_id",
    "transformation_config": {
      "package_name": "package_name",
      "connection_ids": [
        "connection_id1",
        "connection_id2"
      ],
      "excluded_models": [
        "excluded_model1","excluded_model2"
      ],
      "upgrade_available": true
    }
  }`
)

func setupMockClientTransformationGitResource(t *testing.T) {
    tfmock.MockClient().Reset()

    transformationGitPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/transformations").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            body := tfmock.RequestBodyToJson(t, req)
            tfmock.AssertKeyExistsAndHasValue(t, body, "type", "DBT_CORE")
            tfmock.AssertKeyExistsAndHasValue(t, body, "paused", true)

            tfmock.AssertKeyExists(t, body, "transformation_config")
            config := body["transformation_config"].(map[string]interface{})
            tfmock.AssertKeyExistsAndHasValue(t, config, "project_id", "project_id")
            tfmock.AssertKeyExistsAndHasValue(t, config, "name", "name")

            steps := config["steps"].([]interface{})
            tfmock.AssertKeyExistsAndHasValue(t, steps[0].(map[string]interface{}), "name", "name1")
            tfmock.AssertKeyExistsAndHasValue(t, steps[0].(map[string]interface{}), "command", "command1")
            tfmock.AssertKeyExistsAndHasValue(t, steps[1].(map[string]interface{}), "name", "name2")
            tfmock.AssertKeyExistsAndHasValue(t, steps[1].(map[string]interface{}), "command", "command2")

            tfmock.AssertKeyExists(t, body, "schedule")
            schedule := body["schedule"].(map[string]interface{})
            tfmock.AssertKeyExistsAndHasValue(t, schedule, "interval", float64(601))
            tfmock.AssertKeyExistsAndHasValue(t, schedule, "smart_syncing", true)
            tfmock.AssertKeyExistsAndHasValue(t, schedule, "schedule_type", "schedule_type1")
            tfmock.AssertKeyExistsAndHasValue(t, schedule, "time_of_day", "time_of_day1")

            cron := schedule["cron"].([]interface{})
            tfmock.AssertEqual(t, len(cron), 2)
            tfmock.AssertEqual(t, cron[0], "cron1")
            tfmock.AssertEqual(t, cron[1], "cron2")

            connectionIds := schedule["connection_ids"].([]interface{})
            tfmock.AssertEqual(t, len(connectionIds), 2)
            tfmock.AssertEqual(t, connectionIds[0], "connection_id1")
            tfmock.AssertEqual(t, connectionIds[1], "connection_id2")

            daysOfWeek := schedule["days_of_week"].([]interface{})
            tfmock.AssertEqual(t, len(daysOfWeek), 2)
            tfmock.AssertEqual(t, daysOfWeek[0], "days_of_week1")
            tfmock.AssertEqual(t, daysOfWeek[1], "days_of_week2")

            transformationGitData = tfmock.CreateMapFromJsonString(t, gitResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", transformationGitData), nil
        },
    )

    tfmock.MockClient().When(http.MethodGet, "/v1/transformations/transformation_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            tfmock.AssertNotEmpty(t, transformationGitData)
            response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", transformationGitData)
            return response, nil
        },
    )

    transformationGitDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/transformations/transformation_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            tfmock.AssertNotEmpty(t, transformationGitData)
            transformationGitData = nil
            response := tfmock.FivetranSuccessResponse(t, req, 200, "", nil)
            return response, nil
        },
    )
}

func setupMockClientTransformationQuickstartResource(t *testing.T) {
    tfmock.MockClient().Reset()

    transformationQuickstartPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/transformations").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            body := tfmock.RequestBodyToJson(t, req)
            tfmock.AssertKeyExistsAndHasValue(t, body, "type", "QUICKSTART")
            tfmock.AssertKeyExistsAndHasValue(t, body, "paused", true)

            tfmock.AssertKeyExists(t, body, "transformation_config")
            config := body["transformation_config"].(map[string]interface{})
            tfmock.AssertKeyExistsAndHasValue(t, config, "package_name", "package_name")
            connectionIds := config["connection_ids"].([]interface{})
            tfmock.AssertEqual(t, len(connectionIds), 2)
            tfmock.AssertEqual(t, connectionIds[0], "connection_id1")
            tfmock.AssertEqual(t, connectionIds[1], "connection_id2")
            excludedModels := config["excluded_models"].([]interface{})
            tfmock.AssertEqual(t, len(excludedModels), 2)
            tfmock.AssertEqual(t, excludedModels[0], "excluded_model1")
            tfmock.AssertEqual(t, excludedModels[1], "excluded_model2")

            tfmock.AssertKeyExists(t, body, "schedule")
            schedule := body["schedule"].(map[string]interface{})
            tfmock.AssertKeyExistsAndHasValue(t, schedule, "interval", float64(601))
            tfmock.AssertKeyExistsAndHasValue(t, schedule, "smart_syncing", true)
            tfmock.AssertKeyExistsAndHasValue(t, schedule, "schedule_type", "schedule_type1")
            tfmock.AssertKeyExistsAndHasValue(t, schedule, "time_of_day", "time_of_day1")

            cron := schedule["cron"].([]interface{})
            tfmock.AssertEqual(t, len(cron), 2)
            tfmock.AssertEqual(t, cron[0], "cron1")
            tfmock.AssertEqual(t, cron[1], "cron2")

            connectionIds = schedule["connection_ids"].([]interface{})
            tfmock.AssertEqual(t, len(connectionIds), 2)
            tfmock.AssertEqual(t, connectionIds[0], "connection_id1")
            tfmock.AssertEqual(t, connectionIds[1], "connection_id2")

            daysOfWeek := schedule["days_of_week"].([]interface{})
            tfmock.AssertEqual(t, len(daysOfWeek), 2)
            tfmock.AssertEqual(t, daysOfWeek[0], "days_of_week1")
            tfmock.AssertEqual(t, daysOfWeek[1], "days_of_week2")

            transformationQuickstartData = tfmock.CreateMapFromJsonString(t, quickstartResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", transformationQuickstartData), nil
        },
    )

    tfmock.MockClient().When(http.MethodGet, "/v1/transformations/transformation_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            tfmock.AssertNotEmpty(t, transformationQuickstartData)
            response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", transformationQuickstartData)
            return response, nil
        },
    )

    transformationQuickstartDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/transformations/transformation_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            tfmock.AssertNotEmpty(t, transformationQuickstartData)
            transformationQuickstartData = nil
            response := tfmock.FivetranSuccessResponse(t, req, 200, "", nil)
            return response, nil
        },
    )
}

func TestResourceTransformationGitMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        resource "fivetran_transformation" "transformation" {
            provider = fivetran-provider

            type = "DBT_CORE"
            paused = true

            schedule {
                cron = ["cron1","cron2"]
                interval = 601
                smart_syncing = true
                connection_ids = ["connection_id1", "connection_id2"]
                schedule_type = "schedule_type1"
                days_of_week = ["days_of_week1","days_of_week2"]
                time_of_day = "time_of_day1"
            }

            transformation_config {
                project_id = "project_id"
                name = "name"
                steps = [
                    {
                        name = "name1"
                        command = "command1"
                    },
                    {
                        name = "name2"
                        command = "command2"
                    }
                ]
            }
        }
        `,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, transformationGitPostHandler.Interactions, 1)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "id", "transformation_id"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "status", "status"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "created_at", "created_at"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "created_by_id", "created_by_id"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "type", "DBT_CORE"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "paused", "true"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "output_model_names.0", "output_model_name1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "output_model_names.1", "output_model_name2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.project_id", "project_id"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.name", "name"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.steps.0.name", "name1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.steps.0.command", "command1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.steps.1.name", "name2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.steps.1.command", "command2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.smart_syncing", "true"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.interval", "601"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.schedule_type", "schedule_type1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.cron.0", "cron1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.cron.1", "cron2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.connection_ids.0", "connection_id1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.connection_ids.1", "connection_id2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.days_of_week.0", "days_of_week1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.days_of_week.1", "days_of_week2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.time_of_day", "time_of_day1"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientTransformationGitResource(t)
            },
            ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
            CheckDestroy: func(s *terraform.State) error {
                tfmock.AssertEqual(t, transformationGitDeleteHandler.Interactions, 1)
                tfmock.AssertEmpty(t, transformationData)
                return nil
            },

            Steps: []resource.TestStep{
                step1,
            },
        },
    )
}

func TestResourceTransformationQuickstartMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        resource "fivetran_transformation" "transformation" {
            provider = fivetran-provider

            type = "QUICKSTART"
            paused = true

            schedule {
                cron = ["cron1","cron2"]
                interval = 601
                smart_syncing = true
                connection_ids = ["connection_id1", "connection_id2"]
                schedule_type = "schedule_type1"
                days_of_week = ["days_of_week1","days_of_week2"]
                time_of_day = "time_of_day1"
            }

            transformation_config {
                package_name = "package_name"
                connection_ids = ["connection_id1", "connection_id2"]
                excluded_models = ["excluded_model1", "excluded_model2"]
            }
        }
        `,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, transformationQuickstartPostHandler.Interactions, 1)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "id", "transformation_id"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "status", "status"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "created_at", "created_at"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "created_by_id", "created_by_id"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "type", "QUICKSTART"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "paused", "true"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "output_model_names.0", "output_model_name1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "output_model_names.1", "output_model_name2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.package_name", "package_name"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.connection_ids.0", "connection_id1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.connection_ids.1", "connection_id2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.excluded_models.0", "excluded_model1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.excluded_models.1", "excluded_model2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "transformation_config.upgrade_available", "true"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.smart_syncing", "true"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.interval", "601"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.schedule_type", "schedule_type1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.cron.0", "cron1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.cron.1", "cron2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.connection_ids.0", "connection_id1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.connection_ids.1", "connection_id2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.days_of_week.0", "days_of_week1"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.days_of_week.1", "days_of_week2"),
            resource.TestCheckResourceAttr("fivetran_transformation.transformation", "schedule.time_of_day", "time_of_day1"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientTransformationQuickstartResource(t)
            },
            ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
            CheckDestroy: func(s *terraform.State) error {
                tfmock.AssertEqual(t, transformationQuickstartDeleteHandler.Interactions, 1)
                return nil
            },

            Steps: []resource.TestStep{
                step1,
            },
        },
    )
}