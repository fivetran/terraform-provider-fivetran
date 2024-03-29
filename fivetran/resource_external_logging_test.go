package fivetran_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceExternalLoggingE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranExternalLoggingResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "testgroup" {
					provider = fivetran-provider
					name = "test_group_name"
			    }

				resource "fivetran_external_logging" "test_extlog" {
					provider = fivetran-provider

    				group_id = fivetran_group.testgroup.id
    				service = "azure_monitor_log"
    				enabled = "true"
    				run_setup_tests = "false"

				    config {
        				workspace_id = "workspace_id"
        				primary_key = "PASSWORD"
    				}
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranExternalLoggingResourceCreate(t, "fivetran_external_logging.test_extlog"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "service", "azure_monitor_log"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "enabled", "true"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.workspace_id", "workspace_id"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.primary_key", "PASSWORD"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "testgroup" {
					provider = fivetran-provider
					name = "test_group_name"
			    }

				resource "fivetran_external_logging" "test_extlog" {
					provider = fivetran-provider

    				group_id = fivetran_group.testgroup.id
    				service = "azure_monitor_log"
    				enabled = "true"
    				run_setup_tests = "false"

				    config {
        				workspace_id = "workspace_id_1"
        				primary_key = "PASSWORD"
    				}
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranExternalLoggingResourceUpdate(t, "fivetran_external_logging.test_extlog"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "service", "azure_monitor_log"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "enabled", "true"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.workspace_id", "workspace_id_1"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.primary_key", "PASSWORD"),
				),
			},
		},
	})
}

func testFivetranExternalLoggingResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewExternalLoggingDetails().ExternalLoggingId(rs.Primary.ID).Do(context.Background())
		if err != nil {
			fmt.Println(err)
			return err
		}
		//todo: check response _  fields if needed
		return nil
	}
}

func testFivetranExternalLoggingResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewExternalLoggingDetails().ExternalLoggingId(rs.Primary.ID).Do(context.Background())
		if err != nil {
			return err
		}
		//todo: check response _  fields if needed
		return nil
	}
}

func testFivetranExternalLoggingResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_external_logging" {
			continue
		}

		response, err := client.NewExternalLoggingDetails().ExternalLoggingId(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if !strings.HasPrefix(response.Code, "NotFound") {
			return errors.New("External Logging " + rs.Primary.ID + " still exists. Response code: " + response.Code)
		}

	}

	return nil
}
