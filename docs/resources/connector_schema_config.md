---
page_title: "Resource: fivetran_connector_schema_config"
---

# Resource: fivetran_connector_schema_config

This resource allows you to manage connector Standard Config settings. Choose schema change handling policy and enable/disable schemas, table and columns.

The resource is in ALPHA state. Resource schema and behavior may be changed further.

## Usage guide

Once you have defined `schema_change_handling` you should keep in ming that all schema settings will be aligned to chosen policy if not defined in config.
In `schema` you define only **exclusions** that differs from chosen policy. Default value for `enabled` attribute is `true` so it can be omited in most cases.

Allowed `schema_change_handling` policies:
- ALLOW_ALL - all schemas, tables and columns are ENABLED by default, config contains only DISABLED items
- BLOCK_ALL - all schemas, tables and columns are DISABLED by default, config contains only ENABLED items
- ALLOW_COLUMNS - all schemas and tables are DISABLED by default, but all columns are ENABLED, config contains ENABLED schemas and tables, and disabled columns

Policy settings and config can't affect core tables and columns, there is no ability manage these elements. There will be no errors, provider will just ignore them.

## Usage examples

### ALLOW_ALL example

```hcl
resource "fivetran_connector_schema_config" "schema" {
  connector_id = "connector_id"
  schema_change_handling = "ALLOW_ALL"
  schema {
    name = "schema_name"
    table {
      name = "table_name"
      column {
        name = "hashed_column_name"
        hashed = "true"
      }
      column {
        name = "blocked_column_name"
        enabled = "false"
      }
    }
    table {
      name = "blocked_table_name"
      enabled = "false"
    }
  }
  schema{
    name = "blocked_schema"
    enabld = "false"
  }
}
```

Settings we get here:
- All new and existing schemas except `blocked_schema` would be enabled
- All new and existing tables in schema `schema_name` except `blocked_table_name` would be enabled
- All new and existing columns in table `table_name` of schema `schema_name` except `blocked_column_name` would be enabled
- Column with name `hashed_column_name` would be hashed in table `table_name` in schema `schema_name`
- All new columns/tables/schemas would be enabled once captured by connector on sync if not disabled by system

### BLOCK_ALL example

All included into config schemas, tables and columns would be enabled by default (`enabled` default value is `true`)

```hcl
resource "fivetran_connector_schema_config" "schema" {
  connector_id = "connector_id"
  schema_change_handling = "BLOCK_ALL"
  schema {
    name = "schema_name"
    table {
      name = "table_name"
      column {
        name = "hashed_column_name"
        hashed = "true"
      }
    }
    table {
      name = "enabled_table_name"
    }
  }
  schema{
    name = "enabled_schema"
  }
}
```

Settings we get here:

- All new and existing schemas except `enabled_schema` and `schema_name` would be disabled
- Only system-enabled tables and columns would be enabled in `enabled_schema`
- All new and existing tables in schema `schema_name` except `enabled_table_name`, `table_name` and system-enabled tables would be disabled
- All new and existing columns in table `table_name` of schema `schema_name` except `hashed_column_name` and system-enabled columns would be disabled
- Column `hashed_column_name` would be hashed in table `table_name` in schema `schema_name`
- All new non system-enabled columns/tables/schemas would be disables once captured by connector on sync

### ALLOW_COLUMNS example

```hcl
resource "fivetran_connector_schema_config" "schema" {
  connector_id = "connector_id"
  schema_change_handling = "ALLOW_COLUMNS"
  schema {
    name = "schema_name"
    table {
      name = "table_name"
      column {
        name = "hashed_column_name"
        hashed = "true"
      }
      column {
        name = "disabled_columns_name"
        enabled = "false"
      }
    }
    table {
      name = "enabled_table_name"
    }
  }
  schema{
    name = "enabled_schema"
  }
}
```

Settings we get here:

- All included into config schemas, tables and columns would be enabled by default if no `enabled` attribute defined (`enabled` default value is `true`).
- All new and existing schemas except `enabled_schema` and `schema_name` would be disabled
- Only system-enabled tables and columns would be enabled in `enabled_schema`
- All new and existing tables in schema `schema_name` except `enabled_table_name`, `table_name` and system-enabled tables would be disabled
- All new and existing columns in table `table_name` of schema `schema_name` except `disabled_columns_name` and system-enabled columns would be enabled
- Column `hashed_column_name` would be hashed in table `table_name` in schema `schema_name`
- All new non system-enabled tables/schemas would be disabled once captured by connector on sync
- All new non system-enabled columns inside enabled tables (including system enabled-tables) would be enabled once captured by connector on sync


### Locked table columns management

You can't manage Core-table enablement, but you can manage its non-locked columns. For example your schema `schema_name` has Core-table `system_locked_table` that can't be disabled and you want to disable one of its columns named `columns_name`:

```hcl
resource "fivetran_connector_schema_config" "schema" {
  connector_id = "connector_id"
  schema_change_handling = "ALLOW_COLUMNS"
  schema {
    name = "schema_name"
    table {
      name = "system_locked_table"
      column {
        name = "columns_name"
        enabled = "false"
      }
    }
  }
}
```

## Schema

### Required

- `connector_id` - the Fivetran Connector ID of connector which standard config is managed by the resource
- `schema_change_handling` - the policy value (ALLOW_ALL | ALLOW_COLUMNS | BLOCK_ALL)

### Optional

- `schema` - set of schema settings (see [below for nested schema](#nestedblock--schema))

<a id="nestedblock--schema"></a>
## Nested Schema for `schema`

### Required

- `name` - the name of schema in source

### Optional

- `enabled` - is enabled in settings (default: "true")
- `table` - set of table settings (see [below for nested schema](#nestedblock--table))

<a id="nestedblock--table"></a>
## Nested Schema for `table`

### Required
- `name` - table name in source

### Optional

- `enabled` - is enabled in settings (default: "true")
- `column` - set of table settings (see [below for nested schema](#nestedblock--column))

<a id="nestedblock--column"></a>
## Nested Schema for `column`

### Required

- `name` - column name in source

### Optional

- `enabled` - is enabled in settings (default: "true")
- `hashed` - is column set as hashed in settings (default: "false")