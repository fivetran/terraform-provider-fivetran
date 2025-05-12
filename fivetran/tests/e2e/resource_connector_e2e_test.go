package e2e_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceConnectorE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectorResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_name"
			    }

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "fivetran_log"
					
					data_delay_sensitivity = "NORMAL"
					data_delay_threshold = 0

					destination_schema {
						name = "fivetran_log_schema"
					}
					
					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false
				}

				resource "fivetran_connector_schedule" "test_connector_schedule" {
					provider = fivetran-provider

					connector_id = fivetran_connector.test_connector.id
					sync_frequency = 5
					paused = true
					pause_after_trial = true
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_connector"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "fivetran_log"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "fivetran_log_schema"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_sensitivity", "NORMAL"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_threshold", "0"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "schedule_type", "auto"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "sync_frequency", "5"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "paused", "true"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "pause_after_trial", "true"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_name"
			    }

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "fivetran_log"
					
					data_delay_sensitivity = "NORMAL"
					data_delay_threshold = 0

					destination_schema {
						name = "fivetran_log_schema"
					}

					trust_certificates = true
					trust_fingerprints = true
					run_setup_tests = true
				}

				resource "fivetran_connector_schedule" "test_connector_schedule" {
					provider = fivetran-provider
					
					connector_id = fivetran_connector.test_connector.id
					sync_frequency = 15
					paused = false
					pause_after_trial = false
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceUpdate(t, "fivetran_connector.test_connector"),

					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "id"),

					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "fivetran_log"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "fivetran_log_schema"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_sensitivity", "NORMAL"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_threshold", "0"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "true"),

					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "schedule_type", "auto"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "sync_frequency", "15"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "paused", "false"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "pause_after_trial", "false"),
				),
			},
		},
	})
}

func TestResourceConnectorHdE2E(t *testing.T) {
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
					}

					resource "fivetran_hybrid_deployment_agent" "hybrid_deployment_agent2" {
						provider = fivetran-provider
    					display_name = "display_name_2"
    					group_id = fivetran_group.group.id
    					auth_type = "AUTO"
					}

					resource "fivetran_connector" "test_connector" {
						provider = fivetran-provider
      					group_id = fivetran_group.group.id
      					service = "postgres"

      					hybrid_deployment_agent_id = fivetran_hybrid_deployment_agent.hybrid_deployment_agent1.id

      					destination_schema {
        					prefix = "postgres"
      					}

      					trust_certificates = true
      					trust_fingerprints = true
      					run_setup_tests = false

      					config {
        					user = "user1"
        					password = "password1"
        					host = "host"
        					port = "123"
      					}
    				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_connector"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "hybrid_deployment_agent_id"),
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
					}

					resource "fivetran_hybrid_deployment_agent" "hybrid_deployment_agent2" {
						provider = fivetran-provider
    					display_name = "display_name_2"
    					group_id = fivetran_group.group.id
    					auth_type = "AUTO"
					}

					resource "fivetran_connector" "test_connector" {
						provider = fivetran-provider
      					group_id = fivetran_group.group.id
      					service = "postgres"

      					hybrid_deployment_agent_id = fivetran_hybrid_deployment_agent.hybrid_deployment_agent2.id

      					destination_schema {
        					prefix = "postgres"
      					}

      					trust_certificates = true
      					trust_fingerprints = true
      					run_setup_tests = false

      					config {
        					user = "user1"
        					password = "password1"
        					host = "host"
        					port = "123"
      					}
    				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceUpdate(t, "fivetran_connector.test_connector"),

					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "hybrid_deployment_agent_id"),
				),
			},
		},
	})
}

func testFivetranConnectorResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewConnectionDetails().ConnectionID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields if needed
		return nil
	}
}

func testFivetranConnectorResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewConnectionDetails().ConnectionID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields if needed
		return nil
	}
}

func testFivetranConnectorResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_connector" {
			continue
		}

		response, err := client.NewConnectionDetails().ConnectionID(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}

		if !strings.HasPrefix(response.Code, "NotFound_") {
			return fmt.Errorf(`
			There was no error occured on recieving connector after deletion!

			Expected response.Code: 'NotFound_Connector'. 
			Actual response.Code was: '%s'. 
			response.Message: '%s'
			Connector %s still exists.`, response.Code, response.Message, rs.Primary.ID)
		}

	}

	return nil
}
