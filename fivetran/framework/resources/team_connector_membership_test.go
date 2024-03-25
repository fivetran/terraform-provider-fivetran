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
	teamConnectorMembershipPostHandler   *mock.Handler
	teamConnectorMembershipPatchHandler  *mock.Handler
	teamConnectorMembershipDeleteHandler *mock.Handler
	teamConnectorMembershipData     map[string]interface{}
	teamConnectorMembershipListData map[string]interface{}
	teamConnectorMembershipResponse string
)

func setupMockClientTeamConnectorMembershipResource(t *testing.T) {
	tfmock.MockClient().Reset()
	teamConnectorMembershipResponse =
		`{
        "id": "test_connector",
        "role": "Connector Reviewer",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`

	teamConnectorMembershipUpdatedResponse :=
		`{
        "id": "test_connector",
        "role": "Connector Administrator",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`

	teamConnectorMembershipResponse = `{
             "items": [
                {
                    "id": "test_connector",
                    "role": "Connector Reviewer",
                    "created_at": "2020-05-25T15:26:47.306509Z"
                }
                ],
                "next_cursor": null}`

	teamConnectorMembershipPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/teams/test_team/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamConnectorMembershipData = tfmock.CreateMapFromJsonString(t, teamConnectorMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Connector membership has been created", teamConnectorMembershipData)
			return response, nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/teams/test_team/connectors/test_connector").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamConnectorMembershipData = tfmock.CreateMapFromJsonString(t, teamConnectorMembershipUpdatedResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", teamConnectorMembershipData), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/teams/test_team/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamConnectorMembershipListData = tfmock.CreateMapFromJsonString(t, teamConnectorMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", teamConnectorMembershipListData)
			return response, nil
		},
	)

	teamConnectorMembershipPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/teams/test_team/connectors/test_connector").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Connector membership has been updated", teamConnectorMembershipData), nil
		},
	)

	teamConnectorMembershipDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/teams/test_team/connectors/test_connector").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "Connector membership has been deleted", nil), nil
		},
	)
}

func TestConnectorMembershipResourceTeamMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_team_connector_membership" "test_team_connector_membership" {
                 provider = fivetran-provider

                 team_id = "test_team"
                 
                 connector {
                    connector_id = "test_connector"
                    role = "Connector Reviewer"                    
                 }
            }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamConnectorMembershipPostHandler.Interactions, 1)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamConnectorMembershipResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamConnectorMembershipDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
