package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	webhooksDataSourceMockGetHandler *mock.Handler
	webhooksDataSourceMockData       map[string]interface{}
)

const (
	webhooksMappingResponse = `
    {
        "items":[
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
    },
    	{
    	"id": "recur_readable1",
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
    }`
)

func setupMockClientWebhooksDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	webhooksDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/webhooks").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			webhooksDataSourceMockData = tfmock.CreateMapFromJsonString(t, webhooksMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", webhooksDataSourceMockData), nil
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
				tfmock.AssertEqual(t, webhooksDataSourceMockGetHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, webhooksDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.id", "recur_readable"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.type", "group"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.group_id", "_moonbeam"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.url", "https://your-host.your-domain/webhook"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.events.0", "sync_end"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.events.1", "sync_start"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.active", "true"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.secret", "******"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.created_at", "2022-04-29T10:45:00.000Z"),
			resource.TestCheckResourceAttr("data.fivetran_webhooks.test_webhooks", "webhooks.0.created_by", "_airworthy"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientWebhooksDataSourceConfigMapping(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
