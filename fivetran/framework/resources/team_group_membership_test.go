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
	teamGroupMembershipPostHandler   *mock.Handler
	teamGroupMembershipPatchHandler  *mock.Handler
	teamGroupMembershipDeleteHandler *mock.Handler
	teamGroupMembershipData     map[string]interface{}
	teamGroupMembershipListData map[string]interface{}
	teamGroupMembershipResponse string
)

func setupMockClientTeamGroupMembershipResource(t *testing.T) {
	tfmock.MockClient().Reset()
	teamGroupMembershipResponse =
		`{
        "id": "test_group",
        "role": "Destination Reviewer",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`

	teamGroupMembershipUpdatedResponse :=
		`{
        "id": "test_group",
        "role": "Destination Administrator",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`

	teamGroupMembershipResponse =
				`{
                "items": [
                {
                    "id": "test_group",
                    "role": "Destination Reviewer",
                    "created_at": "2020-05-25T15:26:47.306509Z"
                }
                ],
         "next_cursor": null}`

	teamGroupMembershipPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/teams/test_team/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamGroupMembershipData = tfmock.CreateMapFromJsonString(t, teamGroupMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Group membership has been created", teamGroupMembershipData)
			return response, nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/teams/test_team/groups/test_group").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamGroupMembershipData = tfmock.CreateMapFromJsonString(t, teamGroupMembershipUpdatedResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", teamGroupMembershipData), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/teams/test_team/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamGroupMembershipListData = tfmock.CreateMapFromJsonString(t, teamGroupMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", teamGroupMembershipListData)
			return response, nil
		},
	)

	teamGroupMembershipPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/teams/test_team/groups/test_group").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Group membership has been updated", teamGroupMembershipData), nil
		},
	)

	teamGroupMembershipDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/teams/test_team/groups/test_group").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "Group membership has been deleted", nil), nil
		},
	)
}

func TestGroupMembershipResourceTeamMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_team_group_membership" "test_team_group_membership" {
                 provider = fivetran-provider

                 team_id = "test_team"
                 
                 group {
                    group_id = "test_group"
                    role = "Destination Reviewer"                    
                 }
            }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamGroupMembershipPostHandler.Interactions, 1)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamGroupMembershipResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamGroupMembershipDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
