package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestFivetranGroup_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {},
		Providers:    testProviders,
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
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "creator", "_accountworthy"),
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
					resource.TestCheckResourceAttr("fivetran_group.testgroup", "creator", "_accountworthy"),
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