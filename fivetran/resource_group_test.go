package fivetran_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceGroupE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranGroupResourceDestroy,
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
				),
			},
		},
	})
}

func TestResourceGroupWithUsersE2E(t *testing.T) {
	//t.Skip("Endpoint to add user to group doesn't support new RBAC role names. It will be fixed soon")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranGroupResourceDestroy,
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
							email = fivetran_user.userjohn.email
							role = "Destination Reviewer"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranGroupUsersUpdateWithUsers(t, "fivetran_group_users.testgroup_users", []string{"john.black@testmail.com"}),
					resource.TestCheckResourceAttrSet("fivetran_group_users.testgroup_users", "user.0.id"),
					resource.TestCheckResourceAttr("fivetran_group_users.testgroup_users", "user.0.role", "Destination Reviewer"),
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
							email = fivetran_user.userjohn.email
							role = "Destination Administrator"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranGroupUsersUpdateWithUsers(t, "fivetran_group_users.testgroup_users", []string{"john.black@testmail.com"}),
					resource.TestCheckResourceAttrSet("fivetran_group_users.testgroup_users", "user.0.id"),
					resource.TestCheckResourceAttr("fivetran_group_users.testgroup_users", "user.0.role", "Destination Administrator"),
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
					testFivetranGroupUsersUpdate(t, "fivetran_group_users.testgroup_users"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name"),
					resource.TestCheckResourceAttr("fivetran_group_users.testgroup_users", "user.#", "0"),
				),
			},
		},
	})
}

func testFivetranGroupResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		response, err := client.NewGroupDetails().GroupID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}

		if response.Data.Name != "test_group_name_updated" {
			return fmt.Errorf("Group has name %v different from expected (%v)", response.Data.Name, "test_group_name_updated")
		}

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
		if response.Code != "NotFound_Group" {
			return errors.New("Group " + rs.Primary.ID + " still exists. Response code: " + response.Code)
		}

	}

	return nil
}

func testFivetranGroupUsersUpdateWithUsers(t *testing.T, resourceName string, expectedUsers []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		response, err := client.NewGroupListUsers().GroupID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}

		var actualUsers []string
		difference := false
		for _, user := range response.Data.Items {
			if user.Role == "" {
				continue
			}
			actualUsers = append(actualUsers, user.Email)
			found := false
			for _, expectedUser := range expectedUsers {
				if expectedUser == user.Email {
					found = true
				}
			}
			if !found {
				difference = true
			}
		}

		if difference || len(actualUsers) != len(expectedUsers) {
			return fmt.Errorf("Group users different from expected. Was: (" + strings.Join(actualUsers, ",") + "), expected: (" + strings.Join(expectedUsers, ",") + ")")
		}

		return nil
	}
}

func testFivetranGroupUsersUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		response, err := client.NewGroupListUsers().GroupID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}

		var users []string
		for _, user := range response.Data.Items {
			if user.Role == "" {
				continue
			}
			users = append(users, user.ID)
		}

		if len(users) != 0 {
			return fmt.Errorf("Group has extra " + strconv.Itoa(len(users)) + " users (" + strings.Join(users, ",") + ")")
		}

		return nil
	}
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

		var users []string
		for _, user := range response.Data.Items {
			// when group just created it contains one membership for group creator
			if user.ID == PredefinedUserId {
				continue
			}
			users = append(users, user.ID)
		}

		if len(users) != 0 {
			return fmt.Errorf("Group has extra " + strconv.Itoa(len(users)) + " users (" + strings.Join(users, ",") + ")")
		}

		return nil
	}
}
