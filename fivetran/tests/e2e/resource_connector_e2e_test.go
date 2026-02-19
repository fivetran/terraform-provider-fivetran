package e2e_test

import (
	"context"
	"fmt"
	"strconv"
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

            	resource "fivetran_proxy_agent" "test_proxy_agent" {
                	provider = fivetran-provider

                 	display_name = "test_proxy_agent1"
                 	group_region = "GCP_US_EAST4"
            	}

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"
					
					data_delay_sensitivity = "NORMAL"
					data_delay_threshold = 0

					destination_schema {
						prefix = "fivetran_log_schema"
					}
					
					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false

					networking_method  = "ProxyAgent"
					proxy_agent_id = fivetran_proxy_agent.test_proxy_agent.id

      				config {
        				user = "user1"
        				password = "password1"
        				host = "host"
        				port = "123"
        				update_method = "QUERY_BASED"
						connection_type  = "ProxyAgent"
      				}
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
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "postgres"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "fivetran_log_schema"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_sensitivity", "NORMAL"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_threshold", "0"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "networking_method", "ProxyAgent"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "proxy_agent_id"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.connection_type", "ProxyAgent"),

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

            	resource "fivetran_proxy_agent" "test_proxy_agent" {
                	provider = fivetran-provider

                 	display_name = "test_proxy_agent1"
                 	group_region = "GCP_US_EAST4"
            	}

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"
					
					data_delay_sensitivity = "NORMAL"
					data_delay_threshold = 0

					destination_schema {
						prefix = "fivetran_log_schema"
					}

					trust_certificates = true
					trust_fingerprints = true
					run_setup_tests = true

					networking_method  = "SshTunnel"
					proxy_agent_id = null

      				config {
        				user = "user1"
        				password = "password1"
        				host = "host"
        				port = "123"
        				update_method = "QUERY_BASED"
						tunnel_host      = "127.0.0.1"
						tunnel_port      = 22
						tunnel_user      = "fivetran"
						connection_type  = null
      				}
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

					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "postgres"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "fivetran_log_schema"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_sensitivity", "NORMAL"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_threshold", "0"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "networking_method", "SshTunnel"),
					resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "proxy_agent_id"),
					resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "config.connection_type"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.tunnel_host", "127.0.0.1"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.tunnel_port", "22"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.tunnel_user", "fivetran"),

					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "schedule_type", "auto"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "sync_frequency", "15"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "paused", "false"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "pause_after_trial", "false"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_name"
			    }

            	resource "fivetran_proxy_agent" "test_proxy_agent" {
                	provider = fivetran-provider

                 	display_name = "test_proxy_agent1"
                 	group_region = "GCP_US_EAST4"
            	}

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"
					
					data_delay_sensitivity = "NORMAL"
					data_delay_threshold = 0

					destination_schema {
						prefix = "fivetran_log_schema"
					}
					
					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false

					networking_method  = "ProxyAgent"
					proxy_agent_id = fivetran_proxy_agent.test_proxy_agent.id

      				config {
        				user = "user1"
        				password = "password1"
        				host = "host"
        				port = "123"
        				update_method = "QUERY_BASED"
						connection_type  = null
						tunnel_host      = null
						tunnel_port      = null
						tunnel_user      = null
      				}
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
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "postgres"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "fivetran_log_schema"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_sensitivity", "NORMAL"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_threshold", "0"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "networking_method", "ProxyAgent"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "proxy_agent_id"),
					resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "config.connection_type"),
					resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "config.tunnel_host"),
					resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "config.tunnel_port"),
					resource.TestCheckNoResourceAttr("fivetran_connector.test_connector", "config.tunnel_user"),

					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "schedule_type", "auto"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "sync_frequency", "5"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "paused", "true"),
					resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "pause_after_trial", "true"),
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

					depends_on = [
						fivetran_connector.test_connector
					]

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
        					update_method = "QUERY_BASED"
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
        					update_method = "QUERY_BASED"
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

func TestGoogleVideo360ReportsE2E(t *testing.T) {
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

				locals {

					reports = {
						"standard_report_daily" = {
							config_method = "CREATE_NEW"
							report_type   = "STANDARD"
							partners      = []
							dimensions = [
								"FILTER_ADVERTISER",
								"FILTER_DATE"
							]
							metrics = [
								"METRIC_CLICKS"
							]
							update_config_on_each_sync = true
						}
					}
				}

			    resource "fivetran_connector" "google_video_360_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "google_display_and_video_360"
					run_setup_tests = false
					destination_schema {
						name = "google_display_and_video_360" 
					}
					config {
						timeframe_months = "TWELVE"

						dynamic "reports" {
							for_each = local.reports

							content {
								table_name                 = reports.key
								
								config_method              = reports.value.config_method
								partners                   = reports.value.partners
								report_type                = reports.value.report_type
								dimensions                 = reports.value.dimensions
								metrics                    = reports.value.metrics
								update_config_on_each_sync = reports.value.update_config_on_each_sync
							}
						}
					}
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.google_video_360_connection"),
					resource.TestCheckResourceAttr("fivetran_connector.google_video_360_connection", "service", "google_display_and_video_360"),
					resource.TestCheckResourceAttr("fivetran_connector.google_video_360_connection", "config.reports.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connector.google_video_360_connection", "config.reports.0.table_name", "standard_report_daily"),
					resource.TestCheckResourceAttr("fivetran_connector.google_video_360_connection", "config.reports.0.dimensions.#", "2"),
					resource.TestCheckResourceAttr("fivetran_connector.google_video_360_connection", "config.reports.0.dimensions.0", "FILTER_ADVERTISER"),
					resource.TestCheckResourceAttr("fivetran_connector.google_video_360_connection", "config.reports.0.dimensions.1", "FILTER_DATE"),
					resource.TestCheckResourceAttr("fivetran_connector.google_video_360_connection", "config.reports.0.metrics.0", "METRIC_CLICKS"),
					resource.TestCheckResourceAttr("fivetran_connector.google_video_360_connection", "config.reports.0.per_interaction_dimensions.#", "0"),
					resource.TestCheckResourceAttr("fivetran_connector.google_video_360_connection", "config.reports.0.report_configuration_ids.#", "0"),
					resource.TestCheckResourceAttr("fivetran_connector.google_video_360_connection", "config.reports.0.report_type", "STANDARD"),
				),
			},
		},
	})
}

func TestWorkdayAdaptiveReportsE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_workday_adaptive"
			    }

			    resource "fivetran_connector" "workday_adaptive_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "workday_adaptive"
					run_setup_tests = false
					destination_schema {
						name = "workday_adaptive" 
					}
					config {
						login = "mail@mail.com"
						password = "password"
						instance_code = "INSTANCE_CODE1"
						reports {
							table = "table_1"
							version_sync_strategy = "SYNC_SELECT_VERSIONS"
							include_zero_rows = true
							versions = ["2945"]
							accounts  { 
							  id = "3"
							  flag = true 
							}
							accounts { 
							  id = "4"
							  flag = true 
							}
							dimensions = ["3","4"]
						}
					}
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.workday_adaptive_connection"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "service", "workday_adaptive"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.login", "mail@mail.com"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.password", "password"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.instance_code", "INSTANCE_CODE1"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.reports.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.reports.0.table", "table_1"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.reports.0.version_sync_strategy", "SYNC_SELECT_VERSIONS"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.reports.0.include_zero_rows", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.reports.0.versions.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.reports.0.versions.0", "2945"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.reports.0.accounts.#", "2"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.reports.0.dimensions.#", "2"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.reports.0.dimensions.0", "3"),
					resource.TestCheckResourceAttr("fivetran_connector.workday_adaptive_connection", "config.reports.0.dimensions.1", "4"),
				),
			},
		},
	})
}

func TestGoogleCampaignManagerReportsE2E(t *testing.T) {
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

				locals {

					reports = {
						"standard_report_daily" = {
							config_method = "CREATE_NEW"
							report_type   = "STANDARD"
							dimensions = [
								"FILTER_ADVERTISER",
								"FILTER_DATE"
							]
							metrics = [
								"METRIC_CLICKS"
							]
						}
					}
				}

			    resource "fivetran_connector" "google_campaign_manager_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "double_click_campaign_manager"
					run_setup_tests = false
					destination_schema {
						name = "double_click_campaign_manager" 
					}
					config {
						timeframe_months = "TWELVE"

						dynamic "reports" {
							for_each = local.reports

							content {
								table                 = reports.key
								conversion_dimensions = reports.value.dimensions
								custom_floodlight_variables = []
								dimensions                 = reports.value.dimensions
								enable_all_dimension_combinations = false
								metrics                    = reports.value.metrics
								per_interaction_dimensions = []
								report_configuration_ids  = []
								report_type                = reports.value.report_type
							}
						}
					}
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.google_campaign_manager_connection"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "service", "double_click_campaign_manager"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.timeframe_months", "TWELVE"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.reports.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.reports.0.table", "standard_report_daily"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.reports.0.conversion_dimensions.0", "FILTER_ADVERTISER"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.reports.0.dimensions.#", "2"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.reports.0.dimensions.0", "FILTER_ADVERTISER"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.reports.0.dimensions.1", "FILTER_DATE"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.reports.0.enable_all_dimension_combinations", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.reports.0.metrics.0", "METRIC_CLICKS"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.reports.0.per_interaction_dimensions.#", "0"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.reports.0.report_configuration_ids.#", "0"),
					resource.TestCheckResourceAttr("fivetran_connector.google_campaign_manager_connection", "config.reports.0.report_type", "STANDARD"),
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

func TestResourceConnectorPlanOnlyAttributesE2E(t *testing.T) {
	suffix := strconv.Itoa(seededRand.Int())
	groupName := "test_group_plan_" + suffix
	schemaPrefix := "pg_plan_only_" + suffix

	config1 := fmt.Sprintf(`
		resource "fivetran_group" "test_group" {
			provider = fivetran-provider
			name = "%s"
		}

		resource "fivetran_destination" "test_destination" {
			provider = fivetran-provider
			group_id = fivetran_group.test_group.id
			service = "big_query"
			time_zone_offset = "-5"
			region = "GCP_US_EAST4"
			run_setup_tests = false
			config {
				project_id = "%s"
				data_set_location = "US"
			}
		}

		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider
			group_id = fivetran_group.test_group.id
			service = "postgres"

			destination_schema {
				prefix = "%s"
			}

			config {
				user = "test_user"
				password = "test_password"
				host = "test.example.com"
				port = "5432"
				update_method = "QUERY_BASED"
			}

			run_setup_tests = false
			trust_certificates = false
			trust_fingerprints = false

			depends_on = [fivetran_destination.test_destination]
		}
	`, groupName, BqProjectId, schemaPrefix)

	config2 := fmt.Sprintf(`
		resource "fivetran_group" "test_group" {
			provider = fivetran-provider
			name = "%s"
		}

		resource "fivetran_destination" "test_destination" {
			provider = fivetran-provider
			group_id = fivetran_group.test_group.id
			service = "big_query"
			time_zone_offset = "-5"
			region = "GCP_US_EAST4"
			run_setup_tests = false
			config {
				project_id = "%s"
				data_set_location = "US"
			}
		}

		resource "fivetran_connector" "test_connector" {
			provider = fivetran-provider
			group_id = fivetran_group.test_group.id
			service = "postgres"

			destination_schema {
				prefix = "%s"
			}

			config {
				user = "test_user"
				password = "test_password"
				host = "test.example.com"
				port = "5432"
				update_method = "QUERY_BASED"
			}

			run_setup_tests = true
			trust_certificates = false
			trust_fingerprints = false

			depends_on = [fivetran_destination.test_destination]
		}
	`, groupName, BqProjectId, schemaPrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectorResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_connector"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "postgres"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "id"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceUpdate(t, "fivetran_connector.test_connector"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "postgres"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "id"),
				),
			},
		},
	})
}

func TestResourceConnectorImportingWithTableGroupNameE2E(t *testing.T) {
	var connectionId string = ""

	tfConfig := `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_connection"
			    }

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "share_point"

					destination_schema {
						name = "share_point_connection"
						table_group_name = "table_group_name1"
					}

					auth {
						tenant_id = "tenant_id1"

						client_access {
							client_id     = "client_id1"
							client_secret = "client_secret1"
						}
					}

					config {
						access_type          = "CUSTOM_APP"
						is_single_table_mode = true
    					prefix               = "prefix1"
						line_separator       = "\\n"
    					compression          = "uncompressed"
						append_file_option   = "upsert_file"
						file_handling        = "unstructured"
						file_type            = "unstructured"
						json_delivery_mode   = "Packed"
						on_error             = "fail"
						quote_char           = "\""
						quote_character_enabled = true
						empty_header         = false
						sync_permissions     = false

						files {
							table_name = "table_name1"
						}
					}

					run_setup_tests = false
					trust_certificates = false
					trust_fingerprints = false
				}

				data "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					id = fivetran_connector.test_connector.id
				}
		  `

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: tfConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionResourceCreate(t, "fivetran_connector.test_connector"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "id"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "name"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "created_at"),
					resource.TestCheckResourceAttrSet("fivetran_connector.test_connector", "connected_by"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "share_point"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "destination_schema.name", "share_point_connection"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "destination_schema.table_group_name", "table_group_name1"),

					resource.TestCheckResourceAttrSet("data.fivetran_connector.test_connector", "id"),
					resource.TestCheckResourceAttrSet("data.fivetran_connector.test_connector", "name"),
					resource.TestCheckResourceAttrSet("data.fivetran_connector.test_connector", "created_at"),
					resource.TestCheckResourceAttrSet("data.fivetran_connector.test_connector", "connected_by"),
					resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "service", "share_point"),
					resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "destination_schema.name", "share_point_connection"),
					resource.TestCheckResourceAttr("data.fivetran_connector.test_connector", "destination_schema.table_group_name", "table_group_name1"),
					func(s *terraform.State) error {
						rs := GetResource(t, s, "fivetran_connector.test_connector")
						connectionId = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: tfConfig,
				ResourceName:      "fivetran_connector.test_connector",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"run_setup_tests", 
					"trust_certificates", 
					"trust_fingerprints",
					"data_delay_sensitivity",
					"auth",
				},
				ImportStateIdFunc:       func(s *terraform.State) (string, error) {
					return connectionId, nil
				},

				ImportStateCheck: ComposeImportStateCheck(
					CheckImportResourceAttr("fivetran_connector", "service", "share_point"),
					CheckImportResourceAttr("fivetran_connector", "destination_schema.name", "share_point_connection"),
					CheckImportResourceAttr("fivetran_connector", "destination_schema.table_group_name", "table_group_name1"),
					CheckNoImportResourceAttr("fivetran_connector", "destination_schema.table"),
					CheckNoImportResourceAttr("fivetran_connector", "destination_schema.prefix"),
				),
			},
		},
	})
}

func TestResourceConnectorWhenConfigHasEmptyNestedObjectsE2E(t *testing.T) {
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
					service = "pendo"
					
					data_delay_sensitivity = "NORMAL"
					data_delay_threshold = 0
					networking_method = "Directly"

					destination_schema {
						name = "schema1"
					}

					config {
					 	sync_mode = "SpecificAppIds"
						integration_key = "integration_key1"
						app_ids = ["-1", "1", "2"]
						region = "US"

						# aws_credentials, etc., are absent intentionally, to test connection creation without empty nested objects
					}
					
					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_connector"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "pendo"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "schema1"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_sensitivity", "NORMAL"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "data_delay_threshold", "0"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "networking_method", "Directly"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.sync_mode", "SpecificAppIds"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.integration_key", "integration_key1"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.region", "US"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.app_ids.#", "3"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.app_ids.0", "-1"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.app_ids.1", "1"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.app_ids.2", "2"),
				),
			},
		},
	})
}