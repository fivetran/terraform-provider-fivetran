package fivetran_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceDbtProjectE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() {},
		ProviderFactories: providerFactory,
		CheckDestroy:      testFivetranDbtProjectResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "fivetran_group" "test_group" {
						provider = fivetran-provider
						name = "dbt_test_group"
					}

					resource "fivetran_dbt_project" "test_project" {
						group_id = fivetran_group.test_group.id
						dbt_version = "1.0.1"
						threads = 1
						default_schema = "dbt_demo_test_e2e_terraform"
						type = "GIT"
						project_config {
							git_remote_url = "git@github.com:fivetran/dbt_demo.git"
							git_branch = "main"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranDbtProjectResourceCreate(t, "fivetran_dbt_project.test_project"),

					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "dbt_version", "1.0.1"),
					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "threads", "1"),
					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "default_schema", "dbt_demo_test_e2e_terraform"),
					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "type", "GIT"),
					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "project_config.0.git_remote_url", "git@github.com:fivetran/dbt_demo.git"),
					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "project_config.0.git_branch", "main"),
				),
			},
		},
	})
}

func testFivetranDbtProjectResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		response, err := client.NewDbtDetails().DbtProjectID(rs.Primary.ID).Do(context.Background())
		if err != nil {
			return err
		}

		if response.Code != "Success" {
			return errors.New("DBT Project " + rs.Primary.ID + " doesn't exist. Response code: " + response.Code)
		}

		return nil
	}
}

func testFivetranDbtProjectResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "fivetran_group" {
			response, err := client.NewGroupDetails().GroupID(rs.Primary.ID).Do(context.Background())
			if err.Error() != "status code: 404; expected: 200" {
				return err
			}
			if !strings.HasPrefix(response.Code, "NotFound") {
				return errors.New("Group " + rs.Primary.ID + " still exists. Response code: " + response.Code)
			}
		}
		if rs.Type == "fivetran_dbt_project" {
			response, err := client.NewDbtDetails().DbtProjectID(rs.Primary.ID).Do(context.Background())
			if err.Error() != "status code: 404; expected: 200" {
				return err
			}
			if !strings.HasPrefix(response.Code, "NotFound") {
				return errors.New("DBT Project " + rs.Primary.ID + " still exists. Response code: " + response.Code)
			}
		}
	}

	return nil
}
