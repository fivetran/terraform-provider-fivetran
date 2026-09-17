<p align="center">
  <a href="https://www.fivetran.com/">
    <img src="https://cdn.prod.website-files.com/6130fa1501794ed4d11867ba/63d9599008ad50523f8ce26a_logo.svg" alt="Fivetran">
  </a>
</p>

<p align="center">
  Manage your Fivetran resources as infrastructure as code.
</p>

<p align="center">
  <a href="https://registry.terraform.io/providers/fivetran/fivetran/latest" target="_blank"><img src="https://img.shields.io/badge/Terraform%20Registry-fivetran%2Ffivetran-844FBA?logo=terraform" alt="Terraform Registry"></a>
  <a href="https://github.com/fivetran/terraform-provider-fivetran/releases/latest" target="_blank"><img src="https://img.shields.io/github/v/release/fivetran/terraform-provider-fivetran" alt="Latest Release"></a>
  <a href="https://github.com/fivetran/terraform-provider-fivetran/blob/main/LICENSE" target="_blank"><img src="https://img.shields.io/github/license/fivetran/terraform-provider-fivetran" alt="License"></a>
</p>

# Terraform Provider for Fivetran

The `terraform-provider-fivetran` is the official Terraform provider for managing [Fivetran](https://fivetran.com/) resources as infrastructure as code. Use it to provision and manage destinations, connections, users, teams, roles, and other Fivetran resources through version-controlled Terraform configurations.

We strongly encourage you to get acquainted with the Fivetran REST API [documentation](https://fivetran.com/docs/rest-api) before using our Terraform provider.

See our [CHANGELOG](CHANGELOG.md) for information about the latest bug fixes, updates, and features added to the provider. 

## Quickstart

### Prerequisites

- [Terraform CLI](https://developer.hashicorp.com/terraform/install) v1.0 or later
- A Fivetran account
- A [Fivetran API key and API secret](https://fivetran.com/docs/developer-resources/terraform/getting-started#authenticate) with sufficient permissions for the resources you want to manage

### Configure the provider

Create a Terraform configuration:

```hcl
terraform {
  required_version = ">= 1.0"

  required_providers {
    fivetran = {
      source  = "fivetran/fivetran"
      version = "~> 1.9"
    }
  }
}

provider "fivetran" {}

resource "fivetran_group" "example" {
  name = "terraform-example"
}
```

Set your credentials as environment variables:

```bash
export FIVETRAN_APIKEY="your-api-key"
export FIVETRAN_APISECRET="your-api-secret"
```

Initialize Terraform and preview the configuration:

```bash
terraform init
terraform plan
```

Never commit API credentials to `.tf` files or version control. The provider requires the API key and API secret as separate values; do not use a Base64-encoded API key as either value. For authentication options and API-key types, see our [Getting Started Guide](https://fivetran.com/docs/developer-resources/terraform/getting-started#authenticate).

## Managing connections

Fivetran connection settings are service-specific. You can use the legacy [`fivetran_connector`](https://registry.terraform.io/providers/fivetran/fivetran/latest/docs/resources/connector) resource or the [Connection_v2 Configuration Page](https://fivetran.com/docs/developer-resources/terraform/terraform-configuration-connections) to generate the required HCL for a connection service.

The `fivetran_connection_v2` resource uses metadata-driven `config` and `auth` objects and validates their fields during planning. Keep these behaviors in mind:

- `destination_schema_names` is required.
- The accepted `config` and `auth` fields depend on the selected service.
- Sensitive fields are not necessarily placed in `auth`; follow the generated configuration for the service.
- New v2 connections are created paused. Manage their state separately with `fivetran_connection_v2_pause_state`.
- `fivetran_connection_v2` and `fivetran_connection_v2_pause_state` are in **alpha** and are subject to change at any moment.

See the [`fivetran_connection_v2` reference](https://registry.terraform.io/providers/fivetran/fivetran/latest/docs/resources/connection_v2) for the full resource schema and current alpha limitations.

## Important considerations

- Provider capabilities depend on the corresponding Fivetran REST API behavior and the selected connection service.
- Some authorization flows, including interactive OAuth, may require completion outside Terraform.
- Sensitive values may be stored in Terraform state. Use an encrypted remote backend with appropriately restricted access.
- Avoid managing the same setting through overlapping Terraform resources or through both Terraform and the Fivetran dashboard. Out-of-band changes create drift and may be reverted by a later `terraform apply`.

For release-specific changes and known fixes, review the [changelog](CHANGELOG.md) before upgrading.

## Known issues

- Version 1.2.5 was broken, please use version 1.2.6.
- Some lists may have to change to sets for better usability.
- REST API's Response and Request payloads may differ for some connectors, and in some cases data transformations may occur. Setting up and managing some connectors may not be possible due to that limitation. As a workaround, we may deliver individual connectors data sources and resources at the Terraform Provider level instead of using the current REST API approach of a single endpoint to manage all connectors.
- If you receive messages of the following type when planning/applying the fivetran_connector resource:
`unexpected new value: .config.field: was cty.StringVal("value"), but now null`
Or
`unexpected new value: .config.field: inconsistent values for sensitive attribute`
Check that the field causing the problem is actually applicable to the service specified in the resource
- **For SAP ERP for HANA connectors, configuring specific table selections via `fivetran_connector_schema_config` in Terraform may result in a `Table with name [TABLE_NAME] not found in source schema [SCHEMA_NAME]` error.** This occurs because the SAP ERP for HANA connector initially starts with no schema preloaded, and a schema discovery process must complete successfully before tables can be configured. Even setting `validation_level = "NONE"` in Terraform does not resolve this underlying backend requirement. As a workaround, it is recommended to **create the `fivetran_connector` resource without the `fivetran_connector_schema_config` resource first.** Allow the connector to be created and initiate its schema discovery. **After the connector is established and the schema has been discovered (which may require a manual refresh in the Fivetran UI or an API call), then manage table selections and schema configurations directly within the Fivetran UI.**

## Support

For feedback, feature requests, or bug reports, contact Fivetran through the [Support Portal](https://support.fivetran.com/). 

## Contributions

External contributions are not currently accepted. See [CONTRIBUTING.md](CONTRIBUTING.md) for the current contribution policy.

## License

This project is licensed under the [Apache License 2.0](LICENSE).
