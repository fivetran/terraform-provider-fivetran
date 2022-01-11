package fivetran_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceUserE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() {},
		Providers:    testProviders,
		CheckDestroy: testFivetranUserResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
		   	resource "fivetran_user" "userjohn" {
				 provider = fivetran-provider
			     email = "john.fox@testmail.com"
			     family_name = "Fox"
			     given_name = "John"
			     phone = "+19876543210"
			     picture = "https://myPicturecom"
			}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranUserResourceCreate(t, "fivetran_user.userjohn"),
					resource.TestCheckResourceAttr("fivetran_user.userjohn", "email", "john.fox@testmail.com"),
					resource.TestCheckResourceAttr("fivetran_user.userjohn", "family_name", "Fox"),
					resource.TestCheckResourceAttr("fivetran_user.userjohn", "given_name", "John"),
					resource.TestCheckResourceAttr("fivetran_user.userjohn", "phone", "+19876543210"),
					resource.TestCheckResourceAttr("fivetran_user.userjohn", "picture", "https://myPicturecom"),
				),
			},
			{
				Config: `
		   	resource "fivetran_user" "userjohn" {
				provider = fivetran-provider
				email = "john.fox@testmail.com"
				family_name = "Connor"
				given_name = "Jane"
				phone = "+19876543219"
				picture = "https://yourPicturecom"
		    }
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranUserResourceUpdate(t, "fivetran_user.userjohn"),
					resource.TestCheckResourceAttr("fivetran_user.userjohn", "email", "john.fox@testmail.com"),
					resource.TestCheckResourceAttr("fivetran_user.userjohn", "family_name", "Connor"),
					resource.TestCheckResourceAttr("fivetran_user.userjohn", "given_name", "Jane"),
					resource.TestCheckResourceAttr("fivetran_user.userjohn", "phone", "+19876543219"),
					resource.TestCheckResourceAttr("fivetran_user.userjohn", "picture", "https://yourPicturecom"),
				),
			},
		},
	})
}

func testFivetranUserResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := Client().NewUserDetails().UserID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields
		return nil
	}
}

func testFivetranUserResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		_, err := Client().NewUserDetails().UserID(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields
		return nil
	}
}

func testFivetranUserResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_user" {
			continue
		}

		response, err := Client().NewUserDetails().UserID(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if response.Code != "NotFound" {
			return errors.New("User " + rs.Primary.ID + " still exists.")
		}

	}

	return nil
}
