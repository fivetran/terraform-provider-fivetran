package e2e_test

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
					name = "TestResourceExternalLoggingE2E"
			    }

				resource "fivetran_external_logging" "test_extlog" {
					provider = fivetran-provider

    				group_id = fivetran_group.testgroup.id
    				service = "splunkLog"
    				enabled = "true"
    				run_setup_tests = "false"

				    config {
        				host = "1.1.1.1"
						port = 8080
        				token = "PASSWORD"
    				}
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranExternalLoggingResourceCreate(t, "fivetran_external_logging.test_extlog"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "service", "splunkLog"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "enabled", "true"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.host", "1.1.1.1"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.port", "8080"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.token", "PASSWORD"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "testgroup" {
					provider = fivetran-provider
					name = "TestResourceExternalLoggingE2E"
			    }

				resource "fivetran_external_logging" "test_extlog" {
					provider = fivetran-provider

    				group_id = fivetran_group.testgroup.id
    				service = "splunkLog"
    				enabled = "true"
    				run_setup_tests = "false"

				    config {
        				host = "1.1.1.1"
						port = 8080
        				token = "PASSWORD"
    				}
				}

				data "fivetran_external_logging" "data_test_extlog" {
					provider = fivetran-provider

					id = fivetran_external_logging.test_extlog.id
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranExternalLoggingResourceUpdate(t, "fivetran_external_logging.test_extlog"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "service", "splunkLog"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "enabled", "true"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.host", "1.1.1.1"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.port", "8080"),
					resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.token", "PASSWORD"),

					resource.TestCheckResourceAttr("data.fivetran_external_logging.data_test_extlog", "service", "splunkLog"),
					resource.TestCheckResourceAttr("data.fivetran_external_logging.data_test_extlog", "enabled", "true"),
					resource.TestCheckResourceAttr("data.fivetran_external_logging.data_test_extlog", "config.host", "1.1.1.1"),
					resource.TestCheckResourceAttr("data.fivetran_external_logging.data_test_extlog", "config.port", "8080"),
					resource.TestCheckNoResourceAttr("data.fivetran_external_logging.data_test_extlog", "config.token"),
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
