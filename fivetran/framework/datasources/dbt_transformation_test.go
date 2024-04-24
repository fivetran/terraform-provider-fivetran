package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	tfmock.MockClient().Reset()

	dbtModelResponse := `
	{
		"id": "dbt_model_id",
		"model_name": "dbt_model_name",
		"scheduled": true
    }
	`

	dbtTransformationDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/dbt/transformations/transformation_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			dbtTransformationDataSourceMockData = tfmock.CreateMapFromJsonString(t, dbtTransformationResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", dbtTransformationDataSourceMockData), nil
		},
	)
	dbtModelDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/dbt/models/dbt_model_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, dbtModelResponse)), nil
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
				tfmock.AssertEqual(t, dbtTransformationDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, dbtTransformationDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "dbt_model_id", "dbt_model_id"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "dbt_model_name", "dbt_model_name"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "run_tests", "true"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "paused", "false"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "schedule.schedule_type", "TIME_OF_DAY"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "schedule.days_of_week.0", "MONDAY"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "schedule.interval", "15"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_transformation.transformation", "schedule.time_of_day", "12:00"),
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
