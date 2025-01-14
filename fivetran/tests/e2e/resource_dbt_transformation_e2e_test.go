package e2e_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceDbtTransformationE2E(t *testing.T) {
	t.Skip("Skip cause deprecated in upstream.")
	destinationConfig := `
	resource "fivetran_destination" "test_destination" {
		provider = fivetran-provider
		group_id = "` + PredefinedGroupId + `"
		service = "big_query"
		region = "GCP_US_EAST4"
		time_zone_offset = "-5"
		config {
			project_id = "` + BqProjectId + `"
			data_set_location = "US"
		}
	}
	`
	projectConfig := `
	resource "fivetran_dbt_project" "test_project" {
		provider = fivetran-provider
		group_id = fivetran_destination.test_destination.id
		dbt_version = "1.3.2"
		threads = 1
		default_schema = "dbt_demo_test_e2e_terraform"
		type = "GIT"
		project_config {
			git_remote_url = "git@github.com:fivetran/dbt_demo.git"
			git_branch = "main"
		}
	}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranDbtTransformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: destinationConfig + projectConfig + `
				resource "fivetran_dbt_transformation" "test_transformation" {
					provider = fivetran-provider
					dbt_project_id = fivetran_dbt_project.test_project.id
					dbt_model_name = "statistics"
					paused = true
					run_tests = false
					schedule {
					   schedule_type = "INTERVAL"
					   days_of_week = ["MONDAY"]
					   interval = 60
					}
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "status", "READY"),
					resource.TestCheckResourceAttr("fivetran_dbt_transformation.test_transformation", "paused", "true"),
					resource.TestCheckResourceAttr("fivetran_dbt_transformation.test_transformation", "run_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_dbt_transformation.test_transformation", "schedule.schedule_type", "INTERVAL"),
					resource.TestCheckResourceAttr("fivetran_dbt_transformation.test_transformation", "schedule.interval", "60"),
					resource.TestCheckResourceAttr("fivetran_dbt_transformation.test_transformation", "schedule.days_of_week.0", "MONDAY"),
				),
			},
		},
	})
}

func testFivetranDbtTransformationResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "fivetran_dbt_project" {
			response, err := client.NewDbtProjectDetails().DbtProjectID(rs.Primary.ID).Do(context.Background())
			if err.Error() != "status code: 404; expected: 200" {
				return err
			}
			if !strings.HasPrefix(response.Code, "NotFound") {
				return errors.New("DBT Project " + rs.Primary.ID + " still exists. Response code: " + response.Code)
			}
		}
		if rs.Type == "fivetran_destination" {
			response, err := client.NewDestinationDetails().DestinationID(rs.Primary.ID).Do(context.Background())
			if err.Error() != "status code: 404; expected: 200" {
				return err
			}
			if !strings.HasPrefix(response.Code, "NotFound") {
				return errors.New("Destination " + rs.Primary.ID + " still exists. Response code: " + response.Code)
			}
		}
		if rs.Type == "fivetran_dbt_transformation" {
			response, err := client.NewDbtTransformationDetailsService().TransformationId(rs.Primary.ID).Do(context.Background())
			if err.Error() != "status code: 404; expected: 200" {
				return err
			}
			if !strings.HasPrefix(response.Code, "NotFound") {
				return errors.New("Transformation " + rs.Primary.ID + " still exists. Response code: " + response.Code)
			}
		}
	}

	return nil
}
