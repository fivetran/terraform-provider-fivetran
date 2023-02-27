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
- `daily_sync_time` - Defines the sync start time when the sync frequency is already set or being set by the current request to 1440. It can be specified in one hour increments starting from 00:00 to 23:00. If not specified, we will use the baseline sync start time. This parameter has no effect on the 0 to 60 minutes offset used to determine the actual sync start time.
- `run_setup_tests` - Specifies whether the setup tests should be run automatically.
- `trust_certificates` - Specifies whether we should trust the certificate automatically. Applicable only for database connectors.
- `trust_fingerprints` - Specifies whether we should trust the SSH fingerprint automatically. Applicable only for database connectors.

-> To complete connector configuration you should specify `run_setup_tests` to `true`. Default value is `false`.

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

- `abs_connection_string` (String)
- `abs_container_name` (String)
- `access_key_id` (String)
- `access_token` (String, Sensitive)
- `account` (String)
- `account_id` (String)
- `account_ids` (List of String)
- `accounts` (List of String)
- `action_breakdowns` (List of String)
- `action_report_time` (String)
- `adobe_analytics_configurations` (Block List) (see [below for nested schema](#nestedblock--config--adobe_analytics_configurations))
- `advertisables` (List of String)
- `advertisers` (List of String)
- `advertisers_id` (List of String)
- `agent_host` (String)
- `agent_ora_home` (String)
- `agent_password` (String, Sensitive)
- `agent_port` (String)
- `agent_public_cert` (String)
- `agent_user` (String)
- `aggregation` (String)
- `always_encrypted` (String)
- `api_access_token` (String, Sensitive)
- `api_key` (String, Sensitive)
- `api_keys` (List of String)
- `api_quota` (String)
- `api_secret` (String, Sensitive)
- `api_token` (String, Sensitive)
- `api_type` (String)
- `api_url` (String)
- `api_version` (String)
- `app_sync_mode` (String)
- `append_file_option` (String)
- `apps` (List of String)
- `archive_pattern` (String)
- `asm_option` (String)
- `asm_oracle_home` (String)
- `asm_password` (String, Sensitive)
- `asm_tns` (String)
- `asm_user` (String)
- `auth_mode` (String)
- `auth_type` (String)
- `aws_region_code` (String)
- `base_url` (String)
- `breakdowns` (List of String)
- `bucket` (String)
- `bucket_name` (String)
- `bucket_service` (String)
- `certificate` (String)
- `click_attribution_window` (String)
- `client_id` (String)
- `client_secret` (String, Sensitive)
- `cloud_storage_type` (String)
- `columns` (List of String)
- `compression` (String)
- `config_method` (String)
- `config_type` (String)
- `connection_string` (String)
- `connection_type` (String)
- `consumer_group` (String)
- `consumer_key` (String, Sensitive)
- `consumer_secret` (String, Sensitive)
- `container_name` (String)
- `conversion_report_time` (String)
- `conversion_window_size` (String)
- `custom_tables` (Block List) (see [below for nested schema](#nestedblock--config--custom_tables))
- `customer_id` (String)
- `daily_api_call_limit` (String)
- `data_center` (String)
- `database` (String)
- `dataset_id` (String)
- `datasource` (String)
- `date_granularity` (String)
- `delimiter` (String)
- `dimension_attributes` (List of String)
- `dimensions` (List of String)
- `domain` (String)
- `domain_name` (String)
- `elements` (List of String)
- `email` (String)
- `enable_all_dimension_combinations` (String)
- `encryption_key` (String, Sensitive)
- `endpoint` (String)
- `engagement_attribution_window` (String)
- `entity_id` (String)
- `escape_char` (String)
- `eu_region` (String)
- `external_id` (String)
- `fields` (List of String)
- `file_type` (String)
- `finance_account_sync_mode` (String)
- `finance_accounts` (List of String)
- `folder_id` (String)
- `ftp_host` (String)
- `ftp_password` (String, Sensitive)
- `ftp_port` (String)
- `ftp_user` (String)
- `function` (String)
- `function_app` (String)
- `function_key` (String)
- `function_name` (String)
- `function_trigger` (String, Sensitive)
- `gcs_bucket` (String)
- `gcs_folder` (String)
- `home_folder` (String)
- `host` (String)
- `hosts` (List of String)
- `identity` (String)
- `instance` (String)
- `integration_key` (String)
- `is_account_level_connector` (String)
- `is_ftps` (String)
- `is_keypair` (String)
- `is_multi_entity_feature_enabled` (String)
- `is_new_package` (String)
- `is_secure` (String)
- `key` (String)
- `manager_accounts` (List of String)
- `merchant_id` (String)
- `message_type` (String)
- `metrics` (List of String)
- `named_range` (String)
- `network_code` (String)
- `null_sequence` (String)
- `oauth_token` (String, Sensitive)
- `oauth_token_secret` (String, Sensitive)
- `on_error` (String)
- `on_premise` (String)
- `organization_id` (String)
- `organizations` (List of String)
- `packed_mode_tables` (List of String)
- `pages` (List of String)
- `password` (String, Sensitive)
- `pat` (String, Sensitive)
- `path` (String)
- `pattern` (String)
- `pdb_name` (String)
- `pem_certificate` (String, Sensitive)
- `port` (String)
- `post_click_attribution_window_size` (String)
- `prebuilt_report` (String)
- `prefix` (String)
- `private_key` (String, Sensitive)
- `profiles` (List of String)
- `project_credentials` (Block List) (see [below for nested schema](#nestedblock--config--project_credentials))
- `project_id` (String)
- `projects` (List of String)
- `public_key` (String)
- `publication_name` (String)
- `query_id` (String)
- `region` (String)
- `replication_slot` (String)
- `report_configuration_ids` (List of String)
- `report_suites` (List of String)
- `report_type` (String)
- `report_url` (String)
- `reports` (Block List) (see [below for nested schema](#nestedblock--config--reports))
- `repositories` (List of String)
- `resource_url` (String)
- `role` (String)
- `role_arn` (String, Sensitive)
- `s3bucket` (String)
- `s3external_id` (String)
- `s3folder` (String)
- `s3role_arn` (String, Sensitive)
- `sales_account_sync_mode` (String)
- `sales_accounts` (List of String)
- `sap_user` (String)
- `secret` (String, Sensitive)
- `secret_key` (String, Sensitive)
- `secrets` (String, Sensitive)
- `secrets_list` (Block List) (see [below for nested schema](#nestedblock--config--secrets_list))
- `security_protocol` (String)
- `selected_exports` (List of String)
- `server_url` (String)
- `servers` (String)
- `sftp_host` (String)
- `sftp_is_key_pair` (String)
- `sftp_password` (String, Sensitive)
- `sftp_port` (String)
- `sftp_user` (String)
- `share_url` (String)
- `sheet_id` (String)
- `shop` (String)
- `sid` (String)
- `site_urls` (List of String)
- `skip_after` (String)
- `skip_before` (String)
- `soap_uri` (String)
- `source` (String)
- `sub_domain` (String)
- `subdomain` (String)
- `swipe_attribution_window` (String)
- `sync_data_locker` (String)
- `sync_format` (String)
- `sync_method` (String)
- `sync_mode` (String)
- `sync_type` (String)
- `technical_account_id` (String)
- `test_table_name` (String)
- `time_zone` (String)
- `timeframe_months` (String)
- `tns` (String)
- `token_key` (String, Sensitive)
- `token_secret` (String, Sensitive)
- `tunnel_host` (String)
- `tunnel_port` (String)
- `tunnel_user` (String)
- `unique_id` (String)
- `update_config_on_each_sync` (String)
- `update_method` (String)
- `use_api_keys` (String)
- `use_oracle_rac` (String)
- `use_webhooks` (String)
- `user` (String)
- `user_id` (String)
- `user_key` (String)
- `user_name` (String)
- `user_profiles` (List of String)
- `username` (String)
- `view_attribution_window` (String)
- `view_through_attribution_window_size` (String)

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

<a id="nestedblock--config--secrets_list"></a>
### Nested Schema for `config.secrets_list`

Required:

- `key` (String)
- `value` (String, Sensitive)

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
terraform import fivetran_connector.my_imported_connector {your Fivetran Connector ID}
```

5.  Use the `terraform state show` command to get the values from the state:

```
terraform state show 'fivetran_connector.my_imported_connector'
```
6. Copy the values and paste them to your `.tf` configuration.

-> The `config` object in the state contains all properties defined in the schema. You need to remove properties from the `config` that are not related to connectors. See the [Fivetran REST API documentation](https://fivetran.com/docs/rest-api/connectors/config) for reference to find the properties you need to keep in the `config` section.

### How to authorize connector

## GitHub connector example

To authorize a GitHub connector via terraform using personal access token you should specify `auth_mode`, `username` and `pat` inside `config` block instead of `auth` and set `run_setup_tests` to `true`:

```hcl
resource "fivetran_connector" "my_github_connector" {
    group_id = "group_id"
    service = "github"
    sync_frequency = 60
    paused = false
    pause_after_trial = false
    run_setup_tests = true

    destination_schema {
        name = "github_connector"
    } 

    config {
        sync_mode = "AllRepositories"
        use_webhooks = false
        auth_mode = "PersonalAccessToken"
        username = "git-hub-user-name"
        pat = "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    }
}
```
