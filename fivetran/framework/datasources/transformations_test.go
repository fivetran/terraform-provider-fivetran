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
	transformationsDataSourceMockGetHandler *mock.Handler
	transformationsDataSourceMockData       map[string]interface{}
)

const (
	transformationsMappingResponse = `
	{
		"items": [
      {
        "id": "transformation_id1",
        "status": "status1",
        "schedule": {
          "cron": [
            "cron1"
          ],
          "interval": 601,
          "smart_syncing": true,
          "connection_ids": [
            "connection_id1"
          ],
          "schedule_type": "schedule_type1",
          "days_of_week": [
            "days_of_week01",
            "days_of_week11"
          ],
          "time_of_day": "time_of_day1"
        },
        "type": "type1",
        "paused": true,
        "created_at": "created_at1",
        "output_model_names": [
          "output_model_name1"
        ],
        "created_by_id": "created_by_id1",
        "transformation_config": {
          "project_id": "project_id1",
          "name": "name1",
          "steps": [
            {
              "name": "name01",
              "command": "command01"
            },
            {
              "name": "name02",
              "command": "command02"
            }
          ]
        }
      },
{
    "id": "transformation_id2",
    "status": "status2",
    "schedule": {
      "cron": [
        "cron2"
      ],
      "interval": 602,
      "smart_syncing": true,
      "connection_ids": [
        "connection_id2"
      ],
      "schedule_type": "schedule_type2",
      "days_of_week": [
        "days_of_week02",
        "days_of_week12"
      ],
      "time_of_day": "time_of_day2"
    },
    "type": "type2",
    "paused": true,
    "created_at": "created_at2",
    "output_model_names": [
      "output_model_name2"
    ],
    "created_by_id": "created_by_id2",
    "transformation_config": {
      "package_name": "package_name2",
      "connection_ids": [
        "connection_id2"
      ],
      "excluded_models": [
        "excluded_model2"
      ],
      "upgrade_available": true
    }
  }
    ],
		"next_cursor": null	
    }
	`
)

func setupMockClienttransformationsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()
	transformationsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/transformations").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			transformationsDataSourceMockData = tfmock.CreateMapFromJsonString(t, transformationsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", transformationsDataSourceMockData), nil
		},
	)
}

func TestDataSourcetransformationsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_transformations" "transformation" {
			provider = fivetran-provider
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, transformationsDataSourceMockGetHandler.Interactions, 1)
				return nil
			},
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.id", "transformation_id1"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.status", "status1"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.created_at", "created_at1"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.created_by_id", "created_by_id1"),
			resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.type", "type1"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.paused", "true"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.output_model_names.0", "output_model_name1"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.transformation_config.project_id", "project_id1"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.transformation_config.name", "name1"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.transformation_config.steps.0.name", "name01"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.transformation_config.steps.0.command", "command01"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.transformation_config.steps.1.name", "name02"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.transformation_config.steps.1.command", "command02"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.schedule.smart_syncing", "true"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.schedule.interval", "601"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.schedule.schedule_type", "schedule_type1"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.schedule.cron.0", "cron1"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.schedule.connection_ids.0", "connection_id1"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.schedule.days_of_week.0", "days_of_week01"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.schedule.days_of_week.1", "days_of_week11"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.0.schedule.time_of_day", "time_of_day1"),

      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.id", "transformation_id2"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.status", "status2"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.created_at", "created_at2"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.created_by_id", "created_by_id2"),
			resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.type", "type2"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.paused", "true"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.output_model_names.0", "output_model_name2"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.transformation_config.package_name", "package_name2"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.transformation_config.connection_ids.0", "connection_id2"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.transformation_config.excluded_models.0", "excluded_model2"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.transformation_config.upgrade_available", "true"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.schedule.smart_syncing", "true"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.schedule.interval", "602"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.schedule.schedule_type", "schedule_type2"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.schedule.cron.0", "cron2"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.schedule.connection_ids.0", "connection_id2"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.schedule.days_of_week.0", "days_of_week02"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.schedule.days_of_week.1", "days_of_week12"),
      resource.TestCheckResourceAttr("data.fivetran_transformations.transformation", "transformations.1.schedule.time_of_day", "time_of_day2"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClienttransformationsDataSourceConfigMapping(t)
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
