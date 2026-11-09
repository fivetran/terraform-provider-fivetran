package resources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const externalLoggingV2MigratedConfig = `
		resource "fivetran_external_logging_v2" "log" {
			provider = fivetran-provider

			group_id = "group_id"
			service  = "new_relic_log"
			enabled  = true
			config = {
				api_key = "secret"
				region  = "US1"
			}
		}`

// Migrating from fivetran_external_logging with import and removed blocks, as the
// fivetran_connection_v2 migration guide does, keeps the log service: nothing is deleted or
// recreated, even though the imported state lacks the immutable secret api_key.
// The provider caches metadata per service across tests, so this test uses a service no other test does.
func TestExternalLoggingV2MigrationFromV1Mock(t *testing.T) {
	var logData map[string]interface{}
	var postHandler, patchHandler, deleteHandler *mock.Handler

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()
				tfmock.MockClient().When(http.MethodGet, "/v1/metadata/log-service-types/new_relic_log").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, `{
							"id": "new_relic_log",
							"name": "New Relic",
							"type": "Log",
							"config": {
								"type": "object",
								"required": ["api_key"],
								"properties": {
									"api_key": {"type": "string", "format": "password", "immutable": true},
									"region": {"type": "string"}
								}
							}
						}`)), nil
					},
				)

				tfmock.MockClient().When(http.MethodGet, "/v1/external-logging/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if logData == nil {
							return mock.NewResponse(req, http.StatusNotFound,
								`{"code":"NotFound_LogService","message":"Log service not found"}`), nil
						}
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", logData), nil
					},
				)

				postHandler = tfmock.MockClient().When(http.MethodPost, "/v1/external-logging").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						logData = map[string]interface{}{
							"id":      "group_id",
							"service": "new_relic_log",
							"enabled": true,
							"config":  map[string]interface{}{"api_key": "******", "region": "US1"},
						}
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", logData), nil
					},
				)

				patchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/external-logging/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", logData), nil
					},
				)

				deleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/external-logging/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						logData = nil
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK,
							"External logging service has been deleted", nil), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						resource "fivetran_external_logging" "log" {
							provider = fivetran-provider

							group_id = "group_id"
							service  = "new_relic_log"
							enabled  = true

							config {
								api_key = "secret"
								region  = "US1"
							}
						}`,
					Check: func(s *terraform.State) error {
						tfmock.AssertEqual(t, postHandler.Interactions, 1)
						return nil
					},
				},
				{
					Config: externalLoggingV2MigratedConfig + `

						import {
							to = fivetran_external_logging_v2.log
							id = "group_id"
						}

						removed {
							from = fivetran_external_logging.log
							lifecycle {
								destroy = false
							}
						}`,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fivetran_external_logging_v2.log", plancheck.ResourceActionUpdate),
					}},
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, postHandler.Interactions, 1)
							tfmock.AssertEqual(t, deleteHandler.Interactions, 0)
							// the immutable secret is never sent in an update
							tfmock.AssertEqual(t, patchHandler.Interactions, 0)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "id", "group_id"),
						resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "config.api_key", "secret"),
						resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "config.region", "US1"),
					),
				},
				{
					// the migration blocks are removed: nothing is left to change
					Config:   externalLoggingV2MigratedConfig,
					PlanOnly: true,
				},
			},
		},
	)
}
