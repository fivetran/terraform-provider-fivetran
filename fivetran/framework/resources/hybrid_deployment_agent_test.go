package resources_test

import (
	"net/http"
	"testing"
	
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	hybridDeploymentAgentPostHandler   *mock.Handler
	hybridDeploymentAgentDeleteHandler *mock.Handler
	hybridDeploymentAgentData map[string]interface{}
)

func setupMockClientHybridDeploymentAgentResource(t *testing.T) {
	tfmock.MockClient().Reset()
	hybridDeploymentAgentResponse :=
	`{
     	"id": "lpa_id",
       	"display_name": "display_name",
       	"group_id": "group_id",
       	"registered_at": "registered_at",
       	"files": {
          	"config_json": "config_json",
          	"auth_json": "auth_json",
          	"docker_compose_yaml": "docker_compose_yaml"
       	}
    	}`

	hybridDeploymentAgentPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/hybrid-deployment-agents").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			hybridDeploymentAgentData = tfmock.CreateMapFromJsonString(t, hybridDeploymentAgentResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Hybrid Deployment Agent has been created", hybridDeploymentAgentData), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/hybrid-deployment-agents/lpa_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", hybridDeploymentAgentData), nil
		},
	)

	hybridDeploymentAgentDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/hybrid-deployment-agents/lpa_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "Hybrid Deployment Agent has been deleted", nil), nil
		},
	)
}

func TestResourceHybridDeploymentAgentMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_hybrid_deployment_agent" "test_lpa" {
                 provider = fivetran-provider

                 display_name = "display_name"
                 group_id = "group_id"
            }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, hybridDeploymentAgentPostHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_hybrid_deployment_agent.test_lpa", "display_name", "display_name"),
			resource.TestCheckResourceAttr("fivetran_hybrid_deployment_agent.test_lpa", "group_id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_hybrid_deployment_agent.test_lpa", "registered_at", "registered_at"),
			resource.TestCheckResourceAttr("fivetran_hybrid_deployment_agent.test_lpa", "config_json", "config_json"),
			resource.TestCheckResourceAttr("fivetran_hybrid_deployment_agent.test_lpa", "auth_json", "auth_json"),
			resource.TestCheckResourceAttr("fivetran_hybrid_deployment_agent.test_lpa", "docker_compose_yaml", "docker_compose_yaml"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientHybridDeploymentAgentResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, hybridDeploymentAgentDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
