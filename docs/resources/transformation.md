---
page_title: "Resource: fivetran_transformation"
---

# Resource: fivetran_transformation

Resource is in ALPHA state.

This resource allows you to add, manage and delete transformation projects in your account. 

## Example Usage for dbt Core Transformation

```hcl
resource "fivetran_transformation" "transformation" {
    provider = fivetran-provider

    type = "DBT_CORE"
    paused = true

    schedule {
        schedule_type = "TIME_OF_DAY"
        time_of_day = "11:00"
    }

    transformation_config {
        project_id = "project_id"
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

## Example Usage for Quickstart Transformation

```hcl
resource "fivetran_transformation" "transformation" {
    provider = fivetran-provider

    type = "QUICKSTART"
    paused = true

    schedule {
        schedule_type = "TIME_OF_DAY"
        time_of_day = "11:00"
    }

    transformation_config {
        package_name = "package_name"
        connection_ids = ["connection_id1", "connection_id2"]
        excluded_models = ["excluded_model1", "excluded_model2"]
    }
}
```

## Example Usages for Transformation Schedule section

```hcl
schedule {
    schedule_type = "TIME_OF_DAY"
    days_of_week = ["MONDAY", "FRIDAY"]
    time_of_day = "11:00"
}
```

```hcl
schedule {
    schedule_type = "INTEGRATED"
    connection_ids = ["connection_id1", "connection_id2"]
}
```

```hcl
schedule {
    schedule_type = "INTERVAL"
    interval = 601
}
```

```hcl
schedule {
    schedule_type = "CRON"
    cron = ["0 */1 * * *"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `paused` (Boolean) The field indicating whether the transformation will be set into the paused state. By default, the value is false.
- `schedule` (Block, Optional) (see [below for nested schema](#nestedblock--schedule))
- `transformation_config` (Block, Optional) (see [below for nested schema](#nestedblock--transformation_config))
- `type` (String) Transformation type. The following values are supported: DBT_CORE, QUICKSTART.

### Read-Only

- `created_at` (String) The timestamp of when the transformation was created in your account.
- `created_by_id` (String) The unique identifier for the User within the Fivetran system who created the transformation.
- `id` (String) The unique identifier for the Transformation within the Fivetran system.
- `output_model_names` (Set of String) Identifiers of related models.
- `status` (String) Status of transformation Project (NOT_READY, READY, ERROR).

<a id="nestedblock--schedule"></a>
### Nested Schema for `schedule`

Optional:

- `connection_ids` (Set of String) The list of the connection identifiers to be used for the integrated schedule. Not expected for QUICKSTART transformations
- `cron` (Set of String) Cron schedule: list of CRON strings. Used for for CRON schedule type
- `days_of_week` (Set of String) The set of the days of the week the transformation should be launched on. The following values are supported: MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY. Used for for INTEGRATED schedule type
- `interval` (Number) The time interval in minutes between subsequent transformation runs. Used for for INTERVAL schedule type
- `schedule_type` (String) The type of the schedule to run the Transformation on. The following values are supported: INTEGRATED, TIME_OF_DAY, INTERVAL, CRON.
- `smart_syncing` (Boolean) The boolean flag that enables the Smart Syncing schedule
- `time_of_day` (String) The time of the day the transformation should be launched at. Supported values are: "00:00", "01:00", "02:00", "03:00", "04:00", "05:00", "06:00", "07:00", "08:00", "09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00", "20:00", "21:00", "22:00", "23:00". Used for for TIME_OF_DAY schedule type


<a id="nestedblock--transformation_config"></a>
### Nested Schema for `transformation_config`

Optional:

- `connection_ids` (Set of String) The list of the connection identifiers to be used for the integrated schedule. Also used to identify package_name automatically if package_name was not specified
- `excluded_models` (Set of String) The list of excluded output model names
- `name` (String) The transformation name
- `package_name` (String) The Quickstart transformation package name
- `project_id` (String) The unique identifier for the dbt Core project within the Fivetran system
- `steps` (Attributes List) (see [below for nested schema](#nestedatt--transformation_config--steps))
- `upgrade_available` (Boolean) The boolean flag indicating that a newer version is available for the transformation package

<a id="nestedatt--transformation_config--steps"></a>
### Nested Schema for `transformation_config.steps`

Optional:

- `command` (String) The dbt command in the transformation step
- `name` (String) The step name

## Import

1. To import an existing `fivetran_transformation` resource into your Terraform state, you need to get **Transformation ID** via API call `GET https://api.fivetran.com/v1/transformations` to retrieve available projects.
2. Fetch transformation details for particular `transformation-id` using `GET https://api.fivetran.com/v1/transformations/{transformation-id}` to ensure that this is the transformation you want to import.
3. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_transformation" "my_imported_fivetran_transformation" {

}
```

4. Run the `terraform import` command:

```
terraform import fivetran_transformation.my_imported_fivetran_transformation {Transformation ID}
```

4. Use the `terraform state show` command to get the values from the state:

```
terraform state show 'fivetran_transformation.my_imported_fivetran_transformation'
```

5. Copy the values and paste them to your `.tf` configuration.

