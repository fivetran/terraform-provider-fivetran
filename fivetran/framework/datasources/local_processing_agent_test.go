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
    lpaDataSourceMockGetHandler *mock.Handler
    lpaDataSourceMockData map[string]interface{}
)

const (
    lpaMappingResponse = `
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

func setupMockClientLocalProcessingAgentDataSourceConfigMapping(t *testing.T) {
    tfmock.MockClient().Reset()

    lpaDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/local-processing-agents/lpa_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            lpaDataSourceMockData = tfmock.CreateMapFromJsonString(t, lpaMappingResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", lpaDataSourceMockData), nil
        },
    )
}

func TestDataSourceLocalProcessingAgentMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        data "fivetran_local_processing_agent" "test_lpa" {
            provider = fivetran-provider
            id = "lpa_id"
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, lpaDataSourceMockGetHandler.Interactions, 1)
                tfmock.AssertNotEmpty(t, lpaDataSourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agent.test_lpa", "display_name", "display_name"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agent.test_lpa", "group_id", "group_id"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agent.test_lpa", "registered_at", "registered_at"),
            /*resource.TestCheckResourceAttr("data.fivetran_local_processing_agent.test_lpa", "usage.0.connection_id", "connection_id1"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agent.test_lpa", "usage.0.schema", "schema1"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agent.test_lpa", "usage.0.service", "service1"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agent.test_lpa", "usage.1.connection_id", "connection_id2"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agent.test_lpa", "usage.1.schema", "schema2"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agent.test_lpa", "usage.1.service", "service2"),*/
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientLocalProcessingAgentDataSourceConfigMapping(t)
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
