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

- `abs_connection_string` (String)
- `abs_container_name` (String)
- `access_key` (String)
- `domain_host_name` (String)
- `client_name` (String)
- `domain_type` (String)
- `connection_method` (String)
- `access_key_id` (String)
- `access_token` (String)
- `account` (String)
- `account_id` (String)
- `account_ids` (List of String)
- `accounts` (List of String)
- `action_breakdowns` (List of String)
- `action_report_time` (String)
- `adobe_analytics_configurations` (List of Object) (see [below for nested schema](#nestedobjatt--config--adobe_analytics_configurations))
- `advertisables` (List of String)
- `advertisers` (List of String)
- `advertisers_id` (List of String)
- `agent_host` (String)
- `agent_ora_home` (String)
- `agent_password` (String)
- `agent_port` (String)
- `agent_public_cert` (String)
- `agent_user` (String)
- `aggregation` (String)
- `always_encrypted` (String)
- `api_access_token` (String)
- `api_key` (String)
- `api_keys` (List of String)
- `api_quota` (String)
- `api_secret` (String)
- `api_token` (String)
- `api_type` (String)
- `api_url` (String)
- `api_version` (String)
- `app_sync_mode` (String)
- `append_file_option` (String)
- `apps` (List of String)
- `archive_pattern` (String)
- `asm_option` (String)
- `asm_oracle_home` (String)
- `asm_password` (String)
- `asm_tns` (String)
- `asm_user` (String)
- `auth_mode` (String)
- `auth_type` (String)
- `authorization_method` (String)
- `aws_region_code` (String)
- `base_url` (String)
- `breakdowns` (List of String)
- `bucket` (String)
- `bucket_name` (String)
- `bucket_service` (String)
- `certificate` (String)
- `click_attribution_window` (String)
- `client_id` (String)
- `client_secret` (String)
- `cloud_storage_type` (String)
- `columns` (List of String)
- `compression` (String)
- `config_method` (String)
- `config_type` (String)
- `connection_string` (String)
- `connection_type` (String)
- `consumer_group` (String)
- `consumer_key` (String)
- `consumer_secret` (String)
- `container_name` (String)
- `conversion_report_time` (String)
- `conversion_window_size` (String)
- `custom_tables` (List of Object) (see [below for nested schema](#nestedobjatt--config--custom_tables))
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
- `encryption_key` (String)
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
- `ftp_password` (String)
- `ftp_port` (String)
- `ftp_user` (String)
- `function` (String)
- `function_app` (String)
- `function_key` (String)
- `function_name` (String)
- `function_trigger` (String)
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
- `last_synced_changes__utc_` (String)
- `latest_version` (String)
- `manager_accounts` (List of String)
- `merchant_id` (String)
- `message_type` (String)
- `metrics` (List of String)
- `named_range` (String)
- `network_code` (String)
- `null_sequence` (String)
- `oauth_token` (String)
- `oauth_token_secret` (String)
- `on_error` (String)
- `on_premise` (String)
- `organization` (String)
- `organization_id` (String)
- `organizations` (List of String)
- `packed_mode_tables` (List of String)
- `pages` (List of String)
- `password` (String)
- `pat` (String)
- `path` (String)
- `pattern` (String)
- `pdb_name` (String)
- `pem_certificate` (String)
- `port` (String)
- `post_click_attribution_window_size` (String)
- `prebuilt_report` (String)
- `prefix` (String)
- `private_key` (String)
- `profiles` (List of String)
- `project_credentials` (List of Object) (see [below for nested schema](#nestedobjatt--config--project_credentials))
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
- `reports` (List of Object) (see [below for nested schema](#nestedobjatt--config--reports))
- `repositories` (List of String)
- `resource_url` (String)
- `role` (String)
- `role_arn` (String)
- `s3bucket` (String)
- `s3external_id` (String)
- `s3folder` (String)
- `s3role_arn` (String)
- `sales_account_sync_mode` (String)
- `sales_accounts` (List of String)
- `sap_user` (String)
- `secret` (String)
- `secret_key` (String)
- `secrets` (String)
- `secrets_list` (List of Object) (see [below for nested schema](#nestedobjatt--config--secrets_list))
- `security_protocol` (String)
- `selected_exports` (List of String)
- `server_url` (String)
- `servers` (String)
- `service_version` (String)
- `sftp_host` (String)
- `sftp_is_key_pair` (String)
- `sftp_password` (String)
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
- `table` (String)
- `technical_account_id` (String)
- `test_table_name` (String)
- `time_zone` (String)
- `timeframe_months` (String)
- `tns` (String)
- `token_key` (String)
- `token_secret` (String)
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

<a id="nestedobjatt--config--adobe_analytics_configurations"></a>
### Nested Schema for `config.adobe_analytics_configurations`

Read-Only:

- `calculated_metrics` (List of String)
- `elements` (List of String)
- `metrics` (List of String)
- `report_suites` (List of String)
- `segments` (List of String)
- `sync_mode` (String)


<a id="nestedobjatt--config--custom_tables"></a>
### Nested Schema for `config.custom_tables`

Read-Only:

- `action_breakdowns` (List of String)
- `action_report_time` (String)
- `aggregation` (String)
- `breakdowns` (List of String)
- `click_attribution_window` (String)
- `config_type` (String)
- `fields` (List of String)
- `prebuilt_report_name` (String)
- `table_name` (String)
- `view_attribution_window` (String)


<a id="nestedobjatt--config--project_credentials"></a>
### Nested Schema for `config.project_credentials`

Read-Only:

- `api_key` (String)
- `project` (String)
- `secret_key` (String)


<a id="nestedobjatt--config--reports"></a>
### Nested Schema for `config.reports`

Read-Only:

- `config_type` (String)
- `dimensions` (List of String)
- `fields` (List of String)
- `filter` (String)
- `metrics` (List of String)
- `prebuilt_report` (String)
- `report_type` (String)
- `segments` (List of String)
- `table` (String)


<a id="nestedobjatt--config--secrets_list"></a>
### Nested Schema for `config.secrets_list`

Read-Only:

- `key` (String)
- `value` (String)

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