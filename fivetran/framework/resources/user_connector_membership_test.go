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
	userConnectorMembershipPostHandler   *mock.Handler
	userConnectorMembershipPatchHandler  *mock.Handler
	userConnectorMembershipDeleteHandler *mock.Handler
	userConnectorMembershipData     map[string]interface{}
	userConnectorMembershipListData map[string]interface{}
	userConnectorMembershipResponse string
)

func setupMockClientUserConnectorMembershipResource(t *testing.T) {
	tfmock.MockClient().Reset()
	userConnectorMembershipResponse =
		`{
        "id": "test_connector",
        "role": "Connector Reviewer",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`

	userConnectorMembershipUpdatedResponse :=
		`{
        "id": "test_connector",
        "role": "Connector Administrator",
        "created_at": "2020-05-25T15:26:47.306509Z"
    }`

	userConnectorMembershipResponse = `{
             "items": [
                {
                    "id": "test_connector",
                    "role": "Connector Reviewer",
                    "created_at": "2020-05-25T15:26:47.306509Z"
                }
                ],
                "next_cursor": null}`

	userConnectorMembershipPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/users/test_user/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userConnectorMembershipData = tfmock.CreateMapFromJsonString(t, userConnectorMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Connector membership has been created", userConnectorMembershipData)
			return response, nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/users/test_user/connectors/test_connector").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userConnectorMembershipData = tfmock.CreateMapFromJsonString(t, userConnectorMembershipUpdatedResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", userConnectorMembershipData), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/users/test_user/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userConnectorMembershipListData = tfmock.CreateMapFromJsonString(t, userConnectorMembershipResponse)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", userConnectorMembershipListData)
			return response, nil
		},
	)

	userConnectorMembershipPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/users/test_user/connectors/test_connector").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Connector membership has been updated", userConnectorMembershipData), nil
		},
	)

	userConnectorMembershipDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/users/test_user/connectors/test_connector").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "Connector membership has been deleted", nil), nil
		},
	)
}

func TestConnectorMembershipResourceUserMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_user_connector_membership" "test_user_connector_membership" {
                 provider = fivetran-provider

                 user_id = "test_user"
                 
                 connector {
                    connector_id = "test_connector"
                    role = "Connector Reviewer"                    
                 }
            }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, userConnectorMembershipPostHandler.Interactions, 1)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientUserConnectorMembershipResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, userConnectorMembershipDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
