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

func TestReadonlyFieldSetMock(t *testing.T) {
	var testDestinationData map[string]interface{}
	var destinationMappingDeleteHandler *mock.Handler
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_destination" "mydestination" {
				provider = fivetran-provider
				group_id = "group_id"
				service = "new_s3_datalake"
				time_zone_offset = "0"
				region = "GCP_US_EAST4"
				run_setup_tests = "false"

				config {
					bucket = "bucket"
        			fivetran_role_arn = "fivetran_role_arn"
        			region = "region"
				}
			}`,
	}

	step2 := resource.TestStep{
		Config: `
			resource "fivetran_destination" "mydestination" {
				provider = fivetran-provider
				group_id = "group_id"
				service = "new_s3_datalake"
				time_zone_offset = "0"
				region = "GCP_US_EAST4"
				run_setup_tests = "false"

				config {
					bucket = "bucket1"
        			fivetran_role_arn = "fivetran_role_arn"
        			region = "region"
				}
			}`,
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				tfmock.MockClient().When(http.MethodGet, "/v1/destinations/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testDestinationData), nil
					},
				)

				tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testDestinationData = tfmock.CreateMapFromJsonString(t, `
						{
							"id":"group_id",
							"group_id":"group_id",
							"service":"new_s3_datalake",
							"region":"GCP_US_EAST4",
							"time_zone_offset":"0",
							"setup_status":"connected",
							"daylight_saving_time_enabled":true,
							"networking_method":"Directly",
							"config":{
								"external_id": "group_id",
								"bucket": "bucket",
								"fivetran_role_arn": "fivetran_role_arn",
								"region": "region"
							}
						}
						`)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", testDestinationData), nil
					},
				)

				tfmock.MockClient().When(http.MethodPatch, "/v1/destinations/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)

						tfmock.AssertKeyExists(t, body, "config")

						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "bucket", "bucket1")

						tfmock.AssertKeyDoesNotExist(t, config, "external_id")

						testDestinationData = tfmock.CreateMapFromJsonString(t, `
						{
							"id":"group_id",
							"group_id":"group_id",
							"service":"new_s3_datalake",
							"region":"GCP_US_EAST4",
							"time_zone_offset":"0",
							"setup_status":"connected",
							"daylight_saving_time_enabled":true,
							"networking_method":"Directly",
							"config":{
								"external_id": "",
								"bucket": "bucket1",
								"fivetran_role_arn": "fivetran_role_arn",
								"region": "region"
							}
						}
						`)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testDestinationData), nil
					},
				)

				destinationMappingDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/group_id").ThenCall(
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
				step2,
			},
		},
	)
}

func TestResourceMDLSDestinationMock(t *testing.T){ 
	var getDestinationResponse map[string]interface{}
	var postDestinationResponse map[string]interface{}
	var testDestinationData map[string]interface{}
	var destinationGetHandler *mock.Handler
	var destinationPostHandler *mock.Handler
	var destinationDeleteHandler *mock.Handler
	var testHandler *mock.Handler

	postDestinationResponse = tfmock.CreateMapFromJsonString(t, `
	{
		"id": "group_id",
		"group_id": "group_id",
		"service": "managed_data_lake",
		"region": "AWS_US_EAST_1",
		"time_zone_offset": "0",
		"setup_status": "incomplete",
		"daylight_saving_time_enabled": true,
		"private_link_id": null,
		"networking_method": "Directly",
		"proxy_agent_id": null,
		"setup_tests": [
		{
			"title": "AWS Read and write access test",
			"status": "FAILED",
			"message": "User: arn:aws:iam::1234567890:user/gcp_donkey is not authorized to perform: sts:AssumeRole on resource: arn:aws:iam::1234567890:role/smth-us-east-1 (Service: AWSSecurityTokenService; Status Code: 403; Error Code: AccessDenied; Request ID: 0000000-0000-0000-0000-000000000; Proxy: null)"
		},
		{
			"title": "Input fields validation test",
			"status": "PASSED",
			"message": ""
		}
		],
		"config": {
			"storage_provider": "AWS",
			"bucket": "smth-us-east-1-smth",
			"fivetran_role_arn": "arn:aws:iam::1234567890:role/smth-us-east-1",
			"prefix_path": "prefix-path",
			"region": "us-east-1",
			"snapshot_retention_period": "ONE_WEEK",
			"should_maintain_tables_in_databricks": false,
			"port": 443,
			"auth_type": "OAUTH2",
			"databricks_connection_type": "DIRECTLY",
			"should_maintain_tables_in_one_lake": false,
			"connection_type": "PRIVATE_LINK",
			"should_maintain_tables_in_glue": false,
			"should_maintain_tables_in_bqms": false
		}
	}
	`)

	getDestinationResponse = tfmock.CreateMapFromJsonString(t, `
	{
		"id": "group_id",
		"group_id": "group_id",
		"service": "managed_data_lake",
		"region": "AWS_US_EAST_1",
		"time_zone_offset": "0",
		"setup_status": "connected",
		"daylight_saving_time_enabled": true,
		"private_link_id": null,
		"proxy_agent_id": null,
		"networking_method": "Directly",
		"proxy_agent_id": null,
		"config": {
			"storage_provider": "AWS",
			"bucket": "smth-us-east-1-smth",
			"fivetran_role_arn": "arn:aws:iam::1234567890:role/smth-us-east-1",
			"prefix_path": "prefix-path",
			"region": "us-east-1",
			"snapshot_retention_period": "ONE_WEEK",
			"should_maintain_tables_in_databricks": false,
			"port": 443,
			"auth_type": "OAUTH2",
			"databricks_connection_type": "DIRECTLY",
			"should_maintain_tables_in_one_lake": false,
			"connection_type": "PRIVATE_LINK",
			"should_maintain_tables_in_glue": false,
			"should_maintain_tables_in_bqms": false,
			"polaris_catalog_configuration": {
				"polarisServerEndpoint": "https://smth.us-east-1.aws.polaris.fivetran.com/api/catalog",
				"polarisCatalog": "group_id",
				"clientId": "abc1234567890",
				"clientSecret": "********"
			}
		}
	}
	`)

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				destinationPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
					func(req *http.Request) (*http.Response, error) {

						testDestinationData = postDestinationResponse 
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", postDestinationResponse), nil
					},
				)

				testHandler = tfmock.MockClient().When(http.MethodPost, "/v1/destinations/group_id/test").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Setup tests have been completed", testDestinationData)
						return response, nil
					},
				)

				destinationGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/destinations/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", getDestinationResponse), nil
					},
				)

				destinationDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testDestinationData = nil
						response := tfmock.FivetranSuccessResponse(t, req, 200,
							"Destination with id 'group_id' has been deleted", nil)
						return response, nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationDeleteHandler.Interactions, 2)
				return nil
			},
			Steps: []resource.TestStep{
				// create TF resource
				{
					Config: `
					resource "fivetran_destination" "mydestination" {
						provider = fivetran-provider
						
						group_id = "group_id"
						service= "managed_data_lake"
						region= "AWS_US_EAST_1"
						time_zone_offset= "0"
						trust_certificates = "true"
						trust_fingerprints = "true"
						run_setup_tests = "true"
						daylight_saving_time_enabled = "true"
						networking_method= "Directly"
						config {
							#storage_provider= "AWS"
							bucket= "smth-us-east-1-smth"
							fivetran_role_arn= "arn:aws:iam::1234567890:role/smth-us-east-1"
							prefix_path= "prefix-path"
							region= "us-east-1"
							#snapshot_retention_period= "ONE_WEEK"
							#port= 443
							#auth_type= "OAUTH2"
							#databricks_connection_type= "DIRECTLY"
							connection_type= "PrivateLink"
						}
					}`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, destinationGetHandler.Interactions, 1)
							tfmock.AssertEqual(t, destinationPostHandler.Interactions, 1)
							tfmock.AssertEqual(t, testHandler.Interactions, 1)
							tfmock.AssertNotEmpty(t, getDestinationResponse)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "id", "group_id"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "group_id", "group_id"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "service", "managed_data_lake"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "region", "AWS_US_EAST_1"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "time_zone_offset", "0"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "daylight_saving_time_enabled", "true"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "trust_certificates", "true"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "trust_fingerprints", "true"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "networking_method", "Directly"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.bucket", "smth-us-east-1-smth"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.fivetran_role_arn", "arn:aws:iam::1234567890:role/smth-us-east-1"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.prefix_path", "prefix-path"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.region", "us-east-1"),
						resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.connection_type", "PrivateLink"),
					),
				},

				// import created resource
				{
					Config: `
						resource "fivetran_destination" "mydestination"  {
							provider = fivetran-provider
						}`,
					ImportState:            true,
					ResourceName:            "fivetran_destination.mydestination",
					ImportStateId:  "group_id",
					
					ImportStateCheck: tfmock.ComposeImportStateCheck(
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "id", "group_id"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "group_id", "group_id"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "service", "managed_data_lake"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "region", "AWS_US_EAST_1"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "time_zone_offset", "0"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "daylight_saving_time_enabled", "true"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "networking_method", "Directly"),

						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.bucket", "smth-us-east-1-smth"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.fivetran_role_arn", "arn:aws:iam::1234567890:role/smth-us-east-1"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.prefix_path", "prefix-path"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.region", "us-east-1"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.storage_provider", "AWS"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.snapshot_retention_period", "ONE_WEEK"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.port", "443"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.auth_type", "OAUTH2"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.databricks_connection_type", "DIRECTLY"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.connection_type", "PRIVATE_LINK"),
					),
				},

				// remove from TF config to test import verify in the next test step
				{
					Config: `
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, destinationDeleteHandler.Interactions, 1)
							return nil
						},
					),
				},

				// import and persist state
				{
					Config: `
					resource "fivetran_destination" "mydestination" {
						provider = fivetran-provider
						
						group_id = "group_id"
						service= "managed_data_lake"
						region= "AWS_US_EAST_1"
						time_zone_offset= "0"
						trust_certificates = "true"
						trust_fingerprints = "true"
						run_setup_tests = "true"
						daylight_saving_time_enabled = "true"
						networking_method= "Directly"
						config {
							bucket= "smth-us-east-1-smth"
							fivetran_role_arn= "arn:aws:iam::1234567890:role/smth-us-east-1"
							prefix_path= "prefix-path"
							region= "us-east-1"
							storage_provider= "AWS"
							snapshot_retention_period= "ONE_WEEK"
							port= 443
							auth_type= "OAUTH2"
							databricks_connection_type= "DIRECTLY"
							#connection_type= "PRIVATE_LINK"
						}
					}`,
					ImportState:            true,
					ImportStatePersist: 	true,
					ResourceName:            "fivetran_destination.mydestination",
					ImportStateId:  "group_id",
					
					ImportStateCheck: tfmock.ComposeImportStateCheck(
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "id", "group_id"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "group_id", "group_id"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "service", "managed_data_lake"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "region", "AWS_US_EAST_1"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "time_zone_offset", "0"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "daylight_saving_time_enabled", "true"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "networking_method", "Directly"),

						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.bucket", "smth-us-east-1-smth"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.fivetran_role_arn", "arn:aws:iam::1234567890:role/smth-us-east-1"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.prefix_path", "prefix-path"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.region", "us-east-1"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.storage_provider", "AWS"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.snapshot_retention_period", "ONE_WEEK"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.port", "443"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.auth_type", "OAUTH2"),
						tfmock.CheckImportResourceAttr("fivetran_destination", "group_id", "config.databricks_connection_type", "DIRECTLY"),
					),
				},			
			},
		},
	)
}

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
							"networking_method":"Directly",
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
				networking_method = "Directly"

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
		"networking_method":"Directly",
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
				networking_method = "Directly"

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
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "group_id", "test_group_id"),
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
				networking_method = "Directly"

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
				networking_method = "Directly"

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

func TestResourceDestinationPrivateLinkChangeMock(t *testing.T) {
	var destinationPostHandler *mock.Handler
	var destinationPatchHandler *mock.Handler
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
				networking_method = "Directly"
				hybrid_deployment_agent_id = "agent_id_1"
				private_link_id = "private_link_id_1"
				proxy_agent_id = "proxy_agent_id_1"

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
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "hybrid_deployment_agent_id", "agent_id_1"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "private_link_id", "private_link_id_1"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "proxy_agent_id", "proxy_agent_id_1"),
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
				networking_method = "Directly"
				hybrid_deployment_agent_id = "agent_id_2"
				private_link_id = "private_link_id_2"
				proxy_agent_id = "proxy_agent_id_2"

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
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "hybrid_deployment_agent_id", "agent_id_2"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "private_link_id", "private_link_id_2"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "proxy_agent_id", "proxy_agent_id_2"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.host", "test.host"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.port", "5434"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.user", "postgres"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.password", "password123"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.database", "fivetran"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.connection_type", "Directly"),
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

					testDestinationData = body

					response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated,
						"Destination has been created", body)

					return response, nil
				}

				onPatchDestination := func(t *testing.T, req *http.Request) (*http.Response, error) {
					tfmock.AssertNotEmpty(t, testDestinationData)

					body := tfmock.RequestBodyToJson(t, req)

					// Update saved values
					tfmock.UpdateMapDeep(body, testDestinationData)

					response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Destination has been updated", testDestinationData)

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
			},
		},
	)
}

func TestResourceDestinationDatabricksPrivateLinkConfigPreservationMock(t *testing.T) {
	var destinationPostHandler *mock.Handler
	var destinationDeleteHandler *mock.Handler
	var testDestinationData map[string]interface{}

	step1 := resource.TestStep{
		Config: `
			resource "fivetran_destination" "mydestination" {
				provider = fivetran-provider

				group_id = "test_group_id"
				service = "databricks"
				time_zone_offset = "0"
				region = "AZURE_EASTUS"
				trust_certificates = "true"
				trust_fingerprints = "true"
				daylight_saving_time_enabled = "true"
				run_setup_tests = "false"
				networking_method = "PrivateLink"
				private_link_id = "test_private_link_id"

				config {
					auth_type = "PERSONAL_ACCESS_TOKEN"
					catalog = "test_catalog"
					server_host_name = "adb-1234567890123.19.azuredatabricks.net"
					port = 443
					http_path = "/sql/1.0/warehouses/test"
					cloud_provider = "AZURE"
					personal_access_token = "test_token"
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationPostHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, testDestinationData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "service", "databricks"),
			// These top-level fields should also be preserved
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "networking_method", "PrivateLink"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "private_link_id", "test_private_link_id"),
			// Config fields should preserve the original values, not the API-modified ones
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.server_host_name", "adb-1234567890123.19.azuredatabricks.net"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.cloud_provider", "AZURE"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.catalog", "test_catalog"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.port", "443"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {

				onPostDestination := func(t *testing.T, req *http.Request) (*http.Response, error) {
					tfmock.AssertEmpty(t, testDestinationData)

					body := tfmock.RequestBodyToJson(t, req)

					// Verify the request has the correct values
					config := body["config"].(map[string]interface{})
					tfmock.AssertKeyExistsAndHasValue(t, config, "server_host_name", "adb-1234567890123.19.azuredatabricks.net")
					tfmock.AssertKeyExistsAndHasValue(t, config, "cloud_provider", "AZURE")

					// Add response fields
					body["id"] = "destination_id"
					body["created_at"] = time.Now().Format("2006-01-02T15:04:05.000000Z")
					body["setup_status"] = "connected"

					// Simulate PrivateLink behavior: API returns modified values
					// This mimics the real-world bug where Fivetran's API changes these fields
					config["server_host_name"] = "pls-prod-fivetran-eastus-pls-1.eastus.azure.fivetran.com"
					config["cloud_provider"] = "AWS"
					// API also changes top-level networking fields
					body["networking_method"] = "Directly"
					body["private_link_id"] = ""

					testDestinationData = body

					response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated,
						"Destination has been created", body)

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

				destinationDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						tfmock.AssertNotEmpty(t, testDestinationData)
						testDestinationData = nil
						response := tfmock.FivetranSuccessResponse(t, req, 200,
							"Destination with id 'destination_id' has been deleted", nil)
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
			},
		},
	)
}

func TestResourceDestinationSetupTestsPreservesNetworkingFieldsMock(t *testing.T) {
	var destinationPostHandler *mock.Handler
	var destinationTestHandler *mock.Handler
	var destinationDeleteHandler *mock.Handler
	var testDestinationData map[string]interface{}

	// Step 1: Create with run_setup_tests = false
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_destination" "mydestination" {
				provider = fivetran-provider

				group_id = "test_group_id"
				service = "databricks"
				time_zone_offset = "0"
				region = "AZURE_EASTUS"
				trust_certificates = "true"
				trust_fingerprints = "true"
				daylight_saving_time_enabled = "true"
				run_setup_tests = "false"
				networking_method = "PrivateLink"
				private_link_id = "test_private_link_id"

				config {
					auth_type = "PERSONAL_ACCESS_TOKEN"
					catalog = "test_catalog"
					server_host_name = "adb-test.azuredatabricks.net"
					port = 443
					http_path = "/sql/1.0/warehouses/test"
					cloud_provider = "AZURE"
					personal_access_token = "test_token"
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationPostHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, testDestinationData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "private_link_id", "test_private_link_id"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "networking_method", "PrivateLink"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "run_setup_tests", "false"),
		),
	}

	// Step 2: Change run_setup_tests to true - should preserve private_link_id
	step2 := resource.TestStep{
		Config: `
			resource "fivetran_destination" "mydestination" {
				provider = fivetran-provider

				group_id = "test_group_id"
				service = "databricks"
				time_zone_offset = "0"
				region = "AZURE_EASTUS"
				trust_certificates = "true"
				trust_fingerprints = "true"
				daylight_saving_time_enabled = "true"
				run_setup_tests = "true"
				networking_method = "PrivateLink"
				private_link_id = "test_private_link_id"

				config {
					auth_type = "PERSONAL_ACCESS_TOKEN"
					catalog = "test_catalog"
					server_host_name = "adb-test.azuredatabricks.net"
					port = 443
					http_path = "/sql/1.0/warehouses/test"
					cloud_provider = "AZURE"
					personal_access_token = "test_token"
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationTestHandler.Interactions, 1)
				return nil
			},
			// These should still be preserved even after running setup tests
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "private_link_id", "test_private_link_id"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "networking_method", "PrivateLink"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "run_setup_tests", "true"),
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
					body["setup_status"] = "connected"

					testDestinationData = body

					response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated,
						"Destination has been created", body)

					return response, nil
				}

				onTestDestination := func(t *testing.T, req *http.Request) (*http.Response, error) {
					// Legacy response that doesn't include private_link_id in the data
					legacyResponse := map[string]interface{}{
						"id":                           "destination_id",
						"group_id":                     "test_group_id",
						"service":                      "databricks",
						"region":                       "AZURE_EASTUS",
						"time_zone_offset":             "0",
						"setup_status":                 "connected",
						"daylight_saving_time_enabled": true,
						// Note: private_link_id and networking_method NOT included in legacy response
						"setup_tests": []interface{}{
							map[string]interface{}{
								"title":   "Test Title",
								"status":  "PASSED",
								"message": "Test passed",
							},
						},
					}

					response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK,
						"Setup tests have been completed", legacyResponse)
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
							"Destination with id 'destination_id' has been deleted", nil)
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
			},
		},
	)
}
