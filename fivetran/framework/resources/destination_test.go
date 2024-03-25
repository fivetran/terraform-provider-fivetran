package resources_test

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceDestinationMappingMock(t *testing.T) {
	var testDestinationData map[string]interface{}
	var destinationMappingGetHandler *mock.Handler
	var destinationMappingPostHandler *mock.Handler
	var destinationMappingDeleteHandler *mock.Handler
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_destination" "mydestination" {
				provider = fivetran-provider

				group_id = "group_id"
				service = "snowflake"
				time_zone_offset = "0"
				region = "GCP_US_EAST4"
				trust_certificates = "true"
				trust_fingerprints = "true"
				daylight_saving_time_enabled = "true"
				run_setup_tests = "false"

				config {
					host = "host"
					port = "123"
					database = "database"
					auth = "auth"
					user = "user"
					password = "password"
					connection_type = "connection_type"
					tunnel_host = "tunnel_host"
					tunnel_port = "123"
					tunnel_user = "tunnel_user"
					project_id = "project_id"
					data_set_location = "data_set_location"
					bucket = "bucket"
					server_host_name = "server_host_name"
					http_path = "http_path"
					personal_access_token = "personal_access_token"
					create_external_tables = "false"
					external_location = "external_location"
					auth_type = "auth_type"
					role_arn = "role_arn"
					secret_key = "secret_key"
					private_key = "private_key"
					cluster_id = "cluster_id"
					cluster_region = "cluster_region"
					role = "role"
					is_private_key_encrypted = "false"
					passphrase = "passphrase"
					catalog = "catalog"
					fivetran_role_arn = "fivetran_role_arn"
					prefix_path = "prefix_path"
					region = "region"
					storage_account_name = "storage_account_name"
					container_name = "container_name"
					tenant_id = "tenant_id"
					client_id = "client_id"
					secret_value = "secret_value"
					workspace_name = "workspace_name"
					lakehouse_name = "lakehouse_name"
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationMappingGetHandler.Interactions, 0)
				tfmock.AssertEqual(t, destinationMappingPostHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, testDestinationData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				destinationMappingGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testDestinationData), nil
					},
				)

				destinationMappingPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
					func(req *http.Request) (*http.Response, error) {

						body := tfmock.RequestBodyToJson(t, req)

						tfmock.AssertKeyExists(t, body, "config")

						config := body["config"].(map[string]interface{})

						tfmock.AssertKeyExistsAndHasValue(t, config, "host", "host")

						tfmock.AssertKeyExistsAndHasValue(t, config, "port", float64(123))
						tfmock.AssertKeyExistsAndHasValue(t, config, "database", "database")
						tfmock.AssertKeyExistsAndHasValue(t, config, "auth", "auth")
						tfmock.AssertKeyExistsAndHasValue(t, config, "user", "user")
						tfmock.AssertKeyExistsAndHasValue(t, config, "password", "password")
						tfmock.AssertKeyExistsAndHasValue(t, config, "connection_type", "connection_type")
						tfmock.AssertKeyExistsAndHasValue(t, config, "tunnel_host", "tunnel_host")
						tfmock.AssertKeyExistsAndHasValue(t, config, "tunnel_user", "tunnel_user")
						tfmock.AssertKeyExistsAndHasValue(t, config, "project_id", "project_id")
						tfmock.AssertKeyExistsAndHasValue(t, config, "data_set_location", "data_set_location")
						tfmock.AssertKeyExistsAndHasValue(t, config, "bucket", "bucket")
						tfmock.AssertKeyExistsAndHasValue(t, config, "server_host_name", "server_host_name")
						tfmock.AssertKeyExistsAndHasValue(t, config, "http_path", "http_path")
						tfmock.AssertKeyExistsAndHasValue(t, config, "personal_access_token", "personal_access_token")
						tfmock.AssertKeyExistsAndHasValue(t, config, "create_external_tables", false)
						tfmock.AssertKeyExistsAndHasValue(t, config, "external_location", "external_location")
						tfmock.AssertKeyExistsAndHasValue(t, config, "auth_type", "auth_type")
						tfmock.AssertKeyExistsAndHasValue(t, config, "role_arn", "role_arn")
						tfmock.AssertKeyExistsAndHasValue(t, config, "secret_key", "secret_key")
						tfmock.AssertKeyExistsAndHasValue(t, config, "private_key", "private_key")
						tfmock.AssertKeyExistsAndHasValue(t, config, "cluster_id", "cluster_id")
						tfmock.AssertKeyExistsAndHasValue(t, config, "cluster_region", "cluster_region")
						tfmock.AssertKeyExistsAndHasValue(t, config, "role", "role")
						tfmock.AssertKeyExistsAndHasValue(t, config, "is_private_key_encrypted", false)
						tfmock.AssertKeyExistsAndHasValue(t, config, "passphrase", "passphrase")
						tfmock.AssertKeyExistsAndHasValue(t, config, "catalog", "catalog")
						tfmock.AssertKeyExistsAndHasValue(t, config, "fivetran_role_arn", "fivetran_role_arn")
						tfmock.AssertKeyExistsAndHasValue(t, config, "prefix_path", "prefix_path")
						tfmock.AssertKeyExistsAndHasValue(t, config, "region", "region")
						tfmock.AssertKeyExistsAndHasValue(t, config, "storage_account_name", "storage_account_name")
						tfmock.AssertKeyExistsAndHasValue(t, config, "container_name", "container_name")
						tfmock.AssertKeyExistsAndHasValue(t, config, "tenant_id", "tenant_id")
						tfmock.AssertKeyExistsAndHasValue(t, config, "client_id", "client_id")
						tfmock.AssertKeyExistsAndHasValue(t, config, "secret_value", "secret_value")
						tfmock.AssertKeyExistsAndHasValue(t, config, "workspace_name", "workspace_name")
						tfmock.AssertKeyExistsAndHasValue(t, config, "lakehouse_name", "lakehouse_name")

						tfmock.AssertKeyExistsAndHasValue(t, config, "tunnel_port", float64(123))

						testDestinationData = tfmock.CreateMapFromJsonString(t, `
						{
							"id":"destination_id",
							"group_id":"group_id",
							"service":"snowflake",
							"region":"GCP_US_EAST4",
							"time_zone_offset":"0",
							"setup_status":"connected",
							"daylight_saving_time_enabled":true,
							"setup_tests":[
								{
									"title":"Host Connection",
									"status":"PASSED",
									"message":""
								},
								{
									"title":"Database Connection",
									"status":"PASSED",
									"message":""
								},
								{
									"title":"Permission Test",
									"status":"PASSED",
									"message":""
								}
							],
							"config":{
								"host":                     "host",
								"port":                     "123",
								"database":                 "database",
								"auth":                     "auth",
								"user":                     "user",
								"password":                 "******",
								"connection_type":          "connection_type",
								"tunnel_host":              "tunnel_host",
								"tunnel_port":              "123",
								"tunnel_user":              "tunnel_user",
								"project_id":               "project_id",
								"data_set_location":        "data_set_location",
								"bucket":                   "bucket",
								"server_host_name":         "server_host_name",
								"http_path":                "http_path",
								"personal_access_token":    "******",
								"create_external_tables":   "false",
								"external_location":        "external_location",
								"auth_type":                "auth_type",
								"role_arn":                 "******",
								"secret_key":               "******",
								"private_key":              "******",
								"public_key":               "public_key",
								"cluster_id":               "cluster_id",
								"cluster_region":           "cluster_region",
								"role":                     "role",
								"is_private_key_encrypted": "false",
								"passphrase":               "******",
								"catalog": 					"catalog",
								"fivetran_role_arn": 		"fivetran_role_arn",
								"prefix_path": 				"prefix_path",
								"region": 					"region",
								"storage_account_name": 	"storage_account_name",
								"container_name": 			"container_name",
								"tenant_id": 				"tenant_id",
								"client_id": 				"client_id",
								"secret_value":				"******",
								"workspace_name": 			"workspace_name",
								"lakehouse_name": 			"lakehouse_name"
							}
						}
						`)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", testDestinationData), nil
					},
				)

				destinationMappingDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testDestinationData = nil
						response := tfmock.FivetranSuccessResponse(t, req, 200,
							"Destination with id 'destionation_id' has been deleted", nil)
						return response, nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationMappingDeleteHandler.Interactions, 1)
				tfmock.AssertEmpty(t, testDestinationData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestResourceDestinationSetupTests(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_destination" "mydestination" {
				provider = fivetran-provider

				group_id = "group_id"
				service = "snowflake"
				time_zone_offset = "0"
				region = "GCP_US_EAST4"
				trust_certificates = "true"
				trust_fingerprints = "true"
				run_setup_tests = "true"
				daylight_saving_time_enabled = "true"

				config {
					host = "terraform-test.us-east-1.rds.amazonaws.com"
					port = 5432
					user = "postgres"
					password = "password"
					database = "fivetran"
					connection_type = "Directly"
				}
			}`,
	}
	testDestinationData := tfmock.CreateMapFromJsonString(t, `
	{
		"id":"destination_id",
		"group_id":"group_id",
		"service":"snowflake",
		"region":"GCP_US_EAST4",
		"time_zone_offset":"0",
		"daylight_saving_time_enabled":true,
		"setup_status":"incomplete",
		"setup_tests":[
			{
				"title":"Host Connection",
				"status":"FAILED",
				"message":"Host Connection error"
			},
			{
				"title":"Database Connection",
				"status":"FAILED",
				"message":"Database Connection error"
			},
			{
				"title":"Permission Test",
				"status":"FAILED",
				"message":"Permission Test error"
			}
		],
		"config":{
			"host": "terraform-test.us-east-1.rds.amazonaws.com",
			"port": "5432",
			"user": "postgres",
			"password": "password",
			"database": "fivetran",
			"connection_type": "Directly"
		}
	}
	`)

	var getHandler *mock.Handler
	var postHandler *mock.Handler
	var testHandler *mock.Handler
	var deleteHandler *mock.Handler

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				postHandler = tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated,
							"Destination has been created", testDestinationData)
						return response, nil
					},
				)

				getHandler = tfmock.MockClient().When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", testDestinationData)
						return response, nil
					},
				)

				testHandler = tfmock.MockClient().When(http.MethodPost, "/v1/destinations/destination_id/test").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Setup tests have been completed", testDestinationData)
						return response, nil
					},
				)

				deleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testDestinationData = nil
						response := tfmock.FivetranSuccessResponse(t, req, 200,
							"Destination with id 'destionation_id' has been deleted", nil)
						return response, nil
					},
				)

			},
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, postHandler.Interactions, 1)
				tfmock.AssertEqual(t, testHandler.Interactions, 1)
				tfmock.AssertEqual(t, getHandler.Interactions, 2)
				tfmock.AssertEqual(t, deleteHandler.Interactions, 1)
				return nil
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestResourceDestinationMock(t *testing.T) {
	var destinationPostHandler *mock.Handler
	var destinationPatchHandler *mock.Handler
	var destinationTestHandler *mock.Handler
	var destinationDeleteHandler *mock.Handler
	var testDestinationData map[string]interface{}

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
				daylight_saving_time_enabled = "true"
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
				tfmock.AssertEqual(t, destinationPostHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, testDestinationData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "service", "postgres_rds_warehouse"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "time_zone_offset", "0"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "region", "GCP_US_EAST4"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "trust_certificates", "true"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "trust_fingerprints", "true"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "daylight_saving_time_enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "run_setup_tests", "false"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.host", "terraform-test.us-east-1.rds.amazonaws.com"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.port", "5432"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.user", "postgres"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.password", "password"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.database", "fivetran"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.connection_type", "Directly"),
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
				daylight_saving_time_enabled = "true"
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
				tfmock.AssertEqual(t, destinationPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "time_zone_offset", "0"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "region", "GCP_US_EAST4"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "trust_certificates", "true"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "trust_fingerprints", "true"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "daylight_saving_time_enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "run_setup_tests", "false"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.host", "test.host"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.port", "5434"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.user", "postgres"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.password", "password123"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.database", "fivetran"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.connection_type", "Directly"),
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
				daylight_saving_time_enabled = "true"
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
				tfmock.AssertEqual(t, destinationPatchHandler.Interactions, 1)
				tfmock.AssertEqual(t, destinationTestHandler.Interactions, 1)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {

				onPostDestination := func(t *testing.T, req *http.Request) (*http.Response, error) {
					tfmock.AssertEmpty(t, testDestinationData)

					body := tfmock.RequestBodyToJson(t, req)

					// Add response fields
					body["id"] = "destination_id"
					body["created_at"] = time.Now().Format("2006-01-02T15:04:05.000000Z")

					if port, ok := body["config"].(map[string]interface{})["port"]; ok {
						body["config"].(map[string]interface{})["port"] = strconv.Itoa(int(port.(float64)))
					}

					testDestinationData = body

					response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated,
						"Destination has been created", body)

					return response, nil
				}

				onPatchDestination := func(t *testing.T, req *http.Request) (*http.Response, error) {
					tfmock.AssertNotEmpty(t, testDestinationData)

					body := tfmock.RequestBodyToJson(t, req)

					if config, ok := body["config"]; ok {
						if port, ok := config.(map[string]interface{})["port"]; ok {
							body["config"].(map[string]interface{})["port"] = strconv.Itoa(int(port.(float64)))
						}
					}

					// Update saved values
					tfmock.UpdateMapDeep(body, testDestinationData)

					response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Destination has been updated", testDestinationData)

					return response, nil
				}

				onTestDestination := func(t *testing.T, req *http.Request) (*http.Response, error) {
					// setup test results array
					setupTests := make([]interface{}, 0)

					setupTestResult := make(map[string]interface{})
					setupTestResult["title"] = "Test Title"
					setupTestResult["status"] = "PASSED"
					setupTestResult["message"] = "Test passed"

					setupTests = append(setupTests, setupTestResult)

					testDestinationData["setup_tests"] = setupTests

					response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Setup tests have been completed", testDestinationData)
					return response, nil
				}

				tfmock.MockClient().Reset()
				testDestinationData = nil

				destinationPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return onPostDestination(t, req)
					},
				)

				tfmock.MockClient().When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						tfmock.AssertNotEmpty(t, testDestinationData)
						response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", testDestinationData)
						return response, nil
					},
				)

				destinationPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return onPatchDestination(t, req)
					},
				)

				destinationTestHandler = tfmock.MockClient().When(http.MethodPost, "/v1/destinations/destination_id/test").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return onTestDestination(t, req)
					},
				)

				destinationDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						tfmock.AssertNotEmpty(t, testDestinationData)
						testDestinationData = nil
						response := tfmock.FivetranSuccessResponse(t, req, 200,
							"Destination with id 'destionation_id' has been deleted", nil)
						return response, nil
					},
				)

			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationDeleteHandler.Interactions, 1)
				tfmock.AssertEmpty(t, testDestinationData)
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
