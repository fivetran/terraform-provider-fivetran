package fivetran_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceWebhookE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranWebhookResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
            resource "fivetran_webhook" "test_webhook" {
                 provider = fivetran-provider

                 type = "account"
                 url = "https://your-host.your-domain/webhook"
                 secret = "password"
                 active = "false"
                 run_tests = "false"
                 events = ["sync_start","sync_end"]
            }
          `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranWebhookResourceCreate(t, "fivetran_webhook.test_webhook"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "type", "account"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "url", "https://your-host.your-domain/webhook"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "secret", "password"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "active", "false"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.0", "sync_end"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.1", "sync_start"),
				),
			},
			{
				Config: `
            resource "fivetran_webhook" "test_webhook" {
                 provider = fivetran-provider

                 type = "account"
                 url = "https://your-host.your-domain/webhook_1"
                 secret = "password_2"
                 active = "false"
                 run_tests = "false"
                 events = ["sync_start","sync_end", "connection_failure"]
            }
          `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranWebhookResourceUpdate(t, "fivetran_webhook.test_webhook"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "type", "account"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "url", "https://your-host.your-domain/webhook_1"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "secret", "password_2"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "active", "false"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.0", "connection_failure"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.1", "sync_end"),
					resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.2", "sync_start"),
				),
			},
		},
	})
}

func testFivetranWebhookResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		_, err := client.NewWebhookDetails().WebhookId(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields
		return nil
	}
}

func testFivetranWebhookResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)
		_, err := client.NewWebhookDetails().WebhookId(rs.Primary.ID).Do(context.Background())

		if err != nil {
			return err
		}
		//todo: check response _  fields
		return nil
	}
}

func testFivetranWebhookResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_webhook" {
			continue
		}

		response, err := client.NewWebhookDetails().WebhookId(rs.Primary.ID).Do(context.Background())
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if response.Code != "NotFound" {
			return errors.New("Webhook " + rs.Primary.ID + " still exists.")
		}

	}

	return nil
}
