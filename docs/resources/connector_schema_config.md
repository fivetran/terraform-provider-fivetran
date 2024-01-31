 ---
page_title: "Resource: fivetran_connector_schema_config"
---

# Resource: fivetran_connector_schema_config

This resource allows you to manage the Standard Configuration settings of a connector:
 - Define the schema change handling settings
 - Enable and disable schemas, tables, and columns

The resource is in **ALPHA** state. The resource schema and behavior are subject to change without prior notice.

Known issues:
 - Definition of `sync_mode` for table causes infinite drifting changes in plan

## Usage guide

Note that all configuration settings are aligned to the `schema_change_handling` settings,  except the settings explicitly specified in `schema`.
In `schema`, you only override the default settings defined by the chosen `schema_change_handling` option. The default value for the `enabled` attribute is `true` so it can be omitted when you want to enable schemas, tables, or columns.
The allowed `schema_change_handling` options are as follows:
- `ALLOW_ALL`- all schemas, tables and columns are ENABLED by default. You only need  to explicitly specify DISABLED items or hashed tables
- `BLOCK_ALL` - all schemas, tables and columns are DISABLED by default, the configuration only specifies ENABLED items
- `ALLOW_COLUMNS` - all schemas and tables are DISABLED by default, but all columns are ENABLED by default, the configuration specifies ENABLED schemas and tables, and DISABLED columns

Note that system-enabled tables and columns (such as primary and foreign key columns, and [system tables and columns](https://fivetran.com/docs/getting-started/system-columns-and-tables)) are synced regardless of the `schema_change_handling` settings and configuration. You can only [disable non-locked columns in the system-enabled tables](#nestedblock--nonlocked). If the configuration specifies any system tables or locked system table columns as disabled ( `enabled = "false"`), the provider just ignores these statements.

## Usage examples

### Example for the ALLOW_ALL option

In `schema`,  you only need to specify schemas and columns you want to disable (`enabled = "false"`) and tables you want to disable or hash (`hashed = "true"`).

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
    enabled = "false"
  }
}
```

The configuration resulting from the example request is as follows:
- All new and existing schemas except `blocked_schema` are enabled
- All new and existing tables in the `schema_name` schema except the `blocked_table_name` table are enabled
- All new and existing columns in the`table_name` of the `schema_name` schema except the `blocked_column_name` column are enabled
- The `hashed_column_name` column is hashed in the `table_name` table in the `schema_name` schema
- All new schemas, tables, and columns are enabled once captured by the connector during the sync except those disabled by the system

### Example for the BLOCK_ALL option

All schemas, tables, and columns specified in `schema` are enabled by default (the default value for the `"enabled"` parameter is `true`).

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
    name = "enabled_schema_name"
  }
}
```

The configuration resulting from the example request is as follows:

- All new and existing schemas except  the `enabled_schema` and `schema_name` are disabled
- Only system-enabled tables and columns are enabled in the `enabled_schema_name` schema
- All new and existing tables in the `schema_name` schema except  the `enabled_table_name`, `table_name` tables and system tables are disabled
- All new and existing columns in the `table_name` table of the `schema_name` schema are disabled except the `hashed_column_name` column and system columns 
- The `hashed_column_name` column in the `table_name`  table the `schema_name` schema is hashed
- All new columns except the system-enabled columns, all schemas and tables are disabled once captured by the connector during the sync

### Example for the ALLOW_COLUMNS option

In `schema`,  you only need to specify schemas and tables you want to enable `enabled = "true"`) and columns you want to disable (`enabled = "false"`) or hash (`hashed = "true"`).

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
    name = "enabled_schema_name"
  }
}
```

The configuration resulting from the example request is as follows:

- All specified existing schemas and tables are enabled and all columns inside them are enabled by default, unless `enabled = "false"` is specified for the column
- All new and existing schemas except the `enabled_schema_name` and `schema_name` are disabled
- Only system-enabled tables and columns would be enabled in  the`enabled_schema_name` schema
- All new and existing tables in the `schema_name` schema except the `enabled_table_name`, `table_name` and system-enabled tables are disabled
- All new and existing columns in the`table_name` table of the `schema_name` schema except the `disabled_columns_name` and system-enabled columns are enabled
- The `hashed_column_name` would be hashed in table `table_name` in schema `schema_name`
- All new non system-enabled tables/schemas would be disabled once captured by connector on sync
- All new non system-enabled columns inside enabled tables (including system enabled-tables) would be enabled once captured by connector on sync

<a id="nestedblock--nonlocked"></a>
### Non-locked table column management in system-enabled tables

You cannot manage system-enabled tables, but you can manage its non-locked columns. For example, your schema `schema_name` has a system-enabled table `system_enabled_table` that can't be disabled, and you want to disable one of its columns named `columns_name`:

```hcl
resource "fivetran_connector_schema_config" "schema" {
  connector_id = "connector_id"
  schema_change_handling = "ALLOW_COLUMNS"
  schema {
    name = "schema_name"
    table {
      name = "system_enabled_table"
      column {
        name = "columns_name"
        enabled = "false"
      }
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `connector_id` (String) The unique identifier for the connector within the Fivetran system.
- `schema_change_handling` (String)

### Optional

- `schema` (Block Set) (see [below for nested schema](#nestedblock--schema))
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The unique resource identifier (equals to `connector_id`).

<a id="nestedblock--schema"></a>
### Nested Schema for `schema`

Required:

- `name` (String) The schema name within your destination in accordance with Fivetran conventional rules.

Optional:

- `enabled` (Boolean) The boolean value specifying whether the sync for the schema into the destination is enabled.
- `table` (Block Set) (see [below for nested schema](#nestedblock--schema--table))

<a id="nestedblock--schema--table"></a>
### Nested Schema for `schema.table`

Required:

- `name` (String) The table name within your destination in accordance with Fivetran conventional rules.

Optional:

- `column` (Block Set) (see [below for nested schema](#nestedblock--schema--table--column))
- `enabled` (Boolean) The boolean value specifying whether the sync of table into the destination is enabled.
- `sync_mode` (String) This field appears in the response if the connector supports switching sync modes for tables.

<a id="nestedblock--schema--table--column"></a>
### Nested Schema for `schema.table.column`

Required:

- `name` (String) The column name within your destination in accordance with Fivetran conventional rules.

Optional:

- `enabled` (Boolean) The boolean value specifying whether the sync of the column into the destination is enabled.
- `hashed` (Boolean) The boolean value specifying whether a column should be hashed.




<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

You don't need to import this resource as it is synthetic (doesn't create new instances in upstream).