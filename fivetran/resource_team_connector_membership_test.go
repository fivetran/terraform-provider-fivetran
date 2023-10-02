package fivetran_test

import (
    "context"
    "errors"
    "testing"

    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceTeamConnectorMembershipE2E(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:     func() {},
        Providers:    testProviders,
        CheckDestroy: testFivetranTeamConnectorMembershipResourceDestroy,
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

            resource "fivetran_connector" "test_connector" {
                provider = fivetran-provider
                group_id = fivetran_group.test_group.id
                service = "fivetran_log"
                destination_schema {
                    name = "fivetran_log_schema"
                }
                    
                trust_certificates = false
                trust_fingerprints = false
                run_setup_tests = false
            
                config {
                    group_name = fivetran_group.test_group.name
                }
            }

            resource "fivetran_team_connector_membership" "test_team_connector_membership" {
                 provider = fivetran-provider

                 team_id = fivetran_team.testteam.id

                 connector {
                    connector_id = fivetran_connector.test_connector.id
                    role = "Connector Reviewer"                    
                 }
            }
          `,
                Check: resource.ComposeAggregateTestCheckFunc(
                    testFivetranTeamConnectorMembershipResourceCreate(t, "fivetran_team_connector_membership.test_team_connector_membership"),
                    resource.TestCheckResourceAttrSet("fivetran_team_connector_membership.test_team_connector_membership", "team_id"),
                    resource.TestCheckResourceAttrSet("fivetran_team_connector_membership.test_team_connector_membership", "connector.0.connector_id"),
                    resource.TestCheckResourceAttr("fivetran_team_connector_membership.test_team_connector_membership", "connector.0.role", "Connector Reviewer"),
                    resource.TestCheckResourceAttr("fivetran_team_connector_membership.test_team_connector_membership", "connector.#", "1"),
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

            resource "fivetran_connector" "test_connector" {
                provider = fivetran-provider
                group_id = fivetran_group.test_group.id
                service = "fivetran_log"
                destination_schema {
                    name = "fivetran_log_schema"
                }
                    
                trust_certificates = false
                trust_fingerprints = false
                run_setup_tests = false
            
                config {
                    group_name = fivetran_group.test_group.name
                }
            }

            resource "fivetran_team_connector_membership" "test_team_connector_membership" {
                 provider = fivetran-provider

                 team_id = fivetran_team.testteam.id

                 connector {
                    connector_id = fivetran_connector.test_connector.id
                    role = "Connector Administrator"
                 }
            }
          `,
                Check: resource.ComposeAggregateTestCheckFunc(
                    testFivetranTeamConnectorMembershipResourceCreate(t, "fivetran_team_connector_membership.test_team_connector_membership"),
                    resource.TestCheckResourceAttrSet("fivetran_team_connector_membership.test_team_connector_membership", "team_id"),
                    resource.TestCheckResourceAttrSet("fivetran_team_connector_membership.test_team_connector_membership", "connector.0.connector_id"),
                    resource.TestCheckResourceAttr("fivetran_team_connector_membership.test_team_connector_membership", "connector.0.role", "Connector Administrator"),
                    resource.TestCheckResourceAttr("fivetran_team_connector_membership.test_team_connector_membership", "connector.#", "1"),
                ),
            },
        },
    })
}


func testFivetranTeamConnectorMembershipResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        rs := GetResource(t, s, resourceName)

        response, err := client.NewTeamConnectorMembershipsList().
            TeamId(rs.Primary.ID).
            Do(context.Background())

        if err != nil {
            return err
        }

        if response.Code == "NotFound" || len(response.Data.Items) == 0 {
            return errors.New("Team connector membership didn't created.")
        }

        //todo: check response _  fields
        return nil
    }
}

func testFivetranTeamConnectorMembershipResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        rs := GetResource(t, s, resourceName)
        response, err := client.NewTeamConnectorMembershipsList().
            TeamId(rs.Primary.ID).
            Do(context.Background())

        if err != nil {
            return err
        }

        for _, value := range response.Data.Items {
            if value.Role == "Connector Reviewer" {
                return nil
            }
        }

        return errors.New("Team connector membership " + rs.Primary.ID + " didn't updated.")
    }
}

func testFivetranTeamConnectorMembershipResourceDestroy(s *terraform.State) error {
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "fivetran_team" {
            continue
        }

        response, err := client.NewTeamConnectorMembershipsList().
            TeamId(rs.Primary.ID).
            Do(context.Background())

        if err.Error() != "status code: 404; expected: 200" {
            return err
        }
        if response.Code != "NotFound_Team" || len(response.Data.Items) > 0 {
            return errors.New("Team connector membership " + rs.Primary.ID + " still exists.")
        }

    }

    return nil
}
