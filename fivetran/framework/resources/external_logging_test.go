package resources_test

import (
	"net/http"
	"testing"
	"time"
	
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	externalLoggingPostHandler   		*mock.Handler
	externalLoggingPatchHandler  		*mock.Handler
	externalLoggingTestHandler   		*mock.Handler
	externalLoggingDeleteHandler 		*mock.Handler
	testExternalLoggingData      		map[string]interface{}
	testExternalLoggingPatchedData      map[string]interface{}

	externalLoggingMappingGetHandler    *mock.Handler
	externalLoggingMappingPostHandler   *mock.Handler
	externalLoggingMappingPatchHandler  *mock.Handler
	externalLoggingMappingDeleteHandler *mock.Handler
)

const (
	externalLoggingMappingResponse = `
    {
        "id":                 "log_id",
        "group_id":           "log_id",
        "service":            "azure_monitor_log",
        "enabled":            false,
        "config":{
            "workspace_id":   "workspace_id",
            "primary_key":    "******",
            "role_arn":       "role_arn",
            "region":         "region",
            "log_group_name": "log_group_name",
            "sub_domain":     "sub_domain",
            "enable_ssl":     true,
            "channel":        "channel",
            "token":          "******",
            "external_id":    "external_id",
            "api_key":        "******",
            "host":           "host",
            "hostname":       "hostname",
            "port":           443,
            "project_id":     "project_id"
        }
    }
    `

    externalLoggingMappingUpdatedResponse = `
    {
        "id":                 "log_id",
        "group_id":           "log_id",
        "service":            "azure_monitor_log",
        "enabled":            false,
        "config":{
            "workspace_id":   "workspace_id_1",
            "primary_key":    "******",
            "role_arn":       "role_arn",
            "region":         "region",
            "log_group_name": "log_group_name",
            "sub_domain":     "sub_domain",
            "enable_ssl":     true,
            "channel":        "channel",
            "token":          "******",
            "external_id":    "external_id",
            "api_key":        "******",
            "host":           "host",
            "hostname":       "hostname",
            "port":           443,
            "project_id":     "project_id"
        }
    }
    `
)

func setupMockClientExternalLoggingConfigMapping(t *testing.T) {
	var patchInteractions int
	patchInteractions = 0

	tfmock.MockClient().Reset()

	externalLoggingMappingGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/external-logging/log_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testExternalLoggingData), nil
		},
	)

	externalLoggingMappingPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/external-logging").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)

			tfmock.AssertKeyExists(t, body, "config")

			config := body["config"].(map[string]interface{})

			tfmock.AssertKeyExistsAndHasValue(t, config, "workspace_id", "workspace_id")
			tfmock.AssertKeyExistsAndHasValue(t, config, "primary_key", "primary_key")
			tfmock.AssertKeyExistsAndHasValue(t, config, "role_arn", "role_arn")
			tfmock.AssertKeyExistsAndHasValue(t, config, "region", "region")
			tfmock.AssertKeyExistsAndHasValue(t, config, "port", float64(443))
			tfmock.AssertKeyExistsAndHasValue(t, config, "log_group_name", "log_group_name")
			tfmock.AssertKeyExistsAndHasValue(t, config, "sub_domain", "sub_domain")
			tfmock.AssertKeyExistsAndHasValue(t, config, "enable_ssl", true)
			tfmock.AssertKeyExistsAndHasValue(t, config, "channel", "channel")
			tfmock.AssertKeyExistsAndHasValue(t, config, "token", "token")
			tfmock.AssertKeyExistsAndHasValue(t, config, "external_id", "external_id")
			tfmock.AssertKeyExistsAndHasValue(t, config, "api_key", "api_key")
			tfmock.AssertKeyExistsAndHasValue(t, config, "host", "host")
			tfmock.AssertKeyExistsAndHasValue(t, config, "hostname", "hostname")
			tfmock.AssertKeyExistsAndHasValue(t, config, "project_id", "project_id")

			testExternalLoggingData = tfmock.CreateMapFromJsonString(t, externalLoggingMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", testExternalLoggingData), nil
		},
	)

	externalLoggingMappingPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/external-logging/log_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			patchInteractions += 1
			body := tfmock.RequestBodyToJson(t, req)

			tfmock.AssertKeyExists(t, body, "config")

			config := body["config"].(map[string]interface{})

			tfmock.AssertKeyExistsAndHasValue(t, config, "workspace_id", "workspace_id_1")
			tfmock.AssertKeyExistsAndHasValue(t, config, "role_arn", "role_arn")
			tfmock.AssertKeyExistsAndHasValue(t, config, "region", "region")
			tfmock.AssertKeyExistsAndHasValue(t, config, "port", float64(443))
			tfmock.AssertKeyExistsAndHasValue(t, config, "log_group_name", "log_group_name")
			tfmock.AssertKeyExistsAndHasValue(t, config, "sub_domain", "sub_domain")
			tfmock.AssertKeyExistsAndHasValue(t, config, "enable_ssl", true)
			tfmock.AssertKeyExistsAndHasValue(t, config, "channel", "channel")
			tfmock.AssertKeyExistsAndHasValue(t, config, "external_id", "external_id")
			tfmock.AssertKeyExistsAndHasValue(t, config, "host", "host")
			tfmock.AssertKeyExistsAndHasValue(t, config, "hostname", "hostname")
			tfmock.AssertKeyExistsAndHasValue(t, config, "project_id", "project_id")

			if(patchInteractions < 2) {
				tfmock.AssertKeyDoesNotExist(t, config, "primary_key")
				tfmock.AssertKeyDoesNotExist(t, config, "api_key")
				tfmock.AssertKeyDoesNotExist(t, config, "token")
			} else {
				tfmock.AssertKeyExistsAndHasValue(t, config, "primary_key", "primary_key_2")
				tfmock.AssertKeyExistsAndHasValue(t, config, "api_key", "api_key_2")
				tfmock.AssertKeyExistsAndHasValue(t, config, "token", "token_2")
			}

			testExternalLoggingData = tfmock.CreateMapFromJsonString(t, externalLoggingMappingUpdatedResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", testExternalLoggingData), nil
		},
	)

	externalLoggingMappingDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/external-logging/log_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			testExternalLoggingData = nil
			response := tfmock.FivetranSuccessResponse(t, req, 200,
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

	            group_id = "log_id"
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
                    project_id = "project_id"
                }
            }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, externalLoggingMappingPostHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, testExternalLoggingData)
				return nil
			},
		),
	}

	// remove from config to test import verify in the next test step
	step2 := resource.TestStep{
		Config: `
		`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, externalLoggingMappingDeleteHandler.Interactions, 1)
				return nil
			},
		),
	}

	// import verify
	step3 := resource.TestStep{
		PreConfig: func() {
			testExternalLoggingData = tfmock.CreateMapFromJsonString(t, externalLoggingMappingResponse)
		},
		Config: `
				resource "fivetran_external_logging" "test_extlog" {
					provider = fivetran-provider
				}
			`,
		ResourceName:      "fivetran_external_logging.test_extlog",
		ImportState:       true,
		ImportStateId: 	"log_id",
		ImportStatePersist: true,
		ImportStateCheck: tfmock.ComposeImportStateCheck(
			tfmock.CheckImportResourceAttr("fivetran_external_logging", "log_id", "id", "log_id"),
			tfmock.CheckImportResourceAttr("fivetran_external_logging", "log_id", "group_id", "log_id"),
			tfmock.CheckImportResourceAttr("fivetran_external_logging", "log_id", "service", "azure_monitor_log"),
			tfmock.CheckImportResourceAttr("fivetran_external_logging", "log_id", "config.role_arn", "role_arn"),
			tfmock.CheckNoImportResourceAttr("fivetran_external_logging", "log_id", "config.api_key"),
		),
	}

	// sensitive fields being null after import do not trigger an update
	step4 := resource.TestStep{
		Config: `
			resource "fivetran_external_logging" "test_extlog" {
                provider = fivetran-provider

	            group_id = "log_id"
                service = "azure_monitor_log"
                enabled = "false"

                config {
                    workspace_id = "workspace_id"
                    role_arn = "role_arn"
                    region = "region"
                    log_group_name = "log_group_name"
                    sub_domain = "sub_domain"
                    enable_ssl = true
                    channel = "channel"
                    external_id = "external_id"
                    host = "host"
                    hostname = "hostname"
                    port = 443
                    project_id = "project_id"
                }
            }`,
		PlanOnly: true,
	}

	// updating non-sensitive fields after importing is working
	step5 := resource.TestStep{
		Config: `
			resource "fivetran_external_logging" "test_extlog" {
                provider = fivetran-provider

	            group_id = "log_id"
                service = "azure_monitor_log"
                enabled = "false"

                config {
                    workspace_id = "workspace_id_1"
                    role_arn = "role_arn"
                    region = "region"
                    log_group_name = "log_group_name"
                    sub_domain = "sub_domain"
                    enable_ssl = true
                    channel = "channel"
                    external_id = "external_id"
                    host = "host"
                    hostname = "hostname"
                    port = 443
                    project_id = "project_id"
                }
            }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, externalLoggingMappingPatchHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, testExternalLoggingData)
				return nil
			},			
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.workspace_id", "workspace_id_1"),
			resource.TestCheckNoResourceAttr("fivetran_external_logging.test_extlog", "config.api_key"),
			resource.TestCheckNoResourceAttr("fivetran_external_logging.test_extlog", "config.token"),
			resource.TestCheckNoResourceAttr("fivetran_external_logging.test_extlog", "config.primary_key"),
		),
	}

	// updating sensitive fields after importing is working
	step6 := resource.TestStep{
		Config: `
			resource "fivetran_external_logging" "test_extlog" {
                provider = fivetran-provider

	            group_id = "log_id"
                service = "azure_monitor_log"
                enabled = "false"

                config {
					api_key = "api_key_2"
					token = "token_2"
					primary_key = "primary_key_2"
                    workspace_id = "workspace_id_1"
                    role_arn = "role_arn"
                    region = "region"
                    log_group_name = "log_group_name"
                    sub_domain = "sub_domain"
                    enable_ssl = true
                    channel = "channel"
                    external_id = "external_id"
                    host = "host"
                    hostname = "hostname"
                    port = 443
                    project_id = "project_id"
                }
            }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, externalLoggingMappingPostHandler.Interactions, 1)
				tfmock.AssertEqual(t, externalLoggingMappingPatchHandler.Interactions, 2)
				tfmock.AssertNotEmpty(t, testExternalLoggingData)
				return nil
			},			
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.workspace_id", "workspace_id_1"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.api_key", "api_key_2"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.token", "token_2"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.primary_key", "primary_key_2"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientExternalLoggingConfigMapping(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, externalLoggingMappingPostHandler.Interactions, 1)
				tfmock.AssertEqual(t, externalLoggingMappingPatchHandler.Interactions, 2)
				tfmock.AssertEqual(t, externalLoggingMappingDeleteHandler.Interactions, 2)
				tfmock.AssertEmpty(t, testExternalLoggingData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
				step4,
				step5,
				step6,
			},
		},
	)
}

func onPostExternalLogging(t *testing.T, req *http.Request) (*http.Response, error) {
	tfmock.AssertEmpty(t, testExternalLoggingData)

	body := tfmock.RequestBodyToJson(t, req)

	// Add response fields
	body["id"] = "log_id"
	body["created_at"] = time.Now().Format("2006-01-02T15:04:05.000000Z")

	testExternalLoggingData = body

	response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated,
		"External logging service has been added", body)

	return response, nil
}

func onPatchExternalLogging(t *testing.T, req *http.Request) (*http.Response, error) {
	testExternalLoggingPatchedData = tfmock.CreateMapFromJsonString(t, externalLoggingMappingUpdatedResponse)

	tfmock.AssertNotEmpty(t, testExternalLoggingPatchedData)

	body := tfmock.RequestBodyToJson(t, req)

	// Update saved values
	updateMapDeep(body, testExternalLoggingPatchedData)

	response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "External logging service has been updated", testExternalLoggingPatchedData)

	testExternalLoggingData = testExternalLoggingPatchedData
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

	response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Setup tests have been completed", testExternalLoggingData)
	return response, nil
}

func setupMockClientForExternalLogging(t *testing.T) {
	tfmock.MockClient().Reset()
	testExternalLoggingData = nil

	externalLoggingPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/external-logging").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPostExternalLogging(t, req)
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/external-logging/log_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			tfmock.AssertNotEmpty(t, testExternalLoggingData)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", testExternalLoggingData)
			return response, nil
		},
	)

	externalLoggingPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/external-logging/log_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPatchExternalLogging(t, req)
		},
	)

	externalLoggingTestHandler = tfmock.MockClient().When(http.MethodPost, "/v1/external-logging/log_id/test").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onTestExternalLogging(t, req)
		},
	)

	externalLoggingDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/external-logging/log_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			tfmock.AssertNotEmpty(t, testExternalLoggingData)
			testExternalLoggingData = nil
			response := tfmock.FivetranSuccessResponse(t, req, 200,
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

                group_id = "log_id"
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
                    project_id = "project_id"
                }
            }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, externalLoggingPostHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, testExternalLoggingData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "service", "azure_monitor_log"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "group_id", "log_id"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "enabled", "false"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "run_setup_tests", "false"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.workspace_id", "workspace_id"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.primary_key", "primary_key"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.log_group_name", "log_group_name"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.sub_domain", "sub_domain"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.enable_ssl", "true"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.port", "443"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.role_arn", "role_arn"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.channel", "channel"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.token", "token"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.region", "region"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.external_id", "external_id"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.api_key", "api_key"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.host", "host"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.hostname", "hostname"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.project_id", "project_id"),
		),
	}

	step2 := resource.TestStep{
		Config: `
            resource "fivetran_external_logging" "test_extlog" {
                provider = fivetran-provider

                group_id = "log_id"
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
                    project_id = "project_id"
                }
            }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, externalLoggingPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "service", "azure_monitor_log"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "group_id", "log_id"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "enabled", "false"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "run_setup_tests", "false"),

			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.workspace_id", "workspace_id_1"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.primary_key", "primary_key"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.log_group_name", "log_group_name"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.sub_domain", "sub_domain"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.enable_ssl", "true"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.port", "443"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.role_arn", "role_arn"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.channel", "channel"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.token", "token"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.region", "region"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.external_id", "external_id"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.api_key", "api_key"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.host", "host"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.hostname", "hostname"),
			resource.TestCheckResourceAttr("fivetran_external_logging.test_extlog", "config.project_id", "project_id"),
		),
	}

	step3 := resource.TestStep{
		Config: `
            resource "fivetran_external_logging" "test_extlog" {
                provider = fivetran-provider

                group_id = "log_id"
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
                    project_id = "project_id"
                }
            }`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, externalLoggingPatchHandler.Interactions, 1)
				tfmock.AssertEqual(t, externalLoggingTestHandler.Interactions, 1)
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
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, externalLoggingDeleteHandler.Interactions, 1)
				tfmock.AssertEmpty(t, testExternalLoggingData)
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

func updateMapDeep(source map[string]interface{}, target map[string]interface{}) {
	for sk, sv := range source {
		if tv, ok := target[sk]; ok {
			if svmap, ok := sv.(map[string]interface{}); ok {
				if tvmap, ok := tv.(map[string]interface{}); ok {
					updateMapDeep(svmap, tvmap)
					continue
				}
			}
		}
		target[sk] = sv
	}
}
