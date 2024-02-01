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
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.host", "host"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.port", "123"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.database", "database"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.auth", "auth"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.user", "user"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.password", "******"),
			resource.TestCheckResourceAttr("data.fivetran_destination.test_destintion", "config.connection_type", "connection_type"),
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
