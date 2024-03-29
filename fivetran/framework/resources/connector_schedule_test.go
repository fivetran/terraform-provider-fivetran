package resources_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceConnectorScheduleMock(t *testing.T) {
	var patchHandler *mock.Handler

	scheduleState := map[string]interface{}{
		"daily_sync_time":   nil,
		"schedule_type":     "auto",
		"paused":            true,
		"pause_after_trial": true,
		"sync_frequency":    float64(360),
	}

	getOrPanic := func(ss map[string]interface{}, key string) interface{} {
		value, ok := ss[key]
		if !ok {
			panic(fmt.Sprintf("Test map value %v is missing", key))
		}
		return value
	}

	createResponse := func(ss map[string]interface{}) string {
		syncFrequency := getOrPanic(ss, "sync_frequency").(float64)
		dailySyncTime := getOrPanic(ss, "daily_sync_time")
		paused := getOrPanic(ss, "paused").(bool)
		pauseAfterTrial := getOrPanic(ss, "pause_after_trial").(bool)
		scheduletype := getOrPanic(ss, "schedule_type").(string)
		connectorResponseTemplate := `
		{
			"id": "connector_id",
			"group_id": "group_id",
			"service": "service_type",
			"service_version": 0,
			"schema": "schema_name",
			"connected_by": "user_id",
			"created_at": "2020-03-11T15:03:55.743708Z",
			"succeeded_at": "2020-03-17T12:31:40.870504Z",
			"failed_at": "2021-01-15T10:55:00.056497Z",
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
				"password": "******"
			},
			%v 
			"data_delay_sensitivity": "NORMAL",
			"data_delay_threshold": 0,
			"paused": %v,
			"pause_after_trial": %v,
			"sync_frequency": %v,
			"schedule_type": "%v"
		}
		`
		dailySyncTimeElem := ""
		if syncFrequency == float64(1440) && dailySyncTime != nil {
			dailySyncTimeElem = fmt.Sprintf(`"daily_sync_time": "%v",`, dailySyncTime)
		}
		return fmt.Sprintf(
			connectorResponseTemplate,
			dailySyncTimeElem,
			paused,
			pauseAfterTrial,
			syncFrequency,
			scheduletype,
		)
	}

	var responseData map[string]interface{}

	preCheckFunc := func() {
		tfmock.MockClient().Reset()
		tfmock.MockClient().When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				responseData = tfmock.CreateMapFromJsonString(t, createResponse(scheduleState))
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
			},
		)

		patchHandler =
			tfmock.MockClient().When(http.MethodPatch, "/v1/connectors/connector_id").ThenCall(
				func(req *http.Request) (*http.Response, error) {
					body := tfmock.RequestBodyToJson(t, req)

					for k, v := range body {
						if _, ok := scheduleState[k]; ok {
							scheduleState[k] = v
						}
					}

					responseString := createResponse(scheduleState)
					responseData = tfmock.CreateMapFromJsonString(t, responseString)
					return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", responseData), nil
				},
			)
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck:                 preCheckFunc,
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				{
					Config: `
					resource "fivetran_connector_schedule" "test_connector_schedule" {
						provider = fivetran-provider
						connector_id = "connector_id"
						sync_frequency = 360
						schedule_type = "auto"
						daily_sync_time = "12:00"
						paused = true
						pause_after_trial = true
					}`,

					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, patchHandler.Interactions, 1)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "id", "connector_id"),
					),
				},
				{
					Config: `
					resource "fivetran_connector_schedule" "test_connector_schedule" {
						provider = fivetran-provider
						connector_id = "connector_id"
						sync_frequency = 1440
						schedule_type = "auto"
						daily_sync_time = "12:00"
						paused = true
						pause_after_trial = true
					}`,

					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, patchHandler.Interactions, 2)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "id", "connector_id"),
					),
				},
				{
					Config: `
					resource "fivetran_connector_schedule" "test_connector_schedule" {
						provider = fivetran-provider
						connector_id = "connector_id"
						sync_frequency = 60
						schedule_type = "auto"
						daily_sync_time = "12:00"
						paused = true
						pause_after_trial = true
					}`,

					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, patchHandler.Interactions, 3)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "id", "connector_id"),
					),
				},
				{
					Config: `
					resource "fivetran_connector_schedule" "test_connector_schedule" {
						provider = fivetran-provider
						connector_id = "connector_id"
						sync_frequency = 60
						schedule_type = "auto"
						paused = true
						pause_after_trial = true
					}`,

					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, patchHandler.Interactions, 3)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "id", "connector_id"),
					),
				},
				{
					Config: `
					resource "fivetran_connector_schedule" "test_connector_schedule" {
						provider = fivetran-provider
						connector_id = "connector_id"
						sync_frequency = 1440
						schedule_type = "auto"
						daily_sync_time = "15:00"
						paused = true
						pause_after_trial = true
					}`,

					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, patchHandler.Interactions, 4)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schedule.test_connector_schedule", "id", "connector_id"),
					),
				},
			},
		},
	)
}
