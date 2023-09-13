package mock

import (
    "net/http"
    "testing"
    "time"

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

func onPostWebhooks(t *testing.T, req *http.Request) (*http.Response, error) {
    assertEmpty(t, userData)

    body := requestBodyToJson(t, req)

    // Check the request
    assertEqual(t, len(body), 6)
    assertEqual(t, body["email"], "john.fox@testmail.com")
    assertEqual(t, body["given_name"], "John")
    assertEqual(t, body["family_name"], "Fox")
    assertEqual(t, body["phone"], "+19876543210")
    assertEqual(t, body["picture"], "https://myPicturecom")
    assertEqual(t, body["role"], "Account Reviewer")

    // Add response fields
    body["id"] = "john_fox_id"
    body["verified"] = false
    body["invited"] = true
    body["logged_in_at"] = nil
    body["created_at"] = time.Now().Format("2006-01-02T15:04:05.000000Z")
    userData = body

    response := fivetranSuccessResponse(t, req, http.StatusCreated, "Account webhook has been created", body)

    return response, nil
}

func onPatchWebhook(t *testing.T, req *http.Request, updateIteration int) (*http.Response, error) {
    assertNotEmpty(t, userData)

    body := requestBodyToJson(t, req)

    if updateIteration == 0 {
        // Check the request
        assertEqual(t, len(body), 5)
        assertEqual(t, body["given_name"], "Jane")
        assertEqual(t, body["family_name"], "Connor")
        assertEqual(t, body["phone"], "+19876543219")
        assertEqual(t, body["picture"], "https://yourPicturecom")
        assertEqual(t, body["role"], "Account Administrator")

        // Update saved values
        for k, v := range body {
            webhookData[k] = v
        }

        response := fivetranSuccessResponse(t, req, http.StatusOK, "Webhook has been updated", userData)
        return response, nil
    }

    if updateIteration == 1 {
        // Check the request
        assertEqual(t, len(body), 2)
        assertEqual(t, body["phone"], nil)
        assertEqual(t, body["picture"], nil)

        // Update saved values
        for k, v := range body {
            webhookData[k] = v
        }

        response := fivetranSuccessResponse(t, req, http.StatusOK, "Webhook has been updated", userData)
        return response, nil
    }

    return nil, nil
}

func onTestWebhook(t *testing.T, req *http.Request) (*http.Response, error) {
    // setup test results array
    setupTests := make([]interface{}, 0)

    setupTestResult := make(map[string]interface{})
    setupTestResult["title"] = "Test Title"
    setupTestResult["status"] = "PASSED"
    setupTestResult["message"] = "Test passed"

    setupTests = append(setupTests, setupTestResult)

    testWebhookData["setup_tests"] = setupTests

    response := fivetranSuccessResponse(t, req, http.StatusOK, "Setup tests have been completed", testWebhookData)
    return response, nil
}

func setupMockClientWebhookResource(t *testing.T) {
    mockClient.Reset()
    webhookData = nil
    updateCounter := 0

    webhookPostHandler = mockClient.When(http.MethodPost, "/v1/webhooks/account").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return onPostWebhooks(t, req)
        },
    )

    mockClient.When(http.MethodGet, "/v1/webhooks/webhook_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            assertNotEmpty(t, webhookData)
            response := fivetranSuccessResponse(t, req, http.StatusOK, "", webhookData)
            return response, nil
        },
    )

    webhookPatchHandler = mockClient.When(http.MethodPatch, "/v1/webhooks/webhook_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            response, err := onPatchWebhook(t, req, updateCounter)
            updateCounter++
            return response, err
        },
    )

    webhookDeleteHandler = mockClient.When(http.MethodDelete, "/v1/webhooks/webhook_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            assertNotEmpty(t, webhookData)
            webhookData = nil
            response := fivetranSuccessResponse(t, req, 200,
                "Webhook with id 'webhook_id' has been deleted", nil)
            return response, nil
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

                 id = "recur_readable"
                 type = "group"
                 group_id = "_moonbeam"
                 url = "https://your-host.your-domain/webhook"
                 secret = "password"
                 active = false
                 run_setup_tests = false
                 created_at = "2022-04-29T10:45:00.000Z"
                 created_by = "_airworthy"
                 events : ["sync_start","sync_end"]
            }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, webhookPostHandler.Interactions, 1)
                assertNotEmpty(t, webhookData)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "id", "recur_readable"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "type", "group"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "group_id", "_moonbeam"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "url", "https://your-host.your-domain/webhook"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "secret", "password"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "active", "false"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "created_at", "2022-04-29T10:45:00.000Z"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "created_by", "_airworthy"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.0", "sync_start"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.1", "sync_end"),
        ),
    }

    step2 := resource.TestStep{
        Config: `
            resource "fivetran_webhook" "test_webhook" {
                 provider = fivetran-provider

                 id = "recur_readable"
                 type = "group"
                 group_id = "_moonbeam"
                 url = "https://your-host.your-domain/webhook_1"
                 secret = "password"
                 active = false
                 run_setup_tests = false
                 created_at = "2022-04-29T10:45:00.000Z"
                 created_by = "_airworthy"
                 events : ["sync_start","sync_end"]
            }`,
        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, webhookPatchHandler.Interactions, 1)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "id", "recur_readable"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "type", "group"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "group_id", "_moonbeam"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "url", "https://your-host.your-domain/webhook_1"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "secret", "password"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "active", "false"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "created_at", "2022-04-29T10:45:00.000Z"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "created_by", "_airworthy"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.0", "sync_start"),
            resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.1", "sync_end"),
        ),
    }

    step3 := resource.TestStep{
        Config: `
            resource "fivetran_webhook" "test_webhook" {
                 provider = fivetran-provider

                 id = "recur_readable"
                 type = "group"
                 group_id = "_moonbeam"
                 url = "https://your-host.your-domain/webhook_1"
                 secret = "password"
                 active = false
                 run_setup_tests = true
                 created_at = "2022-04-29T10:45:00.000Z"
                 created_by = "_airworthy"
                 events : ["sync_start","sync_end"]
            }`,
        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, webhookTestHandler.Interactions, 1)
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
                assertEmpty(t, webhookData)
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
