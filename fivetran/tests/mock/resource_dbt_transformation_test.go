package mock

import (
	"net/http"
	"testing"
	"time"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	transformationPostHandler   *mock.Handler
	transformationPatchHandler  *mock.Handler
	transformationDeleteHandler *mock.Handler
	transformationData          map[string]interface{}
)

func onPostTranformation(t *testing.T, req *http.Request) (*http.Response, error) {
	assertEmpty(t, userData)

	body := requestBodyToJson(t, req)

	// Check the request
	assertEqual(t, len(body), 4)

	assertKeyExistsAndHasValue(t, body, "dbt_model_id", "dbt_model_id")
	assertKeyExistsAndHasValue(t, body, "paused", false)
	assertKeyExistsAndHasValue(t, body, "run_tests", false)

	requestSchedule := assertKeyExists(t, body, "schedule").(map[string]interface{})

	assertKeyExistsAndHasValue(t, requestSchedule, "schedule_type", "TIME_OF_DAY")
	assertKeyExistsAndHasValue(t, requestSchedule, "time_of_day", "12:00")

	requestScheduleDays := assertKeyExists(t, requestSchedule, "days_of_week").([]interface{})

	expectedDays := make([]interface{}, 0)

	expectedDays = append(expectedDays, "MONDAY")
	//expectedDays = append(expectedDays, "SATURDAY")

	assertArrayItems(t, requestScheduleDays, expectedDays)

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

	response := fivetranSuccessResponse(t, req, http.StatusCreated, "", transformationData)

	return response, nil
}

func onPatchTransformation(t *testing.T, req *http.Request, updateIteration int) (*http.Response, error) {
	assertNotEmpty(t, transformationData)

	body := requestBodyToJson(t, req)

	if updateIteration == 0 {
		// Check the request
		assertEqual(t, len(body), 3)
		assertKeyExistsAndHasValue(t, body, "paused", true)
		assertKeyExistsAndHasValue(t, body, "run_tests", true)
		requestSchedule := assertKeyExists(t, body, "schedule").(map[string]interface{})

		requestScheduleDays := assertKeyExists(t, requestSchedule, "days_of_week").([]interface{})
		expectedDays := make([]interface{}, 0)
		expectedDays = append(expectedDays, "MONDAY")
		expectedDays = append(expectedDays, "SATURDAY")

		assertArrayItems(t, requestScheduleDays, expectedDays)

		// Update saved values
		for k, v := range body {
			if k != "schedule" {
				transformationData[k] = v
			} else {
				stateSchedule := transformationData[k].(map[string]interface{})
				stateSchedule["days_of_week"] = expectedDays
			}
		}

		response := fivetranSuccessResponse(t, req, http.StatusOK, "Transformation has been updated", transformationData)
		return response, nil
	}

	if updateIteration == 1 {
		// Check the request
		assertEqual(t, len(body), 1)
		schedule := assertKeyExists(t, body, "schedule").(map[string]interface{})
		assertKeyExistsAndHasValue(t, schedule, "schedule_type", "INTERVAL")
		assertKeyExistsAndHasValue(t, schedule, "interval", float64(60))

		// Update saved values
		for k, v := range body {
			transformationData[k] = v
		}

		response := fivetranSuccessResponse(t, req, http.StatusOK, "Transformation has been updated", transformationData)
		return response, nil
	}

	response := fivetranSuccessResponse(t, req, http.StatusOK, "", transformationData)

	return response, nil
}

func setupMockClientTransformationResource(t *testing.T) {
	mockClient.Reset()
	transformationData = nil
	updateCounter := 0

	transformationPostHandler = mockClient.When(http.MethodPost, "/v1/dbt/transformations").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPostTranformation(t, req)
		},
	)

	mockClient.When(http.MethodGet, "/v1/dbt/transformations/transformation_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			assertNotEmpty(t, transformationData)
			response := fivetranSuccessResponse(t, req, http.StatusOK, "", transformationData)
			return response, nil
		},
	)

	transformationPatchHandler = mockClient.When(http.MethodPatch, "/v1/dbt/transformations/transformation_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			response, err := onPatchTransformation(t, req, updateCounter)
			updateCounter++
			return response, err
		},
	)

	transformationDeleteHandler = mockClient.When(http.MethodDelete, "/v1/dbt/transformations/transformation_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			assertNotEmpty(t, transformationData)
			transformationData = nil
			response := fivetranSuccessResponse(t, req, 200, "", nil)
			return response, nil
		},
	)
}

func TestResourceTransformationMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_dbt_transformation" "transformation" {
			provider = fivetran-provider

			dbt_model_id = "dbt_model_id"
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
				assertEqual(t, transformationPostHandler.Interactions, 1)
				assertNotEmpty(t, transformationData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "id", "transformation_id"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "dbt_model_id", "dbt_model_id"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "run_tests", "false"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "paused", "false"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.0.schedule_type", "TIME_OF_DAY"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.0.time_of_day", "12:00"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.0.days_of_week.0", "MONDAY"),
		),
	}

	// Update run_tests and paused fields, update days of week in schedule
	step2 := resource.TestStep{
		Config: `
		resource "fivetran_dbt_transformation" "transformation" {
			provider = fivetran-provider

			dbt_model_id = "dbt_model_id"
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
				assertEqual(t, transformationPatchHandler.Interactions, 1)
				assertNotEmpty(t, transformationData)
				return nil
			},

			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "run_tests", "true"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "paused", "true"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.0.days_of_week.1", "SATURDAY"),
		),
	}

	// Update schedule_type and paused fields
	step3 := resource.TestStep{
		Config: `
		resource "fivetran_dbt_transformation" "transformation" {
			provider = fivetran-provider

			dbt_model_id = "dbt_model_id"
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
				assertEqual(t, transformationPatchHandler.Interactions, 2)
				assertNotEmpty(t, transformationData)
				return nil
			},

			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.0.schedule_type", "INTERVAL"),
			resource.TestCheckResourceAttr("fivetran_dbt_transformation.transformation", "schedule.0.interval", "60"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTransformationResource(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, transformationDeleteHandler.Interactions, 1)
				assertEmpty(t, transformationData)
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
