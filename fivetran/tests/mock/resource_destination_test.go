package mock

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	destinationPostHandler   *mock.Handler
	destinationPatchHandler  *mock.Handler
	destinationTestHandler   *mock.Handler
	destinationDeleteHandler *mock.Handler
	testDestinationData      map[string]interface{}
)

func onPostDestination(t *testing.T, req *http.Request) (*http.Response, error) {
	assertEmpty(t, testDestinationData)

	body := requestBodyToJson(t, req)

	// Add response fields
	body["id"] = "destination_id"
	body["created_at"] = time.Now().Format("2006-01-02T15:04:05.000000Z")

	if port, ok := body["config"].(map[string]interface{})["port"]; ok {
		body["config"].(map[string]interface{})["port"] = strconv.Itoa(int(port.(float64)))
	}

	testDestinationData = body

	response := fivetranSuccessResponse(t, req, http.StatusCreated,
		"Destination has been created", body)

	return response, nil
}

func onPatchDestination(t *testing.T, req *http.Request) (*http.Response, error) {
	assertNotEmpty(t, testDestinationData)

	body := requestBodyToJson(t, req)

	if config, ok := body["config"]; ok {
		if port, ok := config.(map[string]interface{})["port"]; ok {
			body["config"].(map[string]interface{})["port"] = strconv.Itoa(int(port.(float64)))
		}
	}

	// Update saved values
	updateMapDeep(body, testDestinationData)

	response := fivetranSuccessResponse(t, req, http.StatusOK, "Destination has been updated", testDestinationData)

	return response, nil
}

func onTestDestination(t *testing.T, req *http.Request) (*http.Response, error) {
	// setup test results array
	setupTests := make([]interface{}, 0)

	setupTestResult := make(map[string]interface{})
	setupTestResult["title"] = "Test Title"
	setupTestResult["status"] = "PASSED"
	setupTestResult["message"] = "Test passed"

	setupTests = append(setupTests, setupTestResult)

	testDestinationData["setup_tests"] = setupTests

	response := fivetranSuccessResponse(t, req, http.StatusOK, "Setup tests have been completed", testDestinationData)
	return response, nil
}

func updateMapDeep(source map[string]interface{}, target map[string]interface{}) {
	for sk, sv := range source {
		if tv, ok := target[sk]; ok {
			if svmap, ok := sv.(map[string]interface{}); ok {
				if tvmap, ok := tv.(map[string]interface{}); ok {
					updateMapDeep(svmap, tvmap)
					continue
				}
			}
		}
		target[sk] = sv
	}
}

func setupMockClientForDestination(t *testing.T) {
	mockClient.Reset()
	testDestinationData = nil

	destinationPostHandler = mockClient.When(http.MethodPost, "/v1/destinations").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPostDestination(t, req)
		},
	)

	mockClient.When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			assertNotEmpty(t, testDestinationData)
			response := fivetranSuccessResponse(t, req, http.StatusOK, "", testDestinationData)
			return response, nil
		},
	)

	destinationPatchHandler = mockClient.When(http.MethodPatch, "/v1/destinations/destination_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPatchDestination(t, req)
		},
	)

	destinationTestHandler = mockClient.When(http.MethodPost, "/v1/destinations/destination_id/test").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onTestDestination(t, req)
		},
	)

	destinationDeleteHandler = mockClient.When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			assertNotEmpty(t, testDestinationData)
			testDestinationData = nil
			response := fivetranSuccessResponse(t, req, 200,
				"Destination with id 'destionation_id' has been deleted", nil)
			return response, nil
		},
	)

}

func TestResourceDestinationMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_destination" "mydestination" {
				provider = fivetran-provider

				group_id = "test_group_id"
				service = "postgres_rds_warehouse"
				time_zone_offset = "0"
				region = "GCP_US_EAST4"
				trust_certificates = "true"
				trust_fingerprints = "true"
				run_setup_tests = "false"

				config {
					host = "terraform-test.us-east-1.rds.amazonaws.com"
					port = 5432
					user = "postgres"
					password = "password"
					database = "fivetran"
					connection_type = "Directly"
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, destinationPostHandler.Interactions, 1)
				assertNotEmpty(t, testDestinationData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "service", "postgres_rds_warehouse"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "time_zone_offset", "0"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "region", "GCP_US_EAST4"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "trust_certificates", "true"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "trust_fingerprints", "true"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "run_setup_tests", "false"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.host", "terraform-test.us-east-1.rds.amazonaws.com"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.port", "5432"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.user", "postgres"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.password", "password"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.database", "fivetran"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.connection_type", "Directly"),
		),
	}

	step2 := resource.TestStep{
		Config: `
			resource "fivetran_destination" "mydestination" {
				provider = fivetran-provider

				group_id = "test_group_id"
				service = "postgres_rds_warehouse"
				time_zone_offset = "0"
				region = "GCP_US_EAST4"
				trust_certificates = "true"
				trust_fingerprints = "true"
				run_setup_tests = "false"

				config {
					host = "test.host"
					port = 5434
					user = "postgres"
					password = "password123"
					database = "fivetran"
					connection_type = "Directly"
				}
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, destinationPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "time_zone_offset", "0"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "region", "GCP_US_EAST4"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "trust_certificates", "true"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "trust_fingerprints", "true"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "run_setup_tests", "false"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.host", "test.host"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.port", "5434"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.user", "postgres"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.password", "password123"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.database", "fivetran"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.0.connection_type", "Directly"),
		),
	}

	step3 := resource.TestStep{
		Config: `
			resource "fivetran_destination" "mydestination" {
				provider = fivetran-provider

				group_id = "test_group_id"
				service = "postgres_rds_warehouse"
				time_zone_offset = "0"
				region = "GCP_US_EAST4"
				trust_certificates = "true"
				trust_fingerprints = "true"
				run_setup_tests = "true"

				config {
					host = "test.host"
					port = 5434
					user = "postgres"
					password = "password123"
					database = "fivetran"
					connection_type = "Directly"
				}
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, destinationPatchHandler.Interactions, 1)
				assertEqual(t, destinationTestHandler.Interactions, 1)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientForDestination(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, destinationDeleteHandler.Interactions, 1)
				assertEmpty(t, testDestinationData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
			},
		},
	)
}
