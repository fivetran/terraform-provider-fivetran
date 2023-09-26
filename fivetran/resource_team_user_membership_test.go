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
        PreCheck:     func() {},
        Providers:    testProviders,
        CheckDestroy: testFivetranTeamUserMembershipResourceDestroy,
        Steps: []resource.TestStep{
            {
                Config: `
            resource "fivetran_team" "testteam" {
                provider = fivetran-provider
                name = "test_team"
                description = "test_team"
                role = "Account Analyst"
            }

            resource "fivetran_user" "testuser" {
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
                 user_id = fivetran_user.testuser.id
                 role = "Team Member"
            }
          `,
                Check: resource.ComposeAggregateTestCheckFunc(
                    testFivetranTeamUserMembershipResourceCreate(t, "fivetran_team_user_membership.test_team_user_membership"),
                    resource.TestCheckResourceAttrSet("fivetran_team_user_membership.test_team_user_membership", "team_id"),
                    resource.TestCheckResourceAttrSet("fivetran_team_user_membership.test_team_user_membership", "user_id"),
                    resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "role", "Team Member"),
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

            resource "fivetran_user" "testuser" {
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
                 user_id = fivetran_user.testuser.id
                 role = "Team Manager"
            }
          `,
                Check: resource.ComposeAggregateTestCheckFunc(
                    testFivetranTeamUserMembershipResourceUpdate(t, "fivetran_team_user_membership.test_team_user_membership"),
                    resource.TestCheckResourceAttrSet("fivetran_team_user_membership.test_team_user_membership", "team_id"),
                    resource.TestCheckResourceAttrSet("fivetran_team_user_membership.test_team_user_membership", "user_id"),
                    resource.TestCheckResourceAttr("fivetran_team_user_membership.test_team_user_membership", "role", "Team Manager"),
                ),
            },
        },
    })
}

func testFivetranTeamUserMembershipResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        rs := GetResource(t, s, resourceName)

        _, err := client.NewTeamUserMembershipDetails().
            TeamId(rs.Primary.Attributes["team_id"]).
            UserId(rs.Primary.Attributes["user_id"]).
            Do(context.Background())

        if err != nil {
            return err
        }
        //todo: check response _  fields
        return nil
    }
}

func testFivetranTeamUserMembershipResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        rs := GetResource(t, s, resourceName)
        _, err := client.NewTeamUserMembershipDetails().
            TeamId(rs.Primary.Attributes["team_id"]).
            UserId(rs.Primary.Attributes["user_id"]).
            Do(context.Background())

        if err != nil {
            return err
        }
        //todo: check response _  fields
        return nil
    }
}

func testFivetranTeamUserMembershipResourceDestroy(s *terraform.State) error {
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "fivetran_team_user_membership" {
            continue
        }

        response, err := client.NewTeamUserMembershipDetails().
            TeamId(rs.Primary.Attributes["team_id"]).
            UserId(rs.Primary.Attributes["user_id"]).
            Do(context.Background())

        if err.Error() != "status code: 404; expected: 200" {
            return err
        }
        if response.Code != "NotFound" && response.Code != "NotFound_Team" {
            return errors.New("Team User memebrship " + rs.Primary.ID + " still exists.")
        }

    }

    return nil
}
