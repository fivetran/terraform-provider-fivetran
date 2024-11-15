package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"regexp"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceHybridDeploymentAgentE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranHybridDeploymentAgentResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "testgroup" {
					provider = fivetran-provider
					name = "TestResourceHybridDeploymentAgentE2E"
			    }

            	resource "fivetran_hybrid_deployment_agent" "test_lpa" {
                	provider = fivetran-provider

                 	display_name = "TestResourceHybridDeploymentAgentE2E"
                 	group_id = fivetran_group.testgroup.id
                 	auth_type = "AUTO"
            	}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranHybridDeploymentAgentResourceCreate(t, "fivetran_hybrid_deployment_agent.test_lpa"),
					resource.TestCheckResourceAttr("fivetran_hybrid_deployment_agent.test_lpa", "display_name", "TestResourceHybridDeploymentAgentE2E"),
					resource.TestCheckResourceAttrSet("fivetran_hybrid_deployment_agent.test_lpa", "token"),
				),
			},
		},
	})
}

func TestResourceConnectorWithHybridDeploymentAgentE2E(t *testing.T) {
	regexp, _ := regexp.Compile("[a-z]*_[a-z]*")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectorResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_name"
			    }

            	resource "fivetran_hybrid_deployment_agent" "test_hda" {
                	provider = fivetran-provider

                 	display_name = "TestResourceHybridDeploymentAgentE2E"
                 	group_id = fivetran_group.test_group.id
                 	auth_type = "AUTO"
            	}

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "fivetran_log"
					hybrid_deployment_agent_id = fivetran_hybrid_deployment_agent.test_hda.id
					destination_schema {
						name = "fivetran_log_schema"
					}
					
					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_connector"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "service", "fivetran_log"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "name", "fivetran_log_schema"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_certificates", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "trust_fingerprints", "false"),
					resource.TestCheckResourceAttr("fivetran_connector.test_connector", "run_setup_tests", "false"),
					resource.TestMatchResourceAttr("fivetran_connector.test_connector", "hybrid_deployment_agent_id", regexp),
				),
			},
		},
	})
}

func testFivetranHybridDeploymentAgentResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	fmt.Printf("sadasasas")

	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		fmt.Printf("sdfsdsdsdf %v", rs.Primary.ID)

		_, err := client.NewHybridDeploymentAgentDetails().AgentId(rs.Primary.ID).Do(context.Background())
		if err != nil {
			fmt.Printf("sdfsdsdsdf %v %v", rs.Primary.ID, err)
			return err
		}
		//todo: check response _  fields if needed
		return nil
	}
}

func testFivetranHybridDeploymentAgentResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_hybrid_deployment_agent" {
			continue
		}

		response, err := client.NewHybridDeploymentAgentDetails().AgentId(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if !strings.HasPrefix(response.Code, "NotFound") {
			return errors.New("Hybrid Deployment Agent " + rs.Primary.ID + " still exists. Response code: " + response.Code)
		}

	}

	return nil
}
