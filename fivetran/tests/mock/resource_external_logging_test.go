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
    externalLoggingPostHandler   *mock.Handler
    externalLoggingPatchHandler  *mock.Handler
    externalLoggingTestHandler   *mock.Handler
    externalLoggingDeleteHandler *mock.Handler
    testExternalLoggingData      map[string]interface{}

    externalLoggingMappingGetHandler    *mock.Handler
    externalLoggingMappingPostHandler   *mock.Handler
    externalLoggingMappingDeleteHandler *mock.Handler
)

const (
    externalLoggingMappingResponse = `
    {
        "id":                 "log_id",
        "group_id":           "group_id",
        "service":            "azure_monitor_log",
        "enabled":            false,
        "config":{
            "workspace_id":   "workspace_id",
            "primary_key":    "primary_key",
            "role_arn":       "role_arn",
            "region":         "region",
            "log_group_name": "log_group_name",
            "sub_domain":     "sub_domain",
            "enable_ssl":     true,
            "channel":        "channel",
            "token":          "token",
            "external_id":    "external_id",
            "api_key":        "api_key",
            "host":           "host",
            "hostname":       "hostname",
            "port":           443
        }
    }
    `
)

func setupMockClientExternalLoggingConfigMapping(t *testing.T) {
    mockClient.Reset()

    externalLoggingMappingGetHandler = mockClient.When(http.MethodGet, "/v1/external-logging/log_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return fivetranSuccessResponse(t, req, http.StatusOK, "Success", testExternalLoggingData), nil
        },
    )

    externalLoggingMappingPostHandler = mockClient.When(http.MethodPost, "/v1/external-logging").ThenCall(
        func(req *http.Request) (*http.Response, error) {

            body := requestBodyToJson(t, req)

            assertKeyExists(t, body, "config")

            config := body["config"].(map[string]interface{})

            assertKeyExistsAndHasValue(t, config, "workspace_id", "workspace_id")
            assertKeyExistsAndHasValue(t, config, "primary_key", "primary_key")
            assertKeyExistsAndHasValue(t, config, "role_arn", "role_arn")
            assertKeyExistsAndHasValue(t, config, "region", "region")
            assertKeyExistsAndHasValue(t, config, "port", float64(443))
            assertKeyExistsAndHasValue(t, config, "log_group_name", "log_group_name")
            assertKeyExistsAndHasValue(t, config, "sub_domain", "sub_domain")
            assertKeyExistsAndHasValue(t, config, "enable_ssl", true)
            assertKeyExistsAndHasValue(t, config, "channel", "channel")
            assertKeyExistsAndHasValue(t, config, "token", "token")
            assertKeyExistsAndHasValue(t, config, "external_id", "external_id")
            assertKeyExistsAndHasValue(t, config, "api_key", "api_key")
            assertKeyExistsAndHasValue(t, config, "host", "host")
            assertKeyExistsAndHasValue(t, config, "hostname", "hostname")

            testExternalLoggingData = createMapFromJsonString(t, externalLoggingMappingResponse)
            return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", testExternalLoggingData), nil
        },
    )

    externalLoggingMappingDeleteHandler = mockClient.When(http.MethodDelete, "/v1/external-logging/log_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            testExternalLoggingData = nil
            response := fivetranSuccessResponse(t, req, 200,
                "External logging service with id 'log_id' has been deleted", nil)
            return response, nil
        },
    )
}

func TestResourceExternalLoggingMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
            resource "fivetran_external_logging" "test_extlog" {
                provider = fivetran-provider

                group_id = "group_id"
                service = "azure_monitor_log"
                enabled = "false"
                run_setup_tests = "false"

                config {
                    workspace_id = "workspace_id"
                    primary_key = "primary_key"
                    role_arn = "role_arn"
                    region = "region"
                    log_group_name = "log_group_name"
                    sub_domain = "sub_domain"
                    enable_ssl = true
                    channel = "channel"
                    token = "token"
                    external_id = "external_id"
                    api_key = "api_key"
                    host = "host"
                    hostname = "hostname"
                    port = 443
                }
            }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, externalLoggingMappingGetHandler.Interactions, 1)
                assertEqual(t, externalLoggingMappingPostHandler.Interactions, 1)
                assertNotEmpty(t, testExternalLoggingData)
                return nil
            },
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientExternalLoggingConfigMapping(t)
            },
            Providers: testProviders,
            CheckDestroy: func(s *terraform.State) error {
                assertEqual(t, externalLoggingMappingDeleteHandler.Interactions, 1)
                assertEmpty(t, testExternalLoggingData)
                return nil
            },

            Steps: []resource.TestStep{
                step1,
            },
        },
    )
}

func onPostExternalLogging(t *testing.T, req *http.Request) (*http.Response, error) {
    assertEmpty(t, testExternalLoggingData)

    body := requestBodyToJson(t, req)

    // Add response fields
    body["id"] = "log_id"
    body["created_at"] = time.Now().Format("2006-01-02T15:04:05.000000Z")

    testExternalLoggingData = body

    response := fivetranSuccessResponse(t, req, http.StatusCreated,
        "External logging service has been added", body)

    return response, nil
}

func onPatchExternalLogging(t *testing.T, req *http.Request) (*http.Response, error) {
    assertNotEmpty(t, testExternalLoggingData)

    body := requestBodyToJson(t, req)

    // Update saved values
    updateMapDeep(body, testExternalLoggingData)

    response := fivetranSuccessResponse(t, req, http.StatusOK, "External logging service has been updated", testExternalLoggingData)

    return response, nil
}

func onTestExternalLogging(t *testing.T, req *http.Request) (*http.Response, error) {
    // setup test results array
    setupTests := make([]interface{}, 0)

    setupTestResult := make(map[string]interface{})
    setupTestResult["title"] = "Test Title"
    setupTestResult["status"] = "PASSED"
    setupTestResult["message"] = "Test passed"

    setupTests = append(setupTests, setupTestResult)

    testExternalLoggingData["setup_tests"] = setupTests

    response := fivetranSuccessResponse(t, req, http.StatusOK, "Setup tests have been completed", testExternalLoggingData)
    return response, nil
}

func setupMockClientForExternalLogging(t *testing.T) {
    mockClient.Reset()
    testExternalLoggingData = nil

    externalLoggingPostHandler = mockClient.When(http.MethodPost, "/v1/external-logging").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return onPostExternalLogging(t, req)
        },
    )

    mockClient.When(http.MethodGet, "/v1/external-logging/log_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            assertNotEmpty(t, testExternalLoggingData)
            response := fivetranSuccessResponse(t, req, http.StatusOK, "", testExternalLoggingData)
            return response, nil
        },
    )

    externalLoggingPatchHandler = mockClient.When(http.MethodPatch, "/v1/external-logging/log_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return onPatchExternalLogging(t, req)
        },
    )

    externalLoggingTestHandler = mockClient.When(http.MethodPost, "/v1/external-logging/log_id/test").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            return onTestExternalLogging(t, req)
        },
    )

    externalLoggingDeleteHandler = mockClient.When(http.MethodDelete, "/v1/external-logging/log_id").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            assertNotEmpty(t, testExternalLoggingData)
            testExternalLoggingData = nil
            response := fivetranSuccessResponse(t, req, 200,
                "External logging service with id 'log_id' has been deleted", nil)
            return response, nil
        },
    )

}

func TestResourceExternalLoggingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
            resource "fivetran_external_logging" "test_extlog" {
                provider = fivetran-provider

                group_id = "group_id"
                service = "azure_monitor_log"
                enabled = "false"
                run_setup_tests = "false"

                config {
                    workspace_id = "workspace_id"
                    primary_key = "primary_key"
                    role_arn = "role_arn"
                    region = "region"
                    log_group_name = "log_group_name"
                    sub_domain = "sub_domain"
                    enable_ssl = true
                    channel = "channel"
                    token = "token"
                    external_id = "external_id"
                    api_key = "api_key"
                    host = "host"
                    hostname = "hostname"
                    port = 443
                }
            }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, externalLoggingPostHandler.Interactions, 1)
                assertNotEmpty(t, testExternalLoggingData)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "service", "azure_monitor_log"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "group_id", "group_id"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "enabled", "false"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "run_setup_tests", "false"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.workspace_id", "workspace_id"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.primary_key", "primary_key"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.log_group_name", "log_group_name"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.sub_domain", "sub_domain"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.enable_ssl", "true"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.port", "443"),           
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.role_arn", "role_arn"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.channel", "channel"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.token", "token"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.region", "region"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.external_id", "external_id"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.api_key", "api_key"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.host", "host"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.hostname", "hostname"),
        ),
    }

    step2 := resource.TestStep{
        Config: `
            resource "fivetran_external_logging" "test_extlog" {
                provider = fivetran-provider

                group_id = "group_id"
                service = "azure_monitor_log"
                enabled = "false"
                run_setup_tests = "false"

                config {
                    workspace_id = "workspace_id_1"
                    primary_key = "primary_key"
                    role_arn = "role_arn"
                    region = "region"
                    log_group_name = "log_group_name"
                    sub_domain = "sub_domain"
                    enable_ssl = true
                    channel = "channel"
                    token = "token"
                    external_id = "external_id"
                    api_key = "api_key"
                    host = "host"
                    hostname = "hostname"
                    port = 443
                }
            }`,
        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, externalLoggingPatchHandler.Interactions, 1)
                return nil
            },
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "service", "azure_monitor_log"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "group_id", "group_id"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "enabled", "false"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "run_setup_tests", "false"),

            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.workspace_id", "workspace_id_1"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.primary_key", "primary_key"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.log_group_name", "log_group_name"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.sub_domain", "sub_domain"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.enable_ssl", "true"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.port", "443"),           
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.role_arn", "role_arn"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.channel", "channel"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.token", "token"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.region", "region"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.external_id", "external_id"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.api_key", "api_key"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.host", "host"),
            resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.0.hostname", "hostname"),
        ),
    }

    step3 := resource.TestStep{
        Config: `
            resource "fivetran_external_logging" "test_extlog" {
                provider = fivetran-provider

                group_id = "group_id"
                service = "azure_monitor_log"
                enabled = "false"
                run_setup_tests = "true"

                config {
                    workspace_id = "workspace_id_1"
                    primary_key = "primary_key"
                    role_arn = "role_arn"
                    region = "region"
                    log_group_name = "log_group_name"
                    sub_domain = "sub_domain"
                    enable_ssl = true
                    channel = "channel"
                    token = "token"
                    external_id = "external_id"
                    api_key = "api_key"
                    host = "host"
                    hostname = "hostname"
                    port = 443
                }
            }`,
        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                assertEqual(t, externalLoggingPatchHandler.Interactions, 1)
                assertEqual(t, externalLoggingTestHandler.Interactions, 1)
                return nil
            },
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientForExternalLogging(t)
            },
            Providers: testProviders,
            CheckDestroy: func(s *terraform.State) error {
                assertEqual(t, externalLoggingDeleteHandler.Interactions, 1)
                assertEmpty(t, testExternalLoggingData)
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
