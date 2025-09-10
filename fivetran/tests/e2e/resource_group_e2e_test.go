package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var groupResourceConfig = `
					resource "fivetran_group" "testgroup" {
						provider = fivetran-provider
						name = "%v"
					}`

var groupResourceWithUsersConfig = `
					resource "fivetran_user" "userjohn" {
						provider = fivetran-provider
						email = "%v"
						family_name = "Black"
						given_name = "John"
						phone = "+19876543210"
						picture = "https://myPicturecom"
						role = "Account Reviewer"
					}

					resource "fivetran_group" "testgroup" {
						provider = fivetran-provider
						name = "%v"
					}

					resource "fivetran_group_users" "testgroup_users" {
						provider = fivetran-provider
						group_id = fivetran_group.testgroup.id

						user {
							email = fivetran_user.userjohn.email
							role = "%v"
						}
					}
				`

var groupResourceWithEmptyUsersConfig = `
					resource "fivetran_user" "userjohn" {
						provider = fivetran-provider
						email = "%v"
						family_name = "Black"
						given_name = "John"
						phone = "+19876543210"
						picture = "https://myPicturecom"
						role = "Account Reviewer"
					}

					resource "fivetran_group" "testgroup" {
						provider = fivetran-provider
						name = "%v"
					}

					resource "fivetran_group_users" "testgroup_users" {
						provider = fivetran-provider
						group_id = fivetran_group.testgroup.id
					}
				`

func TestResourceGroupE2E(t *testing.T) {
	suffix := strconv.Itoa(seededRand.Int())
	groupCreateName := "TestResourceGroupE2E" + suffix + "created"
	groupUpdateName := "TestResourceGroupE2E" + suffix + "updated"

	resourceCreateConfig := fmt.Sprintf(groupResourceConfig, groupCreateName)
	resourceUpdateConfig := fmt.Sprintf(groupResourceConfig, groupUpdateName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceCreateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranGroupResourceCreate(t, "fivetran_group.testgroup"),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", groupCreateName),
					resource.TestCheckResourceAttrSet("fivetran_group.testgroup", "created_at"),
				),
			},
			{
				Config: resourceUpdateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranGroupResourceUpdate(t, "fivetran_group.testgroup", groupUpdateName),
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", groupUpdateName),
					resource.TestCheckResourceAttrSet("fivetran_group.testgroup", "created_at"),
					resource.TestCheckResourceAttrSet("fivetran_group.testgroup", "last_updated"),
				),
			},
		},
	})
}

func TestResourceGroupWithUsersE2E(t *testing.T) {
	suffix := strconv.Itoa(seededRand.Int())
	groupName := "TestResourceGroupE2E" + suffix + "created"
	userName := "john.black" + suffix + "@testmail.com"
	roleCreate := "Destination Reviewer"
	roleUpdate := "Manage Destination"

	resourceWithUsersCreateConfig := fmt.Sprintf(groupResourceWithUsersConfig, userName, groupName, roleCreate)
	resourceWithUsersUpdateConfig := fmt.Sprintf(groupResourceWithUsersConfig, userName, groupName, roleUpdate)
	resourceWithUsersEmptyConfig := fmt.Sprintf(groupResourceWithEmptyUsersConfig, userName, groupName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranGroupResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceWithUsersCreateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fivetran_group_users.testgroup_users", "user.0.id"),
					resource.TestCheckResourceAttr("fivetran_group_users.testgroup_users", "user.0.role", roleCreate),
					resource.TestCheckResourceAttr("fivetran_group_users.testgroup_users", "user.0.email", userName),
				),
			},
			{
				Config: resourceWithUsersUpdateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fivetran_group_users.testgroup_users", "user.0.id"),
					resource.TestCheckResourceAttr("fivetran_group_users.testgroup_users", "user.0.role", roleUpdate),
					resource.TestCheckResourceAttr("fivetran_group_users.testgroup_users", "user.0.email", userName),
				),
			},
			{
				Config: resourceWithUsersEmptyConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", groupName),
					resource.TestCheckResourceAttr("fivetran_group_users.testgroup_users", "user.#", "0"),
				),
			},
		},
	})
}

func testFivetranGroupResourceUpdate(t *testing.T, resourceName string, groupName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		response, err := client.NewGroupDetails().GroupID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}

		if response.Data.Name != groupName {
			return fmt.Errorf("Group has name %v different from expected (%v)", response.Data.Name, groupName)
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
