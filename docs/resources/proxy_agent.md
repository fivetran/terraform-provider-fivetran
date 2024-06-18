---
page_title: "Resource: fivetran_proxy_agent"
---

# Resource: fivetran_proxy_agent

This resource allows you to create, update, and delete proxy agent.

## Example Usage

```hcl
resource "fivetran_proxy_agent" "test_proxy_agent" {
    provider = fivetran-provider

    display_name = "display_name"
    group_region = "group_region"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `display_name` (String) Proxy agent name.
- `group_region` (String) Data processing location. This is where Fivetran will operate and run computation on data.

### Read-Only

- `account_id` (String) The unique identifier for the account.
- `created_by` (String) The actor who created the proxy agent.
- `id` (String) The unique identifier for the proxy within your account.
- `proxy_server_uri` (String) The proxy server URI.
- `registred_at` (String) The timestamp of the time the proxy agent was created in your account.
- `salt` (String) The salt.
- `token` (String) The auth token.