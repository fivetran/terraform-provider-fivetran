----
page_title: "Version Update 0.7.0"
subcategory: "Upgrade Guides"
---

# Version 0.7.0

## What's new in 0.7.0

In version `0.7.0` of Fivetran Terraform provider, resource `fivetran_connector` is separated onto two resources:
- `fivetran_connector` resource
- `fivetran_connector_schedule` resource
With this new structure, it's now possible to create a connector, define the schema config for it, and enable it in one `apply` cycle without intermediate stages.
Before this version, you had to "un-pause" connector after applying initial schema configuration with additional `apply` to avoid unneeded data to be synced.

## Migration guide

### Provider 

Update your provider configuration in the following way:

Previous configuration:

```hcl
required_providers {
   fivetran = {
     version = "~> 0.6.19"
     source  = "fivetran/fivetran"                
   }
 }
```

Updated configuration:

```hcl
required_providers {
   fivetran = {
     version = ">= 0.7.0"
     source  = "fivetran/fivetran"                
   }
 }
```

### Resource `fivetran_connector`

Update all your connector resources (`fivetran_connector`):

Previous configuration:

```hcl
resource "fivetran_connector" "test_connector" {

  group_id  = "worker_tennis"
  service   = "fivetran_log"

  destination_schema {
    name = "fivetran_log_schema"
  }

  sync_frequency     = "1440"
  daily_sync_time    = "6:00"
  paused             = false
  pause_after_trial  = false

  run_setup_tests    = true
  config {
    group_name = "worker_tennis"
  }
}
```

Updated configuration:

```hcl
resource "fivetran_connector" "test_connector" {
  group_id  = "worker_tennis"
  service   = "fivetran_log"

  destination_schema {
    name = "fivetran_log_schema"
  }

  run_setup_tests    = true

  config {
    group_name = "worker_tennis"
  }
}
resource "fivetran_connector_schedule" "test_connector_schedule" {
  connector_id       = fivetran_connector.test_connector.id

  sync_frequency     = "1440"
  daily_sync_time    = "6:00"
  paused             = false
  pause_after_trial  = false

  schedule_type      = "auto"
}

```

### Update terraform state

Once all configurations have been updated, run:

```
terraform init -upgrade
```