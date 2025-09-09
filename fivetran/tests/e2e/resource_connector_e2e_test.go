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

					resource.TestCheckResourceAttrSet("fivetran_connector_schedule.test_connector_schedule", "group_id"),
					resource.TestCheckResourceAttrSet("fivetran_connector_schedule.test_connector_schedule", "connector_name"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "schedule_type", "auto"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "sync_frequency", "15"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "paused", "false"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "pause_after_trial", "false"),
				),
			},
		},
	})
}

func TestResourceConnectorScheduleByGroupIdAndConnectorNameE2E(t *testing.T) {
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

					group_id = fivetran_group.test_group.id
					connector_name = "fivetran_log_schema"
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

					resource.TestCheckResourceAttrSet("fivetran_connector_schedule.test_connector_schedule", "id"),
					resource.TestCheckResourceAttrSet("fivetran_connector_schedule.test_connector_schedule", "connector_id"),
					resource.TestCheckResourceAttrSet("fivetran_connector_schedule.test_connector_schedule", "group_id"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "connector_name", "fivetran_log_schema"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "schedule_type", "auto"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "sync_frequency", "5"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "paused", "true"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "pause_after_trial", "true"),
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
                 		env_type = "DOCKER"
					}

					resource "fivetran_hybrid_deployment_agent" "hybrid_deployment_agent2" {
						provider = fivetran-provider
    					display_name = "display_name_2"
    					group_id = fivetran_group.group.id
    					auth_type = "AUTO"
                 		env_type = "DOCKER"
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
                 		env_type = "DOCKER"
					}

					resource "fivetran_hybrid_deployment_agent" "hybrid_deployment_agent2" {
						provider = fivetran-provider
    					display_name = "display_name_2"
    					group_id = fivetran_group.group.id
    					auth_type = "AUTO"
                 		env_type = "DOCKER"
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

func TestResourceConnectorWithTableGroupNameE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectorResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "fivetran_group" "group" {
						provider = fivetran-provider
    					name = "test_group_name"
					}

					resource "fivetran_connector" "s3connector" {
					    provider = fivetran-provider
					    group_id  = fivetran_group.group.id
    					service  = "s3"
    					run_setup_tests  = false

    					destination_schema {
        					name = "my_s3_example_schema"
      						table_group_name = "table_group_name"
  					    } 

  						config {
      						bucket = "testbucket"
      						is_public = true
      						quote_character_enabled =  true
      						delimiter = ","
      						file_type = "csv"
      						on_error = "fail"
      						auth_type = "PUBLIC_BUCKET"
      						append_file_option = "upsert_file"
      						connection_type = "Directly"
      						compression = "uncompressed"

      						files {
          						table_name = "csvtable2"
          						file_pattern = "connection.csv"
      						}
    
      						files {
          						table_name = "csvtable1"
          						file_pattern = "myfile.csv"
        					}
    					}
  					}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.s3connector"),
					resource.TestCheckResourceAttrSet("fivetran_connector.s3connector", "destination_schema.table_group_name"),
					resource.TestCheckNoResourceAttr("fivetran_connector.s3connector", "destination_schema.table"),
					resource.TestCheckResourceAttrSet("fivetran_connector.s3connector", "destination_schema.name"),
					resource.TestCheckNoResourceAttr("fivetran_connector.s3connector", "destination_schema.prefix"),
				),
			},
		},
	})
}

func TestResourceConnectorNullableConfigFieldsE2E(t *testing.T) {
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

					resource "fivetran_connector" "test_connector" {
						provider = fivetran-provider
  						group_id = fivetran_group.group.id
  						service  = "maria"
  						run_setup_tests = false

  						config {
    						port             = "24020"
    						host             = "host"
    						update_method    = "BINLOG"
    						replica_id       = "12345"
    						tunnel_host      = "tunnel_host"
    						tunnel_user 	 = "tunnel_user"
    						tunnel_port		 = 2233
    						connection_type  = "SshTunnel"
  						}
  						
  						destination_schema {
    						prefix = "maria"
  						}
					}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_connector"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.connection_type", "SshTunnel"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "config.tunnel_host"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "config.tunnel_user"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "config.tunnel_port"),
				),
			},
			{
				Config: `
					resource "fivetran_group" "group" {
						provider = fivetran-provider
    					name = "sdhfkldwshkjshdkj"
					}

					resource "fivetran_connector" "test_connector" {
						provider = fivetran-provider
  						group_id = fivetran_group.group.id
  						service  = "maria"
  						run_setup_tests = false

  						config {
    						port             = "24020"
    						host             = "host"
    						update_method    = "BINLOG"
    						replica_id       = "12345"
    						tunnel_host      = null
    						tunnel_user 	 = null
    						tunnel_port		 = null
    						connection_type  = "Directly"
  						}
  						
  						destination_schema {
    						prefix = "maria"
  						}
					}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceUpdate(t, "fivetran_connector.test_connector"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.connection_type", "Directly"),
					resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "tunnel_host"),
					resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "tunnel_user"),
					resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "tunnel_port"),
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
