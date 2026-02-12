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

## Common Use Cases

### Credential Rotation

One of the primary benefits of `fivetran_connection_config` is the ability to rotate credentials independently without touching the connection structure:

```hcl
# Rotate database credentials without modifying connection settings
resource "fivetran_connection_config" "postgres" {
    connection_id = fivetran_connection.postgres.id

    config = jsonencode({
        host          = "db.example.com"  # Unchanged
        port          = 5432               # Unchanged
        database      = "mydb"             # Unchanged
        update_method = "XMIN"             # Unchanged
    })

    auth = jsonencode({
        user     = var.new_db_user        # Updated credential
        password = var.new_db_password    # Updated credential
    })

    run_setup_tests = true  # Validates new credentials
}
```

Running `terraform plan` will show only the `auth` field changing, making credential rotation safer and more transparent.

### Configuration Updates

Update connection settings without exposing or modifying credentials:

```hcl
# Change database host without touching credentials
resource "fivetran_connection_config" "postgres" {
    connection_id = fivetran_connection.postgres.id

    config = jsonencode({
        host          = "new-db.example.com"  # Updated
        port          = 5433                   # Updated
        database      = "mydb"
        update_method = "XMIN"
    })

    auth = jsonencode({
        user     = var.db_user      # Unchanged - credentials stay secure
        password = var.db_password  # Unchanged
    })

    run_setup_tests = true
}
```

### Integration with Secrets Management

Easily integrate with AWS Secrets Manager, HashiCorp Vault, or other secrets management systems:

```hcl
# AWS Secrets Manager example
data "aws_secretsmanager_secret_version" "db_creds" {
  secret_id = "prod/fivetran/postgres"
}

locals {
  db_secret = jsondecode(data.aws_secretsmanager_secret_version.db_creds.secret_string)
}

resource "fivetran_connection_config" "postgres" {
    connection_id = fivetran_connection.postgres.id

    config = jsonencode({
        host          = local.db_secret["host"]
        port          = local.db_secret["port"]
        database      = local.db_secret["database"]
        user          = local.db_secret["username"]
        update_method = "XMIN"
    })

    auth = jsonencode({
        password = local.db_secret["password"]
    })

    run_setup_tests = true
}
```

### Separate Teams/Workflows

Enable different teams to manage different aspects:

```hcl
# Platform team manages connection structure
resource "fivetran_connection" "postgres" {
    group_id = fivetran_group.example.id
    service  = "postgres"

    destination_schema {
        prefix = "my_postgres"
    }

    networking_method = "ProxyAgent"
    proxy_agent_id    = fivetran_proxy_agent.platform.id
}

# Database team manages credentials (separate Terraform workspace)
resource "fivetran_connection_config" "postgres" {
    connection_id = data.terraform_remote_state.platform.outputs.postgres_connection_id

    config = jsonencode({
        host          = var.db_host
        port          = var.db_port
        database      = var.db_name
        user          = var.db_user
        update_method = "XMIN"
    })

    auth = jsonencode({
        password = var.db_password
    })

    run_setup_tests = true
}
```

### Environment-Specific Configuration

Easily manage different configurations across environments while keeping the connection structure consistent:

```hcl
# Development environment
resource "fivetran_connection_config" "postgres_dev" {
    connection_id = fivetran_connection.postgres_dev.id

    config = jsonencode({
        host          = "dev-db.example.com"
        port          = 5432
        database      = "dev_db"
        user          = var.dev_db_user
        update_method = "XMIN"
    })

    auth = jsonencode({
        password = var.dev_db_password
    })
}

# Production environment (different credentials, host, but same structure)
resource "fivetran_connection_config" "postgres_prod" {
    connection_id = fivetran_connection.postgres_prod.id

    config = jsonencode({
        host          = "prod-db.example.com"
        port          = 5432
        database      = "prod_db"
        user          = var.prod_db_user
        update_method = "XMIN"
    })

    auth = jsonencode({
        password = var.prod_db_password
    })

    run_setup_tests = true  # Extra validation for production
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
- [`fivetran_connector`](connector.md) - Legacy connector resource (planned for deprecation)
- [Migration Guide](../guides/migrating-from-connector-to-connection.md) - Migrate from fivetran_connector
- [Fivetran API Documentation](https://fivetran.com/docs/rest-api/connectors) - Connection configuration reference
