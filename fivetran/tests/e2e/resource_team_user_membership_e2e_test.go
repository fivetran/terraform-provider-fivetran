package e2e_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceTeamUserMembershipE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranTeamUserMembershipResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
            resource "fivetran_team" "testteam" {
                provider = fivetran-provider
                name = "TestResourceTeamUserMembershipE2E"
                description = "test_team_5"
                role = "Account Analyst"
            }

            resource "fivetran_user" "test_user" {
                provider = fivetran-provider
                role = "Account Administrator"
                email = "TestResourceTeamUserMembershipE2E@testmail.com"
                family_name = "Connor"
                given_name = "Jane"
                phone = "+19876543219"
                picture = "https://yourPicturecom"
            }

            resource "fivetran_team_user_membership" "test_team_user_membership" {
                 provider = fivetran-provider

                 team_id = fivetran_team.testteam.id

                 user {
                    user_id = fivetran_user.test_user.id
                    role = "Team Member"
                 }
            }
          `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranTeamUserMembershipResourceCreate(t, "fivetran_team_user_membership.test_team_user_membership"),
					resource.TestCheckResourceAttrSet("fivetran_team_user_membership.test_team_user_membership", "team_id"),
					resource.TestCheckResourceAttrSet("fivetran_team_user_membership.test_team_user_membership", "user.0.user_id"),
					resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "user.0.role", "Team Member"),
					resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "user.#", "1"),
				),
			},
			{
				Config: `
            resource "fivetran_team" "testteam" {
                provider = fivetran-provider
                name = "TestResourceTeamUserMembershipE2E"
                description = "test_team_6"
                role = "Account Analyst"
            }

            resource "fivetran_user" "test_user" {
                provider = fivetran-provider
                role = "Account Administrator"
                email = "TestResourceTeamUserMembershipE2E@testmail.com"
                family_name = "Connor"
                given_name = "Jane"
                phone = "+19876543219"
                picture = "https://yourPicturecom"
            }

            resource "fivetran_team_user_membership" "test_team_user_membership" {
                 provider = fivetran-provider

                 team_id = fivetran_team.testteam.id

                 user {
                    user_id = fivetran_user.test_user.id
                    role = "Team Manager"
                 }
            }
          `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranTeamUserMembershipResourceCreate(t, "fivetran_team_user_membership.test_team_user_membership"),
					resource.TestCheckResourceAttrSet("fivetran_team_user_membership.test_team_user_membership", "team_id"),
					resource.TestCheckResourceAttrSet("fivetran_team_user_membership.test_team_user_membership", "user.0.user_id"),
					resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "user.0.role", "Team Manager"),
					resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "user.#", "1"),
				),
			},
		},
	})
}

func testFivetranTeamUserMembershipResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		response, err := client.NewTeamUserMembershipsList().
			TeamId(rs.Primary.ID).
			Do(context.Background())

		if err != nil {
			return err
		}

		if response.Code == "NotFound" || len(response.Data.Items) == 0 {
			return errors.New("Team user membership didn't created.")
		}

		//todo: check response _  fields
		return nil
	}
}

func testFivetranTeamUserMembershipResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_team" {
			continue
		}

		response, err := client.NewTeamUserMembershipsList().
			TeamId(rs.Primary.ID).
			Do(context.Background())

		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if response.Code != "NotFound_Team" || len(response.Data.Items) > 0 {
			return errors.New("Team user membership " + rs.Primary.ID + " still exists.")
		}

	}

	return nil
}
