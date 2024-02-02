package resources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	webhookPostHandler   *mock.Handler
	webhookPatchHandler  *mock.Handler
	webhookDeleteHandler *mock.Handler
	webhookTestHandler   *mock.Handler
	webhookData          map[string]interface{}
)

func onTestWebhook(t *testing.T, req *http.Request) (*http.Response, error) {
	// setup test results array
	setupTests := make([]interface{}, 0)

	setupTestResult := make(map[string]interface{})
	setupTestResult["title"] = "Test Title"
	setupTestResult["status"] = "PASSED"
	setupTestResult["message"] = "Test passed"

	setupTests = append(setupTests, setupTestResult)

	webhookData["data"] = setupTests

	response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Setup tests have been completed", webhookData)
	return response, nil
}

func setupMockClientWebhookResource(t *testing.T) {
	tfmock.MockClient().Reset()
	webhookResponse :=
		`{
        "id": "webhook_id",
        "type": "account",
        "group_id": "_moonbeam",
        "url": "https://your-host.your-domain/webhook",
        "events": [
            "sync_start",
            "sync_end"
        ],
        "active": false,
        "secret": "******",
        "created_at": "2022-04-29T10:45:00.000Z",
        "created_by": "_airworthy"
    }`

	webhookUpdatedResponse :=
		`{
        "id": "webhook_id",
        "type": "account",
        "group_id": "_moonbeam",
        "url": "https://your-host.your-domain/webhook_1",
        "events": [
            "sync_start",
            "sync_end",
			"connection_failure"
        ],
        "active": false,
        "secret": "******",
        "created_at": "2022-04-29T10:45:00.000Z",
        "created_by": "_airworthy"
    }`

	webhookPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/webhooks/account").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			webhookData = tfmock.CreateMapFromJsonString(t, webhookResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Account webhook has been created", webhookData), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/webhooks/webhook_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", webhookData), nil
		},
	)

	webhookPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/webhooks/webhook_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			webhookData = tfmock.CreateMapFromJsonString(t, webhookUpdatedResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Webhook has been updated", webhookData), nil
		},
	)

	webhookDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/webhooks/webhook_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "Webhook with id 'webhook_id' has been deleted", nil), nil
		},
	)

	webhookTestHandler = tfmock.MockClient().When(http.MethodPost, "/v1/webhooks/webhook_id/test").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onTestWebhook(t, req)
		},
	)
}

func TestResourceWebhookMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_webhook" "test_webhook" {
                 provider = fivetran-provider

                 type = "account"
                 group_id = "_moonbeam"
                 url = "https://your-host.your-domain/webhook"
                 secret = "password"
                 active = false
                 run_tests = false
                 events = ["sync_start","sync_end"]
            }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, webhookPostHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "type", "account"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "group_id", "_moonbeam"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "url", "https://your-host.your-domain/webhook"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "secret", "password"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "active", "false"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.0", "sync_end"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.1", "sync_start"),
		),
	}

	step2 := resource.TestStep{
		Config: `
            resource "fivetran_webhook" "test_webhook" {
                 provider = fivetran-provider

                 type = "account"
                 group_id = "_moonbeam"
                 url = "https://your-host.your-domain/webhook_1"
                 secret = "password_2"
                 active = false
                 run_tests = false
                 events = ["sync_start","sync_end", "connection_failure"]
            }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, webhookPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "type", "account"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "group_id", "_moonbeam"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "url", "https://your-host.your-domain/webhook_1"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "secret", "password_2"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "active", "false"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.0", "connection_failure"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.1", "sync_end"),
			resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.2", "sync_start"),
		),
	}

	step3 := resource.TestStep{
		Config: `
            resource "fivetran_webhook" "test_webhook" {
                 provider = fivetran-provider

                 type = "account"
                 group_id = "_moonbeam"
                 url = "https://your-host.your-domain/webhook_1"
                 secret = "password_2"
                 active = false
                 run_tests = true
                 events = ["sync_start","sync_end", "connection_failure"]
            }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, webhookTestHandler.Interactions, 3)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientWebhookResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, webhookDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
			},
		},
	)
}
