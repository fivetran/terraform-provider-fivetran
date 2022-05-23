# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.4.2...HEAD)

## [0.4.3](https://github.com/fivetran/terraform-provider-fivetran/compare/v0.4.2...v0.4.3) - 2022-06-01

## Fixed
- Issue with `update_method` connector config field (now it can be effectively updated)
- Issue with `connection_type` field isn't marked as readonlu no more

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
