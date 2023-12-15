package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	dbtTransformationResponse = `
	{
		"id": "transformation_id",
		"status": "PENDING",
		"schedule": {
			"schedule_type": "TIME_OF_DAY",
			"days_of_week": ["MONDAY"],
			"interval": 15,
			"time_of_day": "12:00"
		},
		"last_run": "2023-01-01T00:00:00.743708Z",
		"run_tests": true,
		"model_ids": ["model_id_1"],
		"output_model_name": "output_model_name",
		"dbt_project_id": "dbt_project_id",
		"dbt_model_id": "dbt_model_id",
		"connector_ids": ["connector_id_1"],
		"next_run": "2023-01-02T00:00:00.743708Z",
		"created_at": "2023-01-02T00:00:00.743708Z",
		"paused": false
	}
	`
)

var (
	dbtTransformationDataSourceMockGetHandler *mock.Handler
	dbtModelDataSourceMockGetHandler          *mock.Handler

	dbtTransformationDataSourceMockData map[string]interface{}
)

func setupMockClientDbtTransformationDataSourceMappingTest(t *testing.T) {
	mockClient.Reset()

	dbtModelResponse := `
	{
		"id": "dbt_model_id",
		"model_name": "dbt_model_name",
		"scheduled": true
    }
	`

	dbtTransformationDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/dbt/transformations/transformation_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			dbtTransformationDataSourceMockData = createMapFromJsonString(t, dbtTransformationResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", dbtTransformationDataSourceMockData), nil
		},
	)
	dbtModelDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/dbt/models/dbt_model_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, dbtModelResponse)), nil
		},
	)
}

func TestDataSourceDbtTranformationMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_dbt_transformation" "transformation" {
			provider = fivetran-provider
			id = "transformation_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, dbtTransformationDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, dbtTransformationDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "dbt_model_id", "dbt_model_id"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "dbt_model_name", "dbt_model_name"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "run_tests", "true"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "paused", "false"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "schedule.0.schedule_type", "TIME_OF_DAY"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "schedule.0.days_of_week.0", "MONDAY"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "schedule.0.interval", "15"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "schedule.0.time_of_day", "12:00"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "dbt_project_id", "dbt_project_id"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "output_model_name", "output_model_name"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "created_at", "2023-01-02T00:00:00.743708Z"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "connector_ids.0", "connector_id_1"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "model_ids.0", "model_id_1"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientDbtTransformationDataSourceMappingTest(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
