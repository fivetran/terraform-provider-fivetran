----
page_title: "Dbt Project Setup With Git Private Repo"
subcategory: "Getting Started"
---

# How to set up a dbt Project with private Git Repo.

To be able to use private dbt Progect Git repository you have to grant Fivetran access to this repo.
To do that you need to add a Deploy Key to your repository. To get SSH key from Fivetran use `fivetran_group_ssh_key` datasource:

```hcl
resource "fivetran_group" "my_group" {
    name = "My_Group"
}

data "fivetran_group_ssh_key" "my_group_public_key" {
    id = fivetran_group.my_group.id
}
```

Then you need to set up the group SSH key as a deploy key into your repo using [GitHub Provider Repository Deploy Key Resource](https://registry.terraform.io/providers/integrations/github/latest/docs/resources/repository_deploy_key):

```hcl
resource "github_repository_deploy_key" "example_repository_deploy_key" {
  title      = "Repository test key"
  repository = "your-repo"
  key        = data.fivetran_group_ssh_key.my_group_public_key.public_key
  read_only  = true
}
```

And after that you can configure your project:

```hcl
resource "fivetran_dbt_project" "project" {
    group_id = fivetran_group.my_group.id
    dbt_version = "1.3.2"
    threads = 1
    default_schema = "your_project_default_schema"
    type = "GIT"
    project_config {
        git_remote_url = "git@github.com:your-repo.git"
        git_branch = "main"
    }
}
```

