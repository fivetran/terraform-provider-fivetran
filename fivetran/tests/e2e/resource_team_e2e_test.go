package e2e_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceTeamsE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranTeamsResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
            resource "fivetran_team" "test_team" {
                 provider = fivetran-provider

                 name = "test_team"
                 description = "test_description"
                 role = "Account Reviewer"
            }
          `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranTeamResourceCreate(t, "fivetran_team.test_team"),
					resource.TestCheckResourceAttr("fivetran_team.test_team", "name", "test_team"),
					resource.TestCheckResourceAttr("fivetran_team.test_team", "description", "test_description"),
					resource.TestCheckResourceAttr("fivetran_team.test_team", "role", "Account Reviewer"),
				),
			},
			{
				Config: `
            resource "fivetran_team" "test_team" {
                 provider = fivetran-provider

                 name = "test_team_2"
                 description = "test_description"
                 role = "Account Reviewer"
            }
          `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranTeamResourceUpdate(t, "fivetran_team.test_team"),
					resource.TestCheckResourceAttr("fivetran_team.test_team", "name", "test_team_2"),
					resource.TestCheckResourceAttr("fivetran_team.test_team", "description", "test_description"),
					resource.TestCheckResourceAttr("fivetran_team.test_team", "role", "Account Reviewer"),
				),
			},
		},
	})
}

func testFivetranTeamResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewTeamsDetails().TeamId(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields
		return nil
	}
}

func testFivetranTeamResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		_, err := client.NewTeamsDetails().TeamId(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields
		return nil
	}
}

func testFivetranTeamsResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_team" {
			continue
		}

		response, err := client.NewTeamsDetails().TeamId(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if response.Code != "NotFound_Team" {
			return errors.New("Team " + rs.Primary.ID + " still exists.")
		}

	}

	return nil
}
