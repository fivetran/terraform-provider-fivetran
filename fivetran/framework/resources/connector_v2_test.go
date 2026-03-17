package resources_test

import (
	"net/http"
	"regexp"
	"testing"

	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestResourceConnectorV2RetryTableMock verifies that when the first POST attempt fails
// (API rejects "table"), the provider retries with "table_group_name" and succeeds.
func TestResourceConnectorV2RetryTableMock(t *testing.T) {
	var responseData map[string]interface{}
	postCallCount := 0

	preCheck := func() {
		tfmock.MockClient().Reset()
		postCallCount = 0

		tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if responseData == nil {
					return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "NotFound", nil), nil
				}
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
			},
		)
		tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				responseData = nil
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)
		tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				postCallCount++
				body := tfmock.RequestBodyToJson(t, req)
				config, _ := body["config"].(map[string]interface{})

				if postCallCount == 1 {
					// First attempt must use "table"
					tfmock.AssertKeyExists(t, config, "table")
					tfmock.AssertKeyDoesNotExist(t, config, "table_group_name")
					return tfmock.FivetranSuccessResponse(t, req, http.StatusUnprocessableEntity, "Invalid table field", nil), nil
				}

				// Second attempt must use "table_group_name"
				tfmock.AssertKeyExists(t, config, "table_group_name")
				tfmock.AssertKeyDoesNotExist(t, config, "table")

				responseJson := createConnectorTestResponseJsonMock(
					"connector_id", "group_id", "postgres", "my_schema", "my_table",
					`{"user": "u"}`,
				)
				responseData = tfmock.CreateMapFromJsonString(t, responseJson)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", responseData), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 preCheck,
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_connector_v2" "test_connector" {
					provider = fivetran-provider

					group_id = "group_id"
					service  = "postgres"

					destination_schema = "my_schema.my_table"

					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests    = false

					config = {
						user = "u"
					}
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, postCallCount, 2)
						return nil
					},
					resource.TestCheckResourceAttr("fivetran_connector_v2.test_connector", "destination_schema", "my_schema.my_table"),
					resource.TestCheckResourceAttr("fivetran_connector_v2.test_connector", "service", "postgres"),
				),
			},
		},
	})
}

// TestResourceConnectorV2AllCandidatesFailMock verifies that when all POST candidates fail
// the provider surfaces an error containing "All destination_schema configurations failed".
func TestResourceConnectorV2AllCandidatesFailMock(t *testing.T) {
	postCallCount := 0

	preCheck := func() {
		tfmock.MockClient().Reset()
		postCallCount = 0

		tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				postCallCount++
				return tfmock.FivetranSuccessResponse(t, req, http.StatusUnprocessableEntity, "Invalid schema fields", nil), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 preCheck,
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_connector_v2" "test_connector" {
					provider = fivetran-provider

					group_id = "group_id"
					service  = "postgres"

					destination_schema = "my_schema.my_table"

					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests    = false

					config = {
						user = "u"
					}
				}`,
				ExpectError: regexp.MustCompile(`All destination_schema configurations failed`),
			},
		},
	})
}

// TestResourceConnectorV2PlainSchemaRetryMock verifies that for a plain schema value
// (no dot), "schema" is tried first and "schema_prefix" is tried on failure.
func TestResourceConnectorV2PlainSchemaRetryMock(t *testing.T) {
	var responseData map[string]interface{}
	postCallCount := 0

	preCheck := func() {
		tfmock.MockClient().Reset()
		postCallCount = 0

		tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if responseData == nil {
					return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "NotFound", nil), nil
				}
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
			},
		)
		tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				responseData = nil
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)
		tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				postCallCount++
				body := tfmock.RequestBodyToJson(t, req)
				config, _ := body["config"].(map[string]interface{})

				if postCallCount == 1 {
					tfmock.AssertKeyExists(t, config, "schema")
					tfmock.AssertKeyDoesNotExist(t, config, "schema_prefix")
					return tfmock.FivetranSuccessResponse(t, req, http.StatusUnprocessableEntity, "Invalid schema field", nil), nil
				}

				tfmock.AssertKeyExists(t, config, "schema_prefix")
				tfmock.AssertKeyDoesNotExist(t, config, "schema")

				responseJson := createConnectorTestResponseJsonMock(
					"connector_id", "group_id", "postgres", "my_schema", "",
					`{"user": "u"}`,
				)
				responseData = tfmock.CreateMapFromJsonString(t, responseJson)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", responseData), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 preCheck,
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_connector_v2" "test_connector" {
					provider = fivetran-provider

					group_id = "group_id"
					service  = "postgres"

					destination_schema = "my_schema"

					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests    = false

					config = {
						user = "u"
					}
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, postCallCount, 2)
						return nil
					},
					resource.TestCheckResourceAttr("fivetran_connector_v2.test_connector", "destination_schema", "my_schema"),
				),
			},
		},
	})
}

// TestResourceConnectorV2FirstCandidateSucceedsMock verifies that when the first POST
// candidate succeeds, there is no unnecessary second attempt.
func TestResourceConnectorV2FirstCandidateSucceedsMock(t *testing.T) {
	var responseData map[string]interface{}
	postCallCount := 0

	preCheck := func() {
		tfmock.MockClient().Reset()
		postCallCount = 0

		tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if responseData == nil {
					return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "NotFound", nil), nil
				}
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
			},
		)
		tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				responseData = nil
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)
		tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				postCallCount++
				body := tfmock.RequestBodyToJson(t, req)
				config, _ := body["config"].(map[string]interface{})

				// First attempt (schema + table) succeeds immediately.
				tfmock.AssertKeyExists(t, config, "table")
				tfmock.AssertKeyDoesNotExist(t, config, "table_group_name")

				responseJson := createConnectorTestResponseJsonMock(
					"connector_id", "group_id", "postgres", "my_schema", "my_table",
					`{"user": "u"}`,
				)
				responseData = tfmock.CreateMapFromJsonString(t, responseJson)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", responseData), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 preCheck,
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_connector_v2" "test_connector" {
					provider = fivetran-provider

					group_id = "group_id"
					service  = "postgres"

					destination_schema = "my_schema.my_table"

					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests    = false

					config = {
						user = "u"
					}
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						// Exactly one POST — no retry needed.
						tfmock.AssertEqual(t, postCallCount, 1)
						return nil
					},
					resource.TestCheckResourceAttr("fivetran_connector_v2.test_connector", "destination_schema", "my_schema.my_table"),
				),
			},
		},
	})
}

// TestResourceConnectorV2DestinationSchemaPersistsOnRefreshMock verifies that
// destination_schema retains the user's original value (e.g. "my_schema.my_table")
// after a Read/refresh, even though the API only returns "my_schema".
func TestResourceConnectorV2DestinationSchemaPersistsOnRefreshMock(t *testing.T) {
	var responseData map[string]interface{}

	preCheck := func() {
		tfmock.MockClient().Reset()

		tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if responseData == nil {
					return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "NotFound", nil), nil
				}
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
			},
		)
		tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				responseData = nil
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)
		tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				// API only returns the schema portion (no table) — simulates real API behavior.
				responseJson := createConnectorTestResponseJsonMock(
					"connector_id", "group_id", "postgres", "my_schema", "",
					`{"user": "u"}`,
				)
				responseData = tfmock.CreateMapFromJsonString(t, responseJson)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", responseData), nil
			},
		)
	}

	config := `
	resource "fivetran_connector_v2" "test_connector" {
		provider = fivetran-provider

		group_id = "group_id"
		service  = "postgres"

		destination_schema = "my_schema.my_table"

		trust_certificates = false
		trust_fingerprints = false
		run_setup_tests    = false

		config = {
			user = "u"
		}
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 preCheck,
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		CheckDestroy:             func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				// Create
				Config: config,
				Check: resource.TestCheckResourceAttr(
					"fivetran_connector_v2.test_connector", "destination_schema", "my_schema.my_table",
				),
			},
			{
				// Refresh — destination_schema must stay "my_schema.my_table", not drift to "my_schema"
				Config: config,
				Check: resource.TestCheckResourceAttr(
					"fivetran_connector_v2.test_connector", "destination_schema", "my_schema.my_table",
				),
			},
		},
	})
}
