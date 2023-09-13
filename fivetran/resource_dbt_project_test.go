package fivetran_test

import (
	"context"
	"errors"
	"strconv"
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
					resource "fivetran_dbt_project" "test_project" {
						provider = fivetran-provider
						group_id = "` + PredefinedGroupId + `"
						dbt_version = "1.0.1"
						threads = 1
						default_schema = "dbt_demo_test_e2e_terraform"
						type = "GIT"
						project_config {
							folder_path = "/folder/path"
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
					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "project_config.0.folder_path", "/folder/path"),
				),
			},
			{
				Config: `
					resource "fivetran_dbt_project" "test_project" {
						provider = fivetran-provider
						group_id = "` + PredefinedGroupId + `"
						dbt_version = "1.0.0"
						threads = 2
						target_name = "target_name"
						default_schema = "dbt_demo_test_e2e_terraform"
						type = "GIT"
						project_config {
							folder_path = "/folder/path_1"
							git_remote_url = "git@github.com:fivetran/dbt_demo.git"
							git_branch = "not_main"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranDbtProjectResourceUpdate(t, "fivetran_dbt_project.test_project"),

					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "dbt_version", "1.0.0"),
					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "target_name", "target_name"),
					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "threads", "2"),
					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "project_config.0.git_branch", "not_main"),
					resource.TestCheckResourceAttr("fivetran_dbt_project.test_project", "project_config.0.folder_path", "/folder/path_1"),
				),
			},
		},
	})
}

func testFivetranDbtProjectResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		response, err := client.NewDbtProjectDetails().DbtProjectID(rs.Primary.ID).Do(context.Background())
		if err != nil {
			return err
		}

		if response.Code != "Success" {
			return errors.New("DBT Project " + rs.Primary.ID + " doesn't exist. Response code: " + response.Code)
		}

		if response.Data.DefaultSchema != "dbt_demo_test_e2e_terraform" {
			return errors.New("DBT Project " + rs.Primary.ID + " has wrong default_schema value. Actual: " + response.Data.DefaultSchema + " Expected: dbt_demo_test_e2e_terraform")
		}
		if response.Data.DbtVersion != "1.0.1" {
			return errors.New("DBT Project " + rs.Primary.ID + " has wrong dbt_version value. Actual: " + response.Data.DbtVersion + " Expected: 1.0.1")
		}
		if response.Data.Threads != 1 {
			return errors.New("DBT Project " + rs.Primary.ID + " has wrong threads value. Actual: " + strconv.Itoa(response.Data.Threads) + " Expected: 1")
		}
		if response.Data.Type != "GIT" {
			return errors.New("DBT Project " + rs.Primary.ID + " has wrong type value. Actual: " + response.Data.Type + " Expected: GIT")
		}
		if response.Data.ProjectConfig.GitRemoteUrl != "git@github.com:fivetran/dbt_demo.git" {
			return errors.New("DBT Project " + rs.Primary.ID + " has wrong project_config.git_remote_url value. Actual: " + response.Data.ProjectConfig.GitRemoteUrl + " Expected: git@github.com:fivetran/dbt_demo.git")
		}
		if response.Data.ProjectConfig.GitBranch != "main" {
			return errors.New("DBT Project " + rs.Primary.ID + " has wrong project_config.git_branch value. Actual: " + response.Data.ProjectConfig.GitBranch + " Expected: main")
		}
		if response.Data.ProjectConfig.FolderPath != "/folder/path" {
			return errors.New("DBT Project " + rs.Primary.ID + " has wrong project_config.folder_path value. Actual: " + response.Data.ProjectConfig.FolderPath + " Expected: /folder/path")
		}
		return nil
	}
}

func testFivetranDbtProjectResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		response, err := client.NewDbtProjectDetails().DbtProjectID(rs.Primary.ID).Do(context.Background())
		if err != nil {
			return err
		}

		if response.Code != "Success" {
			return errors.New("DBT Project " + rs.Primary.ID + " doesn't exist. Response code: " + response.Code)
		}
		if response.Data.DbtVersion != "1.0.0" {
			return errors.New("DBT Project " + rs.Primary.ID + " has wrong dbt_version value. Actual: " + response.Data.DbtVersion + " Expected: 1.0.0")
		}
		if response.Data.Threads != 2 {
			return errors.New("DBT Project " + rs.Primary.ID + " has wrong threads value. Actual: " + strconv.Itoa(response.Data.Threads) + " Expected: 2")
		}
		if response.Data.ProjectConfig.GitBranch != "not_main" {
			return errors.New("DBT Project " + rs.Primary.ID + " has wrong project_config.git_branch value. Actual: " + response.Data.ProjectConfig.GitBranch + " Expected: not_main")
		}
		if response.Data.ProjectConfig.FolderPath != "/folder/path_1" {
			return errors.New("DBT Project " + rs.Primary.ID + " has wrong project_config.folder_path value. Actual: " + response.Data.ProjectConfig.FolderPath + " Expected: /folder/path_1")
		}
		return nil
	}
}

func testFivetranDbtProjectResourceDestroy(s *terraform.State) error {
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
	}

	return nil
}
