# Terraform Provider for Fivetran

`terraform-provider-fivetran` is the official Terraform Provider for Fivetran. 

Checkout our [CHANGELOG](CHANGELOG.md) for information about the latest bug fixes, updates, and features added to the SDK. 

Make sure you read the Fivetran REST API [documentation](https://fivetran.com/docs/rest-api) before using the Provider.

**NOTE**: `terraform-provider-fivetran` is in [BETA](https://en.wikipedia.org/wiki/Software_release_life_cycle#Beta) development stage (since v0.4.0). Future versions may introduce minor changes and bug-fixes. 

## Known issues

- Some lists may have to change to sets for better usability.
- REST API's Response and Request payloads may differ for some connectors, and in some cases data transformations may occur. Setting up and managing some connectors may not be possible due to that limitation. As a workaround, we may deliver individual connectors data sources and resources at the Terraform Provider level instead of using the current REST API approach of a single endpoint to manage all connectors.

## Support

Please get in touch with us through our [Support Portal](https://support.fivetran.com/) if you 
have any comments, suggestions, support requests, or bug reports.  
