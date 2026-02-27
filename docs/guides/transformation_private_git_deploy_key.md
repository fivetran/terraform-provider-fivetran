----
page_title: "Transformation Project Setup With Git Private Repo"
subcategory: "Getting Started"
---

# How to set up a Transformation Project with private Git Repo.

To be able to use private Transformation Project Git repository you have to grant Fivetran access to this repo.
To do that you need to add a Deploy Key to your repository.
To get SSH key from Fivetran create `fivetran_transformation_project` resource, then set up the public key as a deploy key in your repo. You can optionally use the `fivetran_transformation_project_run_tests` action (Terraform 1.14+) to validate the project setup automatically.

```hcl
resource "fivetran_group" "my_group" {
    name = "My_Group"
}

resource "fivetran_transformation_project" "project" {
    provider = fivetran-provider
    group_id = fivetran_group.my_group.id
    type     = "DBT_GIT"
    run_tests = false

    project_config {
        git_remote_url = "git_remote_url"
        git_branch     = "git_branch"
        folder_path    = "folder_path"
        dbt_version    = "dbt_version"
        default_schema = "default_schema"
        threads        = 0
        target_name    = "target_name"
        environment_vars = ["environment_var"]
    }

    lifecycle {
        action_trigger {
            action = fivetran_transformation_project_run_tests.project_tests
            events = ["after_update"]
        }
    }
}

# GitHub example — for Bitbucket use bitbucket_deploy_key instead
resource "github_repository_deploy_key" "deploy_key" {
    title      = "Fivetran deploy key"
    repository = "repo-owner/repo-name"
    key        = fivetran_transformation_project.project.project_config.public_key
    read_only  = true

    lifecycle {
        action_trigger {
            action = fivetran_transformation_project_run_tests.project_tests
            events = ["after_create"]
        }
    }
}

action "fivetran_transformation_project_run_tests" "project_tests" {
    project_id            = fivetran_transformation_project.project.id
    fail_on_tests_failure = true
}
```

Setting `fail_on_tests_failure = false` will report test failures as warnings instead of errors, allowing Terraform to continue with the rest of the plan.

Since we recommend using third-party providers in this case, please make sure that access to the repositories is provided correctly and the providers are configured correctly for connection.
