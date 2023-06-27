---
page_title: "Resource: fivetran_connector_schedule"
---

# Resource: fivetran_connector_swchedule

This resource allows you to manage connectors schedule: pause/unpause connector, set daily_sync_time and sync_frequency.

## Example Usage

```hcl
resource "fivetran_connector_schedule" "my_connector_schedule" {
    connector_id = fivetran_connector.my_connector.id

    sync_frequency     = "1440"
    daily_sync_time    = "03:00"

    paused             = false
    pause_after_trial  = true

    schedule_type      = "auto"
}
```

## Schema

### Required

- `connector_id` (String) The unique identifier for the connector

### Optional

- `daily_sync_time` (String) The optional parameter that defines the sync start time when the sync frequency is already set or being set by the current request to 1440. It can be specified in one hour increments starting from 00:00 to 23:00. If not specified, we will use [the baseline sync start time](https://fivetran.com/docs/getting-started/syncoverview#syncfrequencyandscheduling). This parameter has no effect on the [0 to 60 minutes offset](https://fivetran.com/docs/getting-started/syncoverview#syncstarttimesandoffsets) used to determine the actual sync start time
- `pause_after_trial` (String) Specifies whether the connector should be paused after the free trial period has ended
- `paused` (String) Specifies whether the connector is paused
- `schedule_type` (String) The connector schedule configuration type. Supported values: auto, manual
- `sync_frequency` (String) The connector sync frequency in minutes

### Read-Only

- `id` (String) The unique identifier for the user within the account.

## Import

1. To import an existing `fivetran_connector_schedule` resource into your Terraform state, you need to get **Fivetran Connector ID** on the **Setup** tab of the connector page in your Fivetran dashboard.

2. Retrieve all connectors in a particular group using the [fivetran_group_connectors data source](/docs/data-sources/group_connectors). To retrieve existing groups, use the [fivetran_groups data source](/docs/data-sources/groups).

3. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_connector_schedule" "my_imported_connector_schedule" {
    connector_id = "<your_fivetran_connector_id>"
}
```

4. Run the `terraform import` command:

```
terraform import fivetran_connector_schedule.my_imported_connector_schedule {your Fivetran Connector ID}
```

5.  Use the `terraform state show` command to get the values from the state:

```
terraform state show 'fivetran_connector_schedule.my_imported_connector_schedule'
```

6. Copy the field values and paste them to your `.tf` configuration.