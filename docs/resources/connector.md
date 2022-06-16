---
page_title: "Resource: fivetran_connector"
---

# Resource: fivetran_connector

This resource allows you to create, update, and delete connectors.

## Example Usage

```hcl
resource "fivetran_connector" "amplitude" {
    group_id = fivetran_group.group.id
    service = "amplitude"
    sync_frequency = 60
    paused = false
    pause_after_trial = false

    destination_schema {
        name = "amplitude_connector"
    } 

    config {
        project_credentials {
            project = "project1"
            api_key = "my_api_key"
            secret_key = "my_secret_key"
        }

        project_credentials {
            project = "project2"
            api_key = "my_api_key"
            secret_key = "my_secret_key"
        }
    }
}
```

### NOTE: resources indirect dependencies

The connector resource receives the `group_id` parameter value from the group resource, but the destination resource depends on the group resource.  When you try to destroy the destination resource infrastructure, the terraform plan is created successfully, but once you run the `terraform apply` command, it returns an error because the Fivetran API doesn't let you delete destinations that have linked connectors. To solve this problem, you should either explicitly define `depends_on` between the connector and destination:

```hcl
resource "fivetran_connector" "amplitude" {
    ...
    depends_on = [
        fivetran_destination.my_destination
    ]
}
```

or get the group ID from the destination:

```hcl
resource "fivetran_connector" "amplitude" {
    group_id = fivetran_destination.my_destination.group_id
    ...
}
```

## Schema

### Required

- `config` - The connector setup configuration. The format is specific for each connector. (see [below for nested schema](#nestedblock--config))
- `group_id` - The unique identifier for the group within the Fivetran system.
- `pause_after_trial` - Specifies whether the connector should be paused after the free trial period has ended.
- `paused` - Specifies whether the connector is paused.
- `destination_schema` - The connector destination schema configuration. Defines connector schema identity in destination. (see [below for nested schema](#nestedblock--schema)) 
- `service` - The name for the connector type within the Fivetran system.
- `sync_frequency` - The connector sync frequency in minutes. The supported values are: `5`, `15`, `30`, `60`, `120`, `180`, `360`, `480`, `720`, `1440`.

### Optional

- `auth` - The connector authorization settings. Can be used to authorize a connector using your external client credentials. The format is specific for each connector. (see [below for nested schema](#nestedblock--auth))
- `run_setup_tests` - Specifies whether the setup tests should be run automatically.
- `trust_certificates` - Specifies whether we should trust the certificate automatically. Applicable only for database connectors.
- `trust_fingerprints` - Specifies whether we should trust the SSH fingerprint automatically. Applicable only for database connectors.

### Read-Only

- `connected_by` 
- `created_at` 
- `failed_at` 
- `id` 
- `last_updated` 
- `name`
- `schedule_type` 
- `service_version` 
- `status` - (see [below for nested schema](#nestedatt--status))
- `succeeded_at` 

<a id="nestedblock--config"></a>
### Nested Schema for `config`

See [Connector Config](https://fivetran.com/docs/rest-api/connectors/config) for details.

Optional:

- `abs_connection_string` 
- `abs_container_name` 
- `access_key_id` 
- `access_token` 
- `account` 
- `account_id` 
- `account_ids` 
- `accounts` 
- `action_breakdowns` 
- `action_report_time` 
- `adobe_analytics_configurations` - see [below for nested schema](#nestedblock--config--adobe_analytics_configurations)
- `advertisables` 
- `advertisers` 
- `advertisers_id` 
- `aggregation` 
- `always_encrypted` 
- `api_access_token` 
- `api_key` 
- `api_keys` 
- `api_quota` 
- `api_secret` 
- `api_token` 
- `api_type`
- `api_url` 
- `api_version` 
- `app_sync_mode` 
- `append_file_option` 
- `apps` 
- `archive_pattern` 
- `auth_mode` 
- `auth_type` 
- `aws_region_code`
- `base_url`
- `breakdowns` 
- `bucket` 
- `bucket_name` 
- `bucket_service` 
- `certificate` 
- `click_attribution_window` 
- `client_id` 
- `client_secret` 
- `cloud_storage_type` 
- `columns` 
- `compression` 
- `config_method` 
- `config_type` 
- `connection_string` 
- `connection_type`
- `consumer_group` 
- `consumer_key` 
- `consumer_secret` 
- `container_name` 
- `conversion_report_time` 
- `conversion_window_size` 
- `custom_tables` - see [below for nested schema](#nestedblock--config--custom_tables)
- `customer_id` 
- `daily_api_call_limit` 
- `data_center` 
- `database` 
- `dataset_id` 
- `datasource` 
- `date_granularity` 
- `delimiter` 
- `dimension_attributes` 
- `dimensions` 
- `domain` 
- `domain_name` 
- `elements` 
- `email` 
- `enable_all_dimension_combinations`
- `encryption_key`
- `endpoint` 
- `engagement_attribution_window`
- `entity_id`  
- `escape_char` 
- `eu_region` 
- `external_id` 
- `fields` 
- `file_type` 
- `finance_account_sync_mode` 
- `finance_accounts` 
- `folder_id`
- `ftp_host` 
- `ftp_password` 
- `ftp_port` 
- `ftp_user` 
- `function` 
- `function_app` 
- `function_key` 
- `function_name` 
- `function_trigger` 
- `gcs_bucket` 
- `gcs_folder` 
- `home_folder` 
- `host` 
- `hosts` 
- `identity` 
- `instance` 
- `integration_key` 
- `is_ftps` 
- `is_multi_entity_feature_enabled`
- `is_new_package`
- `is_secure` 
- `key` 
- `manager_accounts` 
- `merchant_id` 
- `message_type` 
- `metrics` 
- `named_range` 
- `network_code` 
- `null_sequence` 
- `oauth_token` 
- `oauth_token_secret` 
- `on_error` 
- `on_premise` 
- `organization_id` 
- `organizations` 
- `pages` 
- `password` 
- `path` 
- `pattern` 
- `pem_certificate` 
- `port` 
- `post_click_attribution_window_size` 
- `prebuilt_report` 
- `prefix` 
- `private_key` 
- `profiles` 
- `project_credentials` - see [below for nested schema](#nestedblock--config--project_credentials)
- `project_id` 
- `projects` 
- `public_key` 
- `publication_name` 
- `query_id` 
- `region` 
- `replication_slot` 
- `report_configuration_ids` 
- `report_suites` 
- `report_type` 
- `report_url` 
- `reports` - see [below for nested schema](#nestedblock--config--reports)
- `repositories` 
- `resource_url` 
- `role` 
- `role_arn` 
- `s3bucket` 
- `s3external_id` 
- `s3folder` 
- `s3role_arn` 
- `sales_account_sync_mode` 
- `sales_accounts` 
- `secret` 
- `secret_key` 
- `secrets` 
- `security_protocol` 
- `selected_exports` 
- `server_url` 
- `servers` 
- `sftp_host` 
- `sftp_is_key_pair` 
- `sftp_password` 
- `sftp_port` 
- `sftp_user` 
- `sheet_id` 
- `shop` 
- `sid` 
- `site_urls` 
- `skip_after` 
- `skip_before` 
- `soap_uri` 
- `source` 
- `sub_domain` 
- `subdomain` 
- `swipe_attribution_window` 
- `sync_data_locker` 
- `sync_format` 
- `sync_mode` 
- `sync_type` 
- `technical_account_id` 
- `test_table_name` 
- `time_zone` 
- `timeframe_months` 
- `tunnel_host` 
- `tunnel_port` 
- `tunnel_user` 
- `unique_id` 
- `update_config_on_each_sync` 
- `update_method` 
- `use_api_keys` 
- `use_webhooks` 
- `user` 
- `user_id` 
- `user_key` 
- `user_name` 
- `user_profiles` 
- `username` 
- `view_attribution_window` 
- `view_through_attribution_window_size` 

Read-Only:

- `authorization_method` 
- `last_synced_changes__utc_` 
- `latest_version` 
- `service_version` 

<a id="nestedblock--schema"></a>
### Nested Schema for `destination-schema`

Optional:

- `name` - required for all connectors instead of db-like connectors, represents `config.schema` field.
- `table` - required for some non db-like connectors, represents `config.table` field.
- `prefix` - required only for db-like connectors, represents `config.schema_prefix` field.

See [Connector Config](https://fivetran.com/docs/rest-api/connectors/config) for details.

<a id="nestedblock--config--adobe_analytics_configurations"></a>
### Nested Schema for `config.adobe_analytics_configurations`

Optional:

- `sync_mode` 
- `report_suites` 
- `elements` 
- `metrics` 
- `calculated_metrics` 
- `segments` 

<a id="nestedblock--config--custom_tables"></a>
### Nested Schema for `config.custom_tables`

Optional:

- `action_breakdowns` 
- `action_report_time` 
- `aggregation` 
- `breakdowns` 
- `click_attribution_window` 
- `config_type` 
- `fields` 
- `prebuilt_report_name` 
- `table_name` 
- `view_attribution_window` 

<a id="nestedblock--config--project_credentials"></a>
### Nested Schema for `config.project_credentials`

Optional:

- `api_key` 
- `project` 
- `secret_key` 

<a id="nestedblock--config--reports"></a>
### Nested Schema for `config.reports`

Optional:

- `config_type` 
- `dimensions` 
- `fields` 
- `filter` 
- `metrics` 
- `prebuilt_report` 
- `report_type` 
- `segments` 
- `table` 

<a id="nestedblock--auth"></a>
### Nested Schema for `auth`

See [Connector Config](https://fivetran.com/docs/rest-api/connectors/config) for details.

Optional:

- `access_token` 
- `client_access` see [below for nested schema](#nestedblock--auth--client_access)
- `realm_id` 
- `refresh_token` 

<a id="nestedblock--auth--client_access"></a>
### Nested Schema for `auth.client_access`

Optional:

- `client_id` 
- `client_secret` 
- `developer_token` 
- `user_agent` 

<a id="nestedatt--status"></a>
### Nested Schema for `status`

Read-Only:

- `is_historical_sync` 
- `setup_state` 
- `sync_state` 
- `tasks` see [below for nested schema](#nestedobjatt--status--tasks)
- `update_state` 
- `warnings` see [below for nested schema](#nestedobjatt--status--warnings)

<a id="nestedobjatt--status--tasks"></a>
### Nested Schema for `status.tasks`

Read-Only:

- `code` 
- `message` 

<a id="nestedobjatt--status--warnings"></a>
### Nested Schema for `status.warnings`

Read-Only:

- `code` 
- `message` 

## Import

1. To import an existing `fivetran_connector` resource into your Terraform state, you need to get **Fivetran Connector ID** on the **Setup** tab of the connector page in your Fivetran dashboard.

2. Retrieve all connectors in a particular group using the [fivetran_group_connectors data source](/docs/data-sources/group_connectors). To retrieve existing groups, use the [fivetran_groups data source](/docs/data-sources/groups).

3. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_connector" "my_imported_connector" {

}
```

4. Run the `terraform import` command:

```
terraform import fivetran_connector.my_imported_connector <your Fivetran Connector ID>
```

5.  Use the `terraform state show` command to get the values from the state:

```
terraform state show 'fivetran_connector.my_imported_connector'
```
6. Copy the values and paste them to your `.tf` configuration.

-> The `config` object in the state contains all properties defined in the schema. You need to remove properties from the `config` that are not related to connectors. See the [Fivetran REST API documentation](https://fivetran.com/docs/rest-api/connectors/config) for reference to find the properties you need to keep in the `config` section.
