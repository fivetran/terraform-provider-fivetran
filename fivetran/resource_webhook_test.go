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
        PreCheck:     func() {},
        Providers:    testProviders,
        CheckDestroy: testFivetranWebhookResourceDestroy,
        Steps: []resource.TestStep{
            {
                Config: `
            resource "fivetran_webhook" "test_webhook" {
                 provider = fivetran-provider

                 id = "recur_readable"
                 type = "group"
                 group_id = "_moonbeam"
                 url = "https://your-host.your-domain/webhook"
                 secret = "password"
                 active = false
                 created_at = "2022-04-29T10:45:00.000Z"
                 created_by = "_airworthy"
                 events : ["sync_start","sync_end"]
            }
          `,
                Check: resource.ComposeAggregateTestCheckFunc(
                    testFivetranWebhookResourceCreate(t, "fivetran_webhook.test_webhook"),
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
            },
            {
                Config: `
            resource "fivetran_webhook" "test_webhook" {
                 provider = fivetran-provider

                 id = "recur_readable"
                 type = "group"
                 group_id = "_moonbeam"
                 url = "https://your-host.your-domain/webhook_1"
                 secret = "password"
                 active = false
                 created_at = "2022-04-29T10:45:00.000Z"
                 created_by = "_airworthy"
                 events : ["sync_start","sync_end"]
            }
          `,
                Check: resource.ComposeAggregateTestCheckFunc(
                    testFivetranWebhookResourceCreate(t, "fivetran_webhook.test_webhook"),
                    resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "id", "recur_readable"),
                    resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "type", "group"),
                    resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "group_id", "_moonbeam"),
                    resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "url", "https://your-host.your-domain/webhook_1"),
                    resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "secret", "password"),
                    resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "active", "false"),
                    resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "created_at", "2022-04-29T10:45:00.000Z"),
                    resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "created_by", "_airworthy"),
                    resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.0", "sync_start"),
                    resource.TestCheckResourceAttr("fivetran_webhook.test_webhook", "events.1", "sync_end"),                ),
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
