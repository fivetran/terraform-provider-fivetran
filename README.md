# Terraform Provider for Fivetran

`terraform-provider-fivetran` is the official Terraform Provider for Fivetran. 

Checkout our [CHANGELOG](CHANGELOG.md) for information about the latest bug fixes, updates, and features added to the SDK. 

Make sure you read the Fivetran REST API [documentation](https://fivetran.com/docs/rest-api) before using the Provider.

**NOTE**: `terraform-provider-fivetran` is in [BETA](https://en.wikipedia.org/wiki/Software_release_life_cycle#Beta) development stage. Future versions may introduce minor changes and bug fixes. 

## Known issues

- Version 1.2.5 was broken, please use version 1.2.6.
- Some lists may have to change to sets for better usability.
- REST API's Response and Request payloads may differ for some connectors, and in some cases data transformations may occur. Setting up and managing some connectors may not be possible due to that limitation. As a workaround, we may deliver individual connectors data sources and resources at the Terraform Provider level instead of using the current REST API approach of a single endpoint to manage all connectors.
- If you receive messages of the following type when planning/applying the fivetran_connector resource:
`unexpected new value: .config.field: was cty.StringVal("value"), but now null`
Or
`unexpected new value: .config.field: inconsistent values for sensitive attribute`
Check that the field causing the problem is actually applicable to the service specified in the resource

## Support

Please get in touch with us through our [Support Portal](https://support.fivetran.com/) if you 
have any comments, suggestions, support requests, or bug reports.  
