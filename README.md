# Terraform Provider for Fivetran

`terraform-provider-fivetran` is the official Terraform Provider for Fivetran. 

Checkout our [CHANGELOG](CHANGELOG.md) for information about the latest bug fixes, updates, and features added to the SDK. 

Make sure you read the Fivetran REST API [documentation](https://fivetran.com/docs/rest-api) before using the Provider.

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

Please get in touch with us through our [Support Portal](https://support.fivetran.com/) if you 
have any comments, suggestions, support requests, or bug reports.  
