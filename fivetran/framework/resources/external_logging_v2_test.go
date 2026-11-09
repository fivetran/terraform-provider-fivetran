package resources_test

import (
	"context"
	"net/http"
	"regexp"
	"strings"
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

func TestExternalLoggingV2SchemaShape(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := resources.ExternalLoggingV2()
	var schemaResp fwresource.SchemaResponse
	r.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", schemaResp.Diagnostics)
	}

	attrs := schemaResp.Schema.Attributes
	groupId, ok := attrs["group_id"].(resourceSchema.StringAttribute)
	if !ok || !groupId.Required {
		t.Fatalf("group_id must be a required string attribute, got %T", attrs["group_id"])
	}
	service, ok := attrs["service"].(resourceSchema.StringAttribute)
	if !ok || !service.Required {
		t.Fatalf("service must be a required string attribute, got %T", attrs["service"])
	}
	config, ok := attrs["config"].(resourceSchema.DynamicAttribute)
	if !ok {
		t.Fatalf("config has type %T, want DynamicAttribute", attrs["config"])
	}
	// secrets belong in config_secrets, so config stays readable in plans
	if config.Sensitive {
		t.Fatal("config must not be sensitive")
	}
	secrets, ok := attrs["config_secrets"].(resourceSchema.DynamicAttribute)
	if !ok || !secrets.Sensitive || !secrets.WriteOnly {
		t.Fatalf("config_secrets must be a sensitive write-only dynamic attribute, got %T", attrs["config_secrets"])
	}
	if !config.Required || config.Optional || config.Computed {
		t.Fatal("config must be required: the API rejects a create without config")
	}
	if len(schemaResp.Schema.Blocks) != 0 {
		t.Fatal("config must be a dynamic attribute, not a block")
	}
}

func TestExternalLoggingV2IsRegistered(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	for _, factory := range framework.FivetranProvider().Resources(ctx) {
		var metadataResp fwresource.MetadataResponse
		factory().Metadata(ctx, fwresource.MetadataRequest{ProviderTypeName: "fivetran"}, &metadataResp)
		if metadataResp.TypeName == "fivetran_external_logging_v2" {
			return
		}
	}
	t.Fatal("fivetran_external_logging_v2 must be registered")
}

func externalLoggingV2TfConfig(service, enabled, runSetupTests, config string) string {
	return `
		resource "fivetran_external_logging_v2" "log" {
			provider = fivetran-provider

			group_id        = "group_id"
			service         = "` + service + `"
			enabled         = ` + enabled + `
			run_setup_tests = ` + runSetupTests + `
			config = tomap({
				` + config + `
			})
		}`
}

func stubLogMetadataNotFound(t *testing.T, service string) {
	tfmock.MockClient().When(http.MethodGet, "/v1/metadata/log-service-types/"+service).ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return mock.NewResponse(req, http.StatusNotFound,
				`{"code":"NotFound_Service","message":"Service not found"}`), nil
		},
	)
}

// assertExternalLoggingV2UserAgent checks that a request carries the external logging v2 user-agent suffix.
func assertExternalLoggingV2UserAgent(t *testing.T, req *http.Request) {
	t.Helper()
	if userAgent := req.Header.Get("User-Agent"); !strings.HasSuffix(userAgent, "fivetran_external_logging_v2") {
		t.Errorf("%s %s: User-Agent %q does not end with the fivetran_external_logging_v2 suffix", req.Method, req.URL.Path, userAgent)
	}
}

func TestExternalLoggingV2LifecycleMock(t *testing.T) {
	var logData map[string]interface{}
	var postHandler, patchHandler, deleteHandler, testsHandler *mock.Handler

	step1 := resource.TestStep{
		Config: externalLoggingV2TfConfig("datadog_log", "true", "true", `api_key = "secret"
				site = "datadoghq.com"`),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, postHandler.Interactions, 1)
				// the create endpoint doesn't run the tests, the dedicated endpoint does
				tfmock.AssertEqual(t, testsHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "group_id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "service", "datadog_log"),
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "enabled", "true"),
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "run_setup_tests", "true"),
			// the API returns the secret masked, the configured value must be kept
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "config.api_key", "secret"),
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "config.site", "datadoghq.com"),
		),
	}

	step2 := resource.TestStep{
		Config: externalLoggingV2TfConfig("datadog_log", "false", "true", `api_key = "secret"
				site = "datadoghq.eu"`),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, patchHandler.Interactions, 1)
				// the PATCH runs the tests itself
				tfmock.AssertEqual(t, testsHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "enabled", "false"),
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "config.api_key", "secret"),
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "config.site", "datadoghq.eu"),
		),
	}

	// Removing a key from config sends no PATCH when metadata is unavailable. State must still drop the key.
	step3 := resource.TestStep{
		Config: externalLoggingV2TfConfig("datadog_log", "false", "false", `api_key = "secret"`),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, patchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "config.api_key", "secret"),
			resource.TestCheckNoResourceAttr("fivetran_external_logging_v2.log", "config.site"),
		),
	}

	// Turning run_setup_tests on without other changes runs the tests through the dedicated endpoint.
	step4 := resource.TestStep{
		Config: externalLoggingV2TfConfig("datadog_log", "false", "true", `api_key = "secret"`),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, patchHandler.Interactions, 1)
				tfmock.AssertEqual(t, testsHandler.Interactions, 2)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "run_setup_tests", "true"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()
				stubLogMetadataNotFound(t, "datadog_log")

				tfmock.MockClient().When(http.MethodGet, "/v1/external-logging/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						assertExternalLoggingV2UserAgent(t, req)
						if logData == nil {
							return mock.NewResponse(req, http.StatusNotFound,
								`{"code":"NotFound_LogService","message":"Log service not found"}`), nil
						}
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", logData), nil
					},
				)

				postHandler = tfmock.MockClient().When(http.MethodPost, "/v1/external-logging").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						assertExternalLoggingV2UserAgent(t, req)
						body := tfmock.RequestBodyToJson(t, req)
						tfmock.AssertKeyExistsAndHasValue(t, body, "group_id", "group_id")
						tfmock.AssertKeyExistsAndHasValue(t, body, "service", "datadog_log")
						tfmock.AssertKeyExistsAndHasValue(t, body, "enabled", true)
						tfmock.AssertKeyDoesNotExist(t, body, "run_setup_tests")
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "api_key", "secret")
						tfmock.AssertKeyExistsAndHasValue(t, config, "site", "datadoghq.com")

						logData = map[string]interface{}{
							"id":      "group_id",
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

				testsHandler = tfmock.MockClient().When(http.MethodPost, "/v1/external-logging/group_id/test").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						assertExternalLoggingV2UserAgent(t, req)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]interface{}{
							"setup_tests": []interface{}{
								map[string]interface{}{"title": "Connection test", "status": "PASSED", "message": ""},
							},
						}), nil
					},
				)

				patchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/external-logging/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						assertExternalLoggingV2UserAgent(t, req)
						body := tfmock.RequestBodyToJson(t, req)
						tfmock.AssertKeyExistsAndHasValue(t, body, "enabled", false)
						tfmock.AssertKeyExistsAndHasValue(t, body, "run_setup_tests", true)
						// only changed config fields are sent
						config := body["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "site", "datadoghq.eu")
						tfmock.AssertKeyDoesNotExist(t, config, "api_key")

						logData["enabled"] = false
						logData["config"].(map[string]interface{})["site"] = "datadoghq.eu"
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", logData), nil
					},
				)

				deleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/external-logging/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						assertExternalLoggingV2UserAgent(t, req)
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
			Steps: []resource.TestStep{step1, step2, step3, step4},
		},
	)
}

// config_secrets is write-only: it is sent on create and when config_secrets_version changes,
// and is never stored in state.
func TestExternalLoggingV2ConfigSecretsMock(t *testing.T) {
	var logData map[string]interface{}
	var postHandler, patchHandler *mock.Handler
	var lastPatchConfig map[string]interface{}

	tfConfig := func(site, secretsVersion string) string {
		return `
		resource "fivetran_external_logging_v2" "log" {
			provider = fivetran-provider

			group_id = "group_id"
			service  = "datadog_log"
			config = {
				site = "` + site + `"
			}
			config_secrets = {
				api_key = "secret"
			}
			config_secrets_version = ` + secretsVersion + `
		}`
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()
				stubLogMetadataNotFound(t, "datadog_log")

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
						config := tfmock.RequestBodyToJson(t, req)["config"].(map[string]interface{})
						tfmock.AssertKeyExistsAndHasValue(t, config, "api_key", "secret")
						tfmock.AssertKeyExistsAndHasValue(t, config, "site", "datadoghq.com")

						logData = map[string]interface{}{
							"id":      "group_id",
							"service": "datadog_log",
							"enabled": true,
							"config":  map[string]interface{}{"api_key": "******", "site": "datadoghq.com"},
						}
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", logData), nil
					},
				)

				patchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/external-logging/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						lastPatchConfig = tfmock.RequestBodyToJson(t, req)["config"].(map[string]interface{})
						if site, ok := lastPatchConfig["site"]; ok {
							logData["config"].(map[string]interface{})["site"] = site
						}
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", logData), nil
					},
				)

				tfmock.MockClient().When(http.MethodDelete, "/v1/external-logging/group_id").ThenCall(
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
					Config: tfConfig("datadoghq.com", "1"),
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, postHandler.Interactions, 1)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "config.site", "datadoghq.com"),
						// the secret is neither tracked in config nor stored in config_secrets
						resource.TestCheckNoResourceAttr("fivetran_external_logging_v2.log", "config.api_key"),
						resource.TestCheckNoResourceAttr("fivetran_external_logging_v2.log", "config_secrets"),
						resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "config_secrets_version", "1"),
					),
				},
				{
					// an unchanged config_secrets_version doesn't resend the secrets
					Config: tfConfig("datadoghq.eu", "1"),
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, patchHandler.Interactions, 1)
							tfmock.AssertKeyExistsAndHasValue(t, lastPatchConfig, "site", "datadoghq.eu")
							tfmock.AssertKeyDoesNotExist(t, lastPatchConfig, "api_key")
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "config.site", "datadoghq.eu"),
					),
				},
				{
					// a new config_secrets_version sends only the secrets
					Config: tfConfig("datadoghq.eu", "2"),
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, patchHandler.Interactions, 2)
							tfmock.AssertKeyExistsAndHasValue(t, lastPatchConfig, "api_key", "secret")
							tfmock.AssertKeyDoesNotExist(t, lastPatchConfig, "site")
							return nil
						},
						resource.TestCheckNoResourceAttr("fivetran_external_logging_v2.log", "config.api_key"),
						resource.TestCheckResourceAttr("fivetran_external_logging_v2.log", "config_secrets_version", "2"),
					),
				},
			},
		},
	)
}

func TestExternalLoggingV2ImportMock(t *testing.T) {
	logData := map[string]interface{}{
		"id":      "group_id",
		"service": "datadog_log",
		"enabled": true,
		"config": map[string]interface{}{
			"api_key": "******",
			"site":    "datadoghq.com",
		},
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()
				stubLogMetadataNotFound(t, "datadog_log")
				tfmock.MockClient().When(http.MethodGet, "/v1/external-logging/group_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", logData), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						resource "fivetran_external_logging_v2" "log" {
							provider = fivetran-provider
						}`,
					ResourceName:       "fivetran_external_logging_v2.log",
					ImportState:        true,
					ImportStateId:      "group_id",
					ImportStatePersist: false,
					ImportStateCheck: func(states []*terraform.InstanceState) error {
						if len(states) != 1 {
							t.Fatalf("imported %d resources, want 1", len(states))
						}
						attrs := states[0].Attributes
						tfmock.AssertEqual(t, attrs["id"], "group_id")
						tfmock.AssertEqual(t, attrs["group_id"], "group_id")
						tfmock.AssertEqual(t, attrs["service"], "datadog_log")
						tfmock.AssertEqual(t, attrs["config.site"], "datadoghq.com")
						// the masked secret must not be stored as if it were the real value
						if _, ok := attrs["config.api_key"]; ok {
							t.Errorf("masked api_key must not be imported, got %q", attrs["config.api_key"])
						}
						return nil
					},
				},
			},
		},
	)
}

// When the metadata API describes the log service, config is validated at plan time.
func TestExternalLoggingV2ValidatesConfigAgainstMetadataMock(t *testing.T) {
	metadataBody := `{
		"id": "grafana_loki",
		"name": "Grafana Loki",
		"type": "Log",
		"config": {
			"type": "object",
			"required": ["api_key"],
			"properties": {
				"api_key": {"type": "string", "format": "password"},
				"site": {"type": "string"},
				"port": {"type": "integer"},
				"external_id": {"type": "string", "readonly": true}
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
					Config: externalLoggingV2TfConfig("grafana_loki", "true", "false", `api_key = "secret"
					unknown_field = "x"`),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`Unsupported Dynamic Field`),
				},
				{
					Config:      externalLoggingV2TfConfig("grafana_loki", "true", "false", `site = "grafana.net"`),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`Missing Required Dynamic Field`),
				},
				{
					Config: externalLoggingV2TfConfig("grafana_loki", "true", "false", `api_key = "secret"
					port = "not a number"`),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`Invalid Dynamic Field`),
				},
				{
					Config: externalLoggingV2TfConfig("grafana_loki", "true", "false", `api_key = "secret"
					external_id = "x"`),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`Readonly Dynamic Field`),
				},
			},
		},
	)
}
