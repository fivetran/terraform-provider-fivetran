package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	webhooksDataSourceMockGetHandler *mock.Handler
	webhooksDataSourceMockData       map[string]interface{}
)

const (
	webhooksMappingResponse = `
	{
        "items": [
		{
        	"id": "program_quoth",
        	"type": "account",
        	"url": "https://your-host.your-domain/webhook",
        	"events": [
          		"sync_start",
          		"sync_end"
        	],
        	"active": true,
        	"secret": "******",
        	"created_at": "2022-04-29T09:41:08.583Z",
        	"created_by": "_airworthy"
      	},
      	{
        	"id": "recur_readable",
        	"type": "group",
        	"group_id": "_moonbeam",
        	"url": "https://your-host.your-domain/webhook",
        	"events": [
          		"sync_start",
          		"sync_end"
        	],
        	"active": true,
        	"secret": "******",
        	"created_at": "2022-04-29T10:45:00.000Z",
        	"created_by": "_airworthy"
      	}],
        "next_cursor": null
    }
	`
)

func setupMockClientWebhooksDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	webhooksDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/webhooks").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			webhooksDataSourceMockData = createMapFromJsonString(t, webhooksMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", webhooksDataSourceMockData), nil
		},
	)
}

func TestDataSourceWebhooksMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_webhooks" "test_webhooks" {
			provider = fivetran-provider
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, webhooksDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, webhooksDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.id", "program_quoth"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.type", "account"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.url", "https://your-host.your-domain/webhook"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.events.0", "sync_start"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.events.1", "sync_end"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.active", "true"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.secret", "******"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.created_at", "2022-04-29T09:41:08.583Z"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.created_by", "_airworthy"),

			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.1.id", "recur_readable"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.1.type", "group"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.1.group_id", "_moonbeam"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.1.url", "https://your-host.your-domain/webhook"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.1.events.0", "sync_start"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.1.events.1", "sync_end"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.1.active", "true"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.1.secret", "******"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.1.created_at", "2022-04-29T10:45:00.000Z"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.1.created_by", "_airworthy"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientWebhooksDataSourceConfigMapping(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
