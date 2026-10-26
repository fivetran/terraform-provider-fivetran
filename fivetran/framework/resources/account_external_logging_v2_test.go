package resources_test

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/resources"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccountExternalLoggingV2SchemaShape(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := resources.AccountExternalLoggingV2()
	var schemaResp fwresource.SchemaResponse
	r.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", schemaResp.Diagnostics)
	}

	attrs := schemaResp.Schema.Attributes
	if _, ok := attrs["group_id"]; ok {
		t.Fatal("fivetran_account_external_logging_v2 must not expose group_id")
	}
	service, ok := attrs["service"].(resourceSchema.StringAttribute)
	if !ok || !service.Required {
		t.Fatalf("service must be a required string attribute, got %T", attrs["service"])
	}
	config, ok := attrs["config"].(resourceSchema.DynamicAttribute)
	if !ok {
		t.Fatalf("config has type %T, want DynamicAttribute", attrs["config"])
	}
	if !config.Sensitive {
		t.Fatal("config holds credentials and must be sensitive")
	}
	if !config.Required || config.Optional || config.Computed {
		t.Fatal("config must be required: the API rejects a create without config")
	}
}

func TestAccountExternalLoggingV2IsRegistered(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	for _, factory := range framework.FivetranProvider().Resources(ctx) {
		var metadataResp fwresource.MetadataResponse
		factory().Metadata(ctx, fwresource.MetadataRequest{ProviderTypeName: "fivetran"}, &metadataResp)
		if metadataResp.TypeName == "fivetran_account_external_logging_v2" {
			return
		}
	}
	t.Fatal("fivetran_account_external_logging_v2 must be registered")
}

func accountExternalLoggingV2TfConfig(service, enabled, config string) string {
	return `
		resource "fivetran_account_external_logging_v2" "log" {
			provider = fivetran-provider

			service = "` + service + `"
			enabled = ` + enabled + `
			config = {
				` + config + `
			}
		}`
}

// A server without log service metadata answers NotFound.
// The resource must still work, without plan-time validation.
func stubLogMetadataNotFoundV2(t *testing.T, service string) {
	tfmock.MockClient().When(http.MethodGet, "/v1/metadata/log-service-types/"+service).ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return mock.NewResponse(req, http.StatusNotFound,
				`{"code":"NotFound_Service","message":"Service not found"}`), nil
		},
	)
}

func TestAccountExternalLoggingV2LifecycleMock(t *testing.T) {
	var logData map[string]interface{}
	var postHandler, patchHandler, deleteHandler *mock.Handler

	step1 := resource.TestStep{
		Config: accountExternalLoggingV2TfConfig("datadog_log", "true", `api_key = "secret"
				site = "datadoghq.com"`),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, postHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_account_external_logging_v2.log", "id", "log_id"),
			resource.TestCheckResourceAttr("fivetran_account_external_logging_v2.log", "service", "datadog_log"),
			resource.TestCheckResourceAttr("fivetran_account_external_logging_v2.log", "enabled", "true"),
			// the API returns the secret masked, the configured value must be kept
			resource.TestCheckResourceAttr("fivetran_account_external_logging_v2.log", "config.api_key", "secret"),
			resource.TestCheckResourceAttr("fivetran_account_external_logging_v2.log", "config.site", "datadoghq.com"),
		),
	}

	step2 := resource.TestStep{
		Config: accountExternalLoggingV2TfConfig("datadog_log", "false", `api_key = "secret"
				site = "datadoghq.eu"`),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, patchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_account_external_logging_v2.log", "enabled", "false"),
			resource.TestCheckResourceAttr("fivetran_account_external_logging_v2.log", "config.api_key", "secret"),
			resource.TestCheckResourceAttr("fivetran_account_external_logging_v2.log", "config.site", "datadoghq.eu"),
		),
	}

	// Removing a key from config sends no PATCH when metadata is unavailable (a removal cannot
	// be expressed without knowing the field is nullable). State must still drop the key.
	step3 := resource.TestStep{
		Config: accountExternalLoggingV2TfConfig("datadog_log", "false", `api_key = "secret"`),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, patchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_account_external_logging_v2.log", "config.api_key", "secret"),
			resource.TestCheckNoResourceAttr("fivetran_account_external_logging_v2.log", "config.site"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()
				stubLogMetadataNotFoundV2(t, "datadog_log")

				tfmock.MockClient().When(http.MethodGet, "/v1/external-logging/account").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", logData), nil
					},
				)

				postHandler = tfmock.MockClient().When(http.MethodPost, "/v1/external-logging/account").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						tfmock.AssertKeyDoesNotExist(t, body, "group_id")
						tfmock.AssertKeyExistsAndHasValue(t, body, "service", "datadog_log")
						tfmock.AssertKeyExistsAndHasValue(t, body, "enabled", true)
						tfmock.AssertKeyExistsAndHasValue(t, body, "run_setup_tests", false)
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "api_key", "secret")
						tfmock.AssertKeyExistsAndHasValue(t, config, "site", "datadoghq.com")

						logData = map[string]interface{}{
							"id":      "log_id",
							"service": "datadog_log",
							"enabled": true,
							"config": map[string]interface{}{
								"api_key": "******",
								"site":    "datadoghq.com",
							},
						}
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", logData), nil
					},
				)

				patchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/external-logging/account").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						tfmock.AssertKeyExistsAndHasValue(t, body, "enabled", false)
						// only changed config fields are sent
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "site", "datadoghq.eu")
						tfmock.AssertKeyDoesNotExist(t, config, "api_key")

						logData["enabled"] = false
						logData["config"].(map[string]interface{})["site"] = "datadoghq.eu"
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", logData), nil
					},
				)

				deleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/external-logging/account").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						logData = nil
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK,
							"Account external logging service has been deleted", nil), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, deleteHandler.Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{step1, step2, step3},
		},
	)
}

// When the metadata API describes the log service, unknown config fields are rejected at plan time.
func TestAccountExternalLoggingV2ValidatesConfigAgainstMetadataMock(t *testing.T) {
	metadataBody := `{
		"id": "grafana_loki",
		"name": "Grafana Loki",
		"type": "Log",
		"config": {
			"type": "object",
			"required": ["api_key"],
			"properties": {
				"api_key": {"type": "string", "format": "password"},
				"site": {"type": "string"}
			}
		}
	}`

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()
				tfmock.MockClient().When(http.MethodGet, "/v1/metadata/log-service-types/grafana_loki").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success",
							tfmock.CreateMapFromJsonString(t, metadataBody)), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: accountExternalLoggingV2TfConfig("grafana_loki", "true", `api_key = "secret"
					unknown_field = "x"`),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`Unsupported Dynamic Field`),
				},
				{
					Config:      accountExternalLoggingV2TfConfig("grafana_loki", "true", `site = "datadoghq.com"`),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`Missing Required Dynamic Field`),
				},
			},
		},
	)
}

func accountExternalLoggingV2TfConfigExpr(service, enabled, configExpr string) string {
	return `
		resource "fivetran_account_external_logging_v2" "log" {
			provider = fivetran-provider

			service = "` + service + `"
			enabled = ` + enabled + `
			config = ` + configExpr + `
		}`
}

// config can be written as an object literal or as a map (tomap). The state must keep the type of
// the planned value, also when only the type changes and no request is sent.
func TestAccountExternalLoggingV2MapValuedConfigMock(t *testing.T) {
	var logData map[string]interface{}
	var patchHandler *mock.Handler

	const resourceName = "fivetran_account_external_logging_v2.log"
	configChecks := func(hostname string) resource.TestCheckFunc {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "config.api_key", "secret"),
			resource.TestCheckResourceAttr(resourceName, "config.hostname", hostname),
			resource.TestCheckResourceAttr(resourceName, "config.region", "US1"),
		)
	}
	patchCalls := func(want int) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			tfmock.AssertEqual(t, patchHandler.Interactions, want)
			return nil
		}
	}
	mapConfig := func(hostname string) string {
		return accountExternalLoggingV2TfConfigExpr("datadog_log", "true",
			`tomap({ api_key = "secret", hostname = "`+hostname+`", region = "US1" })`)
	}
	objectConfig := accountExternalLoggingV2TfConfigExpr("datadog_log", "true",
		`{ api_key = "secret", hostname = "host", region = "US1" }`)

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()
				stubLogMetadataNotFoundV2(t, "datadog_log")

				tfmock.MockClient().When(http.MethodGet, "/v1/external-logging/account").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", logData), nil
					},
				)

				tfmock.MockClient().When(http.MethodPost, "/v1/external-logging/account").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						config := body["config"].(map[string]interface{})
						logData = map[string]interface{}{
							"id":      "log_id",
							"service": "datadog_log",
							"enabled": true,
							"config": map[string]interface{}{
								"api_key":  "******",
								"hostname": config["hostname"],
								"region":   config["region"],
							},
						}
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", logData), nil
					},
				)

				patchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/external-logging/account").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						body := tfmock.RequestBodyToJson(t, req)
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "hostname", "host2")
						tfmock.AssertKeyDoesNotExist(t, config, "api_key")
						logData["config"].(map[string]interface{})["hostname"] = "host2"
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", logData), nil
					},
				)

				tfmock.MockClient().When(http.MethodDelete, "/v1/external-logging/account").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						logData = nil
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK,
							"Account external logging service has been deleted", nil), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// create with a map
				{Config: mapConfig("host"), Check: resource.ComposeAggregateTestCheckFunc(configChecks("host"), patchCalls(0))},
				// type-only update: map -> object, no request
				{Config: objectConfig, Check: resource.ComposeAggregateTestCheckFunc(configChecks("host"), patchCalls(0))},
				// type-only update: object -> map, no request
				{Config: mapConfig("host"), Check: resource.ComposeAggregateTestCheckFunc(configChecks("host"), patchCalls(0))},
				// real change of a map value sends one PATCH
				{Config: mapConfig("host2"), Check: resource.ComposeAggregateTestCheckFunc(configChecks("host2"), patchCalls(1))},
			},
		},
	)
}
