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
	userGroupMembershipsDataSourceMockGetHandler *mock.Handler
	userGroupMembershipsDataSourceMockData       map[string]interface{}
)

const (
	userGroupMembershipsMappingResponse = `
    {
      "items": [
        {
          "id": "group_id_1",
          "role": "Destination Administrator",
          "created_at": "2020-05-25T15:26:47.306509Z"
        },
        {
          "id": "group_id_2",
          "role": "Destination Reviewer",
          "created_at": "2020-05-25T15:26:47.306509Z"
        }
      ],
      "next_cursor": null
    }`
)

func setupMockClientUserGroupMembershipsDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	userDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userDataSourceMockData = tfmock.CreateMapFromJsonString(t, userMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", userDataSourceMockData), nil
		},
	)

	userGroupMembershipsDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/users/user_id/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userGroupMembershipsDataSourceMockData = tfmock.CreateMapFromJsonString(t, userGroupMembershipsMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", userGroupMembershipsDataSourceMockData), nil
		},
	)
}

func TestDataSourceUserGroupMembershipsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_user_group_memberships" "test_user_group_memberships" {
            provider     = fivetran-provider
            user_id      = "user_id"
        }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, userGroupMembershipsDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, userGroupMembershipsDataSourceMockData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientUserGroupMembershipsDataSourceConfigMapping(t)
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
