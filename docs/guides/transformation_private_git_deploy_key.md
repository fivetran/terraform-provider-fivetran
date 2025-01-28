----
page_title: "Transformation Project Setup With Git Private Repo"
subcategory: "Getting Started"
---

# How to set up a Transformation Project with private Git Repo.

To be able to use private Transformation Project Git repository you have to grant Fivetran access to this repo.
To do that you need to add a Deploy Key to your repository. 
To get SSH key from Fivetran create `fivetran_transformation_project` resource:

```hcl
resource "fivetran_group" "my_group" {
    name = "My_Group"
}

resource "fivetran_transformation_project" "project" {
    provider = fivetran-provider
    group_id = "group_id"
    type = "DBT_GIT"
    run_tests = true

    project_config {
        git_remote_url = "git_remote_url"
        git_branch = "git_branch"
        folder_path = "folder_path"
        dbt_version = "dbt_version"
        default_schema = "default_schema"
        threads = 0
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
