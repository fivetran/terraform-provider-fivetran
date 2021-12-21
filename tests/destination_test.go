package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)


func TestFivetranDestination_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {},
		Providers:    testProviders,
		CheckDestroy: testFivetranDestinationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "testgroup" {
					provider = fivetran-provider
					name = "test_group_name"
			    }

			    resource "fivetran_destination" "testdestination" {
					provider = fivetran-provider
					group_id = fivetran_group.testgroup.id
					service = "postgres_rds_warehouse"
					time_zone_offset = "0"
					region = "US"
					trust_certificates = "true"
					trust_fingerprints = "true"
					run_setup_tests = "false"
			
					config {
						host = "terraform-test.us-east-1.rds.amazonaws.com"
						port = 5432
						user = "postgres"
						password = "password"
						database = "fivetran"
						connection_type = "Directly"
					}
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranDestinationResourceCreate(t, "fivetran_destination.testdestination"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "service", "postgres_rds_warehouse"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "time_zone_offset", "0"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "region", "US"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "trust_certificates", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "trust_fingerprints", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.host", "terraform-test.us-east-1.rds.amazonaws.com"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.port", "5432"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.user", "postgres"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.password", "password"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.database", "fivetran"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.connection_type", "Directly"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "testgroup" {
					provider = fivetran-provider
					name = "test_group_name"
			    }

				resource "fivetran_destination" "testdestination" {
					provider = fivetran-provider
					group_id = fivetran_group.testgroup.id
					service = "postgres_rds_warehouse"
					time_zone_offset = "+4"
					region = "UK"
					trust_certificates = "false"
					trust_fingerprints = "false"
					run_setup_tests = "false"
			
					config {
						host = "terraform-test-updated.us-east-1.rds.amazonaws.com"
						port = 5434
						user = "postgres_updated"
						password = "password_updated"
						database = "fivetran_updated"
						connection_type = "Directly"
					}
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranDestinationResourceUpdate(t, "fivetran_destination.testdestination"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "service", "postgres_rds_warehouse"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "time_zone_offset", "+4"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "region", "UK"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "trust_certificates", "false"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "trust_fingerprints", "false"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.host", "terraform-test-updated.us-east-1.rds.amazonaws.com"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.port", "5434"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.user", "postgres_updated"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.password", "password_updated"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.database", "fivetran_updated"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.0.connection_type", "Directly"),
				),
			},
		},
	})
}

func testFivetranDestinationResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
	
		_, err := Client().NewDestinationDetails().DestinationID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields if needed
		return nil
	}
}

func testFivetranDestinationResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
	
		_, err := Client().NewDestinationDetails().DestinationID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields if needed
		return nil
	}
}

func testFivetranDestinationResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_destination" {
			continue
		}

		response, err := Client().NewDestinationDetails().DestinationID(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if response.Code != "NotFound" {
			return errors.New("Destination " + rs.Primary.ID + " still exists.")
		}

	}

	return nil
}