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

func TestResourceProxyE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranProxyResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
            	resource "fivetran_proxy" "test_proxy" {
                	provider = fivetran-provider

                 	display_name = "display_name"
                 	group_region = "group_region"
            	}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranProxyResourceCreate(t, "fivetran_proxy.test_proxy"),
					resource.TestCheckResourceAttr("fivetran_proxy.test_proxy", "display_name", "display_name"),
					resource.TestCheckResourceAttr("fivetran_proxy.test_proxy", "group_region", "group_region"),
				),
			},
		},
	})
}

func testFivetranProxyResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewProxyDetails().ProxyId(rs.Primary.ID).Do(context.Background())
		if err != nil {
			fmt.Println(err)
			return err
		}
		//todo: check response _  fields if needed
		return nil
	}
}

func testFivetranProxyResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_proxy" {
			continue
		}

		response, err := client.NewProxyDetails().ProxyId(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if !strings.HasPrefix(response.Code, "NotFound") {
			return errors.New("Proxy " + rs.Primary.ID + " still exists. Response code: " + response.Code)
		}

	}

	return nil
}
