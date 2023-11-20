package fivetran_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceTeamUserMembershipE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {},
		//Providers:    testProviders,
		ProtoV5ProviderFactories: protoV5ProviderFactory,
		CheckDestroy:             testFivetranTeamUserMembershipResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
            resource "fivetran_team" "testteam" {
                provider = fivetran-provider
                name = "test_team"
                description = "test_team"
                role = "Account Analyst"
            }

            resource "fivetran_user" "test_user" {
                provider = fivetran-provider
                role = "Account Administrator"
                email = "john.fox@testmail.com"
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
                name = "test_team"
                description = "test_team"
                role = "Account Analyst"
            }

            resource "fivetran_user" "test_user" {
                provider = fivetran-provider
                role = "Account Administrator"
                email = "john.fox@testmail.com"
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

// func testFivetranTeamUserMembershipResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs := GetResource(t, s, resourceName)
// 		response, err := client.NewTeamUserMembershipsList().
// 			TeamId(rs.Primary.ID).
// 			Do(context.Background())

// 		if err != nil {
// 			return err
// 		}

// 		for _, value := range response.Data.Items {
// 			if value.Role == "Team Manager" {
// 				return nil
// 			}
// 		}

// 		return errors.New("Team user membership " + rs.Primary.ID + " didn't updated.")
// 	}
// }

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
