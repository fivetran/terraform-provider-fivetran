package mock

import (
    "net/http"
    "testing"

    "github.com/fivetran/go-fivetran/tests/mock"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
    teamConnectorMembershipPostHandler   *mock.Handler
    teamConnectorMembershipPatchHandler  *mock.Handler
    teamConnectorMembershipDeleteHandler *mock.Handler
    teamConnectorMembershipTestHandler   *mock.Handler
    teamConnectorMembershipData          map[string]interface{}
    teamConnectorMembershipListData      map[string]interface{}
    teamConnectorMembershipResponse      string
)

func setupMockClientTeamConnectorMembershipResource(t *testing.T) {
    mockClient.Reset()
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


    teamConnectorMembershipPostHandler = mockClient.When(http.MethodPost, "/v1/teams/test_team/connectors").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            teamConnectorMembershipData = createMapFromJsonString(t, teamConnectorMembershipResponse)
            response := fivetranSuccessResponse(t, req, http.StatusCreated, "Connector membership has been created", teamConnectorMembershipData)
            return response, nil
        },
    )

    mockClient.When(http.MethodGet, "/v1/teams/test_team/connectors/test_connector").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            teamConnectorMembershipData = createMapFromJsonString(t, teamConnectorMembershipUpdatedResponse)
            return fivetranSuccessResponse(t, req, http.StatusOK, "", teamConnectorMembershipData), nil
        },
    )

    mockClient.When(http.MethodGet, "/v1/teams/test_team/connectors").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            teamConnectorMembershipListData = createMapFromJsonString(t, teamConnectorMembershipResponse)
            response := fivetranSuccessResponse(t, req, http.StatusOK, "", teamConnectorMembershipListData)

            // For list step after insert
            teamConnectorMembershipResponse = 
            `{
                "items": [
                {
                    "id": "test_connector",
                    "role": "Connector Reviewer",
                    "created_at": "2020-05-25T15:26:47.306509Z"
                }
                ],
                 "next_cursor": null}`

            return response, nil
        },
    )

    teamConnectorMembershipPatchHandler = mockClient.When(http.MethodPatch, "/v1/teams/test_team/connectors/test_connector").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return fivetranSuccessResponse(t, req, http.StatusOK, "Connector membership has been updated", teamConnectorMembershipData), nil
        },
    )

    teamConnectorMembershipDeleteHandler = mockClient.When(http.MethodDelete, "/v1/teams/test_team/connectors/test_connector").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return fivetranSuccessResponse(t, req, 200, "Connector membership has been deleted", nil), nil
        },
    )
}

func TestConnectorMembershipResourceTeamMock(t *testing.T) {
    teamConnectorMembershipResponse = `{
        "items": [],
        "next_cursor": null
    }`

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
                assertEqual(t, teamConnectorMembershipPostHandler.Interactions, 1)
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
            Providers: testProviders,
            CheckDestroy: func(s *terraform.State) error {
                assertEqual(t, teamConnectorMembershipDeleteHandler.Interactions, 1)
                return nil
            },

            Steps: []resource.TestStep{
                step1,
            },
        },
    )
}
