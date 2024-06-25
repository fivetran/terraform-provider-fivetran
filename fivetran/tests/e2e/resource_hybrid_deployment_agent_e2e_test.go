package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

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
            	}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranHybridDeploymentAgentResourceCreate(t, "fivetran_hybrid_deployment_agent.test_lpa"),
					resource.TestCheckResourceAttr("fivetran_hybrid_deployment_agent.test_lpa", "display_name", "TestResourceHybridDeploymentAgentE2E"),
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
			fmt.Println(err)
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
