package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceSchemaConfigUpdateTimeoutZeroMeansNoDeadlineMock(t *testing.T) {
	var (
		schemaGetHandler   *mock.Handler
		schemaPatchHandler *mock.Handler
		schemaResponseData map[string]interface{}
	)

	schemaAllowAll := `
	{
		"schema_change_handling": "ALLOW_ALL",
		"schemas": {}
	}`

	schemaBlockAll := `
	{
		"schema_change_handling": "BLOCK_ALL",
		"schemas": {}
	}`

	setupMockClient := func(t *testing.T) {
		mockClient.Reset()
		schemaResponseData = createMapFromJsonString(t, schemaAllowAll)

		schemaGetHandler = mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaResponseData), nil
			},
		)

		schemaPatchHandler = mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				_, hasDeadline := req.Context().Deadline()
				assertEqual(t, hasDeadline, false)
				assertEqual(t, req.Context().Err(), nil)

				body := requestBodyToJson(t, req)
				assertKeyExistsAndHasValue(t, body, "schema_change_handling", "BLOCK_ALL")

				schemaResponseData = createMapFromJsonString(t, schemaBlockAll)
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaResponseData), nil
			},
		)
	}

	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "ALLOW_ALL"

				timeouts {
					update = "0"
				}
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaPatchHandler.Interactions, 0)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_ALL"),
		),
	}

	step2 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider = fivetran-provider
				connector_id = "connector_id"
				schema_change_handling = "BLOCK_ALL"

				timeouts {
					update = "0"
				}
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, schemaGetHandler.Interactions > 0, true)
				assertEqual(t, schemaPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "BLOCK_ALL"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClient(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}
