package resources_test

import (
	"net/http"
	"testing"
	
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	userGroupMembershipPostHandler   *mock.Handler
	userGroupMembershipPatchHandler  *mock.Handler
	userGroupMembershipDeleteHandler *mock.Handler
	userGroupMembershipData     map[string]interface{}
	userGroupMembershipListData map[string]interface{}
	userGroupMembershipResponse string
)

func setupMockClientUserGroupMembershipResource(t *testing.T) {
	tfmock.MockClient().Reset()
	userGroupMembershipResponse =
		`{
        "id": "test_group",
        "role": "Destination Reviewer",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`

	userGroupMembershipUpdatedResponse :=
		`{
        "id": "test_group",
        "role": "Destination Administrator",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`

	userGroupMembershipResponse =
				`{
                "items": [
                {
                    "id": "test_group",
                    "role": "Destination Reviewer",
                    "created_at": "2020-05-25T15:26:47.306509Z"
                }
                ],
         "next_cursor": null}`

	userGroupMembershipPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/users/test_user/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userGroupMembershipData = tfmock.CreateMapFromJsonString(t, userGroupMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Group membership has been created", userGroupMembershipData)
			return response, nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/users/test_user/groups/test_group").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userGroupMembershipData = tfmock.CreateMapFromJsonString(t, userGroupMembershipUpdatedResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", userGroupMembershipData), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/users/test_user/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userGroupMembershipListData = tfmock.CreateMapFromJsonString(t, userGroupMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", userGroupMembershipListData)
			return response, nil
		},
	)

	userGroupMembershipPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/users/test_user/groups/test_group").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Group membership has been updated", userGroupMembershipData), nil
		},
	)

	userGroupMembershipDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/users/test_user/groups/test_group").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "Group membership has been deleted", nil), nil
		},
	)
}

func TestGroupMembershipResourceUserMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_user_group_membership" "test_user_group_membership" {
                 provider = fivetran-provider

                 user_id = "test_user"
                 
                 group {
                    group_id = "test_group"
                    role = "Destination Reviewer"                    
                 }
            }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, userGroupMembershipPostHandler.Interactions, 1)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientUserGroupMembershipResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, userGroupMembershipDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
