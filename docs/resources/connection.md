---
page_title: "Resource: fivetran_connection"
---

# Resource: fivetran_connection

This resource allows you to create, update, and delete connections in Fivetran using JSON-based configuration.

The `fivetran_connection` resource creates a connection with basic metadata (service type, destination schema, networking options). To configure connection details like host, port, credentials, use the `fivetran_connection_config` resource.

## Example Usage

### Basic PostgreSQL Connection

```hcl
resource "fivetran_group" "example" {
    name = "My Destination"
}

resource "fivetran_connection" "postgres" {
    group_id = fivetran_group.example.id
    service  = "postgres"

    destination_schema {
        prefix = "my_postgres"
    }

    run_setup_tests    = false
    trust_certificates = false
    trust_fingerprints = false
}
```

### MySQL Connection with Proxy Agent

```hcl
resource "fivetran_proxy_agent" "example" {
    display_name = "My Proxy Agent"
    group_region = "GCP_US_EAST4"
}

resource "fivetran_connection" "mysql" {
    group_id = fivetran_group.example.id
    service  = "mysql"

    destination_schema {
        prefix = "my_mysql"
    }

    networking_method = "ProxyAgent"
    proxy_agent_id    = fivetran_proxy_agent.example.id

    run_setup_tests = false
}
```

### S3 Connection with Custom Schema Name

```hcl
resource "fivetran_connection" "s3" {
    group_id = fivetran_group.example.id
    service  = "s3"

    destination_schema {
        name = "s3_data_schema"
    }

    run_setup_tests = false
}
```

### Connection with Data Delay Settings

```hcl
resource "fivetran_connection" "snowflake_connector" {
    group_id = fivetran_group.example.id
    service  = "snowflake"

    destination_schema {
        prefix = "snowflake_data"
    }

    data_delay_sensitivity = "HIGH"
    data_delay_threshold   = 30

    run_setup_tests = false
}
```

## Schema

### Required

- `group_id` (String) The unique identifier for the destination group.
- `service` (String) The connector service type (e.g., `postgres`, `mysql`, `s3`, `snowflake`). See [Fivetran connector types documentation](https://fivetran.com/docs/connectors) for available services.

### Optional

- `destination_schema` (Block, Optional) Configuration for the destination schema. See [destination_schema](#nested-schema-for-destination_schema) below.
- `run_setup_tests` (Boolean) Whether to run setup tests when creating the connection. Default: `false`. **Note:** This is a plan-only attribute and will not be stored in state.
- `trust_certificates` (Boolean) Whether to automatically trust SSL certificates. Default: `false`. **Note:** This is a plan-only attribute.
- `trust_fingerprints` (Boolean) Whether to automatically trust SSH fingerprints. Default: `false`. **Note:** This is a plan-only attribute.
- `networking_method` (String) The networking method for the connection. Possible values: `Directly`, `SshTunnel`, `ProxyAgent`.
- `proxy_agent_id` (String) The ID of the proxy agent to use. Required when `networking_method` is `ProxyAgent`.
- `hybrid_deployment_agent_id` (String) The ID of the hybrid deployment agent.
- `private_link_id` (String) The ID of the private link configuration.
- `data_delay_sensitivity` (String) The sensitivity level for data delay notifications. Possible values: `LOW`, `NORMAL`, `HIGH`, `CUSTOM`. Default: `NORMAL`.
- `data_delay_threshold` (Number) Custom delay threshold in minutes. Only used when `data_delay_sensitivity` is `CUSTOM`.

### Read-Only

- `id` (String) The unique identifier for the connection.
- `name` (String) The connection name (typically derived from destination schema).
- `connected_by` (String) The ID of the user who created the connection.
- `created_at` (String) The timestamp when the connection was created.

<a id="nested-schema-for-destination_schema"></a>
### Nested Schema for `destination_schema`

Optional:

- `name` (String) The explicit schema name in the destination. Use this for connectors like S3 that allow explicit schema names.
- `prefix` (String) The schema prefix in the destination. The connection name will be derived from this. Use this for most database connectors.
- `table` (String) The table name for single-table connectors.
- `table_group_name` (String) The table group name for multi-table connectors.

**Note:** Use either `name` or `prefix`, not both.

## Import

Connections can be imported using the connection ID:

```shell
terraform import fivetran_connection.example connection_id_here
```

**Note:** When importing, the `run_setup_tests`, `trust_certificates`, and `trust_fingerprints` attributes will not be imported as they are plan-only attributes.

## Notes

- **Configuration Details:** This resource creates the connection structure. To configure connection-specific details (host, port, credentials, etc.), use the [`fivetran_connection_config`](connection_config.md) resource.
- **Setup Tests:** When `run_setup_tests` is `true`, Fivetran will validate the connection configuration. Any test failures will appear as warnings in the Terraform output.
- **Paused State:** Connections are created in a paused state by default. Use the connector schedule resource or the Fivetran UI to unpause the connection.
- **Service Types:** See the [Fivetran documentation](https://fivetran.com/docs/connectors) for the complete list of available connector services.

## See Also

- [`fivetran_connection_config`](connection_config.md) - Configure connection details (host, credentials, etc.)
- [`fivetran_connector_schedule`](connector_schedule.md) - Manage connection sync schedules
- [`fivetran_connector_schema`](connector_schema_config.md) - Manage connection schema settings
