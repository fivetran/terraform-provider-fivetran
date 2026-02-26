----
page_title: "Modular Schema Management"
subcategory: "Preview"
---

# Modular Schema Management with Terraform

This guide demonstrates how to manage a Fivetran connection's schema configuration using the modular resource hierarchy. Instead of a single monolithic resource, schemas, tables, and columns are managed independently — making it easy to use `for_each` loops, split ownership across teams, and handle large schemas efficiently.

~> **Preview** The resources described in this guide are in preview. Their behavior and interface may change in future releases without prior notice.

## Overview

The modular schema management consists of four components:

| Component | Resource/Action | Manages |
|---|---|---|
| Schema reload | `fivetran_connection_schema_reload` (action) | Discovers schemas/tables/columns from source |
| Schema-level | `fivetran_connection_schemas_config` | Policy + which schemas are enabled/disabled |
| Table-level | `fivetran_connection_schema_tables_config` | Which tables are enabled/disabled + sync modes |
| Column-level | `fivetran_connection_table_columns_config` | Which columns are enabled/disabled + hashing + PKs |

The dependency chain ensures resources are applied in the correct order:

```
fivetran_connector (creates connection + triggers schema reload)
  └→ fivetran_connection_schemas_config (schema-level settings)
       └→ fivetran_connection_schema_tables_config (one per schema)
            └→ fivetran_connection_table_columns_config (one per table)
                 └→ fivetran_connector_schedule (unpause)
```

## How enabled/disabled lists work

Each resource uses two mutually exclusive list attributes. The list you choose defines the **complete desired state**:

- **`disabled_*`** — listed items are **disabled**, all unlisted items are **enabled**
- **`enabled_*`** — listed items are **enabled**, all unlisted items are **disabled**

For example, `disabled_schemas = ["staging"]` means "staging is disabled, every other schema is enabled." If a new schema appears upstream, it will automatically be enabled on the next apply.

Either list can be used with any `schema_change_handling` policy — the list choice is independent of the policy setting.

## Step 1: Create the connection and reload schema

The connection is always created in a paused state. Use a `lifecycle.action_trigger` to reload the schema immediately after creation, discovering all available schemas, tables, and columns.

```hcl
resource "fivetran_connector" "pg" {
  group_id = fivetran_group.my_group.id
  service  = "postgres"

  destination_schema {
    prefix = "postgres_rds"
  }

  run_setup_tests    = true
  trust_certificates = false
  trust_fingerprints = false

  config {
    host     = "db.example.com"
    port     = 5432
    user     = "fivetran"
    password = var.db_password
    database = "production"
  }

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.fivetran_connection_schema_reload.reload]
    }
  }
}

action "fivetran_connection_schema_reload" "reload" {
  config {
    connection_id = fivetran_connector.pg.id
    exclude_mode  = "PRESERVE"

    timeouts = {
      invoke = "30m"
    }
  }
}
```

## Step 2: Configure schema-level settings

Set the `schema_change_handling` policy and declare which schemas to enable or disable.

```hcl
resource "fivetran_connection_schemas_config" "config" {
  connection_id          = fivetran_connector.pg.id
  schema_change_handling = "ALLOW_ALL"

  # All schemas are enabled except these:
  disabled_schemas = ["staging", "temp"]

  depends_on = [fivetran_connector.pg]
}
```

## Step 3: Configure tables per schema using for_each

Use `locals` to centralize table configuration and `for_each` to create one resource per schema:

```hcl
locals {
  schema_table_configs = {
    public = {
      disabled_tables = ["legacy_orders", "tmp_imports"]
      sync_modes      = { "users" = "HISTORY", "orders" = "SOFT_DELETE" }
    }
    analytics = {
      disabled_tables = ["raw_debug_events"]
      sync_modes      = {}
    }
    reporting = {
      disabled_tables = ["scratch_pad"]
      sync_modes      = {}
    }
  }
}

resource "fivetran_connection_schema_tables_config" "schema" {
  for_each = local.schema_table_configs

  connection_id   = fivetran_connector.pg.id
  schema_name     = each.key
  disabled_tables = each.value.disabled_tables
  sync_mode       = length(each.value.sync_modes) > 0 ? each.value.sync_modes : null

  depends_on = [fivetran_connection_schemas_config.config]
}
```

## Step 4: Configure columns per table using for_each

Only define column-level resources for tables that need column management:

```hcl
locals {
  column_configs = {
    "public:users" = {
      schema           = "public"
      table            = "users"
      disabled_columns = ["internal_notes", "debug_flags"]
      hashed_columns   = ["email", "phone"]
      pk_columns       = ["id"]
    }
    "public:orders" = {
      schema           = "public"
      table            = "orders"
      disabled_columns = ["raw_payload"]
      hashed_columns   = null
      pk_columns       = null
    }
    "analytics:events" = {
      schema           = "analytics"
      table            = "events"
      disabled_columns = ["raw_data"]
      hashed_columns   = null
      pk_columns       = null
    }
  }
}

resource "fivetran_connection_table_columns_config" "table" {
  for_each = local.column_configs

  connection_id       = fivetran_connector.pg.id
  schema_name         = each.value.schema
  table_name          = each.value.table
  disabled_columns    = each.value.disabled_columns
  hashed_columns      = each.value.hashed_columns
  primary_key_columns = each.value.pk_columns

  depends_on = [fivetran_connection_schema_tables_config.schema]
}
```

-> **NOTE:** `hashed_columns` and `primary_key_columns` are optional. When set to `null`, the resource does not manage hashing or primary keys for that table. When configured, the resource owns the full setting — any external changes are detected as drift.

-> **NOTE:** Columns listed in `hashed_columns` or `primary_key_columns` must not be in `disabled_columns`.

## Step 5: Unpause the connection

After all schema configuration is applied, unpause the connection to start syncing:

```hcl
resource "fivetran_connector_schedule" "schedule" {
  connector_id   = fivetran_connector.pg.id
  sync_frequency = "360"
  paused         = "false"
  schedule_type  = "auto"

  depends_on = [fivetran_connection_table_columns_config.table]
}
```

## Drift detection

The modular resources detect drift at every level:

- A schema disabled externally → `disabled_schemas` grows, plan shows the change
- A new table appears disabled → `disabled_tables` grows, plan shows the change
- A column gets hashed externally (when `hashed_columns` is managed) → drift detected
- A managed item is dropped from the source → it disappears from the list, plan shows the change

Items that are not managed by the resource (e.g., a table not in either list, or `hashed_columns` not configured) do not trigger drift.

## Concurrency and conflict handling

All schema-modifying resources for the same `connection_id` are serialized using a per-connection mutex to prevent API conflicts. If a 409 Conflict occurs (e.g., from an external API call), the resource automatically retries with exponential backoff up to 5 times.

## Apply configuration

```bash
terraform apply
```

Terraform will execute the resources in dependency order: create connection → reload schema → configure schemas → configure tables → configure columns → unpause.
