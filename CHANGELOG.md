# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.6.17...HEAD)

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
