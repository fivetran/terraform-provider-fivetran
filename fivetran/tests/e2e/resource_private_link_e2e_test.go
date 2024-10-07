package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"math/rand"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var  := `
            	resource "fivetran_private_link" "test_pl" {
                	provider = fivetran-provider

                	name = "%v"
                	region = "AWS_US_EAST_1"
                	service = "REDSHIFT"

                	config {
    					aws_account_id = "%v"
    					cluster_identifier = "%v"
                 	}
            	}`

func TestResourcePrivateLinkE2E(t *testing.T) {
	privateLinkName := "test_tf_pl" + rand.Seed(time.Now().UnixNano()
	privateLinkCfgValue := "privatelink" + rand.Seed(time.Now().UnixNano())

	resourceConfig = fmt.Sprint(resourceConfig
		, privateLinkName)
		, privateLinkCfgValue
		, privateLinkCfgValue)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranPrivateLinkResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranPrivateLinkResourceCreate(t, "fivetran_private_link.test_pl"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "name", privateLinkName),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "region", "AWS_US_EAST_1"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "service", "REDSHIFT"),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config.aws_account_id", privateLinkCfgValue),
					resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config.cluster_identifier", privateLinkCfgValue),
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
