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
	teamUserMembershipPostHandler   *mock.Handler
	teamUserMembershipPatchHandler  *mock.Handler
	teamUserMembershipDeleteHandler *mock.Handler
	teamUserMembershipData     map[string]interface{}
	teamUserMembershipListData map[string]interface{}
	teamUserMembershipResponse string
	teamUserMembershipResponse2 string
	deleteCount						int
)

func setupMockClientTeamUserMembershipResource(t *testing.T) {
	tfmock.MockClient().Reset()
	teamUserMembershipResponse =
		`{
        "user_id": "test_user",
        "role": "Team Member"
    }`

	teamUserMembershipUpdatedResponse :=
		`{
        "user_id": "test_user",
        "role": "Team Manager"
    }`

	teamUserMembershipResponse =
				`{
                "items": [
                {
                    "user_id": "test_user",
                    "role": "Team Manager"
                }
                ],
                 "next_cursor": null}`

	teamUserMembershipPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/teams/test_team/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamUserMembershipData = tfmock.CreateMapFromJsonString(t, teamUserMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "User membership has been created", teamUserMembershipData)
			return response, nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/teams/test_team/users/test_user").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamUserMembershipData = tfmock.CreateMapFromJsonString(t, teamUserMembershipUpdatedResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", teamUserMembershipData), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/teams/test_team/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamUserMembershipListData = tfmock.CreateMapFromJsonString(t, teamUserMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", teamUserMembershipListData)
			return response, nil
		},
	)

	teamUserMembershipPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/teams/test_team/users/test_user").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "User membership has been updated", teamUserMembershipData), nil
		},
	)

	teamUserMembershipDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/teams/test_team/users/test_user").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "User membership has been deleted", nil), nil
		},
	)
}

func TestUserMembershipResourceTeamMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_team_user_membership" "test_team_user_membership" {
                 provider = fivetran-provider

                 team_id = "test_team"
                 
                 user {
                    user_id = "test_user"
                    role = "Team Manager"                    
                 }
            }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamUserMembershipPostHandler.Interactions, 1)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamUserMembershipResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamUserMembershipDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}


func setupMockClientTeamUserMembershipResourceNotFound(t *testing.T) {
	tfmock.MockClient().Reset()
	teamUserMembershipResponse =
		`{
        "user_id": "test_user",
        "role": "Team Member"
    }`

	teamUserMembershipResponse2 =
		`{
        "user_id": "test_user2",
        "role": "Team Member"
    }`

	teamUserMembershipResponse =
				`{
                "items": [
                {
                    "user_id": "test_user",
                    "role": "Team Manager"
                }
                ],
                 "next_cursor": null}`


	callCount := 0
	teamUserMembershipPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/teams/test_team/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount != 0 {
				teamUserMembershipData = tfmock.CreateMapFromJsonString(t, teamUserMembershipResponse)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "User membership has been created", teamUserMembershipData), nil
			}
			teamUserMembershipData = tfmock.CreateMapFromJsonString(t, teamUserMembershipResponse2)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "User not found", teamUserMembershipData), nil
		},
	)

	deleteCount := 0;
	teamUserMembershipDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/teams/test_team/users/test_user").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			deleteCount++
			return tfmock.FivetranSuccessResponse(t, req, 200, "User membership has been deleted", nil), nil
		},
	)
}

func TestUserMembershipResourceTeamMockNotFound(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_team_user_membership" "test_team_user_membership" {
                 provider = fivetran-provider

                 team_id = "test_team"
                 
                 user {
                    user_id = "test_user"
                    role = "Team Manager"                    
                 }
                 user {
                    user_id = "test_user2"
                    role = "Team Manager"                    
                 }
            }`,

		ExpectError: regexp.MustCompile(`Error: Unable to Create Team User Memberships Resource`),

	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamUserMembershipResourceNotFound(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamUserMembershipDeleteHandler.Interactions, 1)
				if (deleteCount != 1) {
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
