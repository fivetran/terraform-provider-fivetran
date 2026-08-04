---
page_title: "Resource: fivetran_connection_v2"
---

# Resource: fivetran_connection_v2

~> **Alpha:** This resource is in alpha and its schema and behavior may change without notice. Not recommended for production use.

Manages a Fivetran connection. Unlike `fivetran_connector`/`fivetran_connection`, `config` and `auth` are provided as dynamic, service-specific objects: their accepted fields, types, and required/readonly/immutable rules are resolved from connector metadata at plan time rather than from a fixed schema. This means the same resource works across services (e.g. `s3`, `postgres`, `fivetran_log`) without a per-service Terraform schema.

Pause state is managed separately via `fivetran_connection_v2_pause_state`; this resource does not expose a `paused` attribute. **New connections are always created paused.** To start syncing, add a linked `fivetran_connection_v2_pause_state` resource with `paused = false`.

## Example Usage

```hcl
resource "fivetran_connection_v2" "connection" {
    service  = "fivetran_log"
    group_id = "group_id"

    config = {
        schema = "my_schema"
    }
}

# The connection above is created paused. Unpause it to start syncing:
resource "fivetran_connection_v2_pause_state" "pause_state" {
    connection_id = fivetran_connection_v2.connection.id
    paused        = false
}
```

## Schema

### Required

- `group_id` (String) The unique identifier for the Group (Destination) within the Fivetran system. Changing this forces the connection to be replaced.
- `service` (String) The connection service type (e.g., `postgres`, `mysql`, `s3`, `snowflake`). Changing this forces the connection to be replaced.

### Optional

- `auth` (Dynamic, Sensitive) Service-specific authorization configuration. The accepted fields are defined by connector metadata at runtime.
- `config` (Dynamic) Service-specific connection configuration. The accepted fields are defined by connector metadata at runtime.
- `connect_card_config` (Attributes) Configuration for the interactive Connect Card setup flow.
- `daily_sync_time` (String) Sync start time. Only used when `sync_frequency` is `1440`.
- `data_delay_sensitivity` (String) The level of data delay notification threshold. Possible values: `LOW`, `NORMAL`, `HIGH`, `CUSTOM`, `SYNC_FREQUENCY`.
- `data_delay_threshold` (Number) Custom sync delay notification threshold in minutes. Only used when `data_delay_sensitivity` is `CUSTOM`.
- `destination_configuration` (Attributes) Destination-specific configuration for the connection.
- `destination_schema_names` (String) Defines how schema names appear in the destination. Supported values: `FIVETRAN_NAMING`, `SOURCE_NAMING`. Changing this forces the connection to be replaced.
- `hybrid_deployment_agent_id` (String) The hybrid deployment agent ID for the group the connection belongs to.
- `networking_method` (String) The networking method for the connection. Possible values: `Directly`, `SshTunnel`, `ProxyAgent`, `PrivateLink`.
- `pause_after_trial` (Boolean) Whether the connection should be paused after the free trial period has ended.
- `private_link_id` (String) The private link ID. Required when `networking_method` is `PrivateLink`.
- `proxy_agent_id` (String) The ID of the proxy agent to use. Required when `networking_method` is `ProxyAgent`.
- `run_setup_tests` (Boolean) Whether to run setup tests when creating or updating the connection.
- `schedule_type` (String) The connection schedule configuration type. Supported values: `auto`, `manual`.
- `sync_frequency` (Number) The connection sync frequency in minutes.
- `trust_certificates` (Boolean) Whether Fivetran should trust certificates automatically.
- `trust_fingerprints` (Boolean) Whether Fivetran should trust SSH fingerprints automatically.

### Nested Schema for `connect_card_config`

Optional:
- `all_fields` (Boolean) Whether Connect Card should show all setup fields.
- `hide_setup_guide` (Boolean) Whether to hide the setup guide in the Connect Card flow.
- `redirect_uri` (String, Sensitive) URI where the Connect Card flow redirects after setup.

### Nested Schema for `destination_configuration`

Optional:
- `virtual_warehouse` (String) Destination virtual warehouse used by the connection.

### Read-Only

- `connected_by` (String) The unique identifier of the user who created the connection in your account.
- `created_at` (String) The timestamp of the time the connection was created in your account.
- `failed_at` (String) The timestamp of the time the connection sync failed last time.
- `id` (String) The unique identifier for the connection within the Fivetran system.
- `name` (String) The connection's name.
- `service_version` (String) The connection type version within the Fivetran system.
- `status` (Attributes) The current connection status (setup state, sync state, tasks, and warnings).
- `succeeded_at` (String) The timestamp of the time the connection sync succeeded last time.
