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
    hdaDataSourceMockGetHandler *mock.Handler
    hdaDataSourceMockData map[string]interface{}
)

const (
    hdaMappingResponse = `
    {
        "id": "lpa_id",
        "display_name": "display_name",
        "group_id": "group_id",
        "registered_at": "registered_at",
        "usage": [
        {
            "connection_id": "connection_id1",
            "schema": "schema1",
            "service": "service1"
        },
        {
            "connection_id": "connection_id2",
            "schema": "schema2",
            "service": "service2"
        }
        ]
    }
    `
)

func setupMockClientHybridDeploymentAgentDataSourceConfigMapping(t *testing.T) {
    tfmock.MockClient().Reset()

    hdaDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/hybrid-deployment-agents/lpa_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            hdaDataSourceMockData = tfmock.CreateMapFromJsonString(t, hdaMappingResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", hdaDataSourceMockData), nil
        },
    )
}

func TestDataSourceHybridDeploymentAgentMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        data "fivetran_hybrid_deployment_agent" "test_lpa" {
            provider = fivetran-provider
            id = "lpa_id"
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, hdaDataSourceMockGetHandler.Interactions, 1)
                tfmock.AssertNotEmpty(t, hdaDataSourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("data.fivetran_hybrid_deployment_agent.test_lpa", "display_name", "display_name"),
            resource.TestCheckResourceAttr("data.fivetran_hybrid_deployment_agent.test_lpa", "group_id", "group_id"),
            resource.TestCheckResourceAttr("data.fivetran_hybrid_deployment_agent.test_lpa", "registered_at", "registered_at"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientHybridDeploymentAgentDataSourceConfigMapping(t)
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
