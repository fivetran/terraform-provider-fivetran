package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	teamGroupMembershipPostHandler   *mock.Handler
	teamGroupMembershipPatchHandler  *mock.Handler
	teamGroupMembershipDeleteHandler *mock.Handler
	//teamGroupMembershipTestHandler   *mock.Handler
	teamGroupMembershipData     map[string]interface{}
	teamGroupMembershipListData map[string]interface{}
	teamGroupMembershipResponse string
)

func setupMockClientTeamGroupMembershipResource(t *testing.T) {
	mockClient.Reset()
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

	teamGroupMembershipPostHandler = mockClient.When(http.MethodPost, "/v1/teams/test_team/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamGroupMembershipData = createMapFromJsonString(t, teamGroupMembershipResponse)
			response := fivetranSuccessResponse(t, req, http.StatusCreated, "Group membership has been created", teamGroupMembershipData)
			return response, nil
		},
	)

	mockClient.When(http.MethodGet, "/v1/teams/test_team/groups/test_group").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamGroupMembershipData = createMapFromJsonString(t, teamGroupMembershipUpdatedResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "", teamGroupMembershipData), nil
		},
	)

	mockClient.When(http.MethodGet, "/v1/teams/test_team/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamGroupMembershipListData = createMapFromJsonString(t, teamGroupMembershipResponse)
			response := fivetranSuccessResponse(t, req, http.StatusOK, "", teamGroupMembershipListData)

			// For list step after insert
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

			return response, nil
		},
	)

	teamGroupMembershipPatchHandler = mockClient.When(http.MethodPatch, "/v1/teams/test_team/groups/test_group").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Group membership has been updated", teamGroupMembershipData), nil
		},
	)

	teamGroupMembershipDeleteHandler = mockClient.When(http.MethodDelete, "/v1/teams/test_team/groups/test_group").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, 200, "Group membership has been deleted", nil), nil
		},
	)
}

func TestGroupMembershipResourceTeamMock(t *testing.T) {
	teamGroupMembershipResponse = `{
        "items": [],
        "next_cursor": null
    }`

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
				assertEqual(t, teamGroupMembershipPostHandler.Interactions, 1)
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
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, teamGroupMembershipDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
