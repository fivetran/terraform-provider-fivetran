---
page_title: "Resource: fivetran_destination"
---

# Resource: fivetran_destination

This resource allows you to create, update, and delete destinations.

## Example Usage

```hcl
resource "fivetran_destination" "dest" {
    group_id = fivetran_group.group.id
    service = "postgres_rds_warehouse"
    time_zone_offset = "0"
    region = "GCP_US_EAST4"
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
- `region` - Data processing location. This is where Fivetran will operate and run computation on data. See [Create destination](https://fivetran.com/docs/rest-api/destinations#payloadparameters) for details. Region also defines cloud service provider for your destination (GCP, AWS ar AZURE). 
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
- `role_arn` 
- `secret_key`
- `server_host_name` 
- `tunnel_host` 
- `tunnel_port` 
- `tunnel_user` 
- `user` 

## Import

To import an existing `fivetran_destination` resource into your terraform state you need to get `Destination Group ID` on the `Destination` page at the Fivetran Dashboard.
To retrieve existing groups use [Data Source: fivetran_groups](/docs/data-sources/groups).
Then define an empty resource in your .tf configuration:

```hcl
resource "fivetran_destination" "my_imported_destination" {

}
```

And call `terraform import` command with the following parameters:

```
terraform import fivetran_destination.my_imported_destination <your Destination Group ID>
```

Then copy-paste destination properties from the state to your .tf config, use `terraform state show`:

```
terraform state show 'fivetran_destination.my_imported_destination'
```

-> You need to get rid of redundant `config` properties that doesn't related to particular destination type - 
`config` in state contains all properties defined in schema, but you actually don't need to keep them all. 
Use [Fivetran public docs](https://fivetran.com/docs/rest-api/destinations/config) for reference to find the fields you need to keep in `config` section.