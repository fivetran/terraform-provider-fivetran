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

// The provider caches metadata per service across tests, so this test uses a service no other test does.
const externalLoggingV2ImportMockMetadata = `
{
	"id": "splunkLog",
	"name": "Splunk",
	"type": "Log",
	"config": {
		"type": "object",
		"required": ["hostname", "token"],
		"properties": {
			"hostname": {"type": "string"},
			"port":     {"type": "integer", "nullable": true},
			"channel":  {"type": "string", "nullable": true},
			"token":    {"type": "string", "format": "password", "nullable": true},
			"api_key":  {"type": "string", "format": "password", "immutable": true}
		}
	}
}`

const externalLoggingV2ImportMockConfig = `
		resource "fivetran_external_logging_v2" "log" {
			provider = fivetran-provider

			group_id = "group_id"
			service  = "splunkLog"
			config = {
				hostname = "splunk.example.com"
			}
			config_secrets = {
				token   = "token"
				api_key = "key"
			}
			config_secrets_version = 1
		}`

// Import tracks the defaults the API returns (port, channel). A configuration that doesn't declare
// them keeps them in state instead of planning their removal, and never sends null for them,
// which would clear them on the log service.
// The first config_secrets_version sends the mutable token, but never the immutable api_key.
func TestExternalLoggingV2ImportUndeclaredFieldsAndSecretsMock(t *testing.T) {
	var logData map[string]interface{}
	var patchHandler, deleteHandler *mock.Handler
	var patchConfigs []map[string]interface{}

	const name = "fivetran_external_logging_v2.log"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			tfmock.MockClient().Reset()
			patchConfigs = nil
			logData = tfmock.CreateMapFromJsonString(t, `
			{
				"id": "group_id",
				"service": "splunkLog",
				"enabled": true,
				"config": {
					"hostname": "splunk.example.com",
					"port": 8088,
					"channel": "main",
					"token": "******",
					"api_key": "******"
				}
			}`)

			tfmock.MockClient().When(http.MethodGet, "/v1/metadata/log-service-types/splunkLog").ThenCall(
				func(req *http.Request) (*http.Response, error) {
					return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, externalLoggingV2ImportMockMetadata)), nil
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
			patchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/external-logging/group_id").ThenCall(
				func(req *http.Request) (*http.Response, error) {
					patchConfig, _ := tfmock.RequestBodyToJson(t, req)["config"].(map[string]interface{})
					patchConfigs = append(patchConfigs, patchConfig)
					config := logData["config"].(map[string]interface{})
					for k, v := range patchConfig {
						switch {
						case v == nil:
							delete(config, k)
						case k == "token" || k == "api_key":
							config[k] = "******"
						default:
							config[k] = v
						}
					}
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
		CheckDestroy: func(s *terraform.State) error {
			tfmock.AssertEqual(t, deleteHandler.Interactions, 1)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: externalLoggingV2ImportMockConfig + `

					import {
						to = fivetran_external_logging_v2.log
						id = "group_id"
					}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, patchHandler.Interactions, 1)
						patch := patchConfigs[0]
						// undeclared imported fields are not sent, so they are not cleared
						tfmock.AssertKeyDoesNotExist(t, patch, "port")
						tfmock.AssertKeyDoesNotExist(t, patch, "channel")
						config := logData["config"].(map[string]interface{})
						tfmock.AssertEqual(t, config["port"], float64(8088))
						tfmock.AssertEqual(t, config["channel"], "main")
						// the mutable secret is sent, the immutable one never is
						tfmock.AssertKeyExistsAndHasValue(t, patch, "token", "token")
						tfmock.AssertKeyDoesNotExist(t, patch, "api_key")
						return nil
					},
					resource.TestCheckResourceAttr(name, "config.hostname", "splunk.example.com"),
					resource.TestCheckResourceAttr(name, "config.port", "8088"),
					resource.TestCheckResourceAttr(name, "config.channel", "main"),
					resource.TestCheckNoResourceAttr(name, "config.token"),
					resource.TestCheckResourceAttr(name, "config_secrets_version", "1"),
				),
			},
			{
				// the import block is removed: nothing is left to change
				Config:   externalLoggingV2ImportMockConfig,
				PlanOnly: true,
			},
		},
	})
}
