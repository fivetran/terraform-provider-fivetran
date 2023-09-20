# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.0...HEAD)

## [1.1.0](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.0.2...v1.1.0)

## Updated
- Reworked `fivetran_dbt_tranformation` resource: dbt_model_name and dbt_project_id are now used for resource creation instead of dbt_model_id
- Resource `fivetran_dbt_project` updated: it now checks if project was created and initialized correctly. Resource considered as succesfully created only in case if project received `READY` status in upstream. 

## Added
- New resource `fivetran_webhook` that allows to manage Webhooks.
- New data source `fivetran_webhook` that allows to retrieve details of the existing webhook for a given identifier.
- New data source `fivetran_webhooks` that allows to retrieve the list of existing webhooks available for the current account.
- Connector `fivetran_connector.config` fields support:
    - Added field `fivetran_connector.config.custom_field_ids` for services: `double_click_publishers`.
    - Added field `fivetran_connector.config.connecting_user_email` for services: `financial_force`, `salesforce`, `salesforce_sandbox`.
    - Added field `fivetran_connector.config.abs_connection_method` for services: `adobe_analytics_data_feed`.
    - Added field `fivetran_connector.config.abs_public_key` for services: `adobe_analytics_data_feed`.
    - Added field `fivetran_connector.config.abs_host_ip` for services: `adobe_analytics_data_feed`.
    - Added field `fivetran_connector.config.client_key` for services: `appfigures`.
    - Added field `fivetran_connector.config.signer_public_key` for services: `azure_blob_storage`, `s3`, `sftp`.
    - Added field `fivetran_connector.config.enable_archive_log_only` for services: `sql_server_hva`.
    - Added field `fivetran_connector.config.archive_log_path` for services: `sql_server_hva`.
    - Added field `fivetran_connector.config.webhook_key` for services: `xero`.
    - Added field `fivetran_connector.config.abs_host_user` for services: `adobe_analytics_data_feed`.
    - Added field `fivetran_connector.config.sandbox_account` for services: `gocardless`.
    - Added field `fivetran_connector.config.social_data_sync_timeframe` for services: `linkedin_company_pages`.
    - Added field `fivetran_connector.config.are_soap_credentials_provided` for services: `marketo`.
    - Added field `fivetran_connector.config.accounts_reddit_ads` for services: `reddit_ads`.
    - Added field `fivetran_connector.config.personal_access_token` for services: `harvest`.
    - Added field `fivetran_connector.config.custom_events` for services: `iterable`.
    - Added field `fivetran_connector.config.archive_log_format` for services: `sql_server_hva`.
    - Added field `fivetran_connector.config.base_currency` for services: `open_exchange_rates`.
    - Added field `fivetran_connector.config.abs_container_address` for services: `adobe_analytics_data_feed`.
    - Added field `fivetran_connector.config.api_id` for services: `aircall`.
    - Added field `fivetran_connector.config.connecting_user` for services: `financial_force`, `salesforce`, `salesforce_sandbox`.
    - Added field `fivetran_connector.config.custom_event_sync_mode` for services: `iterable`.
    - Added field `fivetran_connector.config.events` for services: `iterable`.

## [1.0.2](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.0.1...v1.0.2)

## Fixed 
- Issue with `connector_schema_config` resource: provider crashes with nil pointer error.

## Added
- New resource `fivetran_dbt_project` that allows to manage dbt Project.
- New data source `fivetran_dbt_projects` that allows to retrieve list of dbt Projects for your account.
- New data source `fivetran_dbt_project` that allows to retrieve dbt Project details.
- New data source `fivetran_dbt_models` that allows to retrieve dbt Models list for specified dbt Project.

## [1.0.1](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.0.0...v1.0.1)

## Added
- New resource `fivetran_dbt_transformation` that allows to manage dbt Transfomrations.
- New data source `fivetran_dbt_transformation` that allows to retrieve dbt Transfomration.

## Fixed 
Resource `fivetran_connector_schema_config` issue with table `sync_mode`:
- Non-empty `sync_mode` value for table causes non-empty plan after each apply

## [1.0.0](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.0.0-pre...v1.0.0)

## Fixed
- Issue with rate limits: now rate limit exceeded error will be automatically handled with retry after back-off period

## [1.0.0-pre](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.7.3...v1.0.0-pre) 

## Fixed
- Resource `fivetran_connector` issue: auth values not applied.

## Added
Supported custom timeouts for the following operations:
- Resource `fivetran_connector`: Create (default: 30 minutes)
- Resource `fivetran_connector`: Update (default: 30 minutes)
- Resource `fivetran_destination`: Create (default: 30 minutes)
- Resource `fivetran_destination`: Update (default: 30 minutes)
- Resource `fivetran_connector_schema_config`: Create (default: 2 hours)
- Resource `fivetran_connector_schema_config`: Read   (default: 2 hours)
- Resource `fivetran_connector_schema_config`: Update (default: 2 hours)

Connector config fields support by Open API schema:
- Added field `fivetran_connector.config.account_name` for services: `talkdesk`.
- Added field `fivetran_connector.config.custom_reports.aggregate` for services: `tiktok_ads`.
- Added field `fivetran_connector.config.blob_sas_url` for services: `webhooks`.
- Added field `fivetran_connector.config.workspace_schema` for services: `snowflake_db`.
- Added field `fivetran_connector.config.keystore` for services: `aws_msk`.
- Added field `fivetran_connector.config.key_password` for services: `aws_msk`.
- Added field `fivetran_connector.config.report_timezone` for services: `criteo`.
- Added field `fivetran_connector.config.reports.aggregation` for services: `google_search_console`.
- Added field `fivetran_connector.config.sender_id` for services: `sage_intacct`.
- Added field `fivetran_connector.config.enable_tde` for services: `sql_server_hva`.
- Added field `fivetran_connector.config.custom_reports.report_type` for services: `tiktok_ads`.
- Added field `fivetran_connector.config.sasl_mechanism` for services: `apache_kafka`.
- Added field `fivetran_connector.config.custom_tables.level` for services: `facebook_ads`.
- Added field `fivetran_connector.config.packing_mode` for services: `firebase`, `mongo`, `mongo_sharded`.
- Added field `fivetran_connector.config.reports.segment_ids` for services: `google_analytics`.
- Added field `fivetran_connector.config.service_account_email` for services: `google_cloud_function`.
- Added field `fivetran_connector.config.phone_number` for services: `itunes_connect`.
- Added field `fivetran_connector.config.tde_certificate_name` for services: `sql_server_hva`.
- Added field `fivetran_connector.config.use_service_account` for services: `bigquery_db`.
- Added field `fivetran_connector.config.sync_pull_api` for services: `appsflyer`.
- Added field `fivetran_connector.config.instance_number` for services: `oracle_sap_hva_netweaver`.
- Added field `fivetran_connector.config.show_records_with_no_metrics` for services: `apple_search_ads`.
- Added field `fivetran_connector.config.advertisers_sync_mode` for services: `google_search_ads_360`.
- Added field `fivetran_connector.config.ad_analytics` for services: `linkedin_ads`.
- Added field `fivetran_connector.config.refresh_token_expires_at` for services: `pinterest_ads`.
- Added field `fivetran_connector.config.webhook_url` for services: `pipedrive`, `segment`.
- Added field `fivetran_connector.config.api_requests_per_minute` for services: `qualtrics`.
- Added field `fivetran_connector.config.sync_multiple_accounts` for services: `reddit_ads`.
- Added field `fivetran_connector.config.access_type` for services: `share_point`.
- Added field `fivetran_connector.config.forecast_id` for services: `clari`.
- Added field `fivetran_connector.config.token_authenticated_container` for services: `cosmos`.
- Added field `fivetran_connector.config.accounts_sync_mode` for services: `google_search_ads_360`.
- Added field `fivetran_connector.config.reports.attributes` for services: `google_search_ads_360`.
- Added field `fivetran_connector.config.table_name` for services: `airtable`.
- Added field `fivetran_connector.config.asb_ip` for services: `azure_service_bus`.
- Added field `fivetran_connector.config.token_id` for services: `chargedesk`, `mux`.
- Added field `fivetran_connector.config.account_sync_mode` for services: `itunes_connect`.
- Added field `fivetran_connector.config.salesforce_security_token` for services: `pardot`.
- Added field `fivetran_connector.config.is_private_link_required` for services: `aws_lambda`.
- Added field `fivetran_connector.config.app_id` for services: `open_exchange_rates`.
- Added field `fivetran_connector.config.enable_exports` for services: `braze`.
- Added field `fivetran_connector.config.sasl_scram256_secret` for services: `apache_kafka`.
- Added field `fivetran_connector.config.store_hash` for services: `big_commerce`.
- Added field `fivetran_connector.config.business_id` for services: `birdeye`.
- Added field `fivetran_connector.config.reports.filter_field_name` for services: `google_analytics_4`.
- Added field `fivetran_connector.config.service_account` for services: `google_drive`.
- Added field `fivetran_connector.config.subscription` for services: `retailnext`.
- Added field `fivetran_connector.config.support_connected_accounts_sync` for services: `stripe`, `stripe_test`.
- Added field `fivetran_connector.config.api_secret_key` for services: `alchemer`.
- Added field `fivetran_connector.config.enriched_export` for services: `optimizely`.
- Added field `fivetran_connector.config.reports.filter_value` for services: `google_analytics_4`.
- Added field `fivetran_connector.config.pgp_pass_phrase` for services: `azure_blob_storage`, `s3`, `sftp`.
- Added field `fivetran_connector.config.workspace_same_as_source` for services: `bigquery_db`.
- Added field `fivetran_connector.config.ad_unit_view` for services: `double_click_publishers`.
- Added field `fivetran_connector.config.folder` for services: `dropbox`.
- Added field `fivetran_connector.config.use_template_labels` for services: `mandrill`.
- Added field `fivetran_connector.config.pem_private_key` for services: `apple_search_ads`.
- Added field `fivetran_connector.config.sender_password` for services: `sage_intacct`.
- Added field `fivetran_connector.config.tde_private_key` for services: `sql_server_hva`.
- Added field `fivetran_connector.config.sasl_plain_key` for services: `apache_kafka`.
- Added field `fivetran_connector.config.template_labels` for services: `mandrill`.
- Added field `fivetran_connector.config.tenant_id` for services: `azure_sql_db`, `azure_sql_managed_db`.
- Added field `fivetran_connector.config.sap_schema` for services: `db2i_hva`.
- Added field `fivetran_connector.config.reports_linkedin_ads` for services: `linkedin_ads`.
- Added field `fivetran_connector.config.login` for services: `the_trade_desk`.
- Added field `fivetran_connector.config.attribution_window_size` for services: `tiktok_ads`.
- Added field `fivetran_connector.config.custom_reports.metrics` for services: `tiktok_ads`.
- Added field `fivetran_connector.config.app_ids_appsflyer` for services: `appsflyer`.
- Added field `fivetran_connector.config.schema_registry_credentials_source` for services: `apache_kafka`, `aws_msk`, `confluent_cloud`.
- Added field `fivetran_connector.config.custom_reports.dimensions` for services: `tiktok_ads`.
- Added field `fivetran_connector.config.base_id` for services: `airtable`.
- Added field `fivetran_connector.config.rollback_window_size` for services: `bingads`.
- Added field `fivetran_connector.config.base_domain` for services: `freshteam`.
- Added field `fivetran_connector.config.rest_api_limit` for services: `pardot`.
- Added field `fivetran_connector.config.survey_ids` for services: `qualaroo`.
- Added field `fivetran_connector.config.webhook_endpoint` for services: `appsflyer`.
- Added field `fivetran_connector.config.bearer_token` for services: `freshchat`, `orbit`, `productboard`, `ada`.
- Added field `fivetran_connector.config.service_account_key` for services: `firebase`.
- Added field `fivetran_connector.config.config_repository_url` for services: `snowplow`.
- Added field `fivetran_connector.config.line_separator` for services: `gcs`, `google_drive`, `box`, `dropbox`, `email`, `ftp`, `share_point`, `azure_blob_storage`, `kinesis`, `s3`, `sftp`.
- Added field `fivetran_connector.config.custom_tables.use_unified_attribution_setting` for services: `facebook_ads`.
- Added field `fivetran_connector.config.admin_api_key` for services: `splitio`.
- Added field `fivetran_connector.config.sasl_plain_secret` for services: `apache_kafka`.
- Added field `fivetran_connector.config.app_specific_password` for services: `itunes_connect`.
- Added field `fivetran_connector.config.use_customer_bucket` for services: `appsflyer`.
- Added field `fivetran_connector.config.use_pgp_encryption_options` for services: `sftp`, `azure_blob_storage`, `s3`.
- Added field `fivetran_connector.config.resource_token` for services: `cosmos`.
- Added field `fivetran_connector.config.application_key` for services: `dear`.
- Added field `fivetran_connector.config.reports.search_types` for services: `google_search_console`.
- Added field `fivetran_connector.config.s3_bucket` for services: `webhooks`.
- Added field `fivetran_connector.config.sasl_scram512_key` for services: `apache_kafka`, `aws_msk`.
- Added field `fivetran_connector.config.abs_prefix` for services: `braze`.
- Added field `fivetran_connector.config.tde_password` for services: `sql_server_hva`.
- Added field `fivetran_connector.config.site_address` for services: `teamwork`.
- Added field `fivetran_connector.config.sftp_public_key` for services: `adobe_analytics_data_feed`.
- Added field `fivetran_connector.config.client_cert` for services: `apache_kafka`, `heroku_kafka`.
- Added field `fivetran_connector.config.sasl_scram256_key` for services: `apache_kafka`.
- Added field `fivetran_connector.config.sasl_scram512_secret` for services: `apache_kafka`, `aws_msk`.
- Added field `fivetran_connector.config.schema_registry_key` for services: `apache_kafka`, `aws_msk`, `azure_service_bus`, `confluent_cloud`.
- Added field `fivetran_connector.config.pgp_secret_key` for services: `azure_blob_storage`, `s3`, `sftp`.
- Added field `fivetran_connector.config.s3_export_role_arn` for services: `braze`.
- Added field `fivetran_connector.config.filter` for services: `google_analytics`.
- Added field `fivetran_connector.config.ws_certificate` for services: `adp_workforce_now`.
- Added field `fivetran_connector.config.trusted_cert` for services: `apache_kafka`, `heroku_kafka`.
- Added field `fivetran_connector.config.data_set_name` for services: `bigquery_db`.
- Added field `fivetran_connector.config.token_authenticated_database` for services: `cosmos`.
- Added field `fivetran_connector.config.token_secret_key` for services: `mux`.
- Added field `fivetran_connector.config.business_unit_id` for services: `pardot`.
- Added field `fivetran_connector.config.custom_reports.table_name` for services: `tiktok_ads`.
- Added field `fivetran_connector.config.content_owner_id` for services: `youtube_analytics`.
- Added field `fivetran_connector.config.is_vendor` for services: `amazon_selling_partner`.
- Added field `fivetran_connector.config.report_format_type` for services: `workday`.
- Added field `fivetran_connector.config.keystore_password` for services: `aws_msk`.
- Added field `fivetran_connector.config.team_id` for services: `asana`.
- Added field `fivetran_connector.config.host_ip` for services: `azure_blob_storage`, `azure_service_bus`.
- Added field `fivetran_connector.config.namespace` for services: `azure_service_bus`.
- Added field `fivetran_connector.config.s3_export_folder` for services: `braze`.
- Added field `fivetran_connector.config.trust_store_type` for services: `heroku_kafka`.
- Added field `fivetran_connector.config.custom_reports` for services: `tiktok_ads`, `reddit_ads`.
- Added field `fivetran_connector.config.schema_registry_secret` for services: `apache_kafka`, `aws_msk`, `azure_service_bus`, `confluent_cloud`.
- Added field `fivetran_connector.config.sync_formula_fields` for services: `financial_force`, `salesforce`, `salesforce_sandbox`.
- Added field `fivetran_connector.config.site_name` for services: `microsoft_lists`.
- Added field `fivetran_connector.config.folder_path` for services: `one_drive`.
- Added field `fivetran_connector.config.is_sailthru_connect_enabled` for services: `sailthru`.
- Added field `fivetran_connector.config.is_custom_api_credentials` for services: `twitter_ads`.
- Added field `fivetran_connector.config.sync_metadata` for services: `facebook_ads`.
- Added field `fivetran_connector.config.is_auth2_enabled` for services: `apple_search_ads`.
- Added field `fivetran_connector.config.workspace_name` for services: `bigquery_db`, `snowflake_db`.
- Added field `fivetran_connector.config.currency` for services: `criteo`.
- Added field `fivetran_connector.config.log_journal` for services: `db2i_hva`.
- Added field `fivetran_connector.config.conversation_webhook_url` for services: `helpscout`.
- Added field `fivetran_connector.config.s3path` for services: `sailthru`.
- Added field `fivetran_connector.config.adobe_analytics_configurations.table` for services: `adobe_analytics`.
- Added field `fivetran_connector.config.container_address` for services: `azure_blob_storage`.
- Added field `fivetran_connector.config.host_user` for services: `azure_service_bus`, `azure_blob_storage`.
- Added field `fivetran_connector.config.key_store_type` for services: `heroku_kafka`.
- Added field `fivetran_connector.config.rfc_library_path` for services: `oracle_sap_hva_netweaver`.
- Added field `fivetran_connector.config.enable_enrichments` for services: `snowplow`.
- Added field `fivetran_connector.config.json_delivery_mode` for services: `ftp`, `google_drive`, `kinesis`, `sftp`, `share_point`, `azure_blob_storage`, `box`, `dropbox`, `email`, `gcs`, `s3`.
- Added field `fivetran_connector.config.encoded_public_key` for services: `apple_search_ads`.
- Added field `fivetran_connector.config.has_manage_permissions` for services: `azure_service_bus`.
- Added field `fivetran_connector.config.access_key_secret` for services: `s3`.
- Added field `fivetran_connector.config.instance_url` for services: `sap_business_by_design`.
- Added field `fivetran_connector.config.api_usage` for services: `zendesk`.
- Added field `fivetran_connector.config.client_cert_key` for services: `heroku_kafka`, `apache_kafka`.
- Added field `fivetran_connector.config.truststore` for services: `aws_msk`.
- Added field `fivetran_connector.config.auth_method` for services: `azure_sql_db`, `azure_sql_managed_db`, `webhooks`.
- Added field `fivetran_connector.config.list_sync_mode` for services: `google_analytics_4_export`.
- Added field `fivetran_connector.config.s3_role_arn` for services: `adjust`, `webhooks`.
- Added field `fivetran_connector.config.query_param_value` for services: `alchemer`, `birdeye`.
- Added field `fivetran_connector.config.log_journal_schema` for services: `db2i_hva`.
- Added field `fivetran_connector.config.company_key` for services: `khoros_care`.
- Added field `fivetran_connector.config.word_press_site_id_or_woocommerce_domain_name` for services: `woocommerce`.
- Added field `fivetran_connector.config.subscriber_name` for services: `azure_service_bus`.
- Added field `fivetran_connector.config.use_workspace` for services: `bigquery_db`, `snowflake_db`.
- Added field `fivetran_connector.config.s3_export_bucket` for services: `braze`.
- Added field `fivetran_connector.config.tde_certificate` for services: `sql_server_hva`.
- Added field `fivetran_connector.config.attribution_window` for services: `amazon_ads`.

## [0.7.3](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.7.2...v0.7.3) 

## Fixed
- State migration issue when switching on v0.7.2 from earlier versions:
    - Previously created configurations now can be upgraded to v0.7.3 directly
    - Newly created configurations with v0.7.2 could be also upgraded to v0.7.3

## [0.7.2](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.7.2-pre...v0.7.2) 

## Fixed
- Issue with drifting changes for `fivetran_connector.config.list_strategy` field
- Issue with re-creation of connectors that are using `destination_schema.prefix` field
- Supported config fields for CosmosDB and Snowflake DB connectors
- Supported missing fields for S3 source connector
- Supported `replica_id` in connector config for MySQL connectors
- Supported `short_code`, `site_id` and `customer_list_id` fields for Salesforce Commerce Cloud

## [0.7.2-pre](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.7.1...v0.7.2-pre)

## Added
- New `fivetran_connector_schedule` resource introduced
- `fivetran_destination.config.fivetran_role_arn` missing field added
- `fivetran_destination.config.prefix_path` missing field added
- `fivetran_destination.config.region` missing field added

## Fixed
- Run setup tests in update resource only if `run_setup_tests` = true is set
- Issue with drifting changes for `fivetran_connector.config.list_strategy` field
- Issue with re-creation of connectors that are using `destination_schema.prefix` field
- Supported config fields for CosmosDB and Snowflake DB connectors

## Breaking changes
- Resource `fivetran_connector` updated
    - Field `fivetran_connector.sync_frequency` moved to `fivetran_connector_schedule` resource
    - Field `fivetran_connector.paused` moved to `fivetran_connector_schedule` resource
    - Field `fivetran_connector.pause_after_trial` moved to `fivetran_connector_schedule` resource
    - Field `fivetran_connector.daily_sync_time` moved to `fivetran_connector_schedule` resource
    - Field `fivetran_connector.schedule_type` moved to `fivetran_connector_schedule` resource
    - Readonly field `fivetran_connector.status` removed
    - Readonly field `fivetran_connector.succeeded_at` removed
    - Readonly field `fivetran_connector.failed_at` removed
    - Readonly field `fivetran_connector.service_version` removed

## [0.7.1](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.7.0...v0.7.1)

Release identical to v0.6.19;

## [0.7.0](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.19...v0.7.0)

Release reverted due to unexpected issues; 

## [0.6.19](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.18...v0.6.19)

## Added
- `fivetran_connector.config.primary_keys` field support

## [0.6.18](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.17...v0.6.18)

## Added
- `fivetran_connector.config.support_nested_columns` field support
- `fivetran_connector.config.csv_definition` field support
- `fivetran_connector.config.export_storage_type` field support

## [0.6.17](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.16...v0.6.17)

## Added
- `fivetran_connector.config.company_id` field support
- `fivetran_connector.config.login_password` field support
- `fivetran_connector.config.environment` field support
- `fivetran_connector.config.properties` field support
- `fivetran_connector.config.is_public` bool field support
- `fivetran_connector.config.empty_header` bool field support
- `fivetran_connector.config.list_strategy` string field support

## [0.6.16](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.15...v0.6.16)

## Added
- `fivetran_connector.config.group_name` field support

## Fixed
- Issue with `fivetran_connector.config.packed_mode_tables` order
- All collections transformed into sets to avoid drifting changes caused by elements order.
- E2E tests updated 

## [0.6.15](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.14...v0.6.15)

## Added
- `fivetran_connector.config.is_single_table_mode` field support

## [0.6.14](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.13...v0.6.14)

## Added
- `fivetran_connector.config.domain_host_name` field support
- `fivetran_connector.config.access_key` field support
- `fivetran_connector.config.client_name` field support
- `fivetran_connector.config.domain_type` field support
- `fivetran_connector.config.connection_method` field support

## [0.6.13](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.12...v0.6.13)

## Added
- `fivetran_connector.config.packed_mode_tables` field support
- `fivetran_connector.config.organization` field support

## [0.6.12](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.11...v0.6.12)

## Added support for HVA connectors
- `fivetran_connector.config.pdb_name` missing field added
- `fivetran_connector.config.agent_host` missing field added
- `fivetran_connector.config.agent_port` missing field added
- `fivetran_connector.config.agent_user` missing field added
- `fivetran_connector.config.agent_password` missing field added
- `fivetran_connector.config.agent_public_cert` missing field added
- `fivetran_connector.config.agent_ora_home` missing field added
- `fivetran_connector.config.tns` missing field added
- `fivetran_connector.config.use_oracle_rac` missing field added
- `fivetran_connector.config.asm_option` missing field added
- `fivetran_connector.config.asm_user` missing field added
- `fivetran_connector.config.asm_password` missing field added
- `fivetran_connector.config.asm_oracle_home` missing field added
- `fivetran_connector.config.asm_tns` missing field added
- `fivetran_connector.config.sap_user` missing field added

## Fixed
- Issue with `fivetran_user.picture`: unable to set value to `null`
- Issue with `fivetran_user.phone`: unable to set value to `null`

## [0.6.11](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.10...v0.6.11)

## Added
- `fivetran_connector.config.sync_method` missing field added
- `fivetran_connector.config.is_account_level_connector` missing field added

## [0.6.10](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.9...v0.6.10)

## Fixed
- Issue: `fivetran_connector.config.pattern` was always set even if it doesn't have value

## Added
- `fivetran_connector.config.is_keypair` missing field added
- `fivetran_connector.config.share_url` missing field added
- `fivetran_connector.config.secrets_list` missing field added

## [0.6.9](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.8...v0.6.9)

## Fixed
- Issue: `fivetran_connector_schema_config` when updating an existing resource
- Issue: `connector_resource.config.use_api_keys` field type handling fixed
- Issue: `connector_resource.config.is_secure` field type handling fixed

## Added
- `fivetran_destination.config.catalog` missing field added

## Updated
- `connector_resource.config` is optional. Connector resource now can be created with empty config

## [0.6.8](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.7...v0.6.8)

## Fixed
- Issue: Unable to create `fivetran_connector_schema_config` resource for newly created connector. 
- Issue: `import` command fails on resource `fivetran_connector` with `Error: Plugin did not respond`.

## [0.6.7](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.6...v0.6.7)

## Fixed
- Make `fivetran_destination.run_setup_tests` optional with default value `false`

## [0.6.6](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.5...v0.6.6)

## Fixed
- Issue with plugin crash on `fivetran_destination` resource import

## [0.6.5](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.4...v0.6.5)

## Fixed
- Fixed reading `destination_resource.config.is_private_key_encrypted` field
- Fixed updating `daily_sync_time` field

## [0.6.4](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.3...v0.6.4)

## Added
- `destination_resource.config.role` missing field added (Snowflake)
- `destination_resource.config.is_private_key_encrypted` missing field added (Snowflake)
- `destination_resource.config.passphrase` missing field added (Snowflake)
- Documentation for `daily_sync_time` field added 

## [0.6.3](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.3...v0.6.2)

## Fixed 
- Importing resource `fivetran_connector_schema_config` issue 

## [0.6.2](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.2...v0.6.1)

## Fixed
- Issue with `daily_sync_time` when `sync_frequency` is set to 1440 

## Added
- Resource `fivetran_connector_schema_config` now supports `table.sync_mode`

## [0.6.1](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.1...v0.6.0)

## Fixed
- Added missing `destination_resource.public_key` readonly field
- Added missing `destination_resource.private_key` field
- Issue with `data_set_location` field when configuring `big_query` destination

## [0.6.0](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.0...v0.5.4)

## Added
- New resource `fivetran_connector_schema_config` added

## [0.5.4](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.5.4...v0.5.3) - 2022-07-20

## Fixed
- Added missing `connector_resource.config.token_key` field.
- Added missing `connector_resource.config.token_secret` field.

## [0.5.3](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.5.3...v0.5.2) - 2022-07-13

## Fixed
- Fix drifting for `connector_resource.config.function_trigger` field.
- Handle `connector_resource.config.function_trigger` as sensitive field.

## [0.5.2](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.5.2...v0.5.1) - 2022-07-06

## Fixed
- Added missing `connector_resource.config.pat` field (personal access token for github connector).

## [0.5.1](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.5.1...v0.5.0) - 2022-06-16

## Fixed
- Added missing `connector_resource.config.eu_region` field.
- Field `user_resource.role` is optional.
- Field `connector_resource.config.pattern` may have empty value.
- Fixed provider behavior in case of resource existing in state is missing in upstream infrastructure

## [0.5.0](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.4.5...v0.5.0) - 2022-06-15

## Breaking changes

- Resource `fivetran_group` separated to two resources `fivetran_group` and `fivetran_group_users` corresponding to dataSources
- Schema field `fivetran_group.creator` removed. 

### Migration from v0.4.5

For each group in configuration the following should be done:

```
v0.4.5:

# Example user
resource "fivetran_user" "user1" {
    email = "email1@domain.com"
    family_name = "User 1"
    given_name = "User Name 1"
    phone = "+123 45 678 8990"
    role = "Owner"
}

## fivetran_group
resource "fivetran_group" "group1" {
    name = "My Group 1"

    user {
        id = fivetran_user.user1.id
        role = "<Some Role>"
    }

}

v0.5.0:

# Example user
resource "fivetran_user" "user1" {
    email = "email1@domain.com"
    family_name = "User 1"
    given_name = "User Name 1"
    phone = "+123 45 678 8990"
    role = "Owner"
}

## fivetran_group
resource "fivetran_group" "group1" {
    name = "My Group 1"
}

resource "fivetran_group_users" "group1_users"{
    group_id = fivetran_group.group1.id

    user {
        id = fivetran_user.user1.id
        role = "<Some Role>"
    }
}

```

NOTE: please remove old `fivetran_group` resource form state and re-import it after provider version update to avoid state inconsistency.
To import group users just import `fivetran_group_users` resource with the same group id.

## Fixed
- Destination resource `trust_certificates`, `trust_fingerprints` and `run_setup_tests` properties don't have `ForceNew` attribute no more.

## [0.4.6](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.4.5...v0.4.6) - 2022-06-14

## Compatibility changes
- Handle `adwords` service migration to `google_ads` for existing connectors. 
- Deprecate the `adwords` service in favor of the new `google_ads` service. 

NOTE: All connector creation requests with the service `adwords` will now result in an error.

## [0.4.5](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.4.4...v0.4.5) - 2022-06-10

## Fixed
- Issue with `external_id` resource_connector config field.
- Added missing `connector_resource.config.publication_name` field.

## [0.4.4](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.4.3...v0.4.4) - 2022-06-01

## Documentation
- `connector_resource` documentation update with information about how to solve indirect dependency between connector and destination

## [0.4.3](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.4.2...v0.4.3) - 2022-05-24

## Fixed
- Issue with `update_method` connector config field (now it can be effectively updated)
- Issue with `connection_type` field isn't marked as readonly any more
- Type issue with `resource_destination.create_external_tables`
- Issue with `username` field mapping  

## [0.4.2](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.4.1...v0.4.2) - 2022-05-18

## Fixed
- Issue with `resource_connector.config.skip_before` response value type 
- Issue with `resource_connector.config.skip_after`  response value type

## [0.4.1](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.4.0...v0.4.1) - 2022-05-06

## Fixed
- Upgrading `go-getter` to 1.5.11 in order to address a [dependency security vulnerability](https://nvd.nist.gov/vuln/detail/CVE-2022-29810)

## [0.4.0](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.3.6...v0.4.0) - 2022-05-04

Considering this version as BETA.

## Breaking changes
- New `destination_schema` field for determining `schema`, `table` and `schema_prefix` outside `config` to prevent drifting changes.
- Changes in `destination_schema` leads to resource replacement

### Migration from v0.3.6

You should move the following fields in `connector_resource` configurations:
- `connector_resource.config.schema` -> `connector_resource.destination_schema.name`
- `connector_resource.config.table` -> `connector_resource.destination_schema.table`
- `connector_resource.config.schema_prefix` -> `connector_resource.destination_schema.schema_prefix`

The following field is now excluded from `connector_resource` schema:
- `connector_resource.schema` - replaced with `name` field

The following Computed field was added to `connector_resource` schema:
- `connector_resource.name` - this field contains resulting Fivetran Connector Name you can see on Fivetran Dashboard UI

Example:

```
v0.3.6 :
resource "fivetran_connector" "postgres" {
    group_id = fivetran_group.my_group.id
    service = "postgres"
    sync_frequency = 5
    paused = false
    pause_after_trial = false
    schema = "production_pg"
    config {
        schema_prefix = "production_pg"
        host = "123.456.789.012"
        port = "5432"
        user = "postgres"
        password = "IDontKnowThePassword"
        database = "prod"
        update_method = "XMIN"
    }
}

resource "fivetran_connector" "google_sheets" {
    group_id = fivetran_group.my_group.id
    service = "google_sheets"
    sync_frequency = 5
    paused = false
    pause_after_trial = false
    schema = "connector_schema_name.table_name"
    config {
        schema = "connector_schema_name"
        table = "table_name"
        sheet_id = "1Rmq_FN2kTNwWiT4adZKBxHBBlaHBLAHBLAH..."
        named_range = "Some Range Name"
    }
}

v0.4.0 :
resource "fivetran_connector" "postgres" {
    group_id = fivetran_group.my_group.id
    service = "postgres"
    sync_frequency = 5
    paused = false
    pause_after_trial = false
    destination_schema {
        prefix = "production_pg"
    } 
    config {
        host = "123.456.789.012"
        port = "5432"
        user = "postgres"
        password = "IDontKnowThePassword"
        database = "prod"
        update_method = "XMIN"
    }
}

resource "fivetran_connector" "google_sheets" {
    group_id = fivetran_group.my_group.id
    service = "google_sheets"
    sync_frequency = 5
    paused = false
    pause_after_trial = false
    destination_schema {
        name = "connector_schema_name"
        table = "table_name"
    }
    config {
        sheet_id = "1Rmq_FN2kTNwWiT4adZKBxHBBlaHBLAHBLAH..."
        named_range = "Some Range Name"
    }
}
```


## Fixed
- All sensitive fields marked as sensitive in connector_resource
- Minor connector resource fixes

## [0.3.6](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.3.5...v0.3.6) - 2022-04-26

## Fixed
- Fixed auth fields mapping in `client_access` schema

## [0.3.5](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.3.4...v0.3.5) - 2022-04-20

## Fixed
- `ConnectorConfigRequest.AlwaysEncrypted` missing field added
- `ConnectorConfigRequest.FolderId` missing field added

## [0.3.4](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.3.3...v0.3.4) - 2022-02-22

## Fixed
- `ConnectorConfigRequest.BaseUrl` missing field added
- `ConnectorConfigRequest.EntityId` missing field added
- `ConnectorConfigRequest.SoapUri` missing field added
- `ConnectorConfigRequest.UserId` missing field added
- `ConnectorConfigRequest.EncryptionKey` missing field added
- `ConnectorCreateRequest.DailySyncTime` missing field added

## [0.3.3](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.3.2...v0.3.3) - 2022-02-22

## Fixed
- `ConnectorConfigRequest.ApiType` missing field added

## [0.3.2](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.3.1...v0.3.2) - 2022-02-10

## Fixed
- `ConnectorConfigRequest.IsMultiEntityFeatureEnabled` missing field added

## [0.3.1](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.3.0...v0.3.1) - 2022-01-31

## Fixed
- `ConnectorConfigRequest.ConnectionType` missing field added

## [0.3.0](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.2.3...v0.3.0) - 2022-01-25

## Added
- E2E tests.
- GitHub actions workflow to run tests.

## Fixed
- `ConnectorConfigRequest.IsNewPackage` missing field added
- `ConnectorConfigRequest.AdobeAnalyticsConfigurations` missing field added
- Enabled Role management for users

## [0.2.3](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.2.2...v0.2.3) - 2022-01-10

## Fixed
- Crash on `terraform import fivetran_destination.name <group_id>`

## [0.2.2](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.2.1...v0.2.2) - 2021-12-31

## Fixed
- `cluster_id` field marked as optional in source destination schema.
- `cluster_region` field marked as optional in source destination schema.

## [0.2.1](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.2.0...v0.2.1) - 2021-12-13

## Fixed
- `cluster_id` missing field added to source destination schema.
- `cluster_region` missing field added to source destination schema.

## [0.2.0](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.1.2...v0.2.0) - 2021-11-10

## Added
- Custom User-Agent tag provided to track requests from terraform.


## [0.1.2](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.1.1...v0.1.2) - 2021-09-30

## Fixed
- `host` field added to resource connector config schema.

## [0.1.1](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.1.0...v0.1.1) - 2021-09-28

## Fixed
- `secret_key` missing field added to source destination schema.

## [0.1.0](https://github.com/fivetran/terraform-provider-fivetran/releases/tag/v0.1.0) - 2021-07-27

Initial release. 

### Added

- Resources: `fivetran_user`, `fivetran_group`, `fivetran_destination`, `fivetran_connector`
- Data Sources: `fivetran_user`, `fivetran_users`, `fivetran_group`, `fivetran_groups`, `fivetran_group_connectors`, `fivetran_group_users`, `fivetran_destination`, `fivetran_connectors_metadata`, `fivetran_connector`
