package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	transformationDataSourceMockGetHandler *mock.Handler
	transformationDataSourceMockData       map[string]interface{}
)

func setupMockClienttransformationDataSourceMappingTest(t *testing.T) {
	transformationResponse := `
{
    "id": "transformation_id",
    "status": "status",
    "schedule": {
      "cron": [
        "cron1",
        "cron2"
      ],
      "interval": 60,
      "smart_syncing": true,
      "connection_ids": [
        "connection_id1",
        "connection_id2"
      ],
      "schedule_type": "schedule_type",
      "days_of_week": [
        "days_of_week1",
        "days_of_week2"
      ],
      "time_of_day": "time_of_day"
    },
    "type": "type",
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
      ],
      "package_name": "package_name",
      "connection_ids": [
        "connection_id1",
        "connection_id2"
      ],
      "excluded_models": [
        "excluded_model1",
        "excluded_model2"
      ],
      "upgrade_available": true
    }
  }`
	tfmock.MockClient().Reset()

	transformationDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/transformations/transformation_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			transformationDataSourceMockData = tfmock.CreateMapFromJsonString(t, transformationResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", transformationDataSourceMockData), nil
		},
	)
}

func TestDataSourcetransformationMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_transformation" "transformation" {
			provider = fivetran-provider
			id = "transformation_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, transformationDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, transformationDataSourceMockData)
				return nil
			},
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "id", "transformation_id"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "status", "status"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "created_at", "created_at"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "created_by_id", "created_by_id"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "type", "type"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "paused", "true"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "output_model_names.0", "output_model_name1"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "output_model_names.1", "output_model_name2"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.project_id", "project_id"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.name", "name"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.steps.0.name", "name1"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.steps.0.command", "command1"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.steps.1.name", "name2"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.steps.1.command", "command2"),
						resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.package_name", "package_name"),
						resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.connection_ids.0", "connection_id1"),
						resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.connection_ids.1", "connection_id2"),
						resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.excluded_models.0", "excluded_model1"),
						resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.excluded_models.1", "excluded_model2"),
						resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "transformation_config.upgrade_available", "true"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "schedule.smart_syncing", "true"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "schedule.interval", "60"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "schedule.schedule_type", "schedule_type"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "schedule.cron.0", "cron1"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "schedule.cron.1", "cron2"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "schedule.connection_ids.0", "connection_id1"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "schedule.connection_ids.1", "connection_id2"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "schedule.days_of_week.0", "days_of_week1"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "schedule.days_of_week.1", "days_of_week2"),
            resource.TestCheckResourceAttr("data.fivetran_transformation.transformation", "schedule.time_of_day", "time_of_day"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClienttransformationDataSourceMappingTest(t)
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
