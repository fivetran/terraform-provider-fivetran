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

- `id` - The unique identifier for the destination within your Fivetran account.

### Read-Only

- `config` - see [below for nested schema](#nestedatt--config)
- `group_id` 
- `region` 
- `service` 
- `setup_status` 
- `time_zone_offset` 

<a id="nestedatt--config"></a>
### Nested Schema for `config`

Read-Only:

- `auth` 
- `auth_type` 
- `bucket` 
- `cluster_id`
- `cluster_region`
- `connection_type` 
- `create_external_tables` 
- `data_set_location` 
- `database` 
- `external_location` 
- `host` 
- `http_path` 
- `password` 
- `personal_access_token` 
- `port`
- `project_id` 
- `public_key` 
- `role_arn` 
- `secret_key`
- `server_host_name` 
- `tunnel_host` 
- `tunnel_port` 
- `tunnel_user` 
- `user` 
