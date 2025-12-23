---
page_title: "Resource: fivetran_connection_config"
---

# Resource: fivetran_connection_config

This resource allows you to manage connection configuration (config and auth fields) separately from the connection structure.

The `fivetran_connection_config` resource configures connection-specific details like host, port, database credentials, and authentication settings. Use this in combination with `fivetran_connection` for flexible connection management.

## Example Usage

### PostgreSQL Connection with Config and Auth

```hcl
resource "fivetran_connection" "postgres" {
    group_id = fivetran_group.example.id
    service  = "postgres"

    destination_schema {
        prefix = "my_postgres"
    }

    run_setup_tests = false
}

resource "fivetran_connection_config" "postgres" {
    connection_id = fivetran_connection.postgres.id

    config = jsonencode({
        update_method = "XMIN"
        host          = var.database_host
        port          = 5432
        database      = "production_db"
    })

    auth = jsonencode({
        user     = var.db_user
        password = var.db_password
    })

    run_setup_tests    = true
    trust_certificates = false
    trust_fingerprints = false
}
```

### MySQL Connection with Config Only

```hcl
resource "fivetran_connection" "mysql" {
    group_id = fivetran_group.example.id
    service  = "mysql"

    destination_schema {
        prefix = "my_mysql"
    }
}

resource "fivetran_connection_config" "mysql" {
    connection_id = fivetran_connection.mysql.id

    config = jsonencode({
        update_method = "BINLOG"
        host          = "mysql.example.com"
        port          = 3306
        database      = "mydb"
        user          = "fivetran_user"
        password      = var.mysql_password
    })

    run_setup_tests = true
}
```

### S3 Connection with Auth Only

```hcl
resource "fivetran_connection" "s3" {
    group_id = fivetran_group.example.id
    service  = "s3"

    destination_schema {
        name = "s3_data"
    }
}

resource "fivetran_connection_config" "s3" {
    connection_id = fivetran_connection.s3.id

    auth = jsonencode({
        role_arn = "arn:aws:iam::123456789:role/fivetran-access"
    })

    run_setup_tests = true
}
```

### Connection with Sensitive Credentials

```hcl
resource "fivetran_connection_config" "postgres_secure" {
    connection_id = fivetran_connection.postgres.id

    config = jsonencode({
        update_method = "XMIN"
        host          = data.aws_db_instance.postgres.address
        port          = data.aws_db_instance.postgres.port
        database      = data.aws_db_instance.postgres.db_name
    })

    auth = jsonencode({
        user     = data.aws_secretsmanager_secret_version.db_creds.secret_string["username"]
        password = data.aws_secretsmanager_secret_version.db_creds.secret_string["password"]
    })

    run_setup_tests    = true
    trust_certificates = true
}
```

## Schema

### Required

- `connection_id` (String) The unique identifier for the connection.

### Optional

- `config` (String) Connection configuration as a JSON-encoded string. This field uses semantic JSON equality, so whitespace and key order differences won't trigger updates.
- `auth` (String) Authentication configuration as a JSON-encoded string. This field uses semantic JSON equality. Typically contains credentials like username, password, API keys, or role ARNs.
- `run_setup_tests` (Boolean) Whether to run setup tests when applying configuration. Default: `false`. When `true`, Fivetran validates the configuration by testing the connection. **Note:** This is a plan-only attribute and will not be stored in state.
- `trust_certificates` (Boolean) Whether to automatically trust SSL certificates. Default: `false`. **Note:** This is a plan-only attribute.
- `trust_fingerprints` (Boolean) Whether to automatically trust SSH fingerprints. Default: `false`. **Note:** This is a plan-only attribute.

### Read-Only

- `id` (String) The unique identifier for this configuration (same as connection_id).

## Import

Connection configurations can be imported using the connection ID:

```shell
terraform import fivetran_connection_config.example connection_id_here
```

**Note:** When importing, the `run_setup_tests`, `trust_certificates`, and `trust_fingerprints` attributes will not be imported as they are plan-only attributes.

## Notes

- **Semantic JSON Equality:** The `config` and `auth` fields use semantic JSON comparison, meaning changes in whitespace, key ordering, or formatting won't trigger unnecessary updates.
- **Separation of Concerns:** This resource allows you to manage connection configuration separately from the connection structure, enabling better security practices and workflow flexibility.
- **Setup Tests:** When `run_setup_tests` is `true`, Fivetran will validate the configuration. Any test failures will appear as warnings in the Terraform output.
- **Sensitive Data:** Consider using Terraform's `sensitive` attribute or external secret management for credentials in the `auth` field.
- **Configuration Updates:** Both `config` and `auth` are optional. You can update one without affecting the other.

## See Also

- [`fivetran_connection`](connection.md) - Create the connection structure
- [Fivetran API Documentation](https://fivetran.com/docs/rest-api/connectors) - Connection configuration reference
