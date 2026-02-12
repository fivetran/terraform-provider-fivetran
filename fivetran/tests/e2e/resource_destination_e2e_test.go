package e2e_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceDestinationE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranDestinationResourceDestroy,
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
					region = "GCP_US_EAST4"
					daylight_saving_time_enabled = "true"
					trust_certificates = "true"
					trust_fingerprints = "true"
					run_setup_tests = "false"
					networking_method = "Directly"
			
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
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "region", "GCP_US_EAST4"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "trust_certificates", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "trust_fingerprints", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "daylight_saving_time_enabled", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "networking_method", "Directly"),
					resource.TestCheckNoResourceAttr("fivetran_destination.testdestination", "private_link_id"),
					resource.TestCheckNoResourceAttr("fivetran_destination.testdestination", "proxy_agent_id"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.host", "terraform-test.us-east-1.rds.amazonaws.com"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.port", "5432"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.user", "postgres"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.password", "password"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.database", "fivetran"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.connection_type", "Directly"),
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
					region = "GCP_EUROPE_WEST2"
					daylight_saving_time_enabled = "false"
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
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "region", "GCP_EUROPE_WEST2"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "trust_certificates", "false"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "trust_fingerprints", "false"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "daylight_saving_time_enabled", "false"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "networking_method", "Directly"),
					resource.TestCheckNoResourceAttr("fivetran_destination.testdestination", "private_link_id"),
					resource.TestCheckNoResourceAttr("fivetran_destination.testdestination", "proxy_agent_id"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.host", "terraform-test-updated.us-east-1.rds.amazonaws.com"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.port", "5434"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.user", "postgres_updated"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.password", "password_updated"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.database", "fivetran_updated"),
					resource.TestCheckResourceAttr("fivetran_destination.testdestination", "config.connection_type", "Directly"),
				),
			},
		},
	})
}

func TestResourceDestinationHdE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectorResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "fivetran_group" "group" {
						provider = fivetran-provider
    					name = "sdhfkldwshkjshdkj"
					}

					resource "fivetran_hybrid_deployment_agent" "hybrid_deployment_agent1" {
    					provider = fivetran-provider
    					display_name = "display_name_1"
    					group_id = fivetran_group.group.id
    					auth_type = "AUTO"
                 		env_type = "DOCKER"
					}

					resource "fivetran_hybrid_deployment_agent" "hybrid_deployment_agent2" {
						provider = fivetran-provider
    					display_name = "display_name_2"
    					group_id = fivetran_group.group.id
    					auth_type = "AUTO"
                 		env_type = "DOCKER"
					}

			    	resource "fivetran_destination" "testdestination" {
						provider = fivetran-provider
						group_id = fivetran_group.group.id
						service = "postgres_rds_warehouse"
						time_zone_offset = "0"
						region = "GCP_US_EAST4"
						daylight_saving_time_enabled = "true"
						trust_certificates = "true"
						trust_fingerprints = "true"
						run_setup_tests = "false"
      					hybrid_deployment_agent_id = fivetran_hybrid_deployment_agent.hybrid_deployment_agent1.id
			
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
					testFivetranDestinationResourceUpdate(t, "fivetran_destination.testdestination"),
					resource.TestCheckResourceAttrSet("fivetran_destination.testdestination", "hybrid_deployment_agent_id"),
				),
			},
			{
				Config: `
					resource "fivetran_group" "group" {
						provider = fivetran-provider
    					name = "sdhfkldwshkjshdkj"
					}

					resource "fivetran_hybrid_deployment_agent" "hybrid_deployment_agent1" {
						provider = fivetran-provider
    					display_name = "display_name_1"
    					group_id = fivetran_group.group.id
    					auth_type = "AUTO"
                 		env_type = "DOCKER"
					}

					resource "fivetran_hybrid_deployment_agent" "hybrid_deployment_agent2" {
						provider = fivetran-provider
    					display_name = "display_name_2"
    					group_id = fivetran_group.group.id
    					auth_type = "AUTO"
                 		env_type = "DOCKER"
					}

			    	resource "fivetran_destination" "testdestination" {
						provider = fivetran-provider
						group_id = fivetran_group.group.id
						service = "postgres_rds_warehouse"
						time_zone_offset = "0"
						region = "GCP_US_EAST4"
						daylight_saving_time_enabled = "true"
						trust_certificates = "true"
						trust_fingerprints = "true"
						run_setup_tests = "false"
      					hybrid_deployment_agent_id = fivetran_hybrid_deployment_agent.hybrid_deployment_agent2.id
			
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
					testFivetranDestinationResourceUpdate(t, "fivetran_destination.testdestination"),
					resource.TestCheckResourceAttrSet("fivetran_destination.testdestination", "hybrid_deployment_agent_id"),
				),
			},
		},
	})
}

func testFivetranDestinationResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewDestinationDetails().DestinationID(rs.Primary.ID).Do(context.Background())

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

		_, err := client.NewDestinationDetails().DestinationID(rs.Primary.ID).Do(context.Background())

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

		response, err := client.NewDestinationDetails().DestinationID(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if response.Code != "NotFound_Destination" {
			return errors.New("Destination " + rs.Primary.ID + " still exists. Response code: " + response.Code)
		}

	}

	return nil
}
