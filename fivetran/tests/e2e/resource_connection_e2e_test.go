package e2e_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceConnectionE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_connection"
			    }

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_connection"
					}

					config = jsonencode({
						update_method = "XMIN"
					})

					run_setup_tests = false
					trust_certificates = false
					trust_fingerprints = false
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionResourceCreate(t, "fivetran_connection.test_connection"),
					resource.TestCheckResourceAttrSet("fivetran_connection.test_connection", "id"),
					resource.TestCheckResourceAttrSet("fivetran_connection.test_connection", "name"),
					resource.TestCheckResourceAttrSet("fivetran_connection.test_connection", "created_at"),
					resource.TestCheckResourceAttrSet("fivetran_connection.test_connection", "connected_by"),
					resource.TestCheckResourceAttr("fivetran_connection.test_connection", "service", "postgres"),
					resource.TestCheckResourceAttr("fivetran_connection.test_connection", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_connection.test_connection", "trust_certificates", "false"),
					resource.TestCheckResourceAttr("fivetran_connection.test_connection", "trust_fingerprints", "false"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_connection"
			    }

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_connection"
					}

					config = jsonencode({
						update_method = "XMIN"
					})

					run_setup_tests = true
					trust_certificates = true
					trust_fingerprints = true
					data_delay_sensitivity = "NORMAL"
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionResourceUpdate(t, "fivetran_connection.test_connection"),
					resource.TestCheckResourceAttrSet("fivetran_connection.test_connection", "id"),
					resource.TestCheckResourceAttr("fivetran_connection.test_connection", "service", "postgres"),
					resource.TestCheckResourceAttr("fivetran_connection.test_connection", "run_setup_tests", "true"),
					resource.TestCheckResourceAttr("fivetran_connection.test_connection", "trust_certificates", "true"),
					resource.TestCheckResourceAttr("fivetran_connection.test_connection", "trust_fingerprints", "true"),
					resource.TestCheckResourceAttr("fivetran_connection.test_connection", "data_delay_sensitivity", "NORMAL"),
				),
			},
			{
				ResourceName:      "fivetran_connection.test_connection",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"run_setup_tests", "trust_certificates", "trust_fingerprints", "config", "data_delay_sensitivity"},
			},
		},
	})
}

func TestResourceConnectionMultipleServicesE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_multi"
			    }

			    resource "fivetran_connection" "postgres_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_multi"
					}

					config = jsonencode({
						update_method = "XMIN"
					})

					run_setup_tests = false
				}

			    resource "fivetran_connection" "mysql_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "mysql"

					destination_schema {
						prefix = "mysql_multi"
					}

					config = jsonencode({
						update_method = "BINLOG"
					})

					run_setup_tests = false
				}

			    resource "fivetran_connection" "s3_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "s3"

					destination_schema {
						name = "s3_schema"
					}

					config = jsonencode({
						role_arn = "arn:aws:iam::123456789:role/fivetran"
						table_group_name = "s3_table_group"
					})

					run_setup_tests = false
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionResourceCreate(t, "fivetran_connection.postgres_connection"),
					resource.TestCheckResourceAttr("fivetran_connection.postgres_connection", "service", "postgres"),

					testFivetranConnectionResourceCreate(t, "fivetran_connection.mysql_connection"),
					resource.TestCheckResourceAttr("fivetran_connection.mysql_connection", "service", "mysql"),

					testFivetranConnectionResourceCreate(t, "fivetran_connection.s3_connection"),
					resource.TestCheckResourceAttr("fivetran_connection.s3_connection", "service", "s3"),
				),
			},
		},
	})
}

func testFivetranConnectionResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewConnectionDetails().ConnectionID(rs.Primary.ID).Do(context.Background())
		if err != nil {
			return fmt.Errorf("connection %s not found: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testFivetranConnectionResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewConnectionDetails().ConnectionID(rs.Primary.ID).Do(context.Background())
		if err != nil {
			return fmt.Errorf("connection %s not found: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testFivetranConnectionResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_connection" {
			continue
		}

		response, err := client.NewConnectionDetails().ConnectionID(rs.Primary.ID).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			return err
		}

		if err == nil && !strings.HasPrefix(response.Code, "NotFound_") {
			return fmt.Errorf(`
			Connection %s still exists after deletion.
			Expected response.Code: 'NotFound_Connection'.
			Actual response.Code was: '%s'.
			response.Message: '%s'`, rs.Primary.ID, response.Code, response.Message)
		}
	}

	return nil
}
