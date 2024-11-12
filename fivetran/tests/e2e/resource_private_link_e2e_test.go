package e2e_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	//t.Skip("Private links have a strict limit on the number of entities created. This test should only be used for intermediate tests when changes are made directly to Private links.")
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