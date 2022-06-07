package fivetran_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceGroupE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() {},
		ProviderFactories: providerFactory,
		CheckDestroy:      testFivetranGroupResourceDestroy,
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
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "creator", PredefinedUserId),
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
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "creator", PredefinedUserId),
				),
			},
		},
	})
}

func TestResourceGroupWithUsersE2E(t *testing.T) {
	//t.Skip("Endpoint to add user to group doesn't support new RBAC role names. It will be fixed soon")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() {},
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
						role = "Account Reviewer"
					}

					resource "fivetran_group" "testgroup" {
						provider = fivetran-provider
						name = "test_group_name"
					}

					resource "fivetran_group_users" "testgroup_users" {
						provider = fivetran-provider
						group_id = fivetran_group.testgroup.id

						user {
							id = fivetran_user.userjohn.id
							role = "Destination Reviewer"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fivetran_group_users.testgroup_users", "user.0.id"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup_users", "user.0.role", "Destination Reviewer"),
				),
			},
			{
				Config: `
					resource "fivetran_user" "userjohn" {
						provider = fivetran-provider
						email = "john.black@testmail.com"
						family_name = "Black"
						given_name = "John"
						phone = "+19876543210"
						picture = "https://myPicturecom"
						role = "Account Reviewer"
					}

					resource "fivetran_group" "testgroup" {
						provider = fivetran-provider
						name = "test_group_name"
					}

					resource "fivetran_group_users" "testgroup_users" {
						provider = fivetran-provider
						group_id = fivetran_group.testgroup.id

						user {
							id = fivetran_user.userjohn.id
							role = "Destination Administrator"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fivetran_group_users.testgroup_users", "user.0.id"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup_users", "user.0.role", "Destination Administrator"),
				),
			},
			{
				Config: `
					resource "fivetran_group" "testgroup" {
						provider = fivetran-provider
						name = "test_group_name"
					}

					resource "fivetran_group_users" "testgroup_users" {
						provider = fivetran-provider
						group_id = fivetran_group.testgroup.id
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranGroupUsersUpdate(t, "fivetran_group.testgroup"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name"),
					resource.TestCheckResourceAttr("fivetran_group_users.testgroup_users", "user.#", "0"),
				),
			},
		},
	})
}

func testFivetranGroupResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		_, err := client.NewGroupDetails().GroupID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}

		response, err := client.NewGroupListUsers().GroupID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}

		if len(response.Data.Items) != 0 {
			return fmt.Errorf("Group has extra " + strconv.Itoa(len(response.Data.Items)) + " users (" + response.Data.Items[0].ID + ")")
		}

		return nil
	}
}

func testFivetranGroupResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		_, err := client.NewGroupDetails().GroupID(rs.Primary.ID).Do(context.Background())

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

		response, err := client.NewGroupDetails().GroupID(rs.Primary.ID).Do(context.Background())
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
		response, err := client.NewGroupListUsers().GroupID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}

		if len(response.Data.Items) != 0 {
			return fmt.Errorf("Group has extra " + strconv.Itoa(len(response.Data.Items)) + " users (" + response.Data.Items[0].ID + ")")
		}

		return nil
	}
}
