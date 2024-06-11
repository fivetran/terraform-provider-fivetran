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

func TestResourcePrivateLinkE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranPrivateLinkResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
            	resource "fivetran_private_link" "test_pl" {
                	provider = fivetran-provider

                	name = "test_pl_tf"
                	region = "AWS_US_EAST_1"
                	service = "REDSHIFT"

                	config {
    					aws_account_id = "account_id.cloud_region_name.privatelink.snowflakecomputing.com"
    					cluster_identifier = "account_id.cloud_region_name.privatelink.snowflakecomputing.com"
                 	}
            	}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranPrivateLinkResourceCreate(t, "fivetran_private_link.test_pl"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "name", "test_pl_tf"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "region", "AWS_US_EAST_1"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "service", "SOURCE"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config.aws_account_id", "account_id.cloud_region_name.privatelink.snowflakecomputing.com"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config.cluster_identifier", "account_id.cloud_region_name.privatelink.snowflakecomputing.com"),
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

func testFivetranPrivateLinkResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_private_link" {
			continue
		}

		response, err := client.NewPrivateLinkDetails().PrivateLinkId(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if !strings.HasPrefix(response.Code, "NotFound") {
			return errors.New("Private Link " + rs.Primary.ID + " still exists. Response code: " + response.Code)
		}

	}

	return nil
}
