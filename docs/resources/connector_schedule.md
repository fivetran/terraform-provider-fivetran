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

- `connector_id` - The of the connector you want to schedule.

### Optional
- `sync_frequency` - The connector sync frequency in minutes. The supported values are: `5`, `15`, `30`, `60`, `120`, `180`, `360`, `480`, `720`, `1440`. Default value is `360`.
- `daily_sync_time` - Defines the sync start time when the sync frequency is already set or being set by the current request to 1440. If not specified, we will use the baseline sync start time. This parameter has no effect on the 0 to 60 minutes offset used to determine the actual sync start time. Supported values: `00:00` | `01:00` | `02:00` | `03:00` | `04:00` | `05:00` | `06:00` | `07:00` | `08:00` | `09:00` | `10:00` | `11:00` | `12:00` | `13:00` | `14:00` | `15:00` | `16:00` | `17:00` | `18:00` | `19:00` | `20:00` | `21:00` | `22:00` | `23:00`
- `paused` - Specifies whether the connector is paused. Default value: `true`.
- `pause_after_trial` - Specifies whether the connector should be paused after the free trial period has ended. Default value: `true`.
- `schedule_type` - The connector schedule config type. Supported values: `auto`, `manual`. Lets you disable or enable an automatic data sync on a schedule. If you set this parameter to `manual`, the automatic data sync will be disabled, but you will be able to trigger the data sync using the [Sync Connector Data](https://fivetran.com/docs/rest-api/connectors#syncconnectordata) endpoint.  Default value: `auto`.

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
