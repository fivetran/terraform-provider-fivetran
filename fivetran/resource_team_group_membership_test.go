package fivetran_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceTeamGroupMembershipE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranTeamGroupMembershipResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
            resource "fivetran_team" "testteam" {
                provider = fivetran-provider
                name = "test_team"
                description = "test_team"
                role = "Account Analyst"
            }

            resource "fivetran_group" "test_group" {
                provider = fivetran-provider
                name = "test_group_name"
            }

            resource "fivetran_team_group_membership" "test_team_group_membership" {
                 provider = fivetran-provider

                 team_id = fivetran_team.testteam.id

                 group {
                    group_id = fivetran_group.test_group.id
                    role = "Destination Administrator"                    
                 }
            }
          `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranTeamGroupMembershipResourceCreate(t, "fivetran_team_group_membership.test_team_group_membership"),
					resource.TestCheckResourceAttrSet("fivetran_team_group_membership.test_team_group_membership", "team_id"),
					resource.TestCheckResourceAttrSet("fivetran_team_group_membership.test_team_group_membership", "group.0.group_id"),
					resource.TestCheckResourceAttr("fivetran_team_group_membership.test_team_group_membership", "group.0.role", "Destination Administrator"),
					resource.TestCheckResourceAttr("fivetran_team_group_membership.test_team_group_membership", "group.#", "1"),
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

            resource "fivetran_group" "test_group" {
                provider = fivetran-provider
                name = "test_group_name"
            }

            resource "fivetran_team_group_membership" "test_team_group_membership" {
                 provider = fivetran-provider

                 team_id = fivetran_team.testteam.id

                 group {
                    group_id = fivetran_group.test_group.id
                    role = "Destination Reviewer"
                 }
            }
          `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranTeamGroupMembershipResourceCreate(t, "fivetran_team_group_membership.test_team_group_membership"),
					resource.TestCheckResourceAttrSet("fivetran_team_group_membership.test_team_group_membership", "team_id"),
					resource.TestCheckResourceAttrSet("fivetran_team_group_membership.test_team_group_membership", "group.0.group_id"),
					resource.TestCheckResourceAttr("fivetran_team_group_membership.test_team_group_membership", "group.0.role", "Destination Reviewer"),
					resource.TestCheckResourceAttr("fivetran_team_group_membership.test_team_group_membership", "group.#", "1"),
				),
			},
		},
	})
}

func testFivetranTeamGroupMembershipResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		response, err := client.NewTeamGroupMembershipsList().
			TeamId(rs.Primary.ID).
			Do(context.Background())

		if err != nil {
			return err
		}

		if response.Code == "NotFound" || len(response.Data.Items) == 0 {
			return errors.New("Team group membership didn't created.")
		}

		//todo: check response _  fields
		return nil
	}
}

// func testFivetranTeamGroupMembershipResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs := GetResource(t, s, resourceName)
// 		response, err := client.NewTeamGroupMembershipsList().
// 			TeamId(rs.Primary.ID).
// 			Do(context.Background())

// 		if err != nil {
// 			return err
// 		}

// 		for _, value := range response.Data.Items {
// 			if value.Role == "Destination Administrator" {
// 				return nil
// 			}
// 		}

// 		return errors.New("Team group membership " + rs.Primary.ID + " didn't updated.")
// 	}
// }

func testFivetranTeamGroupMembershipResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_team" {
			continue
		}

		response, err := client.NewTeamGroupMembershipsList().
			TeamId(rs.Primary.ID).
			Do(context.Background())

		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if response.Code != "NotFound_Team" || len(response.Data.Items) > 0 {
			return errors.New("Team group membership " + rs.Primary.ID + " still exists.")
		}

	}

	return nil
}
