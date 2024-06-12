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
	localProcessingAgentPostHandler   *mock.Handler
	localProcessingAgentDeleteHandler *mock.Handler
	localProcessingAgentData map[string]interface{}
)

func setupMockClientLocalProcessingAgentResource(t *testing.T) {
	tfmock.MockClient().Reset()
	localProcessingAgentResponse :=
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

	localProcessingAgentPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/local-processing-agents").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			localProcessingAgentData = tfmock.CreateMapFromJsonString(t, localProcessingAgentResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Local Processing Agent has been created", localProcessingAgentData), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/local-processing-agents/lpa_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", localProcessingAgentData), nil
		},
	)

	localProcessingAgentDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/local-processing-agents/lpa_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "Local Processing Agent has been deleted", nil), nil
		},
	)
}

func TestResourceLocalProcessingAgentMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_local_processing_agent" "test_lpa" {
                 provider = fivetran-provider

                 display_name = "display_name"
                 group_id = "group_id"
            }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, localProcessingAgentPostHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_local_processing_agent.test_lpa", "display_name", "display_name"),
			resource.TestCheckResourceAttr("fivetran_local_processing_agent.test_lpa", "group_id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_local_processing_agent.test_lpa", "registered_at", "registered_at"),
			resource.TestCheckResourceAttr("fivetran_local_processing_agent.test_lpa", "config_json", "config_json"),
			resource.TestCheckResourceAttr("fivetran_local_processing_agent.test_lpa", "auth_json", "auth_json"),
			resource.TestCheckResourceAttr("fivetran_local_processing_agent.test_lpa", "docker_compose_yaml", "docker_compose_yaml"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientLocalProcessingAgentResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, localProcessingAgentDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
