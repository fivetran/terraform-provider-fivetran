---
page_title: "Data Source: fivetran_connector"
---

# Data Source: fivetran_connector

This data source returns a connector object.

## Example Usage

```hcl
data "fivetran_connector" "connector" {
    id = "anonymous_mystery"
}
```

## Schema

### Required

- `id` - The unique identifier for the connector within the Fivetran system.

### Read-Only

- `config` - see [below for nested schema](#nestedatt--config)
- `connected_by` 
- `created_at` 
- `daily_sync_time` 
- `failed_at` 
- `group_id` 
- `name`
- `pause_after_trial` 
- `paused` 
- `schedule_type` 
- `destination_schema` - see [below for nested schema](#nestedatt--schema) 
- `service` 
- `service_version` 
- `status` - see [below for nested schema](#nestedatt--status)
- `succeeded_at` 
- `sync_frequency` 

<a id="nestedatt--schema"></a>
### Nested Schema for `destination_schema`

Read-Only:

- `name`
- `table`
- `prefix`

<a id="nestedatt--config"></a>
### Nested Schema for `config`

Read-Only:

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
- `authorization_method` 
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
- `custom_tables` - see [below for nested schema](#nestedobjatt--config--custom_tables)
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
- `last_synced_changes__utc_` 
- `latest_version` 
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
- `project_credentials` - see [below for nested schema](#nestedobjatt--config--project_credentials)
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
- `reports` - see [below for nested schema](#nestedobjatt--config--reports)
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
- `service_version` 
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
- `token_key` 
- `token_secret`
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

<a id="nestedblock--config--adobe_analytics_configurations"></a>
### Nested Schema for `config.adobe_analytics_configurations`

Read-Only:

- `sync_mode` 
- `report_suites` 
- `elements` 
- `metrics` 
- `calculated_metrics` 
- `segments` 

<a id="nestedobjatt--config--custom_tables"></a>
### Nested Schema for `config.custom_tables`

Read-Only:

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

<a id="nestedobjatt--config--project_credentials"></a>
### Nested Schema for `config.project_credentials`

Read-Only:

- `api_key` 
- `project` 
- `secret_key` 

<a id="nestedobjatt--config--reports"></a>
### Nested Schema for `config.reports`

Read-Only:

- `config_type` 
- `dimensions` 
- `fields` 
- `filter` 
- `metrics` 
- `prebuilt_report` 
- `report_type` 
- `segments` 
- `table` 

<a id="nestedatt--status"></a>
### Nested Schema for `status`

Read-Only:

- `is_historical_sync` 
- `setup_state` 
- `sync_state` 
- `tasks` - see [below for nested schema](#nestedobjatt--status--tasks)
- `update_state` 
- `warnings` - see [below for nested schema](#nestedobjatt--status--warnings)

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