package e2e_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	predefinedPrivateLinkId = "<use the ID of an existing private link>"
	predefinedPrivateLinkName = "<use the name of an existing private link>"

)

var privateLinkResourceConfig = `
            	resource "fivetran_private_link" "test_pl" {
                	provider = fivetran-provider

                	name = "%v"
                	region = "AWS_US_EAST_1"
                	service = "REDSHIFT_AWS"

                	config_map = {
    					aws_account_id = "%v"
    					cluster_identifier = "%v"
                 	}
            	}`

func TestResourcePrivateLinkE2E(t *testing.T) {
	t.Skip("Private links have a strict limit on the number of entities created. This test should only be used for intermediate tests when changes are made directly to Private links.")
	suffix := strconv.Itoa(seededRand.Int())
	privateLinkName := suffix
	privateLinkCfgValue := "privatelink_" + suffix

	resourceConfig := fmt.Sprintf(privateLinkResourceConfig, privateLinkName, privateLinkCfgValue, privateLinkCfgValue)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: resourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranPrivateLinkResourceCreate(t, "fivetran_private_link.test_pl"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "name", privateLinkName),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "region", "AWS_US_EAST_1"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "service", "REDSHIFT_AWS"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config_map.aws_account_id", privateLinkCfgValue),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config_map.cluster_identifier", privateLinkCfgValue),
				),
			},
		},
	})
}

func TestResourcePrivateLinkUsageE2E(t *testing.T) {
	t.Skip("Private links have a strict limit on the number of entities created. This test should only be used for intermediate tests when changes are made directly to Private links.")
	suffix := strconv.Itoa(seededRand.Int())
	privateLinkName := "pl_" + suffix

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
            	resource "fivetran_private_link" "test_pl" {
                	provider = fivetran-provider

                	name = "%s"
                	region = "AZURE_EASTUS2"
                	service = "SOURCE_AZURE"

                	config_map = {
    					connection_service_id="/subscriptions/00000000-1111-2222-3333-444444444444/resourceGroups/fivetran-eastus/providers/Microsoft.Network/privateLinkServices/fivetran-eastus-1234"
    					private_dns_records="fivetran-eastus-1234"
						sub_resource_name="postgresqlServer"
                 	}
            	}			
			 	resource "fivetran_group" "test_group" {
			 		provider = fivetran-provider

			 		name = "test_group_for_private_link_test"
			 	}
					
			 	resource "fivetran_destination" "test_destination" {
			 		provider = fivetran-provider

			 		group_id 			= fivetran_group.test_group.id
  					service             = "databricks"
			 		networking_method 	= "PrivateLink"
			 		private_link_id 	= fivetran_private_link.test_pl.id
					time_zone_offset     = "0"
					region               = "AZURE_EASTUS"
					daylight_saving_time_enabled = true
					trust_certificates   = true
					trust_fingerprints   = true
					run_setup_tests      = false

					config {
						auth_type             = "PERSONAL_ACCESS_TOKEN"
						catalog               = "your_catalog"
						port                  = 443
						http_path             = "/sql/1.0/warehouses/your_warehouse"
						personal_access_token = "a_token"
					}
			    }
				`, privateLinkName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranPrivateLinkResourceCreate(t, "fivetran_private_link.test_pl"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "name", privateLinkName),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "region", "AZURE_EASTUS2"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "service", "SOURCE_AZURE"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "networking_method", "PrivateLink"),
					resource.TestCheckResourceAttrSet("fivetran_destination.test_destination", "private_link_id"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "time_zone_offset", "0"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "region", "AZURE_EASTUS"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "daylight_saving_time_enabled", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "trust_certificates", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "trust_fingerprints", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "config.auth_type", "PERSONAL_ACCESS_TOKEN"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "config.catalog", "your_catalog"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "config.port", "443"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "config.http_path", "/sql/1.0/warehouses/your_warehouse"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "config.personal_access_token", "a_token"),
				),
			},
			{
				Config: fmt.Sprintf(`
            	resource "fivetran_private_link" "test_pl" {
                	provider = fivetran-provider

                	name = "%s"
                	region = "AZURE_EASTUS2"
                	service = "SOURCE_AZURE"

                	config_map = {
    					connection_service_id="/subscriptions/00000000-1111-2222-3333-444444444444/resourceGroups/fivetran-eastus/providers/Microsoft.Network/privateLinkServices/fivetran-eastus-1234"
    					private_dns_records="fivetran-eastus-1234"
						sub_resource_name="postgresqlServer"
                 	}
            	}			
			 	resource "fivetran_group" "test_group" {
			 		provider = fivetran-provider

			 		name = "test_group_for_private_link_test"
			 	}
					
			 	resource "fivetran_destination" "test_destination" {
			 		provider = fivetran-provider

			 		group_id 			= fivetran_group.test_group.id
  					service             = "databricks"
			 		networking_method 	= "PrivateLink"
			 		private_link_id 	= fivetran_private_link.test_pl.id
					time_zone_offset     = "0"
					region               = "AZURE_EASTUS"
					daylight_saving_time_enabled = true
					trust_certificates   = true
					trust_fingerprints   = true
					run_setup_tests      = true

					config {
						auth_type             = "PERSONAL_ACCESS_TOKEN"
						catalog               = "your_catalog"
						port                  = 443
						http_path             = "/sql/1.0/warehouses/your_warehouse"
						personal_access_token = "a_token"
					}
			    }
				`, privateLinkName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "name", privateLinkName),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "region", "AZURE_EASTUS2"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "service", "SOURCE_AZURE"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "networking_method", "PrivateLink"),
					resource.TestCheckResourceAttrSet("fivetran_destination.test_destination", "private_link_id"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "time_zone_offset", "0"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "region", "AZURE_EASTUS"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "daylight_saving_time_enabled", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "trust_certificates", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "trust_fingerprints", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "run_setup_tests", "true"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "config.auth_type", "PERSONAL_ACCESS_TOKEN"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "config.catalog", "your_catalog"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "config.port", "443"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "config.http_path", "/sql/1.0/warehouses/your_warehouse"),
					resource.TestCheckResourceAttr("fivetran_destination.test_destination", "config.personal_access_token", "a_token"),
				),
			},
		},
	})
}

func testFivetranPrivateLinkResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewPrivateLinkDetails().PrivateLinkId(rs.Primary.ID).Do(context.Background())
		if err != nil {
			fmt.Println(err)
			return err
		}
		//todo: check response _  fields if needed
		return nil
	}
}

func testFivetranPrivateLinkExists(privateLinkId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		
		_, err := client.NewPrivateLinkDetails().PrivateLinkId(privateLinkId).Do(context.Background())
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	}
}

func TestResourcePrivateLinkImportingE2E(t *testing.T) {
	t.Skip("Private links have a strict limit on the number of entities created. This test should only be used for intermediate tests when changes are made directly to Private links.")
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "fivetran_private_link" "imported_pl" {
						provider = fivetran-provider

						name = "%v"
						region = "AWS_US_EAST_1"
						service = "REDSHIFT_AWS"
					}`, predefinedPrivateLinkName),
				ImportState:            true,
				ResourceName:            "fivetran_private_link.imported_pl",
				ImportStateId: predefinedPrivateLinkId,
			    
				ImportStateCheck: ComposeImportStateCheck(
					CheckImportResourceAttr("fivetran_private_link", "id", predefinedPrivateLinkId),
					CheckImportResourceAttr("fivetran_private_link", "name", predefinedPrivateLinkName),
					CheckImportResourceAttr("fivetran_private_link", "region", "AWS_US_EAST_1"),
					CheckImportResourceAttr("fivetran_private_link", "service", "REDSHIFT_AWS"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "fivetran_private_link" "imported_pl" {
						provider = fivetran-provider

						name = "%v"
						region = "AWS_US_EAST_1"
						service = "REDSHIFT_AWS"
						config_map = {
						}
					}`, predefinedPrivateLinkName),
				ImportState:            true,
				ImportStatePersist: 	true,
				ResourceName:            "fivetran_private_link.imported_pl",
				ImportStateId: predefinedPrivateLinkId,
			    
				ImportStateCheck: ComposeImportStateCheck(
					CheckImportResourceAttr("fivetran_private_link", "id", predefinedPrivateLinkId),
					CheckImportResourceAttr("fivetran_private_link", "name", predefinedPrivateLinkName),
					CheckImportResourceAttr("fivetran_private_link", "region", "AWS_US_EAST_1"),
					CheckImportResourceAttr("fivetran_private_link", "service", "REDSHIFT_AWS"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "fivetran_private_link" "imported_pl" {
						provider = fivetran-provider

						name = "%v"
						region = "AWS_US_EAST_1"
						service = "REDSHIFT_AWS"
						config_map = {
						}
					}`, predefinedPrivateLinkName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranPrivateLinkExists(predefinedPrivateLinkId),
					resource.TestCheckResourceAttr("fivetran_private_link.imported_pl", "id", predefinedPrivateLinkId),
					resource.TestCheckResourceAttr("fivetran_private_link.imported_pl", "name", predefinedPrivateLinkName),
					resource.TestCheckResourceAttr("fivetran_private_link.imported_pl", "region", "AWS_US_EAST_1"),
					resource.TestCheckResourceAttr("fivetran_private_link.imported_pl", "service", "REDSHIFT_AWS"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "fivetran_private_link" "imported_pl" {
						provider = fivetran-provider

						name = "%v"
						region = "AWS_US_EAST_1"
						service = "REDSHIFT_AWS"
						config_map = {}
					}`, predefinedPrivateLinkName),
				PlanOnly: true,
			},
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranPrivateLinkExists(predefinedPrivateLinkId),
					resource.TestCheckResourceAttr("fivetran_private_link.imported_pl", "id", predefinedPrivateLinkId),
					resource.TestCheckResourceAttr("fivetran_private_link.imported_pl", "name", predefinedPrivateLinkName),
					resource.TestCheckResourceAttr("fivetran_private_link.imported_pl", "region", "AWS_US_EAST_1"),
					resource.TestCheckResourceAttr("fivetran_private_link.imported_pl", "service", "REDSHIFT_AWS"),
				),
			},
			{
				Config: `
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranPrivateLinkExists(predefinedPrivateLinkId),
				),
			},
		},
	})
}