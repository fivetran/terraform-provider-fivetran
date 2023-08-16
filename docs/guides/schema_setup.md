----
page_title: "Connector Schema Setup"
subcategory: "Getting Started"
---

# How to set up Fivetran connector schema config using Terraform

In this guide, we will set up a simple pipeline with one connector and schema using Fivetran Terraform Provider. 

## Create a connector resource

Create the `fivetran_connector` resource:

```hcl
resource "fivetran_connector" "connector" {
   ...
   run_setup_tests = "true" # it is necessary to authorise connector
}
```

Connector will be in the paused state, but ready to sync.

-> Connector should be **authorized** to be able to fetch schema from source. Set `run_setup_tests = "true"`.

## Set up connector schema config

Let's define what exactly we want to sync by using the `fivetran_connector_schema_config` resource:

```hcl
resource "fivetran_connector_schema_config" "connector_schema" {
  connector_id = fivetran_connector.connector.id
  schema_change_handling = "BLOCK_ALL"
  schema {
    name = "my_fivetran_log_connector"
    table {
      name = "log"
      column {
        name = "event"
        enabled = "true"
      }
      column {
        name = "message_data"
        enabled = "true"
      }
      column {
        name = "message_event"
        enabled = "true"
      }
      column {
        name = "sync_id"
        enabled = "true"
      }
    }
  }
  # before applying schema resource will trigger "Reload connector schema config" endpoint
  # it could take time for slow sources or for source with huge connector_schema_setup
  # to prevent timeouts you can set custom timeouts
  timeouts {
      create = "6h"
      read   = "6h"
      update = "6h"
  }
  # if you not sure in timing you can set timeouts to 0 - it means `no timeout`
  # WARNING: not recommended - this could lead to unpredictable apply process hanging
  #timeouts {
  #    create = "0"
  #    read   = "0"
  #    update = "0"
  #}
}
```

## Set up connector schedule configuration

-> The schedule should depend on the schema resource to enable the connector **after** the schema changes are applied.

```hcl
resource "fivetran_connector_schedule" "my_connector_schedule" {
    connector_id = fivetran_connector_schema_config.connector_schema.id

    sync_frequency     = "5"

    paused             = false
    pause_after_trial  = true

    schedule_type      = "auto"
}
```

## Apply configuration

```bash
terraform apply
```

## Example configuration

An example .tf file with the configuration could be found [here](https://github.com/fivetran/terraform-provider-fivetran/tree/main/config-examples/connector_schema_setup.tf).