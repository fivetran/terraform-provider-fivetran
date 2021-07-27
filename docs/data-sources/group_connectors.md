---
page_title: "Data Source: fivetran_group_connectors"
---

# Data Source: fivetran_group_connectors

This data source returns a list of information about all connectors within a group in your Fivetran account.

## Example

```hcl
data "fivetran_group_connectors" "connectors" {
    id = "anonymous_mystery"
}
```

## Schema

### Required

- `id` - The unique identifier for the group within the Fivetran system.

### Optional

- `schema` - Optional filter. When used, the response will only contain information for the connector with the specified schema

### Read-Only

- `connectors` - see [below for nested schema](#nestedatt--connectors)

<a id="nestedatt--connectors"></a>
### Nested Schema for `connectors`

Read-Only:

- `connected_by` 
- `created_at` 
- `daily_sync_time` 
- `failed_at` 
- `group_id` 
- `id` 
- `schedule_type` 
- `schema` 
- `service` 
- `service_version` 
- `status` - see [below for nested schema](#nestedobjatt--connectors--status)
- `succeeded_at` 
- `sync_frequency`

<a id="nestedobjatt--connectors--status"></a>
### Nested Schema for `connectors.status`

Read-Only:

- `is_historical_sync` 
- `setup_state` 
- `sync_state` 
- `tasks`- see [below for nested schema](#nestedobjatt--connectors--status--tasks)
- `update_state` 
- `warnings` - see [below for nested schema](#nestedobjatt--connectors--status--warnings)

<a id="nestedobjatt--connectors--status--tasks"></a>
### Nested Schema for `connectors.status.tasks`

Read-Only:

- `code` 
- `message` 

<a id="nestedobjatt--connectors--status--warnings"></a>
### Nested Schema for `connectors.status.warnings`

Read-Only:

- `code` 
- `message` 
