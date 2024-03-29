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
	webhookDataSourceMockGetHandler *mock.Handler
	webhookDataSourceMockData       map[string]interface{}
)

const (
	webhookMappingResponse = `
	{
    	"id": "webhook_id",
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
    }
	`
)

func setupMockClientWebhookDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	webhookDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/webhooks/webhook_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			webhookDataSourceMockData = tfmock.CreateMapFromJsonString(t, webhookMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", webhookDataSourceMockData), nil
		},
	)
}

func TestDataSourceWebhookMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_webhook" "test_webhook" {
			provider = fivetran-provider
			id = "webhook_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, webhookDataSourceMockGetHandler.Interactions, 2)
				tfmock.AssertNotEmpty(t, webhookDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_webhook.test_webhook", "id", "webhook_id"),
			resource.TestCheckResourceAttr("data.fivetran_webhook.test_webhook", "type", "group"),
			resource.TestCheckResourceAttr("data.fivetran_webhook.test_webhook", "group_id", "_moonbeam"),
			resource.TestCheckResourceAttr("data.fivetran_webhook.test_webhook", "url", "https://your-host.your-domain/webhook"),
			resource.TestCheckResourceAttr("data.fivetran_webhook.test_webhook", "events.0", "sync_end"),
			resource.TestCheckResourceAttr("data.fivetran_webhook.test_webhook", "events.1", "sync_start"),
			resource.TestCheckResourceAttr("data.fivetran_webhook.test_webhook", "active", "true"),
			resource.TestCheckResourceAttr("data.fivetran_webhook.test_webhook", "created_at", "2022-04-29T10:45:00.000Z"),
			resource.TestCheckResourceAttr("data.fivetran_webhook.test_webhook", "created_by", "_airworthy"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientWebhookDataSourceConfigMapping(t)
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
