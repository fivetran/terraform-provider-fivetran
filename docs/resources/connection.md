---
page_title: "Resource: fivetran_connection"
---

# Resource: fivetran_connection

This resource allows you to create, update, and delete connections in Fivetran using JSON-based configuration.

The `fivetran_connection` resource creates a connection with basic metadata (service type, destination schema, networking options). To configure connection details like host, port, credentials, use the `fivetran_connection_config` resource.

## Example Usage

### Basic PostgreSQL Connection with Minimal Config

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

    config = jsonencode({
        update_method = "XMIN"
    })

    run_setup_tests    = false
    trust_certificates = false
    trust_fingerprints = false
}
```

### PostgreSQL Connection with Full Configuration

```hcl
resource "fivetran_connection" "postgres_full" {
    group_id = fivetran_group.example.id
    service  = "postgres"

    destination_schema {
        prefix = "my_postgres"
    }

    config = jsonencode({
        update_method = "XMIN"
        host          = local.database_config.host
        port          = 5432
        database      = "mydb"
        user          = "fivetran_user"
    })

    run_setup_tests = false
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
resource "fivetran_connection" "snowflake_connection" {
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

## Migrating from fivetran_connector

If you're currently using the legacy `fivetran_connector` resource, you can migrate to `fivetran_connection` and `fivetran_connection_config` to benefit from:

- **Better credential management**: Rotate credentials independently without touching connection configuration
- **Improved security**: Separate metadata from sensitive credentials
- **Cleaner state management**: Better resource dependency tracking
- **More flexible workflows**: Update config and auth independently

### Quick Migration Example

**Before** (fivetran_connector):
```hcl
resource "fivetran_connector" "postgres" {
  group_id = fivetran_group.example.id
  service  = "postgres"

  destination_schema {
    prefix = "my_postgres"
  }

  config {
    host          = "db.example.com"
    port          = 5432
    database      = "mydb"
    user          = "fivetran_user"
    password      = var.db_password
    update_method = "XMIN"
  }
}
```

**After** (fivetran_connection + fivetran_connection_config):
```hcl
resource "fivetran_connection" "postgres" {
  group_id = fivetran_group.example.id
  service  = "postgres"

  destination_schema {
    prefix = "my_postgres"
  }

  config = jsonencode({
    update_method = "XMIN"
  })
}

resource "fivetran_connection_config" "postgres" {
  connection_id = fivetran_connection.postgres.id

  config = jsonencode({
    host          = "db.example.com"
    port          = 5432
    database      = "mydb"
    user          = "fivetran_user"
    update_method = "XMIN"
  })

  auth = jsonencode({
    password = var.db_password
  })
}
```

For complete migration steps, including import procedures, troubleshooting, and rollback instructions, see the [Migration Guide](../guides/migrating-from-connector-to-connection.md).

## Schema

### Required

- `group_id` (String) The unique identifier for the destination group.
- `service` (String) The connection service type (e.g., `postgres`, `mysql`, `s3`, `snowflake`). See [Fivetran connection types documentation](https://fivetran.com/docs/connectors) for available services.

### Optional

- `destination_schema` (Block, Optional) Configuration for the destination schema. See [destination_schema](#nested-schema-for-destination_schema) below.
- `config` (String) Optional connection configuration as a JSON-encoded string. This config is merged with destination_schema fields and sent to the API during creation. The connection resource does not read this field back, allowing it to be managed separately by the `fivetran_connection_config` resource. Use this to provide service-specific required fields (e.g., `update_method` for Postgres/MySQL) or full connection configuration.
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

- `name` (String) The explicit schema name in the destination. Use this for connections like S3 that allow explicit schema names.
- `prefix` (String) The schema prefix in the destination. The connection name will be derived from this. Use this for most database connections.
- `table` (String) The table name for single-table connections.
- `table_group_name` (String) The table group name for multi-table connections.

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
- **Paused State:** Connections are created in a paused state by default. Use the connection schedule resource or the Fivetran UI to unpause the connection.
- **Service Types:** See the [Fivetran documentation](https://fivetran.com/docs/connectors) for the complete list of available connection services.

## See Also

- [`fivetran_connection_config`](connection_config.md) - Configure connection details (host, credentials, etc.)
- [`fivetran_connector`](connector.md) - Legacy connector resource (planned for deprecation)
- [Migration Guide](../guides/migrating-from-connector-to-connection.md) - Migrate from fivetran_connector to fivetran_connection
- [`fivetran_connector_schedule`](connector_schedule.md) - Manage connection sync schedules
- [`fivetran_connector_schema`](connector_schema_config.md) - Manage connection schema settings