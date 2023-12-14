package resources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceConnectorMock(t *testing.T) {
	var postHandler *mock.Handler
	//var getHandler *mock.Handler
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		resource "fivetran_connector_new" "test_connector" {
			provider = fivetran-provider

			service = "google_ads"
			group_id = "group_id"

			destination_schema {
				name = "adwords_schema"
			}

			config {
				user = "user_name"
				password = "password"
				port = 5432
				#always_encrypted = true
				account_ids = ["id1", "id2", "id3"]

				reports {
					report_type = "report_1"
					metrics = ["metric1", "metric2"]
				}
				reports {
					report_type = "report_2"
					metrics = ["metric2", "metric3"]
				}
			}
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, postHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "id", "connector_id"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "name", "adwords_schema"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "group_id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "service", "google_ads"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "destination_schema.name", "adwords_schema"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.user", "user_name"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.password", "password"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.port", "5432"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.always_encrypted", "true"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.account_ids.0", "id1"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.account_ids.1", "id2"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.account_ids.2", "id3"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.reports.0.report_type", "report_1"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.reports.0.metrics.0", "metric1"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.reports.0.metrics.1", "metric2"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.reports.1.report_type", "report_2"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.reports.1.metrics.0", "metric2"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.reports.1.metrics.1", "metric3"),
		),
	}

	step2 := resource.TestStep{
		Config: `
		resource "fivetran_connector_new" "test_connector" {
			provider = fivetran-provider

			service = "google_ads"
			group_id = "group_id"

			destination_schema {
				name = "adwords_schema"
			}

			run_setup_tests = true
			trust_certificates = true
			trust_fingerprints = true

			config {
				user = "user_name_1"
				password = "password_1"
				port = 2345
				always_encrypted = false

				reports {
					report_type = "report_1"
					metrics = ["metric1", "metric2"]
				}
			}
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, postHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "id", "connector_id"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "name", "adwords_schema"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "group_id", "group_id"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "service", "google_ads"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "destination_schema.name", "adwords_schema"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.user", "user_name_1"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.password", "password_1"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.port", "2345"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.always_encrypted", "false"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.reports.0.report_type", "report_1"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.reports.0.metrics.0", "metric1"),
			resource.TestCheckResourceAttr("fivetran_connector_new.test_connector", "config.reports.0.metrics.1", "metric2"),
		),
	}

	var responseData map[string]interface{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()

				//getHandler =
				tfmock.MockClient().When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if responseData == nil {
							return tfmock.FivetranSuccessResponse(t, req, http.StatusNotFound, "NotFound", nil), nil
						}
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
					},
				)

				tfmock.MockClient().When(http.MethodDelete, "/v1/connectors/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
					},
				)

				postHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connectors").ThenCall(
					func(req *http.Request) (*http.Response, error) {

						body := tfmock.RequestBodyToJson(t, req)

						// Check the request
						tfmock.AssertKeyExistsAndHasValue(t, body, "service", "google_ads")
						tfmock.AssertKeyExistsAndHasValue(t, body, "group_id", "group_id")
						tfmock.AssertKeyExistsAndHasValue(t, body, "run_setup_tests", false)
						tfmock.AssertKeyExistsAndHasValue(t, body, "trust_certificates", false)
						tfmock.AssertKeyExistsAndHasValue(t, body, "trust_fingerprints", false)

						if config, ok := tfmock.AssertKeyExists(t, body, "config").(map[string]interface{}); ok {
							tfmock.AssertKeyExistsAndHasValue(t, config, "schema", "adwords_schema")
							tfmock.AssertKeyExistsAndHasValue(t, config, "user", "user_name")
							tfmock.AssertKeyExistsAndHasValue(t, config, "password", "password")
							tfmock.AssertKeyExistsAndHasValue(t, config, "port", float64(5432))
							if reports, ok := tfmock.AssertKeyExists(t, config, "reports").([]interface{}); ok {
								tfmock.AssertEqual(t, len(reports), 2)
							}
							if accountIds, ok := tfmock.AssertKeyExists(t, config, "account_ids").([]interface{}); ok {
								tfmock.AssertEqual(t, len(accountIds), 3)
							}
						}

						responseData = tfmock.CreateMapFromJsonString(t,
							`
						{
							"id": "connector_id",
							"group_id": "group_id",
							"service": "google_ads",
							"service_version": 4,
							"schema": "adwords_schema",
							"paused": true,
							"pause_after_trial": true,
							"connected_by": "monitoring_assuring",
							"created_at": "2020-03-11T15:03:55.743708Z",
							"succeeded_at": "2020-03-17T12:31:40.870504Z",
							"failed_at": "2021-01-15T10:55:00.056497Z",
							"sync_frequency": 360,
							"data_delay_sensitivity": "NORMAL",
							"data_delay_threshold": 0,
							"schedule_type": "auto",
							"status": {
								"setup_state": "incomplete",
								"schema_status": "ready",
								"sync_state": "scheduled",
								"update_state": "delayed",
								"is_historical_sync": false,
								"tasks": [
									{
										"code": "reconnect",
										"message": "Reconnect"
									}
								],
								"warnings": []
							},
							"config": {
								"user": "user_name",
								"password": "******",
								"port": 5432,
								"always_encrypted": true,
								"account_ids": ["id1", "id2", "id3"],
								"reports": [
									{
										"report_type": "report_1",
										"metrics": ["metric1", "metric2"]
									},
									{
										"report_type": "report_2",
										"metrics": ["metric2", "metric3"]
									}
								]
							}
						}
						`)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", responseData), nil
					},
				)

				tfmock.MockClient().When(http.MethodPatch, "/v1/connectors/connector_id").ThenCall(
					func(req *http.Request) (*http.Response, error) {

						body := tfmock.RequestBodyToJson(t, req)

						// Check the request
						tfmock.AssertEqual(t, len(body), 4)

						tfmock.AssertKeyExistsAndHasValue(t, body, "run_setup_tests", true)
						tfmock.AssertKeyExistsAndHasValue(t, body, "trust_certificates", true)
						tfmock.AssertKeyExistsAndHasValue(t, body, "trust_fingerprints", true)

						if config, ok := tfmock.AssertKeyExists(t, body, "config").(map[string]interface{}); ok {
							tfmock.AssertKeyExistsAndHasValue(t, config, "account_ids", nil)
							tfmock.AssertKeyExistsAndHasValue(t, config, "user", "user_name_1")
							tfmock.AssertKeyExistsAndHasValue(t, config, "password", "password_1")
							tfmock.AssertKeyExistsAndHasValue(t, config, "port", float64(2345))
							if reports, ok := tfmock.AssertKeyExists(t, config, "reports").([]interface{}); ok {
								tfmock.AssertEqual(t, len(reports), 1)
							}
						}

						responseData = tfmock.CreateMapFromJsonString(t,
							`
						{
							"id": "connector_id",
							"group_id": "group_id",
							"service": "google_ads",
							"service_version": 4,
							"schema": "adwords_schema",
							"paused": true,
							"pause_after_trial": true,
							"connected_by": "monitoring_assuring",
							"created_at": "2020-03-11T15:03:55.743708Z",
							"succeeded_at": "2020-03-17T12:31:40.870504Z",
							"failed_at": "2021-01-15T10:55:00.056497Z",
							"sync_frequency": 360,
							"data_delay_sensitivity": "NORMAL",
							"data_delay_threshold": 0,
							"schedule_type": "auto",
							"status": {
								"setup_state": "connected",
								"schema_status": "ready",
								"sync_state": "scheduled",
								"update_state": "delayed",
								"is_historical_sync": false,
								"tasks": [
									{
										"code": "reconnect",
										"message": "Reconnect"
									}
								],
								"warnings": []
							},
							"config": {
								"user": "user_name_1",
								"password": "******",
								"port": 2345,
								"always_encrypted": false,
								"reports": [
									{
										"report_type": "report_1",
										"metrics": ["metric1", "metric2"]
									}
								]
							}
						}
						`)
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}
