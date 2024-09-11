----
page_title: "Version Update 1.3.0"
subcategory: "Upgrade Guides"
---

# Version 1.3.0

## What's new in 1.3.0

In version `1.3.0` of Fivetran Terraform provider, resource `fivetran_dbt_project` behavior changed:
- installation of the DBT project configuration should now occur in a separate resource `fivetran_dbt_git_project_config`, after installing the key in the repository

## Migration guide

### Provider 

Update your provider configuration in the following way:

Previous configuration:

```hcl
required_providers {
   fivetran = {
     version = "~> 1.2.8"
     source  = "fivetran/fivetran"                
   }
 }
```

Updated configuration:

```hcl
required_providers {
   fivetran = {
     version = ">= 1.3.0"
     source  = "fivetran/fivetran"                
   }
 }
```

### Resource `fivetran_dbt_project`

Update all your connector schema config resources (`fivetran_dbt_project`):

Previous configuration:

```hcl
resource "fivetran_dbt_project" "test_project" {
  provider = fivetran-provider
  group_id = fivetran_destination.test_destination.id
  dbt_version = "1.0.1"
  threads = 1
  default_schema = "dbt_demo_test_e2e_terraform"
  type = "GIT"
  project_config {
    folder_path = "/folder/path"
    git_remote_url = "git@github.com:fivetran/dbt_demo.git"
    git_branch = "main"
  }
}
```

Updated configuration:

```hcl
resource "fivetran_dbt_project" "test_project" {
  provider = fivetran-provider
  group_id = fivetran_destination.test_destination.id
  dbt_version = "1.0.1"
  threads = 1
  default_schema = "dbt_demo_test_e2e_terraform"
  type = "GIT"
}


For GitHub based repositories
```hcl
resource "github_repository_deploy_key" "example_repository_deploy_key" {
  title      = "Repository test key"
  repository = "fivetran/dbt_demo"
  key        = fivetran_dbt_project.test_project.public_key
  read_only  = true
}
```

For Bitbucket based repositories
```hcl
resource "bitbucket_deploy_key" "test" {
  workspace  = "fivetran"
  repository = "dbt_demo"  
  key        = fivetran_dbt_project.test_project.public_key
  label      = "Repository test key"
}
```

```hcl
resource "fivetran_dbt_git_project_config" "test_project_config" {
  project_id = fivetran_dbt_project.test_project.id

  folder_path = "/folder/path"
  git_remote_url = "git@github.com:fivetran/dbt_demo.git"
  git_branch = "main"
}

```

### Update terraform state

Once all configurations have been updated, run:

```
terraform init -upgrade
```