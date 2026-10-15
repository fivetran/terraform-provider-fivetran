---
page_title: "Resource: fivetran_destination_v2"
---

# Resource: fivetran_destination_v2

~> **Public Preview:** This resource is in public preview. Features and behavior may change based on feedback.

Manages a Fivetran destination. Unlike `fivetran_destination`, `config` is provided as a dynamic, service-specific object: its accepted fields, types, enum values, and readonly/immutable rules are resolved from destination metadata (`GET /v1/metadata/destination-types/{service}`) at plan time rather than from a fixed schema. Field names inside `config` are the same as in the Fivetran REST API.

Plan-time checks:
- Unknown `config` fields, wrong value types, and values outside an enum fail the plan.
- Readonly `config` fields (set by Fivetran, for example `public_key`) can't be configured.
- Changing an immutable `config` field plans a replacement of the destination.
- `region`, `networking_method`, `hybrid_deployment_agent_id`, and `private_link_id` produce a warning when the destination type doesn't list them as supported. The API applies account-level rules (for example allowed regions or the PrivateLink permission) at apply time.

Set the provider `skip_plan_time_validation = true` only if metadata can't be fetched and you need to bypass these checks temporarily.

## Example Usage

```hcl
variable "snowflake_password" {
    type      = string
    sensitive = true
}

resource "fivetran_destination_v2" "destination" {
    group_id         = fivetran_group.group.id
    service          = "snowflake"
    time_zone_offset = "0"
    region           = "GCP_US_EAST4"

    config = {
        host     = "your-account.snowflakecomputing.com"
        port     = 443
        database = "fivetran"
        auth     = "PASSWORD"
        user     = "fivetran_user"
        password = var.snowflake_password
    }
}
```

## Sensitive values

`config` is not marked sensitive as a whole, so plan output shows per-field changes. Pass secrets (for example `password`, `private_key`, `secret_key`) through variables declared with `sensitive = true` or wrap them in `sensitive()`; Terraform then redacts those values in plan output. A secret written as a literal string in `config` is shown in plan output.

Fields with `format: password` in destination metadata are never overwritten with the masked value returned by the API, so they don't produce a permanent diff.

## Import

```shell
terraform import fivetran_destination_v2.destination <destination_id>
```

The API doesn't return secrets, so after import add the sensitive `config` fields to the resource block. Readonly fields are not tracked in state after import.

## Migrating from `fivetran_destination`

The two resources have different state shapes, so `terraform state mv` doesn't apply. To move an existing destination without recreating it:

1. Back up state with `terraform state pull > backup.tfstate`.
2. Add a `fivetran_destination_v2` block. Write `config = { ... }` (attribute syntax) with the API field names; nested blocks become nested objects.
3. Remove the old resource from state: `terraform state rm fivetran_destination.destination`.
4. Import it into the new resource: `terraform import fivetran_destination_v2.destination <destination_id>`.
5. Remove the old `fivetran_destination` block and run `terraform plan` to check that no changes are planned apart from secrets you re-added.

## Schema

### Required

- `config` (Dynamic) Service-specific destination configuration. The accepted fields are defined by destination metadata at runtime.
- `group_id` (String) The unique identifier for the Group within the Fivetran system. Changing this forces the destination to be replaced.
- `service` (String) The destination service type (e.g., `snowflake`, `big_query`, `databricks`). Changing this forces the destination to be replaced.
- `time_zone_offset` (String) Determines the time zone for the Fivetran sync schedule. Possible values: -11 to +12.

### Optional

- `daylight_saving_time_enabled` (Boolean) Shift the sync schedule to account for daylight saving time (for `time_zone_offset` values that observe DST).
- `hybrid_deployment_agent_id` (String) The hybrid deployment agent ID that refers to the controller created for the group the destination belongs to.
- `networking_method` (String) The networking method for the destination. Possible values: `Directly`, `SshTunnel`, `ProxyAgent`, `PrivateLink`.
- `private_link_id` (String) The private link ID. Required when `networking_method` is `PrivateLink` and the destination supports self-service PrivateLink.
- `proxy_agent_id` (String) The ID of the proxy agent to use. Required when `networking_method` is `ProxyAgent`.
- `region` (String) Data processing location. This is where Fivetran will operate and run computation on data. If not set, the API default region is used.
- `run_setup_tests` (Boolean) Whether to run setup tests when creating or updating the destination. Defaults to `true`, matching the API.
- `trust_certificates` (Boolean) Specifies whether Fivetran should trust certificates automatically. Defaults to `false`.
- `trust_fingerprints` (Boolean) Specifies whether Fivetran should trust SSH fingerprints automatically. Defaults to `false`.

### Read-Only

- `id` (String) The unique identifier for the destination within the Fivetran system.
- `setup_status` (String) Destination setup status. Possible values: `incomplete`, `connected`, `broken`.
