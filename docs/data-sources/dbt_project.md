---
page_title: "Data Source: fivetran_dbt_project"
---

# Data Source: fivetran_dbt_project

This data source returns a dbt Project object.

## Example Usage

```hcl
data "fivetran_dbt_project" "project" {
    id = "project_id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) The unique identifier for the dbt Project within the Fivetran system.

### Read-Only

- `created_at` (String) The timestamp of the dbt Project creation.
- `created_by_id` (String) The unique identifier for the User within the Fivetran system who created the dbt Project.
- `dbt_version` (String) The version of dbt that should run the project. We support the following versions: 0.18.0 - 0.18.2, 0.19.0 - 0.19.2, 0.20.0 - 0.20.2, 0.21.0 - 0.21.1, 1.0.0, 1.0.1, 1.0.3 - 1.0.9, 1.1.0 - 1.1.3, 1.2.0 - 1.2.4, 1.3.0 - 1.3.2, 1.4.1.
- `default_schema` (String) Default schema in destination. This production schema will contain your transformed data.
- `environment_vars` (Set of String)
- `group_id` (String) The unique identifier for the group within the Fivetran system.
- `models` (Block Set) The collection of dbt Models. (see [below for nested schema](#nestedblock--models))
- `project_config` (List of Object) Type specific dbt Project configuration parameters. (see [below for nested schema](#nestedatt--project_config))
- `public_key` (String) Public key to grant Fivetran SSH access to git repository.
- `status` (String) Status of dbt Project (NOT_READY, READY, ERROR).
- `target_name` (String) Target name to set or override the value from the deployment.yaml
- `threads` (Number) The number of threads dbt will use (from 1 to 32). Make sure this value is compatible with your destination type. For example, Snowflake supports only 8 concurrent queries on an X-Small warehouse.
- `type` (String) Type of dbt Project. Currently only `GIT` supported. Empty value will be considered as default (GIT).

<a id="nestedblock--models"></a>
### Nested Schema for `models`

Read-Only:

- `id` (String) The unique identifier for the dbt Model within the Fivetran system.
- `model_name` (String) The dbt Model name.
- `scheduled` (Boolean) Boolean specifying whether the model is selected for execution.


<a id="nestedatt--project_config"></a>
### Nested Schema for `project_config`

Read-Only:

- `folder_path` (String)
- `git_branch` (String)
- `git_remote_url` (String)