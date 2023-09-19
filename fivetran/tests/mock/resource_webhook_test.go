package mock

import (
    "net/http"
    "testing"

    "github.com/fivetran/go-fivetran/tests/mock"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

    response := fivetranSuccessResponse(t, req, http.StatusOK, "Setup tests have been completed", webhookData)
    return response, nil
}

func setupMockClientWebhookResource(t *testing.T) {
    mockClient.Reset()
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
        "secret": "password",
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
            "sync_end"
        ],
        "active": false,
        "secret": "password",
        "created_at": "2022-04-29T10:45:00.000Z",
        "created_by": "_airworthy"
    }`

    webhookPostHandler = mockClient.When(http.MethodPost, "/v1/webhooks/account").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            webhookData = createMapFromJsonString(t, webhookResponse)
            return fivetranSuccessResponse(t, req, http.StatusOK, "Account webhook has been created", webhookData), nil
        },
    )

    mockClient.When(http.MethodGet, "/v1/webhooks/webhook_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return fivetranSuccessResponse(t, req, http.StatusOK, "", webhookData), nil
        },
    )

    webhookPatchHandler = mockClient.When(http.MethodPatch, "/v1/webhooks/webhook_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            webhookData = createMapFromJsonString(t, webhookUpdatedResponse)
            return fivetranSuccessResponse(t, req, http.StatusOK, "Webhook has been updated", webhookData), nil
        },
    )

    webhookDeleteHandler = mockClient.When(http.MethodDelete, "/v1/webhooks/webhook_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return fivetranSuccessResponse(t, req, 200, "Webhook with id 'webhook_id' has been deleted", nil), nil
        },
    )

    webhookTestHandler = mockClient.When(http.MethodPost, "/v1/webhooks/webhook_id/test").ThenCall(
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
                assertEqual(t, webhookPostHandler.Interactions, 1)
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
                 secret = "password"
                 active = false
                 run_tests = false
                 events = ["sync_start","sync_end"]
            }`,
        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, webhookPatchHandler.Interactions, 1)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "type", "account"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "group_id", "_moonbeam"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "url", "https://your-host.your-domain/webhook_1"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "secret", "password"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "active", "false"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.0", "sync_end"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.1", "sync_start"),
        ),
    }

    step3 := resource.TestStep{
        Config: `
            resource "fivetran_webhook" "test_webhook" {
                 provider = fivetran-provider

                 type = "account"
                 group_id = "_moonbeam"
                 url = "https://your-host.your-domain/webhook_1"
                 secret = "password"
                 active = false
                 run_tests = true
                 events = ["sync_start","sync_end"]
            }`,
        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, webhookTestHandler.Interactions, 2)
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
            Providers: testProviders,
            CheckDestroy: func(s *terraform.State) error {
                assertEqual(t, webhookDeleteHandler.Interactions, 1)
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
