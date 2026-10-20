package resources_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// redshiftDestinationTfConfig returns a redshift destination. extra lines are added to the config block.
func redshiftDestinationTfConfig(extra string) string {
	return fmt.Sprintf(`
		resource "fivetran_destination" "mydestination" {
			provider = fivetran-provider

			group_id = "group_id"
			service = "redshift"
			time_zone_offset = "0"
			region = "GCP_US_EAST4"
			run_setup_tests = "false"

			config {
				host = "host"
				port = "5439"
				database = "database"
				user = "user"
				password = "password"
				%s
			}
		}`, extra)
}

// Returned by the API for redshift without cluster_id. The sensitive password is masked, as the API does.
const redshiftDestinationResponseWithoutClusterId = `
{
	"id":"destination_id",
	"group_id":"group_id",
	"service":"redshift",
	"region":"GCP_US_EAST4",
	"time_zone_offset":"0",
	"setup_status":"connected",
	"daylight_saving_time_enabled":false,
	"networking_method":"Directly",
	"config":{
		"host":"host",
		"port":"5439",
		"database":"database",
		"user":"user",
		"password":"******",
		"connection_type":"Directly"
	}
}
`

// Create: configured config subfields (including a sensitive one) are missing from the successful
// create response. Apply must succeed and keep the planned values in state. A refresh must then
// clear the non-sensitive one, so the next plan is not empty. The sensitive one is kept from prior
// state, as before.
func TestDestinationConfigUnconfirmedOnCreateMock(t *testing.T) {
	var testDestinationData map[string]interface{}
	var destinationPostHandler *mock.Handler
	var destinationDeleteHandler *mock.Handler

	step1 := resource.TestStep{
		Config: redshiftDestinationTfConfig(`cluster_id = "cluster_id"`),

		// Refresh clears the unreturned cluster_id, so the next plan must show it as a change
		ExpectNonEmptyPlan: true,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationPostHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "id", "destination_id"),
			// Unreturned configured values are kept in state for this apply
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.cluster_id", "cluster_id"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.password", "password"),
			// Returned values and their unrelated siblings are preserved
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.host", "host"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.database", "database"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.user", "user"),
		),
	}

	// Refresh only: the API still omits cluster_id, so state must not carry it forward.
	// The API returns the sensitive password masked, so the configured value is kept from prior state.
	step2 := resource.TestStep{
		RefreshState: true,
		// The follow-up plan must propose the unreturned cluster_id again
		ExpectNonEmptyPlan: true,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckNoResourceAttr("fivetran_destination.mydestination", "config.cluster_id"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.password", "password"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.host", "host"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.database", "database"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				tfmock.MockClient().When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testDestinationData), nil
					},
				)

				destinationPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "cluster_id", "cluster_id")
						tfmock.AssertKeyExistsAndHasValue(t, config, "password", "password")

						// cluster_id is accepted but not returned
						testDestinationData = tfmock.CreateMapFromJsonString(t, redshiftDestinationResponseWithoutClusterId)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", testDestinationData), nil
					},
				)

				destinationDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testDestinationData = nil
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK,
							"Destination with id 'destination_id' has been deleted", nil), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationDeleteHandler.Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{step1, step2},
		},
	)
}

// Update: a configured config subfield added in an update is missing from the PATCH response.
// Apply must succeed and keep the planned value. A refresh must then clear it, so the next plan
// shows it as a change again.
func TestDestinationConfigUnconfirmedOnUpdateMock(t *testing.T) {
	var testDestinationData map[string]interface{}
	var destinationPatchHandler *mock.Handler
	var destinationDeleteHandler *mock.Handler

	step1 := resource.TestStep{
		Config: redshiftDestinationTfConfig(""),
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckNoResourceAttr("fivetran_destination.mydestination", "config.cluster_id"),
		),
	}

	step2 := resource.TestStep{
		Config: redshiftDestinationTfConfig(`cluster_id = "cluster_id"`),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.cluster_id", "cluster_id"),
			resource.TestCheckResourceAttr("fivetran_destination.mydestination", "config.host", "host"),
		),
		// Refresh clears cluster_id again, so the next plan shows it as a change
		ExpectNonEmptyPlan: true,
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				tfmock.MockClient().When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testDestinationData), nil
					},
				)

				tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testDestinationData = tfmock.CreateMapFromJsonString(t, redshiftDestinationResponseWithoutClusterId)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", testDestinationData), nil
					},
				)

				destinationPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "cluster_id", "cluster_id")

						// cluster_id is accepted but not returned
						testDestinationData = tfmock.CreateMapFromJsonString(t, redshiftDestinationResponseWithoutClusterId)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testDestinationData), nil
					},
				)

				destinationDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						testDestinationData = nil
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK,
							"Destination with id 'destination_id' has been deleted", nil), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, destinationDeleteHandler.Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{step1, step2},
		},
	)
}
