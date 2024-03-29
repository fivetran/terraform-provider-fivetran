---
page_title: "Data Source: fivetran_webhook"
---

# Data Source: fivetran_webhook

This data source returns a webhook object.

## Example Usage

```hcl
data "fivetran_webhook" "webhook" {
    id = "webhook_id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) The webhook ID

### Read-Only

- `active` (Boolean) Boolean, if set to true, webhooks are immediately sent in response to events
- `created_at` (String) The webhook creation timestamp
- `created_by` (String) The ID of the user who created the webhook.
- `events` (Set of String) The array of event types
- `group_id` (String) The group ID
- `run_tests` (Boolean) Specifies whether the setup tests should be run
- `secret` (String) The secret string used for payload signing and masked in the response.
- `type` (String) The webhook type (group, account)
- `url` (String) Your webhooks URL endpoint for your application