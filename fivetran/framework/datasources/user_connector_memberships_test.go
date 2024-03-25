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
	userConnectorMembershipsDataSourceMockGetHandler *mock.Handler
	userConnectorMembershipsDataSourceMockData       map[string]interface{}
)

const (
	userConnectorMembershipsMappingResponse = `
    {
      "items": [
        {
          "id": "connector_id_1",
          "role": "Connector Administrator",
          "created_at": "2020-05-25T15:26:47.306509Z"
        },
        {
          "id": "connector_id_2",
          "role": "Connector Reviewer",
          "created_at": "2020-05-25T15:26:47.306509Z"
        }
      ],
      "next_cursor": null
    }`
)

func setupMockClientUserConnectorMembershipsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	userDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userDataSourceMockData = tfmock.CreateMapFromJsonString(t, userMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", userDataSourceMockData), nil
		},
	)

	userConnectorMembershipsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/users/user_id/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userConnectorMembershipsDataSourceMockData = tfmock.CreateMapFromJsonString(t, userConnectorMembershipsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", userConnectorMembershipsDataSourceMockData), nil
		},
	)
}

func TestDataSourceUserConnectorMembershipsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_user_connector_memberships" "test_user_connector_memberships" {
            provider     = fivetran-provider
            user_id      = "user_id"
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, userConnectorMembershipsDataSourceMockGetHandler.Interactions, 2)
				tfmock.AssertNotEmpty(t, userConnectorMembershipsDataSourceMockData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientUserConnectorMembershipsDataSourceConfigMapping(t)
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
