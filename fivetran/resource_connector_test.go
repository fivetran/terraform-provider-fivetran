package fivetran_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceConnectorE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() {},
		ProviderFactories: providerFactory,
		CheckDestroy: testFivetranConnectorResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_name"
			    }

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					lifecycle {` +
					//Ignoring `auth_type` cause it returned by default but is abcent in config
					`		
						ignore_changes = ["config[0].auth_type"]
					}
					group_id = fivetran_group.test_group.id
					service = "google_sheets"
					destination_schema {
						name = "google_sheets_schema"
						table = "table"
					}
					sync_frequency = 5
					paused = true
					pause_after_trial = true
					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false
			
					config {
						sheet_id = "1Rmq_FN2kTNwWiT4adZKBxHBRmvfeBTIfKWi5B8ii9qk"
						named_range = "range"
					}
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_connector"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "google_sheets"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service_version", "1"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "schedule_type", "auto"),

					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.is_historical_sync", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.update_state", "on_schedule"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.setup_state", "incomplete"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.sync_state", "paused"),

					//schema_table format mutate schema to `schema` +`.` + `config.table` 
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "google_sheets_schema.table"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "sync_frequency", "5"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "paused", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "pause_after_trial", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.auth_type", "ServiceAccount"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.sheet_id", "1Rmq_FN2kTNwWiT4adZKBxHBRmvfeBTIfKWi5B8ii9qk"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.named_range", "range"),
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
					lifecycle {` +
					//Ignoring `auth_type` cause it returned by default but is abcent in config
					`
						ignore_changes = ["config[0].auth_type"]
					}
					group_id = fivetran_group.test_group.id
					service = "google_sheets"

					destination_schema {
						name = "google_sheets_schema"
						table = "table"
					}
					
					sync_frequency = 15
					paused = false
					pause_after_trial = false
					trust_certificates = true
					trust_fingerprints = true
					run_setup_tests = false
			
					config {
						sheet_id = "1Rmq_RmvfeBTIfKWi5B8ii9qkFN2kTNwWiT4adZKBxHB"
						named_range = "range_updated"
					}
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceUpdate(t, "fivetran_connector.test_connector"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "google_sheets"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service_version", "1"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "schedule_type", "auto"),

					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.is_historical_sync", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.update_state", "on_schedule"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.setup_state", "incomplete"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "status.0.sync_state", "scheduled"),

					//schema_table format mutate schema to `schema` +`.` + `config.table` 
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "google_sheets_schema.table"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "sync_frequency", "15"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "paused", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "pause_after_trial", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "true"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.auth_type", "ServiceAccount"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.sheet_id", "1Rmq_RmvfeBTIfKWi5B8ii9qkFN2kTNwWiT4adZKBxHB"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "config.0.named_range", "range_updated"),
				),
			},
		},
	})
}

func testFivetranConnectorResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewConnectorDetails().ConnectorID(rs.Primary.ID).Do(context.Background())

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

		_, err := client.NewConnectorDetails().ConnectorID(rs.Primary.ID).Do(context.Background())

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

		response, err := client.NewConnectorDetails().ConnectorID(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if response.Code != "NotFound_Connector" {
			return errors.New("Connector " + rs.Primary.ID + " still exists.")
		}
	}

	return nil
}