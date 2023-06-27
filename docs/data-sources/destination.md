---
page_title: "Data Source: fivetran_destination"
---

# Data Source: fivetran_destination

This data source returns a destination object.

## Example Usage

```hcl
data "fivetran_destination" "dest" {
    id = "anonymous_mystery"
}
```

## Schema

### Required

- `id` (String) The unique identifier for the destination within the Fivetran system

### Read-Only

- `config` (Set of Object) (see [below for nested schema](#nestedatt--config))
- `group_id` (String) The unique identifier for the Group within the Fivetran system.
- `region` (String) Data processing location. This is where Fivetran will operate and run computation on data.
- `service` (String) The connector type name within the Fivetran system
- `setup_status` (String) Destination setup status
- `time_zone_offset` (String) Determines the time zone for the Fivetran sync schedule.

<a id="nestedatt--config"></a>
### Nested Schema for `config`

Read-Only:

- `auth` (String)
- `auth_type` (String)
- `bucket` (String)
- `catalog` (String)
- `cluster_id` (String)
- `cluster_region` (String)
- `connection_type` (String)
- `create_external_tables` (String)
- `data_set_location` (String)
- `database` (String)
- `external_location` (String)
- `host` (String)
- `http_path` (String)
- `is_private_key_encrypted` (String)
- `passphrase` (String)
- `password` (String)
- `personal_access_token` (String)
- `port` (Number)
- `private_key` (String)
- `project_id` (String)
- `public_key` (String)
- `role` (String)
- `role_arn` (String)
- `secret_key` (String)
- `server_host_name` (String)
- `tunnel_host` (String)
- `tunnel_port` (String)
- `tunnel_user` (String)
- `user` (String)
