----
page_title: "Version Update 1.4.0"
subcategory: "Upgrade Guides"
---

# Version 1.4.0

## What's new in 1.4.0

In version `1.4.0` of Fivetran Terraform provider, resource `fivetran_local_processing_agent` renamed to `fivetran_hybrid_deployment_agent`

## Migration guide

### Provider 

Update your provider configuration in the following way:

Previous configuration:

```hcl
required_providers {
   fivetran = {
     version = "~> 1.3.0"
     source  = "fivetran/fivetran"                
   }
 }
```

Updated configuration:

```hcl
required_providers {
   fivetran = {
     version = ">= 1.4.0"
     source  = "fivetran/fivetran"                
   }
 }
```

### Resource `fivetran_hybrid_deployment_agent`

Update all your local processing agent resources (`fivetran_local_processing_agent`):

Previous configuration:

```hcl
resource "fivetran_local_processing_agent" "test_agent" {
}
```

Updated configuration:

```hcl
resource "fivetran_hybrid_deployment_agent" "test_agent" {
}
```

### Resource `fivetran_connector`

Update all your connector resources (`fivetran_connector`):

Previous configuration:

```hcl
resource "fivetran_connector" "test_connector" {
  local_processing_agent_id = agent_id
}
```

Updated configuration:

```hcl
resource "fivetran_connector" "test_connector" {
  hybrid_deployment_agent_id = agent_id
}
```

### Resource `fivetran_destination`

Update all your destination resources (`fivetran_destination`):

Previous configuration:

```hcl
resource "fivetran_destination" "test_destination" {
  local_processing_agent_id = agent_id
}
```

Updated configuration:

```hcl
resource "fivetran_destination" "test_destination" {
  hybrid_deployment_agent_id = agent_id
}
```

### Update terraform state

Once all configurations have been updated, run:

```
terraform init -upgrade
```