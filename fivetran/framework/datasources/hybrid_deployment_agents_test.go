package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	hybridDeploymentAgentsDataSourceMockGetHandler *mock.Handler
	hybridDeploymentAgentsDataSourceMockData       map[string]interface{}
)

const (
	hybridDeploymentAgentsMappingResponse = `
    {
        "items": [
          {
            "id": "id1",
            "display_name": "display_name1",
            "group_id": "group_id1",
            "registered_at": "registered_at1",
            "usage": [
              {
                "connection_id": "connection_id11",
                "schema": "schema11",
                "service": "service11"
              },
              {
                "connection_id": "connection_id12",
                "schema": "schema12",
                "service": "service12"
              }
            ]
          },
          {
            "id": "id2",
            "display_name": "display_name2",
            "group_id": "group_id2",
            "registered_at": "registered_at2",
            "usage": [
              {
                "connection_id": "connection_id21",
                "schema": "schema21",
                "service": "service21"
              },
              {
                "connection_id": "connection_id22",
                "schema": "schema22",
                "service": "service22"
              }
            ]
          }
        ],
        "next_cursor": null
    }`
)

func setupMockClientHybridDeploymentAgentsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	hybridDeploymentAgentsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/hybrid-deployment-agents").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			hybridDeploymentAgentsDataSourceMockData = tfmock.CreateMapFromJsonString(t, hybridDeploymentAgentsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", hybridDeploymentAgentsDataSourceMockData), nil
		},
	)
}

func TestDataSourceHybridDeploymentAgentsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_hybrid_deployment_agents" "test_lpa" {
            provider = fivetran-provider
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, hybridDeploymentAgentsDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, hybridDeploymentAgentsDataSourceMockData)
				return nil
			},
            resource.TestCheckResourceAttr("data.fivetran_hybrid_deployment_agents.test_lpa", "items.0.display_name", "display_name1"),
            resource.TestCheckResourceAttr("data.fivetran_hybrid_deployment_agents.test_lpa", "items.0.group_id", "group_id1"),
            resource.TestCheckResourceAttr("data.fivetran_hybrid_deployment_agents.test_lpa", "items.0.registered_at", "registered_at1"),
            resource.TestCheckResourceAttr("data.fivetran_hybrid_deployment_agents.test_lpa", "items.1.display_name", "display_name2"),
            resource.TestCheckResourceAttr("data.fivetran_hybrid_deployment_agents.test_lpa", "items.1.group_id", "group_id2"),
            resource.TestCheckResourceAttr("data.fivetran_hybrid_deployment_agents.test_lpa", "items.1.registered_at", "registered_at2"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientHybridDeploymentAgentsDataSourceConfigMapping(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
