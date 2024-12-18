---
page_title: "Data Source: fivetran_local_processing_agent"
---

# Data Source: fivetran_local_processing_agent

NOTE: In connection with the general availability of the hybrid deployment functionality and in order to synchronize internal terminology, we have deprecate this data source.

This data source returns a local processing agent object.

## Example Usage

```hcl
data "fivetran_local_processing_agent" "local_processing_agent" {
    id = "local_processing_agent_id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) The unique identifier for the local processing agent within your account.

### Read-Only

- `display_name` (String) The unique name for the local processing agent.
- `group_id` (String) The unique identifier for the Group within the Fivetran system.
- `registered_at` (String) The timestamp of the time the local processing agent was created in your account.
- `usage` (Attributes Set) (see [below for nested schema](#nestedatt--usage))

<a id="nestedatt--usage"></a>
### Nested Schema for `usage`

Read-Only:

- `connection_id` (String) The unique identifier of the connection associated with the agent.
- `schema` (String) The connection schema name.
- `service` (String) The connection type.