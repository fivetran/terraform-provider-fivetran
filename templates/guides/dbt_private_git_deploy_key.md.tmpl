----
page_title: "Dbt Project Setup With Git Private Repo"
subcategory: "Getting Started"
---

# How to set up a dbt Project with private Git Repo.

To be able to use private dbt Project Git repository you have to grant Fivetran access to this repo.
To do that you need to add a Deploy Key to your repository. 
To get SSH key from Fivetran create `fivetran_dbt_project` resource:

```hcl
resource "fivetran_group" "my_group" {
    name = "My_Group"
}

resource "fivetran_dbt_project" "project" {
    group_id = fivetran_group.my_group.id
    dbt_version = "1.3.2"
    threads = 1
    default_schema = "your_project_default_schema"
    type = "GIT"
}
```

Then you need to set up the dbt Project public key (field `public_key` in created resource) as a deploy key into your repo using [GitHub Provider Repository Deploy Key Resource](https://registry.terraform.io/providers/integrations/github/latest/docs/resources/repository_deploy_key):

```hcl
resource "github_repository_deploy_key" "example_repository_deploy_key" {
  title      = "Repository test key"
  repository = "your-repo"
  key        = fivetran_dbt_project.project.public_key
  read_only  = true
}
```

And after that you can configure your project in `fivetran_dbt_git_project_config` resource:

```hcl
resource "fivetran_dbt_git_project_config" "project_config" {
    id = fivetran_dbt_project.project.id
    
    git_remote_url = "git@github.com:your-repo.git"
    git_branch = "main"
}
```

