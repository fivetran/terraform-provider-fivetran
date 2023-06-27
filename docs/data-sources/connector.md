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

- `id` (String) The unique identifier for the user within the account.

### Optional

- `config` (Block List, Max: 1) (see [below for nested schema](#nestedblock--config))

### Read-Only

- `api_type` (String)
- `connected_by` (String) The unique identifier of the user who has created the connector in your account
- `created_at` (String) The timestamp of the time the connector was created in your account
- `daily_api_call_limit` (String)
- `daily_sync_time` (String) The optional parameter that defines the sync start time when the sync frequency is already set or being set by the current request to 1440. It can be specified in one hour increments starting from 00:00 to 23:00. If not specified, we will use [the baseline sync start time](https://fivetran.com/docs/getting-started/syncoverview#syncfrequencyandscheduling). This parameter has no effect on the [0 to 60 minutes offset](https://fivetran.com/docs/getting-started/syncoverview#syncstarttimesandoffsets) used to determine the actual sync start time
- `destination_schema` (List of Object) (see [below for nested schema](#nestedatt--destination_schema))
- `elements` (Set of String) The elements that you want to sync.
- `environment` (String)
- `failed_at` (String) The timestamp of the time the connector sync failed last time
- `group_id` (String) The unique identifier for the Group within the Fivetran system.
- `name` (String) The unique identifier for the team within the account
- `oauth_token` (String, Sensitive) The Twitter App access token.
- `oauth_token_secret` (String, Sensitive) The Twitter App access token secret.
- `organization` (String)
- `pause_after_trial` (String) Specifies whether the connector should be paused after the free trial period has ende
- `paused` (String) Specifies whether the connector is paused
- `report_suites` (Set of String) Specific report suites to sync. Must be populated if `sync_mode` is set to `SpecificReportSuites`.
- `schedule_type` (String) The connector schedule configuration type. Supported values: auto, manual
- `service` (String) The connector type name within the Fivetran system
- `service_version` (String) The connector type version within the Fivetran system.
- `status` (List of Object) (see [below for nested schema](#nestedatt--status))
- `succeeded_at` (String) The timestamp of the time the connector sync succeeded last time
- `sync_frequency` (String) The connector sync frequency in minutes
- `test_table_name` (String)
- `unique_id` (String)

<a id="nestedblock--config"></a>
### Nested Schema for `config`

Optional:

- `abs_connection_string` (String) Azure blob storage connection string.
- `abs_container_name` (String) Azure blob storage container name.
- `abs_prefix` (String) Prefix
- `access_key` (String, Sensitive) The access key for API authentication.
- `access_key_id` (String, Sensitive) Your AWS access key ID.
- `access_token` (String, Sensitive) The Stripe API Restricted Key
- `access_type` (String) Access Type
- `account` (String) The NetSuite Account ID.
- `account_id` (String) Your Optimizely account ID.
- `account_ids` (Set of String)
- `account_key` (String, Sensitive) The read-only primary or secondary account key for the database account. Required for the `ACCOUNT_KEY` data access method.
- `account_sync_mode` (String) Account Sync Mode
- `accounts` (Set of String)
- `action_breakdowns` (Set of String)
- `action_report_time` (String) The report time of action stats. [Possible action_report time values](/docs/applications/facebook-ad-insights/api-config#actionreporttime).
- `ad_analytics` (String) Whether to sync all analytic reports or specific. Default value: `AllReports`
- `ad_unit_view` (String) Ad unit view for the report.
- `adobe_analytics_configurations` (Block Set) (see [below for nested schema](#nestedblock--config--adobe_analytics_configurations))
- `adroll_config_v1_dimensions` (Set of String)
- `adroll_config_v1_metrics` (Set of String)
- `advertisables` (Set of String)
- `advertisers` (Set of String)
- `advertisers_id` (Set of String)
- `agent_host` (String) The agent host.
- `agent_ora_home` (String) The home directory of the Oracle database.
- `agent_password` (String, Sensitive) The agent password.
- `agent_port` (Number) The agent port.
- `agent_public_cert` (String) The public certificate for the agent.
- `agent_user` (String) The agent user name.
- `aggregation` (String) Options to select aggregation duration. [Possible aggregation values](/docs/applications/facebook-ad-insights/api-config#aggregation).
- `always_encrypted` (Boolean) Require TLS through Tunnel
- `amazon_ads_config_v1_profiles` (Set of String)
- `apache_kafka_config_v1_schema_registry_urls` (Set of String)
- `apache_kafka_config_v1_servers` (Set of String)
- `api_access_token` (String, Sensitive) API access token of your custom app.
- `api_key` (String, Sensitive) Your Freshservice API Key.
- `api_keys` (Set of String)
- `api_quota` (Number) Allowed number of API requests to Marketo instance per day, the default value is 10000.
- `api_requests_per_minute` (Number) Allowed number of API requests to Qualtrics per minute, the default value is 2000. Maximum allowed number is 3000 because brands may make up to 3000 API requests per minute across all of its API calls.
- `api_secret` (String, Sensitive) The Sailthru API secret.
- `api_token` (String, Sensitive) The Recharge API token.
- `api_url` (String) Your Braze API URL.
- `api_usage` (String) Maximum Zendesk Api Usage allowed
- `api_version` (String) API Version
- `app_ids` (Set of String)
- `app_name` (String) OAuth App Name
- `app_specific_password` (String, Sensitive) Your app-specific password
- `app_sync_mode` (String) Whether to sync all apps or specific apps.
- `append_file_option` (String) If you know that the source completely over-writes the same file with new data, you can append the changes instead of upserting based on filename and line number.
- `apps` (Set of String)
- `archive_pattern` (String) Files inside of compressed archives with filenames matching this regular expression will be synced.
- `asb_ip` (String) The IP address (or) the URL of ASB namespace
- `asm_option` (Boolean) Default value: `false`. Set to `true` if you're using ASM on a non-RAC instance.
- `asm_oracle_home` (String) ASM Oracle Home path.
- `asm_password` (String, Sensitive) ASM password. Mandatory if `use_oracle_rac` or `asm_option` is set to `true`.
- `asm_tns` (String) ASM TNS.
- `asm_user` (String) ASM user. Mandatory if `use_oracle_rac` or `asm_option` is set to `true`.
- `attribution_window` (String) Time period used to attribute conversions based on clicks.
- `auth` (String) Password-based or key-based authentication type
- `auth_method` (String) Authentication Method.
- `auth_mode` (String) Authorization type.
- `auth_type` (String) Authorization type. Required for storage bucket authentication.
- `aws_msk_config_v1_schema_registry_urls` (Set of String)
- `aws_msk_config_v1_servers` (Set of String)
- `aws_region_code` (String) The AWS region code for the DynamoDB instance, e.g. `us-east-1`.
- `base_id` (String) ID of base in Airtable
- `base_url` (String) (Optional) The custom Salesforce domain. Make sure that the `base_url` starts with `https://`.
- `bingads_config_v1_accounts` (Set of String)
- `blob_sas_url` (String, Sensitive) The blob SAS URL of your Azure container. Required if `bucket_service` is set to `AZURE`.
- `breakdowns` (Set of String)
- `bucket` (String) The Google Cloud Storage source bucket.
- `bucket_name` (String) The name of the bucket.
- `bucket_service` (String) Whether to store the events in Fivetran's container service or your S3 bucket. Default value: `Fivetran`.
- `business_unit_id` (String) Business Unit Id
- `certificate` (String, Sensitive) The contents of your PEM certificate file. Must be populated if `auth_mode` is set to `Certificate`.
- `click_attribution_window` (String) Time period to attribute conversions based on clicks. [Possible click_attribution_window values](/docs/applications/facebook-ad-insights/api-config#clickattributionwindow).
- `client_access` (String) Your application client access fields.
- `client_cert` (String, Sensitive) Kafka client certificate.
- `client_cert_key` (String, Sensitive) Kafka client certificate key.
- `client_id` (String) Marketo REST API Client Id.
- `client_name` (String, Sensitive) Medallia company name
- `client_secret` (String, Sensitive) Marketo REST API Client Secret.
- `cloud_storage_type` (String) Cloud storage type Braze Current is connected to.
- `columns` (Set of String)
- `company_id` (String) Company ID
- `compression` (String) The compression format is used to let Fivetran know that even files without a compression extension should be decompressed using the selected compression format.
- `config_method` (String) The report configuration method. Specifies whether a new configuration is defined manually or an existing configuration is reused. The default value is `CREATE_NEW`.
- `config_repository_url` (String) Public repository URL containing JSON configuration files.
- `config_type` (String) Option to select Prebuilt Reports or Custom Reports. [Possible config_type values](/docs/applications/facebook-ad-insights/api-config#configtype).
- `confluent_cloud_config_v1_schema_registry_urls` (Set of String)
- `connection_method` (String) The connection method used to connect to SFTP Server.
- `connection_string` (String) Connection string of the Event Hub Namespace you want to sync.
- `connection_type` (String) Possible values:`SshTunnel`. `SshTunnel` is used as a value if this parameter is omitted in the request and the following parameter's values are specified: `tunnel_host`, `tunnel_port`, `tunnel_user`.
- `consumer_group` (String) Name of consumer group created for Fivetran.
- `consumer_key` (String) The Twitter App consumer key.
- `consumer_secret` (String, Sensitive) The Twitter App consumer secret.
- `container_address` (String) IP address of Azure Storage Container which is accessible from host machine.
- `container_name` (String) The name of the blob container.
- `content_owner_id` (String) Used only for Content Owner reports. The ID of the content owner for whom the API request is being made.
- `conversion_dimensions` (Set of String)
- `conversion_report_time` (String) The date that the user interacted with the ad OR completed a conversion event.
- `conversion_window_size` (Number) A period of time in days during which a conversion is recorded.
- `criteo_config_v1_metrics` (Set of String)
- `csv_definition` (String) CSV definition for the CSV export (https://help.adjust.com/en/article/csv-uploads#how-do-i-format-my-csv-definition).
- `currency` (String) Currency
- `custom_floodlight_variables` (Set of String)
- `custom_reports` (Block Set) (see [below for nested schema](#nestedblock--config--custom_reports))
- `custom_tables` (Block Set) (see [below for nested schema](#nestedblock--config--custom_tables))
- `customer_id` (String) ID of the customer, can be retrieved from your AdWords dashboard.
- `customer_list_id` (String) The parameter to retrieve customer details.
- `data_access_method` (String) The source data access method. Supported values:<br>`ACCOUNT_KEY`- Data access method that uses account keys to authenticate to the source database. It comes in both read-write and read-only variants.<br>`RESOURCE_TOKEN`- Fine-grained permission model based on native Azure Cosmos DB users and permissions.<br> Learn more in our [Cosmos DB Data Access Methods documentation](/docs/databases/cosmos#dataaccessmethods).
- `data_center` (String) Data Center
- `database` (String) The database name.
- `dataset_id` (String) The dataset ID.
- `datasource` (String) The NetSuite data source value: `NetSuite.com`.
- `date_granularity` (String) The aggregation duration you want. Default value: `HOUR` .
- `delimiter` (String) You can specify the delimiter that your CSVs use here. Fivetran generally tries to infer the delimiter, but in some cases this is impossible.
- `dimension_attributes` (Set of String)
- `dimension_filters` (Block Set) (see [below for nested schema](#nestedblock--config--dimension_filters))
- `dimensions` (Set of String)
- `domain` (String) Zendesk domain.
- `domain_host_name` (String) Workday host name.
- `domain_name` (String) The custom domain name associated with Dynamics 365.
- `domain_type` (String) Domain type of your Medallia URL
- `double_click_campaign_manager_config_v1_dimensions` (Set of String)
- `double_click_campaign_manager_config_v1_metrics` (Set of String)
- `dynamodb_config_v1_packed_mode_tables` (Set of String)
- `email` (String) The email of the Pardot user.
- `empty_header` (Boolean) <strong>Optional.</strong> If your CSV generating software doesn't provide header line for the documents, Fivetran can generate the generic column names and sync data rows with them.
- `enable_all_dimension_combinations` (Boolean) Whether to enable all reach dimension combinations in the report. Default value: `false`
- `enable_enrichments` (Boolean) Enable Enrichments
- `enable_exports` (Boolean) Enable User Profile Exports
- `enable_tde` (Boolean) Using Transparent Data Encryption (TDE)
- `encoded_private_key` (String, Sensitive) The encoded contents of your PEM encoded private key file.
- `encryption_key` (String, Sensitive) Marketo SOAP API Encryption Key.
- `endpoint` (String) Connection-specific collector endpoint. The collector endpoint will have the `webhooks.fivetran.com/snowplow/<endpoint_ID>` format. You will need it to configure Snowplow to connect with Fivetran.
- `engagement_attribution_window` (String) The number of days to use as the conversion attribution window for an engagement (i.e. closeup or save) action.
- `enriched_export` (String) Enriched Events S3 bucket
- `entity_id` (String) If `is_multi_entity_feature_enabled` is `true`, then it's `EntityId`.
- `escape_char` (String) If your CSV generator follows non-standard rules for escaping quotation marks, you can set the escape character here.
- `eu_region` (Boolean) Turn it on if your app is on EU region
- `export_storage_type` (String) Export Storage
- `external_id` (String) This is the same as your `group_id`, used for authentication along with the `role_arn`.
- `facebook_ad_account_config_v1_accounts` (Set of String)
- `facebook_ads_config_v1_accounts` (Set of String)
- `facebook_config_v1_accounts` (Set of String)
- `fields` (Set of String)
- `file_type` (String) If your files are saved with improper extensions, you can force them to be synced as the selected filetype.
- `filter` (String) String parameter restricts the data returned for your report. To use the filters parameter, specify a dimension or metric on which to filter, followed by the filter expression
- `finance_account_sync_mode` (String) Whether to sync all finance accounts or specific finance accounts.
- `finance_accounts` (Set of String)
- `folder` (String) Your Dropbox Folder URL.
- `folder_id` (String) Folder URL
- `folder_path` (String) Your OneDrive folder URL
- `ftp_host` (String) FTP host.
- `ftp_password` (String, Sensitive) FTP password.
- `ftp_port` (Number) FTP port.
- `ftp_user` (String) FTP user.
- `function` (String) The name of your AWS Lambda Function.
- `function_app` (String) Function app name in Azure portal.
- `function_key` (String, Sensitive) Function key used for authorization.
- `function_name` (String) Name of the function to be triggered.
- `function_trigger` (String, Sensitive) The trigger URL of the cloud function.
- `gcs_bucket` (String) The GCS bucket name. Required if `bucket_service` is set to `GCS`.
- `gcs_folder` (String) Your GCS folder name. Required if `GCS` is the `cloud_storage_type`
- `google_ads_config_v1_accounts` (Set of String)
- `google_ads_config_v1_reports` (Set of String)
- `google_analytics_4_config_v1_accounts` (Set of String)
- `google_analytics_4_config_v1_reports` (Set of String)
- `google_analytics_config_v1_accounts` (Set of String)
- `google_analytics_config_v1_dimensions` (Set of String)
- `google_analytics_config_v1_profiles` (Set of String)
- `google_analytics_config_v1_reports` (Set of String)
- `google_analytics_mcf_config_v1_accounts` (Set of String)
- `google_display_and_video_360_config_v1_dimensions` (Set of String)
- `google_display_and_video_360_config_v1_metrics` (Set of String)
- `google_display_and_video_360_config_v1_partners` (Set of String)
- `google_search_console_config_v1_reports` (Set of String)
- `group_name` (String) (Optional) The group name of the `target_group_id`.
- `has_manage_permissions` (Boolean) The boolean value specifying whether the connection string has manage permissions
- `heroku_kafka_config_v1_servers` (Set of String)
- `home_folder` (String) Your S3 home folder path of the Data Locker.
- `host` (String) A host address of the primary node. It should be a DB instance host/IP address with a port number.
- `host_ip` (String) IP address of host tunnel machine which is used to connect to Storage Container.
- `host_user` (String) Username in the host machine.
- `hosts` (Set of String)
- `identity` (String) Marketo REST API identity url.
- `instagram_business_config_v1_accounts` (Set of String)
- `instance` (String) ServiceNow Instance ID.
- `instance_number` (String) Two-digit number (00-97) of the SAP instance within its host.
- `instance_url` (String) The SAP Business ByDesign instance URL.
- `integration_key` (String, Sensitive) The integration key of the Pendo account.
- `is_account_level_connector` (Boolean) (Optional) Retrieve account-level logs.
- `is_auth2_enabled` (Boolean) The contents of your PEM certificate file. Default value: `false`
- `is_custom_api_credentials` (Boolean) Custom API credentials
- `is_ftps` (Boolean) Use Secure FTP (FTPS).
- `is_keypair` (Boolean) Whether to use a key pair for authentication.  When `true`, do not use `password`.
- `is_multi_entity_feature_enabled` (Boolean) Set to `true` if there are multiple entities in your Zuora account and you only want to use one entity. Otherwise, set to `false`.
- `is_new_package` (Boolean) Indicates that that your installed package uses OAuth 2.0. Default value: `false`
- `is_private_key_encrypted` (Boolean) Indicates that a private key is encrypted. The default value: `false`. The field can be specified if authentication type is `KEY_PAIR`.
- `is_private_link_required` (Boolean) We use PrivateLink by default if your AWS Lambda is in the same region as Fivetran. Turning on this toggle ensures that Fivetran always connects to AWS lambda over PrivateLink. Learn more in our [PrivateLink documentation](https://fivetran.com/docs/databases/connection-options#awsprivatelinkbeta).
- `is_public` (Boolean) Is the bucket public? (you don't need an AWS account for syncing public buckets!)
- `is_sailthru_connect_enabled` (Boolean) Enable this if you want to sync Sailthru Connect
- `is_secure` (Boolean) Whether the server supports FTPS.
- `is_single_table_mode` (Boolean) Allows the creation of connector using Merge Mode strategy.
- `json_delivery_mode` (String) Control how your JSON data is delivered into your destination
- `key` (String) The UserVoice API key.
- `key_store_type` (String) Key Store Type
- `linkedin_ads_config_v1_accounts` (Set of String)
- `list_strategy` (String) <strong>Optional.</strong> If you have a file structure where new files are always named in lexicographically increasing order such as files being named in increasing order of time, you can select <code>time_based_pattern_listing</code>.
- `log_journal` (String) The log journal name.
- `log_journal_schema` (String) The log journal schema.
- `log_truncater` (String) Log Truncater.
- `login` (String) The Trade Desk email. It is a part of the login credentials.
- `login_password` (String, Sensitive) The login password. It is a part of the login credentials.
- `manager_accounts` (Set of String)
- `merchant_id` (String) Your Braintree merchant ID.
- `message_type` (String) Message type.
- `metrics` (Set of String)
- `mongo_sharded_config_v1_hosts` (Set of String)
- `mongo_sharded_config_v1_packed_mode_tables` (Set of String)
- `named_range` (String) The name of the named data range on the sheet that contains the data to be synced.
- `namespace` (String) The ASB namespace which we have to sync. Required for `AzureActiveDirectory` authentication.
- `network_code` (Number) Network code is a unique, numeric identifier for your Ad Manager network.
- `null_sequence` (String) If your CSVs use a special value indicating null, you can specify it here.
- `oauth_token` (String, Sensitive) The Twitter App access token.
- `oauth_token_secret` (String, Sensitive) The Twitter App access token secret.
- `on_error` (String) If you know that your files contain some errors, you can choose to have poorly formatted lines skipped. We recommend leaving the value as fail unless you are certain that you have undesirable, malformed data.
- `on_premise` (Boolean) Whether the Jira instance is local or in cloud.
- `organization_id` (String) Organization ID from the Service Account (JWT) credentials of your Adobe Project.
- `organizations` (Set of String)
- `packed_mode_tables` (Set of String)
- `packing_mode` (String) Whether to sync all tables in unpacked mode only, all tables in packed mode only, or specific tables in packed mode. Default value: `UseUnpackedModeOnly`.
- `pages` (Set of String)
- `partners` (Set of String)
- `passphrase` (String, Sensitive) In case private key is encrypted, you are required to enter passphrase that was used to encrypt the private key. The field can be specified if authentication type is `KEY_PAIR`.
- `password` (String, Sensitive) The user's password.
- `pat` (String, Sensitive) The `Personal Access Token` generated in Github.
- `path` (String) A URL subdirectory where the Jira instance is working.
- `pattern` (String) All files in your search path matching this regular expression will be synced.
- `pdb_name` (String) (Multi-tenant databases only) The database's PDB name. Exclude this parameter for single-tenant databases.
- `pem_certificate` (String, Sensitive) The contents of your PEM certificate file. Must be populated if `is_auth2_enabled` is set to `false`.
- `pem_private_key` (String, Sensitive) The contents of your PEM secret key file. Must be populated if `is_auth2_enabled` is set to `true`.
- `per_interaction_dimensions` (Set of String)
- `pgp_pass_phrase` (String, Sensitive) The PGP passphrase used to create the key. Must be populated if `use_pgp_encryption_options` is set to `true`.
- `pgp_secret_key` (String, Sensitive) The contents of your PGP secret key file. Must be populated if `use_pgp_encryption_options` is set to `true`.
- `phone_number` (String) Register the number on AppleId Account Page for 2FA
- `pinterest_ads_config_v1_advertisers` (Set of String)
- `port` (Number) The port number.
- `post_click_attribution_window_size` (String) The time period to attribute conversions based on clicks. Default value: `DAY_30`
- `prebuilt_report` (String) The name of report of which connector will sync the data. [Possible prebuilt_report values](/docs/applications/facebook-ad-insights/api-config#prebuiltreport).
- `prefix` (String) All files and folders under this folder path will be searched for files to sync.
- `primary_keys` (Set of String)
- `private_key` (String, Sensitive) Private access key.  The field should be specified if authentication type is `KEY_PAIR`.
- `profiles` (Set of String)
- `project_credentials` (Block Set) (see [below for nested schema](#nestedblock--config--project_credentials))
- `project_id` (String) The project ID.
- `projects` (Set of String)
- `properties` (Set of String)
- `public_key` (String) Public Key
- `publication_name` (String) Publication name. Specify only for `"updated_method": "WAL_PGOUTPUT"`.
- `query_id` (String) The ID of the query whose configuration you want to reuse. This is a required parameter when `config_method` is set to `REUSE_EXISTING`.
- `realm_id` (String) `Realm ID` of your QuickBooks application.
- `refresh_token` (String, Sensitive) The long-lived `Refresh token` along with the `client_id` and `client_secret` parameters carry the information necessary to get a new access token for API resources.
- `refresh_token_expires_at` (String) The expiration date of the refresh token. Unix timestamp in seconds
- `region` (String) The AWS region code for the DynamoDB instance.
- `replication_slot` (String) Replication slot name. Specify only for `"updated_method": "WAL"` or `"WAL_PGOUTPUT"`.
- `report_configuration_ids` (Set of String)
- `report_timezone` (String) Report Timezone
- `report_type` (String) Type of reporting data to sync. Default value: `STANDARD`.
- `report_url` (String) URL for a live custom report.
- `reports` (Set of String)
- `repositories` (Set of String)
- `resource_token` (String, Sensitive) A token that provides access to a specific Cosmos DB resource. Required for the `RESOURCE_TOKEN` data access method.
- `resource_url` (String) URL at which Dynamics 365 is accessed
- `rest_api_limit` (Number) The number of API calls that the connector should not exceed in a day. Default REST API call limit per day: 150,000.
- `rfc_library_path` (String) Directory path containing the SAP NetWeaver RFC SDK library files.
- `role` (String) Snowflake Connector role name
- `role_arn` (String, Sensitive) Role ARN
- `rollback_window_size` (Number) A period of time in days during which a conversion is recorded.
- `s3_bucket` (String) The S3 bucket name. Required if `bucket_service` is set to `S3`.
- `s3_export_bucket` (String) Exports Bucket
- `s3_export_folder` (String) Exports Folder
- `s3_export_role_arn` (String, Sensitive) Exports Role ARN
- `s3_role_arn` (String, Sensitive) The Role ARN required for authentication. Required if `bucket_service` is set to `S3`.
- `s3bucket` (String) The S3 bucket name.
- `s3external_id` (String) This is the same as your `group_id`, used for authentication along with the `role_arn` required if `AWS_S3` is the `cloud_storage_type`
- `s3folder` (String) Your S3 folder name required if `AWS_S3` is the `cloud_storage_type`
- `s3path` (String) Copy and use this to configure Sailthru Connect in your sailthru account.
- `s3role_arn` (String, Sensitive) The Role ARN required for authentication.
- `sales_account_sync_mode` (String) Whether to sync all sales accounts or specific sales accounts.
- `sales_accounts` (Set of String)
- `salesforce_security_token` (String, Sensitive) The Pardot user's Salesforce SSO Account Security Token.
- `sap_schema` (String) The SAP schema.
- `sap_user` (String) The Oracle schema name where the SAP tables reside.
- `sasl_mechanism` (String) SASL Mechanism
- `sasl_plain_key` (String, Sensitive) API Key
- `sasl_plain_secret` (String, Sensitive) API Secret
- `sasl_scram256_key` (String, Sensitive) API Key
- `sasl_scram256_secret` (String, Sensitive) API Secret
- `sasl_scram512_key` (String, Sensitive) If `security_protocol` is set to `SASL`, enter your secret's `saslScram512Key`.
- `sasl_scram512_secret` (String, Sensitive) If `security_protocol` is set to `SASL`, enter your secret's `saslScram512Key`.
- `schema` (String) Destination schema. Schema is permanent and cannot be changed after connection creation
- `schema_prefix` (String) Destination schema prefix. Prefix for each replicated schema. For example with prefix 'x', source schemas 'foo' and 'bar' get replicated as 'x_foo' and 'x_bar'. The prefix is permanent and cannot be changed after connection creation
- `schema_registry_credentials_source` (String) Schema Registry Credentials source
- `schema_registry_key` (String, Sensitive) Schema Registry Key
- `schema_registry_secret` (String, Sensitive) Schema Registry Secret
- `schema_registry_urls` (Set of String)
- `secret` (String, Sensitive) The UserVoice API secret.
- `secret_key` (String, Sensitive) `Client Secret` of your PayPal client application.
- `secrets` (String, Sensitive) The secrets that should be passed to the function at runtime.
- `secrets_list` (Block Set) (see [below for nested schema](#nestedblock--config--secrets_list))
- `security_protocol` (String) The security protocol for Kafka interaction.
- `segments` (Set of String)
- `selected_exports` (Set of String)
- `sender_id` (String) Your Sender ID
- `sender_password` (String, Sensitive) Your Sender Password
- `server_url` (String) The Oracle Fusion Cloud Instance URL.
- `servers` (Set of String)
- `service_account` (String) Share the folder with the email address
- `service_account_email` (String) Provide Invoker role to this service account.
- `sftp_host` (String) SFTP host.
- `sftp_is_key_pair` (Boolean) Log in with key pair or password
- `sftp_password` (String, Sensitive) SFTP password required if sftp_is_key_pair is false
- `sftp_port` (Number) SFTP port.
- `sftp_public_key` (String) Public Key
- `sftp_user` (String) SFTP user.
- `share_url` (String) Your SharePoint folder URL. You can find the folder URL by following the steps mentioned [here](/docs/files/share-point/setup-guide).
- `sheet_id` (String) The URL of the sheet that can be copied from the browser address bar, or the ID of the sheet that can be found in the sheet's URL between **/d/** and **/edit**.
- `shop` (String) The Shopify shop name. Can be found in the URL before **.myshopify.com**.
- `short_code` (String, Sensitive) The Salesforce eight-character string assigned to a realm for routing purposes.
- `show_records_with_no_metrics` (Boolean) Turn the toggle on if you want the reports to also return records without metrics.
- `sid` (String) The Twilio API key SID
- `site_id` (String) The Site ID of the SharePoint site from which you want to sync your lists. The Site ID is the `id` field in the [Graph API](https://docs.microsoft.com/en-us/graph/api/site-search?view=graph-rest-1.0&tabs=http) response for sites.
- `site_name` (String) The Name of the SharePoint site. The Site Name is the `name` field in the Graph API response for sites.
- `site_urls` (Set of String)
- `skip_after` (Number) We will skip over the number of lines specified at the end so as to not introduce aberrant data into your destination.
- `skip_before` (Number) We will skip over the number of lines specified before syncing data.
- `snapchat_ads_config_v1_organizations` (Set of String)
- `soap_uri` (String) Marketo SOAP API Endpoint.
- `source` (String) The data source.
- `sub_domain` (String) Your WooCommerce sub-domain.
- `subdomain` (String) Your company's freshservice subdomain (usually **company**.freshservice.com).
- `subscriber_name` (String) The subscriber name. If the connection string does not have manage permission, you need to specify a subscriber name we can use to fetch data. If not specified, we default to `fivetran_sub_<schema>`
- `support_connected_accounts_sync` (Boolean) Sync Connected Accounts. Connected Account Documentation - https://stripe.com/docs/api/connected_accounts.
- `support_nested_columns` (Boolean) This option is to unpack the nested columns and sync them separately. By default, we sync the nested columns as JSON objects.
- `swipe_attribution_window` (String) The time period to attribute conversions based on swipes. Default value: `DAY_28`
- `sync_data_locker` (Boolean) Sync AppsFlyer Data Locker. Default value is `true`, set it to `false` to sync AppsFlyer data using only webhooks.
- `sync_format` (String) The webhooks sync format.  Default value: `Unpacked`. Unpacked messages must be valid JSON.
- `sync_metadata` (Boolean) Parameter defining whether to enable or disable metadata synchronisation. Default value: `TRUE`.
- `sync_method` (String) Sync Method
- `sync_mode` (String) Whether to sync all accounts or specific accounts.
- `sync_pack_mode` (String) The packing mode type. Supported values:<br>`STANDARD_UNPACKED_MODE`- Unpacks _one_ layer of nested fields and infers types.<br>`PACKED_MODE`- Delivers packed data as a single destination column value.<br>Learn more in our [Cosmos DB Sync Pack Mode Options documentation](/docs/databases/cosmos#packmodeoptions).
- `sync_type` (String) Sync type.  Unpacked messages must be valid JSON.
- `table` (String) Destination table. Table is permanent and cannot be changed after connection creation
- `table_name` (String) Name of table in Airtable
- `tde_certificate` (String, Sensitive) Certificate used to protect a database encryption key
- `tde_certificate_name` (String) Name of the Certificate used to protect a database encryption key
- `tde_password` (String, Sensitive) Password of the TDE private key
- `tde_private_key` (String, Sensitive) Private key associated with the TDE certificate
- `technical_account_id` (String) Technical Account ID from the Service Account (JWT) credentials of your Adobe Project.
- `tenant_id` (String, Sensitive) Azure AD tenant ID.
- `tiktok_ads_config_v1_accounts` (Set of String)
- `time_zone` (String) The time zone configured in your Pardot instance. An empty value defaults to `UTC+00:00`.
- `timeframe_months` (String) Historical sync timeframe in months.
- `tns` (String) Single-tenant database: The database's SID. <br> Multi-tenant database: The database's TNS.
- `token_authenticated_container` (String) The container name. Required for the `RESOURCE_TOKEN` data access method.
- `token_authenticated_database` (String) The database name. Required for the `RESOURCE_TOKEN` data access method.
- `token_key` (String, Sensitive) Token ID
- `token_secret` (String, Sensitive) Token Secret
- `topics` (Set of String)
- `trust_store_type` (String) Trust Store Type
- `trusted_cert` (String, Sensitive) Kafka trusted certificate.
- `tunnel_host` (String) SSH host, only specify when connecting via an SSH tunnel (do not use a load balancer). Required for connector creation.
- `tunnel_port` (Number) SSH port, only specify when connecting via an SSH tunnel. Required for connector creation.
- `tunnel_user` (String) SSH user, specify only to connect via an SSH tunnel. Required for connector creation.
- `twilio_config_v1_accounts` (Set of String)
- `twitter_ads_config_v1_accounts` (Set of String)
- `twitter_config_v1_accounts` (Set of String)
- `update_config_on_each_sync` (Boolean) Specifies whether the configuration is updated before each sync or only when the connector settings are saved. This parameter only takes effect when `config_method` is set to `REUSE_EXISTING`. The default value is `true`.
- `update_method` (String) The method to detect new or changed rows. <br>Supported values:<br>`BINLOG` - Fivetran uses your binary logs (also called binlogs) to request only the data that has changed since our last sync. This is the default value if no value is specified. <br>`TELEPORT` - Fivetran's proprietary replication method that uses compressed snapshots to detect and apply changes.
- `uri` (String) Cosmos resource instance address.
- `use_api_keys` (Boolean) Whether to use multiple API keys for interaction.
- `use_customer_bucket` (Boolean) Use Custom Bucket. Set it to 'true' if the data is being synced to your S3 bucket instead of an AppsFlyer-managed bucket.
- `use_oracle_rac` (Boolean) Default value: `false`. Set to `true` if you're using a RAC instance.
- `use_pgp_encryption_options` (Boolean) Set to `true` if files are encrypted using PGP in the S3 bucket. Default value: `false`.
- `use_webhooks` (Boolean) Set to `true` to capture deletes.
- `use_workspace` (Boolean) Choose a database and schema to create temporary tables for syncs.
- `user` (String) The user name.
- `user_id` (String) Marketo SOAP API User Id.
- `user_key` (String, Sensitive)
- `user_name` (String) Workday username.
- `user_profiles` (Set of String)
- `username` (String)
- `view_attribution_window` (String) Time period to attribute conversions based on views. [Possible view_attribution_window values](/docs/applications/facebook-ad-insights/api-config#viewattributionwindow).
- `view_through_attribution_window_size` (String) The time period to attribute conversions based on views. Default value: `DAY_7`
- `webhook_url` (String) The registered URL for webhooks in your Pipedrive dashboard.
- `workspace_name` (String) The name of the database where the temporary tables will be created.
- `workspace_schema` (String) The name of the schema that belongs to the workspace database where the temporary tables will be created.
- `ws_certificate` (String, Sensitive) Web Services Certificate.

<a id="nestedblock--config--adobe_analytics_configurations"></a>
### Nested Schema for `config.adobe_analytics_configurations`

Optional:

- `calculated_metrics` (Set of String)
- `elements` (Set of String)
- `metrics` (Set of String)
- `report_suites` (Set of String)
- `segments` (Set of String)
- `sync_mode` (String) Whether to sync all report suites or specific report suites. Default value: `AllReportSuites` .
- `table` (String) The table name unique within the schema to which connector will sync the data. Required for connector creation.


<a id="nestedblock--config--custom_reports"></a>
### Nested Schema for `config.custom_reports`

Optional:

- `aggregate` (String) Time aggregation of report
- `conversions_report_included` (Boolean) The boolean value specifying whether to enable or disable event conversions data synchronisation. Default value: `false`
- `custom_events_included` (Boolean) The boolean value specifying whether the custom events are included in event conversions report. Default value: `false`
- `dimensions` (Set of String)
- `event_names` (Set of String)
- `level` (String) Level of custom report.
- `metrics` (Set of String)
- `report_fields` (Set of String)
- `report_name` (String) The table name within the schema to which connector syncs the data of the specific report.
- `report_type` (String) Type of report to be generated
- `segmentation` (String) Level of custom report.
- `table_name` (String) Destination Table name of report


<a id="nestedblock--config--custom_tables"></a>
### Nested Schema for `config.custom_tables`

Optional:

- `action_breakdowns` (Set of String)
- `action_report_time` (String) The report time of action stats. [Possible action_report time values](/docs/applications/facebook-ads-insights/api-config#actionreporttime).
- `aggregation` (String) Options to select aggregation duration. [Possible aggregation values](/docs/applications/facebook-ads-insights/api-config#aggregation).
- `breakdowns` (Set of String)
- `click_attribution_window` (String) Time period to attribute conversions based on clicks. [Possible click_attribution_window values](/docs/applications/facebook-ads-insights/api-config#clickattributionwindow).
- `config_type` (String) Option to select Prebuilt Reports or Custom Reports. [Possible config_type values](/docs/applications/facebook-ads-insights/api-config#configtype).
- `fields` (Set of String)
- `level` (String)
- `prebuilt_report_name` (String) The report name to which connector will sync the data. [Possible prebuilt_report values](/docs/applications/facebook-ads-insights/api-config#prebuiltreport).
- `table_name` (String) The table name within the schema to which the connector will sync the data. It must be unique within the connector and must comply with [Fivetran's naming conventions](/docs/getting-started/core-concepts#namingconventions).
- `use_unified_attribution_setting` (Boolean)
- `view_attribution_window` (String) Time period to attribute conversions based on views. [Possible view_attribution_window values](/docs/applications/facebook-ads-insights/api-config#viewattributionwindow).


<a id="nestedblock--config--dimension_filters"></a>
### Nested Schema for `config.dimension_filters`

Optional:

- `dimension` (String) Filtered Dimension.
- `filter_expression` (String) Filter Expression value.
- `match_type` (String) Match type.


<a id="nestedblock--config--project_credentials"></a>
### Nested Schema for `config.project_credentials`

Optional:

- `api_key` (String, Sensitive) The API key of the project.
- `project` (String) The project name you wish to use with Fivetran.
- `secret_key` (String, Sensitive) The secret key of the project.


<a id="nestedblock--config--secrets_list"></a>
### Nested Schema for `config.secrets_list`

Optional:

- `key` (String) Secret Key.
- `value` (String, Sensitive) Secret Value.



<a id="nestedatt--destination_schema"></a>
### Nested Schema for `destination_schema`

Read-Only:

- `name` (String)
- `prefix` (String)
- `table` (String)


<a id="nestedatt--status"></a>
### Nested Schema for `status`

Read-Only:

- `is_historical_sync` (String)
- `setup_state` (String)
- `sync_state` (String)
- `tasks` (List of Object) (see [below for nested schema](#nestedobjatt--status--tasks))
- `update_state` (String)
- `warnings` (List of Object) (see [below for nested schema](#nestedobjatt--status--warnings))

<a id="nestedobjatt--status--tasks"></a>
### Nested Schema for `status.tasks`

Read-Only:

- `code` (String)
- `message` (String)


<a id="nestedobjatt--status--warnings"></a>
### Nested Schema for `status.warnings`

Read-Only:

- `code` (String)
- `message` (String)
