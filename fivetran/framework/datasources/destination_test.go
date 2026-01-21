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
	destinationDataSourceMockGetHandler *mock.Handler

	destinationDataSourceMockData map[string]interface{}
)

func setupMockClientDestinationDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	destinationDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			destinationDataSourceMockData = tfmock.CreateMapFromJsonString(t, `
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
					"connection_type":          "DIRECTLY",
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
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", destinationDataSourceMockData), nil
		},
	)
}

func TestDataSourceDestinationConfigMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_destination" "test_destintion" {
			provider = fivetran-provider
			id = "destination_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, destinationDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "service", "snowflake"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "group_id", "group_id"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "region", "GCP_US_EAST4"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "time_zone_offset", "0"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "setup_status", "connected"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "daylight_saving_time_enabled", "true"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.host", "host"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.port", "123"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.database", "database"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.auth", "auth"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.user", "user"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.password", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.connection_type", "DIRECTLY"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.tunnel_host", "tunnel_host"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.tunnel_port", "123"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.tunnel_user", "tunnel_user"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.project_id", "project_id"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.data_set_location", "data_set_location"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.bucket", "bucket"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.server_host_name", "server_host_name"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.http_path", "http_path"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.personal_access_token", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.create_external_tables", "false"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.external_location", "external_location"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.auth_type", "auth_type"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.role_arn", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.secret_key", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.private_key", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.public_key", "public_key"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.cluster_id", "cluster_id"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.cluster_region", "cluster_region"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.role", "role"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.is_private_key_encrypted", "false"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.passphrase", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.catalog", "catalog"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.fivetran_role_arn", "fivetran_role_arn"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.prefix_path", "prefix_path"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.region", "region"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.storage_account_name", "storage_account_name"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.container_name", "container_name"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.tenant_id", "tenant_id"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.client_id", "client_id"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.secret_value", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.workspace_name", "workspace_name"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.lakehouse_name", "lakehouse_name"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientDestinationDataSourceConfigMapping(t)
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
