package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	dbtModelsDataSourceMockGetHandler *mock.Handler
	dbtModelsDataSourceMockData       map[string]interface{}
)

const (
	dbtModelsMappingResponseWithCursor = `
	{
		"items":[
			{
				"id": "model_id",
				"model_name": "model_name",
				"scheduled": true
			}
		],
		"next_cursor": "next_cursor"	
    }
	`

	dbtModelsMappingResponse = `
	{
		"items":[
			{
				"id": "model_id_2",
				"model_name": "model_name_2",
				"scheduled": false
			}
		],
		"next_cursor": null	
    }
	`
)

func setupMockClientdbtModelsDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()
	dbtModelsDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/dbt/models").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			project_id := req.URL.Query().Get("project_id")
			assertEqual(t, project_id, "project_id")
			cursor := req.URL.Query().Get("cursor")
			if cursor == "" {
				dbtModelsDataSourceMockData = createMapFromJsonString(t, dbtModelsMappingResponseWithCursor)
			} else {
				assertEqual(t, cursor, "next_cursor")
				dbtModelsDataSourceMockData = createMapFromJsonString(t, dbtModelsMappingResponse)
			}
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", dbtModelsDataSourceMockData), nil
		},
	)
}

func TestDataSourceDbtModelsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_dbt_models" "test_models" {
			provider = fivetran-provider
			project_id = "project_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, dbtModelsDataSourceMockGetHandler.Interactions, 4)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_dbt_models.test_models", "models.0.id", "model_id"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_models.test_models", "models.0.model_name", "model_name"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_models.test_models", "models.0.scheduled", "true"),

			resource.TestCheckResourceAttr("data.fivetran_dbt_models.test_models", "models.1.id", "model_id_2"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_models.test_models", "models.1.model_name", "model_name_2"),
			resource.TestCheckResourceAttr("data.fivetran_dbt_models.test_models", "models.1.scheduled", "false"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientdbtModelsDataSourceConfigMapping(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
