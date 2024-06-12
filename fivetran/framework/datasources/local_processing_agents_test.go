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
	localProcessingAgentsDataSourceMockGetHandler *mock.Handler
	localProcessingAgentsDataSourceMockData       map[string]interface{}
)

const (
	localProcessingAgentsMappingResponse = `
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

func setupMockClientLocalProcessingAgentsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	localProcessingAgentsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/local-processing-agents").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			localProcessingAgentsDataSourceMockData = tfmock.CreateMapFromJsonString(t, localProcessingAgentsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", localProcessingAgentsDataSourceMockData), nil
		},
	)
}

func TestDataSourceLocalProcessingAgentsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_local_processing_agents" "test_lpa" {
            provider = fivetran-provider
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, localProcessingAgentsDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, localProcessingAgentsDataSourceMockData)
				return nil
			},
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.0.display_name", "display_name1"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.0.group_id", "group_id1"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.0.registered_at", "registered_at1"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.0.usage.0.connection_id", "connection_id11"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.0.usage.0.schema", "schema11"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.0.usage.0.service", "service11"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.0.usage.1.connection_id", "connection_id12"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.0.usage.1.schema", "schema12"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.0.usage.1.service", "service12"),

            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.1.display_name", "display_name2"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.1.group_id", "group_id2"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.1.registered_at", "registered_at2"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.1.usage.0.connection_id", "connection_id21"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.1.usage.0.schema", "schema21"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.1.usage.0.service", "service21"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.1.usage.1.connection_id", "connection_id22"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.1.usage.1.schema", "schema22"),
            resource.TestCheckResourceAttr("data.fivetran_local_processing_agents.test_lpa", "items.1.usage.1.service", "service22"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientLocalProcessingAgentsDataSourceConfigMapping(t)
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
