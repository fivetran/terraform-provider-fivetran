----
page_title: "Dbt Project and Transformation Setup"
subcategory: "Getting Started"
---

# How to set up a dbt Project and Transformation schedule.

In this guide, we will set up a simple pipeline with one dbt Transformation using Fivetran Terraform Provider.

## Prerequisites

To create a project you need to have a group and destination. 

You can use existing ones, or configure a new ones using terraform:

```hcl
resource "fivetran_group" "group" {
    name = "MyGroup"
}
```

Once you have created the group, you need to associate a `Destination` with it:

```hcl
resource "fivetran_destination" "destination" {
    group_id = fivetran_group.group.id
    service = "postgres_rds_warehouse"
    time_zone_offset = "0"
    region = "GCP_US_EAST4"
    trust_certificates = "true"
    trust_fingerprints = "true"
    run_setup_tests = "true"

    config {
        host = "destination.host"
        port = 5432
        user = "postgres"
        password = "myPassword"
        database = "myDatabaseName"
        connection_type = "Directly"
    }
}
```

-> Note: you destination need to have `connected` status before dbt Project setup.

## Add `fivetran_dbt_project` resource.

Follow our [dbt Project setup guide](https://fivetran.com/docs/transformations/dbt/setup-guide#prerequisites) to complete prerequisites for project creation.
After that let's configure dbt Project resource:

```hcl
resource "fivetran_dbt_project" "project" {
    group_id = fivetran_destination.destination.id
    dbt_version = "1.3.2"
    threads = 1
    default_schema = "your_project_default_schema"
    type = "GIT"
    project_config {
        git_remote_url = "git@github.com:your_project_git_remote.git"
        git_branch = "main"
    }
}
```

Project creation and initialization takes time, so it's OK if resource creation takes 7-10 minutes.

## Configure your dbt Transformation schedule

You can configure your first Fivetran dbt Transformation with `fivetran_dbt_transformation` resource:

```hcl
resource "fivetran_dbt_transformation" "test_transformation" {
    dbt_project_id = fivetran_dbt_project.project.id
    dbt_model_name = "your_dbt_model_name"
    paused = false
    run_tests = false
    schedule {
        schedule_type = "INTERVAL"
        days_of_week = ["MONDAY"]
        interval = 60
    }
}
```

Above consfiguration will schedule model with name `your_dbt_model_name` on mondays each 60 minutes.

Now we are ready to apply our configuration:

```bash
terraform apply
```