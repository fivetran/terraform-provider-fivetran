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
	userConnectionMembershipsDataSourceMockGetHandler *mock.Handler
	userConnectionMembershipsDataSourceMockData       map[string]interface{}
)

const (
	userConnectionMembershipsMappingResponse = `
    {
      "items": [
        {
          "id": "connection_id_1",
          "role": "Connection Administrator",
          "created_at": "2020-05-25T15:26:47.306509Z"
        },
        {
          "id": "connection_id_2",
          "role": "Connection Reviewer",
          "created_at": "2020-05-25T15:26:47.306509Z"
        }
      ],
      "next_cursor": null
    }`
)

func setupMockClientUserConnectionMembershipsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	userDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userDataSourceMockData = tfmock.CreateMapFromJsonString(t, userMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", userDataSourceMockData), nil
		},
	)

	userConnectionMembershipsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/users/user_id/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userConnectionMembershipsDataSourceMockData = tfmock.CreateMapFromJsonString(t, userConnectionMembershipsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", userConnectionMembershipsDataSourceMockData), nil
		},
	)
}

func TestDataSourceUserConnectionMembershipsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_user_connection_memberships" "test" {
            provider     = fivetran-provider
            id      		 = "user_id"
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, userConnectionMembershipsDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, userConnectionMembershipsDataSourceMockData)
				return nil
			},
      resource.TestCheckResourceAttr("data.fivetran_user_connection_memberships.test", "connections.0.connection_id", "connection_id_1"),
      resource.TestCheckResourceAttr("data.fivetran_user_connection_memberships.test", "connections.0.role", "Connection Administrator"),
      resource.TestCheckResourceAttr("data.fivetran_user_connection_memberships.test", "connections.0.created_at", "2020-05-25T15:26:47.306509Z"),
      resource.TestCheckResourceAttr("data.fivetran_user_connection_memberships.test", "connections.1.connection_id", "connection_id_2"),
			resource.TestCheckResourceAttr("data.fivetran_user_connection_memberships.test", "connections.1.role", "Connection Reviewer"),
      resource.TestCheckResourceAttr("data.fivetran_user_connection_memberships.test", "connections.1.created_at", "2020-05-25T15:26:47.306509Z"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientUserConnectionMembershipsDataSourceConfigMapping(t)
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
