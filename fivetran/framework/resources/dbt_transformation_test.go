package resources_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	transformationPostHandler   *mock.Handler
	transformationPatchHandler  *mock.Handler
	transformationDeleteHandler *mock.Handler
	transformationData          map[string]interface{}
)

func onPostTranformation(t *testing.T, req *http.Request) (*http.Response, error) {
	tfmock.AssertEmpty(t, transformationData)

	body := tfmock.RequestBodyToJson(t, req)

	// Check the request
	tfmock.AssertEqual(t, len(body), 4)

	tfmock.AssertKeyExistsAndHasValue(t, body, "dbt_model_id", "dbt_model_id")
	tfmock.AssertKeyExistsAndHasValue(t, body, "paused", false)
	tfmock.AssertKeyExistsAndHasValue(t, body, "run_tests", false)

	requestSchedule := tfmock.AssertKeyExists(t, body, "schedule").(map[string]interface{})

	tfmock.AssertKeyExistsAndHasValue(t, requestSchedule, "schedule_type", "TIME_OF_DAY")
	tfmock.AssertKeyExistsAndHasValue(t, requestSchedule, "time_of_day", "12:00")

	requestScheduleDays := tfmock.AssertKeyExists(t, requestSchedule, "days_of_week").([]interface{})

	expectedDays := make([]interface{}, 0)

	expectedDays = append(expectedDays, "MONDAY")
	//expectedDays = append(expectedDays, "SATURDAY")

	tfmock.AssertArrayItems(t, requestScheduleDays, expectedDays)

	// Add response fields
	body["id"] = "transformation_id"
	body["dbt_project_id"] = "dbt_project_id"
	body["output_model_name"] = "output_model_name"

	connectorIds := make([]string, 0)
	body["connector_ids"] = append(connectorIds, "connector_id")

	modelIds := make([]string, 0)
	body["model_ids"] = append(modelIds, "model_id")

	body["created_at"] = time.Now().Format("2006-01-02T15:04:05.000000Z")

	transformationData = body

	response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "", transformationData)

	return response, nil
}

func onPatchTransformation(t *testing.T, req *http.Request, updateIteration int) (*http.Response, error) {
	tfmock.AssertNotEmpty(t, transformationData)

	body := tfmock.RequestBodyToJson(t, req)

	if updateIteration == 0 {
		// Check the request
		tfmock.AssertEqual(t, len(body), 3)
		tfmock.AssertKeyExistsAndHasValue(t, body, "paused", true)
		tfmock.AssertKeyExistsAndHasValue(t, body, "run_tests", true)
		requestSchedule := tfmock.AssertKeyExists(t, body, "schedule").(map[string]interface{})

		requestScheduleDays := tfmock.AssertKeyExists(t, requestSchedule, "days_of_week").([]interface{})
		expectedDays := make([]interface{}, 0)
		expectedDays = append(expectedDays, "MONDAY")
		expectedDays = append(expectedDays, "SATURDAY")

		tfmock.AssertArrayItems(t, requestScheduleDays, expectedDays)

		// Update saved values
		for k, v := range body {
			if k != "schedule" {
				transformationData[k] = v
			} else {
				stateSchedule := transformationData[k].(map[string]interface{})
				stateSchedule["days_of_week"] = expectedDays
			}
		}

		response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Transformation has been updated", transformationData)
		return response, nil
	}

	if updateIteration == 1 {
		// Check the request
		tfmock.AssertEqual(t, len(body), 1)
		schedule := tfmock.AssertKeyExists(t, body, "schedule").(map[string]interface{})
		tfmock.AssertKeyExistsAndHasValue(t, schedule, "schedule_type", "INTERVAL")
		tfmock.AssertKeyExistsAndHasValue(t, schedule, "interval", float64(60))

		// Update saved values
		for k, v := range body {
			transformationData[k] = v
		}

		response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Transformation has been updated", transformationData)
		return response, nil
	}

	response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", transformationData)

	return response, nil
}

func setupMockClientTransformationResource(t *testing.T) {
	tfmock.MockClient().Reset()
	transformationData = nil
	updateCounter := 0

	transformationPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/dbt/transformations").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPostTranformation(t, req)
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/dbt/transformations/transformation_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			tfmock.AssertNotEmpty(t, transformationData)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", transformationData)
			return response, nil
		},
	)

	transformationPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/dbt/transformations/transformation_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			response, err := onPatchTransformation(t, req, updateCounter)
			updateCounter++
			return response, err
		},
	)

	transformationDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/dbt/transformations/transformation_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			tfmock.AssertNotEmpty(t, transformationData)
			transformationData = nil
			response := tfmock.FivetranSuccessResponse(t, req, 200, "", nil)
			return response, nil
		},
	)

	projectResponse := `{
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
	}
	`

	tfmock.MockClient().When(http.MethodGet, "/v1/dbt/projects/dbt_project_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, projectResponse)), nil
		},
	)

	modelsMappingResponse := `
	{
		"items":[
			{
				"id": "dbt_model_id",
				"model_name": "dbt_model_name",
				"scheduled": true
			}
		],
		"next_cursor": null	
    }
	`

	modelMappingResponse := `
	{
		"id": "dbt_model_id",
		"model_name": "dbt_model_name",
		"scheduled": true
	}
	`

	tfmock.MockClient().When(http.MethodGet, "/v1/dbt/models/dbt_model_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, modelMappingResponse)), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/dbt/models").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			project_id := req.URL.Query().Get("project_id")
			tfmock.AssertEqual(t, project_id, "dbt_project_id")
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, modelsMappingResponse)), nil
		},
	)

}

func TestResourceTransformationMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_dbt_transformation" "transformation" {
			provider = fivetran-provider

			dbt_project_id = "dbt_project_id"
			dbt_model_name = "dbt_model_name"
			run_tests = "false"
			paused = "false"
			schedule {
				schedule_type = "TIME_OF_DAY"
				time_of_day = "12:00"
				days_of_week = ["MONDAY"]
			}
		}
		`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, transformationPostHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, transformationData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "id", "transformation_id"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "dbt_model_id", "dbt_model_id"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "run_tests", "false"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "paused", "false"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.schedule_type", "TIME_OF_DAY"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.time_of_day", "12:00"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.days_of_week.0", "MONDAY"),
		),
	}

	// Update run_tests and paused fields, update days of week in schedule
	step2 := resource.TestStep{
		Config: `
		resource "fivetran_dbt_transformation" "transformation" {
			provider = fivetran-provider

			dbt_project_id = "dbt_project_id"
			dbt_model_name = "dbt_model_name"
			run_tests = "true"
			paused = "true"
			schedule {
				schedule_type = "TIME_OF_DAY"
				time_of_day = "12:00"
				days_of_week = ["MONDAY", "SATURDAY"]
			}
		}
		`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, transformationPatchHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, transformationData)
				return nil
			},

			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "run_tests", "true"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "paused", "true"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.days_of_week.1", "SATURDAY"),
		),
	}

	// Update schedule_type and paused fields
	step3 := resource.TestStep{
		Config: `
		resource "fivetran_dbt_transformation" "transformation" {
			provider = fivetran-provider

			dbt_project_id = "dbt_project_id"
			dbt_model_name = "dbt_model_name"
			run_tests = "true"
			paused = "true"
			schedule {
				schedule_type = "INTERVAL"
				interval = 60
			}
		}
		`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, transformationPatchHandler.Interactions, 2)
				tfmock.AssertNotEmpty(t, transformationData)
				return nil
			},

			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.schedule_type", "INTERVAL"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.interval", "60"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTransformationResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, transformationDeleteHandler.Interactions, 1)
				tfmock.AssertEmpty(t, transformationData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
			},
		},
	)
}
