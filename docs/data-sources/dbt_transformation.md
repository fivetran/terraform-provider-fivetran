---
page_title: "Data Source: fivetran_dbt_transformation"
---

# Data Source: fivetran_dbt_transformation

This data source returns a dbt Transformation object.

## Example Usage

```hcl
data "fivetran_dbt_transformation" "transformation" {
    id = "transformation_id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `connector_ids` (Set of String) Identifiers of related connectors.
- `created_at` (String) The timestamp of the dbt Transformation creation.
- `dbt_model_id` (String) The unique identifier for the dbt Model within the Fivetran system.
- `dbt_model_name` (String) Target dbt Model name.
- `dbt_project_id` (String) The unique identifier for the dbt Project within the Fivetran system.
- `id` (String) The ID of this resource.
- `model_ids` (Set of String) Identifiers of related models.
- `output_model_name` (String) The dbt Model name.
- `paused` (Boolean) The field indicating whether the transformation will be created in paused state. By default, the value is false.
- `run_tests` (Boolean) The field indicating whether the tests have been configured for dbt Transformation. By default, the value is false.
- `schedule` (List of Object) dbt Transformation schedule parameters. (see [below for nested schema](#nestedatt--schedule))

<a id="nestedatt--schedule"></a>
### Nested Schema for `schedule`

Read-Only:

- `days_of_week` (Set of String)
- `interval` (Number)
- `schedule_type` (String)
- `time_of_day` (String)