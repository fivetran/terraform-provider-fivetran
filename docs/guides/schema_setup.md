---
page_title: "Connector Schema Setup"
subcategory: "Getting Started"
---

# How to set up Fivetran connector schema config using Terraform

In this guide we will set up simple pipeline with one connector and schema using Fivetran Terraform Provider. 

## Initial setup

Please follow our [Initial Setup guide](https://registry.terraform.io/providers/fivetran/fivetran/latest/docs/resources/connector_schema_config) with one minor diff - connector should be in paused state.

-> We have to create *paused* connector to avoid syncing unwanted data before schema config applied

```hcl
resource "fivetran_connector" "connector" {
    ...
    # connector should be paused on first apply
    paused = true 
    ...
}
```

If we apply such configuration - connector will be in paused state, but ready to sync. 

## Set up connector schema config

Let's define what exactly we want to sync using `fivetran_connector_schema_config` resource:

```hcl
resource "fivetran_connector_schema_config" "connector_schema" {
  connector_id = "fivetran_connector.connector.id"
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
}
```

Now we are ready to apply our configuration:

```bash
terraform apply
```

After schema configuration applied we can un-pause our connector:

```hcl
resource "fivetran_connector" "connector" {
    ...

    paused = true 

   ...
}
```

```bash
terraform apply
```

## Example configuration

Example .tf file with configuration could be found [here](https://github.com/fivetran/terraform-provider-fivetran/tree/main/config-examples/connector_schema_setup.tf).