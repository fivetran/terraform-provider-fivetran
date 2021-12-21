package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestFivetranGroupUsers_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {},
		Providers:    testProviders,
		CheckDestroy: testFivetranGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
			resource "fivetran_user" "userjohn" {
				provider = fivetran-provider
				email = "john.black@testmail.com"
				family_name = "Black"
				given_name = "John"
				phone = "+19876543210"
				picture = "https://myPicturecom"
			}

		   	resource "fivetran_group" "testgroup" {
				provider = fivetran-provider
			    name = "test_group_name"

				user {
					id = "_seaworthy"
					role = "Admin"
				}

				user {
					id = fivetran_user.userjohn.id
					role = "ReadOnly"
				}
			}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "user.0.id", "_seaworthy"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "user.0.role", "Admin"),
					resource.TestCheckResourceAttrSet("fivetran_group.testgroup", "user.1.id"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "user.1.role", "ReadOnly"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "testgroup" {
					provider = fivetran-provider
					name = "test_group_name"
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranGroupUsersUpdate(t, "fivetran_group.testgroup"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "user.#", "0"),
				),
			},
		},
	})
}

func testFivetranGroupUsersUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		response , err := Client().NewGroupListUsers().GroupID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}

		if len(response.Data.Items) != 1 || response.Data.Items[0].ID != "_accountworthy" {
			return fmt.Errorf("Group has extra users")
		}
		
		return nil
	}
}