----
page_title: "Version Update 1.5.0"
subcategory: "Upgrade Guides"
---

# Version 1.5.0

## What's new in 1.5.0

In version `1.5.0` of Fivetran Terraform provider, we have implemented new resources for managing Transformations:

## Migration guide

### Provider 

Update your provider configuration in the following way:

Previous configuration:

```hcl
required_providers {
   fivetran = {
     version = "~> 1.4.2"
     source  = "fivetran/fivetran"                
   }
 }
```

Updated configuration:

```hcl
required_providers {
   fivetran = {
     version = ">= 1.5.0"
     source  = "fivetran/fivetran"                
   }
 }
```

### Resource `fivetran_dbt_project`

Replace all your resources `fivetran_dbt_project` with `fivetran_transformation_project`

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
    git_remote_url = "git@github.com:fivetran/repo-name.git"
    git_branch = "main"
  }
}
```

Updated configuration:

```hcl
resource "fivetran_transformation_project" "project" {
    provider = fivetran-provider
    group_id = "group_id"
    type = "DBT_GIT"
    run_tests = true

    project_config {
        git_remote_url = "git@github.com:fivetran/repo-name.git"
        git_branch = "main"
        folder_path = "/folder/path"
        dbt_version = "1.0.1"
        default_schema = "dbt_demo_test_e2e_terraform"
        threads = 1
        target_name = "target_name"
        environment_vars = ["environment_var"]
    }
}
```

Then you need to set up the Transformation Project public key (field `public_key` in created resource) as a deploy key into your repo using:

[GitHub Provider Repository Deploy Key Resource](https://registry.terraform.io/providers/integrations/github/latest/docs/resources/repository_deploy_key):
```hcl
resource "github_repository_deploy_key" "example_repository_deploy_key" {
  title      = "Repository test key"
  repository = "repo-owner/repo-name"
  key        = fivetran_transformation_project.test_project.project_config.public_key
  read_only  = true
}
```

or

[Bitbucket Provider Repository Deploy Key Resource]https://registry.terraform.io/providers/DrFaust92/bitbucket/latest/docs/resources/deploy_key)
```hcl
resource "bitbucket_deploy_key" "test" {
  workspace  = "repo-owner"
  repository = "repo-name"  
  key        = fivetran_transformation_project.test_project.project_config.public_key
  label      = "Repository test key"
}
```

Since we recommend using third-party providers in this case, please make sure that access to the repositories is provided correctly and the providers are configured correctly for connection.

### Resource `fivetran_dbt_transformation`

Replace all your resources `fivetran_dbt_transformation` with `fivetran_transformation`

Previous configuration:

```hcl
resource "fivetran_dbt_transformation" "transformation" {
    dbt_model_name = "dbt_model_name"
    dbt_project_id = "dbt_project_id"
    run_tests = "false"
    paused = "false"
    schedule {
        schedule_type = "TIME_OF_DAY"
        time_of_day = "12:00"
        days_of_week = ["MONDAY", "SATURDAY"]
    }
}
```

Updated configuration:

```hcl
resource "fivetran_transformation" "transformation" {
    provider = fivetran-provider

    type = "DBT_CORE"
    paused = false

    schedule {
        cron = ["cron1","cron2"]
        interval = 60
        smart_syncing = true
        connection_ids = ["connection_id1", "connection_id2"]
        schedule_type = "TIME_OF_DAY"
        days_of_week = ["MONDAY", "SATURDAY"]
        time_of_day = "14:00"
    }

    transformation_config {
        project_id = "dbt_project_id"
        name = "name"
        steps = [
            {
                name = "name1"
                command = "command1"
            },
            {
                name = "name2"
                command = "command2"
            }
        ]
    }
}
```

### Datasources `fivetran_dbt_project`, `fivetran_dbt_projects`, `fivetran_dbt_transformation`, `fivetran_dbt_models`

Replace datasources:
- `fivetran_dbt_project` with `fivetran_transformation_project`
- `fivetran_dbt_projects` with `fivetran_transformation_projects`
- `fivetran_dbt_transformation` with `fivetran_transformation`
Remove datasource `fivetran_dbt_models`

### Update terraform state

Once all configurations have been updated, run:

```
terraform init -upgrade
```