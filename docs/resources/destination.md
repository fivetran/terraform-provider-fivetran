---
page_title: "Resource: fivetran_destination"
---

# Resource: fivetran_destination

This resource allows you to create, update, and delete destinations.

## Example

```hcl
resource "fivetran_destination" "dest" {
    group_id = fivetran_group.group.id
    service = "postgres_rds_warehouse"
    time_zone_offset = "0"
    region = "EU"
    trust_certificates = "true"
    trust_fingerprints = "true"
    run_setup_tests = "true"

    config {
        host = "destination.fqdn"
        port = 5432
        user = "postgres"
        password = "myPass"
        database = "fivetran"
        connection_type = "Directly"
    }
}
```

## Schema

### Required

- `config` - Destination setup configuration. The format is specific for each destination. (see [below for nested schema](#nestedblock--config))
- `group_id` - The unique identifier for the group within the Fivetran system.
- `region` - Data processing location. This is where Fivetran will operate and run computation on data.
- `run_setup_tests` - Specifies whether setup tests should be run automatically.
- `service` - The name for the destination type within the Fivetran system.
- `time_zone_offset` - Determines the time zone for the Fivetran sync schedule.

### Optional

- `trust_certificates` - Specifies whether we should trust the certificate automatically.
- `trust_fingerprints` - Specifies whether we should trust the SSH fingerprint automatically.

### Read-Only

- `id` 
- `last_updated` 
- `setup_status`

<a id="nestedblock--config"></a>
### Nested Schema for `config`

See [Destination Config](https://fivetran.com/docs/rest-api/destinations/config) for details.

Optional:

- `auth`
- `auth_type` 
- `bucket` 
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
- `role_arn` 
- `server_host_name` 
- `tunnel_host` 
- `tunnel_port` 
- `tunnel_user` 
- `user` 
