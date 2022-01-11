package fivetran_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceGroupE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() {},
		ProviderFactories: providerFactory,
		CheckDestroy: testFivetranGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
		   	resource "fivetran_group" "testgroup" {
				 provider = fivetran-provider
			     name = "test_group_name"
			}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranGroupResourceCreate(t, "fivetran_group.testgroup"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name"),
					resource.TestCheckResourceAttrSet("fivetran_group.testgroup", "created_at"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "creator", "cherry_spoilt"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "testgroup" {
					provider = fivetran-provider
					name = "test_group_name_updated"
			    }
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranGroupResourceUpdate(t, "fivetran_group.testgroup"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name_updated"),
					resource.TestCheckResourceAttrSet("fivetran_group.testgroup", "created_at"),
					resource.TestCheckResourceAttrSet("fivetran_group.testgroup", "last_updated"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "creator", "cherry_spoilt"),
				),
			},
		},
	})
}

func TestResourceGroupWithUsersE2E(t *testing.T) {
	t.Skip("Endpoint to add user to group doesn't support new RBAC role names. It will be fixed soon")
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
					id = "cherry_spoilt"
					role = "Account Administrator"
				}

				user {
					id = fivetran_user.userjohn.id
					role = "Account Reviewer"
				}
			}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "user.0.id", "cherry_spoilt"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "user.0.role", "Admin"),
					resource.TestCheckResourceAttrSet("fivetran_group.testgroup", "user.1.id"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "user.1.role", "Account Reviewer"),
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

func testFivetranGroupResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		_, err := Client().NewGroupDetails().GroupID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields
		return nil
	}
}

func testFivetranGroupResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		_, err := Client().NewGroupDetails().GroupID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields
		return nil
	}
}

func testFivetranGroupResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_group" {
			continue
		}

		response, err := Client().NewGroupDetails().GroupID(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if response.Code != "NotFound" {
			return errors.New("Group " + rs.Primary.ID + " still exists.")
		}

	}

	return nil
}

func testFivetranGroupUsersUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		response , err := Client().NewGroupListUsers().GroupID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}

		if len(response.Data.Items) != 1 || response.Data.Items[0].ID != "cherry_spoilt" {
			return fmt.Errorf("Group has extra users")
		}
		
		return nil
	}
}