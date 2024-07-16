----
page_title: "Version Update 1.1.18"
subcategory: "Upgrade Guides"
---

# Version 1.1.18

## What's new in 1.1.18

In version `1.1.18` of Fivetran Terraform provider, resource `fivetran_connector_schema_config` behavior changed:
- If no columns settings specified in `table.columns` no settings will be applied. If table enabled - columns won't be blocked automatically by `BLOCK_ALL` policy.
- Settings for sub-elements won't be managed if root element disabled: for `BLOCK_ALL` policy for disabled schema no settings for tables/columns will be applied.

## Migration guide

### Provider 

Update your provider configuration in the following way:

Previous configuration:

```hcl
required_providers {
   fivetran = {
     version = "~> 1.1.17"
     source  = "fivetran/fivetran"                
   }
 }
```

Updated configuration:

```hcl
required_providers {
   fivetran = {
     version = ">= 1.1.18"
     source  = "fivetran/fivetran"                
   }
 }
```

### Resource `fivetran_connector_schema_config`

Update all your connector schema config resources (`fivetran_connector_schema_config`):

Previous configuration:

```hcl
resource "fivetran_connector_schema_config" "test_schema" {
  connector_id = "connector_id"
  schema_change_handling = "ALLOW_ALL"
  
  schema {
    name = "schema_name"
    table {
      name = "table_name"
      sync_mode = "HISTORY"
      column {
        name = "hashed_column_name"
        hashed = "true"
      }
    }
  }
}
```

Updated configuration:

```hcl
resource "fivetran_connector_schema_config" "test_schema" {
  connector_id = "connector_id"
  schema_change_handling = "ALLOW_ALL"

  schemas = {
    "schema_name" = {
      tables = {
        "table_name" = {
          sync_mode = "HISTORY"
          columns = {
            "hashed_column_name" = {
              hashed = true
            }
          }
        }
      }
    }
  }
}

```

### Update terraform state

Once all configurations have been updated, run:

```
terraform init -upgrade
```