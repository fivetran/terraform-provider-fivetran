package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var hdaResourceConfig = `
				resource "fivetran_group" "testgroup" {
					provider = fivetran-provider
					name = "%v"
			    }

            	resource "fivetran_hybrid_deployment_agent" "test_lpa" {
                	provider = fivetran-provider

                 	display_name = "%v"
                 	group_id = fivetran_group.testgroup.id
                 	auth_type = "AUTO"
            	}`

var connectorWithHdaResourceConfig = `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "%v"
			    }

            	resource "fivetran_hybrid_deployment_agent" "test_hda" {
                	provider = fivetran-provider

                 	display_name = "%v"
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
		  `

func TestResourceHybridDeploymentAgentE2E(t *testing.T) {
	hdaName := strconv.Itoa(seededRand.Int())
	groupName := strconv.Itoa(seededRand.Int())

	resourceConfig := fmt.Sprintf(hdaResourceConfig, groupName, hdaName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranHybridDeploymentAgentResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:resourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranHybridDeploymentAgentResourceCreate(t, "fivetran_hybrid_deployment_agent.test_lpa"),
					resource.TestCheckResourceAttr("fivetran_hybrid_deployment_agent.test_lpa", "display_name", hdaName),
					resource.TestCheckResourceAttrSet("fivetran_hybrid_deployment_agent.test_lpa", "token"),
				),
			},
		},
	})
}

func TestResourceConnectorWithHybridDeploymentAgentE2E(t *testing.T) {
	regexp, _ := regexp.Compile("[a-z]*_[a-z]*")
	
	hdaName := strconv.Itoa(seededRand.Int())
	groupName := strconv.Itoa(seededRand.Int())

	resourceConfig := fmt.Sprintf(connectorWithHdaResourceConfig, groupName, hdaName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectorResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceConfig,
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
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewHybridDeploymentAgentDetails().AgentId(rs.Primary.ID).Do(context.Background())
		if err != nil {
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
