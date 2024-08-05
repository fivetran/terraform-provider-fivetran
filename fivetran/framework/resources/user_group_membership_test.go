package resources_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
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
	userGroupMembershipResponse2 string
	userGroupmembershipdeleteCount int
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

func setupMockClientUserGroupMembershipResourceNotFoud(t *testing.T) {
	tfmock.MockClient().Reset()
	userGroupMembershipResponse =
		`{
        "id": "test_group",
        "role": "Destination Reviewer",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`
	
	userGroupMembershipResponse2 =
		`{
        "id": "test_group2",
        "role": "Destination Reviewer",
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

	callCount := 0
	userGroupMembershipPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/users/test_user/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount != 1 {
				userGroupMembershipData = tfmock.CreateMapFromJsonString(t, userGroupMembershipResponse)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Group membership has been created", userGroupMembershipData), nil
			}

			userGroupMembershipData = tfmock.CreateMapFromJsonString(t, userGroupMembershipResponse2)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "Group not found", userGroupMembershipData), nil
		},
	)

	userGroupmembershipdeleteCount := 0
	userGroupMembershipDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/users/test_user/groups/test_group").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userGroupmembershipdeleteCount++
			return tfmock.FivetranSuccessResponse(t, req, 200, "Group membership has been deleted", nil), nil
		},
	)
}

func TestGroupMembershipResourceUserMockNotFound(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_user_group_membership" "test_user_group_membership" {
                 provider = fivetran-provider

                 user_id = "test_user"
                 
                 group {
                    group_id = "test_group"
                    role = "Destination Reviewer"                    
                 }

                 group {
                    group_id = "test_group2"
                    role = "Destination Reviewer"                    
                 }
            }`,
		ExpectError: regexp.MustCompile(`Error: Unable to Create User Group Memberships Resource`),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientUserGroupMembershipResourceNotFoud(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, userGroupMembershipDeleteHandler.Interactions, 1)
				if (userGroupmembershipdeleteCount != 1) {
					return fmt.Errorf("Failed acctions are not reverted.")
				}
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
