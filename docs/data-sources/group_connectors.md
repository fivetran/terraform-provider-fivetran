---
page_title: "Data Source: fivetran_group_users"
---

# Data Source: fivetran_group_users

This data source returns a list of information about all users within a group in your Fivetran account.

## Example Usage

```hcl
data "fivetran_group_users" "group_users" {
        id = "anonymous_mystery"
}
```

## Schema

### Optional

- `schema` (String)

### Read-Only

- `connectors` (Set of Object) (see [below for nested schema](#nestedatt--connectors))
- `id` (String) The ID of this resource.

<a id="nestedatt--connectors"></a>
### Nested Schema for `connectors`

Read-Only:

- `connected_by` (String)
- `created_at` (String)
- `daily_sync_time` (String)
- `failed_at` (String)
- `group_id` (String)
- `id` (String)
- `schedule_type` (String)
- `schema` (String)
- `service` (String)
- `service_version` (Number)
- `status` (Set of Object) (see [below for nested schema](#nestedobjatt--connectors--status))
- `succeeded_at` (String)
- `sync_frequency` (Number)

<a id="nestedobjatt--connectors--status"></a>
### Nested Schema for `connectors.status`

Read-Only:

- `is_historical_sync` (Boolean)
- `setup_state` (String)
- `sync_state` (String)
- `tasks` (Set of Object) (see [below for nested schema](#nestedobjatt--connectors--status--tasks))
- `update_state` (String)
- `warnings` (Set of Object) (see [below for nested schema](#nestedobjatt--connectors--status--warnings))

<a id="nestedobjatt--connectors--status--tasks"></a>
### Nested Schema for `connectors.status.tasks`

Read-Only:

- `code` (String)
- `message` (String)


<a id="nestedobjatt--connectors--status--warnings"></a>
### Nested Schema for `connectors.status.warnings`

Read-Only:

- `code` (String)
- `message` (String)
