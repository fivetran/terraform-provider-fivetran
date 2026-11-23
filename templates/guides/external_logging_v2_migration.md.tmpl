----
page_title: "Migrating from fivetran_external_logging to fivetran_external_logging_v2"
subcategory: "Migration Guides"
---

# Migrating from `fivetran_external_logging` to `fivetran_external_logging_v2`

[`fivetran_external_logging_v2`](https://registry.terraform.io/providers/fivetran/fivetran/latest/docs/resources/external_logging_v2) manages the same API object as `fivetran_external_logging`, so you can migrate without recreating the log service. The ID stays the same: it's the ID of the destination (group).

There are two ways to migrate. Neither recreates the log service:

- [With `import` and `removed` blocks](#with-import-and-removed-blocks). This needs Terraform 1.7 or later.
- [With CLI commands](#with-cli-commands), for older Terraform versions.

## With `import` and `removed` blocks

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

## With CLI commands

With Terraform versions older than 1.7, remove the old resource from state and import the log service into the new one.

1. Find the ID of the log service:

```
terraform state show fivetran_external_logging.my_log | grep ' id '
```

2. Remove the old resource from state. This doesn't delete the log service:

```shell
terraform state rm fivetran_external_logging.my_log
```

3. Replace the `fivetran_external_logging` block with a `fivetran_external_logging_v2` block, as in step 2 above but without the `import` and `removed` blocks, and import the log service:

```shell
terraform import fivetran_external_logging_v2.my_log <destination ID>
```

4. Run `terraform plan`. Expect an in-place update for the secrets, as described in step 3 above, and no replacement. Then run `terraform apply`.

5. Update references from `fivetran_external_logging.my_log.id` to `fivetran_external_logging_v2.my_log.id`.

## Notes


- Unlike in `fivetran_external_logging`, only the fields set in `config` are tracked.
- `run_setup_tests = true` in the new block runs the setup tests during the migration apply.
- When metadata marks a secret as immutable, the plan warns that Terraform can't compare it with the imported state. The value isn't sent and doesn't recreate the log service.
