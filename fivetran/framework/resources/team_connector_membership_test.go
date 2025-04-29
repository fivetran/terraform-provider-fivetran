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
	teamConnectorMembershipPostHandler   *mock.Handler
	teamConnectorMembershipPatchHandler  *mock.Handler
	teamConnectorMembershipDeleteHandler *mock.Handler
	teamConnectorMembershipData          map[string]interface{}
	teamConnectorMembershipListData      map[string]interface{}
	teamConnectorMembershipResponse      string
	teamConnectorMembershipResponse2     string
	teamConnectormembershipDeleteCount	int
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

	teamConnectorMembershipPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/teams/test_team/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamConnectorMembershipData = tfmock.CreateMapFromJsonString(t, teamConnectorMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Connector membership has been created", teamConnectorMembershipData)
			return response, nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/teams/test_team/connections/test_connector").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamConnectorMembershipData = tfmock.CreateMapFromJsonString(t, teamConnectorMembershipUpdatedResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", teamConnectorMembershipData), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/teams/test_team/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamConnectorMembershipListData = tfmock.CreateMapFromJsonString(t, teamConnectorMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", teamConnectorMembershipListData)
			return response, nil
		},
	)

	teamConnectorMembershipPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/teams/test_team/connections/test_connector").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Connector membership has been updated", teamConnectorMembershipData), nil
		},
	)

	teamConnectorMembershipDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/teams/test_team/connections/test_connector").ThenCall(
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

func setupMockClientTeamConnectorMembershipResourceNotFound(t *testing.T) {
	tfmock.MockClient().Reset()
	teamConnectorMembershipResponse =
		`{
        "id": "test_connector",
        "role": "Connector Reviewer",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`
	teamConnectorMembershipResponse2 =
		`{
        "id": "test_connector",
        "role": "Connector Reviewer",
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

	callCount := 0
	teamConnectorMembershipPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/teams/test_team/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount != 1 {
				teamConnectorMembershipData = tfmock.CreateMapFromJsonString(t, teamConnectorMembershipResponse)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Connector membership has been created", teamConnectorMembershipData), nil
			}
			teamConnectorMembershipData = tfmock.CreateMapFromJsonString(t, teamConnectorMembershipResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "Connector membership not found", teamConnectorMembershipData), nil
		},
	)

	teamConnectormembershipDeleteCount := 0;
	teamConnectorMembershipDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/teams/test_team/connections/test_connector").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamConnectormembershipDeleteCount++
			return tfmock.FivetranSuccessResponse(t, req, 200, "Connector membership has been deleted", nil), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/teams/test_team/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			teamConnectorMembershipListData = tfmock.CreateMapFromJsonString(t, teamConnectorMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", teamConnectorMembershipListData)
			return response, nil
		},
	)

}

func TestConnectorMembershipResourceTeamMockNotFound(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_team_connector_membership" "test_team_connector_membership" {
                 provider = fivetran-provider

                 team_id = "test_team"
                 
                 connector {
                    connector_id = "test_connector"
                    role = "Connector Reviewer"                    
                 }	
                 connector {
                    connector_id = "test_connector2"
                    role = "Connector Reviewer"                    
                 }	

            }`,
		ExpectError: regexp.MustCompile(`Error: Unable to Create Team Connector Memberships Resource`),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientTeamConnectorMembershipResourceNotFound(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, teamConnectorMembershipDeleteHandler.Interactions, 1)
				if (teamConnectormembershipDeleteCount != 1) {
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
