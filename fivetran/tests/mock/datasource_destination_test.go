package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	destinationDataSourceMockGetHandler *mock.Handler

	destinationDataSourceMockData map[string]interface{}
)

func setupMockClientDestinationDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	destinationDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			destinationDataSourceMockData = createMapFromJsonString(t, destinationMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", destinationDataSourceMockData), nil
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
				assertEqual(t, destinationDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, destinationDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "service", "snowflake"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "group_id", "group_id"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "region", "GCP_US_EAST4"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "time_zone_offset", "0"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "setup_status", "connected"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.host", "host"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.port", "123"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.database", "database"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.auth", "auth"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.user", "user"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.password", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.connection_type", "connection_type"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.tunnel_host", "tunnel_host"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.tunnel_port", "123"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.tunnel_user", "tunnel_user"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.project_id", "project_id"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.data_set_location", "data_set_location"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.bucket", "bucket"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.server_host_name", "server_host_name"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.http_path", "http_path"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.personal_access_token", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.create_external_tables", "false"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.external_location", "external_location"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.auth_type", "auth_type"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.role_arn", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.secret_key", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.private_key", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.public_key", "public_key"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.cluster_id", "cluster_id"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.cluster_region", "cluster_region"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.role", "role"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.passphrase", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.catalog", "catalog"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.fivetran_role_arn", "fivetran_role_arn"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.prefix_path", "prefix_path"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.0.region", "region"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientDestinationDataSourceConfigMapping(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
