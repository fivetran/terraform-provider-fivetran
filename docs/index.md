---
page_title: "Fivetran Provider"
description: |-
    Terraform Provider for Fivetran.
---

# Fivetran Provider

This is the official Terraform Provider for [Fivetran](https://fivetran.com). 

Make sure you read the Fivetran REST API [documentation](https://fivetran.com/docs/rest-api) before using the Provider.

**NOTE**: The Fivetran Provider is still in [ALPHA](https://en.wikipedia.org/wiki/Software_release_life_cycle#Alpha) development stage. Future versions may introduce breaking changes. 

## Example Usage

```hcl
provider "fivetran" {
}
```

## Known issues

- Some lists may have to change to sets for better usability.
- REST API's Response and Request payloads may differ for some connectors, and in some cases data transformations may occur. Setting up and managing some connectors may not be possible due to that limitation. As a workaround, we may deliver individual connectors data sources and resources at the Terraform Provider level instead of using the current REST API approach of a single endpoint to manage all connectors.

## Support

Please get in touch with us through our [Support Portal](https://support.fivetran.com/) if you 
have any comments, suggestions, support requests, or bug reports.  

## Schema

### Required

- `api_key` - can be set by the environment variable `FIVETRAN_APIKEY`
- `api_secret` - can be set by the environment variable `FIVETRAN_APISECRET`
