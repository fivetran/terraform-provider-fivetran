package fivetran_test

import (
    "context"
    "errors"
    "testing"

    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceTeamGroupMembershipE2E(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:     func() {},
        Providers:    testProviders,
        CheckDestroy: testFivetranTeamGroupMembershipResourceDestroy,
        Steps: []resource.TestStep{
            {
                Config: `
            resource "fivetran_resource_team_group_membership" "test_resource_team_group_membership" {
                 provider = fivetran-provider

                 team_id = "test_team"
                 group_id = "test_group"
                 role = "Destination Administrator"
            }
          `,
                Check: resource.ComposeAggregateTestCheckFunc(
                    testFivetranTeamGroupMembershipResourceCreate(t, "fivetran_resource_team_group_membership.test_resource_team_group_membership"),
                    resource.TestCheckResourceAttr("fivetran_resource_team_group_membership.test_resource_team_group_membership", "team_id", "test_team"),
                    resource.TestCheckResourceAttr("fivetran_resource_team_group_membership.test_resource_team_group_membership", "group_id", "test_group"),
                    resource.TestCheckResourceAttr("fivetran_resource_team_group_membership.test_resource_team_group_membership", "role", "Destination Administrator"),
                ),
            },
            {
                Config: `
            resource "fivetran_resource_team_group_membership" "test_resource_team_group_membership" {
                 provider = fivetran-provider

                 team_id = "test_team"
                 group_id = "test_group"
                 role = "Destination Reviewer"
            }
          `,
                Check: resource.ComposeAggregateTestCheckFunc(
                    testFivetranTeamGroupMembershipResourceUpdate(t, "fivetran_resource_team_group_membership.test_resource_team_group_membership"),
                    resource.TestCheckResourceAttr("fivetran_resource_team_group_membership.test_resource_team_group_membership", "team_id", "test_team"),
                    resource.TestCheckResourceAttr("fivetran_resource_team_group_membership.test_resource_team_group_membership", "group_id", "test_group"),
                    resource.TestCheckResourceAttr("fivetran_resource_team_group_membership.test_resource_team_group_membership", "role", "Destination Reviewer"),
                ),
            },
        },
    })
}

func testFivetranTeamGroupMembershipResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        rs := GetResource(t, s, resourceName)

        _, err := client.NewTeamGroupMembershipDetails().
            TeamId(rs.Primary.Attributes["team_id"]).
            GroupId(rs.Primary.Attributes["group_id"]).
            Do(context.Background())

        if err != nil {
            return err
        }
        //todo: check response _  fields
        return nil
    }
}

func testFivetranTeamGroupMembershipResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        rs := GetResource(t, s, resourceName)
        _, err := client.NewTeamGroupMembershipDetails().
            TeamId(rs.Primary.Attributes["team_id"]).
            GroupId(rs.Primary.Attributes["group_id"]).
            Do(context.Background())

        if err != nil {
            return err
        }
        //todo: check response _  fields
        return nil
    }
}

func testFivetranTeamGroupMembershipResourceDestroy(s *terraform.State) error {
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "fivetran_team" {
            continue
        }

        response, err := client.NewTeamGroupMembershipDetails().
            TeamId(rs.Primary.Attributes["team_id"]).
            GroupId(rs.Primary.Attributes["group_id"]).
            Do(context.Background())

        if err.Error() != "status code: 404; expected: 200" {
            return err
        }
        if response.Code != "NotFound" {
            return errors.New("Team Group memebrship " + rs.Primary.ID + " still exists.")
        }

    }

    return nil
}
