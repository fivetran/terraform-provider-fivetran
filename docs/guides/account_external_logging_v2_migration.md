----
page_title: "Migrating from fivetran_account_external_logging_v2"
subcategory: "Migration Guides"
---

# Migrating from `fivetran_account_external_logging_v2`

`fivetran_account_external_logging_v2` is deprecated. Use [`fivetran_external_logging_v2`](https://registry.terraform.io/providers/fivetran/fivetran/latest/docs/resources/external_logging_v2) without `group_id` instead: it manages the same account-level log service.

There are two ways to migrate. Neither recreates the log service:

- [Move the resource](#option-1-move-the-resource) with a `moved` block. This needs Terraform 1.8 or later.
- [Re-import the log service](#option-2-re-import-the-log-service) into the new resource. This works with any Terraform version.

The examples below migrate this resource:

```hcl
resource "fivetran_account_external_logging_v2" "account_log" {
    service = "datadog_log"

    config = {
        api_key  = var.datadog_api_key
        hostname = "hostname"
        region   = "US1"
    }
}
```

## Option 1: Move the resource

With Terraform 1.8 or later, a `moved` block moves the resource in state. The log service isn't recreated or changed.

1. Rename the resource type, keep the attributes, and add a `moved` block:

```hcl
resource "fivetran_external_logging_v2" "account_log" {
    service = "datadog_log"

    config = {
        api_key  = var.datadog_api_key
        hostname = "hostname"
        region   = "US1"
    }
}

moved {
    from = fivetran_account_external_logging_v2.account_log
    to   = fivetran_external_logging_v2.account_log
}
```

2. Run `terraform plan`. It should show the move and no changes. Then run `terraform apply`.

3. Delete the `moved` block, and update references from `fivetran_account_external_logging_v2.account_log` to `fivetran_external_logging_v2.account_log`.

## Option 2: Re-import the log service

You can also import the log service into a new `fivetran_external_logging_v2` resource and remove the old resource from state without deleting the log service. The import ID is `account`, or the account ID.

### With `import` and `removed` blocks (Terraform 1.7 or later)

1. Replace the `fivetran_account_external_logging_v2` block with a `fivetran_external_logging_v2` block, and add an `import` block and a `removed` block:

```hcl
resource "fivetran_external_logging_v2" "account_log" {
    service = "datadog_log"

    config = {
        api_key  = var.datadog_api_key
        hostname = "hostname"
        region   = "US1"
    }
}

import {
    to = fivetran_external_logging_v2.account_log
    id = "account"
}

removed {
    from = fivetran_account_external_logging_v2.account_log

    lifecycle {
        destroy = false
    }
}
```

~> Without the `removed` block, or without `destroy = false` in it, Terraform deletes the log service.

2. Run `terraform plan`, review it, and run `terraform apply`.

3. Delete the `import` and `removed` blocks, and update references from `fivetran_account_external_logging_v2.account_log` to `fivetran_external_logging_v2.account_log`. Run `terraform plan` again. It should show no changes.

### With CLI commands (older Terraform versions)

1. Remove the old resource from state. This doesn't delete the log service:

```shell
terraform state rm fivetran_account_external_logging_v2.account_log
```

2. Replace the `fivetran_account_external_logging_v2` block with a `fivetran_external_logging_v2` block with the same attributes, and import the log service:

```shell
terraform import fivetran_external_logging_v2.account_log account
```

3. Update references to the resource, then run `terraform plan`, review it, and run `terraform apply`.

### What to expect after re-importing

- Secret fields (`format: password` in log service metadata) aren't imported into state. When they're set in `config`, the first apply sends them in an in-place update. The log service isn't recreated.
- When metadata marks a secret as immutable, the plan warns that Terraform can't compare it with the imported state. The value isn't sent and doesn't recreate the log service.
- Imported `config` fields that the configuration leaves out aren't shown as removed and aren't cleared. If the same plan changes other `config` fields, the plan shows them as removed and warns about them, and the apply stops tracking them without clearing them.

## After migrating

`config` in `fivetran_external_logging_v2` isn't sensitive, so plan output shows its fields. To hide secrets, pass them from sensitive variables, or move them to the write-only `config_secrets` attribute. See [Keeping secrets out of state](https://registry.terraform.io/providers/fivetran/fivetran/latest/docs/resources/external_logging_v2#keeping-secrets-out-of-state).
