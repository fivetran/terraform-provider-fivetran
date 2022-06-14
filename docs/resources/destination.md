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

1. To import an existing `fivetran_destination` resource into your Terraform state, you need to get **Destination Group ID** on the destination page in your Fivetran dashboard.
To retrieve existing groups, use the [fivetran_groups data source](/docs/data-sources/groups).
2. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_destination" "my_imported_destination" {

}
```

3. Run the `terraform import` command with the following parameters:

```
terraform import fivetran_destination.my_imported_destination <your Destination Group ID>
```

4. Use the `terraform state show` command to get the values from the state:

```
terraform state show 'fivetran_destination.my_imported_destination'
```
5. Copy the values and paste them to your `.tf` configuration.

-> The `config` object in the state contains all properties defined in the schema. You need to remove properties from the `config` that are not related to destinations. See the [Fivetran REST API documentation](https://fivetran.com/docs/rest-api/destinations/config) for reference to find the properties you need to keep in the `config` section.