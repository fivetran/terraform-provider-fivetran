---
page_title: "Resource: fivetran_account_external_logging_v2"
---

# Resource: fivetran_account_external_logging_v2

~> **Public Preview:** This resource is in public preview. Features and behavior may change based on feedback.

Manages the account-level external logging service. An account can have at most one account-level log service. For log services attached to a destination, use `fivetran_external_logging`.

`config` is a dynamic, service-specific object. When the Fivetran log service metadata API (`/metadata/log-service-types/{service}`) describes the log service, the accepted fields, types and required/immutable rules are validated at plan time. Otherwise the configuration is passed to the API as is, and the API validates it.

## Example Usage

```hcl
resource "fivetran_account_external_logging_v2" "account_log" {
    service         = "datadog_log"
    enabled         = true
    run_setup_tests = true

    config = {
        api_key  = "api_key"
        hostname = "hostname"
        region   = "US1"
    }
}
```

## Schema

### Required

- `config` (Dynamic) Service-specific log service configuration. The accepted fields depend on `service`.
- `service` (String) The name of the log service type within the Fivetran system. Supported log services: `azure_monitor_log`, `cloudwatch`, `datadog_log`, `dynatrace`, `grafana_loki`, `new_relic_log`, `splunkLog`, `stackdriver`. Changing this forces the resource to be replaced.

### Optional

- `enabled` (Boolean) The boolean value specifying whether the log service is enabled. Defaults to `true`.
- `run_setup_tests` (Boolean) Specifies whether the setup tests should be run after the log service is created or updated. Defaults to `false`. This is a plan-only attribute and is not read back from the API.

### Read-Only

- `id` (String) The unique identifier for the account-level log service within the Fivetran system.

## Notes

- Values the API masks (for example secrets) are kept from your configuration and are never read back.
- Setting `run_setup_tests` from `false` to `true` without any other change runs the setup tests through the dedicated setup-tests endpoint. Failed tests are reported as warnings.

## Import

The account has a single account-level log service, and its ID is the account ID. Use your account ID as the import ID.

1. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_account_external_logging_v2" "my_imported_log" {

}
```

2. Run the `terraform import` command:

```
terraform import fivetran_account_external_logging_v2.my_imported_log {account ID}
```

3. Use `terraform state show 'fivetran_account_external_logging_v2.my_imported_log'` to get the values, and paste them into your `.tf` configuration.

-> The API masks secrets, so imported state does not contain them. Add the secret values to `config` yourself.
