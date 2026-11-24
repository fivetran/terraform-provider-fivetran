package resources_test

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// The provider caches metadata per service across tests, so each test here uses a service no other
// test does, and stubs its metadata as not found: the API validates the config.

const externalLoggingV2AccountName = "fivetran_external_logging_v2.log"

func externalLoggingV2AccountTfConfig(groupId, service, enabled, config string) string {
	group := ""
	if groupId != "" {
		group = `group_id = "` + groupId + `"`
	}
	return `
		resource "fivetran_external_logging_v2" "log" {
			provider = fivetran-provider

			` + group + `
			service         = "` + service + `"
			enabled         = ` + enabled + `
			run_setup_tests = true
			config = {
				` + config + `
			}
		}`
}

// externalLoggingV2LogServiceMock serves one log service at path, the way the API stores it: id is
// the ID the API assigns, secrets are returned masked.
type externalLoggingV2LogServiceMock struct {
	t      *testing.T
	path   string
	id     string
	data   map[string]interface{}
	post   *mock.Handler
	get    *mock.Handler
	patch  *mock.Handler
	delete *mock.Handler
	tests  *mock.Handler
	// skipUserAgent turns off the User-Agent check for requests the deprecated resource sends.
	skipUserAgent bool
}

func (m *externalLoggingV2LogServiceMock) skipUserAgentCheck() { m.skipUserAgent = true }
func (m *externalLoggingV2LogServiceMock) checkUserAgent()     { m.skipUserAgent = false }

func (m *externalLoggingV2LogServiceMock) assertUserAgent(req *http.Request) {
	if !m.skipUserAgent {
		assertExternalLoggingV2UserAgent(m.t, req)
	}
}

func newExternalLoggingV2LogServiceMock(t *testing.T, path, id string) *externalLoggingV2LogServiceMock {
	m := &externalLoggingV2LogServiceMock{t: t, path: path, id: id}
	client := tfmock.MockClient()
	m.post = client.When(http.MethodPost, m.createPath()).ThenCall(func(req *http.Request) (*http.Response, error) {
		m.assertUserAgent(req)
		body := tfmock.RequestBodyToJson(t, req)
		config, _ := body["config"].(map[string]interface{})
		m.data = map[string]interface{}{
			"id":      id,
			"service": body["service"],
			"enabled": body["enabled"],
			"config":  maskExternalLoggingV2Secrets(config),
		}
		return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", m.data), nil
	})
	m.get = client.When(http.MethodGet, path).ThenCall(func(req *http.Request) (*http.Response, error) {
		m.assertUserAgent(req)
		if m.data == nil {
			return mock.NewResponse(req, http.StatusNotFound,
				`{"code":"NotFound_LogService","message":"Log service not found"}`), nil
		}
		return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", m.data), nil
	})
	m.patch = client.When(http.MethodPatch, path).ThenCall(func(req *http.Request) (*http.Response, error) {
		m.assertUserAgent(req)
		body := tfmock.RequestBodyToJson(t, req)
		if enabled, ok := body["enabled"]; ok {
			m.data["enabled"] = enabled
		}
		if patch, ok := body["config"].(map[string]interface{}); ok {
			config := m.data["config"].(map[string]interface{})
			for k, v := range maskExternalLoggingV2Secrets(patch) {
				config[k] = v
			}
		}
		return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", m.data), nil
	})
	m.delete = client.When(http.MethodDelete, path).ThenCall(func(req *http.Request) (*http.Response, error) {
		m.assertUserAgent(req)
		m.data = nil
		return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "External logging service has been deleted", nil), nil
	})
	m.tests = client.When(http.MethodPost, path+"/test").ThenCall(func(req *http.Request) (*http.Response, error) {
		m.assertUserAgent(req)
		return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Setup tests have been completed",
			map[string]interface{}{"setup_tests": []interface{}{}}), nil
	})
	return m
}

// createPath is where a log service is created: the account-level path itself, or the collection
// for a destination log service, which takes group_id in the body.
func (m *externalLoggingV2LogServiceMock) createPath() string {
	if m.path == "/v1/external-logging/account" {
		return m.path
	}
	return "/v1/external-logging"
}

func maskExternalLoggingV2Secrets(config map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range config {
		if k == "api_key" || k == "token" {
			v = "******"
		}
		result[k] = v
	}
	return result
}

// Without group_id the resource manages the account-level log service through
// /v1/external-logging/account. Setting group_id later moves it to a destination by recreating it.
func TestExternalLoggingV2AccountLevelLifecycleMock(t *testing.T) {
	var account, group *externalLoggingV2LogServiceMock

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			tfmock.MockClient().Reset()
			stubLogMetadataNotFound(t, "cloudwatch")
			account = newExternalLoggingV2LogServiceMock(t, "/v1/external-logging/account", "account_id")
			group = newExternalLoggingV2LogServiceMock(t, "/v1/external-logging/group_id", "group_id")
		},
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			tfmock.AssertEqual(t, group.delete.Interactions, 1)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: externalLoggingV2AccountTfConfig("", "cloudwatch", "true", `api_key = "secret"
					region = "us-east-1"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, account.post.Interactions, 1)
						tfmock.AssertEqual(t, account.tests.Interactions, 1)
						tfmock.AssertEqual(t, group.post.Interactions, 0)
						return nil
					},
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "id", "account_id"),
					resource.TestCheckNoResourceAttr(externalLoggingV2AccountName, "group_id"),
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "service", "cloudwatch"),
					// the API returns the secret masked, the configured value must be kept
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "config.api_key", "secret"),
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "config.region", "us-east-1"),
				),
			},
			{
				Config: externalLoggingV2AccountTfConfig("", "cloudwatch", "false", `api_key = "secret"
					region = "eu-west-1"`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(externalLoggingV2AccountName, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, account.patch.Interactions, 1)
						tfmock.AssertEqual(t, account.data["enabled"], false)
						tfmock.AssertEqual(t, account.data["config"].(map[string]interface{})["region"], "eu-west-1")
						return nil
					},
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "enabled", "false"),
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "config.region", "eu-west-1"),
				),
			},
			{
				Config: externalLoggingV2AccountTfConfig("group_id", "cloudwatch", "false", `api_key = "secret"
					region = "eu-west-1"`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(externalLoggingV2AccountName, plancheck.ResourceActionReplace),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, account.delete.Interactions, 1)
						tfmock.AssertEqual(t, group.post.Interactions, 1)
						return nil
					},
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "id", "group_id"),
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "group_id", "group_id"),
				),
			},
		},
	})
}

// The account ID imports the account-level log service. It is not a group ID, so the destination
// lookup fails first.
func TestExternalLoggingV2ImportAccountLevelMock(t *testing.T) {
	var account *externalLoggingV2LogServiceMock

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			tfmock.MockClient().Reset()
			stubLogMetadataNotFound(t, "dynatrace")
			account = newExternalLoggingV2LogServiceMock(t, "/v1/external-logging/account", "account_id")
			account.data = map[string]interface{}{
				"id":      "account_id",
				"service": "dynatrace",
				"enabled": true,
				"config":  map[string]interface{}{"host": "dynatrace.example.com", "api_key": "******"},
			}
			tfmock.MockClient().When(http.MethodGet, "/v1/external-logging/account_id").ThenCall(
				func(req *http.Request) (*http.Response, error) {
					return mock.NewResponse(req, http.StatusNotFound,
						`{"code":"NotFound_LogService","message":"External logging service with id 'account_id' not found"}`), nil
				},
			)
		},
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			tfmock.AssertEqual(t, account.delete.Interactions, 1)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: `
					resource "fivetran_external_logging_v2" "log" {
						provider = fivetran-provider

						service = "dynatrace"
						config = {
							host = "dynatrace.example.com"
						}
						config_secrets = {
							api_key = "secret"
						}
						config_secrets_version = 1
					}

					import {
						to = fivetran_external_logging_v2.log
						id = "account_id"
					}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(externalLoggingV2AccountName, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, account.post.Interactions, 0)
						// the first config_secrets_version sends the secret to the account-level endpoint
						tfmock.AssertEqual(t, account.patch.Interactions, 1)
						return nil
					},
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "id", "account_id"),
					resource.TestCheckNoResourceAttr(externalLoggingV2AccountName, "group_id"),
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "config.host", "dynatrace.example.com"),
					resource.TestCheckNoResourceAttr(externalLoggingV2AccountName, "config.api_key"),
				),
			},
		},
	})
}

// The import ID "account" is an alias for the account-level log service. State stores the real ID,
// the account ID, so the imported resource is the same as one imported by the account ID.
func TestExternalLoggingV2ImportAccountAliasMock(t *testing.T) {
	var account *externalLoggingV2LogServiceMock

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			tfmock.MockClient().Reset()
			stubLogMetadataNotFound(t, "stackdriver")
			account = newExternalLoggingV2LogServiceMock(t, "/v1/external-logging/account", "account_id")
			account.data = map[string]interface{}{
				"id":      "account_id",
				"service": "stackdriver",
				"enabled": true,
				"config":  map[string]interface{}{"region": "US"},
			}
		},
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			tfmock.AssertEqual(t, account.delete.Interactions, 1)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: `
					resource "fivetran_external_logging_v2" "log" {
						provider = fivetran-provider

						service = "stackdriver"
						config = {
							region = "US"
						}
					}

					import {
						to = fivetran_external_logging_v2.log
						id = "account"
					}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(externalLoggingV2AccountName, plancheck.ResourceActionNoop),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, account.post.Interactions, 0)
						tfmock.AssertEqual(t, account.patch.Interactions, 0)
						return nil
					},
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "id", "account_id"),
					resource.TestCheckNoResourceAttr(externalLoggingV2AccountName, "group_id"),
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "config.region", "US"),
				),
			},
		},
	})
}

// The alias fails the import when the account has no account-level log service.
func TestExternalLoggingV2ImportAccountAliasNotFoundMock(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			tfmock.MockClient().Reset()
			newExternalLoggingV2LogServiceMock(t, "/v1/external-logging/account", "account_id")
		},
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "fivetran_external_logging_v2" "log" {
						provider = fivetran-provider

						service = "stackdriver"
						config = {
							region = "US"
						}
					}

					import {
						to = fivetran_external_logging_v2.log
						id = "account"
					}`,
				ExpectError: regexp.MustCompile(`Unable to Import External Logging Resource`),
			},
		},
	})
}

// A moved block moves fivetran_account_external_logging_v2 to fivetran_external_logging_v2 without
// recreating or changing the log service.
func TestExternalLoggingV2MovedFromAccountExternalLoggingV2Mock(t *testing.T) {
	var account *externalLoggingV2LogServiceMock

	const accountResource = `
		resource "fivetran_account_external_logging_v2" "log" {
			provider = fivetran-provider

			service = "stackdriver"
			config = {
				project_id = "project"
			}
		}`
	const v2Resource = `
		resource "fivetran_external_logging_v2" "log" {
			provider = fivetran-provider

			service = "stackdriver"
			config = {
				project_id = "project"
			}
		}

		moved {
			from = fivetran_account_external_logging_v2.log
			to   = fivetran_external_logging_v2.log
		}`

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			tfmock.MockClient().Reset()
			stubLogMetadataNotFound(t, "stackdriver")
			account = newExternalLoggingV2LogServiceMock(t, "/v1/external-logging/account", "account_id")
		},
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			tfmock.AssertEqual(t, account.delete.Interactions, 1)
			return nil
		},
		Steps: []resource.TestStep{
			{
				// the deprecated resource sends its own User-Agent suffix
				PreConfig: func() { account.skipUserAgentCheck() },
				Config:    accountResource,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, account.post.Interactions, 1)
						return nil
					},
					resource.TestCheckResourceAttr("fivetran_account_external_logging_v2.log", "id", "account_id"),
				),
			},
			{
				PreConfig: func() { account.checkUserAgent() },
				Config:    v2Resource,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(externalLoggingV2AccountName, plancheck.ResourceActionNoop),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, account.post.Interactions, 1)
						tfmock.AssertEqual(t, account.patch.Interactions, 0)
						tfmock.AssertEqual(t, account.delete.Interactions, 0)
						return nil
					},
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "id", "account_id"),
					resource.TestCheckNoResourceAttr(externalLoggingV2AccountName, "group_id"),
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "service", "stackdriver"),
					resource.TestCheckResourceAttr(externalLoggingV2AccountName, "config.project_id", "project"),
				),
			},
		},
	})
}
