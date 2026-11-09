---
page_title: "Resource: fivetran_external_logging_v2"
---

# Resource: fivetran_external_logging_v2

~> **Alpha:** This resource is in alpha and its schema and behavior may change without notice. Not recommended for production use.

Manages the external logging service of a destination (group). A destination can have at most one log service. For the account-level log service, use `fivetran_account_external_logging_v2`.

`config` is a dynamic, service-specific object. When the Fivetran log service metadata API (`/metadata/log-service-types/{service}`) describes the log service, the accepted fields, types, enum values and required/readonly/immutable rules are validated at plan time. Otherwise the configuration is passed to the API as is, and the API validates it.

## Example Usage

```hcl
variable "datadog_api_key" {
    type      = string
    sensitive = true
}

resource "fivetran_external_logging_v2" "log" {
    group_id        = fivetran_destination_v2.destination.id
    service         = "datadog_log"
    enabled         = true
    run_setup_tests = true

    config = {
        api_key  = var.datadog_api_key
        hostname = "hostname"
        region   = "US1"
    }
}
```

`config` is not sensitive, so its fields are shown in plan output. Pass secrets from sensitive variables, as above, so Terraform hides them, or set them in `config_secrets`.

### Keeping secrets out of state

Secrets set in `config` are stored in Terraform state. To keep them out of state, or to pass ephemeral values (for example from an `ephemeral` secret manager resource), set them in the write-only `config_secrets` attribute instead (Terraform 1.11 or later):

```hcl
ephemeral "google_secret_manager_secret_version" "datadog_api_key" {
    secret = "datadog-api-key"
}

resource "fivetran_external_logging_v2" "log" {
    group_id = fivetran_destination_v2.destination.id
    service  = "datadog_log"

    config = {
        hostname = "hostname"
        region   = "US1"
    }

    config_secrets = {
        api_key = ephemeral.google_secret_manager_secret_version.datadog_api_key.secret_data
    }
    config_secrets_version = 1
}
```

- When the log service metadata describes the service, `config_secrets` accepts only fields with `format: password`. A field can't be set in both `config` and `config_secrets`.
- Terraform doesn't store `config_secrets`, so it can't detect when a secret changes. The values are sent on create and whenever `config_secrets_version` changes: increment it after rotating a secret.
- To move a secret from `config` to `config_secrets`, remove it from `config`, add it to `config_secrets` and increment `config_secrets_version`.

## Schema

### Required

- `config` (Dynamic) Service-specific log service configuration. The accepted fields depend on `service`. `config` is not sensitive: set secrets in `config_secrets`, or pass them from sensitive variables.
- `group_id` (String) The unique identifier for the group (destination) the log service belongs to. Changing this forces the resource to be replaced.
- `service` (String) The name of the log service type within the Fivetran system. Supported log services: `azure_monitor_log`, `cloudwatch`, `datadog_log`, `dynatrace`, `grafana_loki`, `new_relic_log`, `splunkLog`, `stackdriver`. Changing this forces the resource to be replaced.

### Optional

- `config_secrets` (Dynamic, Sensitive, [Write-only](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments)) Write-only secret config fields (fields with `format: password` in log service metadata, for example `api_key` or `token`). Accepts ephemeral values and is never stored in state. Values are sent on create and whenever `config_secrets_version` changes. Requires Terraform 1.11 or later.
- `config_secrets_version` (Number) Change this value to send the current `config_secrets` to Fivetran, for example after rotating a secret.
- `enabled` (Boolean) The boolean value specifying whether the log service is enabled. Defaults to `true`.
- `run_setup_tests` (Boolean) Specifies whether the setup tests should be run after the log service is created or updated. Defaults to `false`. This is a plan-only attribute and is not read back from the API.

### Read-Only

- `id` (String) The unique identifier for the log service within the Fivetran system. It matches `group_id`.

## Notes

- `config` is not sensitive, so its fields are shown in plan and apply output, except values that come from sensitive variables. Secrets set in `config` are stored in state: protect the state file, or use `config_secrets` to keep secrets out of it.
- Values the API masks (for example secrets) are kept from your configuration and are never read back.
- Changing `service`, `group_id` or an immutable `config` field replaces the log service. Terraform deletes the existing service before it creates the new one, so log delivery for the destination stops until the new service exists.
- With `run_setup_tests = true`, the setup tests run after the log service is created, with every update, and when the attribute changes from `false` to `true` without any other change. Failed tests are reported as warnings.

## Migrating from `fivetran_external_logging`

`fivetran_external_logging_v2` manages the same API object as `fivetran_external_logging`, so you can migrate without recreating the log service. The ID stays the same: it's the ID of the destination (group). The migration uses `import` and `removed` blocks, which need Terraform 1.7 or later.

1. Find the ID of each log service:

```
terraform state show fivetran_external_logging.my_log | grep ' id '
```

2. Replace the `fivetran_external_logging` block with a `fivetran_external_logging_v2` block, and add an `import` block and a `removed` block. The `config` block becomes a `config` attribute with the same field names: only the syntax changes.

Before:

```hcl
resource "fivetran_external_logging" "my_log" {
    group_id = fivetran_destination.destination.id
    service  = "datadog_log"
    enabled  = true

    config {
        api_key = var.datadog_api_key
        region  = "US1"
    }
}
```

After:

```hcl
resource "fivetran_external_logging_v2" "my_log" {
    group_id = fivetran_destination.destination.id
    service  = "datadog_log"
    enabled  = true

    config = {
        api_key = var.datadog_api_key
        region  = "US1"
    }
}

import {
    to = fivetran_external_logging_v2.my_log
    id = "<destination ID>"
}

removed {
    from = fivetran_external_logging.my_log
    lifecycle {
        destroy = false
    }
}
```

~> Without the `removed` block, or without `destroy = false` in it, Terraform deletes the log service.

To keep secrets out of state, move them to `config_secrets` instead and set `config_secrets_version = 1`, so that the migration sends them once.

3. Run `terraform plan`. Expect `1 to import` and `0 to destroy`. The log service is also planned for an in-place update, because the API doesn't return secrets and the imported state doesn't hold them. If the plan shows a replacement, stop and check that `group_id`, `service` and the non-secret `config` fields match the existing log service. Then run `terraform apply`.

4. Delete the `import` and `removed` blocks and run `terraform plan` again. It should show no changes.

5. Update references from `fivetran_external_logging.my_log.id` to `fivetran_external_logging_v2.my_log.id`.

To migrate several log services, use `for_each` in the `import` and `removed` blocks, or generate the blocks from state:

```
terraform show -json | jq -r '
  .values.root_module.resources[]
  | select(.type == "fivetran_external_logging")
  | "import { to = fivetran_external_logging_v2.\(.name)\n  id = \"\(.values.id)\" }\n" +
    "removed { from = fivetran_external_logging.\(.name)\n  lifecycle { destroy = false } }\n"
' > migration.tf
```

This covers resources in the root module only. Review `migration.tf` before you apply it.

Notes:

- Unlike in `fivetran_external_logging`, only the fields set in `config` are tracked.
- `run_setup_tests = true` in the new block runs the setup tests during the migration apply.
- When metadata marks a secret as immutable, the plan warns that Terraform can't compare it with the imported state. The value isn't sent and doesn't recreate the log service.

## Import

The log service ID is the ID of the destination (group) it belongs to. Use the destination ID as the import ID.

1. Define an empty resource in your `.tf` configuration:

```hcl
resource "fivetran_external_logging_v2" "my_imported_log" {

}
```

2. Run the `terraform import` command:

```
terraform import fivetran_external_logging_v2.my_imported_log {destination ID}
```

3. Use `terraform state show 'fivetran_external_logging_v2.my_imported_log'` to get the values, and paste them into your `.tf` configuration.

-> The API masks secrets, so imported state does not contain them. Add the secret values to `config` or `config_secrets` yourself. Readonly fields are not imported, because `config` doesn't accept them.

### Import with an `import` block

With Terraform 1.5 or later, you can also import the log service with an `import` block. The import then runs as part of `terraform plan` and `terraform apply`, so you can review it before it changes state.

1. Add an `import` block to your `.tf` configuration:

```hcl
import {
    to = fivetran_external_logging_v2.my_imported_log
    id = "{destination ID}"
}
```

2. Add the matching `fivetran_external_logging_v2` resource block yourself, or let Terraform generate it:

```
terraform plan -generate-config-out=generated.tf
```

3. Review the resource block. Add the secrets, which the generated block doesn't contain, to `config` or `config_secrets`. Then run `terraform plan` and check that it shows `1 to import` and `0 to destroy`.

4. Run `terraform apply`, then delete the `import` block.
