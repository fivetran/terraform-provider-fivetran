# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.2.7...HEAD)

## [1.2.7](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.2.6...v1.2.7)

## Fixed
- Issue with rewriting value in `token` field in `fivetran_proxy_agent` resource

## [1.2.6](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.2.5...v1.2.6)

Release fix

## [1.2.5](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.2.4...v1.2.5)

Broken release, please don't use it

## [1.2.4](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.2.3...v1.2.4)

## Updated
- Documentation on using fields `fivetran_connector.networking_method` and `fivetran_destination.networking_method` has been clarified
- Added documentation on migration to version 1.1.18

New connection services supported:
- Supported service: `anvyl`
- Supported service: `boostr`
- Supported service: `cloudbeds`
- Supported service: `databricks_db`
- Supported service: `exact_online`
- Supported service: `factbird`
- Supported service: `invoiced`
- Supported service: `medallia_agile_research`
- Supported service: `rundeck`
- Supported service: `sigma_computing_source`
- Supported service: `smartrecruiters`
- Supported service: `splash`
- Supported service: `statistics_netherlands_cbs`
- Supported service: `teamtailor`
- Supported service: `xandr`
- Supported service: `zoho_inventory`

New connector config fields supported:
- Added field `fivetran_connector.config.service_account_username` for services: `mixpanel`.
- Added field `fivetran_connector.config.identifier` for services: `statistics_netherlands_cbs`.
- Added field `fivetran_connector.config.property_id` for services: `cloudbeds`.
- Added field `fivetran_connector.config.http_path` for services: `databricks_db`.
- Added field `fivetran_connector.config.x_key` for services: `medallia_agile_research`.
- Added field `fivetran_connector.config.type_name` for services: `akamai`, `bubbleio`.
- Added field `fivetran_connector.config.quota_project_id` for services: `bigquery_db`.
- Added field `fivetran_connector.config.skip_empty_reports` for services: `google_ads`.
- Added field `fivetran_connector.config.x_master_key` for services: `medallia_agile_research`.
- Added field `fivetran_connector.config.tenant_configs` for services: `reltio`.
- Added field `fivetran_connector.config.snc_source_name` for services: `sap_hana_db`.
- Added field `fivetran_connector.config.api_server` for services: `sigma_computing_source`.
- Added field `fivetran_connector.config.rollback_window` for services: `appsflyer`.
- Added field `fivetran_connector.config.private_token` for services: `eventbrite`.
- Added field `fivetran_connector.config.service_account_secret` for services: `mixpanel`.
- Added field `fivetran_connector.config.tenant_app_url` for services: `planful`.
- Added field `fivetran_connector.config.catalog` for services: `databricks_db`.
- Added field `fivetran_connector.config.access_id` for services: `planful`.
- Added field `fivetran_connector.config.max_api_requests_per_day` for services: `reltio`.
- Added field `fivetran_connector.config.snc_certificate_source` for services: `sap_hana_db`.
- Added field `fivetran_connector.config.api_key` for services: `elastic_cloud`, `es_self_hosted`.

New destination config fields supported:
- Added field `fivetran_connector.config.snapshot_retention_period` for services: `adls`, `new_s3_datalake`, `onelake`.
- Added field `fivetran_connector.config.table_format` for services: `new_s3_datalake`.

## [1.2.3](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.2.2...v1.2.3)

## Fixed

Issue with destination resource state migration error `error decoding object; expected 14 attributes, got 12` that occures on provider version upgrade.

## [1.2.2](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.2.1...v1.2.2)

## Fixed

Issue `panic: Can't convert value  to int` on version upgrade for some connectors.

## [1.2.1](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.2.0...v1.2.1)

## Fixed

Issue with upgrade `fivetran_connector` resource state on version upgrade.

## [1.2.0](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.27...v1.2.0)

## Added
- New resource `fivetran_local_processing_agent` that allows to manage Local Processing Agents.
- New data source `fivetran_local_processing_agent` that allows to retrieve details of the existing Local Processing Agent for a given identifier.
- New data source `fivetran_local_processing_agents` that allows to retrieve the list of existing Local Processing Agents available for the current account.

## Updated
Updates to support management of local processing agents:
- Resource `fivetran_connector` updates:
  - Added field `fivetran_connector.local_processing_agent_id`.

- Resource `fivetran_destination` updates:
  - Added field `fivetran_destination.local_processing_agent_id`.
  - Added field `fivetran_destination.networking_method`.

- Datasource `fivetran_connector` updates:
  - Added field `fivetran_connector.local_processing_agent_id`.

- Datasource `fivetran_destination` updates:
  - Added field `fivetran_destination.local_processing_agent_id`.
  - Added field `fivetran_destination.networking_method`.

- Resource `fivetran_connector_schema_config` reworked, added new validation logic and new field `validation_level`.
    - `fivetran_connector_schema_config.validation_level` allows to setup desired level of schema validation before apply
        - NONE: no validation needed, the fastest way to apply schema, especially for a newly created connector - it will use new [Create a Connector Schema Config](https://fivetran.com/docs/rest-api/connectors#createaconnectorschemaconfig) endpoint.
        - TABLES: validate only schemas and tables names. This level will require only schema reloading, without requests to retrieve columns for every configured table
        - COLUMNS: performs full configuration validation, but requires to make requests to the source to retrieve all columns for every table.
- Resource `fivetran_connector_schema_config` won't fail if some parts of schema were deleted from source and now able to handle such changes. 
    - Provider will show warnings in output if some configures schema element doesn't exist anymore in source schema.
    - Provider will fail on attempt to patch not existing schema, table or columns if appropriate `validation_level` specified and will prompt to update resource configuration.

## Fixed
- Remove Default value from `networking_method` field in `fivetran_connector`

## [1.1.26](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.25...v1.1.26)

## Added
- New resource `fivetran_proxy_agent` that allows to manage Proxy Agents.
- New data source `fivetran_proxy_agent` that allows to retrieve details of the existing Proxy Agent for a given identifier.
- New data source `fivetran_proxy_agents` that allows to retrieve the list of existing Proxy Agents available for the current account.

Updates to support management of proxies:
- Resource `fivetran_connector` updates:
    - Added field `fivetran_connector.proxy_agent_id`.
    - Added field `fivetran_connector.networking_method`.

- Datasource `fivetran_connector` updates:
    - Added field `fivetran_connector.proxy_agent_id`.
    - Added field `fivetran_connector.networking_method`.

New connection services supported:
- Supported service: `akamai`
- Supported service: `brave_ads`
- Supported service: `bubbleio`
- Supported service: `canvas_by_instructure`
- Supported service: `cin7core`
- Supported service: `clarity`
- Supported service: `cockroachdb`
- Supported service: `doorloop`
- Supported service: `freshsales`
- Supported service: `kevel`
- Supported service: `planful`
- Supported service: `rebound_returns`
- Supported service: `skillstx`
- Supported service: `smarthr`
- Supported service: `ticket_tailor`

New connector config fields supported:
- Added field `fivetran_connector.config.snc_certificate_source` for services: `sap_hana_db`.
- Added field `fivetran_connector.config.type_name` for services: `akamai`, `bubbleio`.
- Added field `fivetran_connector.config.rollback_window` for services: `appsflyer`.
- Added field `fivetran_connector.config.custom_tables.engaged_view_attribution_window` for services: `facebook_ads`.
- Added field `fivetran_connector.config.server_variable` for services: `myosh`.
- Added field `fivetran_connector.config.snc_source_name` for services: `sap_hana_db`.
- Added field `fivetran_connector.config.snc_fivetran_name` for services: `sap_hana_db`.

## [1.1.25](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.24...v1.1.25)

## Fixed

- Fix issue when provider tries to set readonly config fields to `null` on update for connector/destination.

## [1.1.24](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.23...v1.1.24)

## Fixed

Field `fivetran_connector.config.api_usage` handling issue (field is string, but should be handled as float)
Field `fivetran_connector.config.api_utilization_percentage` handling issue (field is string, but should be handled as float)

## [1.1.23](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.22...v1.1.23)

## Updated
- Migration to terraform-provider-framework finished.

## Added
New connection services supported:
- Supported service: `bizzabo`
- Supported service: `ebay`
- Supported service: `freshsuccess`
- Supported service: `lemlist`
- Supported service: `microsoft_dynamics_365_fno`
- Supported service: `myosh`
- Supported service: `picqer`
- Supported service: `proofpoint_security_awareness`
- Supported service: `solarwinds_service_desk`
- Supported service: `toast`
- Supported service: `tymeshift`
- Supported service: `veeqo`
- Supported service: `veevavault`
- Supported service: `vonage_contact_center`
- Supported service: `yougov_sport`

New connector config fields supported:
- Added field `fivetran_connector.config.api` for services: `freshsuccess`.
- Added field `fivetran_connector.config.db` for services: `myosh`.
- Added field `fivetran_connector.config.snc_certificate` for services: `sap_hana_db`.
- Added field `fivetran_connector.config.snc_mode` for services: `sap_hana_db`.
- Added field `fivetran_connector.config.snc_my_name` for services: `sap_hana_db`.
- Added field `fivetran_connector.config.toast_id` for services: `toast`.
- Added field `fivetran_connector.config.account_plan` for services: `tymeshift`.
- Added field `fivetran_connector.config.primary_key` for services: `care_quality_commission`.
- Added field `fivetran_connector.config.scope` for services: `ebay`.
- Added field `fivetran_connector.config.convert_dats_type_to_date` for services: `sap_hana_db`.

New destination config fields supported:
- Added field `fivetran_destination.config.oauth2_client_id` for services: `databricks`.
- Added field `fivetran_destination.config.oauth2_secret` for services: `databricks`.

## [1.1.22](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.21...v1.1.22)

## Fixed 
- Issue with empty string and list values in config. API doesn't return empty values in response, so we had inconsistent state after apply.
- Issue with disabled schemas: disabled schemas were added to state because of configured tables in upstream.

## [1.1.21](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.20...v1.1.21)

## Added
- Output warnings with test details in case if setup tests for instance (connector, destination or external_logger) are failing.
- New connector types supported:
    - Supported service: `7shifts`
    - Supported service: `commercetools`
    - Supported service: `gainsight_product_experience`
    - Supported service: `open_data_dc`
    - Supported service: `testrail`
    - Supported service: `totango`
- New connector config fields supported:
    - Added field `fivetran_connector.config.project_key` for services: `commercetools`.
    
## Fixed
- Issue with sensitive fields for different connectors (same field might be sensitive or not for different connector types)
- Issue with sensitive collections (`pats` for `github` connector)

## [1.1.20](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.19...v1.1.20)

## Fixed
- Issue with unknown services: no more panic.
- Issue with panic on omited `destination_schema` field for connector resource.
- Issue with missing or empty `config` field for connector resource.
- Issue with `fivetran_connector` resource import.

## Docs
- Added a guide about configuring dbt Project with a private repo.

## [1.1.19](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.18...v1.1.19)

## Added
- New resource `fivetran_user_connector_membership` that allows to manage User Connector memberships.
- New resource `fivetran_user_group_membership` that allows to manage User Group memberships.
- New data source `fivetran_user_connector_memberships` that allows to retrieve details of the existing user connector memebrships for a given identifier.
- New data source `fivetran_user_group_memberships` that allows to retrieve the list of existing user group memberships available for the current account.

## Updated
Resources:
- `fivetran_team`
- `fivetran_team_user_membership`
- `fivetran_team_connector_membership`
- `fivetran_team_group_membership`
- `fivetran_webhook`
- `fivetran_external_logging`
- `fivetran_group`
- `fivetran_group_users`
  migrated on `terraform-plugin-framework`

Datasources:
- `fivetran_team`
- `fivetran_teams`
- `fivetran_team_group_memberships`
- `fivetran_team_connector_memberships`
- `fivetran_team_group_memberships`
- `fivetran_webhook`
- `fivetran_webhooks`
- `fivetran_external_logging`
- `fivetran_connectors_metadata`
- `fivetran_roles`
- `fivetran_group`
- `fivetran_groups`
- `fivetran_group_connectors`
- `fivetran_group_users`
  migrated on `terraform-plugin-framework`

## [1.1.18](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.17...v1.1.18)

## Added
- Added field `fivetran_destination.daylight_saving_time_enabled` for shifting UTC offset with daylight savings time (US Only).

## Updated
- Schema updated for resource `fivetran_connector_schema_config`:
    - Field `schema` deprecated due to issue with types.SetTypable performance.
    - Field `schemas` of Map type added instead.
    - Field `schemas_json` added to apply settings directly from `.json` resource file.
- Resource `fivetran_connector_schema_config` behavior changed:
    - If no columns settings specified in `table.columns` no settings will be applied. If table enabled - columns won't be blocked automatically by BLOCK_ALL policy.
    - Settings for sub-elements won't be managed if root element disabled: for BLOCK_ALL policy for disabled schema no settings for tables/columns will be applied.
- Resource `fivetran_connector` schema updated:
    - Added field `fivetran_connector.config.authentication_method` for services: `adobe_analytics`.
    - Added field `fivetran_connector.config.client_public_certificate` for services: `aurora_postgres`, `azure_postgres`, `google_cloud_postgresql`, `heroku_postgres`, `postgres`, `postgres_rds`.
    - Added field `fivetran_connector.config.client_private_key` for services: `postgres_rds`, `aurora_postgres`, `azure_postgres`, `google_cloud_postgresql`, `heroku_postgres`, `postgres`.
    - Added field `fivetran_connector.config.api_password` for services: `duoplane`.
- Following new connector types supported:
    - Supported service: `absorb_lms`
    - Supported service: `chameleon`
    - Supported service: `donus`
    - Supported service: `duoplane`
    - Supported service: `eventsforce`
    - Supported service: `freshdesk_contact_center`
    - Supported service: `g2`
    - Supported service: `google_tasks`
    - Supported service: `instructure`
    - Supported service: `partnerize`
    - Supported service: `quora_ads`
    - Supported service: `quorum`
    - Supported service: `referralhero`
    - Supported service: `sap_success_factors`
    - Supported service: `sistrix`
    - Supported service: `web_scraper`

## [1.1.17](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.16...v1.1.17)

## Fixed
- Issue with too slow `connector_schema_config` resource while applying config 

## [1.1.16](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.15...v1.1.16)

## Fixed 
- Issue with panic in case if connector doesn't return `sync_mode` in schema response.
- Fix issue with connector import: config not imported into state.

## [1.1.15](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.13...v1.1.15)

## Fixed
- Issue with `run_setup_tests`, `trust_certificates` and `trust_fingerprints` fields
- Issue with `405` errors on resource update

## New supported connector service types
- Supported service: `adobe_workfront`
- Supported service: `azure_cosmos_for_mongo`
- Supported service: `callrail`
- Supported service: `clubspeed`
- Supported service: `datadog`
- Supported service: `electronic_tenant_solutions`
- Supported service: `firehydrant`
- Supported service: `fourkites`
- Supported service: `genesys`
- Supported service: `gmail`
- Supported service: `livechat_partner`
- Supported service: `mambu`
- Supported service: `revenuecat`
- Supported service: `ricochet360`
- Supported service: `rithum`
- Supported service: `sharetribe`
- Supported service: `sparkpost`
- Supported service: `starrez`
- Supported service: `teads`
- Supported service: `visit_by_ges`
- Supported service: `walmart_marketplace`

## Updated `config` schema for connector resource
- Added field `fivetran_connector.config.account_region` for services: `iterable`.
- Added field `fivetran_connector.config.tenant_name` for services: `mambu`.
- Added field `fivetran_connector.config.direct_capture_method` for services: `oracle_hva`, `oracle_sap_hva`.
- Added field `fivetran_connector.config.is_sftp_creds_available` for services: `salesforce_marketing_cloud`.
- Added field `fivetran_connector.config.legal_entity_id` for services: `younium`.
- Added field `fivetran_connector.config.organization_domain` for services: `adobe_workfront`.
- Added field `fivetran_connector.config.custom_url` for services: `dbt_cloud`.
- Added field `fivetran_connector.config.pats` for services: `github`.

## [1.1.14](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.13...v1.1.14)

## Skipped due to issues with publishing new versions to terraform registry

## [1.1.13](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.12...v1.1.13)

## Updated `config` schema for connector resource
- Added field `fivetran_connector.config.company_ids` for services: `cj_commission_detail`.
- Added field `fivetran_connector.config.target_entity_id` for services: `culture_amp`.
- Added field `fivetran_connector.config.url_format` for services: `fountain`.
- Added field `fivetran_connector.config.should_sync_events_with_deleted_profiles` for services: `klaviyo`.
- Added field `fivetran_connector.config.pull_archived_campaigns` for services: `outbrain`.
- Added field `fivetran_connector.config.store_id` for services: `reviewsio`.
- Added field `fivetran_connector.config.non_standard_escape_char` for services: `s3`.
- Added field `fivetran_connector.config.product` for services: `webconnex`.
- Added field `fivetran_connector.config.auth_environment` for services: `younium`.
- Added field `fivetran_connector.config.service_authentication` for services: `dsv`.
- Added field `fivetran_connector.config.subscription_key` for services: `dsv`.
- Added field `fivetran_connector.config.escape_char_options` for services: `s3`.

## [1.1.12](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.11...v1.1.12)

## Updated
- Destination Resource migrated on `terraform-plugin-framework`
- Destination Datasource migrated on `terraform-plugin-framework`

## Fixed
- Issue with `daily_sync_time` in `connector_schedule` resource
- Issue with fields in connector config that are not managed by configuration, but returned from upstream (non-nullable)
- Issue with object collection fields with sensitive sub-fields

## [1.1.11](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.10...v1.1.11)

## Added
New supported connector services
- Supported service: `6sense`
- Supported service: `acumatica`
- Supported service: `adobe_commerce`
- Supported service: `affinity`
- Supported service: `afterpay`
- Supported service: `aha`
- Supported service: `algolia`
- Supported service: `amazon_attribution`
- Supported service: `attio`
- Supported service: `aumni`
- Supported service: `auth0`
- Supported service: `autodesk_bim_360`
- Supported service: `avantlink`
- Supported service: `aws_cost`
- Supported service: `azure_boards`
- Supported service: `azure_devops`
- Supported service: `billing_platform`
- Supported service: `bitly`
- Supported service: `buildkite`
- Supported service: `business_central`
- Supported service: `campaignmonitor`
- Supported service: `castor_edc`
- Supported service: `ceridian_dayforce`
- Supported service: `chartmogul`
- Supported service: `checkout`
- Supported service: `checkr`
- Supported service: `chorusai`
- Supported service: `cimis`
- Supported service: `cj_commission_detail`
- Supported service: `close`
- Supported service: `coassemble`
- Supported service: `codefresh`
- Supported service: `column`
- Supported service: `concord`
- Supported service: `confluence`
- Supported service: `contrast_security`
- Supported service: `convex`
- Supported service: `copper`
- Supported service: `crowddev`
- Supported service: `cvent`
- Supported service: `d2l_brightspace`
- Supported service: `db2`
- Supported service: `db2i_sap_hva`
- Supported service: `drata`
- Supported service: `dropbox_sign`
- Supported service: `dsv`
- Supported service: `economic`
- Supported service: `expensify`
- Supported service: `ezofficeinventory`
- Supported service: `factorial`
- Supported service: `fone_dynamics`
- Supported service: `freightview`
- Supported service: `getfeedback`
- Supported service: `gitlab`
- Supported service: `google_business_profile`
- Supported service: `green_power_monitor`
- Supported service: `grepsr`
- Supported service: `grin`
- Supported service: `hana_sap_hva_b1`
- Supported service: `hana_sap_hva_ecc`
- Supported service: `hana_sap_hva_ecc_netweaver`
- Supported service: `hana_sap_hva_s4`
- Supported service: `hana_sap_hva_s4_netweaver`
- Supported service: `happyfox`
- Supported service: `helpshift`
- Supported service: `ilevel`
- Supported service: `incidentio`
- Supported service: `infobip`
- Supported service: `integrate`
- Supported service: `ironsource`
- Supported service: `ivanti`
- Supported service: `jotform`
- Supported service: `keypay`
- Supported service: `klarna`
- Supported service: `konnect_insights`
- Supported service: `learnupon`
- Supported service: `liftoff`
- Supported service: `linksquares`
- Supported service: `lob`
- Supported service: `maxio_chargify`
- Supported service: `maxio_saasoptics`
- Supported service: `megaphone`
- Supported service: `meltwater`
- Supported service: `microsoft_teams`
- Supported service: `mode`
- Supported service: `moloco`
- Supported service: `monday`
- Supported service: `nylas`
- Supported service: `oracle_business_intelligence_publisher`
- Supported service: `oracle_moat_analytics`
- Supported service: `ordway`
- Supported service: `paychex`
- Supported service: `persona`
- Supported service: `personio`
- Supported service: `pingdom`
- Supported service: `pinpoint`
- Supported service: `pinterest_organic`
- Supported service: `pipe17`
- Supported service: `pivotal_tracker`
- Supported service: `piwik_pro`
- Supported service: `planetscale`
- Supported service: `postmark`
- Supported service: `prive`
- Supported service: `rakutenadvertising`
- Supported service: `ramp`
- Supported service: `rarible`
- Supported service: `redshift_db`
- Supported service: `reltio`
- Supported service: `replyio`
- Supported service: `resource_management_by_smartsheet`
- Supported service: `revops`
- Supported service: `revx`
- Supported service: `ringover`
- Supported service: `rocketlane`
- Supported service: `rtb_house`
- Supported service: `safetyculture`
- Supported service: `sage_hr`
- Supported service: `sap_hana`
- Supported service: `sap_s4hana`
- Supported service: `sensor_tower`
- Supported service: `servicetitan`
- Supported service: `shopware`
- Supported service: `shortcut`
- Supported service: `shortio`
- Supported service: `simplecast`
- Supported service: `slab`
- Supported service: `spotify_ads`
- Supported service: `sprout`
- Supported service: `sql_server_sap_ecc_hva`
- Supported service: `standard_metrics`
- Supported service: `statsig`
- Supported service: `statuspage`
- Supported service: `swoogo`
- Supported service: `talkwalker`
- Supported service: `thinkific`
- Supported service: `transcend`
- Supported service: `ukg_pro`
- Supported service: `unicommerce`
- Supported service: `vitally`
- Supported service: `vonage`
- Supported service: `vts`
- Supported service: `vwo`
- Supported service: `wasabi_cloud_storage`
- Supported service: `wordpress`
- Supported service: `workday_financial_management`
- Supported service: `workday_strategic_sourcing`
- Supported service: `workflowmax`
- Supported service: `workramp`
- Supported service: `xray`
- Supported service: `xsolla`
- Supported service: `yahoo_dsp`
- Supported service: `yahoo_search_ads_yahoo_japan`
- Supported service: `yotpo`
- Supported service: `zoho_books`
- Supported service: `zoho_desk`
- Supported service: `zoom`

New connector config fields supported:
- Added field `fivetran_connector.config.client_host` for services: `ceridian_dayforce`.
- Added field `fivetran_connector.config.report_configs` for services: `yahoo_dsp`.
- Added field `fivetran_connector.config.personal_api_token` for services: `monday`.
- Added field `fivetran_connector.config.host_url` for services: `adobe_commerce`.
- Added field `fivetran_connector.config.organization_name` for services: `confluence`.
- Added field `fivetran_connector.config.agent_config_method` for services: `hana_sap_hva_s4_netweaver`, `hana_sap_hva_b1`, `hana_sap_hva_ecc`, `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`.
- Added field `fivetran_connector.config.system_id` for services: `hana_sap_hva_b1`, `hana_sap_hva_ecc`, `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`.
- Added field `fivetran_connector.config.tenant_url` for services: `ivanti`, `reltio`.
- Added field `fivetran_connector.config.collection_address` for services: `rarible`.
- Added field `fivetran_connector.config.selected_event_types` for services: `salesforce_marketing_cloud`.
- Added field `fivetran_connector.config.customer_api_key` for services: `ukg_pro`.
- Added field `fivetran_connector.config.companies` for services: `business_central`.
- Added field `fivetran_connector.config.partner_user_secret` for services: `expensify`.
- Added field `fivetran_connector.config.sap_source_schema` for services: `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`, `hana_sap_hva_b1`, `hana_sap_hva_ecc`.
- Added field `fivetran_connector.config.hana_mode` for services: `hana_sap_hva_b1`, `hana_sap_hva_ecc`, `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`.
- Added field `fivetran_connector.config.token` for services: `mode`, `oracle_moat_analytics`.
- Added field `fivetran_connector.config.email_id` for services: `ordway`.
- Added field `fivetran_connector.config.sync_mode_advertiser` for services: `yahoo_dsp`.
- Added field `fivetran_connector.config.list_of_company_ids` for services: `cj_commission_detail`.
- Added field `fivetran_connector.config.dsv_service_auth` for services: `dsv`.
- Added field `fivetran_connector.config.workspace` for services: `mode`.
- Added field `fivetran_connector.config.project_access_token` for services: `rollbar`.
- Added field `fivetran_connector.config.x_user_email` for services: `workday_strategic_sourcing`.
- Added field `fivetran_connector.config.seats` for services: `yahoo_dsp`.
- Added field `fivetran_connector.config.agreement_grant_token` for services: `economic`.
- Added field `fivetran_connector.config.odbc_sys_ini_path` for services: `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`, `hana_sap_hva_b1`, `hana_sap_hva_ecc`.
- Added field `fivetran_connector.config.company` for services: `ordway`.
- Added field `fivetran_connector.config.facility_codes` for services: `unicommerce`.
- Added field `fivetran_connector.config.workplace_id` for services: `moloco`.
- Added field `fivetran_connector.config.custom_reports.add_metric_variants` for services: `snapchat_ads`.
- Added field `fivetran_connector.config.app_secret_token` for services: `economic`.
- Added field `fivetran_connector.config.account_access_token` for services: `rollbar`.
- Added field `fivetran_connector.config.hostname` for services: `ukg_pro`.
- Added field `fivetran_connector.config.sync_mode_seat` for services: `yahoo_dsp`.
- Added field `fivetran_connector.config.dsv_subscription_key` for services: `dsv`.
- Added field `fivetran_connector.config.log_on_group` for services: `hana_sap_hva_b1`, `hana_sap_hva_ecc`, `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`.
- Added field `fivetran_connector.config.snc_name` for services: `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`, `hana_sap_hva_b1`, `hana_sap_hva_ecc`.
- Added field `fivetran_connector.config.brand_id` for services: `oracle_moat_analytics`.
- Added field `fivetran_connector.config.report_list` for services: `spotify_ads`.
- Added field `fivetran_connector.config.environment_name` for services: `business_central`.
- Added field `fivetran_connector.config.account_type` for services: `freightview`.
- Added field `fivetran_connector.config.snc_library_path` for services: `hana_sap_hva_ecc`, `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`, `hana_sap_hva_b1`.
- Added field `fivetran_connector.config.service_name` for services: `hana_sap_hva_b1`, `hana_sap_hva_ecc`, `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`.
- Added field `fivetran_connector.config.ecommerce_stores` for services: `mailchimp`.
- Added field `fivetran_connector.config.audience` for services: `auth0`.
- Added field `fivetran_connector.config.target_host` for services: `d2l_brightspace`.
- Added field `fivetran_connector.config.account_sid` for services: `fone_dynamics`.
- Added field `fivetran_connector.config.snc_partner_name` for services: `hana_sap_hva_ecc`, `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`, `hana_sap_hva_b1`.
- Added field `fivetran_connector.config.backint_configuration_path` for services: `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`, `hana_sap_hva_b1`, `hana_sap_hva_ecc`.
- Added field `fivetran_connector.config.advertisers_with_seat` for services: `yahoo_dsp`.
- Added field `fivetran_connector.config.x_api_key` for services: `workday_strategic_sourcing`.
- Added field `fivetran_connector.config.application_id` for services: `algolia`.
- Added field `fivetran_connector.config.partner_user_id` for services: `expensify`.
- Added field `fivetran_connector.config.backint_executable_path` for services: `hana_sap_hva_b1`, `hana_sap_hva_ecc`, `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`.
- Added field `fivetran_connector.config.odbc_driver_manager_library_path` for services: `hana_sap_hva_s4_netweaver`, `hana_sap_hva_b1`, `hana_sap_hva_ecc`, `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`.
- Added field `fivetran_connector.config.user_token` for services: `konnect_insights`.
- Added field `fivetran_connector.config.report_keys` for services: `rakutenadvertising`.
- Added field `fivetran_connector.config.client_namespace` for services: `ceridian_dayforce`.
- Added field `fivetran_connector.config.blockchain` for services: `rarible`.
- Added field `fivetran_connector.config.x_user_token` for services: `workday_strategic_sourcing`.
- Added field `fivetran_connector.config.host_name` for services: `coassemble`.
- Added field `fivetran_connector.config.hana_backup_password` for services: `hana_sap_hva_ecc`, `hana_sap_hva_ecc_netweaver`, `hana_sap_hva_s4`, `hana_sap_hva_s4_netweaver`, `hana_sap_hva_b1`.
- Added field `fivetran_connector.config.account_token` for services: `konnect_insights`.

## [1.1.10](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.9...v1.1.10)

Hot-fix:
- Fixed import issues for connector, user and connector_schema_config resources.

## [1.1.9](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.8...v1.1.9)

Hot-fix:
- [Race conditions issue](https://github.com/fivetran/terraform-provider-fivetran/issues/241)

## [1.1.8](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.7...v1.1.8)

Hot-fixes for issues:
- [Concurrency issue](https://github.com/fivetran/terraform-provider-fivetran/issues/241)
- [Connector schedule sync_frequesncy issue](https://github.com/fivetran/terraform-provider-fivetran/issues/243)

## [1.1.7](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.6...v1.1.7)

- Resource `fivetran_external_logging` fields support:
    - Added field `fivetran_external_logging.config.project_id`.
- Deprecated data sources `fivetran_metadata_schemas`, `fivetran_metadata_tables`, `fivetran_metadata_columns` removed.

## [1.1.6](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.5...v1.1.6)

## Updated
- Internal refactoring (we are migrating from Terraform SDK to Terraform Plugin Framework)

## Added
- Following new `fivetran_connector.config` fields support:
    - Added field `fivetran_connector.config.business_accounts` for services: `reddit_ads`.
    - Added field `fivetran_connector.config.region_api_url` for services: `amazon_attribution`.
    - Added field `fivetran_connector.config.custom_payloads` for services: `google_cloud_function`, `aws_lambda`, `azure_function`.
    - Added field `fivetran_connector.config.reports.time_aggregation_granularity` for services: `google_analytics_4`.
    - Added field `fivetran_connector.config.refresh_token` for services: `ironsource`.
    - Added field `fivetran_connector.config.academy_id` for services: `workramp`.
    - Added field `fivetran_connector.config.api_environment` for services: `afterpay`.
    - Added field `fivetran_connector.config.region_token_url` for services: `amazon_attribution`.
    - Added field `fivetran_connector.config.region_auth_url` for services: `amazon_attribution`.
    - Added field `fivetran_connector.config.database_name` for services: `firebase`.
    - Added field `fivetran_connector.config.connection_name` for services: `appsflyer`.
    - Added field `fivetran_connector.config.server` for services: `castor_edc`.
    - Added field `fivetran_connector.config.auth_code` for services: `happyfox`.
- Connector `fivetran_connector.auth` fields support:
    - Added field `fivetran_connector.config.client_access.user_agent` for services: `google_ads`.
    - Added field `fivetran_connector.config.client_access.developer_token` for services: `google_ads`.

## [1.1.5](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.4...v1.1.5)

## Fixed 
- Issue with `fivetran_connector_schema_config` resource when column isn't excluded from schema.

## [1.1.4](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.3...v1.1.4)

## Added
- New datasource `fivetran_group_ssh_key` that provides public key from SSH key pair associated with the group.
- New datasource `fivetran_group_service_account` that provides Fivetran service account associated with the group.
- Connector `fivetran_connector.auth` fields support:
    - Added field `fivetran_connector.auth.custom_field_ids`.
    - Added field `fivetran_connector.auth.previous_refresh_token`.
    - Added field `fivetran_connector.auth.user_access_token`.
    - Added field `fivetran_connector.auth.consumer_secret`.
    - Added field `fivetran_connector.auth.consumer_key`.
    - Added field `fivetran_connector.auth.oauth_token`.
    - Added field `fivetran_connector.auth.oauth_token_secret`.
    - Added field `fivetran_connector.auth.role_arn`.
    - Added field `fivetran_connector.auth.aws_access_key`.
    - Added field `fivetran_connector.auth.aws_secret_key`.
    - Added field `fivetran_connector.auth.client_id`.
    - Added field `fivetran_connector.auth.key_id`.
    - Added field `fivetran_connector.auth.team_id`.
    - Added field `fivetran_connector.auth.client_secret`.
- Resource `fivetran_destination` updates:
    - Added field `fivetran_destination.config.workspace_name` for OneLake.
    - Added field `fivetran_destination.config.lakehouse_name` for OneLake.
- Connector services supported:
    - Supported service: `calabrio`
    - Supported service: `float`
    - Supported service: `globalmeet`
    - Supported service: `linear`
    - Supported service: `power_reviews_enterprise`
    - Supported service: `smartwaiver`
    - Supported service: `uppromote`
    - Supported service: `zenefits`

## [1.1.3](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.2...1.1.3)

## Fixed
- Issue `Invalid Provider Server Combination`.

## [1.1.2](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.1...v1.1.2)

## Fixed
- Issue with `fivetran_connector_schema_config` resource when table `sync_mode` field doesn't affect upstream. 

## Added
New connector service types supported:
- Supported service: `15five`
- Supported service: `appcues`
- Supported service: `attentive`
- Supported service: `awin`
- Supported service: `ballotready`
- Supported service: `brevo`
- Supported service: `buzzsprout`
- Supported service: `canny`
- Supported service: `care_quality_commission`
- Supported service: `circleci`
- Supported service: `customerio`
- Supported service: `dbt_cloud`
- Supported service: `deputy`
- Supported service: `flexport`
- Supported service: `forj_community`
- Supported service: `hopin`
- Supported service: `insightly`
- Supported service: `integral_ad_science`
- Supported service: `justcall`
- Supported service: `katana`
- Supported service: `launchdarkly`
- Supported service: `looker_source`
- Supported service: `loop`
- Supported service: `loopio`
- Supported service: `mention`
- Supported service: `mixmax`
- Supported service: `mountain`
- Supported service: `namely`
- Supported service: `navan`
- Supported service: `ometria`
- Supported service: `pagerduty`
- Supported service: `partnerstack_vendor`
- Supported service: `playvox`
- Supported service: `revel`
- Supported service: `rippling`
- Supported service: `security_journey`
- Supported service: `skilljar`
- Supported service: `smadex`
- Supported service: `stylight`
- Supported service: `toggl_track`
- Supported service: `trisolute`
- Supported service: `zoho_campaigns`

New connector config fields supported:
- Added field `fivetran_connector.config.limit_for_api_calls_to_external_activities_endpoint` for services: `pardot`.
- Added field `fivetran_connector.config.is_external_activities_endpoint_selected` for services: `pardot`.
- Added field `fivetran_connector.config.distributed_connector_cluster_size` for services: `cosmos`.
- Added field `fivetran_connector.config.custom_reports.granularity` for services: `snapchat_ads`.
- Added field `fivetran_connector.config.custom_reports.breakdown` for services: `snapchat_ads`.
- Added field `fivetran_connector.config.custom_reports.dimension` for services: `snapchat_ads`.
- Added field `fivetran_connector.config.custom_reports.sk_ad_metrics_fields` for services: `snapchat_ads`.
- Added field `fivetran_connector.config.custom_reports.breakout` for services: `snapchat_ads`.
- Added field `fivetran_connector.config.partner_code` for services: `care_quality_commission`.
- Added field `fivetran_connector.config.reports.filter_type` for services: `google_analytics_4`.
- Added field `fivetran_connector.config.app_key` for services: `loopio`.
- Added field `fivetran_connector.config.enable_data_extensions_syncing` for services: `salesforce_marketing_cloud`.
- Added field `fivetran_connector.config.reports.rollback_window` for services: `google_analytics_4`.
- Added field `fivetran_connector.config.api_utilization_percentage` for services: `kustomer`.
- Added field `fivetran_connector.config.api_key:api_secret` for services: `revel`.
- Added field `fivetran_connector.config.custom_reports.base_metrics_fields` for services: `snapchat_ads`.
- Added field `fivetran_connector.config.tenant` for services: `workday_hcm`.
- Added field `fivetran_connector.config.enable_distributed_connector_mode` for services: `cosmos`.


## [1.1.1](https://github.com/fivetran/terraform-provider-fivetran/compare/v1.1.0...v1.1.1)

## Added
- Supports setting connector synchronization every minute. Added possible value 1 for the field `fivetran_connector_schedule.sync_frequency`. 
- New resource `fivetran_connector_certificates` that allows to manage the list of certificates approved for connector.
- New resource `fivetran_connector_fingerprints` that allows to manage the list of SSH fingerprints approved for connector.
- New resource `fivetran_destination_certificates` that allows to manage the list of certificates approved for destination.
- New resource `fivetran_destination_fingerprints` that allows to manage the list of SSH fingerprints approved for destination.
- New data source `fivetran_connector_certificates` that allows to retrieve the list of certificates approved for connector.
- New data source `fivetran_connector_fingerprints` that allows to retrieve the list of SSH fingerprints approved for connector.
- New data source `fivetran_destination_certificates` that allows to retrieve the list of certificates approved for destination.
- New data source `fivetran_destination_fingerprints` that allows to retrieve the list of SSH fingerprints approved for destination.
- New resource `fivetran_team` that allows to manage Teams.
- New resource `fivetran_team_connector_membership` that allows to manage Team Management Connector memberships.
- New resource `fivetran_team_group_membership` that allows to manage Team Management Group memberships.
- New resource `fivetran_team_user_membership` that allows to manage Team Management User memberships.
- New data source `fivetran_team` that allows to retrieve details of the existing team for a given identifier.
- New data source `fivetran_teams` that allows to retrieve the list of existing teams available for the current account.
- New data source `fivetran_team_connector_memberships` that allows to retrieve the list of existing connector memberships available for team.
- New data source `fivetran_team_group_memberships` that allows to retrieve the list of existing group memberships available for team.
- New data source `fivetran_team_user_memberships` that allows to retrieve the list of existing user memberships available for team.
- Resource `fivetran_connector` updates:
    - Added field `fivetran_connector.config.company_request_token` for services: `concur`.
    - Added field `fivetran_connector.config.company_uuid` for services: `concur`.
    - Added field `fivetran_connector.config.client` for services: `sap_hana_db`.
    - Added field `fivetran_connector.config.sysnr` for services: `sap_hana_db`.
    - Added field `fivetran_connector.config.pat_name` for services: `tableau_source`.
    - Added field `fivetran_connector.config.server_address` for services: `tableau_source`.
    - Added field `fivetran_connector.config.pat_secret` for services: `tableau_source`.

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

- New resource `fivetran_external_logging` that allows to manage Log Services.
- New data source `fivetran_metadata_schemas` that allows to retrieve schema-level metadata for an existing connector within your Fivetran account.
- New data source `fivetran_metadata_tables` that allows to retrieve table-level metadata for an existing connector within your Fivetran account.
- New data source `fivetran_metadata_columns` that allows to retrieve column-level metadata for an existing connector within your Fivetran account.
- New data source `fivetran_roles` that allows to retrieve a list of all predefined and custom roles within your Fivetran account

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
